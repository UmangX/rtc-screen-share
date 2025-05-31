# rtc screen share ( this is not webrtc because its for lan)

**Stream your screen to any browser over LAN using WebRTC — no client installation required.**

7xScreenShare is a lightweight, local-only screen sharing solution that streams your desktop to a browser via WebRTC. Designed for simplicity and speed, it eliminates the need for STUN/TURN servers, making it ideal for secure, low-latency screen sharing within a local network.

----------

## 🚀 Project Goals

-   **Zero Client Setup:** Enable screen viewing in any modern browser without additional software.
    
-   **LAN-Only Operation:** Ensure all data remains within the local network for enhanced security and privacy.
    
-   **Low Latency:** Utilize FFmpeg and WebRTC to achieve minimal delay in screen streaming.
    
-   **Cross-Platform Support:** While initially tested on macOS, aim for compatibility with Linux and Windows.
    
-   **Modular Design:** Facilitate easy integration and extension for various use cases.
    

----------

## 🛠️ Prerequisites

Before running the project, ensure the following are installed:

-   [Go](https://golang.org/dl/) (version 1.16 or higher)
    
-   [FFmpeg](https://ffmpeg.org/download.html)
    
-   Make
    
-   [Pion WebRTC](https://github.com/pion/webrtc) Go library
    

> **Note:** This system is designed for LAN use and does not utilize STUN/TURN servers.

----------

## 🧪 Tested Environment

-   ✅ **macOS** (tested)
    
-   ⚠️ **Linux** (untested)
    
-   ⚠️ **Windows** (untested)
    

----------

## 🖥️ Usage

1.  **Start Screen Capture with FFmpeg:**
    
    In one terminal window, run:
    
    `make capture` 
    
2.  **Run the WebRTC Server:**
    
    In another terminal window, run:
    
    `make server` 
    
3.  **Access the Stream:**
    
    Open your browser and navigate to:
    
    `http://localhost:8080` 
    
    You should see your screen being streamed in real-time.
    

----------

## 📁 Project Structure

-   `main.go`: The Go server that handles WebRTC signaling and serves the web interface.
    
-   `Makefile`: Contains commands to start the FFmpeg capture and the Go server.
    
-   `index.html`: The client-side interface to view the stream.
    
-   `README.md`: Project documentation.
    

----------

## 🔧 Configuration

-   **FFmpeg Capture Settings:** Modify the `make capture` command in the `Makefile` to adjust resolution, frame rate, or input source.
    
-   **Server Port:** The default server runs on port `8080`. To change this, modify the `main.go` file accordingly.
    

----------

## 📜 License

This project is licensed under the MIT License.

