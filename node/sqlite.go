
package node

import (
	"database/sql"
	"log"
	"os"
	"fmt"
	"path/filepath"
	_ "github.com/mattn/go-sqlite3"
)

// CreateErrorDatabase creates a new SQLite database with the required fields
func CreateErrorDatabase(dbPath string) error {
	// Ensure the database directory exists
	dbDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return err
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	// Create table with the specified fields
	createTableSQL := `CREATE TABLE IF NOT EXISTS errors (
		errorID INTEGER PRIMARY KEY NOT NULL,
		time TEXT NOT NULL,
		errorType TEXT NOT NULL,
		signature TEXT NOT NULL,
		message TEXT NOT NULL,
		stackTrace TEXT NOT NULL
	);`

	_, err = db.Exec(createTableSQL)
	if err != nil {
		return err
	}

	// Test the connection
	err = db.Ping()
	if err != nil {
		return err
	}

	return nil
}

// InsertError inserts error data into the database
func InsertError(dbPath string, errorID string, timestamp string, errorType string, signature string, message string, stackTrace string) error {
	log.Printf("Attempting to insert error into database at: %s", dbPath)
	
	// Ensure the database exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		log.Printf("Database does not exist, creating at: %s", dbPath)
		if err := CreateErrorDatabase(dbPath); err != nil {
			log.Printf("Failed to create database: %v", err)
			return err
		}
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Printf("Failed to open database: %v", err)
		return err
	}
	defer db.Close()

	// Convert errorID to integer
	var errorIDInt int
	_, err = fmt.Sscanf(errorID, "%d", &errorIDInt)
	if err != nil {
		log.Printf("Failed to convert errorID to integer: %v", err)
		return err
	}

	// Begin transaction
	tx, err := db.Begin()
	if err != nil {
		log.Printf("Failed to begin transaction: %v", err)
		return err
	}

	// Prepare the statement
	stmt, err := tx.Prepare(`INSERT INTO errors (errorID, time, errorType, signature, message, stackTrace) VALUES (?, ?, ?, ?, ?, ?)`)
	if err != nil {
		tx.Rollback()
		log.Printf("Failed to prepare statement: %v", err)
		return err
	}
	defer stmt.Close()

	// Execute the statement
	result, err := stmt.Exec(errorIDInt, timestamp, errorType, signature, message, stackTrace)
	if err != nil {
		tx.Rollback()
		log.Printf("Failed to execute statement: %v", err)
		return err
	}

	// Verify the insertion
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		tx.Rollback()
		log.Printf("Failed to get rows affected: %v", err)
		return err
	}

	if rowsAffected != 1 {
		tx.Rollback()
		log.Printf("Expected 1 row to be affected, got %d", rowsAffected)
		return fmt.Errorf("unexpected number of rows affected: %d", rowsAffected)
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		log.Printf("Failed to commit transaction: %v", err)
		return err
	}

	log.Printf("Successfully inserted error into database")
	return nil
}
