
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/nats-io/nats.go"
	
	"github.com/mtmox/AI-cluster/constants"
	"github.com/mtmox/AI-cluster/streams"
	"github.com/mtmox/AI-cluster/nats_server"
	"github.com/mtmox/AI-cluster/frontend"
	"github.com/mtmox/AI-cluster/backend"
	"github.com/mtmox/AI-cluster/node"
)

func main() {
	// Define flags
	isFrontend := flag.Bool("frontend", false, "Run as frontend instance")
	isBackend := flag.Bool("backend", false, "Run as backend instance")

	// Parse flags
	flag.Parse()

	// Create a logger
	logger := log.New(os.Stdout, "", log.LstdFlags)

	// Check if exactly one flag is set
	if (*isFrontend && *isBackend) || (!*isFrontend && !*isBackend) {
		node.HandleError(fmt.Errorf("invalid flag configuration"), node.FATAL, "Please specify either -frontend or -backend flag")
	}

	// Run the appropriate instance type
	if *isFrontend {
		runFrontend(logger)
		node.HandleError(nil, node.SUCCESS, "Frontend instance completed successfully")
	} else {
		runBackend(logger)
		node.HandleError(nil, node.SUCCESS, "Backend instance completed successfully")
	}
}

func runFrontend(logger *log.Logger) {
	// Connect to NATS server and get the JetStream context
	js, err := nats_server.ConnectToNats()
	if err != nil {
		node.HandleError(err, node.FATAL, "Failed to connect to NATS")
	}
	node.HandleError(nil, node.SUCCESS, "Successfully connected to NATS server")

	node.HandleError(nil, node.SUCCESS, "Frontend instance started successfully")
	frontend.StartFrontend(js, logger)
	time.Sleep(1 * time.Second)
}

func runBackend(logger *log.Logger) {
	// Connect to NATS server and get the JetStream context
	js, err := nats_server.ConnectToNats()
	if err != nil {
		node.HandleError(err, node.FATAL, "Failed to connect to NATS")
	}
	node.HandleError(nil, node.SUCCESS, "Successfully connected to NATS server")

	// Sync models without storing the return value
	_, err = syncModels(js, logger)
	if err != nil {
		node.HandleError(err, node.FATAL, "Error syncing models")
	}
	node.HandleError(nil, node.SUCCESS, "Models synced successfully")

	node.HandleError(nil, node.SUCCESS, "Backend instance started successfully")
	backend.StartBackend(js, logger)
	time.Sleep(1 * time.Second)
}

func syncModels(js nats.JetStreamContext, logger *log.Logger) ([]string, error) {
	// Query and write models
	err := constants.QueryAndWriteModels()
	if err != nil {
		node.HandleError(err, node.ERROR, "Error querying and writing models")
		return nil, err
	}
	node.HandleError(nil, node.SUCCESS, "Models have been successfully written to the JSON file")

	// Read models information
	modelsFile := constants.ModelsOutputFile
	modelsResp, err := constants.ReadModelsInfo(modelsFile)
	if err != nil {
		node.HandleError(err, node.ERROR, "Error reading models information")
		return nil, err
	}
	node.HandleError(nil, node.SUCCESS, "Successfully read models information")

	// Extract model names and publish messages for each model
	var modelNames []string
	for _, model := range modelsResp.Models {
		modelNames = append(modelNames, model.Name)

		msg := constants.ConfigSyncModels{
			Name: model.Name,
		}

		// Publish modelNames to NATS
		subject := "config.sync.models"
		err = streams.PublishToNats(js, subject, msg)
		if err != nil {
			node.HandleError(err, node.ERROR, fmt.Sprintf("Failed to publish model message for %s", model.Name))
			return nil, err
		}
		node.HandleError(nil, node.SUCCESS, fmt.Sprintf("Successfully published model sync message for: %s", model.Name))
	}

	return modelNames, nil
}