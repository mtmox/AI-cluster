
package frontend

import (
	"log"
	"fmt"
	"encoding/json"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/nats-io/nats.go"

	"github.com/mtmox/AI-cluster/streams"
	"github.com/mtmox/AI-cluster/constants"

)

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
	w.Resize(fyne.NewSize(1024, 768))
	configSyncModels(js, logger)
	w.ShowAndRun()

}

func configSyncModels(js nats.JetStreamContext, logger *log.Logger) error {
	subject := "config.sync.models"
	durable := "config_sync_models"
	
	_, err := streams.DurablePull(js, "nodes", subject, durable, func(msg *nats.Msg) {
		populateModels(msg, logger)
		logger.Printf("Model: %s", b)
	})
	if err != nil {
		return fmt.Errorf("Failed to set up consumer to populate models: %s: %v", subject, err)
	}
	logger.Printf("Consumer set up for subject: %s", subject)
	return nil
}