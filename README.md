# 7xscreenshare

local realtime screen sharing for macos and windows


learning the macos system :
    things that will works : 
        -   macos is walled garden but the things only works with obj-c and swift and persmissions are pain to handled 
        -   the only possible way to capture the display is the framework called avfoundation this is essentail
        -   creating custom screen recording system for macos is out of bound for me now 
        -   was thinking about using rtmp for streaming it over localhost but rtmp is working with browsers directly 
        -   hence will use the most sane method of using ffmpeg and use libav.. method for capturing the screen 
        -   the output is then piped to the server using the HLS/DASH method 
        -   HLS - HTTP Live Streaming / DASH -  Dynamic Adaptive Streaming over HTTP 
        -   this is done through http 
        -   update : this is now use go for testing instead of rust because of my familarity over the lang 
