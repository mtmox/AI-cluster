
package frontend

import (
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/nats-io/nats.go"
)

// Add this global variable
var modelNames = make(map[string]bool)

var conversations []Conversation
var selectedConversation *Conversation
var threadsList *widget.List
var chatOutput *widget.Entry

func StartFrontend(js nats.JetStreamContext, logger *log.Logger) {
	a := app.New()
	w := a.NewWindow("AI Interface")

	tabs := container.NewAppTabs(
		container.NewTabItem("Home", widget.NewLabel("Home Tab Content")),
		container.NewTabItem("Chat", createChatTab()),
		container.NewTabItem("Generate", widget.NewLabel("Generate Tab Content")),
	)

	w.SetContent(tabs)
	w.Resize(fyne.NewSize(1536, 1152))
	configSyncModels(js, logger)
	w.ShowAndRun()
}

