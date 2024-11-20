#!/bin/bash

# Function to check the last command's status
check_status() {
    if [ $? -ne 0 ]; then
        echo "Error: $1 failed"
        exit 1
    fi
}

# Function to stop a process and verify it's terminated
stop_process() {
    local process_name=$1
    local pkill_pattern=$2

    echo "Stopping $process_name..."
    pkill -f "$pkill_pattern"
    sleep 2

    if pgrep -f "$pkill_pattern" > /dev/null; then
        echo "Failed to stop $process_name. Attempting to force kill..."
        pkill -9 -f "$pkill_pattern"
        sleep 1
        if pgrep -f "$pkill_pattern" > /dev/null; then
            echo "Failed to force kill $process_name. Please check manually."
            return 1
        else
            echo "$process_name forcefully stopped."
        fi
    else
        echo "$process_name stopped successfully."
    fi
}



# Check if NATS server is running
if ! pgrep -x "nats-server" > /dev/null
then
    echo "NATS server is not running."
else
    # Stop the NATS server
    echo "Stopping NATS server..."
    /opt/homebrew/bin/nats-server --signal quit
    
    # Wait for a moment to ensure the server has stopped
    sleep 5

    # Check if the server stopped successfully
    if pgrep -x "nats-server" > /dev/null
    then
        echo "Failed to stop NATS server gracefully. Attempting to force kill..."
        pkill -9 nats-server
        sleep 1
        if pgrep -x "nats-server" > /dev/null
        then
            echo "Failed to force kill NATS server. Please check manually."
            exit 1
        else
            echo "NATS server forcefully stopped."
        fi
    else
        echo "NATS server stopped successfully."
    fi

    echo "NATS server shutdown complete."
    check_status "Stopping NATS server"
fi


echo "All services stopped successfully."
