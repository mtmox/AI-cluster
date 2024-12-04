from http.server import HTTPServer, BaseHTTPRequestHandler
import json
import threading
import time
import os
from termcolor import colored

# Configuration
HOST = '0.0.0.0'
PORT = 8090

# Global variables to control node actions
update_available = False
start_available = False
stop_available = False

class UpdateHandler(BaseHTTPRequestHandler):
    def do_GET(self):
        global update_available, start_available, stop_available

        if self.path == '/cluster_update':
            self.handle_response('update_available', update_available)
        elif self.path == '/start_node':
            self.handle_response('start_available', start_available)
        elif self.path == '/stop_node':
            self.handle_response('stop_available', stop_available)
        else:
            self.send_error(404)

    def handle_response(self, flag_name, flag_value):
        if flag_value:
            # If flag is true, send 200 OK
            self.send_response(200)
            print(colored(f"200 - {self.path} - Action allowed", 'green'))
        else:
            # If flag is false, send 204 No Content
            self.send_response(204)
            print(colored(f"204 - {self.path} - Action skipped", 'red'))
            
        self.send_header('Content-type', 'application/json')
        self.end_headers()
        response = json.dumps({flag_name: flag_value})
        self.wfile.write(response.encode())

def run_server():
    server_address = (HOST, PORT)
    httpd = HTTPServer(server_address, UpdateHandler)
    print(f"Server running on http://{HOST}:{PORT}")
    httpd.serve_forever()

def clear_screen():
    # Clear the terminal screen
    os.system('cls' if os.name == 'nt' else 'clear')

def display_menu():
    global update_available, start_available, stop_available
    print("\nCurrent status:")
    print(f"Update available: {'Yes' if update_available else 'No'}")
    print(f"Start available: {'Yes' if start_available else 'No'}")
    print(f"Stop available: {'Yes' if stop_available else 'No'}")
    print("\nAvailable options:")
    print("1. Toggle Update Cluster")
    print("2. Toggle Start Node")
    print("3. Toggle Stop Node")
    print("4. Quit")

if __name__ == "__main__":
    server_thread = threading.Thread(target=run_server)
    server_thread.start()

    while True:
        clear_screen()
        display_menu()
        choice = input("Enter your choice (1-4): ")

        if choice == '1':
            update_available = not update_available
            print(f"Update Cluster toggled to: {'Available' if update_available else 'Not Available'}")
        elif choice == '2':
            start_available = not start_available
            print(f"Start Node toggled to: {'Available' if start_available else 'Not Available'}")
        elif choice == '3':
            stop_available = not stop_available
            print(f"Stop Node toggled to: {'Available' if stop_available else 'Not Available'}")
        elif choice == '4':
            print("Server shutting down...")
            break
        else:
            print("Invalid choice. Please try again.")

    # Terminate the server thread
    os._exit(0)
