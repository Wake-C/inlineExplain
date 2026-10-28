package main

import (
	"context"
	"embed"
	"errors"
	"io/fs"
	"log"
	"math"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/egoist/mygo"
	"inlineexplain.local/desktop/internal/explain"
)

//go:embed frontend/*
var frontend embed.FS

//go:embed assets/tray.png
var trayIcon []byte

// Release builds receive the version from packaging/Info.plist.
var appVersion = "dev"

var lookupChanged = mygo.NewEvent[LookupState]("lookup:changed")
var preferencesRefresh = mygo.NewEvent[struct{}]("preferences:refresh")

type LookupState struct {
	ID       uint64              `json:"id"`
	Passage  explain.Context     `json:"passage"`
	Answer   explain.Explanation `json:"answer"`
	Language string              `json:"language"`
	Busy     bool                `json:"busy"`
	Error    string              `json:"error"`
	Seconds  float64             `json:"seconds"`
	Pinned   bool                `json:"pinned"`
}

type Desktop struct {
	mu                sync.Mutex
	settingsWriteMu   sync.Mutex
	settingsRevision  uint64
	settings          Settings
	state             LookupState
	cancel            context.CancelFunc
	card, preferences *mygo.Window
	tray              *mygo.Tray
	anchor            mygo.Point
	quitting          bool // Native close and quit callbacks run on the main thread.
	speech            *exec.Cmd
}

func main() {
	settings, loadError := loadSettings()
	desktop := &Desktop{settings: settings}
	mygo.App.SetName("InlineExplain")
	mygo.App.SetVersion(appVersion)
	mygo.App.SetActivationPolicy(mygo.ActivationPolicyAccessory)
	mygo.Theme.SetSource(mygo.ThemeLight)
	mygo.App.SetLocale(localeTag(settings.Language))
	mygo.App.OnWindowAllClosed(func() {})
	mygo.App.OnActivate(func(visible bool) {
		if !visible {
			desktop.OpenSettings()
		}
	})
	mygo.App.OnSecondInstance(func([]string, string) { desktop.OpenSettings() })
	if !mygo.App.RequestSingleInstanceLock() {
		return
	}
	web, _ := fs.Sub(frontend, "frontend")
	mygo.SetFrontend(web)
	mygo.Bind(desktop)
	mygo.App.WhenReady(func() {
		desktop.card = mygo.NewWindow(mygo.WindowOptions{Title: "InlineExplain", URL: "/", Panel: true, Hidden: true,
			Frameless: true, Transparent: false, BackgroundColor: "#ffffff", CornerRadius: 18,
			Width: 410, Height: 280, MinWidth: 360, MaxWidth: 480, MinHeight: 150, MaxHeight: 680, DisableResize: true})
		desktop.card.OnClose(desktop.hideOnClose(desktop.Hide))
		desktop.card.OnBlur(func() {
			desktop.mu.Lock()
			pinned := desktop.state.Pinned
			desktop.mu.Unlock()
			if !pinned {
				desktop.Hide()
			}
		})
		desktop.preferences = mygo.NewWindow(mygo.WindowOptions{Title: nativeText(settings.Language, "InlineExplain Settings", "InlineExplain 设置"), URL: "/?view=settings", Panel: true,
			Hidden: true, Width: 600, Height: 720, MinWidth: 500, MinHeight: 540, BackgroundColor: "#ffffff"})
		desktop.preferences.OnClose(desktop.hideOnClose(desktop.preferences.Hide))
		tray, err := mygo.NewTray(desktop.trayOptions(settings.Language))
		desktop.tray = tray
		if err != nil {
			log.Print("Could not create the menu bar item.")
		}
		desktop.applyLocale(settings.Language)
		if err := desktop.registerShortcut(settings.Shortcut); err != nil {
			loadError = err
		}
		if loadError != nil {
			desktop.mu.Lock()
			desktop.state.Error = loadError.Error()
			desktop.mu.Unlock()
		}
		go func() {
			if _, err := readKey(settings.Endpoint); err != nil || loadError != nil || !screenPermission(false) {
				desktop.OpenSettings()
			}
		}()
	})
	mygo.App.OnBeforeQuit(desktop.beforeQuit)
	if err := mygo.App.Run(); err != nil {
		log.Fatal("InlineExplain could not start: ", err)
	}
}

func (d *Desktop) registerShortcut(shortcut string) error {
	return mygo.GlobalShortcut.Register(shortcut, d.startLookup)
}

func (d *Desktop) startLookup() {
	point := mygo.Screen.CursorScreenPoint()
	d.card.Hide()
	d.mu.Lock()
	if d.cancel != nil {
		d.cancel()
	}
	ctx, cancel := context.WithCancel(context.Background())
	d.cancel = cancel
	d.state = LookupState{ID: d.state.ID + 1, Busy: true, Language: d.settings.Language}
	id := d.state.ID
	d.anchor = point
	d.mu.Unlock()
	go func() {
		started := time.Now()
		capture, err := captureAt(ctx, point.X, point.Y)
		if ctx.Err() != nil {
			return
		}
		d.mu.Lock()
		if id != d.state.ID {
			d.mu.Unlock()
			return
		}
		if err == nil {
			d.state.Passage = capture.Context
			d.anchor = mygo.Point{X: int(capture.Bounds.X), Y: int(capture.Bounds.Y + capture.Bounds.Height)}
		}
		d.mu.Unlock()
		d.placeCard(280)
		d.card.Show() // Nonactivating NSPanel: the source app and full-screen Space stay active.
		d.publish()
		if err != nil {
			d.finish(id, explain.Explanation{}, err, started)
			return
		}
		d.runExplanation(ctx, id, started)
	}()
}

