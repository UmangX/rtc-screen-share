__how to find the users for the connections ?__\
STUN and TURN servers are used for bypassing the nat and finding the ip address for peer connections\
a STUN server isa public server on the internet when the clients to establish connections it first\
sends a binding request to the stun server the stun server and the server send the binding respone back to the client including\
this public ip address and port hen the other peer tries connections tho that public ip and port your router having recently\
sent data out from that specific port often has a hole or mapping open in its nat table this ( hole punching in the internet)
___

__why there is no need for stun and turn server in this project__\
my main motive is implementing this streaming for local networks and this is done through sharing data among same subnet\
hence we don't need ice server for stun and turn  for find the works this can be done by just getting ip address and data\
and there is no NAT for this networks
___

__how to gain the video capture or the screen__\
this is done through ffmpeg as solution for os screen capture this is send send to the stdout the server then reads and sends to the client

```mermaid
graph TD
    subgraph Initialization
        A[Go Server Starts] --> B[Serves client.html via HTTP]
        B --> C[Listens for WebSocket connections on /ws]
    end

    subgraph Client Connects
        D[Client (Browser) Loads client.html] --> E[Browser JS: Creates WebSocket to ws://localhost:8080/ws]
        E --> F{WebSocket Connection Established?}
        F -- Yes --> G[Go Server: new Websocket connection in handleWebsocket]
        F -- No --> H[Error: WebSocket connection failed]
    end

    subgraph WebRTC Setup (Signaling)
        G --> I[Go Server: Create new RTCPeerConnection (no iceServers)]
        I --> J[Go Server: Create and Add Local VideoTrack]
        J --> K[Go Server: Set onICECandidate handler (sends candidates to client via WS)]

        L[Browser JS: Create new RTCPeerConnection (no iceServers)]
        L --> M[Browser JS: Set onTrack handler (attaches stream to <video>)]
        M --> N[Browser JS: Set onICECandidate handler (sends candidates to server via WS)]

        O[Browser JS: Create SDP Offer]
        O --> P[Browser JS: Set Local Description (Offer)]
        P --> Q[Browser JS: Send Offer to Go Server via WebSocket]

        R[Go Server: Receive Offer via WebSocket]
        R --> S[Go Server: Set Remote Description (Offer)]
        S --> T[Go Server: Create SDP Answer]
        T --> U[Go Server: Wait for ICE Gathering Complete (optional but good practice)]
        U --> V[Go Server: Set Local Description (Answer)]
        V --> W[Go Server: Send Answer to Client via WebSocket]
    end

    subgraph ICE Candidate Exchange & Connection
        X[Both: onICECandidate fires for local candidates] --> Y{Candidate Type?}
        Y -- Local IP --> Z[Send Candidate via WebSocket to other peer]
        Z --> AA[Receive Candidate via WebSocket]
        AA --> BB[Add Candidate to RTCPeerConnection]
        BB --> CC[ICE Negotiation Continues]
        CC --> DD{ICE Connection State: Connected/Completed?}
        DD -- Yes --> EE[Direct P2P Connection Established over Local Network]
        DD -- No (Failed) --> FF[WebRTC Connection Failed]
    end

    subgraph Media Streaming
        EE --> GG[Go Server: Start FFmpeg Process for Screen Capture]
        GG --> HH[Go Server: Read Encoded Frames (e.g., VP8) from FFmpeg stdout]
        HH --> II[Go Server: Write Encoded Frames to RTCPeerConnection's VideoTrack (using WriteSample)]
        II --> JJ[Browser: RTCPeerConnection Receives RTP Packets]
        JJ --> KK[Browser: Decodes Video]
        KK --> LL[Browser: Displays Video in <video> element (via onTrack)]

        MM[Continuous Loop] --> HH
        MM --> JJ
    end

    subgraph Disconnection
        NN[Client Closes Tab/Navigates Away OR Go Server Shuts Down] --> OO[WebSocket Connection Closes]
        OO --> PP[Go Server: RTCPeerConnection Closes]
        PP --> QQ[Go Server: FFmpeg Process Terminated]
        QQ --> RR[Stream Ends]
    end
```
