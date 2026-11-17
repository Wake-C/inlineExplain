// Package speech fetches pronunciation audio for a single English word.
package speech

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

const (
	// MaxAudioBytes bounds what one response may hold in memory.
	MaxAudioBytes = 2 << 20

	maxWordLength   = 48
	defaultEndpoint = "https://dict.youdao.com/dictvoice"
	termEdges       = "\"'.,;:!?()[]{}<>«»„“”‘’—–…·"
)

var (
	// ErrNoAudio means the provider answered without usable audio. The caller
	// can then speak the word with the macOS voice instead.
	ErrNoAudio = errors.New("word audio is not available")
	plainWord  = regexp.MustCompile(`^[A-Za-z]+(?:['-][A-Za-z]+)*$`)
)

// WordTerm returns the only text this client may send: one plain English word.
// Quotes and surrounding punctuation are trimmed; phrases and sentences are
// rejected, so reading context never reaches the dictionary service.
func WordTerm(text string) string {
	term := strings.Trim(strings.TrimSpace(text), termEdges)
	term = strings.ReplaceAll(term, "\u2019", "'")
	term = strings.ReplaceAll(term, "\u2018", "'")
	if len(term) > maxWordLength || !plainWord.MatchString(term) {
		return ""
	}
	return term
}

// VoiceType maps an accent to the provider's voice parameter: 1 is UK, 2 is US.
func VoiceType(accent string) (string, bool) {
	switch accent {
	case "US":
		return "2", true
	case "UK":
		return "1", true
	default:
		return "", false
	}
}

// Youdao reads single-word audio from the public dictionary voice endpoint.
// The endpoint is unofficial and key-free.
type Youdao struct {
	Client    *http.Client
	Endpoint  string
	UserAgent string
}

var defaultClient = &http.Client{
	// Keep the spoken word responsive on a slow network: after this bound the
	// app speaks the word with the macOS voice instead.
	Timeout: 4 * time.Second,
	// A redirect must never carry the requested word to a host the app did
	// not choose to send it to.
	CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
}

// Lookup fetches bounded MPEG audio for one word and accent.
func (y Youdao) Lookup(ctx context.Context, term, accent string) ([]byte, error) {
	word := WordTerm(term)
	if word == "" {
		return nil, ErrNoAudio
	}
	voice, ok := VoiceType(accent)
	if !ok {
		return nil, ErrNoAudio
	}
	endpoint := y.Endpoint
	if endpoint == "" {
		endpoint = defaultEndpoint
	}
	address, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("word audio endpoint: %w", err)
	}
	query := address.Query()
	query.Set("audio", word)
	query.Set("type", voice)
	address.RawQuery = query.Encode()

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, address.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("word audio request: %w", err)
	}
	request.Header.Set("Accept", "audio/mpeg")
	agent := y.UserAgent
	if agent == "" {
		agent = "InlineExplain"
	}
	request.Header.Set("User-Agent", agent)

	client := y.Client
	if client == nil {
		client = defaultClient
	}
	response, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("word audio request: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, ErrNoAudio
	}
	if response.ContentLength > MaxAudioBytes {
		return nil, ErrNoAudio
	}
	audio, err := io.ReadAll(io.LimitReader(response.Body, MaxAudioBytes+1))
	if err != nil {
		return nil, fmt.Errorf("word audio response: %w", err)
	}
	if len(audio) == 0 || len(audio) > MaxAudioBytes || !mpegAudio(audio) {
		return nil, ErrNoAudio
	}
	return audio, nil
}

// mpegAudio accepts an MPEG frame sync or an ID3 tag. A provider error body
// must never reach the decoder as audio.
func mpegAudio(data []byte) bool {
	if len(data) < 3 {
		return false
	}
	identity := data[0] == 'I' && data[1] == 'D' && data[2] == '3'
	frame := data[0] == 0xFF && data[1]&0xE0 == 0xE0
	return identity || frame
}
