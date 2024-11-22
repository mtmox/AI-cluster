
package constants

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
)

type Model struct {
	Name       string    `json:"name"`
	ModifiedAt string    `json:"modified_at"`
	Size       int64     `json:"size"`
	Digest     string    `json:"digest"`
	Details    Details   `json:"details"`
}

type Details struct {
	Format            string   `json:"format"`
	Family            string   `json:"family"`
	Families          []string `json:"families"`
	ParameterSize     string   `json:"parameter_size"`
	QuantizationLevel string   `json:"quantization_level"`
}

type ModelsResponse struct {
	Models []Model `json:"models"`
}

func QueryAndWriteModels() error {
	// Make the API request using the constant
	resp, err := http.Get(ListModelsEndpoint)
	if err != nil {
		return fmt.Errorf("error making API request: %v", err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %v", err)
	}

	// Parse the JSON response
	var modelsResp ModelsResponse
	err = json.Unmarshal(body, &modelsResp)
	if err != nil {
		return fmt.Errorf("error parsing JSON response: %v", err)
	}

	// Write the models to a JSON file
	jsonData, err := json.MarshalIndent(modelsResp, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling JSON: %v", err)
	}

	// Ensure the directory exists
	dir := filepath.Dir(ModelsOutputFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("error creating directory: %v", err)
	}

	err = ioutil.WriteFile(ModelsOutputFile, jsonData, 0644)
	if err != nil {
		return fmt.Errorf("error writing to file: %v", err)
	}

	return nil
}

func ReadModelsInfo(inputFile string) (*ModelsResponse, error) {
	// Read the JSON file
	jsonData, err := ioutil.ReadFile(inputFile)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %v", err)
	}

	// Parse the JSON data
	var modelsResp ModelsResponse
	err = json.Unmarshal(jsonData, &modelsResp)
	if err != nil {
		return nil, fmt.Errorf("error parsing JSON: %v", err)
	}

	return &modelsResp, nil
}

func ReadModelField(inputFile, modelName, fieldName string) (string, error) {
	modelsResp, err := ReadModelsInfo(inputFile)
	if err != nil {
		return "", err
	}

	// Find the specified model and field
	for _, model := range modelsResp.Models {
		if model.Name == modelName {
			switch fieldName {
			case "name":
				return model.Name, nil
			case "modified_at":
				return model.ModifiedAt, nil
			case "size":
				return fmt.Sprintf("%d", model.Size), nil
			case "digest":
				return model.Digest, nil
			case "format":
				return model.Details.Format, nil
			case "family":
				return model.Details.Family, nil
			case "parameter_size":
				return model.Details.ParameterSize, nil
			case "quantization_level":
				return model.Details.QuantizationLevel, nil
			default:
				return "", fmt.Errorf("unknown field: %s", fieldName)
			}
		}
	}

	return "", fmt.Errorf("model not found: %s", modelName)
}
