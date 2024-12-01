
package backend

import (
	"encoding/json"
	"os"
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

func SaveNodeSettings(maxParallel int) error {
	settings := NodeSettings{
		MaxParallelRequests: maxParallel,
		MessageDelayMS:      500,
	}

	data, err := json.Marshal(settings)
	if err != nil {
		return err
	}

	return os.WriteFile("node-config.json", data, 0644)
}

func LoadNodeSettings() error {
	data, err := os.ReadFile("node-config.json")
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