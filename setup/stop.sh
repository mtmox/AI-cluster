#!/bin/bash

# Set PATH to include necessary binaries
export PATH="/opt/homebrew/bin:$PATH"

# Function to find the AI-cluster directory
find_ai_cluster_dir() {
    # First try the HOME directory path
    if [ -d "$HOME/AI-cluster" ]; then
        echo "$HOME/AI-cluster"
        return 0
    fi

    # If not found in HOME, try to find it from current directory upwards
    local current_dir=$(pwd)
    while [[ $current_dir != "/" ]]; do
        if [[ -d "$current_dir/AI-cluster" ]]; then
            echo "$current_dir/AI-cluster"
            return 0
        elif [[ $(basename "$current_dir") == "AI-cluster" ]]; then
            echo "$current_dir"
            return 0
        fi
        current_dir=$(dirname "$current_dir")
    done
    return 1
}

# Find the AI-cluster directory
ai_cluster_dir=$(find_ai_cluster_dir)

if [[ -z "$ai_cluster_dir" ]]; then
    echo "Error: AI-cluster directory not found"
    exit 1
fi

# Define the path to the GO binary
GO_BINARY="$ai_cluster_dir/AI-cluster"

# Check if the GO binary exists
if [ ! -f "$GO_BINARY" ]; then
    echo "Error: GO binary not found at $GO_BINARY"
    exit 1
fi

# Find the process IDs of the running GO binary instances
PIDS=$(pgrep -f "$GO_BINARY")

if [ -z "$PIDS" ]; then
    echo "No GO binary instances are currently running."
    exit 0
fi

# Attempt to stop the GO binary instances gracefully
echo "Stopping GO binary instances..."
for PID in $PIDS; do
    echo "Stopping instance with PID: $PID"
    kill "$PID"
done

# Wait for up to 10 seconds for the processes to terminate
for i in {1..10}; do
    if ! pgrep -f "$GO_BINARY" > /dev/null; then
        echo "All GO binary instances stopped successfully."
        exit 0
    fi
    sleep 1
done

# If any processes are still running, force kill them
REMAINING_PIDS=$(pgrep -f "$GO_BINARY")
if [ -n "$REMAINING_PIDS" ]; then
    echo "Some GO binary instances did not stop gracefully. Forcing termination..."
    for PID in $REMAINING_PIDS; do
        echo "Forcefully terminating instance with PID: $PID"
        kill -9 "$PID"
    done
    echo "All GO binary instances forcefully terminated."
else
    echo "All GO binary instances stopped successfully."
fi