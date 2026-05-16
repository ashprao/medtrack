# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v1.2.0] - 2026-05-14
### Added
- **Active/Inactive toggle** — each medication card has an Active checkbox; unchecking it pauses the medication so it is excluded from Daily Intake without deleting it
- **PRN (As-Needed) medications** — new frequency option for on-demand medications; appear in a separate "As Needed" section with a "Record Use" button
- **Smart taken-at recording** — marking a dose within 2 hours of its scheduled time records the scheduled time; otherwise the actual clock time is used
- **Bedtime slot** — dedicated 10 PM scheduling slot added alongside Morning, Afternoon, and Evening
- **Log layout refactor** — historical log groups entries by month and date using an accordion layout (newest first)
- **Soft-delete for medications** — deleting a medication from the UI marks it inactive rather than hard-deleting, preserving log history
- **Calendar date pickers** — medication form start/end dates and log filter From/To dates now have a calendar picker button alongside the text entry
- **Date format hint** — a subtle label under each date field shows the expected `YYYY-MM-DD` format with today's date as an example
- **Min-date enforcement on To field** — the log filter To date cannot be set to a date before the From date, enforced in both the text entry validator and the calendar picker
- **Linux and Windows packaging** — `make package` and `make package-release` are now platform-aware; they detect the host OS and produce `.app`, `.tar.xz`, or `.zip` accordingly
- **Cross-platform dist targets** — `make dist-darwin`, `make dist-linux`, `make dist-windows` build for each platform via `fyne-cross`; `make dist` builds all three in one shot

### Changed
- **Schedule-change reconciliation** — when a medication's frequency or time slots are edited, pending intakes for removed slots are deleted immediately and new pending intakes are created for added slots; taken/skipped/missed entries are always preserved
- **Deactivation reconciliation** — pausing a medication (Active unchecked) removes only pending intakes for today; taken entries are preserved
- **Fyne upgraded v2.5.4 → v2.7.2** — required by the fyne-x Calendar widget
- **Cross-platform first** — MedTrack is a first-class cross-platform app; Linux and Windows are fully supported build targets, not afterthoughts
- **Makefile `_check_macos` guard** — macOS-only targets (sign, DMG, notarize) now fail fast with a clear error on non-macOS hosts instead of silent tool-not-found failures

### Fixed
- **Future-date blocking on log filter** — From and To fields reject dates beyond today; the calendar picker shows an error dialog if a future date is tapped
- **End date ≥ start date validation** — medication form submit now shows an error dialog if the end date is before the start date
- **From ≤ To validation on Refresh** — log filter Refresh shows a clear error dialog if To is before From
- **Timezone bug** — today's date was incorrectly rejected as "future" in UTC+ timezones; fixed by parsing dates in local time instead of UTC

## [v0.2.0] - 2025-03-28
### Added
- New "Log" tab for historical medication intake tracking
  - View medication intake history in a tabular format
  - Filter by date range
  - Group entries by date
  - Display detailed status information including time taken

## [v0.1.0] - 2025-02-08
### Added
- Initial release of MedTrack
- Core medication management functionality
  - Add, edit, and delete medications
  - Track dosage, frequency, and timing
  - Record food-related instructions
  - Store special instructions and notes
  - Set start and end dates
- Daily intake tracking
  - View medications organized by time of day
  - Track medication intake with checkboxes
  - Group medications by morning, afternoon, and night
- SQLite database for local storage
- Cross-platform GUI using Fyne
  - Clean, intuitive interface
  - Mobile-friendly design
  - Responsive layout
