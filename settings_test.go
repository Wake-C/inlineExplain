package main

import "testing"

func TestChineseDefault(t *testing.T) {
	settings := defaultSettings()
	if settings.Language != "zh" || settings.Endpoint != "https://api.deepseek.com" {
		t.Fatal("The first-run default must use Chinese and the official DeepSeek endpoint")
	}
	if localeTag(settings.Language) != "zh-Hans" || localeTag("en") != "en" {
		t.Fatal("Native menu locale does not follow the app language")
	}
	desktop := &Desktop{}
	if desktop.trayMenu("zh").Items()[1].Label != "设置…" || desktop.trayMenu("en").Items()[1].Label != "Settings…" {
		t.Fatal("Native menu labels must follow the chosen language")
	}
}

func TestSettingsPatchPreservesUneditedFields(t *testing.T) {
	language := "zh"
	current := Settings{Endpoint: "https://api.deepseek.com", Model: "deepseek-chat", Temperature: .3, Language: language, EnglishPrompt: "custom English", ChinesePrompt: "custom Chinese", Shortcut: "Alt+S"}
	temperature := 0.4
	updated := (SettingsPatch{Temperature: &temperature}).apply(current)
	if updated.Language != language || updated.EnglishPrompt != current.EnglishPrompt || updated.ChinesePrompt != current.ChinesePrompt || updated.Temperature != .4 {
		t.Fatal("A partial settings update replaced another preference")
	}
}

func TestProviderCredentialIdentity(t *testing.T) {
	for _, pair := range [][2]string{
		{"https://api.deepseek.com", "https://api-deepseek.com"},
		{"http://localhost:8080", "https://localhost:8080"},
		{"https://example.com:443", "https://example.com:8443"},
		{"https://i.example", "https://İ.example"},
	} {
		first, firstAccount := keyIdentity(pair[0])
		second, secondAccount := keyIdentity(pair[1])
		if first == second && firstAccount == secondAccount {
			t.Fatalf("Different provider origins share credentials: %s and %s", pair[0], pair[1])
		}
	}
	first, account := keyIdentity("https://API.DEEPSEEK.COM:443/v1")
	second, otherAccount := keyIdentity("https://api.deepseek.com")
	if first != second || account != otherAccount {
		t.Fatal("Equivalent origins must use the same credential")
	}
	first, account = keyIdentity("https://İ.example")
	second, otherAccount = keyIdentity("https://xn--i-9bb.example")
	if first != second || account != otherAccount {
		t.Fatal("Unicode hostnames must use the same IDNA origin as the HTTP client")
	}
}
