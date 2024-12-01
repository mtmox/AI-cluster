
package backend

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/mtmox/AI-cluster/constants"
)

// LoadedModelInfo represents information about a loaded model
type LoadedModelInfo struct {
	Model     string    `json:"model"`
	LoadedAt  time.Time `json:"loaded_at"`
	LastUsed  time.Time `json:"last_used"`
	RequestID string    `json:"request_id"`
}

// ModelResponse represents the response from the Ollama API for loaded models
type ModelResponse struct {
	Models []struct {
		Name string `json:"name"`
	} `json:"models"`
}

// ModelManager handles the loading and unloading of models
type ModelManager struct {
	loadedModels map[string]*LoadedModelInfo
	mutex        sync.RWMutex
	maxModels    int
}

var (
	modelManager *ModelManager
	once         sync.Once
)

// GetModelManager returns the singleton instance of ModelManager
func GetModelManager() *ModelManager {
	once.Do(func() {
		settingsLock.RLock()
		maxModels := settings.MaxLoadedModels
		settingsLock.RUnlock()
		
		modelManager = &ModelManager{
			loadedModels: make(map[string]*LoadedModelInfo),
			maxModels:    maxModels,
		}
	})
	return modelManager
}

// GetLoadedModels returns information about currently loaded models
func (mm *ModelManager) GetLoadedModels() ([]*LoadedModelInfo, error) {
	resp, err := http.Get(constants.LoadedModels)
	if err != nil {
		return nil, fmt.Errorf("failed to get loaded models: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	// Print the raw response for debugging
	fmt.Printf("Raw response from Ollama API: %s\n", string(body))

	var modelResp ModelResponse
	if err := json.Unmarshal(body, &modelResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %v", err)
	}

	mm.mutex.Lock()
	defer mm.mutex.Unlock()

	// Update our internal state
	loadedModels := make([]*LoadedModelInfo, 0)
	for _, model := range modelResp.Models {
		info, exists := mm.loadedModels[model.Name]
		if !exists {
			info = &LoadedModelInfo{
				Model:    model.Name,
				LoadedAt: time.Now(),
				LastUsed: time.Now(),
			}
			mm.loadedModels[model.Name] = info
		}
		loadedModels = append(loadedModels, info)
	}

	return loadedModels, nil
}

// UnloadModel attempts to unload a specific model from memory
func (mm *ModelManager) UnloadModel(modelName string) error {
	request := struct {
		Model     string        `json:"model"`
		Messages  []ChatMessage `json:"messages"`
		KeepAlive int          `json:"keep_alive"`
	}{
		Model:     modelName,
		Messages:  []ChatMessage{},
		KeepAlive: 0,
	}

	requestBody, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal unload request: %v", err)
	}

	resp, err := http.Post(constants.ChatEndpoint, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return fmt.Errorf("failed to send unload request: %v", err)
	}
	defer resp.Body.Close()

	mm.mutex.Lock()
	delete(mm.loadedModels, modelName)
	mm.mutex.Unlock()

	return nil
}

// CheckAndUnloadModels checks if we need to unload any models before loading a new one
func (mm *ModelManager) CheckAndUnloadModels(requestedModel string) error {
	fmt.Printf("Checking and potentially unloading models for requested model: %s\n", requestedModel)
	fmt.Printf("Current maximum allowed models: %d\n", mm.maxModels)

	loadedModels, err := mm.GetLoadedModels()
	if err != nil {
		return fmt.Errorf("failed to get loaded models: %v", err)
	}

	fmt.Printf("Currently loaded models count: %d\n", len(loadedModels))
	for _, model := range loadedModels {
		fmt.Printf("Loaded model: %s, Last used: %s\n", model.Model, model.LastUsed)
	}

	// If the requested model is already loaded, no need to unload anything
	for _, model := range loadedModels {
		if model.Model == requestedModel {
			fmt.Printf("Requested model %s is already loaded, no unloading needed\n", requestedModel)
			return nil
		}
	}

	// If we have reached the maximum number of loaded models, unload the least recently used
	if len(loadedModels) >= mm.maxModels {
		fmt.Printf("Maximum number of models reached (%d), looking for least recently used model to unload\n", mm.maxModels)
		
		var oldestModel *LoadedModelInfo
		for _, model := range loadedModels {
			if oldestModel == nil || model.LastUsed.Before(oldestModel.LastUsed) {
				oldestModel = model
			}
		}

		if oldestModel != nil {
			fmt.Printf("Unloading least recently used model: %s (Last used: %s)\n", oldestModel.Model, oldestModel.LastUsed)
			if err := mm.UnloadModel(oldestModel.Model); err != nil {
				return fmt.Errorf("failed to unload model %s: %v", oldestModel.Model, err)
			}
			fmt.Printf("Successfully unloaded model: %s\n", oldestModel.Model)
		}
	} else {
		fmt.Printf("No need to unload any models, current count (%d) is below maximum (%d)\n", len(loadedModels), mm.maxModels)
	}

	return nil
}

// UpdateModelUsage marks a model as recently used
func (mm *ModelManager) UpdateModelUsage(modelName string) {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()

	if info, exists := mm.loadedModels[modelName]; exists {
		info.LastUsed = time.Now()
	}
}
