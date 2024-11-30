
package backend

import (
	"log"
	"time"

	"github.com/nats-io/nats.go"
)

func ProcessMessage(js nats.JetStreamContext, logger *log.Logger) {
	// Define stream name and subjects
	streamName := "messages"
	consumerGroup := "message_processors"
	subject := "in.chat.>"

	// Create consumer configuration
	consumerConfig := &nats.ConsumerConfig{
		Durable:       consumerGroup,
		DeliverGroup:  consumerGroup,
		DeliverPolicy: nats.DeliverAllPolicy,
		AckPolicy:     nats.AckExplicitPolicy,
		MaxDeliver:    -1, // Unlimited redeliveries
	}

	// Create or get the consumer
	consumer, err := js.ConsumerInfo(streamName, consumerGroup)
	if consumer == nil {
		_, err = js.AddConsumer(streamName, consumerConfig)
		if err != nil {
			logger.Fatalf("Failed to create consumer: %v", err)
		}
	}

	// Create pull subscription
	sub, err := js.PullSubscribe(
		subject,
		consumerGroup,
		nats.Durable(consumerGroup),
		nats.BindStream(streamName),
	)
	if err != nil {
		logger.Fatalf("Failed to create pull subscription: %v", err)
	}
	defer sub.Unsubscribe()

	// Process messages in a loop
	for {
		// Fetch messages (batch size of 10, wait up to 10 ms)
		msgs, err := sub.Fetch(1, nats.MaxWait(time.Millisecond))
		if err != nil {
			if err == nats.ErrTimeout {
				continue // No messages available, continue polling
			}
			logger.Printf("Error fetching messages: %v", err)
			continue
		}

		// Process received messages
		for _, msg := range msgs {
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
	}
}
