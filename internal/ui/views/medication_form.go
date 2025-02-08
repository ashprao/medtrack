package views

import (
	"fmt"
	"strings"
	"time"

	"github.com/ashprao/medtrack/internal/models"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type MedicationForm struct {
	container     *fyne.Container
	titleLabel    *widget.Label
	nameEntry     *widget.Entry
	dosageEntry   *widget.Entry
	freqSelect    *widget.Select   // How many times per day
	timeSelects   []*widget.Select // Time slots for each frequency
	timeContainer *fyne.Container  // Container for time selects
	purposeEntry  *widget.Entry
	foodSelect    *widget.Select // Before/After food
	instrEntry    *widget.Entry
	notesEntry    *widget.Entry // Additional info like generic name, side effects
	startDate     *widget.Entry
	endDate       *widget.Entry
	saveButton    *widget.Button
	cancelButton  *widget.Button
	clearButton   *widget.Button
	onSubmit      func(med *models.Medication)
	onCancel      func()
	currentMed    *models.Medication // Track current medication being edited
	hasChanges    bool               // Track if form has unsaved changes
}

// Predefined options
var (
	FrequencyOptions = []string{
		"Once Daily",
		"Twice Daily",
		"Three Times Daily",
	}

	TimeOptions = []string{
		"Morning (9:00 AM)",
		"Afternoon (3:00 PM)",
		"Night (9:00 PM)",
	}

	FoodOptions = []string{
		"Before Food",
		"After Food",
		"With Food",
		"No Food Restriction",
	}

	TimeMap = map[string]string{
		"Morning (9:00 AM)":   "09:00",
		"Afternoon (3:00 PM)": "15:00",
		"Night (9:00 PM)":     "21:00",
	}
)

func NewMedicationForm(onSubmit func(med *models.Medication)) *MedicationForm {
	form := &MedicationForm{
		nameEntry:     widget.NewEntry(),
		dosageEntry:   widget.NewEntry(),
		freqSelect:    widget.NewSelect(FrequencyOptions, nil),
		timeSelects:   make([]*widget.Select, 0),
		timeContainer: container.NewVBox(),
		purposeEntry:  widget.NewEntry(),
		foodSelect:    widget.NewSelect(FoodOptions, nil),
		instrEntry:    widget.NewMultiLineEntry(),
		notesEntry:    widget.NewMultiLineEntry(),
		startDate:     widget.NewEntry(),
		endDate:       widget.NewEntry(),
		onSubmit:      onSubmit,
		hasChanges:    false,
	}

	// Create title label
	form.titleLabel = widget.NewLabelWithStyle("Add Medication Details", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	// Create buttons
	form.saveButton = widget.NewButton("Save", form.submit)
	form.saveButton.Importance = widget.HighImportance
	form.saveButton.Disable() // Initially disabled until changes

	form.cancelButton = widget.NewButton("Cancel", form.cancel)
	form.clearButton = widget.NewButton("Clear", form.Clear)

	// Set placeholders and defaults
	form.nameEntry.SetPlaceHolder("Medication Name")
	form.dosageEntry.SetPlaceHolder("Dosage (e.g., 50mg)")
	form.freqSelect.SetSelected(FrequencyOptions[0]) // Default to once daily
	form.purposeEntry.SetPlaceHolder("Purpose/Treatment")
	form.foodSelect.SetSelected(FoodOptions[3]) // Default to "No Food Restriction"

	// Initialize time selects for default frequency
	form.updateTimeSelects(FrequencyOptions[0])
	form.instrEntry.SetPlaceHolder("Special Instructions")
	form.notesEntry.SetPlaceHolder("Additional Notes (e.g., generic name, side effects)")
	form.startDate.SetPlaceHolder("Start Date (YYYY-MM-DD)")
	form.endDate.SetPlaceHolder("End Date (YYYY-MM-DD, optional)")

	// Setup change handlers for all inputs
	form.nameEntry.OnChanged = func(string) { form.markChanged() }
	form.dosageEntry.OnChanged = func(string) { form.markChanged() }
	form.freqSelect.OnChanged = func(value string) {
		form.updateTimeSelects(value)
		form.markChanged()
	}
	form.purposeEntry.OnChanged = func(string) { form.markChanged() }
	form.foodSelect.OnChanged = func(string) { form.markChanged() }
	form.instrEntry.OnChanged = func(string) { form.markChanged() }
	form.notesEntry.OnChanged = func(string) { form.markChanged() }
	form.startDate.OnChanged = func(string) { form.markChanged() }
	form.endDate.OnChanged = func(string) { form.markChanged() }

	// Create button container with improved layout for mobile
	buttonContainer := container.NewHBox(
		form.cancelButton,
		form.clearButton,
		form.saveButton,
	)

	// Create form with items
	formWidget := widget.NewForm(
		widget.NewFormItem("Name", form.nameEntry),
		widget.NewFormItem("Dosage", form.dosageEntry),
		widget.NewFormItem("Frequency", form.freqSelect),
		widget.NewFormItem("Times", form.timeContainer),
		widget.NewFormItem("Purpose", form.purposeEntry),
		widget.NewFormItem("Food Instructions", form.foodSelect),
		widget.NewFormItem("Special Instructions", form.instrEntry),
		widget.NewFormItem("Additional Notes", form.notesEntry),
		widget.NewFormItem("Start Date", form.startDate),
		widget.NewFormItem("End Date", form.endDate),
	)

	// Create form container with title and buttons
	form.container = container.NewVBox(
		form.titleLabel,
		formWidget,
		buttonContainer,
	)

	return form
}

func (f *MedicationForm) getFrequencyString() string {
	var count string
	switch f.freqSelect.Selected {
	case "Once Daily":
		count = "once daily"
	case "Twice Daily":
		count = "twice daily"
	case "Three Times Daily":
		count = "three times daily"
	default:
		count = "once daily"
	}

	// Get selected times
	var times []string
	for _, timeSelect := range f.timeSelects {
		if timeSelect.Selected != "" {
			times = append(times, TimeMap[timeSelect.Selected])
		}
	}

	// Return combined frequency and times
	return fmt.Sprintf("%s at %s", count, strings.Join(times, ", "))
}

func (f *MedicationForm) parseFrequency(freq string) (string, []string) {
	parts := strings.Split(freq, " at ")
	if len(parts) != 2 {
		return "Once Daily", []string{TimeOptions[0]}
	}

	count := parts[0]
	times := strings.Split(parts[1], ", ")

	// Convert times back to display format
	var displayTimes []string
	for _, t := range times {
		for display, value := range TimeMap {
			if value == t {
				displayTimes = append(displayTimes, display)
				break
			}
		}
	}

	// Convert count to display format
	var displayCount string
	switch strings.TrimSpace(count) {
	case "once daily":
		displayCount = "Once Daily"
	case "twice daily":
		displayCount = "Twice Daily"
	case "three times daily":
		displayCount = "Three Times Daily"
	default:
		displayCount = "Once Daily"
	}

	return displayCount, displayTimes
}

func (f *MedicationForm) markChanged() {
	f.hasChanges = true
	f.saveButton.Enable()
}

func (f *MedicationForm) submit() {
	// Validate required fields
	if f.nameEntry.Text == "" || f.dosageEntry.Text == "" || f.startDate.Text == "" {
		f.saveButton.Enable()
		return
	}

	// Parse dates
	startDate, err := time.Parse("2006-01-02", f.startDate.Text)
	if err != nil {
		f.saveButton.Enable()
		return
	}

	var endDate *time.Time
	if f.endDate.Text != "" {
		parsed, err := time.Parse("2006-01-02", f.endDate.Text)
		if err != nil {
			f.saveButton.Enable()
			return
		}
		endDate = &parsed
	}

	// Create medication with all fields
	med := &models.Medication{
		Name:            f.nameEntry.Text,
		Dosage:          f.dosageEntry.Text,
		Frequency:       f.getFrequencyString(),
		Purpose:         f.purposeEntry.Text,
		FoodInstruction: f.foodSelect.Selected,
		Instructions:    f.instrEntry.Text,
		Notes:           f.notesEntry.Text,
		StartDate:       startDate,
		EndDate:         endDate,
	}

	if f.currentMed != nil {
		med.ID = f.currentMed.ID
		med.CreatedAt = f.currentMed.CreatedAt
	}

	// Call onSubmit with the new medication
	if f.onSubmit != nil {
		f.onSubmit(med)
	}
}

func (f *MedicationForm) SetOnSubmit(handler func(med *models.Medication)) {
	f.onSubmit = handler
}

func (f *MedicationForm) SetOnCancel(handler func()) {
	f.onCancel = handler
}

func (f *MedicationForm) cancel() {
	if f.currentMed != nil {
		// Reload original data
		f.LoadMedication(f.currentMed)
	} else {
		// Clear form for new medication
		f.Clear()
	}
	f.hasChanges = false
	f.saveButton.Disable()

	if f.onCancel != nil {
		f.onCancel()
	}
}

func (f *MedicationForm) Container() fyne.CanvasObject {
	return f.container
}

func (f *MedicationForm) updateTimeSelects(freq string) {
	f.timeContainer.RemoveAll()
	f.timeSelects = make([]*widget.Select, 0)

	var count int
	switch freq {
	case "Once Daily":
		count = 1
	case "Twice Daily":
		count = 2
	case "Three Times Daily":
		count = 3
	default:
		count = 1
	}

	for i := 0; i < count; i++ {
		timeSelect := widget.NewSelect(TimeOptions, nil)
		timeSelect.SetSelected(TimeOptions[0]) // Default to morning
		timeSelect.OnChanged = func(string) { f.markChanged() }
		f.timeSelects = append(f.timeSelects, timeSelect)
		f.timeContainer.Add(timeSelect)
	}
}

func (f *MedicationForm) LoadMedication(med *models.Medication) {
	f.currentMed = med
	f.nameEntry.SetText(med.Name)
	f.dosageEntry.SetText(med.Dosage)

	// Update title for editing
	f.titleLabel.SetText("Update Medication Details")

	// Parse frequency and set selections
	freq, times := f.parseFrequency(med.Frequency)
	f.freqSelect.SetSelected(freq)
	f.updateTimeSelects(freq)

	// Set time selections
	for i, timeSelect := range f.timeSelects {
		if i < len(times) {
			timeSelect.SetSelected(times[i])
		}
	}

	f.purposeEntry.SetText(med.Purpose)
	f.foodSelect.SetSelected(med.FoodInstruction)
	f.instrEntry.SetText(med.Instructions)
	f.notesEntry.SetText(med.Notes)
	f.startDate.SetText(med.StartDate.Format("2006-01-02"))
	if med.EndDate != nil {
		f.endDate.SetText(med.EndDate.Format("2006-01-02"))
	}

	// Update button states
	f.saveButton.SetText("Update")
	f.saveButton.Disable() // Disabled until changes
	f.hasChanges = false
}

func (f *MedicationForm) Clear() {
	f.currentMed = nil
	f.nameEntry.SetText("")
	f.dosageEntry.SetText("")
	f.freqSelect.SetSelected(FrequencyOptions[0])
	f.updateTimeSelects(FrequencyOptions[0])
	f.purposeEntry.SetText("")
	f.foodSelect.SetSelected(FoodOptions[3]) // Default to "No Food Restriction"
	f.instrEntry.SetText("")
	f.notesEntry.SetText("")
	f.startDate.SetText("")
	f.endDate.SetText("")

	// Reset title for adding
	f.titleLabel.SetText("Add Medication Details")

	// Reset button states
	f.saveButton.SetText("Save")
	f.saveButton.Disable()
	f.hasChanges = false
}

func (f *MedicationForm) IsEditing() bool {
	return f.currentMed != nil
}
