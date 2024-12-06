
package node

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/mtmox/AI-cluster/constants"
)

// ErrorLevel represents the severity of the error
type ErrorLevel int

const (
	INFO ErrorLevel = iota
	WARNING
	ERROR
	FATAL
)

// CustomError represents our standardized error type
type CustomError struct {
	ErrorID   int
	Level     ErrorLevel
	Message   string
	Err       error
	Timestamp time.Time
	File      string
	Function  string
	Line      int
	Signature string
	Stack     []byte
}

// ErrorCounter holds the current error ID counter
type ErrorCounter struct {
	Increment int `json:"increment"`
}

var (
	// Logger instances
	fileLogger  *log.Logger
	stdLogger   *log.Logger
	initialized bool
	computerID  string // Store computer identification
	
	// Error ID management
	currentErrorID int
	errorIDMutex   sync.Mutex
	counterFile    = filepath.Join(os.ExpandEnv("$HOME"), "AI-cluster", "node", "error_counter.json")
)

func init() {
	// Ensure the directory exists
	dir := filepath.Dir(counterFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Printf("Failed to create directory for error counter: %v", err)
	}

	// Load the initial error counter
	if err := loadErrorCounter(); err != nil {
		log.Printf("Failed to load error counter: %v", err)
	}

	// Initialize the error database
	if err := CreateErrorDatabase(constants.ErrorDatabase); err != nil {
		log.Printf("Failed to create error database: %v", err)
	}
}

// loadErrorCounter loads the current error counter from JSON file
func loadErrorCounter() error {
	if _, err := os.Stat(counterFile); os.IsNotExist(err) {
		// If file doesn't exist, create it with initial value
		counter := ErrorCounter{Increment: 0}
		currentErrorID = 0
		return saveErrorCounter(counter)
	}

	data, err := os.ReadFile(counterFile)
	if err != nil {
		return fmt.Errorf("failed to read counter file: %v", err)
	}

	var counter ErrorCounter
	if err := json.Unmarshal(data, &counter); err != nil {
		return fmt.Errorf("failed to unmarshal counter: %v", err)
	}

	currentErrorID = counter.Increment
	return nil
}

// saveErrorCounter saves the current error counter to JSON file
func saveErrorCounter(counter ErrorCounter) error {
	data, err := json.Marshal(counter)
	if err != nil {
		return fmt.Errorf("failed to marshal counter: %v", err)
	}

	return os.WriteFile(counterFile, data, 0644)
}

// getNextErrorID generates the next error ID
func getNextErrorID() int {
	errorIDMutex.Lock()
	defer errorIDMutex.Unlock()

	currentErrorID++
	
	// Save the new counter value
	counter := ErrorCounter{Increment: currentErrorID}
	if err := saveErrorCounter(counter); err != nil {
		log.Printf("Failed to save error counter: %v", err)
	}

	return currentErrorID
}

// createFileLogger creates a log file at the same path as the source file
func createFileLogger(sourcePath string) (*log.Logger, error) {
	dir := filepath.Dir(sourcePath)
	baseFileName := filepath.Base(sourcePath)
	logFileName := strings.TrimSuffix(baseFileName, filepath.Ext(baseFileName)) + ".log"
	logFilePath := filepath.Join(dir, logFileName)

	file, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return nil, fmt.Errorf("failed to create log file: %v", err)
	}

	return log.New(file, "", 0), nil
}

// Initialize sets up the logging system
func Initialize(logFilePath string, machineID string) error {
	if initialized {
		return nil
	}

	computerID = machineID

	file, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return fmt.Errorf("failed to open log file: %v", err)
	}

	fileLogger = log.New(file, "", 0)
	stdLogger = log.New(os.Stdout, "", 0)
	initialized = true

	return nil
}

// generateSignature creates a unique signature for the error
func generateSignature(file, function string, line int) string {
	return fmt.Sprintf("%s:%s:%s:%d", GetIPWithoutDots(), file, function, line)
}

// getFunctionName returns the name of the function where the error occurred
func getFunctionName(skip int) string {
	pc, _, _, ok := runtime.Caller(skip)
	if !ok {
		return "unknown"
	}
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return "unknown"
	}
	parts := strings.Split(fn.Name(), ".")
	return parts[len(parts)-1]
}

// getCallerInfo returns file, function name, and line number for error tracking
func getCallerInfo(skip int) (string, string, int) {
	pc, file, line, ok := runtime.Caller(skip)
	if !ok {
		return "unknown", "unknown", 0
	}
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return file, "unknown", line
	}
	parts := strings.Split(fn.Name(), ".")
	return file, parts[len(parts)-1], line
}

// New creates a new CustomError
func New(level ErrorLevel, message string, err error) *CustomError {
	file, function, line := getCallerInfo(2)
	
	return &CustomError{
		ErrorID:   getNextErrorID(),
		Level:     level,
		Message:   message,
		Err:       err,
		Timestamp: time.Now(),
		File:      file,
		Function:  function,
		Line:      line,
		Signature: generateSignature(file, function, line),
		Stack:     debug.Stack(),
	}
}

// HandleError is a convenience function for handling errors inline
func HandleError(err error, level ErrorLevel, message string) {
	if err != nil {
		file, function, line := getCallerInfo(2)
		
		customErr := &CustomError{
			ErrorID:   getNextErrorID(),
			Level:     level,
			Message:   message,
			Err:       err,
			Timestamp: time.Now(),
			File:      file,
			Function:  function,
			Line:      line,
			Signature: generateSignature(file, function, line),
			Stack:     debug.Stack(),
		}
		customErr.Log()
	}
}

// Error implements the error interface
func (e *CustomError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Log logs the error to both file and stdout
func (e *CustomError) Log() {
	sourceLogger, err := createFileLogger(e.File)
	if err != nil {
		log.Printf("Failed to create source logger: %v", err)
	}

	// Encode stack trace to base64
	encodedStack := base64.StdEncoding.EncodeToString(e.Stack)

	// Format the log message with error details, timestamp, and encoded stack trace
	logMessage := fmt.Sprintf("\n*[ErrorID:%d][%s][%s][%s][%s][Stack Trace (Base64):%s]*", 
		e.ErrorID,
		e.Timestamp.Format("2006-01-02 15:04:05"),
		e.Level.String(), 
		e.Signature, 
		e.Error(),
		encodedStack)
	
	// Log to database
	err = InsertError(
		constants.ErrorDatabase,
		fmt.Sprintf("%d", e.ErrorID),
		e.Timestamp.Format("2006-01-02 15:04:05"),
		e.Level.String(),
		e.Signature,
		e.Error(),
		encodedStack,
	)
	if err != nil {
		log.Printf("Failed to insert error into database: %v", err)
	}
		
	if !initialized {
		log.Printf("%s", logMessage)
		if sourceLogger != nil {
			sourceLogger.Printf("%s", logMessage)
		}
		return
	}

	fileLogger.Printf("%s", logMessage)
	stdLogger.Printf("%s", logMessage)
	if sourceLogger != nil {
		sourceLogger.Printf("%s", logMessage)
	}

	

	if e.Level == FATAL {
		os.Exit(1)
	}
}

// String converts ErrorLevel to string
func (l ErrorLevel) String() string {
	switch l {
	case INFO:
		return "INFO"
	case WARNING:
		return "WARNING"
	case ERROR:
		return "ERROR"
	case FATAL:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}
