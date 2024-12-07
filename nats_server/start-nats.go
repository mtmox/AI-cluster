
package nats_server

import (
	"log"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/mtmox/AI-cluster/constants"
	"github.com/mtmox/AI-cluster/streams"
	"github.com/mtmox/AI-cluster/node"
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
		node.HandleError(err, node.ERROR, "Failed to connect to NATS after 5 attempts")
		return nil, err
	}

	node.HandleError(nil, node.SUCCESS, "Successfully connected to NATS")

	// Create JetStream context
	js, err = nc.JetStream()
	if err != nil {
		node.HandleError(err, node.ERROR, "Failed to create JetStream context")
		return nil, err
	}

	node.HandleError(nil, node.SUCCESS, "JetStream context created successfully")

	// Create streams for each configuration
	for _, streamConfig := range streams.Streams {
		err := createStream(js, streamConfig)
		if err != nil {
			node.HandleError(err, node.WARNING, "Error creating stream "+streamConfig.Name)
		} else {
			node.HandleError(nil, node.SUCCESS, "Stream "+streamConfig.Name+" created/verified successfully")
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
				node.HandleError(err, node.ERROR, "Failed to create stream "+config.Name)
				return err
			}
			node.HandleError(nil, node.SUCCESS, "Stream "+config.Name+" created successfully")
		} else {
			node.HandleError(err, node.ERROR, "Error getting stream info for "+config.Name)
			return err
		}
	} else {
		node.HandleError(nil, node.INFO, "Stream "+config.Name+" already exists")
		log.Printf("Stream %s already exists with config: %+v", config.Name, streamInfo.Config)
	}
	return nil
}
