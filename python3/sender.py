# sender.py
import socket
import struct
import time
import numpy as np
import cv2
from mss import mss

def main():
    # --- Configuration ---
    # Automatically get host IP (might need manual adjustment if multiple network interfaces)
    # Or replace with your specific local IP address (e.g., '192.168.1.10')
    #host_ip = socket.gethostbyname(socket.gethostname())
    host_ip = "127.0.0.1"
    port = 9999
    quality = 80  # JPEG quality (0-100, higher means better quality but more data)
    width = 1280  # Desired width for streaming (aspect ratio maintained)
    # -------------------

    print(f"[*] Sender starting on {host_ip}:{port}")
    print("[*] Waiting for a connection...")

    server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    server_socket.bind((host_ip, port))
    server_socket.listen(5)

    client_socket, addr = server_socket.accept()
    print(f"[*] Got connection from: {addr}")

    try:
        with mss() as sct:
            # Determine monitor dimensions dynamically if possible, or use a default
            monitor = sct.monitors[1] # Grab the main monitor (index 1)

            while True:
                # 1. Capture the screen
                img = sct.grab(monitor)
                frame = np.array(img)

                # Optional: Convert BGRX (mss format) to BGR (OpenCV format)
                frame = cv2.cvtColor(frame, cv2.COLOR_BGRA2BGR)

                # 2. Resize frame (optional, reduces bandwidth)
                current_height, current_width, _ = frame.shape
                aspect_ratio = current_width / current_height
                new_height = int(width / aspect_ratio)
                resized_frame = cv2.resize(frame, (width, new_height))

                # 3. Encode frame as JPEG
                encode_param = [int(cv2.IMWRITE_JPEG_QUALITY), quality]
                result, encoded_frame = cv2.imencode('.jpg', resized_frame, encode_param)

                if not result:
                    print("Error encoding frame")
                    continue

                # 4. Pack the frame data
                data = encoded_frame.tobytes()
                size = len(data)
                packed_size = struct.pack(">L", size) #'>L' for big-endian unsigned long (4 bytes)

                # 5. Send data: size first, then frame
                try:
                    client_socket.sendall(packed_size)
                    client_socket.sendall(data)
                except (ConnectionResetError, BrokenPipeError, ConnectionAbortedError):
                    print("[!] Client disconnected.")
                    break

                # Optional: Add a small delay to control frame rate
                # time.sleep(0.01) # Adjust as needed

    except KeyboardInterrupt:
        print("\n[*] Stopping sender.")
    finally:
        client_socket.close()
        server_socket.close()
        print("[*] Sockets closed.")

if __name__ == "__main__":
    main()
