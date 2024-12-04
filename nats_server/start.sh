
#!/bin/bash

# Function to check if NATS server is running
check_nats_running() {
    pgrep -x "nats-server" > /dev/null
    return $?
}

# Function to log messages
log_message() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1"
}

# Set the working directory to the script's location
cd "$(dirname "$0")" || exit 1

# Check if NATS server is already running
if check_nats_running; then
    log_message "NATS server is already running."
    exit 0
fi

# Start NATS server with configuration file using absolute path
log_message "Starting NATS server..."
nohup /opt/homebrew/bin/nats-server -c "$HOME/AI-cluster/nats_server/nats-server.conf" > nats.log 2>&1 &

# Give the server a moment to start
sleep 2

# Verify NATS server is running
for i in {1..5}; do
    if check_nats_running; then
        log_message "NATS server is now running."
        exit 0
    fi
    log_message "Waiting for NATS server to start (attempt $i/5)..."
    sleep 2
done

log_message "Failed to start NATS server."
exit 1
