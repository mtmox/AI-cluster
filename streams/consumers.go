
package streams

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/mtmox/AI-cluster/node"
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
		node.HandleError(err, node.ERROR, fmt.Sprintf("Failed to add consumer for subject %s", subject))
		return nil, fmt.Errorf("failed to add consumer: %v", err)
	}

	// Use pull-based subscription
	subscription, err := js.PullSubscribe(subject, durable, nats.BindStream(streamName))
	if err != nil {
		node.HandleError(err, node.ERROR, fmt.Sprintf("Failed to create pull subscription for subject %s", subject))
		return nil, fmt.Errorf("failed to create pull subscription: %v", err)
	}

	// Start a goroutine to fetch messages
	go func() {
		for {
			messages, err := subscription.Fetch(1, nats.MaxWait(100*time.Millisecond))
			if err != nil {
				if err != nats.ErrTimeout {
					node.HandleError(err, node.WARNING, fmt.Sprintf("Error fetching message for subject %s", subject))
				}
				continue
			}
			for _, msg := range messages {
				callback(msg)
				// Acknowledge the message after processing
				if err := msg.Ack(); err != nil {
					node.HandleError(err, node.WARNING, fmt.Sprintf("Error acknowledging message for subject %s", subject))
				}
			}
		}
	}()

	node.HandleError(nil, node.SUCCESS, fmt.Sprintf("Consumer setup complete for subject: %s", subject))
	return subscription, nil
}

func EphemeralPull(js nats.JetStreamContext, streamName string, subject string, callback func(msg *nats.Msg)) (*nats.Subscription, error) {
	subscription, err := js.PullSubscribe(subject, "", nats.BindStream(streamName))
	if err != nil {
		node.HandleError(err, node.ERROR, fmt.Sprintf("Failed to create ephemeral pull subscription for subject %s", subject))
		return nil, fmt.Errorf("failed to create ephemeral pull subscription: %v", err)
	}

	go func() {
		for {
			messages, err := subscription.Fetch(1, nats.MaxWait(1*time.Second))
			if err != nil {
				if err == nats.ErrTimeout {
					node.HandleError(nil, node.INFO, fmt.Sprintf("No more messages for subject %s, stopping consumer", subject))
					subscription.Unsubscribe()
					return
				}
				node.HandleError(err, node.WARNING, fmt.Sprintf("Error fetching message for subject %s", subject))
				continue
			}
			for _, msg := range messages {
				callback(msg)
			}
		}
	}()

	node.HandleError(nil, node.SUCCESS, fmt.Sprintf("Ephemeral consumer setup complete for subject: %s", subject))
	return subscription, nil
}

func DurableGroupPull(
	js nats.JetStreamContext,
	streamName string,
	subject string,
	durableName string,
	queueGroup string,
	callback func(msg *nats.Msg) bool,
) (*nats.Subscription, error) {
	consumerConfig := &nats.ConsumerConfig{
		Durable:       durableName,
		DeliverGroup:  queueGroup,
		AckPolicy:     nats.AckExplicitPolicy,
		FilterSubject: subject,
		DeliverPolicy: nats.DeliverAllPolicy,
		MaxDeliver:    -1,
		AckWait:       30 * time.Second,
	}

	consumer, err := js.ConsumerInfo(streamName, durableName)
	if consumer == nil {
		_, err = js.AddConsumer(streamName, consumerConfig)
		if err != nil && err != nats.ErrConsumerNameAlreadyInUse {
			node.HandleError(err, node.ERROR, fmt.Sprintf("Failed to add consumer for subject %s", subject))
			return nil, fmt.Errorf("failed to add consumer: %v", err)
		}
	}

	subscription, err := js.PullSubscribe(
		subject,
		queueGroup,
		nats.BindStream(streamName),
	)
	if err != nil {
		node.HandleError(err, node.ERROR, fmt.Sprintf("Failed to create pull subscription for subject %s", subject))
		return nil, fmt.Errorf("failed to create pull subscription: %v", err)
	}

	node.HandleError(nil, node.SUCCESS, fmt.Sprintf("Queue group consumer setup complete for subject: %s, queue group: %s", subject, queueGroup))
	return subscription, nil
}

func DurablePush(js nats.JetStreamContext, streamName string, subject string, durable string, callback func(msg *nats.Msg)) (*nats.Subscription, error) {
	subjectHash := hash(subject)

	consumerConfig := &nats.ConsumerConfig{
		Durable:        durable,
		AckPolicy:      nats.AckExplicitPolicy,
		FilterSubject:  subject,
		DeliverPolicy:  nats.DeliverAllPolicy,
		DeliverSubject: fmt.Sprintf("%s.%s", subjectHash, durable),
	}

	_, err := js.AddConsumer(streamName, consumerConfig)
	if err != nil && err != nats.ErrConsumerNameAlreadyInUse {
		node.HandleError(err, node.ERROR, fmt.Sprintf("Failed to add consumer for subject %s", subject))
		return nil, fmt.Errorf("failed to add consumer: %v", err)
	}

	subscription, err := js.Subscribe(consumerConfig.FilterSubject, func(msg *nats.Msg) {
		callback(msg)
	}, nats.Durable(durable), nats.BindStream(streamName))

	if err != nil {
		node.HandleError(err, node.ERROR, fmt.Sprintf("Failed to create push subscription for subject %s", subject))
		return nil, fmt.Errorf("failed to create push subscription: %v", err)
	}

	node.HandleError(nil, node.SUCCESS, fmt.Sprintf("Push-based consumer setup complete for subject: %s", subject))
	return subscription, nil
}

func EphemeralPush(js nats.JetStreamContext, streamName string, subject string, callback func(msg *nats.Msg)) (*nats.Subscription, error) {
	subscription, err := js.PullSubscribe(subject, "", nats.BindStream(streamName))
	if err != nil {
		node.HandleError(err, node.ERROR, fmt.Sprintf("Failed to create ephemeral pull subscription for subject %s", subject))
		return nil, fmt.Errorf("failed to create ephemeral pull subscription: %v", err)
	}

	go func() {
		for {
			messages, err := subscription.Fetch(1, nats.MaxWait(1*time.Second))
			if err != nil {
				if err == nats.ErrTimeout {
					node.HandleError(nil, node.INFO, fmt.Sprintf("No more messages for subject %s, stopping consumer", subject))
					subscription.Unsubscribe()
					return
				}
				node.HandleError(err, node.WARNING, fmt.Sprintf("Error fetching message for subject %s", subject))
				continue
			}
			for _, msg := range messages {
				callback(msg)
			}
		}
	}()

	node.HandleError(nil, node.SUCCESS, fmt.Sprintf("Ephemeral consumer setup complete for subject: %s", subject))
	return subscription, nil
}

func hash(subject string) string {
	hasher := sha256.New()
	hasher.Write([]byte(subject))
	hashBytes := hasher.Sum(nil)
	hashHex := hex.EncodeToString(hashBytes)
	return hashHex[:16]
}







