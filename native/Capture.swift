// Screen capture and word-range geometry adapted from Flow Translate, MIT.
// Copyright (c) 2026 Flow Translate contributors.
// See NOTICE.
import AppKit
import ScreenCaptureKit
import Vision

struct LookupContext: Codable {
    var selection: String
    var before: String
    var after: String
    var sourceTitle: String
    var sectionTitle = ""
}

struct WordBounds: Codable {
    var x: Double
    var y: Double
    var width: Double
    var height: Double
}

struct CaptureResult: Codable {
    var context: LookupContext
    var bounds: WordBounds
}

enum CaptureError: LocalizedError {
    case message(String)
    var errorDescription: String? { if case let .message(text) = self { return text }; return nil }
}

@MainActor
func captureWord(x: CGFloat, y: CGFloat, hostPID: Int32) async throws -> CaptureResult {
    guard CGPreflightScreenCaptureAccess() else {
        throw CaptureError.message("Allow Screen Recording in Settings, then quit and reopen InlineExplain.")
    }
    let title = NSWorkspace.shared.frontmostApplication?.localizedName ?? ""
    let content = try await SCShareableContent.excludingDesktopWindows(false, onScreenWindowsOnly: true)
    let pointer = CGPoint(x: x, y: y)
    guard let display = content.displays.first(where: { $0.frame.contains(pointer) }) else {
        throw CaptureError.message("The pointer is not on an available display.")
    }
    let rect = CGRect(x: x - 460, y: y - 220, width: 920, height: 440).intersection(display.frame)
    let ownApps = content.applications.filter { $0.processID == hostPID || $0.processID == ProcessInfo.processInfo.processIdentifier }
    let filter = SCContentFilter(display: display, excludingApplications: ownApps, exceptingWindows: [])
    let screen = NSScreen.screens.first { ($0.deviceDescription[NSDeviceDescriptionKey("NSScreenNumber")] as? NSNumber)?.uint32Value == display.displayID }
    let scale = screen?.backingScaleFactor ?? 2
    let config = SCStreamConfiguration()
    config.sourceRect = CGRect(x: rect.minX - display.frame.minX, y: rect.minY - display.frame.minY, width: rect.width, height: rect.height)
    config.width = max(1, Int(rect.width * scale))
    config.height = max(1, Int(rect.height * scale))
    config.showsCursor = false
    config.captureResolution = .best
    let image = try await SCScreenshotManager.captureImage(contentFilter: filter, configuration: config)
    return try await Task.detached(priority: .userInitiated) {
        try recognizeWord(image, screenRect: rect, pointer: pointer, title: title)
    }.value
}

func recognizeWord(_ image: CGImage, screenRect: CGRect, pointer: CGPoint, title: String) throws -> CaptureResult {
    let request = VNRecognizeTextRequest()
    request.recognitionLevel = .accurate
    request.automaticallyDetectsLanguage = true
    request.usesLanguageCorrection = true
    let supported = try request.supportedRecognitionLanguages()
    request.recognitionLanguages = ["en-US", "zh-Hans", "zh-Hant", "ja-JP", "ko-KR", "fr-FR", "de-DE", "es-ES"].filter { supported.contains($0) }
    try VNImageRequestHandler(cgImage: image).perform([request])
    func localRect(_ box: CGRect) -> CGRect {
        CGRect(x: box.minX * screenRect.width, y: (1 - box.maxY) * screenRect.height,
               width: box.width * screenRect.width, height: box.height * screenRect.height)
    }
    let entries = (request.results ?? []).compactMap { observation -> (VNRecognizedText, TextLine)? in
        guard let candidate = observation.topCandidates(1).first else { return nil }
        return (candidate, TextLine(text: candidate.string, rect: localRect(observation.boundingBox)))
    }
    let point = CGPoint(x: pointer.x - screenRect.minX, y: pointer.y - screenRect.minY)
    let candidates = entries.filter { $0.1.rect.insetBy(dx: -3, dy: -5).contains(point) }
    guard let entry = candidates.min(by: { abs($0.1.rect.midY - point.y) < abs($1.1.rect.midY - point.y) }) else {
        throw CaptureError.message("No word found. Point directly at readable text and press Option+S again.")
    }
    let candidate = entry.0
    var picked: (String, CGRect, NSRange)?
    candidate.string.enumerateSubstrings(in: candidate.string.startIndex..<candidate.string.endIndex, options: [.byWords, .localized]) { substring, range, _, _ in
        guard substring != nil else { return }
        let expanded = compoundRange(range, in: candidate.string)
        guard let observation = try? candidate.boundingBox(for: expanded) else { return }
        let box = localRect(observation.boundingBox)
        if box.insetBy(dx: -3, dy: -5).contains(point) {
            picked = (String(candidate.string[expanded]), box, NSRange(expanded, in: candidate.string))
        }
    }
    guard let picked else { throw CaptureError.message("Point directly at a word, then press Option+S again.") }
    let (before, after) = ContextWindow.extract(anchor: entry.1, wordRange: picked.2, lines: entries.map(\.1))
    return CaptureResult(context: LookupContext(selection: picked.0, before: before, after: after, sourceTitle: title),
                         bounds: WordBounds(x: screenRect.minX + picked.1.minX, y: screenRect.minY + picked.1.minY,
                                            width: picked.1.width, height: picked.1.height))
}

private func compoundRange(_ range: Range<String.Index>, in text: String) -> Range<String.Index> {
    func latin(_ c: Character) -> Bool { c.isASCII && (c.isLetter || c.isNumber || "-_’'".contains(c)) }
    guard text[range].allSatisfy(latin) else { return range }
    var lower = range.lowerBound, upper = range.upperBound
    while lower > text.startIndex, latin(text[text.index(before: lower)]) { lower = text.index(before: lower) }
    while upper < text.endIndex, latin(text[upper]) { upper = text.index(after: upper) }
    return lower..<upper
}
