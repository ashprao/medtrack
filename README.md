# MedTrack

MedTrack is a desktop application built in Go using the Fyne toolkit for managing medication schedules and tracking daily intake. This project was developed entirely through AI pair programming using Claude Sonnet (via the Cline extension for VSCode), demonstrating the capabilities of AI-assisted software development. Every line of code, from architecture decisions to implementation details, was generated through natural language interaction with the AI, without manual coding.

## Features

- **Medication Management**
  - Add and edit medications with detailed information
  - Track dosage, frequency, and timing
  - Record food-related instructions (before/after meals)
  - Store special instructions and additional notes
  - Set start and end dates for medications

- **Daily Intake Tracking**
  - View medications organized by time of day
  - Track medication intake with checkboxes
  - See food instructions prominently
  - View special instructions when needed
  - Group medications by morning, afternoon, and night

- **User Interface**
  - Clean, intuitive interface
  - Mobile-friendly design with scrollable forms
  - Context-aware titles (Add vs Update)
  - Responsive layout that adapts to screen size
  - Cross-platform support (Desktop & Mobile)
  - Optimized for touch interactions

## Installation

1. Ensure you have Go 1.23 or later installed
2. Install Fyne dependencies:
   ```bash
   # macOS
   brew install golang gcc pkg-config
   ```
3. Clone the repository:
   ```bash
   git clone https://github.com/ashprao/medtrack.git
   cd medtrack
   ```
4. Install dependencies:
   ```bash
   go mod download
   ```
5. Build and run:
   ```bash
   go run cmd/medtrack/main.go
   ```

## Project Structure

```
medtrack/
├── cmd/
│   └── medtrack/
│       └── main.go           # Application entry point
├── internal/
│   ├── db/
│   │   ├── database_test.go  # Database tests
│   │   └── database.go       # Database management
│   ├── models/
│   │   ├── intake_test.go    # Intake model tests
│   │   ├── intake.go         # Intake model
│   │   ├── medication_test.go # Medication model tests
│   │   └── medication.go     # Medication model
│   └── ui/
│       ├── theme/            # UI theme customization
│       └── views/
│           ├── about.go      # About view
│           ├── daily_intake.go # Daily intake view
│           ├── help.go       # Help documentation view
│           ├── medication_form.go # Medication form
│           ├── medication_list_test.go # List view tests
│           └── medication_list.go # Medication list view
├── go.mod
└── README.md
```

## Usage

1. **Adding Medications**
   - Click "Add New Medication" in the Medications tab
   - Form appears with "Add Medication Details" title
   - Fill in medication details:
     - Name and dosage
     - Frequency and timing
     - Food instructions (before/after meals)
     - Special instructions and notes
     - Additional notes (generic names, side effects)
     - Start and end dates
   - Scroll through form to access all fields
   - Save button enables after making changes
   - Click Save to confirm and return to list
   - Click Cancel to discard and return to list

2. **Editing Medications**
   - Find medication in the list
   - Click Edit button
   - Form appears with "Update Medication Details" title
   - Update any details
   - Scroll through form to access all fields
   - Save button enables after making changes
   - Click Update to save and return to list
   - Click Cancel to discard and return to list
   - Click Clear to reset all fields

3. **Tracking Daily Intake**
   - Switch to Daily Intake tab
   - View medications grouped by time
   - Check off medications as taken
   - View food and special instructions
   - Track multiple doses throughout the day

## Development

- Built with Go and Fyne UI toolkit
- Uses SQLite for local data storage
- Follows Go project layout conventions
- Implements MVC-like architecture
- Mobile-first responsive design
- iOS and Android support via Fyne

## Dependencies

- Go 1.23+
- Fyne v2.5.3
- SQLite3
- Other dependencies as specified in go.mod

## Contributing

Contributions are welcome! Here's how you can help:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

Please make sure to:
- Write clear commit messages
- Add tests for new features
- Update documentation as needed
- Follow the existing code style

## License

This project is licensed under the MIT License - see below for details:

```
MIT License

Copyright (c) 2025 Ashwin Rao

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```
