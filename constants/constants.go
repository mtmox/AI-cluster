
package constants

import (
	"os"
	"path/filepath"
)

const (
	// NatsURL is the computer running the queue
	NatsURL = "nats://192.168.1.16:4222"
	// OllamaURL is the base URL for the API
	OllamaURL = "http://localhost:11434"

	// ListModelsEndpoint is the endpoint for listing available models
	ListModelsEndpoint = OllamaURL + "/api/tags"
)

var (
	// ModelsOutputFile is the path to the output JSON file
	ModelsOutputFile = filepath.Join(os.ExpandEnv("$HOME"), "AI-cluster", "constants", "models.json")
)
