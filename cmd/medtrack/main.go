package main

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/ashprao/medtrack/internal/db"
	"github.com/ashprao/medtrack/internal/models"
	"github.com/ashprao/medtrack/internal/ui/views"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
)

// filterActiveMeds returns only medications where Active is true.
func filterActiveMeds(meds []models.Medication) []models.Medication {
	active := make([]models.Medication, 0, len(meds))
	for _, m := range meds {
		if m.Active {
			active = append(active, m)
		}
	}
	return active
}

// filterIntakesForActiveMeds filters intakes for the Daily Intake view.
// Pending intakes for inactive medications are dropped (the DB has already
// deleted them via EnsureDailyIntakes, but this guards the display layer).
// Taken/skipped intakes are always kept regardless of active state — they
// are medical records and must remain visible for the day they were recorded.
func filterIntakesForActiveMeds(intakes []models.Intake, meds []models.Medication) []models.Intake {
	activeIDs := make(map[int64]bool, len(meds))
	for _, m := range meds {
		if m.Active {
			activeIDs[m.ID] = true
		}
	}
	filtered := make([]models.Intake, 0, len(intakes))
	for _, i := range intakes {
		if activeIDs[i.MedicationID] {
			filtered = append(filtered, i)
			continue
		}
		// Preserve taken/skipped rows for inactive meds.
		if i.Status == models.IntakeStatusTaken || i.Status == models.IntakeStatusSkipped {
			filtered = append(filtered, i)
		}
	}
	return filtered
}

