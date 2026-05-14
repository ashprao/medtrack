package db

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/ashprao/medtrack/internal/models"

	_ "github.com/mattn/go-sqlite3"
)

type Manager struct {
	db *sql.DB
}

func NewManager(dbPath string) (*Manager, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %v", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("error connecting to database: %v", err)
	}

	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return nil, fmt.Errorf("error enabling foreign keys: %v", err)
	}

	manager := &Manager{db: db}
	if err := manager.createTables(); err != nil {
		return nil, fmt.Errorf("error creating tables: %v", err)
	}
	if err := manager.migrateSchema(); err != nil {
		return nil, fmt.Errorf("error migrating schema: %v", err)
	}

	return manager, nil
}

func (m *Manager) createTables() error {
	// First create tables
	queries := []string{
		`CREATE TABLE IF NOT EXISTS medications (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			dosage TEXT NOT NULL,
			frequency TEXT NOT NULL,
			purpose TEXT,
			food_instruction TEXT DEFAULT 'No Food Restriction',
			instructions TEXT,
			notes TEXT,
			start_date DATETIME NOT NULL,
			end_date DATETIME,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS intakes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			medication_id INTEGER NOT NULL,
			scheduled_for DATETIME NOT NULL,
			taken_at DATETIME,
			status TEXT NOT NULL,
			notes TEXT,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL,
			FOREIGN KEY(medication_id) REFERENCES medications(id)
		)`,
	}

	// Execute table creation queries
	for _, query := range queries {
		if _, err := m.db.Exec(query); err != nil {
			return fmt.Errorf("error executing query %q: %v", query, err)
		}
	}

	return nil
}

// migrateSchema applies incremental schema changes to existing databases.
// Each ALTER TABLE is safe to run multiple times — the duplicate column error
// from SQLite is caught and ignored (SQLite has no ADD COLUMN IF NOT EXISTS).
func (m *Manager) migrateSchema() error {
	migrations := []string{
		`ALTER TABLE medications ADD COLUMN deleted_at DATETIME`,
		`ALTER TABLE medications ADD COLUMN active INTEGER NOT NULL DEFAULT 1`,
	}
	for _, q := range migrations {
		if _, err := m.db.Exec(q); err != nil {
			// SQLite returns "duplicate column name" when the column already exists.
			// Ignore that specific error; surface everything else.
			if !isDuplicateColumnError(err) {
				return fmt.Errorf("migration failed %q: %v", q, err)
			}
		}
	}
	return nil
}

func isDuplicateColumnError(err error) bool {
	return err != nil && len(err.Error()) >= 21 && err.Error()[:21] == "duplicate column name"
}

// Add a utility function for executing queries
func executeQuery(db *sql.DB, query string, args ...interface{}) (sql.Result, error) {
	result, err := db.Exec(query, args...)
	if err != nil {
		return nil, fmt.Errorf("error executing query %q: %v", query, err)
	}
	return result, nil
}

// Replace repeated query execution logic with the utility function
func (m *Manager) SaveMedication(med *models.Medication) error {
	// If name of the medicine is empty return error
	if med.Name == "" {
		return errors.New("name of the medicine cannot be empty")
	}

	now := time.Now()
	if med.ID == 0 {
		// New medication
		query := `
			INSERT INTO medications (
				name, dosage, frequency, purpose, food_instruction, instructions, notes,
				start_date, end_date, active, created_at, updated_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`
		result, err := executeQuery(m.db, query,
			med.Name, med.Dosage, med.Frequency, med.Purpose, med.FoodInstruction, med.Instructions, med.Notes,
			med.StartDate, med.EndDate, med.Active, now, now,
		)
		if err != nil {
			return err
		}
		id, err := result.LastInsertId()
		if err != nil {
			return fmt.Errorf("error getting last insert id: %v", err)
		}
		med.ID = id
		med.CreatedAt = now
	} else {
		// Update existing medication
		query := `
			UPDATE medications SET
				name = ?, dosage = ?, frequency = ?, purpose = ?, food_instruction = ?,
				instructions = ?, notes = ?, start_date = ?, end_date = ?,
				active = ?, updated_at = ?
			WHERE id = ?
		`
		_, err := executeQuery(m.db, query,
			med.Name, med.Dosage, med.Frequency, med.Purpose, med.FoodInstruction, med.Instructions, med.Notes,
			med.StartDate, med.EndDate, med.Active, now, med.ID,
		)
		if err != nil {
			return err
		}
	}
	med.UpdatedAt = now
	return nil
}

