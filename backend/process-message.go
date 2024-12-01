
package backend

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

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

	messageHandler := func(msg *nats.Msg) bool {
		if !CanProcessMessage() {
			return false
		}

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
							return false
						}
						logger.Printf("%s [ConvID: %s, ThreadID: %d] %s", 
							idColor("Response"),
							idColor(convID),
							idColor(threadID),
							responseColor(response))
						return true
					}
				}
				logger.Printf("Model %s not found in models file", modelName)
				return false
			}
		}
		
		logger.Printf("Message without model header: %s", string(msg.Data))
		return false
	}

	_, err = streams.DurableGroupPull(
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
