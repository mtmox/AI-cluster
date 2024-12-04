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
    filename=os.path.join(SCRIPT_DIR, 'check_update.log')
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
            elif response.getcode() == 204:
                logging.info("Update check skipped - no action required")
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
        # Set up the environment variables needed for the build script
        env = os.environ.copy()
        env['PATH'] = '/opt/homebrew/bin:/opt/homebrew/sbin:/usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin:' + env.get('PATH', '')
        env['GOPATH'] = os.path.expanduser('~/go')
        env['GOROOT'] = '/opt/homebrew/opt/go/libexec'  # Adjust this path if needed

        # Run the build script with the enhanced environment
        result = subprocess.run(
            ["bash", UPDATE_BUILD],
            check=True,
            env=env,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            text=True
        )
        logging.info("Build script executed successfully")
        logging.debug(f"Build output: {result.stdout}")
    except subprocess.CalledProcessError as e:
        logging.error(f"Error executing build script: {e}")
        logging.error(f"Build script stderr: {e.stderr}")
        logging.error(f"Build script stdout: {e.stdout}")

if __name__ == "__main__":
    if not is_server():
        check_for_update()
    else:
        logging.info("Script running on server, skipping execution")
