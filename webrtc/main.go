package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"

	"github.com/pion/webrtc/v3"
)

const webPort = ":8080"

var peerConnection *webrtc.PeerConnection

func main() {
	fmt.Println("Starting WebRTC screen-sharing server on http://localhost" + webPort)

	root := "./static"
	fs := http.FileServer(http.Dir(root))
	http.Handle("/", fs)

	// WebRTC configuration
	config := webrtc.Configuration{
		// zero config because there is no stun/turn server for local area networks
	}

	var err error
	peerConnection, err = webrtc.NewPeerConnection(config)
	if err != nil {
		fmt.Printf("❌ Error creating PeerConnection: %v\n", err)
		return
	}

	// Log ICE connection state changes
	peerConnection.OnICEConnectionStateChange(func(state webrtc.ICEConnectionState) {
		fmt.Printf(" ICE Connection State: %s\n", state.String())
	})

	// Log new ICE candidates
	peerConnection.OnICECandidate(func(c *webrtc.ICECandidate) {
		if c != nil {
			fmt.Printf(" New ICE candidate: %s\n", c.String())
		}
	})

	// Create video track for FFmpeg to write into
	videoTrack, err := webrtc.NewTrackLocalStaticRTP(
		webrtc.RTPCodecCapability{
			MimeType:    webrtc.MimeTypeH264,
			ClockRate:   90000,
			SDPFmtpLine: "packetization-mode=1",
		},
		"video", "screen",
	)
	if err != nil {
		fmt.Printf(" Error creating video track: %v\n", err)
		return
	}

	transceiver, err := peerConnection.AddTransceiverFromTrack(videoTrack, webrtc.RtpTransceiverInit{
		Direction: webrtc.RTPTransceiverDirectionSendonly,
	})
	if err != nil {
		fmt.Printf(" Error adding transceiver: %v\n", err)
		return
	}
	fmt.Println(" Transceiver added:", transceiver)

	// Start listening for RTP packets from FFmpeg
	go func() {
		fmt.Println(" Waiting for RTP packets on UDP port 5004...")
		conn, err := net.ListenPacket("udp", ":5004")
		if err != nil {
			fmt.Printf(" Error opening UDP port: %v\n", err)
			return
		}
		defer conn.Close()

		// we should make the buffer bigger there was some lag on high framerate videos
		buffer := make([]byte, 1500)
		for {
			n, _, err := conn.ReadFrom(buffer)
			if err != nil {
				fmt.Printf("❌ Error reading UDP packet: %v\n", err)
				break
			}

			_, err = videoTrack.Write(buffer[:n])
			if err != nil {
				fmt.Printf("❌ Error writing to video track: %v\n", err)
			}
		}
	}()

	// Handle SDP offer POSTs
	http.HandleFunc("/sdp", handleSDP)
	if err := http.ListenAndServe(webPort, nil); err != nil {
		fmt.Printf("❌ HTTP server error: %v\n", err)
	}
}

func handleSDP(w http.ResponseWriter, r *http.Request) {
	// CORS (if needed)
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method != http.MethodPost {
		http.Error(w, "❌ Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, "❌ Failed to read request body", http.StatusBadRequest)
		return
	}

	fmt.Printf(" SDP Offer Received:\n%s\n", string(body))

	if peerConnection == nil {
		http.Error(w, "❌ PeerConnection not initialized", http.StatusInternalServerError)
		return
	}

	var offer webrtc.SessionDescription
	if err := json.Unmarshal(body, &offer); err != nil {
		http.Error(w, "❌ Failed to parse SDP offer", http.StatusBadRequest)
		fmt.Println("❌ JSON parse error:", err)
		return
	}

	if err := peerConnection.SetRemoteDescription(offer); err != nil {
		http.Error(w, "❌ Failed to set remote description", http.StatusInternalServerError)
		fmt.Println("❌ SetRemoteDescription error:", err)
		return
	}
	fmt.Println(" Remote SDP set")

	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		http.Error(w, "❌ Failed to create SDP answer", http.StatusInternalServerError)
		fmt.Println("❌ CreateAnswer error:", err)
		return
	}

	if err := peerConnection.SetLocalDescription(answer); err != nil {
		http.Error(w, "❌ Failed to set local description", http.StatusInternalServerError)
		fmt.Println("❌ SetLocalDescription error:", err)
		return
	}

	<-webrtc.GatheringCompletePromise(peerConnection)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(peerConnection.LocalDescription()); err != nil {
		fmt.Printf("❌ Error encoding local SDP: %v\n", err)
	}
	fmt.Println(" SDP Answer sent to client")
}
