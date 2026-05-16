# MedTrack

A cross-platform desktop application for managing medication schedules and tracking daily intake. Built with Go and Fyne v2 — runs natively on macOS, Linux, and Windows from a single codebase.

![Go 1.23+](https://img.shields.io/badge/Go-1.23+-00ADD8?logo=go&logoColor=white)
![Fyne v2](https://img.shields.io/badge/Fyne-v2.7.2-informational)
![License: MIT](https://img.shields.io/badge/License-MIT-yellow)

---

## Overview

MedTrack lets you maintain a medication list, view a daily schedule grouped by time of day, mark doses as taken, and review a historical log with date-range filtering. Data is stored locally in a SQLite database — no network connection required.

![About MedTrack](screenshots/about.png)
*MedTrack v1.2.0 — About dialog*

---

## Screenshots

#### Daily Intake
![Daily Intake](screenshots/daily_intake.png)
*Daily medication schedule grouped by morning, afternoon, evening, and bedtime*

#### Medication Management
![Medication List](screenshots/medication_list.png)
*Medication cards with dosage, timing, and instructions*

#### Add / Edit Medication
![Edit Medication](screenshots/edit_medication.png)
*Form for adding and editing medication details*

---

## Features

- **Medication management** — add, edit, and soft-delete medications with name, dosage, frequency, food instructions, special instructions, notes, and start/end dates
- **Active/Inactive toggle** — pause a medication with the Active checkbox to exclude it from Daily Intake without deleting it; pending intakes are removed, taken entries are preserved
- **Flexible scheduling** — once, twice, three times, or four times daily; bedtime slot (10 PM); or *As Needed* (PRN) for on-demand medications
- **Daily intake view** — doses grouped into Morning / Afternoon / Evening / Bedtime sections; PRN medications shown separately with a "Record Use" button
- **Smart taken-at recording** — if a dose is marked within 2 hours of its scheduled time, the scheduled time is recorded; otherwise the actual clock time is used
- **Schedule-change reconciliation** — editing a medication's schedule immediately removes pending intakes for dropped slots and creates pending intakes for new slots; taken entries are never affected
- **Medication log** — historical intake records with date-range filter, grouped by month and date (newest first), with per-entry delete
- **Data management** — clear intake history or reset all data via Preferences (two-step confirmation required)

---

## Prerequisites

All platforms require **Go 1.23+** and a C compiler (CGO is required by `go-sqlite3`).

### macOS
- **Xcode Command Line Tools**:
  ```bash
  xcode-select --install
  ```
- **gcc and pkg-config** via Homebrew:
  ```bash
  brew install gcc pkg-config
  ```
- **make** — pre-installed via Xcode Command Line Tools
- **fyne CLI** (for packaging and install targets):
  ```bash
  make tools
  ```

### Linux
- **gcc, pkg-config, and OpenGL/X11 headers**:
  ```bash
  sudo apt install gcc pkg-config libgl1-mesa-dev xorg-dev
  ```
  (Fedora/RHEL: `dnf install gcc pkg-config mesa-libGL-devel libXcursor-devel libXrandr-devel libXinerama-devel libXi-devel`)
- **make** — usually pre-installed; otherwise `sudo apt install make`
- **fyne CLI**:
  ```bash
  make tools
  ```

### Windows
- **TDM-GCC** (or MSYS2 MinGW) for CGO support — download from [tdm-gcc.tdragon.net](https://tdm-gcc.tdragon.net)
- **pkg-config** — available via MSYS2 (`pacman -S mingw-w64-x86_64-pkg-config`) or Chocolatey
- **make** — via MSYS2 or `choco install make`
- **fyne CLI**:
  ```bash
  make tools
  ```

---

## Installation

### Option 1 — Quick install (end users)

```bash
go install github.com/ashprao/medtrack/cmd/medtrack@v1.2.0
```

The binary is installed to `$GOPATH/bin` (or `~/go/bin` if `GOPATH` is not set). Ensure that directory is on your `$PATH`.

### Option 2 — Build from source (developers)

```bash
git clone https://github.com/ashprao/medtrack.git
cd medtrack
make tools   # install fyne and fyne-cross CLIs (first time only)
make build   # compile → build/MedTrack
make run     # build and launch
```

Or without make:

```bash
go build -o MedTrack ./cmd/medtrack
./MedTrack
```

### Option 3 — Install as a native app bundle

```bash
make install  # packages and installs to the system app location
```

Fyne resolves the install location per platform: `/Applications` on macOS, `~/.local/share/applications` on Linux, and Program Files on Windows.

---

## Build & Packaging

All common tasks are covered by the `Makefile`. Run `make` (or `make help`) to see all targets.

### Development

| Command | Description |
|---------|-------------|
| `make build` | Compile native binary → `build/MedTrack` |
| `make test` | Run all tests (`CGO_ENABLED=1 go test ./...`) |
| `make run` | Build and launch the app |
| `make clean` | Remove all build artifacts |

### Packaging

`make package` and `make package-release` are platform-aware — they detect the host OS and produce the appropriate artifact:

| Host OS | Artifact |
|---------|----------|
| macOS | `MedTrack.app` bundle |
| Linux | `MedTrack.tar.xz` |
| Windows | `MedTrack.zip` |

| Command | Description |
|---------|-------------|
| `make package` | Create native package (debug build) |
| `make package-release` | Create native package with debug symbols stripped |
| `make install` | Package and install to system app location |

### Cross-platform distribution

Build for any platform from any host using `fyne-cross` (no virtual machine or Docker required when targeting the host OS):

| Command | Description |
|---------|-------------|
| `make dist-darwin` | Build macOS bundles (amd64 + arm64) → `fyne-cross/dist/` |
| `make dist-linux` | Build Linux packages (amd64 + arm64) → `fyne-cross/dist/` |
| `make dist-windows` | Build Windows packages (amd64) → `fyne-cross/dist/` |
| `make dist` | Build all three platforms in one shot |

### macOS Code Signing

| Command | Description |
|---------|-------------|
| `make sign-adhoc` | Ad-hoc sign `MedTrack.app` (runs on your Mac only, no Apple account needed) |
| `make sign` | Developer ID sign (requires `DEVELOPER_ID` env var) |
| `make dmg` | Release app + ad-hoc sign + wrap into `MedTrack-<version>.dmg` |
| `make dmg-signed` | Release app + Developer ID sign + wrap + sign DMG |

**Ad-hoc signed DMG (no Apple account needed)**

`make dmg` produces an ad-hoc signed DMG. Recipients will see a Gatekeeper warning on first launch. Two ways to bypass it:

- **Right-click (or Control-click) → Open** on `MedTrack.app`, then click **Open** in the dialog — grants a permanent exception for that copy.
- Or remove the quarantine flag from the terminal after mounting the DMG:

  ```bash
  xattr -dr com.apple.quarantine /Applications/MedTrack.app
  ```

After the first launch the app runs normally. No certificate or Apple Developer account required.

**For distribution to other Macs with a Developer ID certificate**, set your certificate before running `make dmg-signed`:

```bash
export DEVELOPER_ID="Developer ID Application: Your Name (TEAMID)"
make dmg-signed
```

### Notarization (Apple distribution)

Required for internet distribution without Gatekeeper warnings (macOS 10.15+):

```bash
export DEVELOPER_ID="Developer ID Application: Your Name (TEAMID)"
export APPLE_ID="you@example.com"
export APP_PASSWORD="xxxx-xxxx-xxxx-xxxx"  # app-specific password
export TEAM_ID="YOURTEAMID"
make release-signed   # version check + sign + DMG + notarize + staple
```

---

## Usage

| Tab | Purpose |
|-----|---------|
| **Daily Intake** | View today's doses grouped by time of day; check doses off as taken; record PRN doses |
| **Medications** | Add, edit, or delete medications |
| **Log** | Review historical intake records; set a date range and refresh |
| **Help** | In-app usage guidance |

**About** and **Preferences** (data management) are accessible from the app menu on macOS, and from the menu bar on Linux and Windows.

---

## Project Structure

```
medtrack/
├── cmd/
│   └── medtrack/
│       ├── main.go          # Entry point; owns db.Manager; wires views via closures
│       └── FyneApp.toml     # Bundle ID, version, icon metadata (read by fyne CLI)
├── internal/
│   ├── db/
│   │   ├── database.go      # SQLite manager; schema creation; all DB methods
│   │   └── database_test.go
│   ├── models/
│   │   ├── medication.go    # Medication struct; ParseFrequency
│   │   ├── medication_test.go
│   │   ├── intake.go        # Intake struct; IntakeStatus constants
│   │   └── intake_test.go
│   └── ui/
│       ├── theme/           # Custom Fyne theme (if any)
│       └── views/
│           ├── daily_intake.go      # Daily schedule view
│           ├── medication_list.go   # Medication list + inline form
│           ├── medication_form.go   # Add/edit form component
│           ├── medication_log.go    # Historical log view
│           ├── preferences.go       # Preferences window
│           ├── about.go             # About dialog
│           └── help.go              # Help view
├── docs/
│   ├── architecture.md      # Technical architecture reference
│   └── ai-experiment.md     # Historical: how this project was created
├── Icon.png                 # App icon (512×512 PNG; auto-converted to .icns by fyne)
├── Makefile                 # Build, package, sign, DMG, notarize targets
├── go.mod
└── CHANGELOG.md
```

For a detailed walkthrough of the architecture, data flow, and key design decisions, see [docs/architecture.md](docs/architecture.md).

---

## Running Tests

```bash
make test
```

Or directly:

```bash
CGO_ENABLED=1 go test ./...
```

CGO is required (SQLite). All tests use a temporary on-disk database created via `os.CreateTemp`.

---

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/my-feature`)
3. Add tests for new behaviour
4. Ensure `make test` passes and `make build` is clean
5. Submit a pull request

Please read [docs/architecture.md](docs/architecture.md) before making structural changes — the push-based data flow and view wiring conventions are intentional.

---

## Version History

See [CHANGELOG.md](CHANGELOG.md) for full details.

| Version | Date | Summary |
|---------|------|---------|
| v1.2.0 | May 2026 | Calendar date pickers, date validation, cross-platform packaging |
| v1.1.0 | May 2026 | PRN medications, smart taken-at recording, log layout refactor, bedtime scheduling, soft-delete |
| v0.2.0 | March 2025 | Medication log with date filtering |
| v0.1.0 | February 2025 | Initial release |

---

## License

MIT — see [LICENSE](LICENSE) for details.

---

*This project originated as an AI pair programming experiment. See [docs/ai-experiment.md](docs/ai-experiment.md) for that story.*
