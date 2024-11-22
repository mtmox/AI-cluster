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
RUN_START = os.path.join(os.environ['HOME'], 'AI-cluster', 'setup', 'start.sh')

def check_for_update():
    try:
        with urllib.request.urlopen(f"{SERVER_URL}/start_node") as response:
            if response.getcode() == 200:
                data = json.loads(response.read().decode())
                if data.get('start_available'):
                    logging.info("Starting NODE...")
                    run_start_script()
                else:
                    logging.info("Not available.")
            else:
                logging.error(f"Error checking for update: HTTP {response.getcode()}")
    except urllib.error.URLError as e:
        logging.error(f"Error connecting to server: {e}")

def run_start_script():
    try:
        subprocess.run(["bash", RUN_START], check=True)
        logging.info("Start script executed successfully")
    except subprocess.CalledProcessError as e:
        logging.error(f"Error executing start script: {e}")

if __name__ == "__main__":
    logging.info(f"Starting NODE at {datetime.now()}")
    check_for_update()