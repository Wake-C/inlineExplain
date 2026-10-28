package explain

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/idna"
)

type Config struct {
	Endpoint    string  `json:"endpoint"`
	Model       string  `json:"model"`
	Temperature float64 `json:"temperature"`
	Language    string  `json:"language"`
	Prompt      string  `json:"prompt"`
}

func Endpoint(value string) (*url.URL, error) {
	u, err := url.Parse(strings.TrimSpace(value))
	if err != nil || u.Hostname() == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || (u.Scheme != "https" && !(u.Scheme == "http" && (u.Hostname() == "localhost" || u.Hostname() == "127.0.0.1" || u.Hostname() == "::1"))) {
		return nil, errors.New("Enter an HTTPS provider URL, or an HTTP localhost URL.")
	}
	// Match net/http's IDNA lookup before any case folding. Unicode lowercasing
	// alone can collapse different DNS origins (for example i and capital İ).
	if strings.IndexFunc(u.Hostname(), func(r rune) bool { return r > 127 }) >= 0 {
		host, err := idna.Lookup.ToASCII(u.Hostname())
		if err != nil {
			return nil, errors.New("Enter a valid provider hostname.")
		}
		if port := u.Port(); port != "" {
			u.Host = net.JoinHostPort(host, port)
		} else {
			u.Host = host
		}
	}
	path := strings.TrimRight(u.Path, "/")
	if !strings.HasSuffix(path, "/chat/completions") {
		if path == "" {
			path = "/v1"
		}
		path += "/chat/completions"
	}
	u.Path = path
	return u, nil
}

func Explain(ctx context.Context, cfg Config, key string, passage Context, update func(Explanation)) (Explanation, error) {
	endpoint, err := Endpoint(cfg.Endpoint)
	if err != nil {
		return Explanation{}, err
	}
	if strings.TrimSpace(cfg.Model) == "" || strings.TrimSpace(key) == "" {
		return Explanation{}, errors.New("Add your model and API key in Settings.")
	}
	if cfg.Temperature < 0 || cfg.Temperature > 2 {
		return Explanation{}, errors.New("Temperature must be between 0 and 2.")
	}
	prompt := cfg.Prompt
	if strings.TrimSpace(prompt) == "" {
		prompt = DefaultPrompt(cfg.Language)
	}
	schema := Schema()
	format := map[string]any{"type": "json_object"}
	if endpoint.Hostname() == "api.openai.com" {
		format = map[string]any{"type": "json_schema", "json_schema": map[string]any{"name": "explanation", "strict": true, "schema": schema}}
	} else {
		encoded, _ := json.Marshal(schema)
		prompt += " Required JSON schema: " + string(encoded)
	}
	data, _ := json.Marshal(passage.Normalize())
	parameters := map[string]any{
		"model": strings.TrimSpace(cfg.Model), "temperature": cfg.Temperature, "stream": true, "response_format": format,
		"messages": []map[string]string{{"role": "system", "content": prompt}, {"role": "user", "content": string(data)}},
	}
	// DeepSeek's explicit API switch disables reasoning generation, including
	// when its model is served through a compatible provider.
	if endpoint.Hostname() == "api.deepseek.com" || strings.HasPrefix(strings.ToLower(strings.TrimSpace(cfg.Model)), "deepseek-") {
		parameters["thinking"] = map[string]string{"type": "disabled"}
	}
	body, _ := json.Marshal(parameters)
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint.String(), bytes.NewReader(body))
	if err != nil {
		return Explanation{}, errors.New("Could not create the provider request.")
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "text/event-stream")
	request.Header.Set("Authorization", "Bearer "+strings.TrimSpace(key))
	client := &http.Client{Timeout: 75 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	response, err := client.Do(request)
	if err != nil {
		if ctx.Err() != nil {
			return Explanation{}, ctx.Err()
		}
		return Explanation{}, errors.New("Could not reach the provider. Check your connection and provider URL.")
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return Explanation{}, fmt.Errorf("Provider returned HTTP %d. Check your API key, model and provider settings.", response.StatusCode)
	}
	if !strings.Contains(response.Header.Get("Content-Type"), "text/event-stream") {
		return Explanation{}, errors.New("This provider did not return a streaming response. Check model compatibility.")
	}
	return readStream(response.Body, cfg.Language, update)
}

func readStream(reader io.Reader, language string, update func(Explanation)) (Explanation, error) {
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 4096), 256*1024)
	var event []string
	var output []byte
	completed := false
	consume := func() (bool, error) {
		if len(event) == 0 {
			return false, nil
		}
		payload := strings.Join(event, "\n")
		event = nil
		if payload == "[DONE]" {
			completed = true
			return true, nil
		}
		var envelope struct {
			Choices []struct {
				Delta struct {
					Content string `json:"content"`
					Refusal string `json:"refusal"`
				} `json:"delta"`
				FinishReason *string `json:"finish_reason"`
			} `json:"choices"`
			Error json.RawMessage `json:"error"`
		}
		if json.Unmarshal([]byte(payload), &envelope) != nil || (len(envelope.Error) > 0 && !bytes.Equal(bytes.TrimSpace(envelope.Error), []byte("null"))) {
			return false, errors.New("The provider returned an invalid stream. Try again.")
		}
		for _, choice := range envelope.Choices {
			if choice.Delta.Refusal != "" {
				return false, errors.New("The model declined this request.")
			}
			if choice.FinishReason != nil {
				if *choice.FinishReason != "stop" {
					return false, errors.New("The provider stopped before completing the explanation. Try again.")
				}
				completed = true
			}
			output = append(output, choice.Delta.Content...)
			if len(output) > 32*1024 {
				return false, errors.New("The explanation is too long. Shorten your prompt.")
			}
			if choice.Delta.Content != "" {
				update(Partial(output))
			}
		}
		return false, nil
	}
	for scanner.Scan() {
		line := strings.TrimSuffix(scanner.Text(), "\r")
		if line == "" {
			done, err := consume()
			if err != nil {
				return Partial(output), err
			}
			if done {
				return Parse(output, language)
			}
		} else if strings.HasPrefix(line, "data:") {
			event = append(event, strings.TrimPrefix(strings.TrimPrefix(line, "data:"), " "))
		}
	}
	if scanner.Err() != nil {
		return Partial(output), errors.New("The connection was interrupted. Try again.")
	}
	if _, err := consume(); err != nil {
		return Partial(output), err
	}
	if !completed {
		return Partial(output), errors.New("The connection ended before the explanation was complete. Try again.")
	}
	return Parse(output, language)
}
