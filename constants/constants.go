
package constants

import (
	"os"
	"path/filepath"
)

const (
	// NatsURL is the computer running the queue
	NatsURL = "nats://192.168.1.140:4222"
	// OllamaURL is the base URL for the API
	OllamaURL = "http://localhost:11434"

	// ListModelsEndpoint is the endpoint for listing available models
	ListModelsEndpoint = OllamaURL + "/api/tags"
	ChatEndpoint = OllamaURL + "/api/chat"
	GenerateEndpoint = OllamaURL + "/api/generate"
	LoadedModels = OllamaURL + "/api/ps"
	PullModels = OllamaURL + "/api/pull"
	DeleteModels = OllamaURL + "/api/delete"
)

var (
	// ModelsOutputFile is the path to the output JSON file
	ModelsOutputFile = filepath.Join(os.ExpandEnv("$HOME"), "AI-cluster", "constants", "models.json")
	ErrorDatabase = filepath.Join(os.ExpandEnv("$HOME"), "AI-cluster", "node", "errors.db")
)
