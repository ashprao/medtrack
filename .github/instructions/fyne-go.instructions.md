---
description: "Use when writing, editing, or reviewing Go code for this Fyne v2 desktop app. Covers view construction, layout, data flow, database access, widget usage, and naming conventions."
applyTo: "internal/ui/**/*.go, cmd/**/*.go, internal/db/**/*.go, internal/models/**/*.go"
---

# MedTrack — Fyne v2 Go Desktop App Guidelines

## Architecture

- This is a **Fyne v2** desktop app (v2.5.4), Go 1.23.2, SQLite via `go-sqlite3`.
- Views never import or call `db.Manager` directly — except `MedicationLog`, which holds a `*db.Manager` for date-range refreshes. All other data mutations go through closures injected by `main.go`.

## App Menu (macOS)

- Use `fyne.NewMainMenu` in `main.go`. The menu is macOS-only; Fyne ignores it on other platforms.
- The About item label must be **exactly `"About"`** (not `"About MedTrack"`). Fyne's macOS driver only moves items named `"About"` into the system app menu automatically.
- Use `app.NewWithID("com.ashprao.medtrack")` (not `app.New()`). The bundle ID is required for stable macOS app-menu integration and must match `FyneApp.toml`.
- Destructive / settings actions (Preferences, data management) belong in the app menu, not on a main tab.

## Navigation Tabs

- Navigation is a single `container.NewAppTabs` in `main.go`. Do not add routers, navigators, or additional windows.
- This app has **4 tabs**: Daily Intake, Medications, Log, Help. Do not add an About tab — About is an app-menu item.

## View Construction Pattern

Every view must follow this exact structure — no deviations:

```go
type ViewName struct {
    container *fyne.Container
    // widget fields (unexported)
    // callback fields: onSave func(...) error, onDelete func(...), etc.
}

func NewViewName(/* callbacks, deps */) *ViewName {
    v := &ViewName{...}
    // build widget tree here
    v.container = container.NewBorder(header, nil, nil, nil, scrollContent)
    return v
}

func (v *ViewName) Container() fyne.CanvasObject { return v.container }
func (v *ViewName) UpdateXxx(data ...) { /* mutate widget state, call .Refresh() */ }
```

- Constructors always use the `New<TypeName>` prefix.
- Expose a `Container() fyne.CanvasObject` method — this is what `main.go` embeds into tabs.
- Push-update methods are named `Update<Data>(...)` — views never pull data themselves.

## Data Flow (Push Model)

```
main.go owns *db.Manager
  └── injects closures into view constructors
        └── after every DB write: calls view.UpdateXxx(freshData)
```

- After any save/delete/intake operation in `main.go`, call **all affected views'** `UpdateXxx` methods.
- Views must never own their own refresh timers or polling loops.

## Layout

- Use `container.NewBorder(header, nil, nil, nil, scrollContent)` as the dominant full-screen layout.
- Scrollable content: wrap `container.NewVBox(...)` in `container.NewVScroll(...)`.
- Use `container.NewStack()` (alias: `container.NewMax()`) only for manual push/pop sub-navigation within a tab.
- Window size is **800×600**. Never call `SetMinSize` on the main window — it causes horizontal overflow. Do not make the window resizable without explicit request.
- Prefer `container.NewPadded` at content boundaries for consistent spacing.
- Do not place item cards with `widget.NewCard` in Daily Intake — use flat `container.NewVBox` rows + `widget.Separator` between items.

## Widgets

| Need | Use |
|---|---|
| Item card | `widget.NewCard(title, subtitle, content)` |
| Data entry form | `widget.NewForm` + `widget.NewFormItem` |
| Dropdowns | `widget.NewSelect` |
| Boolean toggle | `widget.NewCheck` |
| Rich text (Help/About) | `widget.NewRichTextWithText` |
| Large decorative text | `widget.NewRichText` (not `canvas.NewText` — not theme-colour-aware) |
| Section headers | `widget.NewLabelWithStyle` with `Bold: true` |
| Icon buttons | `widget.NewButtonWithIcon` |
| Refresh icon | `theme.ViewRefreshIcon()` |

- Never use `widget.NewTable` for rendered card-style lists — use `container.NewVBox` of card-like rows inside a `container.NewVScroll`.
- Never use `canvas.NewText` for user-facing text — it ignores the Fyne theme colour and won't adapt to light/dark mode.

## Dialogs

- Gate destructive actions (save, delete, clear) with `dialog.ShowConfirm` from `fyne.io/fyne/v2/dialog`.
- For dialogs that need a specific size (e.g. About), use `dialog.NewCustom(...)` and call `dlg.Resize(fyne.NewSize(w, h))` before `dlg.Show()`. Never rely on `dialog.ShowCustom` auto-sizing for text-heavy content — it wraps letter-by-letter.
- Validation failures (empty required fields, bad date format) must surface as `dialog.ShowInformation` or `dialog.ShowError` — never as silent no-ops or `log.Println` only.
- Use `fyne.CurrentApp().Driver().AllWindows()[0]` as the parent window inside closures where a `fyne.Window` reference is not in scope.

## Callbacks and Events

