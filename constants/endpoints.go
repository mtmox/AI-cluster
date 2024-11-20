
package constants

import (
	"os"
	"path/filepath"
)

const (
	// BaseURL is the base URL for the API
	BaseURL = "http://localhost:11434"

	// ListModelsEndpoint is the endpoint for listing available models
	ListModelsEndpoint = BaseURL + "/api/tags"
)

var (
	// ModelsOutputFile is the path to the output JSON file
	ModelsOutputFile = filepath.Join(os.ExpandEnv("$HOME"), "AI-cluster", "constants", "models.json")
)
