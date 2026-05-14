package db

import (
	"os"
	"testing"
	"time"

	"github.com/ashprao/medtrack/internal/models"
	"github.com/stretchr/testify/assert"
)

func setupTestDB(t *testing.T) (*Manager, func()) {
	tmpfile, err := os.CreateTemp("", "medtrack-test-*.db")
	if err != nil {
		t.Fatalf("Could not create temp file: %v", err)
	}
	tmpfile.Close()

	manager, err := NewManager(tmpfile.Name())
	if err != nil {
		os.Remove(tmpfile.Name())
		t.Fatalf("Could not create database manager: %v", err)
	}

	cleanup := func() {
		manager.Close()
		os.Remove(tmpfile.Name())
	}

	return manager, cleanup
}

func TestCreateTables(t *testing.T) {
	manager, cleanup := setupTestDB(t)
	defer cleanup()

	err := manager.createTables()
	assert.NoError(t, err, "Creating tables on existing DB should not fail")
}

func TestSaveMedication(t *testing.T) {
	manager, cleanup := setupTestDB(t)
	defer cleanup()

	med := &models.Medication{
		Name:            "Test Med",
		Dosage:          "10mg",
		Frequency:       "once daily at 09:00",
		Purpose:         "Testing",
		FoodInstruction: "After food",
		Instructions:    "Take with water",
		Notes:           "Test notes",
		StartDate:       time.Now(),
	}

	assert.NoError(t, manager.SaveMedication(med))
	assert.NotZero(t, med.ID)

	med.Name = "Updated Med"
	assert.NoError(t, manager.SaveMedication(med))
	meds, err := manager.GetMedications()
	assert.NoError(t, err)
	assert.Len(t, meds, 1)
	assert.Equal(t, "Updated Med", meds[0].Name)
}

func TestGetMedications(t *testing.T) {
	manager, cleanup := setupTestDB(t)
	defer cleanup()

	meds := []models.Medication{
		{Name: "Med A", Dosage: "5mg", Frequency: "once daily at 09:00", StartDate: time.Now()},
		{Name: "Med B", Dosage: "10mg", Frequency: "twice daily at 09:00, 21:00", StartDate: time.Now()},
	}

	for i := range meds {
		assert.NoError(t, manager.SaveMedication(&meds[i]))
	}

	retrieved, err := manager.GetMedications()
	assert.NoError(t, err)
	assert.Len(t, retrieved, 2)
	assert.Equal(t, "Med A", retrieved[0].Name)
	assert.Equal(t, "Med B", retrieved[1].Name)
}

func TestDeleteMedication(t *testing.T) {
	manager, cleanup := setupTestDB(t)
	defer cleanup()

	med := &models.Medication{Name: "Test Med", Dosage: "5mg", Frequency: "once daily at 09:00", StartDate: time.Now()}
	assert.NoError(t, manager.SaveMedication(med))

	intake := &models.Intake{MedicationID: med.ID, ScheduledFor: time.Now(), Status: models.IntakeStatusPending}
	assert.NoError(t, manager.AddIntake(intake))

	assert.NoError(t, manager.DeleteMedication(med.ID))

	// Soft delete: medication no longer appears in active list
	meds, err := manager.GetMedications()
	assert.NoError(t, err)
	assert.Empty(t, meds, "Soft-deleted medication should not appear in GetMedications")

	// Intake history is preserved
	intakes, err := manager.GetDailyIntakes(time.Now())
	assert.NoError(t, err)
	assert.Len(t, intakes, 1, "Intakes should be preserved after soft delete")
}

func TestGetAllMedications(t *testing.T) {
	manager, cleanup := setupTestDB(t)
	defer cleanup()

	med := &models.Medication{Name: "Test Med", Dosage: "5mg", Frequency: "once daily at 09:00", StartDate: time.Now()}
	assert.NoError(t, manager.SaveMedication(med))
	assert.NoError(t, manager.DeleteMedication(med.ID))

	// Active list excludes the deleted medication
	active, err := manager.GetMedications()
	assert.NoError(t, err)
	assert.Empty(t, active)

	// GetAllMedications includes it
	all, err := manager.GetAllMedications()
	assert.NoError(t, err)
	assert.Len(t, all, 1)
	assert.NotNil(t, all[0].DeletedAt, "DeletedAt should be set after soft delete")
}

