package views

import (
	"fmt"
	"image/color"
	"log"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/ashprao/medtrack/internal/models"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// DailyIntake struct holds the data and UI components for displaying a daily intake of medications
// This is the Publisher or Subject in the Observer pattern.
type DailyIntake struct {
	container   *fyne.Container
	medications []models.Medication
	// prnMedications holds medications with TimesPerDay == 0 ("as needed")
	prnMedications []models.Medication
	dateLabel      *widget.Label
	content        *fyne.Container
	emptyState     *fyne.Container
	scrollArea     *container.Scroll
	// onTake is a function called when a medication is taken or skipped. This is the Observer
	// in the Observer pattern
	onTake func(med *models.Medication, scheduledTime time.Time, taken bool)
	// onPRNUse is called when the user taps "Record Use" for a PRN medication
	onPRNUse func(medID int64)
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
	doseNum    int // 1-based position among doses of this medication today
	doseTotal  int // total doses of this medication today (1 means no label needed)
}

func NewDailyIntake(onTake func(med *models.Medication, scheduledTime time.Time, taken bool), onPRNUse func(medID int64)) *DailyIntake {
	d := &DailyIntake{
		dateLabel:      widget.NewLabel(""),
		onTake:         onTake,
		onPRNUse:       onPRNUse,
		medications:    make([]models.Medication, 0),
		prnMedications: make([]models.Medication, 0),
		intakes:        make([]models.Intake, 0),
		cachedItems:    make([]intakeItem, 0),
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
	d.scrollArea = container.NewVScroll(d.content)

	// Empty state fills the center area when there are no items
	emptyLabel := widget.NewLabelWithStyle(
		"No medications scheduled for today",
		fyne.TextAlignCenter,
		fyne.TextStyle{},
	)
	emptyLabel.Importance = widget.MediumImportance
	d.emptyState = container.NewCenter(emptyLabel)

	// Stack holds both; visibility is toggled in refresh()
	center := container.NewStack(d.scrollArea, d.emptyState)

	// Create main container
	d.container = container.NewBorder(
		header, // Top
		nil,    // Bottom
		nil,    // Left
		nil,    // Right
		center, // Center
	)

	return d
}

func (d *DailyIntake) refresh() {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.updateDateLabel()

	// Clear existing content
	d.content.RemoveAll()

	// Show message if no items
	if len(d.cachedItems) == 0 {
		d.scrollArea.Hide()
		d.emptyState.Show()
		return
	}
	d.emptyState.Hide()
	d.scrollArea.Show()

	// Group items by time of day
	morning := make([]intakeItem, 0)
	afternoon := make([]intakeItem, 0)
	evening := make([]intakeItem, 0)
	bedtime := make([]intakeItem, 0)

	for _, item := range d.cachedItems {
		hour := item.intake.ScheduledFor.Hour()
		switch {
		case hour < 12: // Morning (before noon)
			morning = append(morning, item)
		case hour < 17: // Afternoon (12:00–16:59)
			afternoon = append(afternoon, item)
		case hour < 21: // Evening (17:00–20:59)
			evening = append(evening, item)
		default: // Bedtime (21:00+)
			bedtime = append(bedtime, item)
		}
	}

	// Add a helper function for creating widget.Check components
	createCheckBox := func(initialState bool, onChange func(bool)) *widget.Check {
		check := widget.NewCheck("", nil)
		check.SetChecked(initialState)
		check.OnChanged = onChange
		return check
	}

	// Helper function to create a flat row for an intake item.
	// Uses a separator instead of a card border to save vertical space.
	createRow := func(item intakeItem) fyne.CanvasObject {
		check := createCheckBox(item.intake.Status == models.IntakeStatusTaken, func(taken bool) {
			if d.onTake != nil {
				med := item.medication
				log.Printf("Marking medication %s for time %v as %v", med.Name, item.intake.ScheduledFor, taken)
				d.onTake(&med, item.intake.ScheduledFor, taken)
			}
		})

		// Checkbox + name (bold) + optional dose badge + dosage (muted) — all on one line
		nameLabel := widget.NewLabelWithStyle(item.medication.Name, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
		dosageLabel := widget.NewLabel(item.medication.Dosage)
		dosageLabel.Importance = widget.MediumImportance
		titleRow := container.NewHBox(check, nameLabel)
		if item.doseTotal > 1 {
			doseLabel := widget.NewLabel(fmt.Sprintf("Dose %d of %d", item.doseNum, item.doseTotal))
			doseLabel.Importance = widget.MediumImportance
			titleRow.Add(doseLabel)
		}
		titleRow.Add(layout.NewSpacer())
		titleRow.Add(dosageLabel)

		// Food instruction as a muted second line (not bold — secondary info)
		rows := container.NewVBox(titleRow)
		if item.medication.FoodInstruction != "" && item.medication.FoodInstruction != "No Food Restriction" {
			foodLabel := widget.NewLabel(fmt.Sprintf("Take %s", strings.ToLower(item.medication.FoodInstruction)))
			foodLabel.Importance = widget.MediumImportance
			rows.Add(foodLabel)
		}

		// Optional special instructions
		if item.medication.Instructions != "" {
			rows.Add(widget.NewLabel(item.medication.Instructions))
		}

		// Separator acts as the row divider (replaces card border)
		rows.Add(widget.NewSeparator())

		return container.NewPadded(rows)
	}

	addTimeSection := func(title string, items []intakeItem, bgColor color.Color) {
		if len(items) > 0 {
			header := widget.NewLabelWithStyle(title, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
			headerContent := container.NewHBox(layout.NewSpacer(), header, layout.NewSpacer())
			bg := canvas.NewRectangle(bgColor)
			bg.CornerRadius = 6
			d.content.Add(container.NewStack(bg, container.NewPadded(headerContent)))
			d.content.Add(widget.NewSeparator())

			for _, item := range items {
				d.content.Add(createRow(item))
			}
		}
	}

	// Apple-inspired time-of-day tints — low alpha (~15%) works in both light and dark mode
	addTimeSection("Morning", morning, color.NRGBA{R: 255, G: 190, B: 50, A: 38})
	addTimeSection("Afternoon", afternoon, color.NRGBA{R: 30, G: 140, B: 255, A: 38})
	addTimeSection("Evening", evening, color.NRGBA{R: 255, G: 100, B: 40, A: 38})
	addTimeSection("Bedtime", bedtime, color.NRGBA{R: 90, G: 70, B: 210, A: 38})

	// As-needed (PRN) section
	if len(d.prnMedications) > 0 {
		header := widget.NewLabelWithStyle("As Needed", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
		headerContainer := container.NewHBox(layout.NewSpacer(), header, layout.NewSpacer())
		d.content.Add(headerContainer)
		d.content.Add(widget.NewSeparator())

		for _, med := range d.prnMedications {
			m := med // capture
			nameLabel := widget.NewLabelWithStyle(m.Name, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
			dosageLabel := widget.NewLabel(m.Dosage)
			dosageLabel.Importance = widget.MediumImportance
			titleRow := container.NewHBox(nameLabel, layout.NewSpacer(), dosageLabel)

			recordBtn := widget.NewButton("Record Use", func() {
				if d.onPRNUse != nil {
					d.onPRNUse(m.ID)
				}
			})

			btnRow := container.NewHBox(recordBtn)

			rows := container.NewVBox(titleRow, btnRow)
			if m.FoodInstruction != "" && m.FoodInstruction != "No Food Restriction" {
				foodLabel := widget.NewLabel(fmt.Sprintf("Take %s", strings.ToLower(m.FoodInstruction)))
				foodLabel.Importance = widget.MediumImportance
				rows.Add(foodLabel)
			}
			if m.Instructions != "" {
				rows.Add(widget.NewLabel(m.Instructions))
			}
			rows.Add(widget.NewSeparator())
			d.content.Add(container.NewPadded(rows))
		}
	}
}

// updateCache rebuilds the cached items list
func (d *DailyIntake) updateCache() {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Clear existing cache
	d.cachedItems = make([]intakeItem, 0)
	d.prnMedications = make([]models.Medication, 0)

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

	// Build all time slots, separating PRN medications
	for _, med := range d.medications {
		if med.ID == 0 {
			continue // Skip invalid medications
		}
		freq := models.ParseFrequency(med.Frequency)
		if freq.TimesPerDay == 0 {
			d.prnMedications = append(d.prnMedications, med)
			continue
		}
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

	// Compute dose numbers: count totals per medication, then assign doseNum.
	totals := make(map[int64]int)
	for _, it := range d.cachedItems {
		totals[it.medication.ID]++
	}
	counters := make(map[int64]int)
	for i := range d.cachedItems {
		medID := d.cachedItems[i].medication.ID
		counters[medID]++
		d.cachedItems[i].doseNum = counters[medID]
		d.cachedItems[i].doseTotal = totals[medID]
	}
}

func (d *DailyIntake) UpdateMedications(medications []models.Medication) {
	d.medications = medications
	d.updateCache()
	d.refresh()
}

func (d *DailyIntake) UpdateIntakes(intakes []models.Intake) {
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
