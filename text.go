// Copyright (c) 2026, Anton Bespalov. Licensed under the MIT License.
package csirender

import (
	"math"
	"strings"

	"github.com/fogleman/gg"
)

func renderText(dc *gg.Context, rc *RenderContext, el *TextElement) {
	text := rc.resolvePlaceholders(el.Template)
	rc.loadFont(el.Font.Face)
	dc.SetHexColor(rc.parseColor(el.Font.Fg))

	var tx float64 = el.Position.X + el.Size.W/2
	var tax float64 = 0.5
	if el.Align.H == "left" {
		tx = el.Position.X
		tax = 0.0
	}

	lines := strings.Split(text, "\n")
	for i, line := range lines {
		// Anchored text drawing helper
		w, h := dc.MeasureString(line)
		lx := tx - tax*w
		ly := el.Position.Y + float64(i)*25 + 15 + (1.0-0.5)*h
		dc.DrawString(line, math.Round(lx), math.Round(ly))
	}

}
