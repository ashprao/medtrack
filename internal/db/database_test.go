package db

import (
	"os"
	"testing"
	"time"

	"github.com/ashprao/medtrack/internal/models"
)

func setupTestDB(t *testing.T) (*Manager, func()) {
	// Create a temporary database file
	tmpfile, err := os.CreateTemp("", "medtrack-test-*.db")
	if err != nil {
		t.Fatalf("Could not create temp file: %v", err)
	}
	tmpfile.Close()

	// Initialize database
	manager, err := NewManager(tmpfile.Name())
	if err != nil {
		os.Remove(tmpfile.Name())
		t.Fatalf("Could not create database manager: %v", err)
	}

	// Return cleanup function
	cleanup := func() {
		manager.Close()
		os.Remove(tmpfile.Name())
	}

	return manager, cleanup
}

func TestSaveMedication(t *testing.T) {
	manager, cleanup := setupTestDB(t)
	defer cleanup()

	// Test creating new medication
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

	if err := manager.SaveMedication(med); err != nil {
		t.Errorf("Error saving medication: %v", err)
	}

	if med.ID == 0 {
		t.Error("Expected medication ID to be set after save")
	}

	// Test updating existing medication
	med.Name = "Updated Med"
	if err := manager.SaveMedication(med); err != nil {
		t.Errorf("Error updating medication: %v", err)
	}

	// Verify update
	meds, err := manager.GetMedications()
	if err != nil {
		t.Errorf("Error getting medications: %v", err)
	}

	if len(meds) != 1 {
		t.Errorf("Expected 1 medication, got %d", len(meds))
	}

	if meds[0].Name != "Updated Med" {
		t.Errorf("Expected updated name 'Updated Med', got '%s'", meds[0].Name)
	}
}

func TestGetMedications(t *testing.T) {
	manager, cleanup := setupTestDB(t)
	defer cleanup()

	// Add test medications
	meds := []models.Medication{
		{
			Name:      "Med A",
			Dosage:    "5mg",
			Frequency: "once daily at 09:00",
			StartDate: time.Now(),
		},
		{
			Name:      "Med B",
			Dosage:    "10mg",
			Frequency: "twice daily at 09:00, 21:00",
			StartDate: time.Now(),
		},
	}

	for i := range meds {
		if err := manager.SaveMedication(&meds[i]); err != nil {
			t.Fatalf("Error saving test medication: %v", err)
		}
	}

	// Test retrieval
	retrieved, err := manager.GetMedications()
	if err != nil {
		t.Errorf("Error getting medications: %v", err)
	}

	if len(retrieved) != 2 {
		t.Errorf("Expected 2 medications, got %d", len(retrieved))
	}

	// Verify medications are sorted by name
	if retrieved[0].Name != "Med A" || retrieved[1].Name != "Med B" {
		t.Error("Medications not sorted correctly by name")
	}
}

func TestDeleteMedication(t *testing.T) {
	manager, cleanup := setupTestDB(t)
	defer cleanup()

	// Add a test medication
	med := &models.Medication{
		Name:      "Test Med",
		Dosage:    "5mg",
		Frequency: "once daily at 09:00",
		StartDate: time.Now(),
	}

	if err := manager.SaveMedication(med); err != nil {
		t.Fatalf("Error saving test medication: %v", err)
	}

	// Add an intake for this medication
	intake := &models.Intake{
		MedicationID: med.ID,
		ScheduledFor: time.Now(),
		Status:       models.IntakeStatusPending,
	}

	if err := manager.AddIntake(intake); err != nil {
		t.Fatalf("Error saving test intake: %v", err)
	}

	// Delete the medication
	if err := manager.DeleteMedication(med.ID); err != nil {
		t.Errorf("Error deleting medication: %v", err)
	}

	// Verify medication is deleted
	meds, err := manager.GetMedications()
	if err != nil {
		t.Errorf("Error getting medications: %v", err)
	}

	if len(meds) != 0 {
		t.Error("Medication not deleted")
	}

	// Verify associated intakes are deleted
	intakes, err := manager.GetDailyIntakes(time.Now())
	if err != nil {
		t.Errorf("Error getting intakes: %v", err)
	}

	if len(intakes) != 0 {
		t.Error("Associated intakes not deleted")
	}
}

