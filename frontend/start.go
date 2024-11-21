
package frontend

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

var conversations []Conversation
var selectedConversation *Conversation
var threadsList *widget.List
var chatOutput *widget.Entry

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
