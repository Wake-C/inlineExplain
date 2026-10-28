package explain

import (
	"bytes"
	"embed"
	"encoding/json"
	"errors"
	"io"
	"strings"
)

//go:embed prompts/*.txt
var prompts embed.FS

func DefaultPrompt(language string) string {
	if language != "zh" {
		language = "en"
	}
	data, _ := prompts.ReadFile("prompts/" + language + ".txt")
	return strings.TrimSpace(string(data))
}

type Example struct {
	Text        string `json:"text"`
	Translation string `json:"translation"`
}

type Explanation struct {
	Term         string    `json:"term"`
	PartOfSpeech string    `json:"partOfSpeech"`
	UKIPA        string    `json:"ukIpa"`
	USIPA        string    `json:"usIpa"`
	Meaning      string    `json:"meaning"`
	Context      string    `json:"context"`
	Examples     []Example `json:"examples"`
	Synonyms     string    `json:"synonyms"`
	Antonyms     string    `json:"antonyms"`
}

// Partial publishes only values whose JSON is complete, including escaped text.
// The standard decoder does the parsing; incomplete values remain hidden.
func Partial(data []byte) Explanation {
	decoder := json.NewDecoder(bytes.NewReader(data))
	if token, err := decoder.Token(); err != nil || token != json.Delim('{') {
		return Explanation{}
	}
	fields := make(map[string]json.RawMessage)
	for decoder.More() {
		token, err := decoder.Token()
		if err != nil {
			break
		}
		key, ok := token.(string)
		if !ok {
			break
		}
		var value json.RawMessage
		if decoder.Decode(&value) != nil {
			break
		}
		fields[key] = value
	}
	encoded, _ := json.Marshal(fields)
	var answer Explanation
	_ = json.Unmarshal(encoded, &answer)
	return answer
}

func Parse(data []byte, language string) (Explanation, error) {
	var answer Explanation
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&answer) != nil {
		return answer, errors.New("The model returned an incomplete explanation. Try again.")
	}
	if decoder.Decode(new(any)) != io.EOF {
		return answer, errors.New("The model returned extra text. Check your prompt and try again.")
	}
	var fields map[string]json.RawMessage
	_ = json.Unmarshal(data, &fields)
	if len(fields) != 9 {
		return answer, errors.New("The explanation is missing required fields. Check your prompt.")
	}
	for _, value := range fields {
		if bytes.Equal(value, []byte("null")) {
			return answer, errors.New("The model returned an empty explanation field. Try again.")
		}
	}
	if strings.TrimSpace(answer.Term) == "" || strings.TrimSpace(answer.PartOfSpeech) == "" || strings.TrimSpace(answer.Meaning) == "" || strings.TrimSpace(answer.Context) == "" || len(answer.Examples) < 1 || len(answer.Examples) > 2 {
		return answer, errors.New("The explanation is missing its meaning, context or examples. Try again.")
	}
	for _, e := range answer.Examples {
		if strings.TrimSpace(e.Text) == "" || (language == "zh" && strings.TrimSpace(e.Translation) == "") || (language != "zh" && e.Translation != "") {
			return answer, errors.New("The example language does not match your settings. Check your prompt.")
		}
	}
	for _, relation := range []string{answer.Synonyms, answer.Antonyms} {
		if len([]rune(relation)) > 160 || strings.ContainsAny(relation, "\r\n") {
			return answer, errors.New("Word relations must be short, single-line text. Check your prompt.")
		}
	}
	return answer, nil
}

func Schema() map[string]any {
	properties := map[string]any{}
	keys := []string{"term", "partOfSpeech", "ukIpa", "usIpa", "meaning", "context", "examples", "synonyms", "antonyms"}
	for _, key := range keys {
		properties[key] = map[string]any{"type": "string"}
	}
	properties["examples"] = map[string]any{"type": "array", "minItems": 1, "maxItems": 2, "items": map[string]any{
		"type": "object", "additionalProperties": false,
		"properties": map[string]any{"text": map[string]string{"type": "string"}, "translation": map[string]string{"type": "string"}},
		"required":   []string{"text", "translation"},
	}}
	return map[string]any{"type": "object", "additionalProperties": false, "properties": properties, "required": keys}
}
