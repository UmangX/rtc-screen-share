// swift-tools-version:5.5
import PackageDescription

let package = Package(
    name: "screenshare",
    platforms: [
        .macOS(.v12)
    ],
    dependencies: [],
    targets: [
        .executableTarget(
            name: "screenshare",
            dependencies: [],
            path: "Sources"
        ),
    ]
)
