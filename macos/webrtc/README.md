__how to find the users for the connections ?__

STUN and TURN servers are used for bypassing the nat and finding the ip address for peer connections

a STUN server isa public server on the internet when the clients to establish connections it first

sends a binding request to the stun server the stun server and the server send the binding respone back to the client including

this public ip address and port hen the other peer tries connections tho that public ip and port your router having recently

sent data out from that specific port often has a hole or mapping open in its nat table this ( hole punching in the internet)

___

__why there is no need for stun and turn server in this project__

my main motive is implementing this streaming for local networks and this is done through sharing data among same subnet

hence we don't need ice server for stun and turn  for find the works this can be done by just getting ip address and data

and there is no NAT for this networks

___

__how to gain the video capture or the screen__

this is done through ffmpeg as solution for os screen capture this is send send to the stdout the server then reads and encodes this