func TestDeleteIntake(t *testing.T) {
	manager, cleanup := setupTestDB(t)
	defer cleanup()

	med := &models.Medication{Name: "Test Med", Dosage: "5mg", Frequency: "once daily at 09:00", StartDate: time.Now()}
	assert.NoError(t, manager.SaveMedication(med))

	intake := &models.Intake{MedicationID: med.ID, ScheduledFor: time.Now(), Status: models.IntakeStatusPending}
	assert.NoError(t, manager.AddIntake(intake))
	assert.NotZero(t, intake.ID)

	assert.NoError(t, manager.DeleteIntake(intake.ID))

	intakes, err := manager.GetDailyIntakes(time.Now())
	assert.NoError(t, err)
	assert.Empty(t, intakes, "Intake should be gone after DeleteIntake")
}

func TestIntakeOperations(t *testing.T) {
	manager, cleanup := setupTestDB(t)
	defer cleanup()

	med := &models.Medication{Name: "Test Med", Dosage: "5mg", Frequency: "once daily at 09:00", StartDate: time.Now()}
	assert.NoError(t, manager.SaveMedication(med))

	now := time.Now()
	intake := &models.Intake{MedicationID: med.ID, ScheduledFor: now, Status: models.IntakeStatusPending}
	assert.NoError(t, manager.AddIntake(intake))
	assert.NotZero(t, intake.ID)

	intakes, err := manager.GetDailyIntakes(now)
	assert.NoError(t, err)
	assert.Len(t, intakes, 1)

	assert.NoError(t, manager.UpdateIntakeStatus(med.ID, now, true))

	intakes, err = manager.GetDailyIntakes(now)
	assert.NoError(t, err)
	assert.Equal(t, models.IntakeStatusTaken, intakes[0].Status)
}

func TestEnsureDailyIntakes(t *testing.T) {
	manager, cleanup := setupTestDB(t)
	defer cleanup()

	meds := []models.Medication{
		{Name: "Once Daily", Dosage: "5mg", Frequency: "once daily at 09:00", StartDate: time.Now()},
		{Name: "Twice Daily", Dosage: "10mg", Frequency: "twice daily at 09:00, 21:00", StartDate: time.Now()},
	}

	for i := range meds {
		assert.NoError(t, manager.SaveMedication(&meds[i]))
	}

	assert.NoError(t, manager.EnsureDailyIntakes(time.Now(), meds))

	intakes, err := manager.GetDailyIntakes(time.Now())
	assert.NoError(t, err)
	expectedIntakes := 3
	assert.Len(t, intakes, expectedIntakes)

	assert.NoError(t, manager.EnsureDailyIntakes(time.Now(), meds))

	intakes, err = manager.GetDailyIntakes(time.Now())
	assert.NoError(t, err)
	assert.Len(t, intakes, expectedIntakes)
}

func TestSaveMedicationWithEmptyName(t *testing.T) {
	manager, cleanup := setupTestDB(t)
	defer cleanup()

	medication := &models.Medication{
		Name:      "",
		Dosage:    "200mg",
		Frequency: "Once a day",
		StartDate: time.Now(),
	}

	err := manager.SaveMedication(medication)
	assert.Error(t, err, "Saving medication with empty name should fail")
}

func TestAddIntakeWithInvalidMedicationID(t *testing.T) {
	manager, cleanup := setupTestDB(t)
	defer cleanup()

	intake := &models.Intake{
		MedicationID: 999,
		ScheduledFor: time.Now(),
		Status:       models.IntakeStatusPending,
	}

	err := manager.AddIntake(intake)
	assert.Error(t, err, "Adding intake with invalid medication ID should fail")
}

