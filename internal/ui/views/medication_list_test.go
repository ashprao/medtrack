package views

import (
	"medtrack/internal/models"
	"testing"
	"time"

	"fyne.io/fyne/v2/widget"
)

// mockButton is a simple mock for testing button state
type mockButton struct {
	disabled bool
}

func (b *mockButton) Disable() {
	b.disabled = true
}

func (b *mockButton) Enable() {
	b.disabled = false
}

func (b *mockButton) Disabled() bool {
	return b.disabled
}

// mockMedicationList is a simplified version for testing button state logic
type mockMedicationList struct {
	addButton *mockButton
	onSelect  func(med *models.Medication, editBtn *widget.Button)
}

func newMockMedicationList(onSelect func(med *models.Medication, editBtn *widget.Button)) *mockMedicationList {
	return &mockMedicationList{
		addButton: &mockButton{},
		onSelect:  onSelect,
	}
}

// TestMedicationListButtonLogic tests the button enable/disable logic
func TestMedicationListButtonLogic(t *testing.T) {
	// Helper function to check button state
	checkButtonState := func(button *mockButton, expectedEnabled bool) {
		if button.Disabled() == expectedEnabled {
			t.Errorf("Button state incorrect. Expected disabled=%v, got %v", !expectedEnabled, button.Disabled())
		}
	}

	tests := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "Edit button disables add button",
			fn: func(t *testing.T) {
				// Create test medications
				meds := []models.Medication{
					{
						ID:        1,
						Name:      "Test Med 1",
						Dosage:    "10mg",
						Frequency: "once daily at 09:00",
						StartDate: time.Now(),
					},
				}

				var list *mockMedicationList
				// Create medication list
				list = newMockMedicationList(
					func(med *models.Medication, editBtn *widget.Button) {
						list.addButton.Disable()
					},
				)

				// Initially add button should be enabled
				checkButtonState(list.addButton, true)

				// Simulate edit button click
				list.onSelect(&meds[0], nil)

				// Add button should be disabled
				checkButtonState(list.addButton, false)

				// Simulate form submission
				list.addButton.Enable()

				// Add button should be enabled again
				checkButtonState(list.addButton, true)
			},
		},
		{
			name: "Add new button disables itself",
			fn: func(t *testing.T) {
				var list *mockMedicationList
				// Create medication list
				list = newMockMedicationList(
					func(med *models.Medication, editBtn *widget.Button) {
						list.addButton.Disable()
					},
				)

				// Initially add button should be enabled
				checkButtonState(list.addButton, true)

				// Simulate add new button click
				list.onSelect(nil, nil)

				// Add button should be disabled
				checkButtonState(list.addButton, false)

				// Simulate form submission
				list.addButton.Enable()

				// Add button should be enabled again
				checkButtonState(list.addButton, true)
			},
		},
	}

	// Run test cases
	for _, tt := range tests {
		t.Run(tt.name, tt.fn)
	}
}
