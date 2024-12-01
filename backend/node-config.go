
package backend

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/shirou/gopsutil/v3/mem"
)

func nodeConfig() error {
	// Query system for total RAM
	v, err := mem.VirtualMemory()
	if err != nil {
		return fmt.Errorf("error getting system memory: %v", err)
	}
	ramGB := int(v.Total / (1024 * 1024 * 1024))

	maxParallel := ramGB / 2

	envVars := map[string]map[string]string{
		"OLLAMA_MAX_LOADED_MODELS": {
			"value":       "2",
			"description": "Maximum number of loaded models per GPU",
		},
		"OLLAMA_NUM_PARALLEL": {
			"value":       strconv.Itoa(maxParallel),
			"description": "Maximum number of parallel requests",
		},
	}

	for key, data := range envVars {
		os.Setenv(key, data["value"])
		fmt.Printf("Set %s=%s - %s\n", key, data["value"], data["description"])
	}

	// Save the node settings
	if err := SaveNodeSettings(maxParallel); err != nil {
		return fmt.Errorf("failed to save node settings: %v", err)
	}

	fmt.Println("\nOllama configuration complete.")
	fmt.Printf("System RAM: %d GB\n", ramGB)
	fmt.Printf("OLLAMA_NUM_PARALLEL: %d\n", maxParallel)
	fmt.Println("Note: You may need to restart Ollama for changes to take effect.")
	fmt.Println("To stop Ollama, use the following command:")
	fmt.Println("pkill ollama")

	// Execute pkill ollama
	cmd := exec.Command("pkill", "ollama")
	err = cmd.Run()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			if exitError.ExitCode() == 1 {
				fmt.Println("Ollama process was not running.")
			} else {
				fmt.Printf("Failed to terminate Ollama process: %v\n", err)
			}
		} else {
			fmt.Printf("Error executing pkill command: %v\n", err)
		}
	} else {
		fmt.Println("Ollama process has been terminated.")
	}

	fmt.Println("Waiting for 2 seconds before Ollama restarts...")
	time.Sleep(2 * time.Second)

	fmt.Println("Ollama will restart automatically after being stopped.")
	return nil
}


