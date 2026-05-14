package views

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/ashprao/medtrack/internal/version"
)

// ShowAboutDialog shows the About dialog.
// On macOS the menu item must be labelled exactly "About" so Fyne's macOS
// driver automatically moves it into the application menu (top-left menu bar).
func ShowAboutDialog(parent fyne.Window) {
	// Use metadata embedded by fyne package; fall back to the version constant
	// for unpackaged dev builds (go run / go build without fyne package).
	ver := fyne.CurrentApp().Metadata().Version
	if ver == "" {
		ver = version.Version
	}

	appName := widget.NewRichText(&widget.TextSegment{
		Style: widget.RichTextStyle{
			TextStyle: fyne.TextStyle{Bold: true},
			SizeName:  widget.RichTextStyleHeading.SizeName,
			Alignment: fyne.TextAlignCenter,
		},
		Text: "MedTrack",
	})

	verLabel := widget.NewLabel("Version " + ver)
	verLabel.Alignment = fyne.TextAlignCenter

	desc := widget.NewLabel(
		"A desktop application for managing\nmedication schedules and daily intake tracking.")
	desc.Alignment = fyne.TextAlignCenter

	credits := widget.NewLabel("Created by Ashwin Rao\nBuilt with Go and Fyne")
	credits.Alignment = fyne.TextAlignCenter
	credits.Importance = widget.MediumImportance

	content := container.NewVBox(
		container.NewCenter(appName),
		container.NewCenter(verLabel),
		widget.NewSeparator(),
		container.NewCenter(desc),
		widget.NewSeparator(),
		container.NewCenter(credits),
	)

	d := dialog.NewCustom("About MedTrack", "Close", container.NewPadded(content), parent)
	d.Resize(fyne.NewSize(360, 260))
	d.Show()
}
