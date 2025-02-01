package views

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type Help struct {
	container *fyne.Container
}

func NewHelp() *Help {
	h := &Help{}

	// Create scrollable content
	content := container.NewVBox()
	scroll := container.NewVScroll(content)
	scroll.SetMinSize(fyne.NewSize(600, 400))

	// Add sections using rich text
	addSection := func(title string, text string) {
		// Add title
		titleText := widget.NewRichTextWithText(title + "\n")
		titleText.Wrapping = fyne.TextWrapWord
		titleText.Segments[0].(*widget.TextSegment).Style.TextStyle = fyne.TextStyle{Bold: true}
		content.Add(container.NewPadded(titleText))

		// Add content
		richText := widget.NewRichTextWithText(text + "\n")
		richText.Wrapping = fyne.TextWrapWord
		content.Add(richText)
		content.Add(widget.NewSeparator())
	}

	// Overview
	addSection("Overview",
		"MedTrack helps you manage your medications and track daily intake. "+
			"The app has two main sections: Medications and Daily Intake.")

	// Managing Medications
	addSection("Managing Medications",
		"The Medications tab shows your list of medications. From here you can:\n\n"+
			"• View all your medications\n"+
			"• Add new medications\n"+
			"• Edit existing medications\n"+
			"• Delete medications\n\n"+
			"Each medication card shows:\n"+
			"• Name and dosage\n"+
			"• Frequency and timing\n"+
			"• Food instructions (if any)\n"+
			"• Special instructions (if any)\n"+
			"• Additional notes\n\n"+
			"The interface is optimized for both desktop and mobile use, with a responsive design "+
			"that adapts to your screen size.\n")

	// Adding Medications
	addSection("Adding Medications",
		"To add a new medication:\n\n"+
			"1. Go to the Medications tab\n"+
			"2. Click 'Add New Medication'\n"+
			"3. The form will appear with 'Add Medication Details' title\n"+
			"4. Fill in the medication details:\n"+
			"   • Name: Enter the medication name\n"+
			"   • Dosage: Specify amount (e.g., 50mg)\n"+
			"   • Frequency: Choose how often to take\n"+
			"   • Times: Select specific times\n"+
			"   • Purpose: What the medication is for\n"+
			"   • Food Instructions: Choose before/after food\n"+
			"   • Special Instructions: Any specific instructions\n"+
			"   • Additional Notes: Generic names, side effects, etc.\n"+
			"   • Start Date: When to begin (YYYY-MM-DD)\n"+
			"   • End Date: Optional end date\n"+
			"5. Click Save (button will be enabled after making changes)\n\n"+
			"Note: You can click Cancel at any time to discard changes and return to the medication list.\n"+
			"The form is scrollable, so you can easily access all fields on any device.")

	// Editing Medications
	addSection("Editing Medications",
		"To edit an existing medication:\n\n"+
			"1. Go to the Medications tab\n"+
			"2. Find the medication in the list\n"+
			"3. Click the 'Edit' button\n"+
			"4. The form will appear with 'Update Medication Details' title\n"+
			"5. Update any details\n"+
			"6. Click Update when done (button will be enabled after making changes)\n\n"+
			"Note: You can:\n"+
			"• Click Cancel to discard changes and return to the list\n"+
			"• Click Clear to reset all fields\n\n"+
			"Changes will be reflected in your daily intake schedule.")

	// Deleting Medications
	addSection("Deleting Medications",
		"To delete a medication:\n\n"+
			"1. Go to the Medications tab\n"+
			"2. Find the medication in the list\n"+
			"3. Click the 'Delete' button\n"+
			"4. Confirm the deletion\n\n"+
			"Note: Deleting a medication will also remove all its intake records.")

	// Daily Intake
	addSection("Daily Intake",
		"The Daily Intake view shows your medications organized by time:\n\n"+
			"• Morning: Before noon\n"+
			"• Afternoon: 12 PM - 5 PM\n"+
			"• Night: After 5 PM\n\n"+
			"For each medication, you'll see:\n"+
			"• Name and dosage\n"+
			"• Food instructions (if any)\n"+
			"• Special instructions (if any)\n"+
			"• Checkbox to mark as taken")

	// Tracking Intake
	addSection("Tracking Intake",
		"To track your medication intake:\n\n"+
			"1. Go to the Daily Intake tab\n"+
			"2. Find the medication in the appropriate time section\n"+
			"3. Check the box when taken\n\n"+
			"The app will remember which medications you've taken each day.")

	// Tips
	addSection("Tips",
		"• Set up medications in advance\n"+
			"• Check the Daily Intake tab regularly\n"+
			"• Pay attention to food instructions\n"+
			"• Use notes for important reminders\n"+
			"• Keep medication information up to date\n"+
			"• Use in portrait orientation on mobile devices for best experience\n"+
			"• Scroll through the form to access all fields when adding/editing medications\n"+
			"• Look for the title change ('Add' vs 'Update') to know which mode you're in")

	h.container = container.NewBorder(
		widget.NewLabelWithStyle("Help & Documentation", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		nil, nil, nil,
		scroll,
	)

	return h
}

func (h *Help) Container() fyne.CanvasObject {
	return h.container
}
