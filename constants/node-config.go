
package constants

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/shirou/gopsutil/v3/mem"
)

func test() {
	// Query system for total RAM
	v, err := mem.VirtualMemory()
	if err != nil {
		fmt.Printf("Error getting system memory: %v\n", err)
		return
	}
	ramGB := int(v.Total / (1024 * 1024 * 1024))

	envVars := map[string]map[string]string{
		"OLLAMA_MAX_LOADED_MODELS": {
			"value":       "3",
			"description": "Maximum number of loaded models per GPU",
		},
		"OLLAMA_NUM_PARALLEL": {
			"value":       strconv.Itoa(ramGB/2),
			"description": "Maximum number of parallel requests",
		},
	}

	for key, data := range envVars {
		os.Setenv(key, data["value"])
		fmt.Printf("Set %s=%s - %s\n", key, data["value"], data["description"])
	}

	fmt.Println("\nOllama configuration complete.")
	fmt.Printf("System RAM: %d GB\n", ramGB)
	fmt.Printf("OLLAMA_NUM_PARALLEL: %d\n", ramGB/2)
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
}
