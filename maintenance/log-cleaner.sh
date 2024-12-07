#!/bin/bash

# Configuration
DB_PATH="$HOME/AI-cluster/node/errors.db"
ROOT_DIR="$HOME/AI-cluster/"

# List of log files to clear instead of process
CLEAR_LOGS=(
    "$HOME/AI-cluster/nats_server/flag.log"
    "$HOME/AI-cluster/nats_server/nats-server.log"
    "$HOME/AI-cluster/setup/check_update.log"
    "$HOME/AI-cluster/setup/start_node.log"
    "$HOME/AI-cluster/setup/stop_node.log"
)

# Function to check if file should be cleared
should_clear_file() {
    local file="$1"
    for clear_file in "${CLEAR_LOGS[@]}"; do
        if [ "$(realpath "$file")" = "$(realpath "$clear_file")" ]; then
            return 0
        fi
    done
    return 1
}

# Function to decode base64
decode_base64() {
    echo "$1" | base64 -d
}

# Function to check if error exists in database
error_exists_in_db() {
    local error_id=$1
    sqlite3 "$DB_PATH" "SELECT EXISTS(SELECT 1 FROM errors WHERE errorID = $error_id);" | grep -q 1
    return $?
}

# Function to parse and process a log entry
parse_log_entry() {
    local line="$1"
    
    # Extract information using regex
    if [[ $line =~ \*\[ErrorID:([0-9]+)\]\[([^\]]+)\]\[([^\]]+)\]\[([^:]+):([^:]+):([^:]+):([0-9]+)\]\[([^\]]+)\]\[Stack\ Trace\ \(Base64\):([^\]]+)\]\* ]]; then
        error_id="${BASH_REMATCH[1]}"
        timestamp="${BASH_REMATCH[2]}"
        level="${BASH_REMATCH[3]}"
        ip="${BASH_REMATCH[4]}"
        filepath="${BASH_REMATCH[5]}"
        function="${BASH_REMATCH[6]}"
        line_num="${BASH_REMATCH[7]}"
        message="${BASH_REMATCH[8]}"
        stack_trace=$(decode_base64 "${BASH_REMATCH[9]}")
        
        # Print parsed entry
        echo "Log Entry Details:"
        echo "ErrorID: $error_id"
        echo "Timestamp: $timestamp"
        echo "Level: $level"
        echo "IP: $ip"
        echo "FilePath: $filepath"
        echo "Function: $function"
        echo "Line: $line_num"
        echo "Message: $message"
        echo "StackTrace: $stack_trace"
        
        # Check if error exists in database
        if error_exists_in_db "$error_id"; then
            return 1
        fi
    fi
    return 0
}

# Function to process a single log file
process_log_file() {
    local file="$1"
    local temp_file="${file}.tmp"
    
    echo "Processing: $file"
    
    # Check if file should be cleared instead of processed
    if should_clear_file "$file"; then
        echo "Clearing file: $file"
        : > "$file"  # This erases the contents of the file
        return
    fi
    
    # Process file line by line
    while IFS= read -r line; do
        if [[ $line == \** && $line == *\* ]]; then
            if parse_log_entry "$line"; then
                echo "$line" >> "$temp_file"
            fi
        else
            echo "$line" >> "$temp_file"
        fi
    done < "$file"
    
    # Replace original file with processed content
    mv "$temp_file" "$file"
}

# Main script
main() {
    # Check if sqlite3 is installed
    if ! command -v sqlite3 &> /dev/null; then
        echo "Error: sqlite3 is required but not installed"
        exit 1
    fi  # Added missing closing brace here
    
    # Process all .log files recursively
    find "$ROOT_DIR" -type f -name "*.log" | while IFS= read -r file; do
        process_log_file "$file"
    done
    
    echo "Log cleaning completed successfully"
}

# Run main function
main