// UpdateMedicationActive toggles the active flag for a single medication without
// re-validating or rewriting all other fields. Called directly from the UI toggle.
func (m *Manager) UpdateMedicationActive(id int64, active bool) error {
	_, err := executeQuery(m.db,
		`UPDATE medications SET active = ?, updated_at = ? WHERE id = ?`,
		active, time.Now(), id,
	)
	return err
}

func (m *Manager) GetMedications() ([]models.Medication, error) {
	return m.queryMedications(`WHERE deleted_at IS NULL`)
}

// GetAllMedications returns all medications including soft-deleted ones.
// Used by the Log view so historical entries still resolve a medication name.
func (m *Manager) GetAllMedications() ([]models.Medication, error) {
	return m.queryMedications(``)
}

func (m *Manager) queryMedications(whereClause string) ([]models.Medication, error) {
	query := `
		SELECT id, name, dosage, frequency, purpose, food_instruction, instructions, notes,
			   start_date, end_date, created_at, updated_at, deleted_at, active
		FROM medications
		` + whereClause + `
		ORDER BY name
	`

	rows, err := m.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error querying medications: %v", err)
	}
	defer rows.Close()

	var medications []models.Medication
	for rows.Next() {
		var med models.Medication
		var purpose, foodInstruction, instructions, notes sql.NullString
		var deletedAt sql.NullTime
		var activeInt int
		err := rows.Scan(
			&med.ID, &med.Name, &med.Dosage, &med.Frequency,
			&purpose, &foodInstruction, &instructions, &notes,
			&med.StartDate, &med.EndDate, &med.CreatedAt, &med.UpdatedAt, &deletedAt, &activeInt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning medication row: %v", err)
		}

		med.Purpose = purpose.String
		med.FoodInstruction = foodInstruction.String
		if med.FoodInstruction == "" {
			med.FoodInstruction = "No Food Restriction"
		}
		med.Instructions = instructions.String
		med.Notes = notes.String
		if deletedAt.Valid {
			med.DeletedAt = &deletedAt.Time
		}
		med.Active = activeInt != 0

		medications = append(medications, med)
	}

	return medications, nil
}

