package views

import (
	"fmt"
	"log"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/ashprao/medtrack/internal/models"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// DailyIntake struct holds the data and UI components for displaying a daily intake of medications
// This is the Publisher or Subject in the Observer pattern.
type DailyIntake struct {
	container   *fyne.Container
	medications []models.Medication
	dateLabel   *widget.Label
	content     *fyne.Container
	// onTake is a function called when a medication is taken or skipped. This is the Observer
	// in the Observer pattern
	onTake func(med *models.Medication, scheduledTime time.Time, taken bool)
	// intakes is a slice of all recorded intakes to track multiple doses
	intakes []models.Intake
	// cachedItems is a cache for the calculated intake items
	cachedItems []intakeItem
	// mu protects concurrent access to data by locking it when accessing or modifying shared resources
	mu sync.Mutex
}

type intakeItem struct {
	medication models.Medication
	intake     models.Intake
}

func NewDailyIntake(onTake func(med *models.Medication, scheduledTime time.Time, taken bool)) *DailyIntake {
	d := &DailyIntake{
		dateLabel:   widget.NewLabel(""),
		onTake:      onTake,
		medications: make([]models.Medication, 0),
		intakes:     make([]models.Intake, 0),
		cachedItems: make([]intakeItem, 0),
	}

	// Set initial date label with larger text
	d.dateLabel.TextStyle = fyne.TextStyle{Bold: true}
	d.dateLabel.Alignment = fyne.TextAlignCenter
	d.updateDateLabel()

	// Create header
	header := container.NewVBox(
		d.dateLabel,
		widget.NewSeparator(),
	)

	// Create scrollable content area
	d.content = container.NewVBox()
	scrollContent := container.NewVScroll(d.content)
	scrollContent.SetMinSize(fyne.NewSize(600, 400))

	// Create main container
	d.container = container.NewBorder(
		header,        // Top
		nil,           // Bottom
		nil,           // Left
		nil,           // Right
		scrollContent, // Center
	)

	log.Printf("Daily intake view created")
	return d
}

func (d *DailyIntake) refresh() {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Clear existing content
	d.content.RemoveAll()

	// Show message if no items
	if len(d.cachedItems) == 0 {
		d.content.Add(widget.NewLabel("No medications scheduled for today"))
		return
	}

	// Group items by time
	morning := make([]intakeItem, 0)
	afternoon := make([]intakeItem, 0)
	night := make([]intakeItem, 0)

	for _, item := range d.cachedItems {
		hour := item.intake.ScheduledFor.Hour()
		switch {
		case hour < 12: // Morning (before noon)
			morning = append(morning, item)
		case hour < 17: // Afternoon (12-5 PM)
			afternoon = append(afternoon, item)
		default: // Night (5 PM onwards)
			night = append(night, item)
		}
	}

	// Helper function to create a card for an intake
	createCard := func(item intakeItem) *widget.Card {
		check := widget.NewCheck("", nil)
		check.SetChecked(item.intake.Status == models.IntakeStatusTaken)
		check.OnChanged = func(taken bool) {
			if d.onTake != nil {
				med := item.medication
				log.Printf("Marking medication %s for time %v as %v", med.Name, item.intake.ScheduledFor, taken)
				d.onTake(&med, item.intake.ScheduledFor, taken)
			}
		}

		// Create main content container
		mainContent := container.NewVBox()

		// Create top row with checkbox and food instructions
		topRow := container.NewHBox(check)
		if item.medication.FoodInstruction != "" && item.medication.FoodInstruction != "No Food Restriction" {
			foodLabel := widget.NewLabelWithStyle(
				fmt.Sprintf("Take %s", strings.ToLower(item.medication.FoodInstruction)),
				fyne.TextAlignLeading,
				fyne.TextStyle{Bold: true},
			)
			topRow.Add(foodLabel)
		}
		mainContent.Add(topRow)

		// Add special instructions if any
		if item.medication.Instructions != "" {
			mainContent.Add(widget.NewLabel(item.medication.Instructions))
		}

		// Create card with all details
		return widget.NewCard(
			item.medication.Name,
			item.medication.Dosage,
			mainContent,
		)
	}

	// Helper function to add a time section
	addTimeSection := func(title string, items []intakeItem) {
		if len(items) > 0 {
			// Create centered section header
			header := widget.NewLabelWithStyle(
				title,
				fyne.TextAlignCenter,
				fyne.TextStyle{Bold: true},
			)
			separator := widget.NewSeparator()

			// Add header with padding and center alignment
			headerContainer := container.NewHBox(
				layout.NewSpacer(),
				container.NewPadded(header),
				layout.NewSpacer(),
			)
			d.content.Add(headerContainer)
			d.content.Add(separator)

			// Add cards for this time period
			for _, item := range items {
				d.content.Add(createCard(item))
			}

			// Add spacing after section
			d.content.Add(widget.NewLabel("")) // Empty label for spacing
		}
	}

	// Add each time section
	addTimeSection("Morning (9:00 AM)", morning)
	addTimeSection("Afternoon (3:00 PM)", afternoon)
	addTimeSection("Night (9:00 PM)", night)
}

// updateCache rebuilds the cached items list
func (d *DailyIntake) updateCache() {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Clear existing cache
	d.cachedItems = make([]intakeItem, 0)

	// Return early if no medications
	if len(d.medications) == 0 {
		log.Printf("No medications to process")
		return
	}

	log.Printf("Processing %d medications", len(d.medications))

	var allSlots []struct {
		med  models.Medication
		time time.Time
	}

	// Build all time slots
	for _, med := range d.medications {
		if med.ID == 0 {
			continue // Skip invalid medications
		}
		freq := models.ParseFrequency(med.Frequency)
		times := freq.GetDailyTimes()
		for _, t := range times {
			allSlots = append(allSlots, struct {
				med  models.Medication
				time time.Time
			}{med, t})
		}
	}

	// Sort slots by time
	sort.Slice(allSlots, func(i, j int) bool {
		return allSlots[i].time.Before(allSlots[j].time)
	})

	// Create a map of valid time slots for each medication
	validSlots := make(map[int64]map[string]bool)
	for _, slot := range allSlots {
		if _, exists := validSlots[slot.med.ID]; !exists {
			validSlots[slot.med.ID] = make(map[string]bool)
		}
		timeKey := fmt.Sprintf("%02d:%02d", slot.time.Hour(), slot.time.Minute())
		validSlots[slot.med.ID][timeKey] = true
	}

	// Filter out intakes that don't match current medication schedules
	var validIntakes []models.Intake
	for _, intake := range d.intakes {
		if slots, exists := validSlots[intake.MedicationID]; exists {
			timeKey := fmt.Sprintf("%02d:%02d", intake.ScheduledFor.Hour(), intake.ScheduledFor.Minute())
			if slots[timeKey] {
				validIntakes = append(validIntakes, intake)
			} else {
				log.Printf("Removing stale intake for medication %d at %s", intake.MedicationID, timeKey)
			}
		}
	}
	d.intakes = validIntakes

	// Create intake items for each valid slot
	for _, slot := range allSlots {
		item := intakeItem{
			medication: slot.med,
		}

		// Find matching intake from filtered list
		found := false
		for _, intake := range validIntakes {
			if intake.MedicationID == slot.med.ID {
				intakeTime := intake.ScheduledFor
				if intakeTime.Hour() == slot.time.Hour() && intakeTime.Minute() == slot.time.Minute() {
					item.intake = intake
					found = true
					break
				}
			}
		}

		// If no matching intake found, create a pending one
		if !found {
			item.intake = models.Intake{
				MedicationID: slot.med.ID,
				ScheduledFor: slot.time,
				Status:       models.IntakeStatusPending,
			}
		}

		d.cachedItems = append(d.cachedItems, item)
	}

	log.Printf("Updated cache with %d items", len(d.cachedItems))
}

// func (d *DailyIntake) getItemCount() int {
// 	d.mu.Lock()
// 	defer d.mu.Unlock()
// 	return len(d.cachedItems)
// }

// func (d *DailyIntake) getItemAtIndex(index int) intakeItem {
// 	d.mu.Lock()
// 	defer d.mu.Unlock()
// 	if index < 0 || index >= len(d.cachedItems) {
// 		return intakeItem{}
// 	}
// 	// Return a copy to prevent race conditions
// 	return d.cachedItems[index]
// }

func (d *DailyIntake) UpdateMedications(medications []models.Medication) {
	log.Printf("UpdateMedications: starting update with %d medications", len(medications))
	d.medications = medications
	d.updateCache()
	d.refresh()
}

func (d *DailyIntake) UpdateIntakes(intakes []models.Intake) {
	log.Printf("UpdateIntakes: starting update with %d intakes", len(intakes))
	d.intakes = intakes
	d.updateCache()
	d.refresh()
}

func (d *DailyIntake) updateDateLabel() {
	d.dateLabel.SetText("Today's Medications - " + time.Now().Format("Monday, January 2, 2006"))
}

func (d *DailyIntake) Container() fyne.CanvasObject {
	return d.container
}
