// Adapted from Flow Translate's ContextSentence.swift, MIT licensed.
// Copyright (c) 2026 Flow Translate contributors.
// See NOTICE. Modified to retain the exact
// selected occurrence plus before/after text for Peekturn's context policy.
import Foundation
import CoreGraphics

struct TextLine {
    let text: String
    let rect: CGRect
}

enum ContextWindow {
    static func extract(anchor: TextLine, wordRange: NSRange, lines: [TextLine]) -> (String, String) {
        var block = [anchor]
        for preceding in [true, false] {
            var edge = anchor
            for _ in 0..<12 {
                let candidates = lines.filter { line in
                    let dy = line.rect.midY - edge.rect.midY
                    guard preceding ? dy < 0 : dy > 0 else { return false }
                    let height = max(edge.rect.height, line.rect.height)
                    let ratio = line.rect.height / max(anchor.rect.height, 1)
                    let gap = preceding ? edge.rect.minY - line.rect.maxY : line.rect.minY - edge.rect.maxY
                    return ratio >= 0.72 && ratio <= 1.38
                        && gap >= -height * 0.15 && gap <= height * 1.1
                        && abs(line.rect.minX - anchor.rect.minX) <= max(24, anchor.rect.height * 2)
                        && line.text.contains(where: \.isLetter)
                }
                guard let next = candidates.min(by: {
                    abs($0.rect.midY - edge.rect.midY) < abs($1.rect.midY - edge.rect.midY)
                }) else { break }
                let upper = preceding ? next : edge
                let lower = preceding ? edge : next
                if upper.rect.width < lower.rect.width * 0.5,
                   !upper.text.contains(where: { ".!?。！？,，:：;；-".contains($0) }),
                   lower.text.first?.isUppercase == true { break }
                if preceding { block.insert(next, at: 0) } else { block.append(next) }
                edge = next
            }
        }
        var text = "", offset = 0
        for line in block {
            if !text.isEmpty { text += separator(after: text, before: line.text) }
            if line.rect == anchor.rect && line.text == anchor.text {
                offset = text.utf16.count + wordRange.location
            }
            text += line.text
        }
        let value = text as NSString
        return (value.substring(to: offset), value.substring(from: offset + wordRange.length))
    }

    private static func separator(after left: String, before right: String) -> String {
        guard let last = left.last, let first = right.first else { return "" }
        if last.isWhitespace || first.isWhitespace || last == "-" { return "" }
        func cjk(_ character: Character) -> Bool {
            character.unicodeScalars.contains {
                (0x2E80...0x9FFF).contains($0.value) || (0xF900...0xFAFF).contains($0.value)
                    || (0x20000...0x3134F).contains($0.value)
            }
        }
        return cjk(last) && cjk(first) ? "" : " "
    }
}