func (m *Manager) AddIntake(intake *models.Intake) error {
	query := `
		INSERT INTO intakes (
			medication_id, scheduled_for, taken_at, status,
			notes, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now()
	result, err := m.db.Exec(query,
		intake.MedicationID, intake.ScheduledFor, intake.TakenAt,
		intake.Status, intake.Notes, now, now,
	)
	if err != nil {
		return fmt.Errorf("error inserting intake: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("error getting last insert id: %v", err)
	}

	intake.ID = id
	intake.CreatedAt = now
	intake.UpdatedAt = now

	return nil
}

func (m *Manager) GetDailyIntakes(date time.Time) ([]models.Intake, error) {
	// Get start and end of the day
	year, month, day := date.Date()
	startOfDay := time.Date(year, month, day, 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	return m.GetIntakes(startOfDay, endOfDay)
}

func (m *Manager) EnsureDailyIntakes(date time.Time, medications []models.Medication) error {
	// Get existing intakes for the day.
	intakes, err := m.GetDailyIntakes(date)
	if err != nil {
		return err
	}

	// Build the set of valid "medID_HH:MM" keys from the current active schedule.
	// Any pending row not in this set is stale (med deactivated or time removed).
	validKeys := make(map[string]bool)
	for _, med := range medications {
		freq := models.ParseFrequency(med.Frequency)
		if freq.TimesPerDay == 0 {
			continue
		}
		for _, timeStr := range freq.Times {
			hour, min := 0, 0
			if _, err := fmt.Sscanf(timeStr, "%d:%d", &hour, &min); err != nil {
				continue
			}
			year, month, day := date.Date()
			t := time.Date(year, month, day, hour, min, 0, 0, date.Location())
			validKeys[fmt.Sprintf("%d_%s", med.ID, t.Format("15:04"))] = true
		}
	}

	// Delete stale pending rows: pending intakes whose med/time is no longer in
	// the active schedule. Covers both deactivation and schedule changes.
	// Taken/skipped rows are never touched — they are permanent medical records.
	for _, intake := range intakes {
		if intake.Status != models.IntakeStatusPending {
			continue
		}
		key := fmt.Sprintf("%d_%s", intake.MedicationID, intake.ScheduledFor.Format("15:04"))
		if !validKeys[key] {
			if err := m.DeleteIntake(intake.ID); err != nil {
				return fmt.Errorf("error removing stale pending intake %d: %v", intake.ID, err)
			}
		}
	}

	// Build a set of time-of-day buckets that already have a taken/skipped dose
	// for each medication. Used below to avoid creating a duplicate pending row
	// when a schedule time shifts within the same bucket (e.g. 9am → 8am, both Morning).
	takenBuckets := make(map[string]bool) // key: "medID_bucket"
	for _, intake := range intakes {
		if intake.Status == models.IntakeStatusTaken || intake.Status == models.IntakeStatusSkipped {
			bucket := timeBucket(intake.ScheduledFor.Hour())
			takenBuckets[fmt.Sprintf("%d_%s", intake.MedicationID, bucket)] = true
		}
	}

	// Track existing intake slots (by exact time) to avoid duplicate inserts.
	existingIntakes := make(map[string]bool) // key: "medID_HH:MM"
	for _, intake := range intakes {
		key := fmt.Sprintf("%d_%s", intake.MedicationID, intake.ScheduledFor.Format("15:04"))
		existingIntakes[key] = true
	}

	// Create pending intakes for medications at their scheduled times.
	for _, med := range medications {
		freq := models.ParseFrequency(med.Frequency)

		// PRN (as-needed) medications do not get scheduled intakes.
		if freq.TimesPerDay == 0 {
			continue
		}

		for _, timeStr := range freq.Times {
			hour, min := 0, 0
			if _, err := fmt.Sscanf(timeStr, "%d:%d", &hour, &min); err != nil {
				continue
			}

			year, month, day := date.Date()
			scheduledTime := time.Date(year, month, day, hour, min, 0, 0, date.Location())
			key := fmt.Sprintf("%d_%s", med.ID, scheduledTime.Format("15:04"))

			if existingIntakes[key] {
				continue
			}

			// Skip if this time-of-day bucket already has a taken/skipped dose.
			// Prevents a phantom pending row when the schedule shifts within a bucket.
			if takenBuckets[fmt.Sprintf("%d_%s", med.ID, timeBucket(hour))] {
				continue
			}

			intake := &models.Intake{
				MedicationID: med.ID,
				ScheduledFor: scheduledTime,
				Status:       models.IntakeStatusPending,
			}
			if err := m.AddIntake(intake); err != nil {
				return fmt.Errorf("error creating intake for %s at %s: %v", med.Name, timeStr, err)
			}
		}
	}

	return nil
}

// timeBucket returns the time-of-day bucket name for a given hour.
// Thresholds match those used in daily_intake.go and medication_log.go.
func timeBucket(hour int) string {
	switch {
	case hour < 12:
		return "Morning"
	case hour < 17:
		return "Afternoon"
	case hour < 21:
		return "Evening"
	default:
		return "Bedtime"
	}
}

// AddPRNIntake records an immediate "as needed" dose for the given medication.
// The intake is created with status=taken and takenAt=now.
func (m *Manager) AddPRNIntake(medID int64) error {
	now := time.Now()
	intake := &models.Intake{
		MedicationID: medID,
		ScheduledFor: now,
		Status:       models.IntakeStatusTaken,
		TakenAt:      &now,
	}
	return m.AddIntake(intake)
}

func (m *Manager) UpdateIntakeStatus(medicationID int64, scheduledTime time.Time, taken bool) error {
	status := models.IntakeStatusTaken
	if !taken {
		status = models.IntakeStatusPending
	}

	// Get the specific intake for this medication and time
	query := `
		SELECT id
		FROM intakes
		WHERE medication_id = ? AND 
			  strftime('%H:%M', scheduled_for) = strftime('%H:%M', ?)
		ORDER BY ABS(STRFTIME('%s', scheduled_for) - STRFTIME('%s', ?))
		LIMIT 1
	`

	var intakeID int64
	now := time.Now()
	err := m.db.QueryRow(query, medicationID, scheduledTime, now).Scan(&intakeID)
	if err != nil {
		return fmt.Errorf("error finding intake record: %v", err)
	}

	// Update the specific intake record
	updateQuery := `
		UPDATE intakes SET
			status = ?,
			taken_at = ?,
			updated_at = ?
		WHERE id = ?
	`

	var takenAt *time.Time
	if taken {
		// Always record the scheduled time as TakenAt — compliance tracking cares
		// about whether the dose was taken, not the exact clock time of the tap.
		t := scheduledTime
		takenAt = &t
	}

	_, err = m.db.Exec(updateQuery, status, takenAt, now, intakeID)
	if err != nil {
		return fmt.Errorf("error updating intake status: %v", err)
	}

	return nil
}

func (m *Manager) GetIntakes(from, to time.Time) ([]models.Intake, error) {
	query := `
		SELECT id, medication_id, scheduled_for, taken_at,
			   status, notes, created_at, updated_at
		FROM intakes
		WHERE scheduled_for BETWEEN ? AND ?
		ORDER BY scheduled_for
	`

	rows, err := m.db.Query(query, from, to)
	if err != nil {
		return nil, fmt.Errorf("error querying intakes: %v", err)
	}
	defer rows.Close()

	var intakes []models.Intake
	for rows.Next() {
		var intake models.Intake
		err := rows.Scan(
			&intake.ID, &intake.MedicationID, &intake.ScheduledFor,
			&intake.TakenAt, &intake.Status, &intake.Notes,
			&intake.CreatedAt, &intake.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning intake row: %v", err)
		}
		intakes = append(intakes, intake)
	}

	return intakes, nil
}

func (m *Manager) DeleteIntake(id int64) error {
	_, err := m.db.Exec(`DELETE FROM intakes WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("error deleting intake: %v", err)
	}
	return nil
}

