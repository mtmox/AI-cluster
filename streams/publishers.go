
package streams

import (
    "context"
    "encoding/json"
    "fmt"
    "time"

    "github.com/nats-io/nats.go"
    "github.com/mtmox/AI-cluster/node"
)

func PublishToNats(js nats.JetStreamContext, subject string, data interface{}) error {
    jsonData, err := json.Marshal(data)
    if err != nil {
        node.HandleError(err, node.ERROR, fmt.Sprintf("Failed to marshal data for subject %s", subject))
        return fmt.Errorf("failed to marshal data: %v", err)
    }

    // Print the message exactly as it's being published
    fmt.Printf("Publishing message to subject %s:\n%s\n", subject, string(jsonData))
    node.HandleError(nil, node.INFO, fmt.Sprintf("Attempting to publish message to subject: %s", subject))

    // Set a timeout for the publish operation
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    // Attempt to publish with context
    ack, err := js.PublishMsg(&nats.Msg{
        Subject: subject,
        Data:    jsonData,
    }, nats.Context(ctx))

    if err != nil {
        if err == context.DeadlineExceeded {
            node.HandleError(err, node.ERROR, fmt.Sprintf("Publish operation timed out for subject %s", subject))
            return fmt.Errorf("publish operation timed out: %v", err)
        }
        node.HandleError(err, node.ERROR, fmt.Sprintf("Failed to publish message to subject %s", subject))
        return fmt.Errorf("failed to publish message: %v", err)
    }

    if ack == nil {
        node.HandleError(fmt.Errorf("no acknowledgment"), node.ERROR, fmt.Sprintf("No acknowledgment received from stream for subject %s", subject))
        return fmt.Errorf("no acknowledgment received from stream")
    }

    node.HandleError(nil, node.SUCCESS, fmt.Sprintf("Message published successfully to subject %s, sequence: %d", subject, ack.Sequence))
    return nil
}

func PublishToNatsOutMessages(js nats.JetStreamContext, subject string, data []byte) error {
    // Print the message exactly as it's being published
    fmt.Printf("Publishing message to subject %s:\n%s\n", subject, string(data))
    node.HandleError(nil, node.INFO, fmt.Sprintf("Attempting to publish message to subject: %s", subject))

    // Set a timeout for the publish operation
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    // Attempt to publish with context
    ack, err := js.PublishMsg(&nats.Msg{
        Subject: subject,
        Data:    data,
    }, nats.Context(ctx))

    if err != nil {
        if err == context.DeadlineExceeded {
            node.HandleError(err, node.ERROR, fmt.Sprintf("Publish operation timed out for subject %s", subject))
            return fmt.Errorf("publish operation timed out: %v", err)
        }
        node.HandleError(err, node.ERROR, fmt.Sprintf("Failed to publish message to subject %s", subject))
        return fmt.Errorf("failed to publish message: %v", err)
    }

    if ack == nil {
        node.HandleError(fmt.Errorf("no acknowledgment"), node.ERROR, fmt.Sprintf("No acknowledgment received from stream for subject %s", subject))
        return fmt.Errorf("no acknowledgment received from stream")
    }

    node.HandleError(nil, node.SUCCESS, fmt.Sprintf("Message published successfully to subject %s, sequence: %d", subject, ack.Sequence))
    return nil
}

func PublishToNatsWithHeader(js nats.JetStreamContext, subject string, data []byte, header nats.Header) error {
    node.HandleError(nil, node.INFO, fmt.Sprintf("Attempting to publish message to subject: %s", subject))
    fmt.Printf("Message size: %d bytes\n", len(data))
    fmt.Printf("Header: %v\n", header)

    // Set a timeout for the publish operation
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    // Create the message
    msg := &nats.Msg{
        Subject: subject,
        Header:  header,
        Data:    data,
    }

    // Publish the message with options
    ack, err := js.PublishMsg(msg, nats.Context(ctx))

    if err != nil {
        if err == context.DeadlineExceeded {
            node.HandleError(err, node.ERROR, fmt.Sprintf("Publish operation timed out for subject %s", subject))
            return fmt.Errorf("publish operation timed out: %v", err)
        }
        node.HandleError(err, node.ERROR, fmt.Sprintf("Failed to publish message to subject %s", subject))
        return fmt.Errorf("failed to publish message: %v", err)
    }

    if ack == nil {
        node.HandleError(fmt.Errorf("no acknowledgment"), node.ERROR, fmt.Sprintf("No acknowledgment received from stream for subject %s", subject))
        return fmt.Errorf("no acknowledgment received from stream")
    }

    node.HandleError(nil, node.SUCCESS, fmt.Sprintf("Message successfully published to subject: %s, sequence: %d", subject, ack.Sequence))
    return nil
}