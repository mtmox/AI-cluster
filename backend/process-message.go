
package backend

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/nats-io/nats.go"
	"github.com/mtmox/AI-cluster/streams"
	"github.com/mtmox/AI-cluster/constants"
)

// ChatMessage represents the structure for chat messages
type ChatMessage struct {
	Role    string `json:"role"`    // Role can be "user", "assistant", or "system"
	Content string `json:"content"`
}

// ChatRequest represents the structure for the Ollama API request
type ChatRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
	Stream   bool         `json:"stream"`
}

// ChatResponse represents the structure for the Ollama API response
type ChatResponse struct {
	Model     string      `json:"model"`
	CreatedAt string      `json:"created_at"`
	Message   ChatMessage `json:"message"`
	Done      bool        `json:"done"`
}

// IncomingMessage represents the structure of incoming messages
type IncomingMessage struct {
	ConversationID string        `json:"conversation_id"`
	ThreadID       int           `json:"thread_id"`
	Model          string        `json:"model"`
	SystemPrompt   string        `json:"system_prompt"`
	Messages       []ChatMessage `json:"messages"`
}

// NATSMessage represents the structure for messages published to NATS
type NATSMessage struct {
	ConversationID string `json:"conversation_id"`
	ThreadID       int    `json:"thread_id"`
	Content        string `json:"content"`
}

// Initialize color functions
var (
	incomingColor = color.New(color.FgCyan).SprintFunc()
	responseColor = color.New(color.FgGreen).SprintFunc()
	idColor       = color.New(color.FgYellow).SprintFunc()
)

func ProcessMessage(js nats.JetStreamContext, logger *log.Logger) {
	if err := nodeConfig(); err != nil {
		logger.Printf("Failed to create node config: %v", err)
		return
	}

	if err := LoadNodeSettings(); err != nil {
		logger.Printf("Failed to load node settings: %v", err)
		return
	}

	streamName := "messages"
	consumerGroup := "message_processors"
	subject := "in.chat.>"

	modelsInfo, err := constants.ReadModelsInfo(constants.ModelsOutputFile)
	if err != nil {
		logger.Printf("Failed to read models info: %v", err)
		return
	}

	// Create a worker pool
	var wg sync.WaitGroup

	messageHandler := func(msg *nats.Msg) bool {
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer FinishProcessing()

			if msg.Header != nil {
				for key, values := range msg.Header {
					for _, value := range values {
						logger.Printf("Header - %s: %s", key, value)
					}
				}

				if modelName := msg.Header.Get("model"); modelName != "" {
					for _, model := range modelsInfo.Models {
						if model.Name == modelName {
							logger.Printf("Processing message for model: %s", modelName)
							logger.Printf(incomingColor("Incoming Message Data: %s"), string(msg.Data))

							response, convID, threadID, err := sendToLLM(msg.Data, logger)
							if err != nil {
								logger.Printf("Error processing message with LLM: %v", err)
								return
							}
							logger.Printf("%s [ConvID: %s, ThreadID: %d] %s",
								idColor("Response"),
								idColor(convID),
								idColor(threadID),
								responseColor(response))

							natsMsg := &NATSMessage{
								ConversationID: convID,
								ThreadID: 		threadID,
								Content:		response,
							}

							if err := publishMessage(js, natsMsg); err != nil {
								logger.Printf("Error publishing message to NATS: %v", err)
								return
							}	
							return
						}
					}
					logger.Printf("Model %s not found in models file", modelName)
					return
				}
			}

			logger.Printf("Message without model header: %s", string(msg.Data))
		}()

		return true
	}

	subscription, err := streams.DurableGroupPull(
		js,
		streamName,
		subject,
		consumerGroup,
		consumerGroup,
		messageHandler,
	)
	if err != nil {
		logger.Fatalf("Failed to create durable group pull subscription: %v", err)
	}

	// Start a goroutine for message processing
	go func() {
		ticker := time.NewTicker(200 * time.Millisecond)
		for {
			select {
			case <-ticker.C:
				availableSlots := GetAvailableProcessingSlots()
				if availableSlots > 0 {
					err := FetchMessages(subscription, messageHandler, availableSlots)
					if err != nil {
						logger.Printf("Error fetching messages: %v", err)
					}
				}
			}
		}
	}()
}

