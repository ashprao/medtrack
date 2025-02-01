package views

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

const (
	AppVersion = "1.1.0"
)

type About struct {
	container *fyne.Container
}

func NewAbout() *About {
	a := &About{}

	// Create content
	content := container.NewVBox()

	// App name with large text
	appName := canvas.NewText("MedTrack", nil)
	appName.TextSize = 24
	appName.TextStyle = fyne.TextStyle{Bold: true}
	content.Add(container.NewCenter(appName))

	// Version
	version := widget.NewRichText()
	version.Segments = []widget.RichTextSegment{
		&widget.TextSegment{
			Style: widget.RichTextStyle{
				TextStyle: fyne.TextStyle{Bold: true},
			},
			Text: "Version: ",
		},
		&widget.TextSegment{
			Text: AppVersion + "\n\n",
		},
	}
	content.Add(version)

	// Description
	description := widget.NewRichTextWithText(
		"MedTrack is a desktop application for managing medication schedules " +
			"and tracking daily intake. It helps users maintain their medication " +
			"regimen by providing clear schedules and easy tracking.\n\n")
	description.Wrapping = fyne.TextWrapWord
	content.Add(description)

	// Features header
	featuresHeader := widget.NewRichTextWithText("Key Features:\n\n")
	featuresHeader.Wrapping = fyne.TextWrapWord
	featuresHeader.Segments[0].(*widget.TextSegment).Style.TextStyle = fyne.TextStyle{Bold: true}
	content.Add(featuresHeader)

	// Features list
	featuresList := widget.NewRichTextWithText(
		"• Medication management with detailed information\n" +
			"• Daily intake tracking with time-based organization\n" +
			"• Food instruction support (before/after meals)\n" +
			"• Special instructions and additional notes\n" +
			"• Easy-to-use interface\n\n")
	featuresList.Wrapping = fyne.TextWrapWord
	content.Add(featuresList)

	// Credits header
	creditsHeader := widget.NewRichTextWithText("Credits:\n\n")
	creditsHeader.Wrapping = fyne.TextWrapWord
	creditsHeader.Segments[0].(*widget.TextSegment).Style.TextStyle = fyne.TextStyle{Bold: true}
	content.Add(creditsHeader)

	// Credits content
	creditsContent := widget.NewRichTextWithText("Created by Ashwin Rao\n\n")
	creditsContent.Wrapping = fyne.TextWrapWord
	content.Add(creditsContent)

	// Built with
	builtWith := widget.NewRichTextWithText("Built with Go and Fyne UI Toolkit\n")
	builtWith.Wrapping = fyne.TextWrapWord
	builtWith.Segments[0].(*widget.TextSegment).Style.TextStyle = fyne.TextStyle{Italic: true}
	content.Add(builtWith)

	// Create scrollable container
	scroll := container.NewVScroll(content)
	scroll.SetMinSize(fyne.NewSize(600, 400))

	// Add padding around content
	a.container = container.NewPadded(scroll)

	return a
}

func (a *About) Container() fyne.CanvasObject {
	return a.container
}
