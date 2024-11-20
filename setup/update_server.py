
from http.server import HTTPServer, BaseHTTPRequestHandler
import json
import threading
import time

# Configuration
HOST = '0.0.0.0'
PORT = 8090

class UpdateHandler(BaseHTTPRequestHandler):
    def do_GET(self):
        if self.path == '/cluster_update':
            self.send_response(200)
            self.send_header('Content-type', 'application/json')
            self.end_headers()
            response = json.dumps({"update_available": True})
            self.wfile.write(response.encode())
        else:
            self.send_error(404)

def run_server():
    server_address = (HOST, PORT)
    httpd = HTTPServer(server_address, UpdateHandler)
    print(f"Server running on http://{HOST}:{PORT}")
    httpd.serve_forever()

if __name__ == "__main__":
    server_thread = threading.Thread(target=run_server)
    server_thread.start()

    while True:
        command = input("Enter 'quit' to stop the server: ")
        if command.lower() == 'quit':
            break

    print("Server shutting down...")
