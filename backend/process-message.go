
package backend

import (
	"fmt"
	"log"

	"github.com/nats-io/nats.go"

	"github.com/mtmox/AI-cluster/streams"
	"github.com/mtmox/AI-cluster/node"

)

func ProcessMessage(js nats.JetStreamContext, logger *log.Logger) {
	// Define stream name and subjects
	streamName := "messages"
	salt := node.GetIPWithoutDots()
	durable := fmt.Sprintf("process_messages_%d", salt)
	subject := fmt.Sprintf("in.chat.>")

	// Setup message handler
	messageHandler := func(msg *nats.Msg) {
		// Get header values
		if msg.Header != nil {
			log.Printf("Data: %s", string(msg.Data))
		} else {
			log.Printf("Received message without headers: %s", string(msg.Data))
		}
	}

	// Create durable consumer
	_, err := streams.DurablePull(js, streamName, subject, durable, messageHandler)
	if err != nil {
		log.Fatalf("Failed to create durable consumer: %v", err)
	}
}