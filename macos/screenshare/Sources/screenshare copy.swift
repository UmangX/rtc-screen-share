import ScreenCaptureKit

class StreamDelegate: NSObject, SCStreamDelegate {
    @available(macOS 12.3, *)
    func stream(
        _ stream: SCStream, didOutputSampleBuffer sampleBuffer: CMSampleBuffer,
        of outputType: SCStreamOutputType
    ) {
        if outputType == .screen {
            print("Got screen frame at \(Date())")
        }
    }
}

@main
struct screencapture {
    static func main() async {

        let app = NSApplication.shared
        app.setActivationPolicy(.prohibited)
        do {
            // first we have to get the display for this stream to focus on
            if #available(macOS 12.3, *) {
                let displays = try await SCShareableContent.excludingDesktopWindows(
                    false, onScreenWindowsOnly: false)
                // got the displays we need codes for the displays
                // this guard is like reverse if else for the variable
                guard let display = displays.displays.first else {
                    print("finding the displays has being a problem")
                    return
                }
                print("using the display \(display.displayID)")

                let config = SCStreamConfiguration()
                config.showsCursor = true
                if #available(macOS 13.0, *) {
                    config.capturesAudio = false
                } else {
                    print("the audio thing is not working")
                }
                let stream = SCStream(
                    filter: SCContentFilter(
                        display: display, excludingApplications: [], exceptingWindows: []),
                    configuration: config, delegate: StreamDelegate())
                try await stream.startCapture()
                print("Started capture. Press Ctrl+C to stop.")

                RunLoop.main.run()

            } else {
                print("the macos version is not working")
            }
        } catch let error {
            print("this is the error that has happened \(error)")
        }
    }
}