func (d *Desktop) runExplanation(ctx context.Context, id uint64, started time.Time) {
	d.mu.Lock()
	settings, passage := d.settings, d.state.Passage
	d.mu.Unlock()
	prompt := settings.EnglishPrompt
	if settings.Language == "zh" {
		prompt = settings.ChinesePrompt
	}
	key, err := readKey(settings.Endpoint)
	if err != nil {
		d.finish(id, explain.Explanation{}, err, started)
		return
	}
	answer, err := explain.Explain(ctx, explain.Config{Endpoint: settings.Endpoint, Model: settings.Model, Temperature: settings.Temperature, Language: settings.Language, Prompt: prompt}, key, passage, func(answer explain.Explanation) {
		d.mu.Lock()
		if d.state.ID != id || ctx.Err() != nil {
			d.mu.Unlock()
			return
		}
		d.state.Answer = answer
		d.state.Seconds = time.Since(started).Seconds()
		d.mu.Unlock()
		d.publish()
	})
	if ctx.Err() == nil {
		d.finish(id, answer, err, started)
	}
}

func (d *Desktop) finish(id uint64, answer explain.Explanation, err error, started time.Time) {
	d.mu.Lock()
	if d.state.ID != id {
		d.mu.Unlock()
		return
	}
	d.state.Answer = answer
	d.state.Busy = false
	d.state.Seconds = time.Since(started).Seconds()
	if err != nil {
		d.state.Error = err.Error()
	}
	d.mu.Unlock()
	d.publish()
}

func (d *Desktop) publish() { _ = lookupChanged.Emit(d.card, d.Read()) }

func (d *Desktop) Read() LookupState {
	d.mu.Lock()
	defer d.mu.Unlock()
	state := d.state
	if state.Language == "" {
		state.Language = d.settings.Language
	}
	return state
}

func (d *Desktop) Resize(height int) {
	if height < 150 {
		height = 150
	}
	if height > 680 {
		height = 680
	}
	d.placeCard(height)
}

func (d *Desktop) placeCard(height int) {
	d.mu.Lock()
	anchor := d.anchor
	d.mu.Unlock()
	area := mygo.Screen.DisplayNearestPoint(anchor).WorkArea
	height = min(height, area.Height-24)
	x, y := anchor.X+22, anchor.Y+14
	if x+410 > area.X+area.Width-12 {
		x = anchor.X - 410 - 22
	}
	x = max(area.X+12, min(x, area.X+area.Width-422))
	if y+height > area.Y+area.Height-12 {
		y = anchor.Y - height - 18
	}
	y = max(area.Y+12, min(y, area.Y+area.Height-height-12))
	d.card.SetBounds(mygo.Rectangle{X: x, Y: y, Width: 410, Height: height})
}

func (d *Desktop) Hide() { d.stop(); d.card.Hide() }

func (d *Desktop) hideOnClose(hide func()) func(*mygo.CloseEvent) {
	return func(event *mygo.CloseEvent) {
		if d.quitting {
			return
		}
		event.PreventDefault()
		hide()
	}
}

func (d *Desktop) beforeQuit(*mygo.QuitEvent) {
	d.quitting = true
	d.stop()
}

func (d *Desktop) stop() {
	d.mu.Lock()
	if d.cancel != nil {
		d.cancel()
		d.cancel = nil
	}
	if d.speech != nil && d.speech.Process != nil {
		_ = d.speech.Process.Kill()
		d.speech = nil
	}
	d.mu.Unlock()
}

func (d *Desktop) Retry() {
	d.mu.Lock()
	if d.state.Passage.Selection == "" {
		d.mu.Unlock()
		return
	}
	if d.cancel != nil {
		d.cancel()
	}
	ctx, cancel := context.WithCancel(context.Background())
	d.cancel = cancel
	d.state.ID++
	id := d.state.ID
	d.state.Busy = true
	d.state.Error = ""
	d.state.Answer = explain.Explanation{}
	d.state.Language = d.settings.Language
	d.mu.Unlock()
	d.publish()
	go d.runExplanation(ctx, id, time.Now())
}

func (d *Desktop) ChangeLanguage(language string) error {
	if language != "en" && language != "zh" {
		return errors.New("Choose English or Chinese.")
	}
	d.settingsWriteMu.Lock()
	defer d.settingsWriteMu.Unlock()
	d.mu.Lock()
	settings := d.settings
	settings.Language = language
	if err := saveSettings(settings); err != nil {
		d.mu.Unlock()
		return err
	}
	d.settings = settings
	d.settingsRevision++
	d.mu.Unlock()
	d.applyLocale(language)
	d.Retry()
	d.refreshPreferences()
	return nil
}

