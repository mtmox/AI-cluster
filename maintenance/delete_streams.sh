#!/bin/bash

# Check if nats CLI tool is installed
if ! command -v nats &> /dev/null; then
    echo "Error: nats CLI tool is not installed. Please install it first."
    echo "Visit: https://github.com/nats-io/natscli#installation"
    exit 1
fi

# NATS server URL - modify if needed
NATS_URL="nats://192.168.1.140:4222"

# Get list of all streams
echo "Fetching streams..."
STREAMS=$(nats stream list --server="$NATS_URL" --names 2>/dev/null)

if [ $? -ne 0 ]; then
    echo "Error connecting to NATS server at $NATS_URL"
    exit 1
fi

if [ -z "$STREAMS" ]; then
    echo "No streams found."
    exit 0
fi

# Count number of streams
STREAM_COUNT=$(echo "$STREAMS" | wc -l)
echo "Found $STREAM_COUNT stream(s). Deleting..."

# Delete each stream
while IFS= read -r stream; do
    if [ ! -z "$stream" ]; then
        echo "Deleting stream: $stream"
        nats stream rm "$stream" --server="$NATS_URL" --force
        if [ $? -eq 0 ]; then
            echo "Deleted stream: $stream"
        else
            echo "Error deleting stream: $stream"
        fi
    fi
done <<< "$STREAMS"

echo "All streams have been deleted."
