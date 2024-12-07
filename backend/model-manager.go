
package backend

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/mtmox/AI-cluster/constants"
	"github.com/mtmox/AI-cluster/node"
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
		node.HandleError(err, node.ERROR, "Failed to get loaded models")
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		node.HandleError(err, node.ERROR, "Failed to read response body")
		return nil, err
	}

	node.HandleError(nil, node.INFO, "Raw response from Ollama API: "+string(body))

	var modelResp ModelResponse
	if err := json.Unmarshal(body, &modelResp); err != nil {
		node.HandleError(err, node.ERROR, "Failed to unmarshal response")
		return nil, err
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
		node.HandleError(err, node.ERROR, "Failed to marshal unload request")
		return err
	}

	resp, err := http.Post(constants.ChatEndpoint, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		node.HandleError(err, node.ERROR, "Failed to send unload request")
		return err
	}
	defer resp.Body.Close()

	mm.mutex.Lock()
	delete(mm.loadedModels, modelName)
	mm.mutex.Unlock()

	return nil
}

// CheckAndUnloadModels checks if we need to unload any models before loading a new one
func (mm *ModelManager) CheckAndUnloadModels(requestedModel string) error {
	node.HandleError(nil, node.INFO, "Checking and potentially unloading models for requested model: "+requestedModel)
	node.HandleError(nil, node.INFO, "Current maximum allowed models: "+string(mm.maxModels))

	loadedModels, err := mm.GetLoadedModels()
	if err != nil {
		node.HandleError(err, node.ERROR, "Failed to get loaded models")
		return err
	}

	node.HandleError(nil, node.INFO, "Currently loaded models count: "+string(len(loadedModels)))
	for _, model := range loadedModels {
		node.HandleError(nil, node.INFO, "Loaded model: "+model.Model+", Last used: "+model.LastUsed.String())
	}

	// If the requested model is already loaded, no need to unload anything
	for _, model := range loadedModels {
		if model.Model == requestedModel {
			node.HandleError(nil, node.INFO, "Requested model "+requestedModel+" is already loaded, no unloading needed")
			return nil
		}
	}

	// If we have reached the maximum number of loaded models, unload the least recently used
	if len(loadedModels) >= mm.maxModels {
		node.HandleError(nil, node.INFO, "Maximum number of models reached ("+string(mm.maxModels)+"), looking for least recently used model to unload")
		
		var oldestModel *LoadedModelInfo
		for _, model := range loadedModels {
			if oldestModel == nil || model.LastUsed.Before(oldestModel.LastUsed) {
				oldestModel = model
			}
		}

		if oldestModel != nil {
			node.HandleError(nil, node.INFO, "Unloading least recently used model: "+oldestModel.Model+" (Last used: "+oldestModel.LastUsed.String()+")")
			if err := mm.UnloadModel(oldestModel.Model); err != nil {
				node.HandleError(err, node.ERROR, "Failed to unload model "+oldestModel.Model)
				return err
			}
			node.HandleError(nil, node.SUCCESS, "Successfully unloaded model: "+oldestModel.Model)
		}
	} else {
		node.HandleError(nil, node.INFO, "No need to unload any models, current count ("+string(len(loadedModels))+") is below maximum ("+string(mm.maxModels)+")")
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
