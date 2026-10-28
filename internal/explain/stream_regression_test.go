package explain

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func responseEvent(answer Explanation, extra map[string]any) string {
	content, _ := json.Marshal(answer)
	envelope := map[string]any{"choices": []any{map[string]any{"delta": map[string]string{"content": string(content)}}}}
	for key, value := range extra {
		envelope[key] = value
	}
	encoded, _ := json.Marshal(envelope)
	return "data: " + string(encoded) + "\n\ndata: [DONE]\n\n"
}

type transportErrorAfterDone struct{}

func (transportErrorAfterDone) Read([]byte) (int, error) {
	return 0, errors.New("transport reset after terminal marker")
}

func TestDoneStopsBeforeFurtherTransportReads(t *testing.T) {
	reader := io.MultiReader(strings.NewReader(responseEvent(validAnswer("en"), nil)), transportErrorAfterDone{})
	answer, err := readStream(reader, "en", func(Explanation) {})
	if err != nil || answer.Term != "bank" {
		t.Fatalf("Completed response was rejected: %v", err)
	}
}

func TestNullProviderErrorIsSuccess(t *testing.T) {
	answer, err := readStream(strings.NewReader(responseEvent(validAnswer("en"), map[string]any{"error": nil})), "en", func(Explanation) {})
	if err != nil || answer.Term != "bank" {
		t.Fatalf("Null error rejected a successful response: %v", err)
	}
	_, err = readStream(strings.NewReader(responseEvent(validAnswer("en"), map[string]any{"error": map[string]string{"message": "test failure"}})), "en", func(Explanation) {})
	if err == nil {
		t.Fatal("A real provider error must still fail")
	}
}

func TestDeepSeekThinkingDisabled(t *testing.T) {
	for _, model := range []string{"deepseek-chat", "test-model"} {
		t.Run(model, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var body map[string]json.RawMessage
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Error(err)
				}
				if model == "deepseek-chat" {
					var thinking struct {
						Type string `json:"type"`
					}
					_ = json.Unmarshal(body["thinking"], &thinking)
					if thinking.Type != "disabled" {
						t.Error("DeepSeek thinking was not disabled")
					}
				} else if _, exists := body["thinking"]; exists {
					t.Error("A provider-specific parameter leaked into an unrelated model request")
				}
				w.Header().Set("Content-Type", "text/event-stream")
				_, _ = w.Write([]byte(responseEvent(validAnswer("en"), nil)))
			}))
			defer server.Close()
			_, err := Explain(context.Background(), Config{Endpoint: server.URL, Model: model, Language: "en", Temperature: .3}, "test-only", Context{Selection: "bank", Before: "She sat on the", After: "of the river."}, func(Explanation) {})
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}
