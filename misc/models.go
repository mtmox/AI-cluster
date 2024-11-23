
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Model struct {
	Name string `json:"name"`
}

type ModelResponse struct {
	Models []Model `json:"models"`
}

type ActiveModel struct {
	Name      string `json:"name"`
	ID        string `json:"id"`
	Size      string `json:"size"`
	Processor string `json:"processor"`
	Until     string `json:"until"`
}

func ensureModelAvailability(modelName string) (bool, string, int) {
	availableModels := listModels()
	if availableModels != nil {
		for _, model := range availableModels {
			if model.Name == modelName {
				fmt.Printf("Model %s is already available.\n", modelName)
				return true, fmt.Sprintf("Model %s is available", modelName), 200
			}
		}
		fmt.Printf("Model %s is not available.\n", modelName)
		return false, fmt.Sprintf("Model %s is not available", modelName), 404
	}
	fmt.Println("Unable to check model availability.")
	return false, "Unable to check model availability", 500
}

func listModels() []Model {
	url := "http://localhost:11434/api/tags"
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Error listing models: %v\n", err)
		return nil
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response body: %v\n", err)
		return nil
	}

	if resp.StatusCode == 200 {
		var modelResp ModelResponse
		err = json.Unmarshal(body, &modelResp)
		if err != nil {
			fmt.Printf("Error unmarshaling JSON: %v\n", err)
			return nil
		}
		return modelResp.Models
	}

	fmt.Printf("Error listing models: %d, %s\n", resp.StatusCode, string(body))
	return nil
}

func activeModels() []ActiveModel {
	url := "http://localhost:11434/api/ps"
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Error getting active models: %v\n", err)
		return nil
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response body: %v\n", err)
		return nil
	}

	if resp.StatusCode == 200 {
		var data map[string][]map[string]interface{}
		err = json.Unmarshal(body, &data)
		if err != nil {
			fmt.Printf("Error unmarshaling JSON: %v\n", err)
			return nil
		}

		var activeModels []ActiveModel
		for _, model := range data["models"] {
			size := fmt.Sprintf("%.1f GB", float64(model["size"].(float64))/(1024*1024*1024))
			processor := fmt.Sprintf("%.0f%% GPU", model["gpu"].(float64))
			activeModel := ActiveModel{
				Name:      model["name"].(string),
				ID:        model["id"].(string),
				Size:      size,
				Processor: processor,
				Until:     model["until"].(string),
			}
			activeModels = append(activeModels, activeModel)
		}
		return activeModels
	}

	return nil
}
