#!/bin/bash

# Set PATH to include Go
export PATH="/opt/homebrew/bin:$PATH"

# Function to find the AI-cluster directory
find_ai_cluster_dir() {
    local current_dir=$(pwd)
    while [[ $current_dir != "/" ]]; do
        if [[ $(basename "$current_dir") == "AI-cluster" ]]; then
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

# Change to the AI-cluster directory
cd "$ai_cluster_dir" || {
    echo "Error: Unable to change to AI-cluster directory"
    exit 1
}

# Print debugging information
echo "Current directory: $(pwd)"
echo "Go version: $(go version)"
echo "Go path: $(which go)"

# Run go build
if ! go build .; then
    echo "Error: Failed to run go build"
    exit 1
fi

echo "Build completed successfully"