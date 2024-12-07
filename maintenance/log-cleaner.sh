#!/bin/bash

# Configuration
DB_PATH="$HOME/AI-cluster/node/errors.db"
ROOT_DIR="$HOME/AI-cluster/"

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
    fi  # Added missing 'fi'
    
    # Process all .log files recursively
    while IFS= read -r -d '' file; do
        process_log_file "$file"
    done < <(find "$ROOT_DIR" -type f -name "*.log" -print0)
    
    echo "Log cleaning completed successfully"
}

# Run main function
main
