
package backend

import (
	"encoding/json"
	"os"
	"os/user"
	"path/filepath"
	"sync"
	"time"
)

type NodeSettings struct {
	MaxParallelRequests int `json:"max_parallel_requests"`
	MessageDelayMS      int `json:"message_delay_ms"`
}

var (
	settings     NodeSettings
	settingsLock sync.RWMutex
	lastMessage  time.Time
	activeTasks  int
	tasksLock    sync.Mutex
)

func getConfigPath() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	return filepath.Join(usr.HomeDir, "AI-cluster", "backend", "node-config.json"), nil
}

func SaveNodeSettings(maxParallel int) error {
	settings := NodeSettings{
		MaxParallelRequests: maxParallel,
		MessageDelayMS:      500,
	}

	data, err := json.Marshal(settings)
	if err != nil {
		return err
	}

	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	// Create the directory structure if it doesn't exist
	err = os.MkdirAll(filepath.Dir(configPath), 0755)
	if err != nil {
		return err
	}

	// Create the file if it doesn't exist, or truncate it if it does
	file, err := os.OpenFile(configPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write the data to the file
	_, err = file.Write(data)
	return err
}

func LoadNodeSettings() error {
	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	settingsLock.Lock()
	defer settingsLock.Unlock()

	return json.Unmarshal(data, &settings)
}

func CanProcessMessage() bool {
	tasksLock.Lock()
	defer tasksLock.Unlock()

	settingsLock.RLock()
	defer settingsLock.RUnlock()

	now := time.Now()
	if now.Sub(lastMessage) < time.Duration(settings.MessageDelayMS)*time.Millisecond {
		return false
	}

	if activeTasks >= settings.MaxParallelRequests {
		return false
	}

	lastMessage = now
	activeTasks++
	return true
}

func FinishProcessing() {
	tasksLock.Lock()
	defer tasksLock.Unlock()
	
	if activeTasks > 0 {
		activeTasks--
	}
}