func (m *Manager) DeleteMedication(id int64) error {
	now := time.Now()
	_, err := m.db.Exec(
		`UPDATE medications SET deleted_at = ?, updated_at = ? WHERE id = ?`,
		now, now, id,
	)
	if err != nil {
		return fmt.Errorf("error soft-deleting medication: %v", err)
	}
	return nil
}

// GetCounts returns the total number of medications and intake records.
// Used to show the user the exact impact before a destructive action.
func (m *Manager) GetCounts() (medCount int, intakeCount int, err error) {
	if err = m.db.QueryRow(`SELECT COUNT(*) FROM medications`).Scan(&medCount); err != nil {
		return 0, 0, fmt.Errorf("error counting medications: %v", err)
	}
	if err = m.db.QueryRow(`SELECT COUNT(*) FROM intakes`).Scan(&intakeCount); err != nil {
		return 0, 0, fmt.Errorf("error counting intakes: %v", err)
	}
	return medCount, intakeCount, nil
}

// MarkMissedIntakes marks all pending intakes whose scheduled time is before
// 'before' as missed. Call this on app start to catch doses from previous days
// that were never confirmed.
func (m *Manager) MarkMissedIntakes(before time.Time) error {
	_, err := m.db.Exec(
		`UPDATE intakes SET status = ?, updated_at = ? WHERE status = ? AND scheduled_for < ?`,
		models.IntakeStatusMissed, time.Now(), models.IntakeStatusPending, before,
	)
	if err != nil {
		return fmt.Errorf("error marking missed intakes: %v", err)
	}
	return nil
}

// ClearIntakeHistory permanently deletes all intake records.
func (m *Manager) ClearIntakeHistory() error {
	if _, err := m.db.Exec(`DELETE FROM intakes`); err != nil {
		return fmt.Errorf("error clearing intake history: %v", err)
	}
	return nil
}

// ResetAllData permanently deletes all intakes and medications in a single
// transaction so a partial failure cannot leave the DB in an inconsistent state.
func (m *Manager) ResetAllData() error {
	tx, err := m.db.Begin()
	if err != nil {
		return fmt.Errorf("error starting transaction: %v", err)
	}
	if _, err = tx.Exec(`DELETE FROM intakes`); err != nil {
		tx.Rollback()
		return fmt.Errorf("error deleting intakes: %v", err)
	}
	if _, err = tx.Exec(`DELETE FROM medications`); err != nil {
		tx.Rollback()
		return fmt.Errorf("error deleting medications: %v", err)
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("error committing reset: %v", err)
	}
	return nil
}

func (m *Manager) Close() error {
	return m.db.Close()
}
