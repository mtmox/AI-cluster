
from http.server import HTTPServer, BaseHTTPRequestHandler
import json
import threading
import time
import os
from termcolor import colored

# Configuration
HOST = '0.0.0.0'
PORT = 8091

# Global variables to control node actions
start_available = False
stop_available = False

class UpdateHandler(BaseHTTPRequestHandler):
    def do_GET(self):
        global start_available, stop_available

        if self.path == '/flag':
            # Send both flags in the response
            response_data = {
                'start_available': start_available,
                'stop_available': stop_available
            }
            
            # If either flag is true, send 200 OK
            if start_available or stop_available:
                self.send_response(200)
                print(colored(f"200 - {self.path} - Action allowed", 'green'))
            else:
                # If both flags are false, send 204 No Content
                self.send_response(204)
                print(colored(f"204 - {self.path} - Action skipped", 'red'))
                
            self.send_header('Content-type', 'application/json')
            self.end_headers()
            self.wfile.write(json.dumps(response_data).encode())
        else:
            self.send_error(404)

def run_server():
    server_address = (HOST, PORT)
    httpd = HTTPServer(server_address, UpdateHandler)
    print(f"Server running on http://{HOST}:{PORT}")
    httpd.serve_forever()

def clear_screen():
    # Clear the terminal screen
    os.system('cls' if os.name == 'nt' else 'clear')

def display_menu():
    global start_available, stop_available
    print("\nCurrent status:")
    print(f"Start available: {'Yes' if start_available else 'No'}")
    print(f"Stop available: {'Yes' if stop_available else 'No'}")
    print("\nAvailable options:")
    print("1. Toggle Start Node")
    print("2. Toggle Stop Node")
    print("3. Quit")

if __name__ == "__main__":
    server_thread = threading.Thread(target=run_server)
    server_thread.start()

    while True:
        clear_screen()
        display_menu()
        choice = input("Enter your choice (1-3): ")

        if choice == '1':
            start_available = not start_available
            print(f"Start Node toggled to: {'Available' if start_available else 'Not Available'}")
        elif choice == '2':
            stop_available = not stop_available
            print(f"Stop Node toggled to: {'Available' if stop_available else 'Not Available'}")
        elif choice == '3':
            print("Server shutting down...")
            break
        else:
            print("Invalid choice. Please try again.")

    # Terminate the server thread
    os._exit(0)
