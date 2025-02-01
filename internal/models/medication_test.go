package models

import (
	"testing"
	"time"
)

func TestParseFrequency(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		wantTimesDaily int
		wantTimes      []string
	}{
		{
			name:           "once daily morning",
			input:          "once daily at 09:00",
			wantTimesDaily: 1,
			wantTimes:      []string{"09:00"},
		},
		{
			name:           "twice daily",
			input:          "twice daily at 09:00, 21:00",
			wantTimesDaily: 2,
			wantTimes:      []string{"09:00", "21:00"},
		},
		{
			name:           "three times daily",
			input:          "three times daily at 09:00, 14:00, 21:00",
			wantTimesDaily: 3,
			wantTimes:      []string{"09:00", "14:00", "21:00"},
		},
		{
			name:           "invalid format defaults to once daily",
			input:          "invalid format",
			wantTimesDaily: 1,
			wantTimes:      []string{"09:00"},
		},
		{
			name:           "empty string defaults to once daily",
			input:          "",
			wantTimesDaily: 1,
			wantTimes:      []string{"09:00"},
		},
		{
			name:           "unsorted times get sorted",
			input:          "three times daily at 21:00, 09:00, 14:00",
			wantTimesDaily: 3,
			wantTimes:      []string{"09:00", "14:00", "21:00"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseFrequency(tt.input)

			if got.TimesPerDay != tt.wantTimesDaily {
				t.Errorf("ParseFrequency() TimesPerDay = %v, want %v", got.TimesPerDay, tt.wantTimesDaily)
			}

			if len(got.Times) != len(tt.wantTimes) {
				t.Errorf("ParseFrequency() Times length = %v, want %v", len(got.Times), len(tt.wantTimes))
			}

			for i, wantTime := range tt.wantTimes {
				if got.Times[i] != wantTime {
					t.Errorf("ParseFrequency() Time[%d] = %v, want %v", i, got.Times[i], wantTime)
				}
			}

			// Verify default days of week (should be all days)
			if len(got.DaysOfWeek) != 7 {
				t.Errorf("ParseFrequency() DaysOfWeek length = %v, want 7", len(got.DaysOfWeek))
			}
			for i := 0; i < 7; i++ {
				found := false
				for _, day := range got.DaysOfWeek {
					if day == i {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("ParseFrequency() DaysOfWeek missing day %d", i)
				}
			}
		})
	}
}

func TestGetDailyTimes(t *testing.T) {
	tests := []struct {
		name     string
		freq     Frequency
		wantLen  int
		checkDay bool // if true, checks if today's times are returned
	}{
		{
			name: "once daily",
			freq: Frequency{
				TimesPerDay: 1,
				Times:       []string{"09:00"},
				DaysOfWeek:  []int{0, 1, 2, 3, 4, 5, 6},
			},
			wantLen:  1,
			checkDay: true,
		},
		{
			name: "twice daily",
			freq: Frequency{
				TimesPerDay: 2,
				Times:       []string{"09:00", "21:00"},
				DaysOfWeek:  []int{0, 1, 2, 3, 4, 5, 6},
			},
			wantLen:  2,
			checkDay: true,
		},
		{
			name: "not scheduled today",
			freq: Frequency{
				TimesPerDay: 1,
				Times:       []string{"09:00"},
				DaysOfWeek:  []int{}, // no days scheduled
			},
			wantLen:  0,
			checkDay: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.freq.GetDailyTimes()

			if len(got) != tt.wantLen {
				t.Errorf("GetDailyTimes() returned %v times, want %v", len(got), tt.wantLen)
			}

			if tt.checkDay && len(got) > 0 {
				now := time.Now()
				today := now.Day()

				for _, timeVal := range got {
					if timeVal.Day() != today {
						t.Errorf("GetDailyTimes() returned time for wrong day: got %v, want day %v", timeVal.Day(), today)
					}
				}
			}

			// Verify times are in chronological order
			for i := 1; i < len(got); i++ {
				if !got[i].After(got[i-1]) {
					t.Errorf("GetDailyTimes() times not in chronological order at index %d", i)
				}
			}
		})
	}
}

func TestMedicationValidation(t *testing.T) {
	now := time.Now()
	tomorrow := now.Add(24 * time.Hour)

	tests := []struct {
		name    string
		med     Medication
		wantErr bool
	}{
		{
			name: "valid medication",
			med: Medication{
				Name:      "Test Med",
				Dosage:    "10mg",
				Frequency: "once daily at 09:00",
				StartDate: now,
			},
			wantErr: false,
		},
		{
			name: "valid medication with end date",
			med: Medication{
				Name:      "Test Med",
				Dosage:    "10mg",
				Frequency: "once daily at 09:00",
				StartDate: now,
				EndDate:   &tomorrow,
			},
			wantErr: false,
		},
		{
			name: "missing name",
			med: Medication{
				Dosage:    "10mg",
				Frequency: "once daily at 09:00",
				StartDate: now,
			},
			wantErr: true,
		},
		{
			name: "missing dosage",
			med: Medication{
				Name:      "Test Med",
				Frequency: "once daily at 09:00",
				StartDate: now,
			},
			wantErr: true,
		},
		{
			name: "missing frequency",
			med: Medication{
				Name:      "Test Med",
				Dosage:    "10mg",
				StartDate: now,
			},
			wantErr: true,
		},
		{
			name: "missing start date",
			med: Medication{
				Name:      "Test Med",
				Dosage:    "10mg",
				Frequency: "once daily at 09:00",
			},
			wantErr: true,
		},
		{
			name: "end date before start date",
			med: Medication{
				Name:      "Test Med",
				Dosage:    "10mg",
				Frequency: "once daily at 09:00",
				StartDate: tomorrow,
				EndDate:   &now,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Note: This assumes you'll add a Validate method to Medication
			err := tt.med.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Medication.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
