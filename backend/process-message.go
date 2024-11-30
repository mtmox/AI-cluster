
package backend

import (
	"log"

	"github.com/nats-io/nats.go"

	"github.com/mtmox/AI-cluster/streams"
)

func ProcessMessage(js nats.JetStreamContext, logger *log.Logger) {
	// Define stream name and subjects
	streamName := "messages"
	consumerGroup := "message_processors"
	subject := "in.chat.>"

	// Define message handler callback
	messageHandler := func(msg *nats.Msg) {
		// Print all headers
		if msg.Header != nil {
			logger.Println("Message Headers:")
			for key, values := range msg.Header {
				for _, value := range values {
					logger.Printf("Header - %s: %s", key, value)
				}
			}
			logger.Printf("Message Data: %s", string(msg.Data))
		} else {
			logger.Printf("Received message without headers: %s", string(msg.Data))
		}
	}

	// Set up durable pull subscription with queue group
	_, err := streams.DurableGroupPull(
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
