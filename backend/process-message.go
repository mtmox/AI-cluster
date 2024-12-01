
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
	Sender  string `json:"Sender,omitempty"`  // For incoming messages
	Role    string `json:"role,omitempty"`    // For Ollama API
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
	// First create the node config file
	if err := nodeConfig(); err != nil {
		logger.Printf("Failed to create node config: %v", err)
		return
	}
	
	// Load the node settings at startup
	if err := LoadNodeSettings(); err != nil {
		logger.Printf("Failed to load node settings: %v", err)
		return
	}

	streamName := "messages"
	consumerGroup := "message_processors"
	subject := "in.chat.>"

	// Get models information from the file
	modelsInfo, err := constants.ReadModelsInfo(constants.ModelsOutputFile)
	if err != nil {
		logger.Printf("Failed to read models info: %v", err)
		return
	}

	// Modified message handler that returns bool
	messageHandler := func(msg *nats.Msg) bool {
		// Check if we can process a new message
		if !CanProcessMessage() {
			return false
		}

		defer FinishProcessing()

		// Print all headers
		if msg.Header != nil {
			for key, values := range msg.Header {
				for _, value := range values {
					logger.Printf("Header - %s: %s", key, value)
				}
			}
			
			// Check if this consumer can handle the model specified in the header
			if modelName := msg.Header.Get("model"); modelName != "" {
				// Check if the model exists in the models file
				for _, model := range modelsInfo.Models {
					if model.Name == modelName {
						logger.Printf("Processing message for model: %s", modelName)
						logger.Printf(incomingColor("Incoming Message Data: %s"), string(msg.Data))
						
						// Process the message with the LLM
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

	// Set up durable pull subscription with queue group
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
	// Parse the incoming message
	var incomingMsg IncomingMessage
	if err := json.Unmarshal(messageData, &incomingMsg); err != nil {
		return "", "", 0, fmt.Errorf("failed to unmarshal message data: %v", err)
	}

	// Log the parsed incoming message
	logger.Printf("%s [ConvID: %s, ThreadID: %d] %+v", 
		idColor("Parsed Incoming Message:"),
		idColor(incomingMsg.ConversationID),
		idColor(incomingMsg.ThreadID),
		incomingMsg)

	// Prepare the chat messages
	messages := make([]ChatMessage, 0)
	
	// Add system message if present
	if incomingMsg.SystemPrompt != "" {
		messages = append(messages, ChatMessage{
			Role:    "system",
			Content: incomingMsg.SystemPrompt,
		})
	}
	
	// Add the conversation messages with proper role mapping
	for _, msg := range incomingMsg.Messages {
		role := "user"
		if msg.Sender != "" {
			// Convert Sender to lowercase role
			role = convertSenderToRole(msg.Sender)
		}
		messages = append(messages, ChatMessage{
			Role:    role,
			Content: msg.Content,
		})
	}

	// Prepare the request to Ollama
	chatRequest := ChatRequest{
		Model:    incomingMsg.Model,
		Messages: messages,
		Stream:   false, // We're not using streaming for now
	}

	// Convert the request to JSON
	requestBody, err := json.Marshal(chatRequest)
	if err != nil {
		return "", "", 0, fmt.Errorf("failed to marshal chat request: %v", err)
	}

	// Log the request being sent to Ollama
	logger.Printf("Sending request to Ollama: %s", string(requestBody))

	// Make the HTTP request to Ollama
	resp, err := http.Post(constants.ChatEndpoint, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return "", "", 0, fmt.Errorf("failed to send request to Ollama: %v", err)
	}
	defer resp.Body.Close()

	// Read the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", 0, fmt.Errorf("failed to read response body: %v", err)
	}

	// Parse the response
	var chatResponse ChatResponse
	if err := json.Unmarshal(body, &chatResponse); err != nil {
		return "", "", 0, fmt.Errorf("failed to unmarshal response: %v", err)
	}

	return chatResponse.Message.Content, incomingMsg.ConversationID, incomingMsg.ThreadID, nil
}

// convertSenderToRole converts the Sender field to the appropriate role
func convertSenderToRole(sender string) string {
	switch sender {
	case "User":
		return "user"
	case "Assistant":
		return "assistant"
	case "System":
		return "system"
	default:
		return "user" // Default to user if unknown
	}
}
