package views

import (
	"fmt"

	"github.com/ashprao/medtrack/internal/models"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

type MedicationList struct {
	medications    []models.Medication
	container      *fyne.Container
	mainContainer  *fyne.Container // Container to hold both list and form
	listView       *fyne.Container // Container for the list view
	formView       *MedicationForm // Form view
	onSave         func(med *models.Medication) error
	onDelete       func(med *models.Medication)
	onToggleActive func(med *models.Medication, active bool)
	window         fyne.Window
	addButton      *widget.Button
	editButtons    []*widget.Button
	deleteButtons  []*widget.Button
}

func NewMedicationList(onSave func(med *models.Medication) error, onDelete func(med *models.Medication), onToggleActive func(med *models.Medication, active bool), window fyne.Window) *MedicationList {
	// Create list view with add button
	list := &MedicationList{
		medications:    make([]models.Medication, 0),
		onSave:         onSave,
		onDelete:       onDelete,
		onToggleActive: onToggleActive,
		window:         window,
		editButtons:    make([]*widget.Button, 0),
		deleteButtons:  make([]*widget.Button, 0),
	}

	list.addButton = widget.NewButton("Add Medication", func() {
		// Create form view when needed
		if list.formView == nil {
			list.formView = NewMedicationForm(list.handleFormSubmit, list.window)
			list.formView.SetOnCancel(list.handleFormCancel)
		}
		list.showFormView(nil) // nil indicates new medication
	})

	// Create header
	header := container.NewBorder(nil, nil, nil, list.addButton,
		widget.NewLabelWithStyle("My Medications", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
	)

	// Create list view with header and scrollable content
	list.listView = container.NewVBox(header)
	scrollContainer := container.NewVScroll(list.listView)

	// Create main container with scrollable list view initially visible
	list.mainContainer = container.NewMax(scrollContainer)
	list.container = list.mainContainer
	return list
}

func (m *MedicationList) UpdateMedications(medications []models.Medication) {
	m.medications = medications
	m.refresh()
}

func (m *MedicationList) disableAllButtons() {
	m.addButton.Disable()
	for _, btn := range m.editButtons {
		btn.Disable()
	}
	for _, btn := range m.deleteButtons {
		btn.Disable()
	}
}

func (m *MedicationList) enableAllButtons() {
	m.addButton.Enable()
	for _, btn := range m.editButtons {
		btn.Enable()
	}
	for _, btn := range m.deleteButtons {
		btn.Enable()
	}
}

func (m *MedicationList) refresh() {
	// Clear button slices
	m.editButtons = make([]*widget.Button, 0)
	m.deleteButtons = make([]*widget.Button, 0)

	// Keep the header (first child)
	header := m.listView.Objects[0]
	m.listView.RemoveAll()
	m.listView.Add(header)

	if len(m.medications) == 0 {
		m.listView.Add(container.NewCenter(container.NewPadded(
			widget.NewLabel("No medications yet. Tap Add Medication to get started."),
		)))
		return
	}

	for _, med := range m.medications {
		med := med // Create new variable for closure

		// Active / Inactive toggle in the card header
		activeCheck := widget.NewCheck("Active", nil)
		activeCheck.SetChecked(med.Active)
		activeCheck.OnChanged = func(on bool) {
			if m.onToggleActive != nil {
				m.onToggleActive(&med, on)
			}
		}

		// Create button row (at top of card content for quick access)
		editBtn := widget.NewButton("Edit", nil)
		editBtn.OnTapped = func() {
			// Create form view when needed
			if m.formView == nil {
				m.formView = NewMedicationForm(m.handleFormSubmit, m.window)
				m.formView.SetOnCancel(m.handleFormCancel)
			}
			m.showFormView(&med)
		}
		m.editButtons = append(m.editButtons, editBtn)

		deleteBtn := widget.NewButton("Delete", func() {
			dlg := dialog.NewConfirm(
				"Delete Medication",
				fmt.Sprintf("Are you sure you want to delete %s? This will also remove all intake records.", med.Name),
				func(confirmed bool) {
					if confirmed && m.onDelete != nil {
						m.onDelete(&med)
					}
				},
				m.window,
			)
			dlg.Show()
		})
		deleteBtn.Importance = widget.DangerImportance
		m.deleteButtons = append(m.deleteButtons, deleteBtn)

		buttonContainer := container.NewHBox(editBtn, layout.NewSpacer(), deleteBtn)

		// Build structured key/value info using widget.NewForm
		infoForm := widget.NewForm(
			widget.NewFormItem("Frequency", widget.NewLabel(med.Frequency)),
			widget.NewFormItem("Purpose", widget.NewLabel(med.Purpose)),
		)

		content := container.NewVBox(infoForm)

		// Add food instructions if specified
		if med.FoodInstruction != "" && med.FoodInstruction != "No Food Restriction" {
			content.Add(widget.NewLabelWithStyle(
				"Take "+med.FoodInstruction,
				fyne.TextAlignLeading,
				fyne.TextStyle{Bold: true},
			))
		}

		// Add special instructions if any
		if med.Instructions != "" {
			content.Add(widget.NewForm(
				widget.NewFormItem("Instructions", widget.NewLabel(med.Instructions)),
			))
		}

		// Add notes if any
		if med.Notes != "" {
			content.Add(widget.NewForm(
				widget.NewFormItem("Notes", widget.NewLabelWithStyle(med.Notes, fyne.TextAlignLeading, fyne.TextStyle{Italic: true})),
			))
		}

		// Edit/Delete at the bottom — destructive actions trail content per Apple HIG
		content.Add(widget.NewSeparator())
		content.Add(buttonContainer)

		// Create card with all details; show inactive meds in a muted state
		var cardTitle string
		if med.Active {
			cardTitle = med.Name
		} else {
			cardTitle = med.Name + "  [Paused]"
		}
		card := widget.NewCard(
			cardTitle,
			med.Dosage,
			content,
		)
		// Place the active checkbox trailing in the card's top-right area
		cardWithToggle := container.NewBorder(nil, nil, nil, activeCheck, card)
		m.listView.Add(cardWithToggle)
	}
}

func (m *MedicationList) Container() fyne.CanvasObject {
	return m.container
}

// Show form view and hide list view
func (m *MedicationList) showFormView(med *models.Medication) {
	m.mainContainer.RemoveAll()
	if med != nil {
		m.formView.LoadMedication(med)
	} else {
		m.formView.Clear()
	}
	scroll := container.NewVScroll(m.formView.Container())
	m.mainContainer.Add(scroll)
	m.disableAllButtons()
}

// Show list view and hide form view
func (m *MedicationList) showListView() {
	m.mainContainer.RemoveAll()
	scrollContainer := container.NewVScroll(m.listView)
	m.mainContainer.Add(scrollContainer)
	m.enableAllButtons()
}

// Handle form submission
func (m *MedicationList) handleFormSubmit(med *models.Medication) {
	// Save medication through the onSave handler
	if m.onSave != nil {
		// Disable save button while saving
		if m.formView != nil {
			m.formView.saveButton.Disable()
		}

		// Call onSave which will handle the database save
		if err := m.onSave(med); err != nil {
			dialog.ShowError(err, m.window)
			if m.formView != nil {
				m.formView.saveButton.Enable()
			}
			return
		}

		// Return to list view after successful save
		m.showListView()
		m.enableAllButtons()
	}
}

// Handle form cancellation
func (m *MedicationList) handleFormCancel() {
	m.showListView()
}
