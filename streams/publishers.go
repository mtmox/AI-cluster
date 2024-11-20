
package streams

import (
    "encoding/json"
    "fmt"
    "time"
    "context"

    "github.com/nats-io/nats.go"
)

func PublishToNats(js nats.JetStreamContext, subject string, data interface{}) error {
    jsonData, err := json.Marshal(data)
    if err != nil {
        return fmt.Errorf("failed to marshal data: %v", err)
    }

    // Print the message exactly as it's being published
    // fmt.Printf("Publishing message to subject %s:\n%s\n", subject, string(jsonData))

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
            return fmt.Errorf("publish operation timed out: %v", err)
        }
        return fmt.Errorf("failed to publish message: %v", err)
    }

    if ack == nil {
        return fmt.Errorf("no acknowledgment received from stream")
    }

    // fmt.Printf("Message published to JetStream subject %s, sequence: %d\n", subject, ack.Sequence)
    return nil
}

func PublishToNatsWithHeader(js nats.JetStreamContext, subject string, data []byte, header nats.Header) error {
    fmt.Printf("Attempting to publish message to subject: %s\n", subject)
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
            return fmt.Errorf("publish operation timed out: %v", err)
        }
        fmt.Printf("Failed to publish message: %v\n", err)
        return fmt.Errorf("failed to publish message: %v", err)
    }

    if ack == nil {
        fmt.Printf("No acknowledgment received from stream\n")
        return fmt.Errorf("no acknowledgment received from stream")
    }

    fmt.Printf("Message successfully published to subject: %s\n", subject)
    fmt.Printf("Received acknowledgment: Stream=%s, Sequence=%d, Domain=%s\n", ack.Stream, ack.Sequence, ack.Domain)

    return nil
}