// GetAvailableProcessingSlots returns the number of available processing slots
func GetAvailableProcessingSlots() int {
	tasksLock.Lock()
	defer tasksLock.Unlock()

	settingsLock.RLock()
	defer settingsLock.RUnlock()

	available := settings.MaxParallelRequests - activeTasks
	if available < 0 {
		return 0
	}
	return available
}

// FetchMessages attempts to fetch and process multiple messages up to the specified limit
func FetchMessages(subscription *nats.Subscription, callback func(msg *nats.Msg) bool, limit int) error {
	messages, err := subscription.Fetch(limit)
	if err != nil {
		if err != nats.ErrTimeout {
			return fmt.Errorf("error fetching messages: %v", err)
		}
		return nil
	}

	for _, msg := range messages {
		tasksLock.Lock()
		activeTasks++
		tasksLock.Unlock()

		shouldProcess := callback(msg)
		if !shouldProcess {
			tasksLock.Lock()
			activeTasks--
			tasksLock.Unlock()
			continue
		}

		if err := msg.Ack(); err != nil {
			return fmt.Errorf("error acknowledging message: %v", err)
		}
	}

	return nil
}

func sendToLLM(messageData []byte, logger *log.Logger) (string, string, int, error) {
	var incomingMsg IncomingMessage
	if err := json.Unmarshal(messageData, &incomingMsg); err != nil {
		return "", "", 0, fmt.Errorf("failed to unmarshal message data: %v", err)
	}

	modelManager := GetModelManager()
	if err := modelManager.CheckAndUnloadModels(incomingMsg.Model); err != nil {
		logger.Printf("Warning: Error checking/unloading models: %v", err)
	}

	logger.Printf("%s [ConvID: %s, ThreadID: %d] %+v",
		idColor("Parsed Incoming Message:"),
		idColor(incomingMsg.ConversationID),
		idColor(incomingMsg.ThreadID),
		incomingMsg)

	messages := make([]ChatMessage, 0)

	if incomingMsg.SystemPrompt != "" {
		messages = append(messages, ChatMessage{
			Role:    "system",
			Content: incomingMsg.SystemPrompt,
		})
	}

	messages = append(messages, incomingMsg.Messages...)

	chatRequest := ChatRequest{
		Model:    incomingMsg.Model,
		Messages: messages,
		Stream:   false,
	}

	requestBody, err := json.Marshal(chatRequest)
	if err != nil {
		return "", "", 0, fmt.Errorf("failed to marshal chat request: %v", err)
	}

	logger.Printf("Sending request to Ollama: %s", string(requestBody))

	resp, err := http.Post(constants.ChatEndpoint, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return "", "", 0, fmt.Errorf("failed to send request to Ollama: %v", err)
	}
	defer resp.Body.Close()

	modelManager.UpdateModelUsage(incomingMsg.Model)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", 0, fmt.Errorf("failed to read response body: %v", err)
	}

	var chatResponse ChatResponse
	if err := json.Unmarshal(body, &chatResponse); err != nil {
		return "", "", 0, fmt.Errorf("failed to unmarshal response: %v", err)
	}

	return chatResponse.Message.Content, incomingMsg.ConversationID, incomingMsg.ThreadID, nil
}

func publishMessage(js nats.JetStreamContext, msg *NATSMessage) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("error marshaling message: %v", err)
	}

	// Send to NATS subject (you can modify the subject as needed)
	subject := fmt.Sprintf("out.chat.%s.%d", msg.ConversationID, msg.ThreadID)
	
	err = streams.PublishToNatsOutMessages(js, subject, data)
	if err != nil {
		return fmt.Errorf("error publishing to NATS: %v", err)
	}

	return nil
}