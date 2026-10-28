import AppKit
import Foundation

@main
struct Helper {
    @MainActor static func main() async {
        let arguments = CommandLine.arguments
        do {
            switch arguments.dropFirst().first {
            case "capture":
                guard arguments.count == 5, let x = Double(arguments[2]), let y = Double(arguments[3]), let pid = Int32(arguments[4]) else {
                    throw CaptureError.message("Invalid capture coordinates.")
                }
                let result = try await captureWord(x: x, y: y, hostPID: pid)
                let data = try JSONEncoder().encode(result)
                FileHandle.standardOutput.write(data)
            case "permission":
                let allowed = arguments.contains("--request") ? CGRequestScreenCaptureAccess() : CGPreflightScreenCaptureAccess()
                FileHandle.standardOutput.write(try JSONEncoder().encode(["screenRecording": allowed]))
            default: throw CaptureError.message("Unknown capture action.")
            }
        } catch {
            let message = (error as? CaptureError)?.localizedDescription ?? "Screen capture failed. Check Screen Recording access and try again."
            let data = (try? JSONEncoder().encode(["error": message])) ?? Data()
            FileHandle.standardOutput.write(data)
            exit(1)
        }
    }
}
