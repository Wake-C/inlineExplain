package speech

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestWordTermAcceptsOnlyOnePlainWord(t *testing.T) {
	for _, input := range []string{"computer", " don't ", "state-of-the-art", `"computer,"`, "computer’s", "“word”"} {
		if WordTerm(input) == "" {
			t.Errorf("WordTerm rejected a single word: %q", input)
		}
	}
	if got := WordTerm("computer’s"); got != "computer's" {
		t.Errorf("Curly apostrophe was not normalized: %q", got)
	}
	for _, input := range []string{"", "   ", "hello world", "com puter", "电脑", "word123", "a/b", strings.Repeat("a", maxWordLength+1)} {
		if got := WordTerm(input); got != "" {
			t.Errorf("WordTerm accepted more than one plain word: %q became %q", input, got)
		}
	}
}

func TestVoiceTypeDefaultsToTheAmericanParameter(t *testing.T) {
	if voice, ok := VoiceType("US"); !ok || voice != "2" {
		t.Fatal("US must use voice type 2")
	}
	if voice, ok := VoiceType("UK"); !ok || voice != "1" {
		t.Fatal("UK must use voice type 1")
	}
	if _, ok := VoiceType(""); ok {
		t.Fatal("An unknown accent must not select a provider voice")
	}
}

func TestYoudaoLookupRequestsOneWord(t *testing.T) {
	audio := []byte{0xFF, 0xFB, 0x90, 0x64}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/dictvoice" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("audio"); got != "computer" {
			t.Errorf("Unexpected word: %q", got)
		}
		if got := r.URL.Query().Get("type"); got != "2" {
			t.Errorf("Unexpected voice type: %q", got)
		}
		if got := r.Header.Get("Accept"); got != "audio/mpeg" {
			t.Errorf("Unexpected Accept header: %q", got)
		}
		if got := r.Header.Get("User-Agent"); got != "test-agent" {
			t.Errorf("Unexpected User-Agent: %q", got)
		}
		_, _ = w.Write(audio)
	}))
	defer server.Close()

	got, err := Youdao{Client: server.Client(), Endpoint: server.URL + "/dictvoice", UserAgent: "test-agent"}.Lookup(context.Background(), "computer,", "US")
	if err != nil {
		t.Fatalf("Lookup returned an error: %v", err)
	}
	if string(got) != string(audio) {
		t.Fatalf("Lookup returned different audio: %v", got)
	}
}

func TestYoudaoLookupRejectsUnusableAnswers(t *testing.T) {
	cases := map[string]http.HandlerFunc{
		"HTTP error": func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, "no audio", http.StatusNotFound)
		},
		"HTML body": func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte("<html>not audio</html>"))
		},
		"empty body": func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		},
		"oversized body": func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Length", "2097153")
			w.WriteHeader(http.StatusOK)
		},
	}
	for name, handler := range cases {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(handler)
			defer server.Close()
			got, err := Youdao{Client: server.Client(), Endpoint: server.URL}.Lookup(context.Background(), "computer", "US")
			if !errors.Is(err, ErrNoAudio) || got != nil {
				t.Fatalf("Lookup accepted an unusable answer: audio=%v err=%v", got, err)
			}
		})
	}
}

func TestYoudaoLookupRejectsAStreamPastTheLimit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.(http.Flusher).Flush()
		chunk := make([]byte, 64<<10)
		chunk[0], chunk[1] = 0xFF, 0xFB
		written := 0
		for written <= MaxAudioBytes {
			n, err := w.Write(chunk)
			if err != nil {
				return
			}
			written += n
		}
	}))
	defer server.Close()

	if _, err := (Youdao{Client: server.Client(), Endpoint: server.URL}).Lookup(context.Background(), "computer", "US"); !errors.Is(err, ErrNoAudio) {
		t.Fatalf("A stream past the limit must count as no audio, got %v", err)
	}
}

func TestYoudaoLookupDoesNotFollowRedirects(t *testing.T) {
	var reached atomic.Bool
	target := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		reached.Store(true)
	}))
	defer target.Close()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, target.URL, http.StatusFound)
	}))
	defer server.Close()
	client := &http.Client{
		Timeout:       time.Second,
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}

	if _, err := (Youdao{Client: client, Endpoint: server.URL}).Lookup(context.Background(), "computer", "US"); !errors.Is(err, ErrNoAudio) {
		t.Fatalf("A redirect must count as no audio, got %v", err)
	}
	if reached.Load() {
		t.Fatal("A redirect sent the requested word to another host")
	}
}

func TestYoudaoLookupNeverSendsAPhrase(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requests.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	for _, term := range []string{"hello world", "", "你好"} {
		if _, err := (Youdao{Client: server.Client(), Endpoint: server.URL}).Lookup(context.Background(), term, "US"); !errors.Is(err, ErrNoAudio) {
			t.Fatalf("Lookup accepted %q: %v", term, err)
		}
	}
	if requests.Load() != 0 {
		t.Fatal("A phrase or invalid word reached the dictionary service")
	}
}
