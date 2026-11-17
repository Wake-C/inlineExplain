package main

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"strings"

	"inlineexplain.local/desktop/internal/speech"
)

// Word audio is American English by default. The card shows both IPA forms;
// the speaker button reads the US one.
const wordAudioAccent = "US"

// beginSpeech makes a new request current, stops the previous voice and
// invalidates any word lookup that is still in flight.
func (d *Desktop) beginSpeech() uint64 {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.speechID++
	if d.speech != nil && d.speech.Process != nil {
		_ = d.speech.Process.Kill()
	}
	d.speech = nil
	d.dropSpeechFileLocked()
	return d.speechID
}

// dropSpeechFileLocked removes the temporary recording of the current voice.
// The caller holds d.mu. Removing an open file is safe; playback keeps its
// descriptor until it stops.
func (d *Desktop) dropSpeechFileLocked() {
	if d.speechFile != "" {
		_ = os.Remove(d.speechFile)
		d.speechFile = ""
	}
}

// startSpeech plays command only if no newer request replaced it meanwhile.
// A temporary audio file is removed after playback or when it is superseded.
func (d *Desktop) startSpeech(id uint64, command *exec.Cmd, file string) (bool, error) {
	remove := func() {
		if file != "" {
			_ = os.Remove(file)
		}
	}
	d.mu.Lock()
	if id != d.speechID {
		d.mu.Unlock()
		remove()
		return false, nil
	}
	if d.speech != nil && d.speech.Process != nil {
		_ = d.speech.Process.Kill()
	}
	d.dropSpeechFileLocked()
	if err := command.Start(); err != nil {
		d.speech = nil
		d.mu.Unlock()
		remove()
		return false, errors.New("Speech could not start. Check your macOS voices.")
	}
	d.speech = command
	d.speechFile = file
	d.mu.Unlock()
	go func() {
		_ = command.Wait()
		if file != "" {
			_ = os.Remove(file)
			d.mu.Lock()
			if d.speechFile == file {
				d.speechFile = ""
			}
			d.mu.Unlock()
		}
	}()
	return true, nil
}

// Speak reads an example sentence with the system voice.
func (d *Desktop) Speak(text string) error {
	if strings.TrimSpace(text) == "" || len(text) > 2000 {
		return nil
	}
	command := exec.Command("/usr/bin/say")
	command.Stdin = strings.NewReader(text)
	_, err := d.startSpeech(d.beginSpeech(), command, "")
	return err
}

// SpeakWord plays the dictionary recording for one word. When the free
// dictionary service has no usable audio, the system voice says the word.
func (d *Desktop) SpeakWord(term string) error {
	word := speech.WordTerm(term)
	if word == "" {
		return nil
	}
	id := d.beginSpeech()
	audio, err := (speech.Youdao{UserAgent: "InlineExplain/" + appVersion + " (macOS)"}).Lookup(context.Background(), word, wordAudioAccent)
	if err == nil {
		if path, writeErr := writeWordAudio(audio); writeErr == nil {
			started, startErr := d.startSpeech(id, exec.Command("/usr/bin/afplay", path), path)
			if started || startErr == nil {
				return startErr
			}
		}
	}
	command := exec.Command("/usr/bin/say")
	command.Stdin = strings.NewReader(word)
	_, err = d.startSpeech(id, command, "")
	return err
}

func writeWordAudio(audio []byte) (string, error) {
	file, err := os.CreateTemp("", "inlineexplain-word-*.mp3")
	if err != nil {
		return "", err
	}
	path := file.Name()
	if _, err := file.Write(audio); err != nil {
		_ = file.Close()
		_ = os.Remove(path)
		return "", err
	}
	if err := file.Close(); err != nil {
		_ = os.Remove(path)
		return "", err
	}
	return path, nil
}
