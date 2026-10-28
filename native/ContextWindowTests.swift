import Foundation
import CoreGraphics

@main
struct ContextWindowTests {
    static func main() {
        func line(_ text: String, _ y: CGFloat, x: CGFloat = 50) -> TextLine {
            TextLine(text: text, rect: CGRect(x: x, y: y, width: 500, height: 20))
        }
        let lines = [line("An earlier sentence. The bank was", 0), line("covered with wild flowers along", 30), line("the river. We went home.", 60), line("A different column must stay out.", 30, x: 700)]
        let range = (lines[0].text as NSString).range(of: "bank")
        let (before, after) = ContextWindow.extract(anchor: lines[0], wordRange: range, lines: lines)
        precondition(before == "An earlier sentence. The ")
        precondition(after == " was covered with wild flowers along the river. We went home.")
        let repeated = line("We went to the bank. The river bank was green.", 0)
        let lastRange = (repeated.text as NSString).range(of: "bank", options: .backwards)
        let (lastBefore, lastAfter) = ContextWindow.extract(anchor: repeated, wordRange: lastRange, lines: [repeated])
        precondition(lastBefore == "We went to the bank. The river ")
        precondition(lastAfter == " was green.")
        print("PASS: wrapped context, columns and the exact repeated-word occurrence")
    }
}
