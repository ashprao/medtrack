# MedTrack — Architecture Reference

This document is the primary technical reference for contributors. It covers the tech stack, data flow, database design, models, views, and the reasoning behind key design decisions.

---

## Tech Stack

| Component | Technology | Version |
|-----------|-----------|---------|
| Language | Go | 1.23.2 |
| GUI framework | [Fyne v2](https://fyne.io) | 2.7.2 |
| Database driver | [go-sqlite3](https://github.com/mattn/go-sqlite3) | 1.14.24 |
| App metadata | FyneApp.toml | — |

**CGO is required.** `go-sqlite3` compiles a C SQLite amalgamation. Ensure `gcc` and `pkg-config` are available before building (see README prerequisites for per-platform instructions).

**Cross-platform by design.** Go and Fyne v2 were chosen specifically so MedTrack compiles and runs natively on macOS, Linux, and Windows from a single codebase. The data layer, models, and core UI are fully platform-agnostic. Some UI behaviours noted in this document (app menu placement, About/Preferences lift) are Fyne platform-specific features on macOS — they are implementation details, not platform restrictions.

**App identity:** `com.ashprao.medtrack` (defined in `FyneApp.toml`). Fyne uses this as the macOS bundle identifier, Linux desktop entry ID, and Windows registry key for `fyne/app.NewWithID`.

---

## Project Layout

```
medtrack/
├── cmd/
│   └── medtrack/
│       └── main.go          # Entry point; owns db.Manager; wires views together
├── internal/
│   ├── db/
│   │   ├── database.go      # Manager type; schema; all DB methods
│   │   └── database_test.go
│   ├── models/
│   │   ├── medication.go    # Medication struct; Frequency; ParseFrequency
│   │   ├── medication_test.go
│   │   ├── intake.go        # Intake struct; IntakeStatus constants
│   │   └── intake_test.go
│   └── ui/
│       ├── theme/           # Custom Fyne theme overrides (if any)
│       └── views/
│           ├── daily_intake.go      # Today's dose schedule
│           ├── medication_list.go   # List view + inline form
│           ├── medication_form.go   # Reusable add/edit form component
│           ├── medication_log.go    # Historical log with date filter
│           ├── preferences.go       # Preferences window (data management)
│           ├── about.go             # About dialog
│           └── help.go              # Help tab content
├── docs/
│   ├── architecture.md      # This file
│   └── ai-experiment.md     # Historical origin story
├── FyneApp.toml             # Bundle ID, version, build number
├── go.mod
└── CHANGELOG.md
```

---

## Data Flow

MedTrack uses a **push-based, closure-injection pattern**. There are no global variables, no Fyne data bindings, and no circular package imports.

```
main.go
  │
  ├── creates *db.Manager
  │
  ├── defines closures that capture db.Manager:
  │     saveHandler    func(med *models.Medication) error
  │     deleteHandler  func(med *models.Medication)
  │     onTake         func(med *models.Medication, scheduledTime time.Time, taken bool)
  │     onPRNUse       func(medID int64)
  │
  ├── passes closures into view constructors
  │     NewMedicationList(saveHandler, deleteHandler, window)
  │     NewDailyIntake(onTake, onPRNUse)
  │     NewMedicationLog(dbManager)   ← direct reference; log is read-heavy
  │
  └── calls view.Update*() methods to push fresh data after any mutation
```

**Views never call the database directly**, except `MedicationLog` which holds a `*db.Manager` reference for its own refresh cycle. All mutations go through the closures defined in `main.go`.

**UI refresh lifecycle:**
1. User action triggers a closure (e.g., `onTake`)
2. Closure calls the relevant `db.Manager` method
3. Closure re-fetches affected data (medications, intakes)
4. Closure calls `view.UpdateMedications()` / `view.UpdateIntakes()` with fresh data
5. View calls its internal `updateCache()` then `refresh()` to rebuild the UI

---

## Database

**File location:** `~/.medtrack.db` (SQLite)

### Schema

#### `medications`

```sql
CREATE TABLE IF NOT EXISTS medications (
    id               INTEGER PRIMARY KEY AUTOINCREMENT,
    name             TEXT    NOT NULL,
    dosage           TEXT    NOT NULL,
    frequency        TEXT    NOT NULL,
    purpose          TEXT,
    food_instruction TEXT    DEFAULT 'No Food Restriction',
    instructions     TEXT,
    notes            TEXT,
    start_date       DATETIME NOT NULL,
    end_date         DATETIME,
    created_at       DATETIME NOT NULL,
    updated_at       DATETIME NOT NULL,
    deleted_at       DATETIME            -- NULL = active; non-NULL = soft-deleted
)
```

#### `intakes`

```sql
CREATE TABLE IF NOT EXISTS intakes (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    medication_id INTEGER  NOT NULL,
    scheduled_for DATETIME NOT NULL,
    taken_at      DATETIME,             -- NULL until marked taken
    status        TEXT     NOT NULL,    -- "pending" | "taken" | "missed" | "skipped"
    notes         TEXT,
    created_at    DATETIME NOT NULL,
    updated_at    DATETIME NOT NULL,
    FOREIGN KEY(medication_id) REFERENCES medications(id)
)
```

### Migrations

`migrateSchema()` is called in `NewManager` after `createTables`. Each migration is idempotent — it checks for the column before running `ALTER TABLE`. Currently one migration: adding `deleted_at` to `medications`.

Pattern for new migrations:
```go
_, err = db.Exec(`ALTER TABLE foo ADD COLUMN bar TEXT`)
// ignore "duplicate column" errors
```

### Foreign key enforcement

```go
db.Exec(`PRAGMA foreign_keys = ON`)
```

Called once on every new connection. Without this, SQLite ignores FK constraints by default.

### Soft-delete

`DeleteMedication(id)` sets `deleted_at = NOW()` rather than issuing a `DELETE`. This preserves intake history for deleted medications.

- `GetMedications()` — filters `WHERE deleted_at IS NULL` (active medications only)
- `GetAllMedications()` — no filter (used for admin/reporting)
- Intake records are never cascade-deleted when a medication is soft-deleted

---

## Models

### `Medication` — `internal/models/medication.go`

```go
type Medication struct {
    ID              int64
    Name            string
    Dosage          string
    Frequency       string      // See frequency string format below
    Purpose         string
    FoodInstruction string      // "Before Food" | "After Food" | "With Food" | "No Food Restriction"
    Instructions    string
    Notes           string
    StartDate       time.Time
    EndDate         *time.Time  // nil = no end date
    CreatedAt       time.Time
    UpdatedAt       time.Time
    DeletedAt       *time.Time  // nil = active; non-nil = soft-deleted
}
```

#### Frequency string format

Frequency is stored as a human-readable string in the database:

| Display name | Stored string |
|---|---|
| Once Daily | `"once daily at 09:00"` |
| Twice Daily | `"twice daily at 09:00, 14:00"` |
| Three Times Daily | `"three times daily at 09:00, 14:00, 18:00"` |
| Four Times Daily | `"four times daily at 09:00, 14:00, 18:00, 22:00"` |
| As Needed (PRN) | `"as needed"` |

`ParseFrequency(freq string) Frequency` parses this string into structured data:

```go
type Frequency struct {
    TimesPerDay int      // 0 for PRN, 1–4 for scheduled
    Times       []string // "HH:MM" strings, sorted ascending
    DaysOfWeek  []int    // defaults to [0,1,2,3,4,5,6] (every day)
}
```

`GetDailyTimes() []time.Time` expands `Times` into concrete `time.Time` values for today.

### `Intake` — `internal/models/intake.go`

```go
type Intake struct {
    ID           int64
    MedicationID int64
    ScheduledFor time.Time   // When the dose is scheduled
    TakenAt      *time.Time  // nil until marked taken; see proximity window below
    Status       IntakeStatus
    Notes        string
    CreatedAt    time.Time
    UpdatedAt    time.Time
}

type IntakeStatus string

const (
    IntakeStatusPending IntakeStatus = "pending"
    IntakeStatusTaken   IntakeStatus = "taken"
    IntakeStatusMissed  IntakeStatus = "missed"
    IntakeStatusSkipped IntakeStatus = "skipped"
)
```

---

## Database Manager — `internal/db/database.go`

```go
type Manager struct { db *sql.DB }

func NewManager(dbPath string) (*Manager, error)
```

### Method reference

| Method | Description |
|--------|-------------|
| `SaveMedication(med)` | INSERT or UPDATE based on `med.ID == 0` |
| `GetMedications()` | Active medications only (`deleted_at IS NULL`) |
| `GetAllMedications()` | All medications including soft-deleted |
| `DeleteMedication(id)` | Soft-delete: sets `deleted_at = now` |
| `AddIntake(intake)` | INSERT a new intake record |
| `EnsureDailyIntakes(date, meds)` | Idempotently creates pending intakes for all scheduled doses on `date`; skips PRN medications |
| `GetDailyIntakes(date)` | Returns all intakes for a given date |
| `GetIntakes(from, to)` | Returns intakes within a date range |
| `UpdateIntakeStatus(medID, scheduledTime, taken)` | Marks a dose taken/pending; applies proximity window logic for `taken_at` |
| `AddPRNIntake(medID)` | Records an immediate PRN dose with `status=taken`, `taken_at=now` |
| `DeleteIntake(id)` | Hard-deletes a single intake record |
| `GetCounts()` | Returns `(medCount, intakeCount)` for Preferences UI |
| `ClearIntakeHistory()` | Deletes all intake records |
| `ResetAllData()` | Deletes all intakes then all medications in a transaction |
| `Close()` | Closes the database connection |

---

## Views — `internal/ui/views/`

All views expose a `Container() fyne.CanvasObject` method that returns the root widget for embedding into the tab layout.

### `MedicationList`

- Holds a `*MedicationForm` internally; toggles between list view and form view by swapping content in `mainContainer`
- `showFormView(med)` — `nil` for new, non-nil for edit
- `UpdateMedications([]Medication)` — rebuilds the card list

### `DailyIntake`

- Holds `medications []Medication` and `prnMedications []Medication` (split in `updateCache`)
- `updateCache()` — called under mutex; separates PRN medications from scheduled ones; resolves each scheduled slot against the `intakes` slice to determine current status
- `refresh()` — rebuilds `content` VBox; groups scheduled items into Morning / Afternoon / Evening / Bedtime; appends PRN section if `len(prnMedications) > 0`
- Concurrency: `mu sync.Mutex` guards `cachedItems` and `prnMedications`; `UpdateMedications` and `UpdateIntakes` both call `updateCache` then `refresh`

**Time buckets:**

| Section | Hour range |
|---------|-----------|
| Morning | `hour < 12` |
| Afternoon | `12 ≤ hour < 17` |
| Evening | `17 ≤ hour < 21` |
| Bedtime | `hour ≥ 21` |

### `MedicationLog`

- Holds a direct `*db.Manager` reference — the only view that does so
- Default date range: last 7 days
- `RefreshData()` re-queries and calls `renderTable()`
- `renderTable()` groups entries by date (newest first), renders a 2-line card per entry: `[HH:MM · TimeSlot]` + status + delete button on top, bold wrapping medication name + dosage below

### `MedicationForm`

- Dynamically creates 0–4 time `widget.Select` widgets based on the selected frequency
- `getFrequencyString()` serialises the current selections to the frequency string format
- `parseFrequency(freq)` deserialises a stored frequency string back into form selections
- `"As Needed"` selection produces `count = 0`, no time selects rendered

### `Preferences`

- Fixed-size window (`420×260`)
- Fetches counts on open; disables "Clear History" when `intakeCount == 0`, disables "Reset All" when both counts are zero
- Two-step confirmation: primary dialog → type `"DELETE"` to confirm

### `About`

- `ShowAboutDialog(parent fyne.Window)` — `dialog.NewCustom` resized to `360×260`
- Menu item label must be exactly `"About"` for Fyne's macOS driver to move it to the app menu

---

## Key Design Decisions

### 1. Soft-delete for medications

Deleting a medication sets `deleted_at` instead of issuing a `DELETE`. This means historical intake records remain intact and queryable in the log view. The active medication list filters to `deleted_at IS NULL`.

**Alternative considered:** Cascade-delete intakes on medication delete. Rejected because it destroys meaningful history (e.g., a discontinued medication should still appear in past log entries).

### 2. `EnsureDailyIntakes` is idempotent

Called on every app launch and after every medication save. It checks for existing intake records before inserting, so running it multiple times is safe. PRN medications (`TimesPerDay == 0`) are skipped — they generate no scheduled intakes.

### 3. Proximity window for `taken_at`

```go
const proximityWindow = 2 * time.Hour
```

When `UpdateIntakeStatus` marks a dose as taken:
- If `|now - scheduledFor| ≤ 2 hours` → `taken_at = scheduledFor` (user logged on time, just forgot to tap immediately)
- If `|now - scheduledFor| > 2 hours` → `taken_at = now` (genuinely late dose)

This avoids the common UX issue of a 9 AM dose showing "Taken at 2:00 PM" just because the user opened the app in the afternoon to check it off.

### 4. PRN ("As Needed") medications

Medications with `Frequency = "as needed"` parse to `TimesPerDay = 0`. They:
- Are excluded from `EnsureDailyIntakes`
- Are separated from scheduled medications in `DailyIntake.updateCache`
- Appear in their own "As Needed" section with a "Record Use" button
- Generate an intake via `AddPRNIntake` with `status=taken` and `taken_at=now`

### 5. Push-based UI updates (no data bindings)

Fyne provides reactive data bindings, but MedTrack does not use them. Instead, `main.go` explicitly calls `view.UpdateMedications()` / `view.UpdateIntakes()` after every mutation. This was a deliberate choice for predictability — it's always clear what data is being pushed and when.

**Trade-off:** More boilerplate in `main.go` closures; fewer surprise re-renders.

---

## Testing

```bash
go test ./...
```

All tests require CGO (`go-sqlite3`). Cross-compilation without a C toolchain will fail.

**Test helper — `internal/db/database_test.go`:**

```go
func setupTestDB(t *testing.T) (*Manager, func()) {
    // Creates a temp file, opens a Manager, returns a cleanup func
}
```

Each test gets an isolated on-disk database. The cleanup function closes the manager and removes the file.

**Coverage areas:**
- `internal/models` — `ParseFrequency` (all frequency strings including PRN), `GetDailyTimes`
- `internal/db` — CRUD for medications and intakes, soft-delete, `EnsureDailyIntakes` (including PRN skip), `UpdateIntakeStatus` proximity window (within-window and outside-window cases), `AddPRNIntake`, `ClearIntakeHistory`, `ResetAllData`
