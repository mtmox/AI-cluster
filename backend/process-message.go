
package backend

import (
	// "fmt"
	"log"

	"github.com/nats-io/nats.go"

	// "github.com/mtmox/AI-cluster/streams"
	// "github.com/mtmox/AI-cluster/node"
)

func ProcessMessage(js nats.JetStreamContext, logger *log.Logger) {
	// Define stream name and subjects
	streamName := "messages"
	consumerGroup := "message_processors"
	subject := "in.chat.>"

	// Setup message handler
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

		// Acknowledge the message
		if err := msg.Ack(); err != nil {
			logger.Printf("Failed to acknowledge message: %v", err)
		}
	}

	// Create consumer configuration
	consumerConfig := &nats.ConsumerConfig{
		Durable:       consumerGroup,
		DeliverGroup: consumerGroup,
		DeliverPolicy: nats.DeliverAllPolicy,
		AckPolicy:    nats.AckExplicitPolicy,
		MaxDeliver:   -1, // Unlimited redeliveries
	}

	// Create or get the consumer
	consumer, err := js.ConsumerInfo(streamName, consumerGroup)
	if consumer == nil {
		_, err = js.AddConsumer(streamName, consumerConfig)
		if err != nil {
			logger.Fatalf("Failed to create consumer: %v", err)
		}
	}

	// Subscribe to the consumer group
	sub, err := js.QueueSubscribe(
		subject,
		consumerGroup,
		messageHandler,
		nats.Durable(consumerGroup),
		nats.ManualAck(),
		nats.BindStream(streamName),
	)
	if err != nil {
		logger.Fatalf("Failed to subscribe to consumer group: %v", err)
	}

	// Keep the subscription active
	defer sub.Unsubscribe()

	// Keep the process running
	// select {}
}
