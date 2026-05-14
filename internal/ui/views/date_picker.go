package views

import (
	"errors"
	"time"

	xcalendar "fyne.io/x/fyne/widget"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// newDateEntryWithPicker returns a VBox containing:
//   - a row with a text Entry (YYYY-MM-DD) and a calendar picker button
//   - a small hint label showing the expected format
//
// The entry is pre-filled with initial if non-zero. window is required to
// parent the picker dialog. When allowFuture is false, selecting or typing a
// date beyond today is blocked. An optional minDateEntry reference enforces
// that this entry's date must be on or after the value in that entry.
func newDateEntryWithPicker(initial time.Time, window fyne.Window, allowFuture bool, minDateEntry ...*widget.Entry) (*widget.Entry, *fyne.Container) {
	today := func() time.Time {
		y, m, d := time.Now().Date()
		return time.Date(y, m, d, 0, 0, 0, 0, time.Local)
	}

	// minEntry is the optional lower-bound entry (e.g. the From field for a To field).
	var minEntry *widget.Entry
	if len(minDateEntry) > 0 {
		minEntry = minDateEntry[0]
	}

	entry := widget.NewEntry()
	entry.SetPlaceHolder("YYYY-MM-DD")
	if !initial.IsZero() {
		entry.SetText(initial.Format("2006-01-02"))
	}

	entry.Validator = func(s string) error {
		if s == "" {
			return nil
		}
		t, err := time.ParseInLocation("2006-01-02", s, time.Local)
		if err != nil {
			return errors.New("use YYYY-MM-DD format")
		}
		if !allowFuture && t.After(today()) {
			return errors.New("date cannot be in the future")
		}
		if minEntry != nil {
			min, err := time.ParseInLocation("2006-01-02", minEntry.Text, time.Local)
			if err == nil && t.Before(min) {
				return errors.New("date cannot be before From date")
			}
		}
		return nil
	}

	pickerBtn := widget.NewButtonWithIcon("", theme.HistoryIcon(), func() {
		// Determine the starting month for the calendar.
		t := time.Now()
		if parsed, err := time.Parse("2006-01-02", entry.Text); err == nil {
			t = parsed
		}

		var d dialog.Dialog
		cal := xcalendar.NewCalendar(t, func(selected time.Time) {
			// Truncate to midnight for clean comparisons.
			sel := time.Date(selected.Year(), selected.Month(), selected.Day(), 0, 0, 0, 0, time.Local)
			if !allowFuture && sel.After(today()) {
				dialog.ShowInformation("Invalid Date", "You cannot select a future date.", window)
				return
			}
			if minEntry != nil {
				min, err := time.ParseInLocation("2006-01-02", minEntry.Text, time.Local)
				if err == nil && sel.Before(min) {
					dialog.ShowInformation("Invalid Date", "To date cannot be before From date.", window)
					return
				}
			}
			entry.SetText(selected.Format("2006-01-02"))
			d.Hide()
		})
		d = dialog.NewCustom("Select Date", "Cancel", cal, window)
		d.Show()
	})

	hint := widget.NewLabel("Format: YYYY-MM-DD  e.g. " + time.Now().Format("2006-01-02"))
	hint.Importance = widget.LowImportance

	row := container.NewBorder(nil, nil, nil, pickerBtn, entry)
	content := container.NewVBox(row, hint)

	return entry, content
}
