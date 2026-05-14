# MedTrack Makefile
# Wraps fyne CLI and fyne-cross for packaging and distribution.
#
# Quick start:
#   make tools          — install fyne and fyne-cross CLIs
#   make build          — compile native binary to build/MedTrack
#   make package        — produce MedTrack.app (debug, local use)
#   make sign-adhoc     — ad-hoc sign MedTrack.app (runs on your Mac only)
#   make release        — version check + release .app + sign + .dmg
#
# Distribution signing:
#   Set DEVELOPER_ID to your certificate common name before calling make:
#     export DEVELOPER_ID="Developer ID Application: Your Name (TEAMID)"
#   Set APPLE_ID, APP_PASSWORD (app-specific), TEAM_ID for notarization:
#     export APPLE_ID="you@example.com"
#     export APP_PASSWORD="xxxx-xxxx-xxxx-xxxx"
#     export TEAM_ID="YOURTEAMID"

# ── Variables ──────────────────────────────────────────────────────────────
APP_NAME    := MedTrack
APP_ID      := com.ashprao.medtrack
BINARY      := $(APP_NAME)
APP_BUNDLE  := $(APP_NAME).app
BUILD_DIR   := build
ICON        := Icon.png
CMD_PKG     := ./cmd/medtrack

# Extract version from source of truth
VERSION     := $(shell grep 'Version = ' internal/version/version.go | cut -d'"' -f2)
DMG_NAME    := $(APP_NAME)-$(VERSION).dmg

# Detect GOPATH/bin for tool availability checks
GOBIN       := $(shell go env GOPATH)/bin

# ── Code signing identities (override via environment) ─────────────────────
# Ad-hoc signing is always available (no Apple account needed).
# Set DEVELOPER_ID to enable Developer ID signing for Gatekeeper-accepted
# distribution. Set APPLE_ID, APP_PASSWORD, TEAM_ID for notarization.
DEVELOPER_ID  ?=
APPLE_ID      ?=
APP_PASSWORD  ?=
TEAM_ID       ?=

.DEFAULT_GOAL := help

# ── Help ───────────────────────────────────────────────────────────────────
.PHONY: help
help:
	@echo ""
	@echo "  MedTrack v$(VERSION) — build targets"
	@echo ""
	@echo "  Setup"
	@echo "    make tools              Install fyne and fyne-cross CLIs"
	@echo ""
	@echo "  Development"
	@echo "    make build              Compile native binary → build/$(BINARY)"
	@echo "    make test               Run all tests"
	@echo "    make run                Build and launch the app"
	@echo ""
	@echo "  Packaging (macOS)"
	@echo "    make package            fyne package → $(APP_BUNDLE)  (debug build)"
	@echo "    make package-release    fyne package → $(APP_BUNDLE)  (stripped)"
	@echo "    make install            fyne install  → /Applications"
	@echo ""
	@echo "  Code Signing"
	@echo "    make sign-adhoc         Ad-hoc sign (local use only, no Apple account)"
	@echo "    make sign               Developer ID sign (set DEVELOPER_ID env var)"
	@echo ""
	@echo "  DMG"
	@echo "    make dmg                Build + ad-hoc sign + wrap into $(DMG_NAME)"
	@echo "    make dmg-signed         Build + Developer ID sign + wrap + sign DMG"
	@echo ""
	@echo "  Notarization  (requires APPLE_ID, APP_PASSWORD, TEAM_ID)"
	@echo "    make notarize           Submit $(DMG_NAME) to Apple for notarization"
	@echo "    make staple             Staple notarization ticket to $(DMG_NAME)"
	@echo ""
	@echo "  Cross-platform"
	@echo "    make dist               fyne-cross darwin (amd64 + arm64)"
	@echo ""
	@echo "  Release"
	@echo "    make check-version      Validate version consistency across files"
	@echo "    make release            check-version + package-release + sign-adhoc + dmg"
	@echo "    make release-signed     check-version + package-release + sign + dmg-signed + notarize + staple"
	@echo ""
	@echo "  Maintenance"
	@echo "    make clean              Remove build artifacts"
	@echo ""

# ── Setup ──────────────────────────────────────────────────────────────────
.PHONY: tools
tools:
	@echo "→ Installing fyne CLI..."
	go install fyne.io/tools/cmd/fyne@latest
	@echo "→ Installing fyne-cross..."
	go install github.com/fyne-io/fyne-cross@latest
	@echo "✓ Tools installed to $(GOBIN)"

# ── Development ────────────────────────────────────────────────────────────
.PHONY: build
build:
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=1 go build -o $(BUILD_DIR)/$(BINARY) $(CMD_PKG)
	@echo "✓ Binary → $(BUILD_DIR)/$(BINARY)"

.PHONY: test
test:
	CGO_ENABLED=1 go test ./...

.PHONY: run
run: build
	./$(BUILD_DIR)/$(BINARY)

# ── Packaging (macOS) ──────────────────────────────────────────────────────
.PHONY: _check_icon
_check_icon:
	@test -f $(ICON) || { echo "✗ $(ICON) not found — add a 512×512 PNG to the project root"; exit 1; }

.PHONY: _check_fyne
_check_fyne:
	@command -v $(GOBIN)/fyne >/dev/null 2>&1 || { \
		echo "✗ fyne CLI not found — run: make tools"; exit 1; }

.PHONY: package
package: _check_icon _check_fyne
	cd $(CMD_PKG) && $(GOBIN)/fyne package -os darwin
	@mv $(CMD_PKG)/$(APP_BUNDLE) . 2>/dev/null || true
	@echo "✓ $(APP_BUNDLE) created"

