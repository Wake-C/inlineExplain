import AppKit
import CoreText
import Foundation

@main
struct OCRTests {
    static func main() throws {
        let text = "We went to the bank. The river bank was green."
        let width = 1100, height = 160
        let canvas = CGContext(data: nil, width: width, height: height, bitsPerComponent: 8,
                               bytesPerRow: width * 4, space: CGColorSpaceCreateDeviceRGB(),
                               bitmapInfo: CGImageAlphaInfo.premultipliedLast.rawValue)!
        canvas.setFillColor(NSColor.white.cgColor)
        canvas.fill(CGRect(x: 0, y: 0, width: width, height: height))
        let attributed = NSAttributedString(string: text, attributes: [.font: NSFont.systemFont(ofSize: 36), .foregroundColor: NSColor.black])
        let line = CTLineCreateWithAttributedString(attributed)
        canvas.textPosition = CGPoint(x: 30, y: 75)
        CTLineDraw(line, canvas)
        let image = canvas.makeImage()!
        let last = (text as NSString).range(of: "bank", options: .backwards)
        let start = CTLineGetOffsetForStringIndex(line, last.location, nil)
        let end = CTLineGetOffsetForStringIndex(line, last.location + last.length, nil)
        let screen = CGRect(x: 100, y: 100, width: width, height: height)
        let pointer = CGPoint(x: screen.minX + 30 + (start+end)/2, y: screen.minY + CGFloat(height) - 75 - 12)
        let result = try recognizeWord(image, screenRect: screen, pointer: pointer, title: "Synthetic acceptance sample")
        precondition(result.context.selection == "bank", "Apple Vision must find the pointed word")
        precondition(result.context.before.contains("The river "), "The context must come from the last bank occurrence")
        precondition(result.context.after.contains("was green"), "The following text must come from the same occurrence")
        print("PASS: real Apple Vision OCR, character-range hit test and repeated-word context")
    }
}
