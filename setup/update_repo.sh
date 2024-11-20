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

update_home_paths() {
    local REPO_PATH="$1"
    local CURRENT_HOME=$(eval echo ~$USER)
    local MASTER_HOME="/Users/ui3u"

    echo "Updating home paths in $REPO_PATH..."
    cd "$REPO_PATH" || error_exit "Failed to change directory to $REPO_PATH"

    # Find all files (excluding .git directory) and replace paths
    find . -type f -not -path "./.git/*" | while read -r file; do
        if grep -q "$MASTER_HOME" "$file"; then
            sed -i.bak "s|$MASTER_HOME|$CURRENT_HOME|g" "$file"
            if cmp -s "$file" "$file.bak"; then
                rm "$file.bak"
            else
                echo "Modified: $file"
            fi
        fi
    done

    echo "Home paths updated in $REPO_PATH"
}

# Ensure the script is run from the correct directory
cd "$(dirname "$0")" || error_exit "Failed to change to script directory"

# List of repositories to update
REPOS="AI-cluster:$HOME/AI-cluster"

# Loop through each repository and update it
echo "$REPOS" | tr ' ' '\n' | while IFS=':' read -r repo_name repo_path; do
    update_repo "$repo_name" "$repo_path"
done

#### Fix the alteranting MASTER_HOME updating, not correct in mini nodes.

# Update home paths only in the AI-cluster repository
update_home_paths "$HOME/AI-cluster"

# Remove all .bak files in all repositories
echo "$REPOS" | tr ' ' '\n' | while IFS=':' read -r _ repo_path; do
    echo "Removing .bak files in $repo_path..."
    find "$repo_path" -name "*.bak" -type f -delete
done

echo "All updates completed successfully!"
