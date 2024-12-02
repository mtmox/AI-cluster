
package frontend

import (
	"strconv"
	"encoding/json"
	"log"
	"fmt"
	"strings"

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
	Role  string
	Content string
}

// NATSMessage represents the format we'll send to the NATS queue
type NATSMessage struct {
	ConversationID string    `json:"conversation_id"`
	ThreadID       int       `json:"thread_id"`
	Model         string     `json:"model"`
	SystemPrompt  string     `json:"system_prompt"`
	Messages      []Message  `json:"messages"`
}

// NATSResponse represents the format we receive from the NATS queue
type NATSResponse struct {
	ConversationID string `json:"conversation_id"`
	ThreadID       int    `json:"thread_id"`
	Content        string `json:"content"`
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
		content += msg.Role + ": " + msg.Content + "\n"
	}
	output.SetText(content)
}

func createThreadBox(number int) fyne.CanvasObject {
	return widget.NewLabel(strconv.Itoa(number))
}

func appendMessage(output *widget.Entry, role, content string) {
	currentText := output.Text
	newMessage := role + ": " + content + "\n"
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

func formatMessageForNATS(conv *Conversation, thread Thread, model, promptName string) (*NATSMessage, error) {
	if conv == nil {
		return nil, fmt.Errorf("conversation cannot be nil")
	}

	// Get the system prompt content
	systemPrompt, exists := constants.SystemPrompts[promptName]
	if !exists {
		return nil, fmt.Errorf("system prompt %s not found", promptName)
	}

	natsMsg := &NATSMessage{
		ConversationID: conv.ID,
		ThreadID:       thread.ID,
		Model:         model,
		SystemPrompt:  systemPrompt,
		Messages:      thread.Messages,
	}

	return natsMsg, nil
}

func sendMessageToNATS(js nats.JetStreamContext, msg *NATSMessage) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("error marshaling message: %v", err)
	}

	// Send to NATS subject (you can modify the subject as needed)
	subject := fmt.Sprintf("in.chat.%s.%d", msg.ConversationID, msg.ThreadID)
	
	// Create proper NATS header
	header := make(nats.Header)
	header.Set("model", msg.Model)
	
	err = streams.PublishToNatsWithHeader(js, subject, data, header)
	if err != nil {
		return fmt.Errorf("error publishing to NATS: %v", err)
	}

	return nil
}

func populateAssistants(msg *nats.Msg, logger *log.Logger) {
	parts := strings.Split(msg.Subject, ".")
	if len(parts) != 4 {
		logger.Printf("Invalid subject format: %s", msg.Subject)
		return
	}

	// Parse the response message using the new NATSResponse struct
	var response NATSResponse
	err := json.Unmarshal(msg.Data, &response)
	if err != nil {
		logger.Printf("Error unmarshaling response: %v", err)
		return
	}

	// Find the conversation
	var targetConv *Conversation
	for i := range conversations {
		if conversations[i].ID == response.ConversationID {
			targetConv = &conversations[i]
			break
		}
	}

	if targetConv == nil {
		logger.Printf("Conversation not found: %s", response.ConversationID)
		return
	}

	// Find the thread
	var targetThread *Thread
	for i := range targetConv.Threads {
		if targetConv.Threads[i].ID == response.ThreadID {
			targetThread = &targetConv.Threads[i]
			break
		}
	}

	if targetThread == nil {
		logger.Printf("Thread not found: %d", response.ThreadID)
		return
	}

	// Create and append the assistant message
	newMessage := Message{
		Role:    "Assistant",
		Content: response.Content,
	}
	targetThread.Messages = append(targetThread.Messages, newMessage)

	// Update the chat output if this is the currently selected conversation and thread
	if selectedConversation != nil && 
	   selectedConversation.ID == response.ConversationID && 
	   currentThreadIndex < len(selectedConversation.Threads) &&
	   selectedConversation.Threads[currentThreadIndex].ID == response.ThreadID {
		updateChatOutput(chatOutput, targetThread.Messages)
	}
}

func consumeOutChatMessages(js nats.JetStreamContext, logger *log.Logger) error {
	subject := "out.chat.>"
	durable := "out_chat_messages"
	
	_, err := streams.DurablePull(js, "messages", subject, durable, func(msg *nats.Msg) {
		populateAssistants(msg, logger)
	})
	if err != nil {
		return fmt.Errorf("Failed to set up consumer to populate models: %s: %v", subject, err)
	}
	logger.Printf("Consumer set up for subject: %s", subject)
	return nil
}