func TestGetIntakesForNonExistentDate(t *testing.T) {
	manager, cleanup := setupTestDB(t)
	defer cleanup()

	nonExistentStartDate := time.Date(3000, 1, 1, 0, 0, 0, 0, time.UTC)
	nonExistentEndDate := nonExistentStartDate.Add(24 * time.Hour)

	intakes, err := manager.GetIntakes(nonExistentStartDate, nonExistentEndDate)
	assert.NoError(t, err, "Querying inexistent intakes should not fail")
	assert.Empty(t, intakes, "Non-existent date should return no intakes")
}

func TestUpdateIntakeStatusForInvalidRecord(t *testing.T) {
	manager, cleanup := setupTestDB(t)
	defer cleanup()

	invalidMedicationID := int64(999)
	scheduledTime := time.Now()

	err := manager.UpdateIntakeStatus(invalidMedicationID, scheduledTime, true)
	assert.Error(t, err, "Updating intake status for non-existent record should fail")
}

func TestEnsureDailyIntakesWithNoMedications(t *testing.T) {
	manager, cleanup := setupTestDB(t)
	defer cleanup()

	date := time.Now()

	err := manager.EnsureDailyIntakes(date, []models.Medication{})
	assert.NoError(t, err)

	intakes, err := manager.GetDailyIntakes(date)
	assert.NoError(t, err)
	assert.Empty(t, intakes, "Ensuring daily intakes with no medications should result in no intakes")
}

func TestGetCounts(t *testing.T) {
	manager, cleanup := setupTestDB(t)
	defer cleanup()

	// Empty DB
	meds, intakes, err := manager.GetCounts()
	assert.NoError(t, err)
	assert.Equal(t, 0, meds)
	assert.Equal(t, 0, intakes)

	// Add a medication and ensure daily intakes
	med := &models.Medication{
		Name:      "Count Test Med",
		Dosage:    "10mg",
		Frequency: "once daily at 09:00",
		StartDate: time.Now(),
	}
	assert.NoError(t, manager.SaveMedication(med))
	assert.NoError(t, manager.EnsureDailyIntakes(time.Now(), []models.Medication{*med}))

	meds, intakes, err = manager.GetCounts()
	assert.NoError(t, err)
	assert.Equal(t, 1, meds)
	assert.Equal(t, 1, intakes)
}

func TestClearIntakeHistory(t *testing.T) {
	manager, cleanup := setupTestDB(t)
	defer cleanup()

	med := &models.Medication{
		Name:      "Clear Test Med",
		Dosage:    "10mg",
		Frequency: "once daily at 09:00",
		StartDate: time.Now(),
	}
	assert.NoError(t, manager.SaveMedication(med))
	assert.NoError(t, manager.EnsureDailyIntakes(time.Now(), []models.Medication{*med}))

	// Verify intakes exist
	intakes, err := manager.GetDailyIntakes(time.Now())
	assert.NoError(t, err)
	assert.NotEmpty(t, intakes)

	// Clear and verify
	assert.NoError(t, manager.ClearIntakeHistory())

	intakes, err = manager.GetDailyIntakes(time.Now())
	assert.NoError(t, err)
	assert.Empty(t, intakes, "All intakes should be removed after ClearIntakeHistory")

	// Medications should still exist
	meds, err := manager.GetMedications()
	assert.NoError(t, err)
	assert.Len(t, meds, 1, "Medications should be untouched by ClearIntakeHistory")
}

func TestResetAllData(t *testing.T) {
	manager, cleanup := setupTestDB(t)
	defer cleanup()

	med := &models.Medication{
		Name:      "Reset Test Med",
		Dosage:    "10mg",
		Frequency: "once daily at 09:00",
		StartDate: time.Now(),
	}
	assert.NoError(t, manager.SaveMedication(med))
	assert.NoError(t, manager.EnsureDailyIntakes(time.Now(), []models.Medication{*med}))

	// Verify data exists
	meds, intakes, err := manager.GetCounts()
	assert.NoError(t, err)
	assert.Equal(t, 1, meds)
	assert.Equal(t, 1, intakes)

	// Reset and verify everything is gone
	assert.NoError(t, manager.ResetAllData())

	meds, intakes, err = manager.GetCounts()
	assert.NoError(t, err)
	assert.Equal(t, 0, meds, "All medications should be removed after ResetAllData")
	assert.Equal(t, 0, intakes, "All intakes should be removed after ResetAllData")
}

