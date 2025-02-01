package models

import (
	"testing"
	"time"
)

func TestIntakeStatus(t *testing.T) {
	tests := []struct {
		name   string
		status IntakeStatus
		valid  bool
	}{
		{
			name:   "pending status",
			status: IntakeStatusPending,
			valid:  true,
		},
		{
			name:   "taken status",
			status: IntakeStatusTaken,
			valid:  true,
		},
		{
			name:   "missed status",
			status: IntakeStatusMissed,
			valid:  true,
		},
		{
			name:   "skipped status",
			status: IntakeStatusSkipped,
			valid:  true,
		},
		{
			name:   "invalid status",
			status: IntakeStatus("invalid"),
			valid:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Note: This assumes you'll add an IsValid method to IntakeStatus
			if got := tt.status.IsValid(); got != tt.valid {
				t.Errorf("IntakeStatus.IsValid() = %v, want %v", got, tt.valid)
			}
		})
	}
}

func TestIntakeValidation(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		intake  Intake
		wantErr bool
	}{
		{
			name: "valid pending intake",
			intake: Intake{
				MedicationID: 1,
				ScheduledFor: now,
				Status:       IntakeStatusPending,
			},
			wantErr: false,
		},
		{
			name: "valid taken intake",
			intake: Intake{
				MedicationID: 1,
				ScheduledFor: now,
				TakenAt:      &now,
				Status:       IntakeStatusTaken,
			},
			wantErr: false,
		},
		{
			name: "missing medication ID",
			intake: Intake{
				ScheduledFor: now,
				Status:       IntakeStatusPending,
			},
			wantErr: true,
		},
		{
			name: "missing scheduled time",
			intake: Intake{
				MedicationID: 1,
				Status:       IntakeStatusPending,
			},
			wantErr: true,
		},
		{
			name: "missing status",
			intake: Intake{
				MedicationID: 1,
				ScheduledFor: now,
			},
			wantErr: true,
		},
		{
			name: "invalid status",
			intake: Intake{
				MedicationID: 1,
				ScheduledFor: now,
				Status:       IntakeStatus("invalid"),
			},
			wantErr: true,
		},
		{
			name: "taken status without taken time",
			intake: Intake{
				MedicationID: 1,
				ScheduledFor: now,
				Status:       IntakeStatusTaken,
			},
			wantErr: true,
		},
		{
			name: "pending status with taken time",
			intake: Intake{
				MedicationID: 1,
				ScheduledFor: now,
				TakenAt:      &now,
				Status:       IntakeStatusPending,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Note: This assumes you'll add a Validate method to Intake
			err := tt.intake.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Intake.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIntakeStatusTransitions(t *testing.T) {
	now := time.Now()
	intake := Intake{
		MedicationID: 1,
		ScheduledFor: now,
		Status:       IntakeStatusPending,
	}

	// Test marking as taken
	if err := intake.MarkAsTaken(now); err != nil {
		t.Errorf("MarkAsTaken() error = %v", err)
	}
	if intake.Status != IntakeStatusTaken {
		t.Errorf("Expected status %v, got %v", IntakeStatusTaken, intake.Status)
	}
	if intake.TakenAt == nil || !intake.TakenAt.Equal(now) {
		t.Error("TakenAt not set correctly")
	}

	// Test marking as pending
	if err := intake.MarkAsPending(); err != nil {
		t.Errorf("MarkAsPending() error = %v", err)
	}
	if intake.Status != IntakeStatusPending {
		t.Errorf("Expected status %v, got %v", IntakeStatusPending, intake.Status)
	}
	if intake.TakenAt != nil {
		t.Error("TakenAt should be nil when pending")
	}

	// Test marking as missed
	if err := intake.MarkAsMissed(); err != nil {
		t.Errorf("MarkAsMissed() error = %v", err)
	}
	if intake.Status != IntakeStatusMissed {
		t.Errorf("Expected status %v, got %v", IntakeStatusMissed, intake.Status)
	}

	// Test marking as skipped
	if err := intake.MarkAsSkipped(); err != nil {
		t.Errorf("MarkAsSkipped() error = %v", err)
	}
	if intake.Status != IntakeStatusSkipped {
		t.Errorf("Expected status %v, got %v", IntakeStatusSkipped, intake.Status)
	}
}
