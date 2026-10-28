package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"inlineexplain.local/desktop/internal/explain"
)

type Bounds struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

type Capture struct {
	Context explain.Context `json:"context"`
	Bounds  Bounds          `json:"bounds"`
}

func helperPath() string {
	executable, _ := os.Executable()
	return filepath.Join(filepath.Dir(executable), "InlineExplainOCR")
}

func captureAt(ctx context.Context, x, y int) (Capture, error) {
	command := exec.CommandContext(ctx, helperPath(), "capture", fmt.Sprint(x), fmt.Sprint(y), fmt.Sprint(os.Getpid()))
	data, err := command.Output()
	if ctx.Err() != nil {
		return Capture{}, ctx.Err()
	}
	var failure struct {
		Error string `json:"error"`
	}
	_ = json.Unmarshal(data, &failure)
	if failure.Error != "" {
		return Capture{}, errors.New(failure.Error)
	}
	if err != nil {
		return Capture{}, errors.New("Could not capture the screen. Check Screen Recording access in Settings.")
	}
	var result Capture
	if json.Unmarshal(data, &result) != nil || result.Context.Selection == "" {
		return result, errors.New("No word found. Point directly at text and try again.")
	}
	result.Context = result.Context.Normalize()
	return result, nil
}

func screenPermission(request bool) bool {
	args := []string{"permission"}
	if request {
		args = append(args, "--request")
	}
	data, err := exec.Command(helperPath(), args...).Output()
	var result struct {
		ScreenRecording bool `json:"screenRecording"`
	}
	return err == nil && json.Unmarshal(data, &result) == nil && result.ScreenRecording
}