func TestIntakeOperations(t *testing.T) {
	manager, cleanup := setupTestDB(t)
	defer cleanup()

	// Add a test medication
	med := &models.Medication{
		Name:      "Test Med",
		Dosage:    "5mg",
		Frequency: "once daily at 09:00",
		StartDate: time.Now(),
	}

	if err := manager.SaveMedication(med); err != nil {
		t.Fatalf("Error saving test medication: %v", err)
	}

	// Test adding intake
	now := time.Now()
	intake := &models.Intake{
		MedicationID: med.ID,
		ScheduledFor: now,
		Status:       models.IntakeStatusPending,
	}

	if err := manager.AddIntake(intake); err != nil {
		t.Errorf("Error adding intake: %v", err)
	}

	if intake.ID == 0 {
		t.Error("Expected intake ID to be set after save")
	}

	// Test getting daily intakes
	intakes, err := manager.GetDailyIntakes(now)
	if err != nil {
		t.Errorf("Error getting daily intakes: %v", err)
	}

	if len(intakes) != 1 {
		t.Errorf("Expected 1 intake, got %d", len(intakes))
	}

	// Test updating intake status
	if err := manager.UpdateIntakeStatus(med.ID, now, true); err != nil {
		t.Errorf("Error updating intake status: %v", err)
	}

	// Verify status update
	intakes, err = manager.GetDailyIntakes(now)
	if err != nil {
		t.Errorf("Error getting daily intakes: %v", err)
	}

	if intakes[0].Status != models.IntakeStatusTaken {
		t.Errorf("Expected status %s, got %s", models.IntakeStatusTaken, intakes[0].Status)
	}
}

func TestEnsureDailyIntakes(t *testing.T) {
	manager, cleanup := setupTestDB(t)
	defer cleanup()

	// Add test medications with different frequencies
	meds := []models.Medication{
		{
			Name:      "Once Daily",
			Dosage:    "5mg",
			Frequency: "once daily at 09:00",
			StartDate: time.Now(),
		},
		{
			Name:      "Twice Daily",
			Dosage:    "10mg",
			Frequency: "twice daily at 09:00, 21:00",
			StartDate: time.Now(),
		},
	}

	for i := range meds {
		if err := manager.SaveMedication(&meds[i]); err != nil {
			t.Fatalf("Error saving test medication: %v", err)
		}
	}

	// Test ensuring daily intakes
	if err := manager.EnsureDailyIntakes(time.Now(), meds); err != nil {
		t.Errorf("Error ensuring daily intakes: %v", err)
	}

	// Verify intakes were created
	intakes, err := manager.GetDailyIntakes(time.Now())
	if err != nil {
		t.Errorf("Error getting daily intakes: %v", err)
	}

	expectedIntakes := 3 // 1 for once daily + 2 for twice daily
	if len(intakes) != expectedIntakes {
		t.Errorf("Expected %d intakes, got %d", expectedIntakes, len(intakes))
	}

	// Test idempotency - running again shouldn't create duplicates
	if err := manager.EnsureDailyIntakes(time.Now(), meds); err != nil {
		t.Errorf("Error ensuring daily intakes (second run): %v", err)
	}

	intakes, err = manager.GetDailyIntakes(time.Now())
	if err != nil {
		t.Errorf("Error getting daily intakes after second run: %v", err)
	}

	if len(intakes) != expectedIntakes {
		t.Errorf("Expected %d intakes after second run, got %d", expectedIntakes, len(intakes))
	}
}
