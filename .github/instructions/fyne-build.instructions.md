---
description: "Use when writing, editing, or reviewing Makefile, FyneApp.toml, build scripts, packaging, code signing, or DMG creation for this Fyne v2 macOS desktop app."
applyTo: "Makefile, FyneApp.toml, cmd/medtrack/FyneApp.toml, scripts/**"
---

# MedTrack — Build & Packaging Guidelines

## Makefile as the Primary Interface

- All build, test, package, sign, and release tasks are driven by `make`. Developers should not need to remember raw `fyne` or `hdiutil` invocations.
- `make help` (the default target) must always print an up-to-date, formatted list of all targets with one-line descriptions.
- Add new targets to the help output at the same time as the target itself.

## FyneApp.toml Placement

- `FyneApp.toml` must live in `cmd/medtrack/` — the directory where `fyne package` and `fyne install` are run. Fyne only reads the metadata file from the current working directory.
- The `Icon` path in `FyneApp.toml` is relative to `cmd/medtrack/`. Point it to `../../Icon.png` so the icon lives at the repo root and is easy to find.
- `FyneApp.toml` carries `[Details]` with `Name`, `ID`, `Version`, `Build`, `Icon`. The `ID` must match `app.NewWithID(...)` in `main.go`.

```toml
Website = "https://github.com/ashprao/medtrack"

[Details]
  Name    = "MedTrack"
  ID      = "com.ashprao.medtrack"
  Version = "1.1.0"
  Build   = 1
  Icon    = "../../Icon.png"
```

## Version Source of Truth

- `internal/version/version.go` is the canonical source: `const Version = "x.y.z"`.
- `FyneApp.toml` `Version` field must be kept in sync manually when bumping.
- The Makefile derives `VERSION` at build time:
  ```makefile
  VERSION := $(shell grep 'Version = ' internal/version/version.go | cut -d'"' -f2)
  ```
- `scripts/check-version.sh` validates that `version.go`, `FyneApp.toml`, `CHANGELOG.md`, and `README.md` all agree. Run it via `make check-version` before cutting a release.

## fyne CLI Usage

- Always `cd` into the main package directory before invoking `fyne`:
  ```makefile
  cd ./cmd/medtrack && $(GOBIN)/fyne package -os darwin
  ```
- `fyne package` produces `MedTrack.app` inside `cmd/medtrack/`. Move it to the repo root afterward so downstream targets (`dmg`, `sign`) can find it:
  ```makefile
  @mv $(CMD_PKG)/$(APP_BUNDLE) . 2>/dev/null || true
  ```
- `fyne install` has no `--src` flag — it must also be run from the package directory.
- `fyne package` flags:
  - `-os darwin` — target platform (single dash)
  - `-release` — strip debug symbols for distribution (single dash)
  - No `-src` / `--src` needed when running from the package directory.
- Install the fyne CLI via `go install fyne.io/tools/cmd/fyne@latest`. Wrap this in `make tools`.

## App Bundle Structure

After `fyne package -os darwin`, `MedTrack.app` contains:

```
MedTrack.app/
  Contents/
    Info.plist          ← populated from FyneApp.toml
    MacOS/
      medtrack          ← the compiled binary (lowercase)
    Resources/
      icon.icns         ← auto-converted from Icon.png by fyne
```

- `fyne` auto-converts `Icon.png` → `icon.icns`. No manual `iconutil` step is needed.
- `Icon.png` should be **512×512** (or larger, square). Fyne scales down as needed.

## Code Signing

Three tiers — choose based on distribution intent:

### Ad-hoc (local use only)

No Apple account required. The app runs on your Mac but Gatekeeper blocks it on others.

```makefile
codesign --sign - --deep --force "$(APP_BUNDLE)"
```

### Developer ID (Gatekeeper-accepted distribution)

Requires an Apple Developer account and a "Developer ID Application" certificate in Keychain.

```makefile
codesign \
  --sign "$(DEVELOPER_ID)" \
  --deep \
  --force \
  --options runtime \
  --timestamp \
  "$(APP_BUNDLE)"
```

- `--options runtime` enables Hardened Runtime — **required** for notarization.
- `--timestamp` embeds Apple's trusted timestamp — required for long-term validity.
- Set `DEVELOPER_ID` as an environment variable, never hardcode it:
  ```bash
  export DEVELOPER_ID="Developer ID Application: Your Name (TEAMID)"
  ```
- Also sign the DMG itself after wrapping:
  ```makefile
  codesign --sign "$(DEVELOPER_ID)" --timestamp "$(DMG_NAME)"
  ```

### Notarization (internet distribution, macOS 10.15+)

Required to avoid Gatekeeper warnings when users download from the internet.

```makefile
xcrun notarytool submit "$(DMG_NAME)" \
  --apple-id "$(APPLE_ID)" \
  --password "$(APP_PASSWORD)" \
  --team-id "$(TEAM_ID)" \
  --wait
xcrun stapler staple "$(DMG_NAME)"
```

- `APP_PASSWORD` is an **app-specific password** generated at appleid.apple.com — never the Apple ID login password.
- All four env vars (`DEVELOPER_ID`, `APPLE_ID`, `APP_PASSWORD`, `TEAM_ID`) must be set before calling `make release-signed`.
- Stapling embeds the notarization ticket into the DMG so offline Gatekeeper checks work.

## DMG Creation

Use `hdiutil` (macOS built-in — no extra dependency):

```makefile
hdiutil create \
  -volname "$(APP_NAME)" \
  -srcfolder "$(APP_BUNDLE)" \
  -ov \
  -format UDZO \
  "$(DMG_NAME)"
```

- `-format UDZO` = zlib-compressed read-only image (standard for distribution).
- `-ov` = overwrite if DMG already exists.
- Name the DMG `AppName-x.y.z.dmg` using the `VERSION` variable.

## Release Targets

| Target | When to Use |
|---|---|
| `make release` | Local or internal distribution — ad-hoc signed DMG, no Apple account |
| `make release-signed` | Public internet distribution — Developer ID + notarized + stapled DMG |

`make release-signed` pipeline:
1. `make check-version` — abort if versions are inconsistent
2. `make package-release` — stripped `.app` bundle
3. `make sign` — Developer ID codesign with Hardened Runtime
4. `make _make_dmg` — wrap into versioned DMG
5. DMG signing — sign the DMG itself
6. `make notarize` — submit to Apple, wait for approval
7. `make staple` — staple the notarization ticket

## Cross-Platform Builds (fyne-cross)

- Use `fyne-cross darwin -arch amd64,arm64` for CI builds or building arm64 + amd64 separately.
- On macOS, `fyne-cross` delegates to the native `fyne` CLI (no Docker required for darwin→darwin).
- Output lands in `fyne-cross/dist/` — add to `.gitignore`.
- Install via `go install github.com/fyne-io/fyne-cross@latest` (included in `make tools`).

## What Not to Do

- Do not place `FyneApp.toml` at the repo root — `fyne` won't find it when run from `cmd/medtrack/`.
- Do not use `-src` or `--src` with `fyne package` or `fyne install` — these flags do not exist.
- Do not hardcode signing identities, Apple IDs, or passwords in the Makefile — always use env vars.
- Do not skip `--options runtime` on `codesign` if you intend to notarize — it is a hard prerequisite.
- Do not use `APP_PASSWORD` for the Apple ID login password — it must be an app-specific password.
- Do not run `make release-signed` without first confirming `make check-version` passes.
