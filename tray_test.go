package main

import (
	"bytes"
	"image/png"
	"testing"
)

func TestTrayUsesTransparentTemplateIcon(t *testing.T) {
	for _, language := range []string{"zh", "en"} {
		t.Run(language, func(t *testing.T) {
			options := (&Desktop{}).trayOptions(language)
			if options.Title != "" {
				t.Error("The menu bar must show the app symbol without the old text title")
			}
			if !options.IconIsTemplate {
				t.Error("The menu bar icon must follow the macOS menu bar appearance")
			}
			icon, err := png.Decode(bytes.NewReader(options.Icon))
			if err != nil {
				t.Fatal("The tray must receive a valid embedded PNG:", err)
			}
			if icon.Bounds().Dx() != 36 || icon.Bounds().Dy() != 36 {
				t.Fatal("The menu bar asset must have a 36px Retina canvas")
			}
			var visible, transparent int
			for y := 0; y < 36; y++ {
				for x := 0; x < 36; x++ {
					r, g, b, alpha := icon.At(x, y).RGBA()
					if alpha == 0 {
						transparent++
					} else {
						visible++
						if r != 0 || g != 0 || b != 0 {
							t.Fatal("A template icon must contain only black and transparent pixels")
						}
					}
				}
			}
			if visible < 100 || transparent < 100 {
				t.Fatal("The template must contain a visible symbol on a transparent canvas")
			}
			if options.Menu == nil || len(options.Menu.Items()) != 4 {
				t.Fatal("The icon must keep the existing menu")
			}
		})
	}
}
