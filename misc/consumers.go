
package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/nats-io/nats.go"
)

func test() {
	
	// Get the number of parallel requests from environment variable
	numParallel, err := strconv.Atoi(os.Getenv("OLLAMA_NUM_PARALLEL"))
	if err != nil {
		log.Fatalf("Error parsing OLLAMA_NUM_PARALLEL: %v", err)
	}

	// Connect to NATS
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = "nats://192.168.1.34:4222" // Default to localhost if not set
	}
	nc, err := nats.Connect(natsURL)
	if err != nil {
		log.Fatalf("Error connecting to NATS: %v", err)
	}
	defer nc.Close()

	// Create JetStream context
	js, err := nc.JetStream()
	if err != nil {
		log.Fatalf("Error creating JetStream context: %v", err)
	}

	// Initialize processor
	processor := NewProcessor(numParallel, natsURL)

	// Ensure the stream exists
	streamName := "incoming_requests_generate"
	subject := "incoming.requests.generate"
	err = ensureStream(js, streamName, subject)
	if err != nil {
		log.Fatalf("Error ensuring stream: %v", err)
	}

	// Subscribe to the JetStream
	sub, err := js.PullSubscribe(subject, "generate_consumer_group")
	if err != nil {
		log.Fatalf("Error subscribing to JetStream: %v", err)
	}
	defer sub.Unsubscribe()

	fmt.Printf("Connected to NATS server at %s\n", natsURL)
	fmt.Println("Main process is running. Listening for messages...")

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start message processing loop
	go processMessages(sub, processor)

	// Wait for termination signal
	<-sigChan

	fmt.Println("Shutting down...")
	processor.Wait()
	fmt.Println("All tasks completed. Exiting.")
}

func processMessages(sub *nats.Subscription, processor *Processor) {
	for {
		msgs, err := sub.Fetch(10, nats.MaxWait(5*time.Second))
		if err != nil {
			if err == nats.ErrTimeout {
				continue
			}
			log.Printf("Error fetching messages: %v", err)
			continue
		}

		for _, msg := range msgs {
			log.Printf("Received message on subject: %s", msg.Subject)
			log.Printf("Message data: %s", string(msg.Data))
			
			// Process the message
			processor.AddTask(msg.Data)

			// Acknowledge the message to remove it from the queue
			if err := msg.Ack(); err != nil {
				log.Printf("Error acknowledging message: %v", err)
				// If acknowledgement fails, we can optionally requeue the message
				// msg.Nak()
			}
		}
	}
}

func ensureStream(js nats.JetStreamContext, streamName, subject string) error {
	stream, err := js.StreamInfo(streamName)
	if err != nil {
		// Stream doesn't exist, let's create it
		_, err = js.AddStream(&nats.StreamConfig{
			Name:     streamName,
			Subjects: []string{subject},
		})
		if err != nil {
			return fmt.Errorf("error creating stream: %v", err)
		}
	} else {
		// Stream exists, let's make sure it has the correct subject
		found := false
		for _, s := range stream.Config.Subjects {
			if s == subject {
				found = true
				break
			}
		}
		if !found {
			stream.Config.Subjects = append(stream.Config.Subjects, subject)
			_, err = js.UpdateStream(&stream.Config)
			if err != nil {
				return fmt.Errorf("error updating stream: %v", err)
			}
		}
	}
	return nil
}