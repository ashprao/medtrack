package views

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// ShowPreferences opens the Preferences window.
// On macOS, Fyne moves a menu item labelled "Preferences…" into the application
// menu automatically. getCounts, onClearHistory, and onResetAll are injected
// from main.go so the view layer has no direct DB dependency.
func ShowPreferences(
	parent fyne.Window,
	getCounts func() (medCount int, intakeCount int, err error),
	onClearHistory func() error,
	onResetAll func() error,
) {
	win := fyne.CurrentApp().NewWindow("MedTrack Preferences")
	win.Resize(fyne.NewSize(460, 320))
	win.SetFixedSize(true)

	medCount, intakeCount, countsErr := getCounts()

	// ── Data Management section ──────────────────────────────────────────────

	sectionLabel := widget.NewLabelWithStyle("Data Management", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	sectionLabel.Importance = widget.MediumImportance

	// Clear History row
	clearTitle := widget.NewLabel("Clear Intake History")
	clearTitle.TextStyle = fyne.TextStyle{Bold: true}

	clearDesc := widget.NewLabel("Remove all past intake records. Medications and their schedules are not affected.")
	clearDesc.Wrapping = fyne.TextWrapWord
	clearDesc.Importance = widget.MediumImportance

	clearBtn := widget.NewButton("Clear History…", func() {
		showClearHistoryFlow(win, getCounts, onClearHistory)
	})
	clearBtn.Importance = widget.WarningImportance
	if countsErr != nil || intakeCount == 0 {
		clearBtn.Disable()
	}

	clearRow := container.NewVBox(
		container.NewBorder(nil, nil, nil, clearBtn, clearTitle),
		clearDesc,
	)

	// Erase All Data row
	eraseTitle := widget.NewLabel("Erase All Data")
	eraseTitle.TextStyle = fyne.TextStyle{Bold: true}

	eraseDesc := widget.NewLabel("Permanently remove all medications and their complete intake history.")
	eraseDesc.Wrapping = fyne.TextWrapWord
	eraseDesc.Importance = widget.MediumImportance

	eraseBtn := widget.NewButton("Erase All Data…", func() {
		showResetAllFlow(win, getCounts, onResetAll)
	})
	eraseBtn.Importance = widget.DangerImportance
	if countsErr != nil || (medCount == 0 && intakeCount == 0) {
		eraseBtn.Disable()
	}

	eraseRow := container.NewVBox(
		container.NewBorder(nil, nil, nil, eraseBtn, eraseTitle),
		eraseDesc,
	)

	content := container.NewVBox(
		sectionLabel,
		widget.NewSeparator(),
		clearRow,
		widget.NewSeparator(),
		eraseRow,
	)

	doneBtn := widget.NewButton("Done", func() { win.Close() })

	win.SetContent(container.NewPadded(
		container.NewBorder(nil, doneBtn, nil, nil, content),
	))
	win.Show()
}

// showClearHistoryFlow shows a confirmation sheet before clearing intake history.
func showClearHistoryFlow(
	win fyne.Window,
	getCounts func() (int, int, error),
	onClearHistory func() error,
) {
	_, intakeCount, err := getCounts()
	if err != nil {
		dialog.ShowError(err, win)
		return
	}

	message := fmt.Sprintf(
		"Your %d intake log %s will be permanently removed.\n\nMedications and their schedules will not be affected.",
		intakeCount, plural("record", intakeCount),
	)

	showConfirmSheet(win, "Clear Intake History", message, "Clear History",
		onClearHistory, "Intake history has been cleared.")
}

// showResetAllFlow shows a confirmation sheet before erasing all data.
func showResetAllFlow(
	win fyne.Window,
	getCounts func() (int, int, error),
	onResetAll func() error,
) {
	medCount, intakeCount, err := getCounts()
	if err != nil {
		dialog.ShowError(err, win)
		return
	}

	message := fmt.Sprintf(
		"Your %d %s and %d intake log %s will be permanently removed.\n\nThis will return MedTrack to its initial state.",
		medCount, plural("medication", medCount),
		intakeCount, plural("record", intakeCount),
	)

	showConfirmSheet(win, "Erase All Data", message, "Erase All Data",
		onResetAll, "All data has been erased.")
}

// showConfirmSheet presents a single confirmation dialog with a clearly-labelled
// destructive button. Apple HIG: one well-worded confirmation is sufficient for
// destructive actions — no secondary typing challenge required.
func showConfirmSheet(
	win fyne.Window,
	title string,
	message string,
	confirmLabel string,
	action func() error,
	successMessage string,
) {
	msg := widget.NewLabel(message)
	msg.Wrapping = fyne.TextWrapWord

	dlg := dialog.NewCustomConfirm(
		title,
		confirmLabel,
		"Cancel",
		container.NewPadded(msg),
		func(confirmed bool) {
			if !confirmed {
				return
			}
			if err := action(); err != nil {
				dialog.ShowError(err, win)
				return
			}
			dialog.ShowInformation(title, successMessage, win)
		},
		win,
	)
	dlg.Resize(fyne.NewSize(380, 200))
	dlg.Show()
}

func plural(word string, count int) string {
	if count == 1 {
		return word
	}
	return word + "s"
}
