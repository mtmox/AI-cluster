
package node

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
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
	Level     ErrorLevel
	Message   string
	Err       error
	Timestamp time.Time
	File      string
	Function  string
	Line      int
	Signature string
}

var (
	// Logger instances
	fileLogger  *log.Logger
	stdLogger   *log.Logger
	initialized bool
	computerID  string // Store computer identification
)

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

	return log.New(file, "", log.Ldate|log.Ltime), nil
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

	fileLogger = log.New(file, "", log.Ldate|log.Ltime)
	stdLogger = log.New(os.Stdout, "", log.Ldate|log.Ltime)
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

// New creates a new CustomError
func New(level ErrorLevel, message string, err error) *CustomError {
	_, file, line, _ := runtime.Caller(1)
	function := getFunctionName(1)
	
	return &CustomError{
		Level:     level,
		Message:   message,
		Err:       err,
		Timestamp: time.Now(),
		File:      file,
		Function:  function,
		Line:      line,
		Signature: generateSignature(file, function, line),
	}
}

// HandleError is a convenience function for handling errors inline
func HandleError(err error, level ErrorLevel, message string) {
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		function := getFunctionName(1)
		
		customErr := &CustomError{
			Level:     level,
			Message:   message,
			Err:       err,
			Timestamp: time.Now(),
			File:      file,
			Function:  function,
			Line:      line,
			Signature: generateSignature(file, function, line),
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

	// New format with brackets and simplified signature
	logMessage := fmt.Sprintf("*[%s]:[%s]* *[%s]*", 
		e.Level.String(), e.Signature, e.Error())
	
	if !initialized {
		log.Printf("%s\n", logMessage)
		if sourceLogger != nil {
			sourceLogger.Println(logMessage)
		}
		return
	}

	fileLogger.Println(logMessage)
	stdLogger.Println(logMessage)
	if sourceLogger != nil {
		sourceLogger.Println(logMessage)
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
