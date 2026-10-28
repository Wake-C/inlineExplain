package main

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/egoist/mygo"
	"inlineexplain.local/desktop/internal/explain"
)

type Settings struct {
	Endpoint      string  `json:"endpoint"`
	Model         string  `json:"model"`
	Temperature   float64 `json:"temperature"`
	Language      string  `json:"language"`
	EnglishPrompt string  `json:"englishPrompt"`
	ChinesePrompt string  `json:"chinesePrompt"`
	Shortcut      string  `json:"shortcut"`
}

type Preferences struct {
	Settings        Settings `json:"settings"`
	Revision        uint64   `json:"revision"`
	Version         string   `json:"version"`
	HasKey          bool     `json:"hasKey"`
	ScreenRecording bool     `json:"screenRecording"`
}

// SettingsPatch changes only fields edited in the form. An older open window
// cannot overwrite a language change made on the reading card.
type SettingsPatch struct {
	Endpoint      *string  `json:"endpoint,omitempty"`
	Model         *string  `json:"model,omitempty"`
	Temperature   *float64 `json:"temperature,omitempty"`
	Language      *string  `json:"language,omitempty"`
	EnglishPrompt *string  `json:"englishPrompt,omitempty"`
	ChinesePrompt *string  `json:"chinesePrompt,omitempty"`
	Shortcut      *string  `json:"shortcut,omitempty"`
}

func (p SettingsPatch) apply(settings Settings) Settings {
	if p.Endpoint != nil {
		settings.Endpoint = *p.Endpoint
	}
	if p.Model != nil {
		settings.Model = *p.Model
	}
	if p.Temperature != nil {
		settings.Temperature = *p.Temperature
	}
	if p.Language != nil {
		settings.Language = *p.Language
	}
	if p.EnglishPrompt != nil {
		settings.EnglishPrompt = *p.EnglishPrompt
	}
	if p.ChinesePrompt != nil {
		settings.ChinesePrompt = *p.ChinesePrompt
	}
	if p.Shortcut != nil {
		settings.Shortcut = *p.Shortcut
	}
	return settings
}

func loadSettings() (Settings, error) {
	settings := defaultSettings()
	data, err := os.ReadFile(settingsPath())
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return settings, errors.New("Could not read your settings.")
	}
	if err == nil && json.Unmarshal(data, &settings) != nil {
		return settings, errors.New("Your settings file could not be read. Open Settings to save it again.")
	}
	if settings.Language != "zh" && settings.Language != "en" {
		settings.Language = "zh"
	}
	if settings.Shortcut == "" {
		settings.Shortcut = "Alt+S"
	}
	return settings, nil
}

func defaultSettings() Settings {
	return Settings{Endpoint: "https://api.deepseek.com", Model: "deepseek-chat", Temperature: .3, Language: "zh", Shortcut: "Alt+S"}
}

func settingsPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "Library", "Application Support", "InlineExplain", "settings.json")
}

func saveSettings(settings Settings) error {
	path := settingsPath()
	if os.MkdirAll(filepath.Dir(path), 0700) != nil {
		return errors.New("Could not create the settings folder.")
	}
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return errors.New("Could not encode settings.")
	}
	if os.WriteFile(path+".tmp", data, 0600) != nil {
		return errors.New("Could not save settings.")
	}
	return os.Rename(path+".tmp", path)
}

func keyIdentity(endpoint string) (string, string) {
	u, err := explain.Endpoint(endpoint)
	if err != nil {
		return "com.inlineexplain.desktop.provider", "default"
	}
	account := "user"
	if current, err := user.Current(); err == nil {
		account = current.Username
	}
	port := u.Port()
	if port == "" {
		if u.Scheme == "https" {
			port = "443"
		} else {
			port = "80"
		}
	}
	origin := strings.ToLower(u.Scheme) + "://" + net.JoinHostPort(strings.ToLower(u.Hostname()), port)
	digest := sha256.Sum256([]byte(origin))
	return fmt.Sprintf("com.inlineexplain.desktop.provider.%x", digest), account
}

func readKey(endpoint string) (string, error) {
	service, account := keyIdentity(endpoint)
	key, err := mygo.Secrets.Service(service).Get(account)
	if err != nil {
		return "", errors.New("Add your provider API key in Settings.")
	}
	return string(key), nil
}

func writeKey(endpoint, key string) error {
	service, account := keyIdentity(endpoint)
	if err := mygo.Secrets.Service(service).Set(account, []byte(strings.TrimSpace(key))); err != nil {
		return errors.New("Could not save your API key to Keychain.")
	}
	return nil
}
