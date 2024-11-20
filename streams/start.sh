#!/bin/bash

# Function to check the last command's status
check_status() {
    if [ $? -ne 0 ]; then
        echo "Error: $1 failed"
        exit 1
    fi
}

# Check if NATS server is already running
if pgrep -x "nats-server" > /dev/null
then
    echo "NATS server is already running."
else
    # Start NATS server with configuration file using absolute path
    echo "Starting NATS server..."
    /opt/homebrew/bin/nats-server -c "$HOME/moai/messageStream/nats-server.conf" &

    # Give the server a moment to start
    sleep 5

    # Verify NATS server is running
    if pgrep -x "nats-server" > /dev/null
    then
        echo "NATS server is now running."
        echo "You can use the following subjects for messaging:"
    else
        echo "Failed to start NATS server."
        exit 1
    fi
fi
check_status "Starting NATS server"

echo "All services started successfully."
