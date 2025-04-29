from flask import Flask, Response, render_template_string
import cv2
import numpy as np
from mss import mss
import time
import io # Not strictly needed if directly yielding bytes, but good practice

# --- Configuration ---
port = 4221  # Port the web server will run on
quality = 100 # JPEG quality (0-100)
width = 1280 # Desired width for streaming (aspect ratio maintained)
# -------------------

app = Flask(__name__)

def generate_frames():
    """Generator function to capture screen frames and yield them as JPEG bytes."""
    with mss() as sct:
        monitor = sct.monitors[1] # Grab the main monitor

        while True:
            try:
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
                    # Optionally yield a placeholder image or just continue
                    continue

                frame_bytes = encoded_frame.tobytes()

                # 4. Yield the frame in MJPEG format
                yield (b'--frame\r\n'
                       b'Content-Type: image/jpeg\r\n\r\n' + frame_bytes + b'\r\n')

                # Control frame rate slightly (adjust as needed for performance)
                time.sleep(0.03) # ~30 FPS target, reduce sleep for higher FPS

            except Exception as e:
                print(f"Error in frame generation: {e}")
                # You might want to break or add more robust error handling
                time.sleep(1) # Avoid rapid error loops


@app.route('/')
def index():
    """Serves the HTML page that displays the video stream."""
    # Simple HTML page with an img tag pointing to the video feed endpoint
    html_content = """
    <html>
    <head>
        <title>Screen Share</title>
        <style>
            body { margin: 0; background-color: #222; display: flex; justify-content: center; align-items: center; height: 100vh; }
            img { max-width: 100%; max-height: 100%; object-fit: contain; }
        </style>
    </head>
    <body>
        <img src="{{ url_for('video_feed') }}">
    </body>
    </html>
    """
    return render_template_string(html_content)

@app.route('/video_feed')
def video_feed():
    """Route that serves the MJPEG stream."""
    return Response(generate_frames(),
                    mimetype='multipart/x-mixed-replace; boundary=frame')

if __name__ == '__main__':
    print(f"[*] Starting web server on port {port}...")
    print(f"[*] Open your browser and navigate to http://<your-ip-address>:{port}")
    print("[*] Or http://127.0.0.1:{port} on this machine.")
    # Use host='0.0.0.0' to make it accessible from other devices on the network
    # threaded=True allows handling multiple requests concurrently (HTML page + stream)
    app.run(host='0.0.0.0', port=port, threaded=True, debug=False)
    # Note: debug=True is useful for development but should be False for stability.
