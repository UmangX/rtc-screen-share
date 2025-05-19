# 7xscreenshare

building local network screen share with realtime streaming 

 - has to be single command for the host os 
 - has be accessible through browser ( no client software required ) 
 
development choices 

 - macos is a walled garden this is bad for me cause i don't want to deepdive into avfounations 
 - i tried coding using screencapture kit but the permissions is very crumblesome to configure 
 - swift is a very different language 
 - current solution for live screen recording and encoding is through ffmpeg 
 - ffmpeg is using the libav file for getting the frames for screen this is not a huge overhead for host 
 - the last method of multiple per second screenshot and then encoding it is not very fast there is a huge latency 
 
 current status 
 
 - python mlss libraries is used for screenshot and then encoded which is working fine 
 - used ffmpeg-dash for hls but this setup is not realtime 
 - dash hls method is good for monitoring maybe but not fast enough like screensharing 
 - i did consider rtmp protocol but the point of browser streaming is not working for rtmp we need client program to decode the stream
 -  we can do transmuxing of converting rtmp feed into dash then using hls but this is not robust 
 - looking into webrtc using go for this + ffmpeg for encoding this stream into  browser 
 
 why ? 

- this program / project is just a side project to understand how protocols works and encoding and all that  

