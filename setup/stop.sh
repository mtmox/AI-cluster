#!/bin/bash

# Define the path to the GO binary
GO_BINARY="$HOME/AI-cluster/AI-cluster"

# Check if the GO binary exists
if [ ! -f "$GO_BINARY" ]; then
    echo "Error: GO binary not found at $HOME/AI-cluster/AI-cluster"
    exit 1
fi

# Find the process ID of the running GO binary
PID=$(pgrep -f "$GO_BINARY")

if [ -z "$PID" ]; then
    echo "GO binary is not currently running."
    exit 0
fi

# Attempt to stop the GO binary gracefully
echo "Stopping GO binary (PID: $PID)..."
kill "$PID"

# Wait for up to 10 seconds for the process to terminate
for i in {1..10}; do
    if ! ps -p "$PID" > /dev/null; then
        echo "GO binary stopped successfully."
        exit 0
    fi
    sleep 1
done

# If the process is still running, force kill it
if ps -p "$PID" > /dev/null; then
    echo "GO binary did not stop gracefully. Forcing termination..."
    kill -9 "$PID"
    echo "GO binary forcefully terminated."
else
    echo "GO binary stopped successfully."
fi
