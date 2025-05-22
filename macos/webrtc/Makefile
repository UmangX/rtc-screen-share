server:
	go run .
capture:
	ffmpeg -re -f avfoundation -pixel_format uyvy422 -i 0 -r 30 -vf "format=yuv420p" -c:v libx264 -preset ultrafast -tune zerolatency -f rtp rtp://127.0.0.1:5004

