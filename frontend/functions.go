
package frontend

import (
	"strconv"
	"encoding/json"
	"log"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"

	"github.com/nats-io/nats.go"

	"github.com/mtmox/AI-cluster/streams"
	"github.com/mtmox/AI-cluster/constants"
)

type Conversation struct {
	ID           string
	Threads      []Thread
	ThreadCounter int
}

type Thread struct {
	ID       int
	Messages []Message
}

type Message struct {
	Sender  string
	Content string
}

func updateConversationList(list *widget.List, conversations []Conversation) {
	list.Length = func() int { return len(conversations) }
	list.UpdateItem = func(id widget.ListItemID, item fyne.CanvasObject) {
		item.(*widget.Label).SetText(conversations[id].ID)
	}
	list.Refresh()
}

func updateThreadsList(list *widget.List, threads []Thread) {
	list.Length = func() int { return len(threads) }
	list.UpdateItem = func(id widget.ListItemID, item fyne.CanvasObject) {
		item.(*widget.Label).SetText(strconv.Itoa(threads[id].ID))
	}
	list.Refresh()
}

func updateChatOutput(output *widget.Entry, messages []Message) {
	var content string
	for _, msg := range messages {
		content += msg.Sender + ": " + msg.Content + "\n"
	}
	output.SetText(content)
}

func createThreadBox(number int) fyne.CanvasObject {
	return widget.NewLabel(strconv.Itoa(number))
}

func appendMessage(output *widget.Entry, sender, content string) {
	currentText := output.Text
	newMessage := sender + ": " + content + "\n"
	output.SetText(currentText + newMessage)
}

// Add this function to update the model selector from outside
func updateModelSelector() {
	if modelSelector != nil {
		var names []string
		for name := range modelNames {
			names = append(names, name)
		}
		modelSelector.Options = names
		modelSelector.Refresh()
	}
}

func configSyncModels(js nats.JetStreamContext, logger *log.Logger) error {
	subject := "config.sync.models"
	durable := "config_sync_models"
	
	_, err := streams.DurablePull(js, "nodes", subject, durable, func(msg *nats.Msg) {
		populateModels(msg, logger)
	})
	if err != nil {
		return fmt.Errorf("Failed to set up consumer to populate models: %s: %v", subject, err)
	}
	logger.Printf("Consumer set up for subject: %s", subject)
	return nil
}

func populateModels(msg *nats.Msg, logger *log.Logger) {
    var model constants.ConfigSyncModels
	err := json.Unmarshal(msg.Data, &model)
	if err != nil {
		logger.Printf("Error unmarshaling model data: %v", err)
		return
	}
	
	if !modelNames[model.Name] {
		modelNames[model.Name] = true
		logger.Printf("Added new model: %s", model.Name)
		updateModelSelector()
	}
}
