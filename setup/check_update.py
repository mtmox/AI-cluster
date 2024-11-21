#!/opt/homebrew/bin/python3

import urllib.request
import urllib.error
import json
import subprocess
import logging
from datetime import datetime
import os

# Configuration
SERVER_URL = 'http://ui3u.local:8090'  # Replace with your server's IP and port
UPDATE_CHECK = os.path.join(os.environ['HOME'], 'AI-cluster', 'setup', 'update_repo.sh')
UPDATE_BUILD = os.path.join(os.environ['HOME'], 'AI-cluster', 'setup', 'build.sh')

def check_for_update():
    try:
        with urllib.request.urlopen(f"{SERVER_URL}/cluster_update") as response:
            if response.getcode() == 200:
                data = json.loads(response.read().decode())
                if data.get('update_available'):
                    logging.info("Update available. Running update script...")
                    run_update_script()
                else:
                    logging.info("No update available.")
            else:
                logging.error(f"Error checking for update: HTTP {response.getcode()}")
    except urllib.error.URLError as e:
        logging.error(f"Error connecting to server: {e}")

def run_update_script():
    try:
        subprocess.run(["bash", UPDATE_CHECK], check=True)
        logging.info("Update script executed successfully")
        run_build_script()
    except subprocess.CalledProcessError as e:
        logging.error(f"Error executing update script: {e}")

def run_build_script():
    try:
        subprocess.run(["bash", UPDATE_BUILD], check=True)
        logging.info("Build script executed successfully")
    except subprocess.CalledProcessError as e:
        logging.error(f"Error executing build script: {e}")

if __name__ == "__main__":
    logging.info(f"Running update check at {datetime.now()}")
    check_for_update()