
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
)

type Request struct {
	ID struct {
		Wallet    string `json:"wallet"`
		RequestID string `json:"request_id"`
	} `json:"id"`
	Type     string `json:"type"`
	RunTimes int    `json:"runTimes"`
	Synth    bool   `json:"synth"`
	Payload  struct {
		Model   string `json:"model"`
		Prompt  string `json:"prompt"`
		Stream  bool   `json:"stream"`
		System  string `json:"system"`
		Options struct {
			NumCtx      int     `json:"num_ctx"`
			Temperature float64 `json:"temperature"`
		} `json:"options"`
	} `json:"payload"`
}

type Task struct {
	Type         string  `json:"type"`
	ModelName    string  `json:"model_name"`
	SystemPrompt string  `json:"system_prompt"`
	Prompt       string  `json:"prompt"`
	OutputDir    string  `json:"output_dir"`
	Temperature  float64 `json:"temperature"`
	ID           struct {
		Wallet    string `json:"wallet"`
		RequestID string `json:"request_id"`
	} `json:"id"`
	ResponseNum int `json:"response_num"`
}

type Result struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Result  string `json:"result,omitempty"`
}

type Processor struct {
	taskQueue   chan []byte
	resultQueue chan Result
	workerPool  chan struct{}
	wg          sync.WaitGroup
	nc          *nats.Conn
}

func NewProcessor(numWorkers int, natsURL string) *Processor {
	nc, err := nats.Connect(natsURL)
	if err != nil {
		log.Fatalf("Failed to connect to NATS: %v", err)
	}

	p := &Processor{
		taskQueue:   make(chan []byte),
		resultQueue: make(chan Result),
		workerPool:  make(chan struct{}, numWorkers),
		nc:          nc,
	}

	for i := 0; i < numWorkers; i++ {
		go p.worker(i)
	}

	go p.resultsCollector()

	return p
}

func (p *Processor) AddTask(taskData []byte) {
	p.wg.Add(1)
	p.taskQueue <- taskData
}

func (p *Processor) worker(workerID int) {
	for taskData := range p.taskQueue {
		p.workerPool <- struct{}{}
		func() {
			defer func() {
				<-p.workerPool
				p.wg.Done()
			}()

			var req Request
			if err := json.Unmarshal(taskData, &req); err != nil {
				log.Printf("Error unmarshaling task: %v", err)
				return
			}

			log.Printf("Worker %d started processing task for request ID: %s", workerID, req.ID.RequestID)

			requestIDParts := strings.SplitN(req.ID.RequestID, ":", 2)
			baseRequestID := requestIDParts[0]

			task := Task{
				Type:         req.Type,
				ModelName:    req.Payload.Model,
				SystemPrompt: req.Payload.System,
				Prompt:       req.Payload.Prompt,
				OutputDir:    filepath.Join("output", baseRequestID),
				Temperature:  req.Payload.Options.Temperature,
				ID:           req.ID,
				ResponseNum:  1,
			}

			startTime := time.Now()
			result := p.executeTask(task)
			duration := time.Since(startTime)

			log.Printf("Worker %d completed task for request ID: %s in %v", workerID, req.ID.RequestID, duration)

			p.resultQueue <- result
		}()
	}
}

func (p *Processor) executeTask(task Task) Result {
	err := os.MkdirAll(task.OutputDir, 0755)
	if err != nil {
		return Result{
			Success: false,
			Message: fmt.Sprintf("Error creating output directory: %v", err),
		}
	}

	ollamaRequest := map[string]interface{}{
		"model":  task.ModelName,
		"prompt": task.Prompt,
		"system": task.SystemPrompt,
		"stream": false,
		"options": map[string]interface{}{
			"temperature": task.Temperature,
		},
	}

	jsonData, err := json.Marshal(ollamaRequest)
	if err != nil {
		return Result{
			Success: false,
			Message: fmt.Sprintf("Error marshaling Ollama request: %v", err),
		}
	}

	resp, err := http.Post("http://localhost:11434/api/generate", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return Result{
			Success: false,
			Message: fmt.Sprintf("Error sending request to Ollama API: %v", err),
		}
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return Result{
			Success: false,
			Message: fmt.Sprintf("Error reading Ollama API response: %v", err),
		}
	}

	log.Printf("Ollama API response: %s", string(body))

	var ollamaResponse map[string]interface{}
	err = json.Unmarshal(body, &ollamaResponse)
	if err != nil {
		return Result{
			Success: false,
			Message: fmt.Sprintf("Error unmarshaling Ollama API response: %v", err),
		}
	}

	if ollamaResponse == nil {
		return Result{
			Success: false,
			Message: "Ollama API response is nil",
		}
	}

	responseStr, ok := ollamaResponse["response"].(string)
	if !ok {
		return Result{
			Success: false,
			Message: fmt.Sprintf("Ollama API response does not contain 'response' field or it's not a string. Full response: %v", ollamaResponse),
		}
	}

	result := Result{
		Success: true,
		Message: fmt.Sprintf("Processed task for model: %s", task.ModelName),
		Result:  responseStr,
	}

	// Send result to proxyOut via NATS
	resultJSON, err := json.Marshal(map[string]interface{}{
		"id":     task.ID,
		"result": result,
	})
	if err != nil {
		log.Printf("Error marshaling result for NATS: %v", err)
	} else {
		subject := fmt.Sprintf("outgoing.responses.%s", task.Type)
		err = p.nc.Publish(subject, resultJSON)
		if err != nil {
			log.Printf("Error publishing result to NATS: %v", err)
		}
	}

	// Write result to output directory
	messageParts := strings.SplitN(task.ID.RequestID, ":", 2)
	messageNumber := "1"
	if len(messageParts) > 1 {
		messageNumber = messageParts[1]
	}

	outputFile := filepath.Join(task.OutputDir, fmt.Sprintf("%s.json", messageNumber))
	err = os.WriteFile(outputFile, []byte(result.Result), 0644)
	if err != nil {
		result.Success = false
		result.Message = fmt.Sprintf("Error writing result to file: %v", err)
	}

	return result
}

func (p *Processor) resultsCollector() {
	for result := range p.resultQueue {
		if result.Success {
			log.Printf("Task completed successfully: %s", result.Message)
		} else {
			log.Printf("Task failed: %s", result.Message)
		}
	}
}

func (p *Processor) Wait() {
	close(p.taskQueue)
	p.wg.Wait()
	close(p.resultQueue)
	p.nc.Close()
}