//go:build live

package main

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"inlineexplain.local/desktop/internal/explain"
)

// Uses the explicitly selected verification profile, with its key in Keychain.
// Passages are synthetic. The key and provider response bodies are not logged.
func TestLiveConfiguredProvider(t *testing.T) {
	service := os.Getenv("INLINEEXPLAIN_LIVE_KEYCHAIN_SERVICE")
	account := os.Getenv("INLINEEXPLAIN_LIVE_KEYCHAIN_ACCOUNT")
	if service == "" || account == "" {
		t.Fatal("Choose the authorized verification profile with INLINEEXPLAIN_LIVE_KEYCHAIN_SERVICE and INLINEEXPLAIN_LIVE_KEYCHAIN_ACCOUNT.")
	}
	raw, err := exec.Command("security", "find-generic-password", "-s", service, "-a", account, "-w").Output()
	if err != nil {
		t.Fatal("The selected verification profile is unavailable.")
	}
	var provider struct {
		Endpoint string `json:"endpoint"`
		APIKey   string `json:"apiKey"`
		Model    string `json:"model"`
	}
	if json.Unmarshal(raw, &provider) != nil || provider.Endpoint == "" || provider.APIKey == "" {
		t.Fatal("The selected profile needs endpoint and apiKey JSON fields.")
	}
	if provider.Model == "" {
		provider.Model = "deepseek-chat"
	}
	for _, test := range []struct{ name, language, before, after, prompt string }{
		{"river-English", "en", "She sat on the", "of the river and watched the water.", ""},
		{"river-Chinese", "zh", "She sat on the", "of the river and watched the water.", ""},
		{"finance-custom", "en", "He went to the", "to borrow money for his new home.", explain.DefaultPrompt("en") + " Use exactly one example. Keep the meaning under fourteen words."},
	} {
		t.Run(test.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 75*time.Second)
			defer cancel()
			answer, err := explain.Explain(ctx, explain.Config{Endpoint: provider.Endpoint, Model: provider.Model, Temperature: 0.3, Language: test.language, Prompt: test.prompt}, provider.APIKey,
				explain.Context{Selection: "bank", Before: test.before, After: test.after, SourceTitle: "Acceptance sample"}, func(explain.Explanation) {})
			if err != nil {
				t.Fatal(err)
			}
			if answer.Term != "bank" {
				t.Fatal("The selected term was not retained.")
			}
			contextText := strings.ToLower(strings.TrimSpace(answer.Context))
			for _, prefix := range []string{"here it names", "it describes", "it means", "in this passage", "in this sentence", "这里用于", "此处指", "这里指", "这里表示", "它描述", "这个词"} {
				if strings.HasPrefix(contextText, prefix) {
					t.Fatalf("The contextual explanation still uses a redundant prefix: %s", prefix)
				}
			}
			if test.prompt != "" && len(answer.Examples) != 1 {
				t.Fatal("The custom one-example instruction was not followed.")
			}
			combined := strings.ToLower(answer.Meaning + " " + answer.Context)
			if test.name == "river-English" && !strings.Contains(combined, "river") && !strings.Contains(combined, "water") {
				t.Fatal("The explanation did not use the river context.")
			}
			if test.name == "finance-custom" && !strings.Contains(combined, "money") && !strings.Contains(combined, "loan") {
				t.Fatal("The explanation did not use the financial context.")
			}
			t.Logf("Live provider passed: language=%s, examples=%d, contextual meaning and response fields validated", test.language, len(answer.Examples))
		})
	}
}
