package db

import (
	"database/sql"
	"fmt"
	"time"

	"medtrack/internal/models"

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

	manager := &Manager{db: db}
	if err := manager.createTables(); err != nil {
		return nil, fmt.Errorf("error creating tables: %v", err)
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

func (m *Manager) SaveMedication(med *models.Medication) error {
	now := time.Now()

	if med.ID == 0 {
		// New medication
		query := `
			INSERT INTO medications (
				name, dosage, frequency, purpose, food_instruction, instructions, notes,
				start_date, end_date, created_at, updated_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`

		result, err := m.db.Exec(query,
			med.Name, med.Dosage, med.Frequency, med.Purpose, med.FoodInstruction, med.Instructions, med.Notes,
			med.StartDate, med.EndDate, now, now,
		)
		if err != nil {
			return fmt.Errorf("error inserting medication: %v", err)
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
				updated_at = ?
			WHERE id = ?
		`

		_, err := m.db.Exec(query,
			med.Name, med.Dosage, med.Frequency, med.Purpose, med.FoodInstruction, med.Instructions, med.Notes,
			med.StartDate, med.EndDate, now, med.ID,
		)
		if err != nil {
			return fmt.Errorf("error updating medication: %v", err)
		}
	}

	med.UpdatedAt = now
	return nil
}

func (m *Manager) GetMedications() ([]models.Medication, error) {
	query := `
		SELECT id, name, dosage, frequency, purpose, food_instruction, instructions, notes,
			   start_date, end_date, created_at, updated_at
		FROM medications
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
		err := rows.Scan(
			&med.ID, &med.Name, &med.Dosage, &med.Frequency,
			&purpose, &foodInstruction, &instructions, &notes, &med.StartDate,
			&med.EndDate, &med.CreatedAt, &med.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning medication row: %v", err)
		}

		// Convert NULL strings to empty strings
		med.Purpose = purpose.String
		med.FoodInstruction = foodInstruction.String
		if med.FoodInstruction == "" {
			med.FoodInstruction = "No Food Restriction"
		}
		med.Instructions = instructions.String
		med.Notes = notes.String

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
	// Get existing intakes for the day
	intakes, err := m.GetDailyIntakes(date)
	if err != nil {
		return err
	}

	// Track which medications already have intakes for specific times
	existingIntakes := make(map[string]bool) // key: "medID_time"
	for _, intake := range intakes {
		key := fmt.Sprintf("%d_%s", intake.MedicationID, intake.ScheduledFor.Format("15:04"))
		existingIntakes[key] = true
	}

	// Create intakes for medications at their scheduled times
	for _, med := range medications {
		freq := models.ParseFrequency(med.Frequency)
		times := freq.Times

		for _, timeStr := range times {
			// Parse the time string (HH:MM)
			hour, min := 0, 0
			if _, err := fmt.Sscanf(timeStr, "%d:%d", &hour, &min); err != nil {
				continue
			}

			// Create scheduled time for this dose
			year, month, day := date.Date()
			scheduledTime := time.Date(year, month, day, hour, min, 0, 0, date.Location())
			key := fmt.Sprintf("%d_%s", med.ID, scheduledTime.Format("15:04"))

			if !existingIntakes[key] {
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
	}

	return nil
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
		takenAt = &now
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

func (m *Manager) DeleteMedication(id int64) error {
	// First delete all intakes for this medication
	_, err := m.db.Exec(`DELETE FROM intakes WHERE medication_id = ?`, id)
	if err != nil {
		return fmt.Errorf("error deleting medication intakes: %v", err)
	}

	// Then delete the medication
	_, err = m.db.Exec(`DELETE FROM medications WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("error deleting medication: %v", err)
	}

	return nil
}

func (m *Manager) Close() error {
	return m.db.Close()
}
