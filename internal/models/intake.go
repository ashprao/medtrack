package models

import (
	"fmt"
	"time"
)

type Intake struct {
	ID           int64
	MedicationID int64
	ScheduledFor time.Time
	TakenAt      *time.Time // Null until medication is marked as taken
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

// IsValid checks if the status is one of the defined constants
func (s IntakeStatus) IsValid() bool {
	switch s {
	case IntakeStatusPending, IntakeStatusTaken, IntakeStatusMissed, IntakeStatusSkipped:
		return true
	default:
		return false
	}
}

// Validate checks if the intake has all required fields and valid values
func (i *Intake) Validate() error {
	if i.MedicationID == 0 {
		return fmt.Errorf("medication ID is required")
	}
	if i.ScheduledFor.IsZero() {
		return fmt.Errorf("scheduled time is required")
	}
	if i.Status == "" {
		return fmt.Errorf("status is required")
	}
	if !i.Status.IsValid() {
		return fmt.Errorf("invalid status: %s", i.Status)
	}
	if i.Status == IntakeStatusTaken && i.TakenAt == nil {
		return fmt.Errorf("taken time is required when status is taken")
	}
	if i.Status != IntakeStatusTaken && i.TakenAt != nil {
		return fmt.Errorf("taken time should be nil when status is not taken")
	}
	return nil
}

// MarkAsTaken marks the intake as taken at the specified time
func (i *Intake) MarkAsTaken(takenAt time.Time) error {
	i.Status = IntakeStatusTaken
	i.TakenAt = &takenAt
	return i.Validate()
}

// MarkAsPending marks the intake as pending
func (i *Intake) MarkAsPending() error {
	i.Status = IntakeStatusPending
	i.TakenAt = nil
	return i.Validate()
}

// MarkAsMissed marks the intake as missed
func (i *Intake) MarkAsMissed() error {
	i.Status = IntakeStatusMissed
	i.TakenAt = nil
	return i.Validate()
}

// MarkAsSkipped marks the intake as skipped
func (i *Intake) MarkAsSkipped() error {
	i.Status = IntakeStatusSkipped
	i.TakenAt = nil
	return i.Validate()
}
