package explain

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func TestContextMatchesPeekturnSentenceEdges(t *testing.T) {
	for _, test := range []struct{ before, after, wantBefore, wantAfter string }{
		{"the bookChapter 1, passage 1. She was", "to leave the familiar streets behind. A new city", "She was", "to leave the familiar streets behind."},
		{"Earlier sentence here. was", "measured at 3.5 units", "Earlier sentence here. was", "measured at 3.5 units"},
		{"  a quiet\n window   of text ", "the U.S. market", "a quiet window of text", "the U.S. market"},
	} {
		got := (Context{Selection: "ready", Before: test.before, After: test.after}).Normalize()
		if got.Before != test.wantBefore || got.After != test.wantAfter {
			t.Fatalf("wrong context: %#v", got)
		}
	}
	got := (Context{Before: strings.Repeat("before ", 60), After: strings.Repeat("after ", 60)}).Normalize()
	if len(strings.Fields(got.Before)) != 30 || len(strings.Fields(got.After)) != 30 {
		t.Fatal("context must be limited to 30 words on each side")
	}
}

func validAnswer(language string) Explanation {
	translation := ""
	if language == "zh" {
		translation = "她坐在河岸上。"
	}
	return Explanation{Term: "bank", PartOfSpeech: "noun", Meaning: "The land beside a river.", Context: "She sits on the land next to the water.", Examples: []Example{{Text: "She sat on the river bank.", Translation: translation}}}
}

func stream(t *testing.T, answer Explanation, done bool) string {
	t.Helper()
	encoded, _ := json.Marshal(answer)
	var result strings.Builder
	// Split inside UTF-8 and JSON escapes, as network chunks may do.
	for _, r := range string(encoded) {
		data, _ := json.Marshal(map[string]any{"choices": []any{map[string]any{"delta": map[string]string{"content": string(r)}}}})
		fmt.Fprintf(&result, "data: %s\r\n\r\n", data)
	}
	if done {
		result.WriteString("data: [DONE]\r\n\r\n")
	}
	return result.String()
}

func TestStreamingExplanationAndIncompleteResponse(t *testing.T) {
	want := validAnswer("zh")
	got, err := readStream(strings.NewReader(stream(t, want, true)), "zh", func(Explanation) {})
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("stream failed: %#v %v", got, err)
	}
	_, err = readStream(strings.NewReader(stream(t, want, false)), "zh", func(Explanation) {})
	if err == nil {
		t.Fatal("an interrupted stream must not be reported complete")
	}
	if Partial([]byte(`{"meaning":"half`)).Meaning != "" {
		t.Fatal("unfinished strings must not be shown")
	}
}

func TestRequestSendsExactContextAndCustomPrompt(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			t.Errorf("wrong provider path: %s", r.URL.Path)
		}
		var body struct {
			Messages       []struct{ Role, Content string }
			ResponseFormat map[string]any `json:"response_format"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)
		}
		if !strings.Contains(body.Messages[0].Content, "Use very simple English") || !strings.Contains(body.Messages[0].Content, "Required JSON schema") {
			t.Error("custom prompt or schema was lost")
		}
		var passage Context
		_ = json.Unmarshal([]byte(body.Messages[1].Content), &passage)
		if passage.Selection != "bank" || passage.Before != "She sat on the" || passage.After != "of the river." {
			t.Errorf("context was changed: %#v", passage)
		}
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte(stream(t, validAnswer("en"), true)))
	}))
	defer server.Close()
	_, err := Explain(context.Background(), Config{Endpoint: server.URL, Model: "test", Language: "en", Prompt: "Use very simple English", Temperature: .3}, "test-only", Context{Selection: "bank", Before: "She sat on the", After: "of the river."}, func(Explanation) {})
	if err != nil {
		t.Fatal(err)
	}
}

func TestSchemaAndLanguageValidation(t *testing.T) {
	answer := validAnswer("en")
	encoded, _ := json.Marshal(answer)
	if _, err := Parse(encoded, "en"); err != nil {
		t.Fatal(err)
	}
	if _, err := Parse(encoded, "zh"); err == nil {
		t.Fatal("Chinese examples require a translation")
	}
	answer.Synonyms = "land\nshore"
	encoded, _ = json.Marshal(answer)
	if _, err := Parse(encoded, "en"); err == nil {
		t.Fatal("multiline relations should be rejected")
	}
	if !strings.Contains(DefaultPrompt("en"), "Use elementary English") || !strings.Contains(DefaultPrompt("zh"), "只解释用户选中的英文单词或短语") {
		t.Fatal("Peekturn defaults are missing")
	}
}

func TestEndpointNormalization(t *testing.T) {
	for _, input := range []string{"https://example.com", "https://example.com/v1", "https://example.com/v1/chat/completions"} {
		u, err := Endpoint(input)
		if err != nil || u.Path != "/v1/chat/completions" {
			t.Fatalf("invalid normalization: %s %v", input, err)
		}
	}
	for _, input := range []string{"http://example.com", "https://user:password@example.com", "https://example.com?key=secret"} {
		if _, err := Endpoint(input); err == nil {
			t.Fatal("invalid provider endpoint accepted")
		}
	}
}

func TestProviderFailureAndCancellation(t *testing.T) {
	t.Run("status does not expose response body", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte("private-provider-detail"))
		}))
		defer server.Close()
		_, err := Explain(context.Background(), Config{Endpoint: server.URL, Model: "test", Language: "en"}, "test-only", Context{Selection: "bank"}, func(Explanation) {})
		if err == nil || !strings.Contains(err.Error(), "401") || strings.Contains(err.Error(), "private-provider-detail") {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	t.Run("dismiss cancels request", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		_, err := Explain(ctx, Config{Endpoint: "http://127.0.0.1:1", Model: "test", Language: "en"}, "test-only", Context{Selection: "bank"}, func(Explanation) {})
		if err != context.Canceled {
			t.Fatalf("expected cancellation, got %v", err)
		}
	})
}
