// Copyright (c) 2026, Anton Bespalov
package csirender

import (
	"fmt"
	"image/color"
	"strings"
)

// parseHexColor parses a hex color string (e.g. "#FF0000" or "red") into color.RGBA.
func parseHexColor(s string) (color.RGBA, error) {
	s = strings.TrimPrefix(s, "#")
	if len(s) == 6 {
		var r, g, b uint8
		_, err := fmt.Sscanf(s, "%02x%02x%02x", &r, &g, &b)
		return color.RGBA{R: r, G: g, B: b, A: 255}, err
	}
	return color.RGBA{}, fmt.Errorf("invalid hex color %s", s)
}

// colorDistance calculates Euclidean distance between two colors in RGB space.
func colorDistance(c1, c2 color.RGBA) float64 {
	dr := float64(c1.R) - float64(c2.R)
	dg := float64(c1.G) - float64(c2.G)
	db := float64(c1.B) - float64(c2.B)
	return dr*dr + dg*dg + db*db
}

// parseColor resolves a color reference from the global palette or uses it directly.
func (rc *RenderContext) parseColor(c string) string {
	if c == "" {
		return ""
	}
	if rc.cfg.Screen.Palette != nil {
		if hex, ok := rc.cfg.Screen.Palette[c]; ok {
			return hex
		}
	}
	return c
}
