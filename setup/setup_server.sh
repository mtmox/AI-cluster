#!/bin/bash

# Function to check if a command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Install Homebrew if not already installed
if ! command_exists brew; then
    echo "Installing Homebrew..."
    /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
    
    # Add Homebrew to PATH
    echo 'eval "$(/opt/homebrew/bin/brew shellenv)"' >> ~/.zprofile
    eval "$(/opt/homebrew/bin/brew shellenv)"
else
    echo "Homebrew is already installed."
fi

# Install Oh My Zsh if not already installed
if [ ! -d "$HOME/.oh-my-zsh" ]; then
    echo "Installing Oh My Zsh..."
    sh -c "$(curl -fsSL https://raw.github.com/ohmyzsh/ohmyzsh/master/tools/install.sh)"
else
    echo "Oh My Zsh is already installed."
fi

# Install Homebrew packages
echo "Installing Homebrew packages..."
brew install python3 go git nats-server nats-io/nats-tools/nats


# Check and add cron job if not present
echo "Checking cron job..."
CRON_JOB="* * * * * /opt/homebrew/bin/python3 $HOME/AI-cluster/setup/check_update.py"
if ! crontab -l 2>/dev/null | grep -q "check_update.py"; then
    echo "Setting up cron job..."
    (crontab -l 2>/dev/null; echo "$CRON_JOB") | crontab -
    echo "Cron job added."
else
    echo "Cron job for check_update.py already exists. Skipping."
fi

# Function to create a deploy key for a repository
create_deploy_key() {
    local repo_name="$1"
    local key_file="$HOME/.ssh/deploy_key_${repo_name}"
    
    if [ ! -f "$key_file" ]; then
        ssh-keygen -t ed25519 -f "$key_file" -N "" -C "deploy_key_${repo_name}"
        echo "Deploy key created for $repo_name"
        echo "Public key for $repo_name:"
        cat "${key_file}.pub"
        echo "ADD THIS KEY TO GITHUB AS A DEPLOY KEY FOR $repo_name"
    else
        echo "Deploy key for $repo_name already exists. Skipping."
    fi
}

create_deploy_key "AI-cluster"

# Start the SSH agent
eval "$(ssh-agent -s)"

# Add the deploy keys to the agent
ssh-add $HOME/.ssh/deploy_key_AI-cluster

# Function to clone a repository if it doesn't exist
clone_repo() {
    local repo_url="$1"
    local repo_name=$(basename "$repo_url" .git)
    local key_file="$HOME/.ssh/deploy_key_${repo_name}"
    
    if [ ! -d "$HOME/$repo_name" ]; then
        GIT_SSH_COMMAND="ssh -i $key_file" git clone "$repo_url" "$HOME/$repo_name" && echo "Successfully cloned $repo_name" || echo "Failed to clone $repo_name"
    else
        echo "$repo_name already exists. Skipping."
    fi
}

# Change to the home directory
cd $HOME

# Loop to try cloning repositories every 60 seconds
while true; do
    clone_repo "git@github.com:mtmox/AI-cluster.git"
    
    # Check if the repository exists
    if [ -d "$HOME/AI-cluster" ]; then
        echo "Repository has been cloned successfully."
        break
    else
        echo "Waiting 60 seconds before trying again..."
        sleep 60
    fi
done

echo "Setup complete!"