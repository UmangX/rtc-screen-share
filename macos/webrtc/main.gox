package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"

	"github.com/pion/webrtc/v3"
)

const (
	webPort = ":8080"
)

var peer_connection *webrtc.PeerConnection

func main() {
	fmt.Println("This is running the webrtc server ...")

	// this is the config for web server file serving
	root := "./static"
	fs := http.FileServer(http.Dir(root))
	http.Handle("/", fs)

	// this is the config for webrtc
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{},
	}
	var err error
	peer_connection, err = webrtc.NewPeerConnection(config)
	if err != nil {
		fmt.Printf("error : there is some error intializing the peer connection %e\n", err)
	}

	// --- Add ICE State Handlers ---
	peer_connection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		fmt.Printf("ICE Connection State has changed to: %s\n", connectionState.String())
	})
	// --- End Add ICE State Handlers ---

	videotrack, err := webrtc.NewTrackLocalStaticRTP(
		webrtc.RTPCodecCapability{
			MimeType:    webrtc.MimeTypeH264,
			ClockRate:   90000,
			SDPFmtpLine: "packetization-mode=1",
		},
		"video", "pion",
	)

	if err != nil {
		panic(err)
	}
	transceiver, err := peer_connection.AddTransceiverFromTrack(videotrack, webrtc.RtpTransceiverInit{
		Direction: webrtc.RTPTransceiverDirectionSendonly,
	})
	if err != nil {
		panic(err)
	}
	fmt.Println("Added transceiver:", transceiver)

	go func() {
		conn, err := net.ListenPacket("udp", ":5004")
		if err != nil {
			panic(err)
		}
		defer conn.Close()

		buffer := make([]byte, 1500)
		for {
			n, _, err := conn.ReadFrom(buffer)
			if err != nil {
				fmt.Println("Read error:", err)
				break
			}

			_, err = videotrack.Write(buffer[:n])
			if err != nil {
				fmt.Println("Write to track error:", err)
			}
		}
	}()

	// this is handling for web server
	http.HandleFunc("/sdp", handle_sdp)
	http.ListenAndServe(webPort, nil)
}

func handle_sdp(w http.ResponseWriter, r *http.Request) {
	//this is barebone http handling i think hence deter if the request is post
	if r.Method != http.MethodPost {
		http.Error(w, "Method is not alloweed", http.StatusMethodNotAllowed)
		return
	}

	// reading the files
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "error reading the request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// this is debugging for printing the recived data
	fmt.Printf("this is the recieved data : %s \n", string(body))

	// send the sdp answer back to the client
	if peer_connection == nil {
		http.Error(w, "error in peer intialization", 500)
		return
	}

	// read the body get the sdp and generate the answer
	// then config for the host to works with the config
	var offer webrtc.SessionDescription
	if err := json.Unmarshal(body, &offer); err != nil {
		http.Error(w, "Error : failed to parse sdp offer", http.StatusBadRequest)
		return
	}

	// once parse set the host with the sdp config by the user
	if err := peer_connection.SetRemoteDescription(offer); err != nil {
		http.Error(w, "Error : failed to set remote descriptions", http.StatusInternalServerError)
		return
	}

	// create a answer instance
	answer, err := peer_connection.CreateAnswer(nil)
	if err != nil {
		http.Error(w, "error : creating the answer", http.StatusInternalServerError)
		return
	}

	// this is used for the server setup for communications
	err = peer_connection.SetLocalDescription(answer)
	if err != nil {
		http.Error(w, "error : setting the local description on the host", http.StatusInternalServerError)
		return
	}

	complete_gather := webrtc.GatheringCompletePromise(peer_connection)
	<-complete_gather
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(peer_connection.LocalDescription())
}