- Callback fields must be named `onSave`, `onDelete`, `onTake`, `onSubmit`, `onCancel`, `onPRNUse` — extend with the `onVerb` pattern for any new callbacks.
- Store callbacks as unexported struct fields of function type (e.g., `onSave func(*models.Medication) error`).
- Inject all callbacks via the constructor — never set them after construction.

## Database Layer (`internal/db`)

- All public methods on `db.Manager` use positional `?` placeholders (SQLite style).
- Wrap `db.Exec` errors through the private `executeQuery` helper — do not call `db.Exec` directly in new methods.
- Use `CREATE TABLE IF NOT EXISTS` — schema changes must be idempotent.
- New queries follow existing patterns: prepare statement → scan rows → close rows before returning.
- Enable `PRAGMA foreign_keys = ON` on every new connection open. Failure to do so causes FK constraint tests to silently pass.
- Schema migrations use `ALTER TABLE` inside `migrateSchema` — never drop and recreate tables.
- Soft-delete pattern: `deleted_at DATETIME` column; filter with `WHERE deleted_at IS NULL` in all reads.

## Daily Intake — Domain Conventions

- Time periods use clinically correct buckets: **Morning** (before 12:00), **Afternoon** (12:00–17:00), **Evening** (17:00–20:59), **Bedtime** (21:00+). Do not add or rename buckets without explicit request.
- Time option labels must show human ranges: `"Morning (9:00 AM – 12:00 PM)"`, `"Afternoon (12:00 PM – 5:00 PM)"`, `"Evening (5:00 PM – 9:00 PM)"`, `"Bedtime (10:00 PM)"`. Stored HH:MM values: `09:00`, `12:00`, `18:00`, `22:00`.
- **PRN (As Needed)** medications (`TimesPerDay == 0`) appear in a dedicated "As Needed" section at the bottom of Daily Intake with a "Record Use" button. They must never generate scheduled phantom intakes via `EnsureDailyIntakes`.
- **Smart taken-at proximity**: if a dose is marked within 2 hours of `ScheduledFor`, record `TakenAt = ScheduledFor`; otherwise record `TakenAt = time.Now()`. The `proximityWindow` constant is `2 * time.Hour`.

## Models (`internal/models`)

- Plain structs with value fields; optional fields use pointer types (e.g., `EndDate *time.Time`).
- Add a `Validate() error` method to any new model.
- `IntakeStatus` is a typed `string` — add new status constants using the same `type IntakeStatus string` pattern with typed constants.
- `Frequency` is parsed from a human-readable string — extend `ParseFrequency` if adding new frequency types; do not store parsed structs in the DB.
- Frequency string format stored in DB: `"<count> at HH:MM[, HH:MM...]"` or `"as needed"`. `ParseFrequency` returns `Frequency{TimesPerDay int, Times []string}` where `TimesPerDay == 0` means PRN.
- At runtime, use `fyne.CurrentApp().Metadata().Version` for the displayed version string; fall back to `internal/version.Version` when `Metadata().Version` is empty (i.e. during `go run`).

## Form UX Conventions

- Placeholder hints use human-readable examples: `"e.g. 2025-06-30"`, not `"YYYY-MM-DD"`.
- Cancel button importance: `widget.LowImportance`. Do not place a separate Clear button on forms.
- **Never use `widget.LowImportance` for visible text labels.** Use `widget.MediumImportance` for secondary/muted text (e.g. dosage, food instructions, captions). Reserve `LowImportance` on labels only for genuinely disabled or placeholder states.
- A form's save action must validate first and show `dialog.ShowError` or `dialog.ShowInformation` for any violation before writing to the DB.

## Naming Conventions

| Item | Convention |
|---|---|
| Files | `snake_case.go` |
| Types | `PascalCase` |
| Constructors | `New<TypeName>(...)` |
| Private helpers | lowercase, e.g. `refresh()`, `updateCache()` |
| Update methods | `Update<Data>(...)` |
| Callback fields | `onVerb` e.g. `onSave`, `onDelete` |

## Go Idioms to Follow

- **Loop variable capture**: always use `med := med` (or equivalent) inside `range` loops that reference `med` in closures.
- **Lazy initialization**: only create sub-views (e.g., form) on first use with a `if v.formView == nil` guard.
- **Mutex**: use `sync.Mutex` only when a view maintains a cache accessed from multiple goroutines (see `DailyIntake`). Do not add mutexes speculatively.

## What Not to Do

- Do not add a router, state manager, or reactive binding library.
- Do not let views import `internal/db` except where already established (`MedicationLog`).
- Do not change the window size or make the window resizable without explicit request.
- Do not add a custom theme unless explicitly asked — the app uses Fyne's default theme.
- Do not use `widget.NewTable` for card-style lists.
- Do not use `canvas.NewText` for user-facing text — use `widget.NewLabel` or `widget.NewRichText`.
- Do not name menu items `"About MedTrack"` — Fyne only auto-integrates items named exactly `"About"` into the macOS app menu.
- Do not call `dialog.ShowCustom` for dialogs that contain paragraphs of text — it auto-sizes and wraps letter-by-letter. Use `dialog.NewCustom` + explicit `Resize`.
- Do not export helper methods that are only used within the package (e.g. `enableAllButtons` must stay unexported).
- Do not poll or use timers inside views for data refresh.