.PHONY: package-release
package-release: _check_icon _check_fyne
	cd $(CMD_PKG) && $(GOBIN)/fyne package -os darwin -release
	@mv $(CMD_PKG)/$(APP_BUNDLE) . 2>/dev/null || true
	@echo "✓ $(APP_BUNDLE) created (release build)"

# ── Code Signing ───────────────────────────────────────────────────────────
# Ad-hoc signing: the app will run on your Mac only.
# Gatekeeper will block it on other Macs unless they right-click → Open.
.PHONY: sign-adhoc
sign-adhoc:
	@test -d $(APP_BUNDLE) || { echo "✗ $(APP_BUNDLE) not found — run make package first"; exit 1; }
	codesign --sign - --deep --force "$(APP_BUNDLE)"
	@echo "✓ Ad-hoc signed $(APP_BUNDLE)"

# Developer ID signing: accepted by Gatekeeper on any Mac.
# Requires: export DEVELOPER_ID="Developer ID Application: Your Name (TEAMID)"
.PHONY: sign
sign:
	@test -d $(APP_BUNDLE) || { echo "✗ $(APP_BUNDLE) not found — run make package-release first"; exit 1; }
	@test -n "$(DEVELOPER_ID)" || { \
		echo "✗ DEVELOPER_ID is not set."; \
		echo "  Run: export DEVELOPER_ID=\"Developer ID Application: Your Name (TEAMID)\""; \
		echo "  Or use 'make sign-adhoc' for local testing (no Apple account needed)."; \
		exit 1; }
	codesign \
		--sign "$(DEVELOPER_ID)" \
		--deep \
		--force \
		--options runtime \
		--timestamp \
		"$(APP_BUNDLE)"
	@echo "✓ Developer ID signed $(APP_BUNDLE)"
	@codesign --verify --verbose "$(APP_BUNDLE)"

# ── DMG ────────────────────────────────────────────────────────────────────
.PHONY: _make_dmg
_make_dmg:
	@test -d $(APP_BUNDLE) || { echo "✗ $(APP_BUNDLE) not found"; exit 1; }
	hdiutil create \
		-volname "$(APP_NAME)" \
		-srcfolder "$(APP_BUNDLE)" \
		-ov \
		-format UDZO \
		"$(DMG_NAME)"
	@echo "✓ $(DMG_NAME) created"

# DMG with ad-hoc signing (no Apple account required)
.PHONY: dmg
dmg: package-release sign-adhoc _make_dmg

# DMG with Developer ID signing (DEVELOPER_ID must be set)
.PHONY: dmg-signed
dmg-signed: package-release sign _make_dmg
	codesign --sign "$(DEVELOPER_ID)" --timestamp "$(DMG_NAME)"
	@echo "✓ $(DMG_NAME) signed"

# ── Notarization ───────────────────────────────────────────────────────────
# Requires: APPLE_ID, APP_PASSWORD (app-specific password), TEAM_ID
# After notarization, Apple embeds a ticket — staple it so offline verification works.
.PHONY: _check_notarize_env
_check_notarize_env:
	@test -n "$(APPLE_ID)"     || { echo "✗ APPLE_ID is not set";     exit 1; }
	@test -n "$(APP_PASSWORD)" || { echo "✗ APP_PASSWORD is not set"; exit 1; }
	@test -n "$(TEAM_ID)"      || { echo "✗ TEAM_ID is not set";      exit 1; }

.PHONY: notarize
notarize: _check_notarize_env
	@test -f "$(DMG_NAME)" || { echo "✗ $(DMG_NAME) not found — run make dmg-signed first"; exit 1; }
	xcrun notarytool submit "$(DMG_NAME)" \
		--apple-id "$(APPLE_ID)" \
		--password "$(APP_PASSWORD)" \
		--team-id "$(TEAM_ID)" \
		--wait
	@echo "✓ Notarization complete"

.PHONY: staple
staple:
	@test -f "$(DMG_NAME)" || { echo "✗ $(DMG_NAME) not found"; exit 1; }
	xcrun stapler staple "$(DMG_NAME)"
	@echo "✓ Notarization ticket stapled to $(DMG_NAME)"

.PHONY: install
install: _check_icon _check_fyne
	cd $(CMD_PKG) && $(GOBIN)/fyne install
	@echo "✓ $(APP_NAME) installed to /Applications/$(APP_BUNDLE)"

# ── Cross-platform ─────────────────────────────────────────────────────────
.PHONY: dist
dist:
	@command -v $(GOBIN)/fyne-cross >/dev/null 2>&1 || { \
		echo "✗ fyne-cross not found — run: make tools"; exit 1; }
	$(GOBIN)/fyne-cross darwin -arch amd64,arm64 -app-id $(APP_ID) $(CMD_PKG)
	@echo "✓ Cross-compiled bundles → fyne-cross/dist/"

# ── Release ────────────────────────────────────────────────────────────────
.PHONY: check-version
check-version:
	@bash scripts/check-version.sh

# Local release: ad-hoc signed DMG, no Apple account needed
.PHONY: release
release: check-version dmg
	@echo ""
	@echo "✓ Release v$(VERSION) ready: $(DMG_NAME)  (ad-hoc signed)"

# Distribution release: Developer ID signed + notarized DMG
# Requires DEVELOPER_ID, APPLE_ID, APP_PASSWORD, TEAM_ID to be set
.PHONY: release-signed
release-signed: check-version dmg-signed notarize staple
	@echo ""
	@echo "✓ Distribution release v$(VERSION) ready: $(DMG_NAME)  (notarized)"

# ── Maintenance ────────────────────────────────────────────────────────────
.PHONY: clean
clean:
	rm -rf $(BUILD_DIR) $(APP_BUNDLE) *.dmg fyne-cross/
	@echo "✓ Clean"
