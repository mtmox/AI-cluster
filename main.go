
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/nats-io/nats.go"
	
	"github.com/mtmox/AI-cluster/constants"
	"github.com/mtmox/AI-cluster/streams"
	"github.com/mtmox/AI-cluster/nats_server"
	"github.com/mtmox/AI-cluster/frontend"
	"github.com/mtmox/AI-cluster/backend"
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
		logger.Fatal("Please specify either -frontend or -backend flag")
	}

	// Run the appropriate instance type
	if *isFrontend {
		// Start NATS server
		logger.Println("Starting NATS queue")
		nats_server.StartNats()
		runFrontend(logger) 
	} else {
		runBackend(logger)
	}
}

func runFrontend(logger *log.Logger) {
	// Connect to NATS server and get the JetStream context
	js, err := nats_server.ConnectToNats()
	if err != nil {
		logger.Fatalf("Failed to connect to NATS: %v", err)
	}

	// Sync models without storing the return value
	_, err = syncModels(js, logger)
	if err != nil {
		logger.Fatalf("Error syncing models: %v", err)
	}

	logger.Println("Starting frontend instance")
	frontend.StartFrontend(js, logger)
	nats_server.StopNats()
}

func runBackend(logger *log.Logger) {
	// Connect to NATS server and get the JetStream context
	js, err := nats_server.ConnectToNats()
	if err != nil {
		logger.Fatalf("Failed to connect to NATS: %v", err)
	}

	// Sync models without storing the return value
	_, err = syncModels(js, logger)
	if err != nil {
		logger.Fatalf("Error syncing models: %v", err)
	}

	logger.Println("Starting backend instance")
	backend.StartBackend(js, logger)
}

func syncModels(js nats.JetStreamContext, logger *log.Logger) ([]string, error) {
	// Query and write models
	err := constants.QueryAndWriteModels()
	if err != nil {
		return nil, fmt.Errorf("error querying and writing models: %v", err)
	}
	fmt.Println("Models have been successfully written to the JSON file.")

	// Read models information
	modelsFile := constants.ModelsOutputFile
	modelsResp, err := constants.ReadModelsInfo(modelsFile)
	if err != nil {
		return nil, fmt.Errorf("error reading models information: %v", err)
	}

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
			return nil, fmt.Errorf("failed to publish model message for %s: %v", model.Name, err)
		}

		logger.Printf("Published model sync message for: %s", model.Name)
	}

	return modelNames, nil
}