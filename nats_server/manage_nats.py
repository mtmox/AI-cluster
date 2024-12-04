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
    filename=os.path.join(SCRIPT_DIR, 'flag.log')
)

# Configuration
SERVER_IPS = [
    '192.168.1.16',
    '192.168.1.22',
    '192.168.1.34',
    '192.168.1.35',
    '192.168.1.83',
    '192.168.1.140'
]

def get_nats_config():
    try:
        with open(os.path.join(SCRIPT_DIR, 'nats-server-url.json'), 'r') as f:
            config = json.load(f)
            return config.get('nats_url', '')
    except Exception as e:
        logging.error(f"Error reading NATS config: {e}")
        return ''

def extract_ip_from_nats_url(nats_url):
    try:
        # Extract IP from nats://192.168.1.140:4222 format
        return nats_url.split('://')[1].split(':')[0]
    except:
        return ''

def is_server():
    try:
        host_ip = socket.gethostbyname(socket.gethostname())
        nats_url = get_nats_config()
        nats_ip = extract_ip_from_nats_url(nats_url)
        
        # Check if the host IP matches the NATS IP
        return host_ip == nats_ip
    except:
        return False

def check_flag():
    for server_ip in SERVER_IPS:
        try:
            server_url = f'http://{server_ip}:8091'
            with urllib.request.urlopen(f"{server_url}/flag") as response:
                if response.getcode() == 200:
                    data = json.loads(response.read().decode())
                    if data.get('start_available'):
                        logging.info("Start flag is true, running start script...")
                        run_start_script()
                        break
                    elif data.get('stop_available'):
                        logging.info("Stop flag is true, running stop script...")
                        run_stop_script()
                        break
                    else:
                        logging.info("No flag skipping actions...")
                elif response.getcode() == 204:
                    logging.info(f"Start check skipped for {server_ip} - no action required")
                else:
                    logging.error(f"Error checking start status on {server_ip}: HTTP {response.getcode()}")
        except urllib.error.URLError as e:
            logging.error(f"Error connecting to server {server_ip}: {e}")
            continue

def run_start_script():
    try:
        subprocess.run(["bash", os.path.join(os.environ['HOME'], 'AI-cluster', 'nats_server', 'start.sh')], check=True)
        logging.info("Start script executed successfully")
    except subprocess.CalledProcessError as e:
        logging.error(f"Error executing start script: {e}")

def run_stop_script():
    try:
        subprocess.run(["bash", os.path.join(os.environ['HOME'], 'AI-cluster', 'nats_server', 'stop.sh')], check=True)
        logging.info("Stop script executed successfully")
    except subprocess.CalledProcessError as e:
        logging.error(f"Error executing stop script: {e}")        

if __name__ == "__main__":
    if is_server():
        check_flag()
    else:
        logging.info("Script not running on NATS server, skipping execution")
