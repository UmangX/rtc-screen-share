# receiver.py
import socket
import struct
import cv2
import numpy as np

def recvall(sock, count):
    """Helper function to receive exactly 'count' bytes."""
    buf = b''
    while len(buf) < count:
        newbuf = sock.recv(count - len(buf))
        if not newbuf: return None # Connection closed
        buf += newbuf
    return buf

def main():
    # --- Configuration ---
    # Replace with the actual IP address of the sender computer
    sender_ip = "127.0.0.1" # <-- IMPORTANT: Change this!
    port = 9999
    # -------------------

    client_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)

    print(f"[*] Connecting to {sender_ip}:{port}...")
    try:
        client_socket.connect((sender_ip, port))
        print("[*] Connected to sender.")
    except ConnectionRefusedError:
        print(f"[!] Connection failed. Is the sender running at {sender_ip}:{port}?")
        return
    except socket.gaierror:
         print(f"[!] Hostname could not be resolved. Check the sender IP address: {sender_ip}")
         return
    except Exception as e:
        print(f"[!] Error connecting: {e}")
        return

    data = b""
    payload_size = struct.calcsize(">L") # Should match the sender's packing format

    try:
        while True:
            # 1. Receive the packed size
            packed_msg_size = recvall(client_socket, payload_size)
            if not packed_msg_size:
                print("[!] Sender closed the connection (packed_msg_size is None).")
                break

            # 2. Unpack the size
            try:
                 msg_size = struct.unpack(">L", packed_msg_size)[0]
            except struct.error as e:
                print(f"[!] Error unpacking size: {e}. Received data: {packed_msg_size}")
                break


            # 3. Receive the frame data based on the unpacked size
            frame_data = recvall(client_socket, msg_size)
            if not frame_data:
                 print("[!] Sender closed the connection (frame_data is None).")
                 break

            # 4. Decode the frame
            frame = cv2.imdecode(np.frombuffer(frame_data, dtype=np.uint8), cv2.IMREAD_COLOR)

            if frame is None:
                print("[!] Error decoding frame.")
                continue

            # 5. Display the frame
            cv2.imshow('Screen Share', frame)

            # Exit on 'q' key press
            if cv2.waitKey(1) & 0xFF == ord('q'):
                break

    except KeyboardInterrupt:
        print("\n[*] Stopping receiver.")
    except (ConnectionResetError, BrokenPipeError, ConnectionAbortedError):
         print("[!] Connection to sender lost.")
    except Exception as e:
        print(f"\n[!] An error occurred: {e}")
    finally:
        client_socket.close()
        cv2.destroyAllWindows()
        print("[*] Socket closed and windows destroyed.")


if __name__ == "__main__":
    main()
