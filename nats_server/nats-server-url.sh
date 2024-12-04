#!/bin/bash

# Define the source and output files
SOURCE_FILE="$HOME/AI-cluster/constants/constants.go"
OUTPUT_FILE="$HOME/AI-cluster/nats_server/nats-server-url.json"

# Check if the source file exists
if [ ! -f "$SOURCE_FILE" ]; then
    echo "Error: File '$SOURCE_FILE' not found"
    exit 1
fi

# Use grep and sed to extract the NatsURL value
# This looks for lines containing "NatsURL" and extracts the string value between quotes
nats_url=$(grep "NatsURL.*=.*\".*\"" "$SOURCE_FILE" | sed -E 's/.*"(.*)".*/\1/')

# Check if we found a value
if [ -z "$nats_url" ]; then
    echo "Error: NatsURL not found in the file"
    exit 1
fi

# Create JSON output
json_output="{\"nats_url\": \"$nats_url\"}"

# Write to the output file
echo "$json_output" > "$OUTPUT_FILE"

echo "NATS URL has been written to $OUTPUT_FILE"
