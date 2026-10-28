package explain

import (
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Context has the same provider fields as Peekturn's LookupContext.
type Context struct {
	Selection    string `json:"selection"`
	Before       string `json:"before"`
	After        string `json:"after"`
	SourceTitle  string `json:"sourceTitle"`
	SectionTitle string `json:"sectionTitle"`
}

var sentenceEnd = regexp.MustCompile(`[.!?]["'”’)\]]*`)

// Normalize applies Peekturn's sentence-edge policy and 30-word limit per side.
// OCR supplies the original word occurrence and its local text, never AI text.
func (c Context) Normalize() Context {
	c.Before = strings.Join(strings.Fields(c.Before), " ")
	c.After = strings.Join(strings.Fields(c.After), " ")
	ends := sentenceEdges(c.Before)
	if len(ends) > 0 {
		tail := strings.TrimSpace(c.Before[ends[len(ends)-1]:])
		if len(strings.Fields(tail)) >= 2 {
			c.Before = tail
		}
	}
	ends = sentenceEdges(c.After)
	if len(ends) > 0 {
		head := strings.TrimSpace(c.After[:ends[0]])
		if len(strings.Fields(head)) >= 2 {
			c.After = head
		}
	}
	before, after := strings.Fields(c.Before), strings.Fields(c.After)
	if len(before) > 30 {
		before = before[len(before)-30:]
	}
	if len(after) > 30 {
		after = after[:30]
	}
	c.Before, c.After = strings.Join(before, " "), strings.Join(after, " ")
	return c
}

func sentenceEdges(text string) []int {
	var result []int
	for _, match := range sentenceEnd.FindAllStringIndex(text, -1) {
		rest := text[match[1]:]
		if strings.TrimSpace(rest) == "" {
			result = append(result, match[1])
			continue
		}
		first, _ := utf8.DecodeRuneInString(rest)
		if !unicode.IsSpace(first) {
			continue
		}
		rest = strings.TrimLeft(strings.TrimSpace(rest), "\"'“”‘’(")
		first, _ = utf8.DecodeRuneInString(rest)
		if first >= 'A' && first <= 'Z' {
			result = append(result, match[1])
		}
	}
	return result
}
