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
cd "$ai_cluster_dir"
"$GO_BINARY" "-$INSTANCE_TYPE"