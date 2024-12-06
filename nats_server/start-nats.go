
package nats_server

import (
	"log"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/mtmox/AI-cluster/constants"
	"github.com/mtmox/AI-cluster/streams"
)

func ConnectToNats() (nats.JetStreamContext, error) {
	var nc *nats.Conn
	var js nats.JetStreamContext
	var err error

	// Connect to NATS with retry
	for i := 0; i < 5; i++ {
		nc, err = nats.Connect(constants.NatsURL, nats.Timeout(2*time.Second))
		if err == nil {
			break
		}
		log.Printf("Failed to connect to NATS (attempt %d): %v", i+1, err)
		time.Sleep(100 * time.Millisecond)
	}
	if err != nil {
		return nil, err
	}

	log.Println("Successfully connected to NATS")

	// Create JetStream context
	js, err = nc.JetStream()
	if err != nil {
		return nil, err
	}

	log.Println("JetStream context created successfully")

	// Create streams for each configuration
	for _, streamConfig := range streams.Streams {
		err := createStream(js, streamConfig)
		if err != nil {
			log.Printf("Error creating stream %s: %v", streamConfig.Name, err)
		}
	}

	return js, nil
}

func createStream(js nats.JetStreamContext, config streams.StreamConfig) error {
	streamInfo, err := js.StreamInfo(config.Name)
	if err != nil {
		if err == nats.ErrStreamNotFound {
			log.Printf("Stream %s not found, creating it", config.Name)
			_, err = js.AddStream(&nats.StreamConfig{
				Name:      config.Name,
				Subjects:  config.Subjects,
				Retention: config.Retention,
				Discard:   config.Discard,
				Storage:   nats.FileStorage,
				MaxAge:    config.MaxAge,
				MaxMsgs:   config.MaxMsgs,
			})
			if err != nil {
				return err
			}
			log.Printf("Stream %s created successfully", config.Name)
		} else {
			return err
		}
	} else {
		log.Printf("Stream %s already exists with config: %+v", config.Name, streamInfo.Config)
	}
	return nil
}