package models

import (
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Medication struct {
	ID              int64
	Name            string
	Dosage          string
	Frequency       string
	Purpose         string
	FoodInstruction string // Before/After food
	Instructions    string // Other special instructions
	Notes           string // Additional info like generic name, side effects
	StartDate       time.Time
	EndDate         *time.Time // Optional end date
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// Validate checks if the medication has all required fields and valid values
func (m *Medication) Validate() error {
	if m.Name == "" {
		return fmt.Errorf("name is required")
	}
	if m.Dosage == "" {
		return fmt.Errorf("dosage is required")
	}
	if m.Frequency == "" {
		return fmt.Errorf("frequency is required")
	}
	if m.StartDate.IsZero() {
		return fmt.Errorf("start date is required")
	}
	if m.EndDate != nil && !m.EndDate.After(m.StartDate) {
		return fmt.Errorf("end date must be after start date")
	}
	return nil
}

type Frequency struct {
	TimesPerDay int
	Times       []string // Time strings in "HH:MM" format
	DaysOfWeek  []int    // 0 = Sunday, 1 = Monday, etc.
}

// ParseFrequency parses a frequency string into structured data
func ParseFrequency(freq string) Frequency {
	f := Frequency{
		DaysOfWeek: []int{0, 1, 2, 3, 4, 5, 6}, // Default to every day
	}

	log.Printf("Parsing frequency: %s", freq)

	// Split frequency into count and times
	parts := strings.Split(freq, " at ")
	if len(parts) != 2 {
		// Default to once daily at 9 AM if format is invalid
		f.TimesPerDay = 1
		f.Times = []string{"09:00"}
		return f
	}

	// Parse frequency count
	count := strings.TrimSpace(parts[0])
	switch strings.ToLower(count) {
	case "once daily":
		f.TimesPerDay = 1
	case "twice daily":
		f.TimesPerDay = 2
	case "three times daily":
		f.TimesPerDay = 3
	default:
		f.TimesPerDay = 1
	}

	// Parse times
	times := strings.Split(parts[1], ", ")
	f.Times = make([]string, 0, len(times))
	for _, t := range times {
		f.Times = append(f.Times, strings.TrimSpace(t))
	}

	// Sort times to ensure consistent order
	sort.Strings(f.Times)

	log.Printf("Parsed frequency: TimesPerDay=%d, Times=%v", f.TimesPerDay, f.Times)
	return f
}

// GetDailyTimes returns all times this medication should be taken today
func (f Frequency) GetDailyTimes() []time.Time {
	now := time.Now()
	year, month, day := now.Date()
	var times []time.Time

	// Check if today is a scheduled day
	today := int(now.Weekday())
	isScheduledDay := false
	for _, d := range f.DaysOfWeek {
		if d == today {
			isScheduledDay = true
			break
		}
	}

	if !isScheduledDay {
		return times
	}

	// Add each scheduled time for today
	for _, t := range f.Times {
		hour, min := 9, 0 // Default to 9 AM if parsing fails
		if parts := strings.Split(t, ":"); len(parts) == 2 {
			hour, _ = strconv.Atoi(parts[0])
			min, _ = strconv.Atoi(parts[1])
		}
		times = append(times, time.Date(year, month, day, hour, min, 0, 0, now.Location()))
	}

	return times
}