func main() {
	// Initialize application.
	// ID must match FyneApp.toml [Details].ID so that fyne package embeds
	// the correct metadata and macOS uses the right bundle identifier.
	myApp := app.NewWithID("com.ashprao.medtrack")
	mainWindow := myApp.NewWindow("MedTrack")

	mainWindow.Resize(fyne.NewSize(800, 600))

	// Setup database
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("Error getting home directory:", err)
	}

	dbPath := filepath.Join(homeDir, ".medtrack.db")
	dbManager, err := db.NewManager(dbPath)
	if err != nil {
		log.Fatal("Error initializing database:", err)
	}
	defer dbManager.Close()

	// Initialize views
	var medicationList *views.MedicationList
	var dailyIntake *views.DailyIntake
	var medicationLog *views.MedicationLog

	// Create main layouts
	medicationsContainer := container.NewStack()

	// Create handlers
	saveHandler := func(med *models.Medication) error {
		// Save to database
		if err := dbManager.SaveMedication(med); err != nil {
			return err
		}

		// Refresh medication list and daily intake
		medications, err := dbManager.GetMedications()
		if err != nil {
			return err
		}
		medicationList.UpdateMedications(medications)

		// Update daily intakes for the new/updated medication (active only)
		activeMeds := filterActiveMeds(medications)
		if err := dbManager.EnsureDailyIntakes(time.Now(), activeMeds); err != nil {
			return err
		}

		// Refresh daily intake view
		intakes, err := dbManager.GetDailyIntakes(time.Now())
		if err != nil {
			return err
		}
		dailyIntake.UpdateMedications(filterActiveMeds(medications))
		dailyIntake.UpdateIntakes(filterIntakesForActiveMeds(intakes, medications))

		return nil
	}

	deleteHandler := func(med *models.Medication) {
		if err := dbManager.DeleteMedication(med.ID); err != nil {
			dialog.ShowError(err, mainWindow)
			return
		}

		// Refresh medication list
		medications, err := dbManager.GetMedications()
		if err != nil {
			dialog.ShowError(err, mainWindow)
			return
		}
		medicationList.UpdateMedications(medications)

		// Refresh daily intake view
		intakes, err := dbManager.GetDailyIntakes(time.Now())
		if err != nil {
			dialog.ShowError(err, mainWindow)
			return
		}
		dailyIntake.UpdateMedications(filterActiveMeds(medications))
		dailyIntake.UpdateIntakes(filterIntakesForActiveMeds(intakes, medications))
	}

	toggleActiveHandler := func(med *models.Medication, active bool) {
		if err := dbManager.UpdateMedicationActive(med.ID, active); err != nil {
			dialog.ShowError(err, mainWindow)
			return
		}
		med.Active = active

		medications, err := dbManager.GetMedications()
		if err != nil {
			dialog.ShowError(err, mainWindow)
			return
		}
		medicationList.UpdateMedications(medications)

		activeMeds := filterActiveMeds(medications)
		if err := dbManager.EnsureDailyIntakes(time.Now(), activeMeds); err != nil {
			dialog.ShowError(err, mainWindow)
			return
		}

		intakes, err := dbManager.GetDailyIntakes(time.Now())
		if err != nil {
			dialog.ShowError(err, mainWindow)
			return
		}
		dailyIntake.UpdateMedications(filterActiveMeds(medications))
		dailyIntake.UpdateIntakes(filterIntakesForActiveMeds(intakes, medications))
	}

	medicationList = views.NewMedicationList(saveHandler, deleteHandler, toggleActiveHandler, mainWindow)
	medicationsContainer.Add(medicationList.Container())

	// Load initial medications and setup daily intakes
	medications, err := dbManager.GetMedications()
	if err != nil {
		log.Printf("Error loading medications: %v", err)
	} else {
		medicationList.UpdateMedications(medications)

		// Initialize daily intake view
		dailyIntake = views.NewDailyIntake(func(med *models.Medication, scheduledTime time.Time, taken bool) {
			if err := dbManager.UpdateIntakeStatus(med.ID, scheduledTime, taken); err != nil {
				dialog.ShowError(err, mainWindow)
				return
			}

			// Get latest medications and intakes
			medications, err := dbManager.GetMedications()
			if err != nil {
				dialog.ShowError(err, mainWindow)
				return
			}

			intakes, err := dbManager.GetDailyIntakes(time.Now())
			if err != nil {
				dialog.ShowError(err, mainWindow)
				return
			}

			// Update with fresh data
			dailyIntake.UpdateMedications(filterActiveMeds(medications))
			dailyIntake.UpdateIntakes(filterIntakesForActiveMeds(intakes, medications))

			// Also refresh the medication log if it's initialized
			if medicationLog != nil {
				medicationLog.RefreshData()
			}
		}, func(medID int64, takenAt time.Time) {
			// PRN "Record Use" handler
			if err := dbManager.AddPRNIntake(medID, takenAt); err != nil {
				dialog.ShowError(err, mainWindow)
				return
			}
			if medicationLog != nil {
				medicationLog.RefreshData()
			}
		})

		// Initialize medication log view
		medicationLog = views.NewMedicationLog(dbManager, mainWindow)

		// Ensure and load initial daily intakes (active medications only)
		activeMeds := filterActiveMeds(medications)
		if err := dbManager.EnsureDailyIntakes(time.Now(), activeMeds); err != nil {
			log.Printf("Error ensuring daily intakes: %v", err)
		}
		// Mark any pending intakes from previous days as missed
		startOfToday := time.Now().Truncate(24 * time.Hour)
		if err := dbManager.MarkMissedIntakes(startOfToday); err != nil {
			log.Printf("Error marking missed intakes: %v", err)
		}
		intakes, err := dbManager.GetDailyIntakes(time.Now())
		if err != nil {
			log.Printf("Error loading daily intakes: %v", err)
		} else {
			dailyIntake.UpdateMedications(filterActiveMeds(medications))
			dailyIntake.UpdateIntakes(filterIntakesForActiveMeds(intakes, medications))
		}
	}

	// "About" (exactly) causes Fyne's macOS driver to move this item into the
	// application menu automatically, giving native macOS About behaviour.
	// "Preferences…" (with ellipsis) is likewise moved to the application menu.
	aboutItem := fyne.NewMenuItem("About", func() {
		views.ShowAboutDialog(mainWindow)
	})
	prefsItem := fyne.NewMenuItem("Preferences…", func() {
		clearHistoryHandler := func() error {
			if err := dbManager.ClearIntakeHistory(); err != nil {
				return err
			}
			// Rebuild daily intake with fresh (empty) intake list
			medications, err := dbManager.GetMedications()
			if err != nil {
				return err
			}
			intakes, err := dbManager.GetDailyIntakes(time.Now())
			if err != nil {
				return err
			}
			dailyIntake.UpdateMedications(filterActiveMeds(medications))
			dailyIntake.UpdateIntakes(filterIntakesForActiveMeds(intakes, medications))
			medicationLog.RefreshData()
			return nil
		}

		resetAllHandler := func() error {
			if err := dbManager.ResetAllData(); err != nil {
				return err
			}
			empty := []models.Medication{}
			medicationList.UpdateMedications(empty)
			dailyIntake.UpdateMedications(empty)
			dailyIntake.UpdateIntakes([]models.Intake{})
			medicationLog.RefreshData()
			return nil
		}

		views.ShowPreferences(mainWindow, dbManager.GetCounts, clearHistoryHandler, resetAllHandler)
	})
	prefsItem.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyComma, Modifier: fyne.KeyModifierSuper}

	helpItem := fyne.NewMenuItem("MedTrack Help", func() {
		views.ShowHelp(mainWindow)
	})

	// "MedTrack" menu: About and Preferences… are auto-lifted by Fyne into the
	// macOS application menu. The separate "Help" menu places MedTrack Help at
	// the right of the native menu bar, matching macOS conventions.
	mainWindow.SetMainMenu(fyne.NewMainMenu(
		fyne.NewMenu("MedTrack", aboutItem, prefsItem),
		fyne.NewMenu("Help", helpItem),
	))

	tabs := container.NewAppTabs(
		container.NewTabItemWithIcon("Daily Intake", theme.HomeIcon(), dailyIntake.Container()),
		container.NewTabItemWithIcon("Medications", theme.ListIcon(), medicationsContainer),
		container.NewTabItemWithIcon("Log", theme.HistoryIcon(), medicationLog.Container()),
	)

	// Add tab change listener to refresh data when switching to Log tab
	tabs.OnChanged = func(tab *container.TabItem) {
		if tab.Text == "Log" {
			// Refresh the log data when switching to the Log tab
			medicationLog.RefreshData()
		}
	}

	mainWindow.SetContent(tabs)
	mainWindow.ShowAndRun()
}
