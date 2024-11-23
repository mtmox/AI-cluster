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
    filename='check_update.log'
)

# Configuration
SERVER_IP = '192.168.1.16'  # Replace with your server's hostname
SERVER_URL = f'http://{SERVER_IP}:8090'
UPDATE_CHECK = os.path.join(os.environ['HOME'], 'AI-cluster', 'setup', 'update_repo.sh')
UPDATE_BUILD = os.path.join(os.environ['HOME'], 'AI-cluster', 'setup', 'build.sh')

def is_server():
    try:
        host_ip = socket.gethostbyname(socket.gethostname())
        server_ip = socket.gethostbyname(SERVER_IP)
        return host_ip == server_ip
    except:
        return False

def check_for_update():
    try:
        with urllib.request.urlopen(f"{SERVER_URL}/cluster_update") as response:
            if response.getcode() == 200:
                data = json.loads(response.read().decode())
                if data.get('update_available'):
                    logging.info("Update flag is true, running update script...")
                    run_update_script()
                else:
                    logging.info("Update flag is false, skipping update.")
                    return
            else:
                logging.error(f"Error checking update status: HTTP {response.getcode()}")
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
    if not is_server():
        check_for_update()
    else:
        logging.info("Script running on server, skipping execution")
