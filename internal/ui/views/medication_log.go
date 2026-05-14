package views

import (
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/ashprao/medtrack/internal/db"
	"github.com/ashprao/medtrack/internal/models"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// MedicationLog struct holds the data and UI components for displaying a log of medication intakes
type MedicationLog struct {
	container    *fyne.Container
	dbManager    *db.Manager
	fromDate     *widget.Entry
	toDate       *widget.Entry
	intakes      []models.Intake
	medications  []models.Medication
	dateRangeBox *fyne.Container
	content      *fyne.Container
}

// logEntry represents a single row in the log table
type logEntry struct {
	intakeID    int64
	scheduledAt time.Time
	timeSlot    string
	medication  string
	dosage      string
	status      models.IntakeStatus
	takenAt     *time.Time
}

// NewMedicationLog creates a new medication log view
func NewMedicationLog(dbManager *db.Manager) *MedicationLog {
	m := &MedicationLog{
		dbManager:   dbManager,
		intakes:     make([]models.Intake, 0),
		medications: make([]models.Medication, 0),
	}

	// Create title
	title := widget.NewLabelWithStyle("Medication Intake History", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	// Create date range selector
	now := time.Now()
	oneWeekAgo := now.AddDate(0, 0, -7)

	m.fromDate = widget.NewEntry()
	m.fromDate.SetText(oneWeekAgo.Format("2006-01-02"))
	m.fromDate.SetPlaceHolder("e.g. 2025-06-30")

	m.toDate = widget.NewEntry()
	m.toDate.SetText(now.Format("2006-01-02"))
	m.toDate.SetPlaceHolder("e.g. 2025-06-30")

	refreshButton := widget.NewButtonWithIcon("Refresh", theme.ViewRefreshIcon(), func() {
		from, err := time.Parse("2006-01-02", m.fromDate.Text)
		if err != nil {
			log.Printf("Invalid from date: %v", err)
			return
		}

		to, err := time.Parse("2006-01-02", m.toDate.Text)
		if err != nil {
			log.Printf("Invalid to date: %v", err)
			return
		}

		// Add one day to include the end date fully
		to = to.AddDate(0, 0, 1)

		m.UpdateData(from, to)
	})

	dateEntries := container.NewGridWithColumns(4,
		widget.NewLabel("From:"),
		m.fromDate,
		widget.NewLabel("To:"),
		m.toDate,
	)
	m.dateRangeBox = container.NewVBox(
		dateEntries,
		container.NewBorder(nil, nil, nil, refreshButton, nil),
	)

	// Create scrollable content area
	m.content = container.NewVBox()
	scrollContent := container.NewVScroll(m.content)

	// Create main container
	m.container = container.NewBorder(
		container.NewVBox(
			title,
			widget.NewSeparator(),
			m.dateRangeBox,
			widget.NewSeparator(),
		), // Top
		nil,           // Bottom
		nil,           // Left
		nil,           // Right
		scrollContent, // Center
	)

	// Load initial data
	from, _ := time.Parse("2006-01-02", m.fromDate.Text)
	to, _ := time.Parse("2006-01-02", m.toDate.Text)
	to = to.AddDate(0, 0, 1) // Add one day to include the end date fully

	m.UpdateData(from, to)

	return m
}

// UpdateData refreshes the log data for the specified date range
func (m *MedicationLog) UpdateData(from, to time.Time) {
	// Use GetAllMedications so historical log entries for soft-deleted
	// medications still resolve a name instead of showing "Unknown".
	medications, err := m.dbManager.GetAllMedications()
	if err != nil {
		log.Printf("Error loading medications: %v", err)
		return
	}
	m.medications = medications

	// Get intakes for the date range
	intakes, err := m.dbManager.GetIntakes(from, to)
	if err != nil {
		log.Printf("Error loading intakes: %v", err)
		return
	}
	m.intakes = intakes

	// Refresh the table
	m.renderTable()
}

// renderTable updates the table with the current data
func (m *MedicationLog) renderTable() {
	// Clear existing content
	m.content.RemoveAll()

	// Show message if no items
	if len(m.intakes) == 0 {
		m.content.Add(widget.NewLabel("No medication intake records found for the selected date range"))
		return
	}

	// Create a map of medication IDs to names
	medNames := make(map[int64]string)
	medDosages := make(map[int64]string)
	for _, med := range m.medications {
		medNames[med.ID] = med.Name
		medDosages[med.ID] = med.Dosage
	}

	// Create log entries from intakes
	entries := make([]logEntry, 0, len(m.intakes))
	for _, intake := range m.intakes {
		medName := "Unknown"
		if name, ok := medNames[intake.MedicationID]; ok {
			medName = name
		}

		dosage := ""
		if dose, ok := medDosages[intake.MedicationID]; ok {
			dosage = dose
		}

		// Determine time slot (must match daily_intake.go bucket thresholds)
		hour := intake.ScheduledFor.Hour()
		var timeSlot string
		switch {
		case hour < 12:
			timeSlot = "Morning"
		case hour < 17:
			timeSlot = "Afternoon"
		case hour < 21:
			timeSlot = "Evening"
		default:
			timeSlot = "Bedtime"
		}

		entries = append(entries, logEntry{
			intakeID:    intake.ID,
			scheduledAt: intake.ScheduledFor,
			timeSlot:    timeSlot,
			medication:  medName,
			dosage:      dosage,
			status:      intake.Status,
			takenAt:     intake.TakenAt,
		})
	}

	// Sort entries newest-first (date+time descending)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].scheduledAt.After(entries[j].scheduledAt)
	})

	// Two-level grouping: month → date, both preserving newest-first order.
	type dateGroup struct {
		date    time.Time
		entries []logEntry
	}
	type monthGroup struct {
		month      time.Time // first day of the month
		dateGroups []dateGroup
	}

	var months []monthGroup
	monthIndex := make(map[string]int) // "2006-01" → index in months
	dateIndex := make(map[string]int)  // "2006-01-02" → index in dateGroups slice

	for _, entry := range entries {
		monthKey := entry.scheduledAt.Format("2006-01")
		dateKey := entry.scheduledAt.Format("2006-01-02")

		mi, monthExists := monthIndex[monthKey]
		if !monthExists {
			monthStart, _ := time.Parse("2006-01", monthKey)
			months = append(months, monthGroup{month: monthStart})
			mi = len(months) - 1
			monthIndex[monthKey] = mi
		}

		if _, dateExists := dateIndex[dateKey]; !dateExists {
			d, _ := time.Parse("2006-01-02", dateKey)
			months[mi].dateGroups = append(months[mi].dateGroups, dateGroup{date: d})
			dateIndex[dateKey] = len(months[mi].dateGroups) - 1
		}

		di := dateIndex[dateKey]
		months[mi].dateGroups[di].entries = append(months[mi].dateGroups[di].entries, entry)
	}

	createLogRow := func(entry logEntry) fyne.CanvasObject {
		// Format status — always Title Case for consistency.
		var statusText string
		switch entry.status {
		case models.IntakeStatusTaken:
			statusText = "Taken"
		case models.IntakeStatusSkipped:
			statusText = "Skipped"
		case models.IntakeStatusMissed:
			statusText = "Missed"
		case models.IntakeStatusPending:
			// A pending entry from a previous day is effectively missed.
			startOfToday := time.Now().Truncate(24 * time.Hour)
			if entry.scheduledAt.Before(startOfToday) {
				statusText = "Missed"
			} else {
				statusText = "Pending"
			}
		default:
			statusText = string(entry.status)
		}

		// Compact time + slot label on the left
		metaLabel := widget.NewLabel(fmt.Sprintf("%s · %s", entry.scheduledAt.Format("15:04"), entry.timeSlot))
		metaLabel.Importance = widget.MediumImportance

		// Status right-aligned
		statusLabel := widget.NewLabelWithStyle(statusText, fyne.TextAlignTrailing, fyne.TextStyle{})
		switch statusText {
		case "Pending":
			statusLabel.Importance = widget.WarningImportance
		case "Missed":
			statusLabel.Importance = widget.DangerImportance
		}
		// Delete button
		e := entry
		deleteBtn := widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
			msg := fmt.Sprintf(
				"Delete the %s log entry for %s (%s)?\n\nThis cannot be undone.",
				e.timeSlot, e.medication, e.scheduledAt.Format("15:04"),
			)
			dialog.NewConfirm("Delete Log Entry", msg, func(confirmed bool) {
				if !confirmed {
					return
				}
				if err := m.dbManager.DeleteIntake(e.intakeID); err != nil {
					log.Printf("Error deleting intake %d: %v", e.intakeID, err)
					return
				}
				m.RefreshData()
			}, fyne.CurrentApp().Driver().AllWindows()[0]).Show()
		})
		deleteBtn.Importance = widget.LowImportance

		// Top row: meta left, status centre, delete right
		topRow := container.NewBorder(nil, nil, metaLabel, deleteBtn, statusLabel)

		// Medication name (bold, wrapping) + dosage (muted)
		medLabel := widget.NewLabelWithStyle(
			fmt.Sprintf("%s  %s", entry.medication, entry.dosage),
			fyne.TextAlignLeading,
			fyne.TextStyle{Bold: true},
		)
		medLabel.Wrapping = fyne.TextWrapWord

		rows := container.NewVBox(topRow, medLabel, widget.NewSeparator())
		return container.NewPadded(rows)
	}

	// Render: month section header + per-date accordion
	today := time.Now().Format("2006-01-02")
	firstAccordion := true // track so we auto-open only the most recent date

	for _, mg := range months {
		// Month section label — non-interactive landmark for scrolling
		monthLabel := widget.NewLabelWithStyle(
			mg.month.Format("January 2006"),
			fyne.TextAlignLeading,
			fyne.TextStyle{Bold: true},
		)
		monthLabel.Importance = widget.MediumImportance
		m.content.Add(container.NewPadded(monthLabel))

		// Build accordion items for each date in this month
		var accordionItems []*widget.AccordionItem
		for _, dg := range mg.dateGroups {
			dg := dg // capture
			n := len(dg.entries)
			entryWord := "entries"
			if n == 1 {
				entryWord = "entry"
			}
			title := fmt.Sprintf("%s  ·  %d %s",
				dg.date.Format("Monday, January 2"),
				n, entryWord,
			)
			rows := container.NewVBox()
			for _, entry := range dg.entries {
				rows.Add(createLogRow(entry))
			}
			accordionItems = append(accordionItems, widget.NewAccordionItem(title, rows))
		}

		acc := widget.NewAccordion(accordionItems...)
		acc.MultiOpen = true

		// Auto-open the most recent date (index 0) of the first (most recent) month.
		// Also auto-open any accordion item whose date is today.
		for i, dg := range mg.dateGroups {
			if (firstAccordion && i == 0) || dg.date.Format("2006-01-02") == today {
				acc.Open(i)
			}
		}
		firstAccordion = false

		m.content.Add(acc)
		m.content.Add(widget.NewLabel("")) // breathing room between months
	}
}

// Container returns the main container for this view
func (m *MedicationLog) Container() fyne.CanvasObject {
	return m.container
}

// GetDateRange returns the current from and to dates
func (m *MedicationLog) GetDateRange() (from, to time.Time) {
	from, _ = time.Parse("2006-01-02", m.fromDate.Text)
	to, _ = time.Parse("2006-01-02", m.toDate.Text)
	to = to.AddDate(0, 0, 1) // Add one day to include the end date fully
	return from, to
}

// RefreshData refreshes the log data with the current date range
func (m *MedicationLog) RefreshData() {
	from, to := m.GetDateRange()
	m.UpdateData(from, to)
}