func TestAddPRNIntake(t *testing.T) {
	manager, cleanup := setupTestDB(t)
	defer cleanup()

	med := &models.Medication{
		Name:      "PRN Med",
		Dosage:    "5mg",
		Frequency: "as needed",
		StartDate: time.Now(),
	}
	assert.NoError(t, manager.SaveMedication(med))

	before := time.Now()
	assert.NoError(t, manager.AddPRNIntake(med.ID))
	after := time.Now()

	intakes, err := manager.GetIntakes(before.Add(-time.Second), after.Add(time.Second))
	assert.NoError(t, err)
	assert.Len(t, intakes, 1)
	assert.Equal(t, models.IntakeStatusTaken, intakes[0].Status)
	assert.NotNil(t, intakes[0].TakenAt)
	// TakenAt should be within the before/after window
	assert.False(t, intakes[0].TakenAt.Before(before), "TakenAt should not be before the call")
	assert.False(t, intakes[0].TakenAt.After(after), "TakenAt should not be after the call")
}

func TestEnsureDailyIntakesPRNSkipped(t *testing.T) {
	manager, cleanup := setupTestDB(t)
	defer cleanup()

	prn := &models.Medication{
		Name:      "PRN Med",
		Dosage:    "5mg",
		Frequency: "as needed",
		StartDate: time.Now(),
	}
	assert.NoError(t, manager.SaveMedication(prn))
	assert.NoError(t, manager.EnsureDailyIntakes(time.Now(), []models.Medication{*prn}))

	_, intakeCount, err := manager.GetCounts()
	assert.NoError(t, err)
	assert.Equal(t, 0, intakeCount, "PRN medications should not generate scheduled intakes")
}

func TestUpdateIntakeStatusAlwaysUsesScheduledTime(t *testing.T) {
	manager, cleanup := setupTestDB(t)
	defer cleanup()

	med := &models.Medication{
		Name:      "Prox Med",
		Dosage:    "10mg",
		Frequency: "once daily at 09:00",
		StartDate: time.Now(),
	}
	assert.NoError(t, manager.SaveMedication(med))

	now := time.Now()
	year, month, day := now.Date()

	t.Run("recent dose uses scheduled time", func(t *testing.T) {
		scheduled := time.Date(year, month, day, now.Hour(), now.Minute()-30, 0, 0, now.Location())
		if now.Minute() < 30 {
			scheduled = time.Date(year, month, day, now.Hour()-1, now.Minute()+30, 0, 0, now.Location())
		}
		intake := &models.Intake{
			MedicationID: med.ID,
			ScheduledFor: scheduled,
			Status:       models.IntakeStatusPending,
		}
		assert.NoError(t, manager.AddIntake(intake))
		assert.NoError(t, manager.UpdateIntakeStatus(med.ID, scheduled, true))

		intakes, err := manager.GetIntakes(scheduled.Add(-time.Minute), scheduled.Add(time.Hour))
		assert.NoError(t, err)
		assert.Len(t, intakes, 1)
		assert.NotNil(t, intakes[0].TakenAt)
		diff := intakes[0].TakenAt.Sub(scheduled)
		if diff < 0 {
			diff = -diff
		}
		assert.Less(t, diff, time.Second, "TakenAt should always equal scheduled time")
	})

	t.Run("late dose also uses scheduled time", func(t *testing.T) {
		if now.Hour() < 3 {
			t.Skip("cannot run late-dose test before 3:00 AM")
		}
		scheduled := time.Date(year, month, day, now.Hour()-3, now.Minute(), 0, 0, now.Location())
		intake := &models.Intake{
			MedicationID: med.ID,
			ScheduledFor: scheduled,
			Status:       models.IntakeStatusPending,
		}
		assert.NoError(t, manager.AddIntake(intake))
		assert.NoError(t, manager.UpdateIntakeStatus(med.ID, scheduled, true))

		intakes, err := manager.GetIntakes(scheduled.Add(-time.Minute), scheduled.Add(time.Hour))
		assert.NoError(t, err)
		assert.NotEmpty(t, intakes)
		for _, it := range intakes {
			if it.TakenAt == nil {
				continue
			}
			diff := it.TakenAt.Sub(scheduled)
			if diff < 0 {
				diff = -diff
			}
			assert.Less(t, diff, time.Second, "TakenAt should equal scheduled time even for late doses")
		}
	})
}

