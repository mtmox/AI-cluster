#!/bin/bash

# Set PATH to include Go
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

# Print current directory for debugging
echo "Starting directory: $(pwd)"

# Find the AI-cluster directory
ai_cluster_dir=$(find_ai_cluster_dir)

if [[ -z "$ai_cluster_dir" ]]; then
    echo "Error: AI-cluster directory not found"
    echo "Searched in:"
    echo "- $HOME/AI-cluster"
    echo "- Current directory and its parents: $(pwd)"
    exit 1
fi

echo "Found AI-cluster directory at: $ai_cluster_dir"

# Change to the AI-cluster directory
cd "$ai_cluster_dir" || {
    echo "Error: Unable to change to AI-cluster directory at $ai_cluster_dir"
    exit 1
}

# Print debugging information
echo "Current directory: $(pwd)"
echo "Go version: $(go version)"
echo "Go path: $(which go)"

# Check if go.mod exists
if [ ! -f "go.mod" ]; then
    echo "Error: go.mod not found in $(pwd)"
    echo "Contents of directory:"
    ls -la
    exit 1
fi

# Run go build
if ! go build .; then
    echo "Error: Failed to run go build"
    echo "Contents of directory:"
    ls -la
    exit 1
fi

echo "Build completed successfully"