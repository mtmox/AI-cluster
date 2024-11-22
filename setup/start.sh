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

# Set default instance type to backend
INSTANCE_TYPE="backend"

# Check for the instance type flag
if [ "$1" == "-frontend" ]; then
    INSTANCE_TYPE="frontend"
elif [ "$1" == "-backend" ] || [ -z "$1" ]; then
    INSTANCE_TYPE="backend"
else
    echo "Error: Invalid argument. Use -frontend for frontend instance, or no argument (or -backend) for backend instance."
    echo "Usage: $0 [-frontend|-backend]"
    exit 1
fi

# Start the GO binary with the appropriate flag
echo "Starting GO binary as $INSTANCE_TYPE instance..."
"$GO_BINARY" "-$INSTANCE_TYPE"
