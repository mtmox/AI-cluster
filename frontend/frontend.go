
package frontend

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"strconv"
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
	w.Resize(fyne.NewSize(1024, 768))
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
	chatOutput := widget.NewMultiLineEntry()
	chatOutput.Disable()

	// Navigation arrows
	leftArrow := widget.NewButton("<", func() {})
	rightArrow := widget.NewButton(">", func() {})

	// Send button
	sendButton := widget.NewButton("Send", func() {
		message := messageBar.Text
		if message != "" {
			appendMessage(chatOutput, "User", message)
			messageBar.SetText("")
		}
	})

	// Layout
	topContainer := container.NewBorder(nil, nil, nil, sendButton, messageBar)
	leftSide := container.NewVSplit(conversationList, threadsList)
	rightSide := settingsContainer
	middleContent := container.NewBorder(nil, nil, leftArrow, rightArrow, chatOutput)
	content := container.NewBorder(topContainer, nil, leftSide, rightSide, middleContent)

	return content
}

func createThreadBox(number int) fyne.CanvasObject {
	return widget.NewLabel(strconv.Itoa(number))
}

func appendMessage(output *widget.Entry, sender, content string) {
	currentText := output.Text
	newMessage := sender + ": " + content + "\n"
	output.SetText(currentText + newMessage)
}
