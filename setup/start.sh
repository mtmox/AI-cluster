#!/bin/bash

# Define the path to the GO binary
GO_BINARY="$HOME/AI-cluster/AI-cluster"

# Check if the GO binary exists
if [ ! -f "$GO_BINARY" ]; then
    echo "Error: GO binary not found at $GO_BINARY"
    exit 1
fi

# Check if the GO binary is executable
if [ ! -x "$GO_BINARY" ]; then
    echo "Error: GO binary is not executable. Setting executable permission."
    chmod +x "$GO_BINARY"
fi

# Start the GO binary
echo "Starting GO binary..."
"$GO_BINARY"