#!/opt/homebrew/bin/python3

import urllib.request
import urllib.error
import json
import subprocess
import logging
from datetime import datetime
import os
import socket

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s',
    filename='start_node.log'
)

# Configuration
SERVER_IP = '192.168.1.16'  # Replace with your server's hostname
SERVER_URL = f'http://{SERVER_IP}:8090'
RUN_START = os.path.join(os.environ['HOME'], 'AI-cluster', 'setup', 'start.sh')

def is_server():
    try:
        host_ip = socket.gethostbyname(socket.gethostname())
        server_ip = socket.gethostbyname(SERVER_IP)
        return host_ip == server_ip
    except:
        return False

def check_start_flag():
    try:
        with urllib.request.urlopen(f"{SERVER_URL}/start_node") as response:
            if response.getcode() == 200:
                data = json.loads(response.read().decode())
                if data.get('start_available'):
                    logging.info("Start flag is true, running start script...")
                    run_start_script()
                else:
                    logging.info("Start flag is false, skipping start.")
                    return
            else:
                logging.error(f"Error checking start status: HTTP {response.getcode()}")
    except urllib.error.URLError as e:
        logging.error(f"Error connecting to server: {e}")

def run_start_script():
    try:
        subprocess.run(["bash", RUN_START], check=True)
        logging.info("Start script executed successfully")
    except subprocess.CalledProcessError as e:
        logging.error(f"Error executing start script: {e}")

if __name__ == "__main__":
    if not is_server():
        check_start_flag()
    else:
        logging.info("Script running on server, skipping execution")
