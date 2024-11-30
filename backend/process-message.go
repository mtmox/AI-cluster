
package backend

import (
	"log"

	"github.com/nats-io/nats.go"
	"github.com/mtmox/AI-cluster/streams"
	"github.com/mtmox/AI-cluster/constants"
)

func ProcessMessage(js nats.JetStreamContext, logger *log.Logger) {
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
		// Print all headers
		if msg.Header != nil {
			logger.Println("Message Headers:")
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
