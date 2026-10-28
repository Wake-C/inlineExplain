package main

import (
	"context"
	"testing"

	"github.com/egoist/mygo"
)

func TestWindowCloseHidesUntilApplicationQuit(t *testing.T) {
	desktop := &Desktop{}
	hidden := map[string]int{}
	callbacks := map[string]func(*mygo.CloseEvent){}
	for _, window := range []string{"card", "preferences"} {
		callbacks[window] = desktop.hideOnClose(func() { hidden[window]++ })
		event := &mygo.CloseEvent{}
		callbacks[window](event)
		if !event.DefaultPrevented() || hidden[window] != 1 {
			t.Fatalf("Normal %s close must hide and keep the window reusable", window)
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	desktop.cancel = func() {
		if !desktop.quitting {
			t.Error("The quit state must be set before stopping the active lookup")
		}
		cancel()
	}
	desktop.beforeQuit(&mygo.QuitEvent{})
	if ctx.Err() == nil || desktop.cancel != nil {
		t.Fatal("Application quit must cancel the active lookup")
	}
	for _, window := range []string{"card", "preferences"} {
		event := &mygo.CloseEvent{}
		callbacks[window](event)
		if event.DefaultPrevented() {
			t.Errorf("%s close must allow application quit", window)
		}
		if hidden[window] != 1 {
			t.Errorf("Application quit must close %s instead of hiding it again", window)
		}
	}
}