func (d *Desktop) TogglePin() {
	d.mu.Lock()
	d.state.Pinned = !d.state.Pinned
	d.mu.Unlock()
	d.publish()
}

func (d *Desktop) Copy() {
	s := d.Read()
	parts := []string{s.Answer.Term, s.Answer.Meaning, s.Answer.Context}
	for _, example := range s.Answer.Examples {
		parts = append(parts, example.Text)
		if example.Translation != "" {
			parts = append(parts, example.Translation)
		}
	}
	mygo.Clipboard.WriteText(strings.Join(parts, "\n\n"))
}

func (d *Desktop) Speak(text string) error {
	if strings.TrimSpace(text) == "" || len(text) > 2000 {
		return nil
	}
	d.mu.Lock()
	if d.speech != nil && d.speech.Process != nil {
		_ = d.speech.Process.Kill()
	}
	command := exec.Command("/usr/bin/say")
	command.Stdin = strings.NewReader(text)
	if err := command.Start(); err != nil {
		d.mu.Unlock()
		return errors.New("Speech could not start. Check your macOS voices.")
	}
	d.speech = command
	d.mu.Unlock()
	go func() { _ = command.Wait() }()
	return nil
}

func (d *Desktop) OpenSettings() {
	if d.preferences != nil {
		d.preferences.Show()
		d.preferences.Focus()
		d.refreshPreferences()
	}
}

func (d *Desktop) refreshPreferences() {
	if d.preferences != nil {
		_ = preferencesRefresh.Emit(d.preferences, struct{}{})
	}
}

func (d *Desktop) Preferences() Preferences {
	d.mu.Lock()
	settings := d.settings
	revision := d.settingsRevision
	d.mu.Unlock()
	if settings.EnglishPrompt == "" {
		settings.EnglishPrompt = explain.DefaultPrompt("en")
	}
	if settings.ChinesePrompt == "" {
		settings.ChinesePrompt = explain.DefaultPrompt("zh")
	}
	key, _ := readKey(settings.Endpoint)
	return Preferences{Settings: settings, Revision: revision, Version: appVersion, HasKey: key != "", ScreenRecording: screenPermission(false)}
}

func (d *Desktop) Save(patch SettingsPatch, apiKey string) (Preferences, error) {
	d.settingsWriteMu.Lock()
	defer d.settingsWriteMu.Unlock()
	d.mu.Lock()
	old := d.settings
	d.mu.Unlock()
	settings := patch.apply(old)
	if _, err := explain.Endpoint(settings.Endpoint); err != nil {
		return Preferences{}, err
	}
	if strings.TrimSpace(settings.Model) == "" {
		return Preferences{}, errors.New("Enter a model name.")
	}
	if math.IsNaN(settings.Temperature) || settings.Temperature < 0 || settings.Temperature > 2 {
		return Preferences{}, errors.New("Temperature must be between 0 and 2.")
	}
	if settings.Language != "en" && settings.Language != "zh" {
		return Preferences{}, errors.New("Choose English or Chinese.")
	}
	if settings.Shortcut != "Alt+S" && settings.Shortcut != "Alt+Shift+S" {
		return Preferences{}, errors.New("Choose a supported shortcut.")
	}
	if settings.Shortcut != old.Shortcut {
		if err := d.registerShortcut(settings.Shortcut); err != nil {
			return Preferences{}, errors.New("That shortcut is not available. Choose another.")
		}
	}
	if strings.TrimSpace(apiKey) != "" {
		if err := writeKey(settings.Endpoint, apiKey); err != nil {
			if settings.Shortcut != old.Shortcut {
				mygo.GlobalShortcut.Unregister(settings.Shortcut)
			}
			return Preferences{}, err
		}
	}
	if err := saveSettings(settings); err != nil {
		if settings.Shortcut != old.Shortcut {
			mygo.GlobalShortcut.Unregister(settings.Shortcut)
		}
		return Preferences{}, err
	}
	if settings.Shortcut != old.Shortcut {
		mygo.GlobalShortcut.Unregister(old.Shortcut)
	}
	d.mu.Lock()
	d.settings = settings
	d.settingsRevision++
	d.mu.Unlock()
	if settings.Language != old.Language {
		d.applyLocale(settings.Language)
		if d.card != nil && d.card.IsVisible() {
			d.Retry()
		} else {
			d.stop()
			d.mu.Lock()
			d.state.ID++
			d.state.Language = settings.Language
			d.state.Answer = explain.Explanation{}
			d.state.Busy = false
			d.state.Error = ""
			d.mu.Unlock()
			if d.card != nil {
				d.publish()
			}
		}
	}
	return d.Preferences(), nil
}

func (d *Desktop) DefaultPrompt(language string) string { return explain.DefaultPrompt(language) }

func (d *Desktop) RequestScreenRecording() bool { return screenPermission(true) }

func (d *Desktop) OpenScreenSettings() {
	_ = mygo.Shell.OpenExternal("x-apple.systempreferences:com.apple.preference.security?Privacy_ScreenCapture")
}
