#!/bin/bash

# Update Git Repository Script

# Function to display error messages and exit
error_exit() {
    echo "Error: $1" >&2
    exit 1
}

update_repo() {
    local REPO_NAME="$1"
    local REPO_PATH="$2"
    local BRANCH_NAME="master"  # or "main", depending on your repository
    local DEPLOY_KEY="$HOME/.ssh/deploy_key_${REPO_NAME}"

    # Check if the deploy key exists
    if [ ! -f "$DEPLOY_KEY" ]; then
        error_exit "Deploy key not found for $REPO_NAME"
    fi

    # Use the specific deploy key for this repository
    export GIT_SSH_COMMAND="ssh -i $DEPLOY_KEY -o IdentitiesOnly=yes"

    # Navigate to the repository
    cd "$REPO_PATH" || error_exit "Failed to change directory to $REPO_PATH"

    # Fetch the latest changes
    echo "Fetching latest changes for $REPO_PATH..."
    git fetch origin || error_exit "Failed to fetch changes from remote"

    # Check current branch
    CURRENT_BRANCH=$(git rev-parse --abbrev-ref HEAD)
    if [ "$CURRENT_BRANCH" != "$BRANCH_NAME" ]; then
        echo "Switching to branch $BRANCH_NAME..."
        git checkout "$BRANCH_NAME" || error_exit "Failed to switch to branch $BRANCH_NAME"
    fi

    # Reset local changes and pull the latest changes
    echo "Resetting local changes and pulling latest changes..."
    git reset --hard origin/"$BRANCH_NAME" || error_exit "Failed to reset and pull changes"

    echo "Update completed successfully for $REPO_PATH!"

    # Unset the GIT_SSH_COMMAND to avoid affecting other operations
    unset GIT_SSH_COMMAND
}

echo "All updates completed successfully!"
