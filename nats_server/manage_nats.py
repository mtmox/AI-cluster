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
SERVER_IP = '192.168.1.16'  # Changed from SERVER_IPS to SERVER_IP

START_SCRIPT = os.path.join(os.environ['HOME'], 'AI-cluster', 'nats_server', 'start.sh')
STOP_SCRIPT = os.path.join(os.environ['HOME'], 'AI-cluster', 'nats_server', 'stop.sh')
NATS_URL = os.path.join(os.environ['HOME'], 'AI-cluster', 'nats_server', 'nats-server-url.sh')

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

def get_local_ip():
    try:
        # Create a socket and connect to an external address
        # This won't actually establish a connection but will help us
        # determine which local interface would be used
        s = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
        s.settimeout(0)  # Non-blocking mode
        try:
            # We don't need to actually reach this address
            s.connect(('10.254.254.254', 1))
            local_ip = s.getsockname()[0]
        except Exception:
            local_ip = '127.0.0.1'
        finally:
            s.close()
        return local_ip
    except Exception as e:
        logging.error(f"Error getting local IP: {e}")
        return '127.0.0.1'

def is_server():
    subprocess.run(["bash", NATS_URL], check=True)
    try:
        host_ip = get_local_ip()
        nats_url = get_nats_config()
        nats_ip = extract_ip_from_nats_url(nats_url)
        
        logging.info(f"host_ip={host_ip} and nats_ip={nats_ip}")
        # Check if the host IP matches the NATS IP
        return host_ip == nats_ip
    except Exception as e:
        logging.error(f"Error in is_server(): {e}")
        return False

def check_flag():
    try:
        server_url = f'http://{SERVER_IP}:8091'  # Using single SERVER_IP
        with urllib.request.urlopen(f"{server_url}/flag") as response:
            if response.getcode() == 200:
                data = json.loads(response.read().decode())
                if data.get('start_available'):
                    logging.info("Start flag is true, running start script...")
                    run_start_script()
                elif data.get('stop_available'):
                    logging.info("Stop flag is true, running stop script...")
                    run_stop_script()
                else:
                    logging.info("No flag skipping actions...")
            elif response.getcode() == 204:
                logging.info(f"Start check skipped for {SERVER_IP} - no action required")
            else:
                logging.error(f"Error checking start status on {SERVER_IP}: HTTP {response.getcode()}")
    except urllib.error.URLError as e:
        logging.error(f"Error connecting to server {SERVER_IP}: {e}")

def run_start_script():
    try:
        subprocess.run(["bash", START_SCRIPT], check=True) 
        logging.info("Start script executed successfully")
    except subprocess.CalledProcessError as e:
        logging.error(f"Error executing start script: {e}")

def run_stop_script():
    try:
        subprocess.run(["bash", STOP_SCRIPT], check=True)
        logging.info("Stop script executed successfully")
    except subprocess.CalledProcessError as e:
        logging.error(f"Error executing stop script: {e}")        

if __name__ == "__main__":
    if is_server():
        check_flag()
    else:
        logging.info("Script not running on NATS server, skipping execution")
