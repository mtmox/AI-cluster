
package frontend

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func StartFrontend() {
	a := app.New()
	w := a.NewWindow("AI Interface")

	tabs := container.NewAppTabs(
		container.NewTabItem("Home", widget.NewLabel("Home Tab Content")),
		container.NewTabItem("Chat", createChatTab()),
		container.NewTabItem("Generate", widget.NewLabel("Generate Tab Content")),
	)

	w.SetContent(tabs)
	w.Resize(fyne.NewSize(800, 600))
	w.ShowAndRun()
}

func createChatTab() fyne.CanvasObject {
	// Message bar (Area 1)
	messageBar := widget.NewEntry()
	messageBar.SetPlaceHolder("Type your message here...")

	// Conversation tab (Area 2)
	conversationList := widget.NewList(
		func() int { return 0 },
		func() fyne.CanvasObject { return widget.NewLabel("Conversation") },
		func(id widget.ListItemID, item fyne.CanvasObject) {},
	)

	// Conversation message threads tab (Area 3)
	threadsList := widget.NewList(
		func() int { return 0 },
		func() fyne.CanvasObject { return widget.NewLabel("Thread") },
		func(id widget.ListItemID, item fyne.CanvasObject) {},
	)

	// Settings tab (Area 4)
	killButton := widget.NewButton("Kill", func() {})
	settingsContainer := container.NewVBox(killButton)

	// Chat output area (Area 5)
	chatOutput := createChatOutputArea()

	// Layout
	leftSide := container.NewHSplit(conversationList, threadsList)
	rightSide := container.NewHSplit(chatOutput, settingsContainer)
	content := container.NewBorder(nil, messageBar, leftSide, rightSide)

	return content
}

func createChatOutputArea() fyne.CanvasObject {
	output := widget.NewMultiLineEntry()
	output.Disable()

	leftArrow := widget.NewButton("<", func() {})
	rightArrow := widget.NewButton(">", func() {})

	return container.NewBorder(nil, nil, leftArrow, rightArrow, output)
}
