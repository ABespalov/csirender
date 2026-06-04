// Copyright (c) 2026, Anton Bespalov
package csirender

import (
	"github.com/fogleman/gg"
)

func renderLine(dc *gg.Context, rc *RenderContext, el *LineElement) {
	c := el.Color
	if c == "" {
		c = "black" // Default color
	}
	dc.SetHexColor(rc.parseColor(c))

	t := 1.0
	if el.T != nil {
		t = *el.T
	} else if el.Size.H > 0 && el.Size.W > 0 {
		if el.Size.W > el.Size.H {
			t = el.Size.H
		} else {
			t = el.Size.W
		}
	}

	l := 1.0
	if el.L != nil {
		l = *el.L
	} else if el.Size.H > 0 && el.Size.W > 0 {
		if el.Size.W > el.Size.H {
			l = el.Size.W
		} else {
			l = el.Size.H
		}
	}

	angle := 0.0
	if el.Angle != nil {
		angle = *el.Angle
	} else if el.Size.H > el.Size.W && el.L == nil && el.T == nil {
		angle = 90.0
	}

	dc.Push()
	defer dc.Pop()

	if angle != 0 {
		dc.Translate(el.Position.X, el.Position.Y)
		dc.Rotate(gg.Radians(angle))
		dc.MoveTo(0, t/2.0)
		dc.LineTo(l, t/2.0)
	} else {
		dc.MoveTo(el.Position.X, el.Position.Y+t/2.0)
		dc.LineTo(el.Position.X+l, el.Position.Y+t/2.0)
	}

	dc.SetLineWidth(t)

	dashes := parseLineStyle(el.Style, t)
	if len(dashes) > 0 {
		dc.SetDash(dashes...)
	}

	dc.Stroke()
}