// TestEnsureDailyIntakes_DeactivateMed verifies that deactivating a medication
// (by omitting it from the medications list) deletes its pending intakes for
// today while preserving any already-taken rows.
func TestEnsureDailyIntakes_DeactivateMed(t *testing.T) {
	manager, cleanup := setupTestDB(t)
	defer cleanup()

	now := time.Now()
	med := &models.Medication{
		Name:      "Deactivate Test",
		Dosage:    "10mg",
		Frequency: "twice daily at 09:00, 18:00",
		StartDate: now,
	}
	assert.NoError(t, manager.SaveMedication(med))

	// Create today's two scheduled pending intakes.
	assert.NoError(t, manager.EnsureDailyIntakes(now, []models.Medication{*med}))
	intakes, err := manager.GetDailyIntakes(now)
	assert.NoError(t, err)
	assert.Len(t, intakes, 2)

	// Mark the 9 AM dose as taken.
	year, month, day := now.Date()
	nineAM := time.Date(year, month, day, 9, 0, 0, 0, now.Location())
	assert.NoError(t, manager.UpdateIntakeStatus(med.ID, nineAM, true))

	// Deactivate: call EnsureDailyIntakes with an empty list.
	assert.NoError(t, manager.EnsureDailyIntakes(now, []models.Medication{}))

	// Only the taken 9 AM row should survive; the pending 6 PM row must be deleted.
	intakes, err = manager.GetDailyIntakes(now)
	assert.NoError(t, err)
	assert.Len(t, intakes, 1, "pending intake should be deleted on deactivation")
	assert.Equal(t, models.IntakeStatusTaken, intakes[0].Status, "taken row must be preserved")
}

// TestEnsureDailyIntakes_ScheduleChange_StalePending verifies that changing a
// medication's schedule removes pending rows at times that no longer exist and
// creates new pending rows at the newly added times.
func TestEnsureDailyIntakes_ScheduleChange_StalePending(t *testing.T) {
	manager, cleanup := setupTestDB(t)
	defer cleanup()

	now := time.Now()
	med := &models.Medication{
		Name:      "Schedule Change Test",
		Dosage:    "10mg",
		Frequency: "twice daily at 09:00, 18:00",
		StartDate: now,
	}
	assert.NoError(t, manager.SaveMedication(med))

	// Create today's intakes with the original schedule (9 AM + 6 PM).
	assert.NoError(t, manager.EnsureDailyIntakes(now, []models.Medication{*med}))
	intakes, err := manager.GetDailyIntakes(now)
	assert.NoError(t, err)
	assert.Len(t, intakes, 2)

	// Update to new schedule: 8 AM instead of 9 AM; 6 PM unchanged.
	med.Frequency = "twice daily at 08:00, 18:00"
	assert.NoError(t, manager.SaveMedication(med))
	assert.NoError(t, manager.EnsureDailyIntakes(now, []models.Medication{*med}))

	// 9 AM pending deleted, 8 AM pending created, 6 PM pending preserved → 2 rows.
	intakes, err = manager.GetDailyIntakes(now)
	assert.NoError(t, err)
	assert.Len(t, intakes, 2, "stale 9am should be replaced by 8am")

	byTime := make(map[string]models.IntakeStatus)
	for _, i := range intakes {
		byTime[i.ScheduledFor.Format("15:04")] = i.Status
	}
	assert.Equal(t, models.IntakeStatusPending, byTime["08:00"], "8am pending should be created")
	assert.Equal(t, models.IntakeStatusPending, byTime["18:00"], "6pm pending should be preserved")
	_, has9 := byTime["09:00"]
	assert.False(t, has9, "stale 9am pending should be deleted")
}

