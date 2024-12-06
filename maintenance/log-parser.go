
package main

import (
	"bufio"
	"database/sql"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/mtmox/AI-cluster/constants"
)

type LogEntry struct {
	ErrorID    int
	Timestamp  time.Time
	Level      string
	IP         string
	FilePath   string
	Function   string
	Line       int
	Message    string
	StackTrace string
}

// Add a method to print LogEntry
func (l *LogEntry) Print() {
	fmt.Printf("Log Entry Details:\n"+
		"ErrorID: %d\n"+
		"Timestamp: %v\n"+
		"Level: %s\n"+
		"IP: %s\n"+
		"FilePath: %s\n"+
		"Function: %s\n"+
		"Line: %d\n"+
		"Message: %s\n"+
		"StackTrace: %s\n",
		l.ErrorID,
		l.Timestamp,
		l.Level,
		l.IP,
		l.FilePath,
		l.Function,
		l.Line,
		l.Message,
		l.StackTrace)
}

var logPattern = regexp.MustCompile(`\*\[ErrorID:(\d+)\]\[([^\]]+)\]\[([^\]]+)\]\[([^:]+):([^:]+):([^:]+):(\d+)\]\[([^\]]+)\]\[Stack Trace \(Base64\):([^\]]+)\]\*`)

func parseLogEntry(line string) (*LogEntry, error) {
	matches := logPattern.FindStringSubmatch(line)
	if matches == nil {
		return nil, fmt.Errorf("invalid log format")
	}

	errorID := 0
	fmt.Sscanf(matches[1], "%d", &errorID)

	timestamp, err := time.Parse("2006-01-02 15:04:05", matches[2])
	if err != nil {
		return nil, fmt.Errorf("failed to parse timestamp: %v", err)
	}

	lineNum := 0
	fmt.Sscanf(matches[7], "%d", &lineNum)

	stackTrace, err := base64.StdEncoding.DecodeString(matches[9])
	if err != nil {
		return nil, fmt.Errorf("failed to decode stack trace: %v", err)
	}

	entry := &LogEntry{
		ErrorID:    errorID,
		Timestamp:  timestamp,
		Level:      matches[3],
		IP:         matches[4],
		FilePath:   matches[5],
		Function:   matches[6],
		Line:       lineNum,
		Message:    matches[8],
		StackTrace: string(stackTrace),
	}
	
	// Print the entry after parsing
	entry.Print()
	
	return entry, nil
}

func processLogFile(db *sql.DB, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	tempFile, err := os.CreateTemp(filepath.Dir(filePath), "temp_*.log")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	scanner := bufio.NewScanner(file)
	writer := bufio.NewWriter(tempFile)

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "*") && strings.HasSuffix(line, "*") {
			entry, err := parseLogEntry(line)
			if err != nil {
				// Write invalid entries to output
				writer.WriteString(line + "\n")
				continue
			}

			// Check if error exists in database
			var exists bool
			err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM errors WHERE errorID = ?)", entry.ErrorID).Scan(&exists)
			if err != nil {
				return fmt.Errorf("database query failed: %v", err)
			}

			if !exists {
				writer.WriteString(line + "\n")
			}
		} else {
			writer.WriteString(line + "\n")
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scanner error: %v", err)
	}

	writer.Flush()

	// Replace original file with processed content
	if err := os.Rename(tempFile.Name(), filePath); err != nil {
		return fmt.Errorf("failed to replace original file: %v", err)
	}

	return nil
}

func main() {
	dbPath := constants.ErrorDatabase
	searchPath := constants.RootDirectory

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		fmt.Printf("Failed to open database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	err = filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(path, ".log") {
			fmt.Printf("Processing: %s\n", path)
			if err := processLogFile(db, path); err != nil {
				fmt.Printf("Error processing %s: %v\n", path, err)
				return nil
			}
		}
		return nil
	})

	if err != nil {
		fmt.Printf("Error walking directory: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Log cleaning completed successfully")
}
