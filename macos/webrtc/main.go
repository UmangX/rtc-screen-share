package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"time" // Import time for timeouts

	"github.com/pion/webrtc/v3"
)

const (
	webPort = ":8080"
	udpPort = ":5004" // Define UDP port as a constant
)

var peer_connection *webrtc.PeerConnection

func main() {
	fmt.Println("This is running the webrtc server ...")

	// this is the config for web server file serving
	root := "./static"
	fs := http.FileServer(http.Dir(root))
	http.Handle("/", fs)

	// this is the config for webrtc
	// Added a public STUN server for better ICE candidate gathering,
	// even for local testing, as it helps in discovering host candidates.
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{},
	}
	var err error
	peer_connection, err = webrtc.NewPeerConnection(config)
	if err != nil {
		fmt.Printf("Error: there is some error initializing the peer connection: %v\n", err) // More detailed error
		return                                                                               // Exit if peer connection fails to initialize
	}

	// --- Debugging for PeerConnection States ---
	peer_connection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		fmt.Printf("ICE Connection State has changed: %s\n", connectionState.String())
	})

	peer_connection.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		fmt.Printf("Peer Connection State has changed: %s\n", state.String())
		if state == webrtc.PeerConnectionStateFailed || state == webrtc.PeerConnectionStateDisconnected {
			fmt.Println("Peer connection failed or disconnected. Investigate why!")
			// You might want to add logic here to clean up or attempt re-connection
		} else if state == webrtc.PeerConnectionStateConnected {
			fmt.Println("Peer Connection is CONNECTED!")
		}
	})
	// --- End Debugging for PeerConnection States ---

	// Setup video track
	videotrack, err := webrtc.NewTrackLocalStaticRTP(
		webrtc.RTPCodecCapability{
			MimeType:    webrtc.MimeTypeH264,
			ClockRate:   90000,
			SDPFmtpLine: "packetization-mode=1",
		},
		"video", "pion",
	)

	if err != nil {
		fmt.Printf("Error: failed to create video track: %v\n", err)
		return // Exit if track creation fails
	}
	transceiver, err := peer_connection.AddTransceiverFromTrack(videotrack, webrtc.RtpTransceiverInit{
		Direction: webrtc.RTPTransceiverDirectionSendonly,
	})
	if err != nil {
		fmt.Printf("Error: failed to add transceiver from track: %v\n", err)
		return // Exit if adding transceiver fails
	}
	fmt.Println("Added transceiver:", transceiver.Kind()) // Print transceiver kind for clarity

	// Goroutine to read UDP and write to WebRTC track
	go func() {
		conn, err := net.ListenPacket("udp", udpPort)
		if err != nil {
			fmt.Printf("Error: failed to listen on UDP port %s: %v\n", udpPort, err)
			return // Use return instead of panic for graceful shutdown in goroutine
		}
		defer func() {
			fmt.Printf("Closing UDP listener on %s\n", udpPort)
			conn.Close()
		}()

		fmt.Printf("Listening for H264 RTP on UDP %s...\n", udpPort)
		buffer := make([]byte, 1500) // Standard MTU size
		for {
			n, _, err := conn.ReadFrom(buffer)
			if err != nil {
				// Check for temporary network errors to avoid exiting on transient issues
				if netErr, ok := err.(net.Error); ok && netErr.Temporary() {
					fmt.Printf("Temporary read error from UDP: %v\n", err)
					time.Sleep(100 * time.Millisecond) // Wait a bit before retrying
					continue
				}
				fmt.Printf("Fatal read error from UDP: %v\n", err)
				break // Exit loop on fatal error
			}

			// fmt.Printf("Received %d bytes from %s on UDP\n", n, remoteAddr.String()) // Uncomment for verbose UDP debug

			// Write to the WebRTC track
			_, err = videotrack.Write(buffer[:n])
			if err != nil {
				// Ignore errors if the peer connection is not yet established or closed
				if err == io.ErrClosedPipe || err == io.EOF {
					// fmt.Println("Write to track error: Peer connection likely closed or not ready:", err) // Uncomment for verbose closed pipe debug
				} else {
					fmt.Printf("Write to track error: %v\n", err) // Log other errors
				}
			}
		}
	}()

	// HTTP handler for SDP exchange
	http.HandleFunc("/sdp", handle_sdp)

	fmt.Printf("WebRTC server listening on %s and serving static files from %s\n", webPort, root)
	err = http.ListenAndServe(webPort, nil)
	if err != nil {
		fmt.Printf("Error: HTTP server failed to start: %v\n", err)
	}
}

// handle_sdp handles the SDP offer/answer exchange
func handle_sdp(w http.ResponseWriter, r *http.Request) {
	fmt.Println("\n--- Handling /sdp request ---") // Start of request debug

	if r.Method != http.MethodPost {
		fmt.Printf("Error: Method not allowed: %s\n", r.Method)
		http.Error(w, "Method is not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("Error: reading the request body: %v\n", err)
		http.Error(w, "Error reading the request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	fmt.Printf("Received SDP data from client: %s\n", string(body))

	if peer_connection == nil {
		fmt.Println("Error: peer connection not initialized.")
		http.Error(w, "Error in peer initialization", http.StatusInternalServerError)
		return
	}

	var offer webrtc.SessionDescription
	if err := json.Unmarshal(body, &offer); err != nil {
		fmt.Printf("Error: failed to parse sdp offer: %v\n", err)
		http.Error(w, "Error: failed to parse sdp offer", http.StatusBadRequest)
		return
	}
	fmt.Printf("Successfully parsed SDP offer. Type: %s\n", offer.Type)

	// Set the remote description (the offer from the client)
	if err := peer_connection.SetRemoteDescription(offer); err != nil {
		fmt.Printf("Error: failed to set remote description: %v\n", err)
		http.Error(w, "Error: failed to set remote description", http.StatusInternalServerError)
		return
	}
	fmt.Println("Remote description set successfully.")

	// Create an SDP answer
	answer, err := peer_connection.CreateAnswer(nil)
	if err != nil {
		fmt.Printf("Error: creating the SDP answer: %v\n", err)
		http.Error(w, "Error: creating the answer", http.StatusInternalServerError)
		return
	}
	fmt.Println("SDP answer created.")

	// Set the local description (the answer generated by the server)
	err = peer_connection.SetLocalDescription(answer)
	if err != nil {
		fmt.Printf("Error: setting the local description on the host: %v\n", err)
		http.Error(w, "Error: setting the local description on the host", http.StatusInternalServerError)
		return
	}
	fmt.Println("Local description set successfully.")

	// Wait for ICE gathering to complete. This is crucial as it ensures
	// all local candidates are gathered before sending the answer.
	// Add a timeout to prevent indefinite waiting if ICE gathering fails for some reason.
	select {
	case <-webrtc.GatheringCompletePromise(peer_connection):
		fmt.Println("ICE gathering complete.")
	case <-time.After(10 * time.Second): // 10 second timeout for ICE gathering
		fmt.Println("Warning: ICE gathering timed out after 10 seconds.")
		// It might still work, but could indicate a problem if no candidates were found.
	}

	// Send the SDP answer back to the client
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(peer_connection.LocalDescription()); err != nil {
		fmt.Printf("Error: failed to encode and send local description: %v\n", err)
		http.Error(w, "Error: failed to encode and send local description", http.StatusInternalServerError)
		return
	}
	fmt.Println("SDP answer sent to client successfully.")
	fmt.Println("--- /sdp request handling complete ---") // End of request debug
}