// TestEnsureDailyIntakes_TimeBucketGuard verifies that shifting a schedule time
// within the same time-of-day bucket (e.g. 9 AM → 8 AM, both Morning) does not
// create a new pending row when the original time has already been taken.
func TestEnsureDailyIntakes_TimeBucketGuard(t *testing.T) {
	manager, cleanup := setupTestDB(t)
	defer cleanup()

	now := time.Now()
	med := &models.Medication{
		Name:      "Bucket Guard Test",
		Dosage:    "10mg",
		Frequency: "once daily at 09:00",
		StartDate: now,
	}
	assert.NoError(t, manager.SaveMedication(med))

	// Create the 9 AM intake and mark it taken.
	assert.NoError(t, manager.EnsureDailyIntakes(now, []models.Medication{*med}))
	year, month, day := now.Date()
	nineAM := time.Date(year, month, day, 9, 0, 0, 0, now.Location())
	assert.NoError(t, manager.UpdateIntakeStatus(med.ID, nineAM, true))

	// Shift schedule to 8 AM (still Morning bucket).
	med.Frequency = "once daily at 08:00"
	assert.NoError(t, manager.SaveMedication(med))
	assert.NoError(t, manager.EnsureDailyIntakes(now, []models.Medication{*med}))

	// Taken 9 AM preserved; no new 8 AM pending (Morning bucket already taken).
	intakes, err := manager.GetDailyIntakes(now)
	assert.NoError(t, err)
	assert.Len(t, intakes, 1, "bucket guard should prevent a duplicate pending row")
	assert.Equal(t, models.IntakeStatusTaken, intakes[0].Status)
	assert.Equal(t, "09:00", intakes[0].ScheduledFor.Format("15:04"), "taken 9am row must be preserved")
}

// TestEnsureDailyIntakes_ScheduleChange_CrossBucket verifies the full cross-bucket
// case: the morning time shifts within the same bucket (suppressed by guard), and
// the evening time moves to a different bucket (old pending deleted, new pending created).
func TestEnsureDailyIntakes_ScheduleChange_CrossBucket(t *testing.T) {
	manager, cleanup := setupTestDB(t)
	defer cleanup()

	now := time.Now()
	med := &models.Medication{
		Name:      "Cross Bucket Test",
		Dosage:    "10mg",
		Frequency: "twice daily at 09:00, 18:00",
		StartDate: now,
	}
	assert.NoError(t, manager.SaveMedication(med))

	// Create intakes and mark 9 AM taken.
	assert.NoError(t, manager.EnsureDailyIntakes(now, []models.Medication{*med}))
	year, month, day := now.Date()
	nineAM := time.Date(year, month, day, 9, 0, 0, 0, now.Location())
	assert.NoError(t, manager.UpdateIntakeStatus(med.ID, nineAM, true))

	// New schedule: 9 AM → 8 AM (same Morning bucket), 6 PM → 9 PM (Evening → Bedtime).
	med.Frequency = "twice daily at 08:00, 21:00"
	assert.NoError(t, manager.SaveMedication(med))
	assert.NoError(t, manager.EnsureDailyIntakes(now, []models.Medication{*med}))

	// Expected: taken 9 AM preserved, 8 AM suppressed (Morning taken),
	// old 6 PM deleted, new 9 PM pending created → 2 rows total.
	intakes, err := manager.GetDailyIntakes(now)
	assert.NoError(t, err)
	assert.Len(t, intakes, 2, "should have taken 9am + pending 9pm only")

	byTime := make(map[string]models.IntakeStatus)
	for _, i := range intakes {
		byTime[i.ScheduledFor.Format("15:04")] = i.Status
	}
	assert.Equal(t, models.IntakeStatusTaken, byTime["09:00"], "taken 9am must be preserved")
	assert.Equal(t, models.IntakeStatusPending, byTime["21:00"], "9pm pending must be created")
	_, has8 := byTime["08:00"]
	assert.False(t, has8, "8am pending must be suppressed by bucket guard")
	_, has18 := byTime["18:00"]
	assert.False(t, has18, "stale 6pm pending must be deleted")
}
