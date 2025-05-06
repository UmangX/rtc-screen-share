import Foundation
import AppKit
import ScreenCaptureKit

class StreamDelegate: NSObject, SCStreamDelegate {
    @available(macOS 12.3, *)
    func stream(_ stream: SCStream, didOutputSampleBuffer sampleBuffer: CMSampleBuffer, of outputType: SCStreamOutputType) {
        if outputType == .screen {
            print("Got screen frame at \(Date())")
        }
    }
}

@main
struct ScreenCapture {
    static func main() async {
        let app = NSApplication.shared
        app.setActivationPolicy(.prohibited) // No Dock icon

        do {
            if #available(macOS 12.3, *) {
                let content = try await SCShareableContent.excludingDesktopWindows(false, onScreenWindowsOnly: false)

                guard let display = content.displays.first else {
                    print("No display found.")
                    return
                }

                print("Capturing display ID: \(display.displayID)")

                let config = SCStreamConfiguration()
                config.showsCursor = true

                if #available(macOS 13.0, *) {
                    config.capturesAudio = false
                }

                let stream = SCStream(
                    filter: SCContentFilter(display: display, excludingApplications: [], exceptingWindows: []),
                    configuration: config,
                    delegate: StreamDelegate()
                )


try await stream.startCapture()
print("Stream started. Press Ctrl+C to stop.")

await Task.never()

            } else {
                print("macOS 12.3+ is required.")
            }
        } catch {
            print("Error starting stream: \(error)")
        }
    }
}

