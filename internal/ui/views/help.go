package views

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// ShowHelp opens a dedicated help window. On macOS this is triggered from the
// Help menu; it is a non-modal, closeable window following macOS conventions.
func ShowHelp(parent fyne.Window) {
	content := container.NewVBox()
	scroll := container.NewVScroll(content)

	addSection := func(title, text string) {
		titleText := widget.NewRichTextWithText(title + "\n")
		titleText.Wrapping = fyne.TextWrapWord
		titleText.Segments[0].(*widget.TextSegment).Style.TextStyle = fyne.TextStyle{Bold: true}
		// Spacer before title for breathing room; bare titleText avoids the
		// container.NewPadded box artefact visible in dark mode.
		content.Add(widget.NewLabel(""))
		content.Add(titleText)

		richText := widget.NewRichTextWithText(text + "\n")
		richText.Wrapping = fyne.TextWrapWord
		content.Add(richText)
		content.Add(widget.NewSeparator())
	}

	// Overview
	addSection("Overview",
		"MedTrack helps you manage your medications and track your daily intake. "+
			"The app has three main tabs:\n\n"+
			"• Daily Intake — see and record today's doses\n"+
			"• Medications — manage your medication list\n"+
			"• Log — review intake history over a date range\n\n"+
			"About and Preferences are available in the MedTrack menu in the menu bar.")

	// Daily Intake
	addSection("Daily Intake",
		"Your doses for today are organised into time-of-day sections:\n\n"+
			"• Morning (before 12:00 PM)\n"+
			"• Afternoon (12:00 PM – 5:00 PM)\n"+
			"• Evening (5:00 PM – 9:00 PM)\n"+
			"• Bedtime (10:00 PM onwards)\n\n"+
			"Each row shows the medication name, dosage, and a Taken checkbox. "+
			"If a medication has multiple doses scheduled for the same time slot, "+
			"each dose appears as a separate row with a dose badge (e.g. \"Dose 1 of 2\").\n\n"+
			"To record a dose as taken:\n"+
			"1. Check the Taken checkbox next to the dose\n"+
			"2. A \"When did you take it?\" dialog will appear, pre-filled with the scheduled time\n"+
			"3. Adjust the time if needed (HH:MM, 24-hour format)\n"+
			"4. Click Confirm — the row updates to show \"Taken at HH:MM\"\n\n"+
			"To undo a taken dose, uncheck the checkbox.\n\n"+
			"Doses that were not taken before midnight are automatically marked as Missed "+
			"the next time the app starts.")

	// PRN / As-Needed Medications
	addSection("PRN (As-Needed) Medications",
		"Medications set to frequency \"as needed\" appear in a separate "+
			"\"As Needed\" section at the bottom of the Daily Intake tab.\n\n"+
			"To record a PRN dose:\n"+
			"1. Find the medication in the As Needed section\n"+
			"2. Click \"Record Use\"\n"+
			"3. The use is logged immediately with the current time\n\n"+
			"PRN medications do not generate scheduled doses and never appear in the "+
			"timed sections (Morning, Afternoon, Evening, Bedtime).")

	// Managing Medications
	addSection("Managing Medications",
		"The Medications tab lists all your current medications. Each card shows:\n\n"+
			"• Name and dosage\n"+
			"• Frequency and scheduled times\n"+
			"• Purpose\n"+
			"• Food instructions (if set)\n"+
			"• Special instructions and notes\n\n"+
			"Each card also has an Active checkbox. Uncheck it to pause the medication — "+
			"it will be hidden from Daily Intake (pending intakes removed) but its history is kept. "+
			"Re-check it to resume scheduling.\n\n"+
			"Use the Edit and Delete buttons on each card to manage individual medications. "+
			"Deleting a medication also removes all of its intake records.")

	// Adding Medications
	addSection("Adding Medications",
		"1. Go to the Medications tab and click \"Add Medication\"\n"+
			"2. Fill in the form:\n"+
			"   • Name — medication name\n"+
			"   • Dosage — e.g. 50 mg\n"+
			"   • Frequency — how often to take it (e.g. \"2 times daily\", \"as needed\")\n"+
			"   • Times — morning, afternoon, evening, or bedtime slots\n"+
			"   • Purpose — what the medication is for\n"+
			"   • Food Instructions — before food, after food, etc.\n"+
			"   • Special Instructions — any clinical notes\n"+
			"   • Notes — generic names, side effects, reminders\n"+
			"   • Start Date — e.g. 2025-06-01\n"+
			"   • End Date — optional\n"+
			"3. Click Save\n\n"+
			"The new medication immediately appears in Daily Intake for today.")

	// Editing Medications
	addSection("Editing Medications",
		"1. Go to the Medications tab\n"+
			"2. Click Edit on the medication card\n"+
			"3. Update any fields\n"+
			"4. Click Update\n\n"+
			"Click Cancel at any time to discard changes. "+
			"If you change the frequency or time slots, Daily Intake is updated immediately: "+
			"pending intakes for removed slots are deleted and new pending intakes are created "+
			"for added slots. Taken, skipped, and missed entries are never affected.")

	// Deleting Medications
	addSection("Deleting Medications",
		"1. Go to the Medications tab\n"+
			"2. Click Delete on the medication card\n"+
			"3. Confirm the deletion in the dialog\n\n"+
			"Deleting a medication permanently removes it and all associated intake history.")

	// Medication Log
	addSection("Medication Log",
		"The Log tab shows your intake history across a date range.\n\n"+
			"• Use the From and To date fields (format: YYYY-MM-DD) to set the range\n"+
			"• Click Refresh to load results for that range\n\n"+
			"Each row shows the date, time slot, medication, dosage, and status:\n"+
			"• Taken — dose was recorded as taken (the scheduled time is shown on the left)\n"+
			"• Missed — scheduled dose was not taken before midnight\n"+
			"• Skipped — dose was explicitly skipped\n"+
			"• Pending — dose is scheduled for later today (current day only)")

	// Preferences
	addSection("Preferences",
		"Open Preferences from the MedTrack menu (shortcut: ⌘,).\n\n"+
			"Preferences contains data-management actions:\n\n"+
			"• Clear Intake History — removes all historical intake records but keeps "+
			"your medication list intact. Today's schedule is rebuilt automatically.\n"+
			"• Reset All Data — removes all medications and all intake records, "+
			"returning the app to a clean state.\n\n"+
			"Both actions require a two-step confirmation: review the summary, then "+
			"type DELETE to confirm. This prevents accidental data loss.")

	// Tips
	addSection("Tips",
		"• Set up all your medications before your first daily check-in\n"+
			"• The scheduled time shown in the \"When did you take it?\" dialog is your target — "+
			"adjust it only if you actually took the dose at a different time\n"+
			"• PRN medications are for occasional use; frequent use may indicate a need to "+
			"review your regimen with your doctor\n"+
			"• Use the Log tab to spot patterns (consistently missed doses, timing drift)\n"+
			"• Keep medication notes up to date with generic names and known interactions")

	win := fyne.CurrentApp().NewWindow("MedTrack Help")
	win.SetContent(container.NewBorder(
		container.NewPadded(widget.NewLabelWithStyle("Help & Documentation", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})),
		nil, nil, nil,
		scroll,
	))
	win.Resize(fyne.NewSize(620, 520))
	win.Show()
}
