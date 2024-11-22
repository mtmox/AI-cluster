#!/bin/bash

# Define the path to the GO binary
GO_BINARY="$HOME/AI-cluster/AI-cluster"

# Check if the GO binary exists
if [ ! -f "$GO_BINARY" ]; then
    echo "Error: GO binary not found at $HOME/AI-cluster/AI-cluster"
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
