
package streams

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"time"

	"github.com/nats-io/nats.go"
)

func DurablePull(js nats.JetStreamContext, streamName string, subject string, durable string, callback func(msg *nats.Msg)) (*nats.Subscription, error) {
	consumerConfig := &nats.ConsumerConfig{
		Durable:       durable,
		AckPolicy:     nats.AckExplicitPolicy,
		FilterSubject: subject,
		DeliverPolicy: nats.DeliverAllPolicy,
	}

	_, err := js.AddConsumer(streamName, consumerConfig)
	if err != nil && err != nats.ErrConsumerNameAlreadyInUse {
		return nil, fmt.Errorf("failed to add consumer: %v", err)
	}

	// Use pull-based subscription
	subscription, err := js.PullSubscribe(subject, durable, nats.BindStream(streamName))
	if err != nil {
		return nil, fmt.Errorf("failed to create pull subscription: %v", err)
	}

	// Start a goroutine to fetch messages
	go func() {
		for {
			messages, err := subscription.Fetch(1, nats.MaxWait(10*time.Millisecond))
			if err != nil {
				// log.Printf("Error fetching message: %v", err)
				continue
			}
			for _, msg := range messages {
				callback(msg)
				// Acknowledge the message after processing
				if err := msg.Ack(); err != nil {
					log.Printf("Error acknowledging message: %v", err)
				}
			}
		}
	}()

	log.Printf("Consumer setup complete for subject: %s", subject)

	return subscription, nil
}

func EphemeralPull(js nats.JetStreamContext, streamName string, subject string, callback func(msg *nats.Msg)) (*nats.Subscription, error) {
	// Use pull-based subscription with an ephemeral consumer
	subscription, err := js.PullSubscribe(subject, "", nats.BindStream(streamName))
	if err != nil {
		return nil, fmt.Errorf("failed to create ephemeral pull subscription: %v", err)
	}

	// Start a goroutine to fetch messages
	go func() {
		for {
			messages, err := subscription.Fetch(1, nats.MaxWait(1*time.Second))
			if err != nil {
				if err == nats.ErrTimeout {
					log.Printf("No more messages for subject %s, stopping consumer", subject)
					subscription.Unsubscribe()
					return
				}
				log.Printf("Error fetching message: %v", err)
				continue
			}
			for _, msg := range messages {
				callback(msg)
			}
		}
	}()

	log.Printf("Ephemeral consumer setup complete for subject: %s", subject)

	return subscription, nil
}

// SetupPushConsumer sets up a push-based consumer for the given stream and subject
func DurablePush(js nats.JetStreamContext, streamName string, subject string, durable string, callback func(msg *nats.Msg)) (*nats.Subscription, error) {
	
	subjectHash := hash(subject) 

	consumerConfig := &nats.ConsumerConfig{
		Durable:       durable,
		AckPolicy:     nats.AckExplicitPolicy,
		FilterSubject: subject,
		DeliverPolicy: nats.DeliverAllPolicy,
		DeliverSubject: fmt.Sprintf("%s.%s", subjectHash, durable), // Create a unique delivery subject
	}

	_, err := js.AddConsumer(streamName, consumerConfig)
	if err != nil && err != nats.ErrConsumerNameAlreadyInUse {
		return nil, fmt.Errorf("failed to add consumer: %v", err)
	}

	// Use push-based subscription
	subscription, err := js.Subscribe(consumerConfig.FilterSubject, func(msg *nats.Msg) {
		callback(msg)
	}, nats.Durable(durable), nats.BindStream(streamName))

	if err != nil {
		return nil, fmt.Errorf("failed to create push subscription: %v", err)
	}

	log.Printf("Push-based consumer setup complete for subject: %s", subject)

	return subscription, nil
}

func EphemeralPush(js nats.JetStreamContext, streamName string, subject string, callback func(msg *nats.Msg)) (*nats.Subscription, error) {
	// Use pull-based subscription with an ephemeral consumer
	subscription, err := js.PullSubscribe(subject, "", nats.BindStream(streamName))
	if err != nil {
		return nil, fmt.Errorf("failed to create ephemeral pull subscription: %v", err)
	}

	// Start a goroutine to fetch messages
	go func() {
		for {
			messages, err := subscription.Fetch(1, nats.MaxWait(1*time.Second))
			if err != nil {
				if err == nats.ErrTimeout {
					log.Printf("No more messages for subject %s, stopping consumer", subject)
					subscription.Unsubscribe()
					return
				}
				log.Printf("Error fetching message: %v", err)
				continue
			}
			for _, msg := range messages {
				callback(msg)
			}
		}
	}()

	log.Printf("Ephemeral consumer setup complete for subject: %s", subject)

	return subscription, nil
}

// hash takes a subject string, hashes it, and returns an alphanumeric output
func hash(subject string) string {
	// Create a new SHA256 hash
	hasher := sha256.New()
	
	// Write the subject to the hasher
	hasher.Write([]byte(subject))
	
	// Get the hash sum as bytes
	hashBytes := hasher.Sum(nil)
	
	// Convert the hash to a hex string
	hashHex := hex.EncodeToString(hashBytes)
	
	// Take the first 16 characters of the hex string to keep it shorter
	return hashHex[:16]
}







