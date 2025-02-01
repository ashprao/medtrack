package views

import (
	"fmt"

	"medtrack/internal/models"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

type MedicationList struct {
	medications   []models.Medication
	container     *fyne.Container
	mainContainer *fyne.Container // Container to hold both list and form
	listView      *fyne.Container // Container for the list view
	formView      *MedicationForm // Form view
	onSave        func(med *models.Medication) error
	onDelete      func(med *models.Medication)
	window        fyne.Window
	addButton     *widget.Button
	editButtons   []*widget.Button
	deleteButtons []*widget.Button
}

func NewMedicationList(onSave func(med *models.Medication) error, onDelete func(med *models.Medication), window fyne.Window) *MedicationList {
	// Create list view with add button
	list := &MedicationList{
		medications:   make([]models.Medication, 0),
		onSave:        onSave,
		onDelete:      onDelete,
		window:        window,
		editButtons:   make([]*widget.Button, 0),
		deleteButtons: make([]*widget.Button, 0),
	}

	list.addButton = widget.NewButton("Add New Medication", func() {
		// Create form view when needed
		if list.formView == nil {
			list.formView = NewMedicationForm(list.handleFormSubmit)
			list.formView.SetOnCancel(list.handleFormCancel)
		}
		list.showFormView(nil) // nil indicates new medication
	})

	list.listView = container.NewVBox(
		container.NewHBox(
			widget.NewLabel("Medications"),
			list.addButton,
		),
	)

	// Create main container with list view initially visible
	list.mainContainer = container.NewMax(list.listView)
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

func (m *MedicationList) EnableAllButtons() {
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
		m.listView.Add(widget.NewLabel("No medications added yet"))
		return
	}

	for _, med := range m.medications {
		med := med // Create new variable for closure
		// Create content for card
		content := container.NewVBox(
			widget.NewLabel("Frequency: "+med.Frequency),
			widget.NewLabel("Purpose: "+med.Purpose),
		)

		// Add food instructions if specified
		if med.FoodInstruction != "" && med.FoodInstruction != "No Food Restriction" {
			foodLabel := widget.NewLabelWithStyle(
				"Take "+med.FoodInstruction,
				fyne.TextAlignLeading,
				fyne.TextStyle{Bold: true},
			)
			content.Add(foodLabel)
		}

		// Add special instructions if any
		if med.Instructions != "" {
			content.Add(widget.NewLabel("Instructions: " + med.Instructions))
		}

		// Add notes if any
		if med.Notes != "" {
			content.Add(widget.NewLabelWithStyle(
				"Notes: "+med.Notes,
				fyne.TextAlignLeading,
				fyne.TextStyle{Italic: true},
			))
		}

		// Create button container
		buttonContainer := container.NewHBox()

		// Add edit button
		editBtn := widget.NewButton("Edit", nil) // Create button first
		editBtn.OnTapped = func() {
			// Create form view when needed
			if m.formView == nil {
				m.formView = NewMedicationForm(m.handleFormSubmit)
				m.formView.SetOnCancel(m.handleFormCancel)
			}
			m.showFormView(&med)
		}
		buttonContainer.Add(editBtn)
		m.editButtons = append(m.editButtons, editBtn)

		// Add delete button
		deleteBtn := widget.NewButton("Delete", func() {
			// Show confirmation dialog
			dialog := dialog.NewConfirm(
				"Delete Medication",
				fmt.Sprintf("Are you sure you want to delete %s? This will also remove all intake records.", med.Name),
				func(confirmed bool) {
					if confirmed && m.onDelete != nil {
						m.onDelete(&med)
					}
				},
				fyne.CurrentApp().Driver().AllWindows()[0],
			)
			dialog.Show()
		})
		deleteBtn.Importance = widget.DangerImportance
		buttonContainer.Add(deleteBtn)
		m.deleteButtons = append(m.deleteButtons, deleteBtn)

		content.Add(buttonContainer)

		// Create card with all details
		card := widget.NewCard(
			med.Name,
			med.Dosage,
			content,
		)
		m.listView.Add(card)
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
	scroll.SetMinSize(fyne.NewSize(300, 400))
	m.mainContainer.Add(scroll)
	m.disableAllButtons()
}

// Show list view and hide form view
func (m *MedicationList) showListView() {
	m.mainContainer.RemoveAll()
	m.mainContainer.Add(m.listView)
	m.EnableAllButtons()
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
		m.EnableAllButtons()
	}
}

// Handle form cancellation
func (m *MedicationList) handleFormCancel() {
	m.showListView()
}
