package main

import (
	"fmt"
	"net/http"
)

const (
	webPort = ":8080"
)

func main() {
	fmt.Println("This is running the webrtc server ...")
	http.ListenAndServe(webPort, http.FileServer(http.Dir("./static")))
}
