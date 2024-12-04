#!/opt/homebrew/bin/python3

import urllib.request
import urllib.error
import json
import subprocess
import logging
from datetime import datetime
import os
import socket

# Get the directory where the script is located
SCRIPT_DIR = os.path.dirname(os.path.abspath(__file__))

# Configure logging with absolute path
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s',
    filename=os.path.join(SCRIPT_DIR, 'stop_node.log')
)

# Configuration
SERVER_IP = '192.168.1.16'  # Replace with your server's hostname
SERVER_URL = f'http://{SERVER_IP}:8090'
RUN_STOP = os.path.join(os.environ['HOME'], 'AI-cluster', 'setup', 'stop.sh')

def is_server():
    try:
        host_ip = socket.gethostbyname(socket.gethostname())
        server_ip = socket.gethostbyname(SERVER_IP)
        return host_ip == server_ip
    except:
        return False

def check_stop_flag():
    try:
        with urllib.request.urlopen(f"{SERVER_URL}/stop_node") as response:
            if response.getcode() == 200:
                data = json.loads(response.read().decode())
                if data.get('stop_available'):
                    logging.info("Stop flag is true, running stop script...")
                    run_stop_script()
                else:
                    logging.info("Stop flag is false, skipping stop.")
            elif response.getcode() == 204:
                logging.info("Stop check skipped - no action required")
            else:
                logging.error(f"Error checking stop status: HTTP {response.getcode()}")
    except urllib.error.URLError as e:
        logging.error(f"Error connecting to server: {e}")

def run_stop_script():
    try:
        subprocess.run(["bash", RUN_STOP], check=True)
        logging.info("Stop script executed successfully")
    except subprocess.CalledProcessError as e:
        logging.error(f"Error executing stop script: {e}")

if __name__ == "__main__":
    if not is_server():
        check_stop_flag()
    else:
        logging.info("Script running on server, skipping execution")
