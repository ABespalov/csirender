// Copyright (c) 2026, Anton Bespalov. Licensed under the MIT License.
package csirender

import (
	"math"

	"github.com/fogleman/gg"
)

func renderShape(dc *gg.Context, rc *RenderContext, el *ShapeElement) {
	// 1. Determine base styling from element and its border configuration
	fillColor := el.Color
	if fillColor == "" && el.Border != nil {
		fillColor = el.Border.Bg
	}

	borderColor := ""
	borderWidth := 0.0
	if el.Border != nil {
		borderColor = el.Border.Color
		if el.Border.T != nil {
			borderWidth = *el.Border.T
		}
	}

	// 2. Prepare telemetry for thresholds
	localTel := &LocalDataProvider{
		Parent:    rc.provider,
		Overrides: make(map[string]interface{}),
	}

	if el.Source != "" {
		if val, ok := rc.provider.GetValue(el.Source); ok {
			localTel.Overrides[el.Source] = val
			localTel.Overrides[cleanBase(el.Source)] = val
		}
	}

	// 3. Evaluate element-level thresholds
	for _, th := range el.Thresholds {
		if evalCondition(th.Condition, localTel) {
			if th.Bg != "" {
				fillColor = th.Bg
			}
			if th.Color != "" {
				fillColor = th.Color
			}
			if th.Fg != "" {
				borderColor = th.Fg
			}
			break
		}
	}

	// Evaluate border-specific thresholds
	if el.Border != nil {
		for _, th := range el.Border.Thresholds {
			if evalCondition(th.Condition, localTel) {
				if th.Bg != "" {
					fillColor = th.Bg
				}
				if th.Color != "" {
					fillColor = th.Color
				}
				if th.Fg != "" {
					borderColor = th.Fg
				}
				break
			}
		}
	}

	// 4. Determine rotation angle
	rotation := el.Rotation
	if el.RotationExpr != "" {
		if dynRot, ok := rc.provider.ResolveThreshold(el.RotationExpr, el.Source); ok {
			rotation += dynRot
		} else {
			// Fallback: check if expression is exactly a telemetry key
			if v, ok := rc.provider.GetValue(el.RotationExpr); ok {
				if vf, ok := v.(float64); ok {
					rotation += vf
				}
			}
		}
	}

	// 5. Setup context and apply rotation
	dc.Push()
	defer dc.Pop()

	cx := el.Position.X + el.Size.W/2.0
	cy := el.Position.Y + el.Size.H/2.0

	if rotation != 0 {
		dc.RotateAbout(rotation*math.Pi/180.0, cx, cy)
	}

	// 6. Construct path based on kind
	kind := el.Kind
	if kind == "" {
		kind = "rect" // Default
	}

	switch kind {
	case "circle":
		r := math.Min(el.Size.W, el.Size.H) / 2.0
		dc.DrawCircle(cx, cy, r)
	case "ellipse":
		rx := el.Size.W / 2.0
		ry := el.Size.H / 2.0
		dc.DrawEllipse(cx, cy, rx, ry)
	case "polygon":
		sides := el.Sides
		if sides < 3 {
			sides = 3 // Minimum for a polygon is a triangle
		}
		r := math.Min(el.Size.W, el.Size.H) / 2.0
		dc.DrawRegularPolygon(sides, cx, cy, r, 0)
	case "rect":
		fallthrough
	default:
		if el.CornerRadius > 0 {
			dc.DrawRoundedRectangle(el.Position.X, el.Position.Y, el.Size.W, el.Size.H, el.CornerRadius)
		} else {
			dc.DrawRectangle(el.Position.X, el.Position.Y, el.Size.W, el.Size.H)
		}
	}

	// 7. Draw the constructed path
	hasFill := fillColor != "" && fillColor != "transparent"
	hasStroke := borderWidth > 0 && borderColor != "" && borderColor != "transparent"

	if hasFill {
		c := rc.parseColor(fillColor)
		if c != "" {
			dc.SetHexColor(c)
			if hasStroke {
				dc.FillPreserve()
			} else {
				dc.Fill()
			}
		}
	}

	if hasStroke {
		c := rc.parseColor(borderColor)
		if c != "" {
			dc.SetHexColor(c)
			dc.SetLineWidth(borderWidth)
			
			if el.Border != nil {
				dashes := parseLineStyle(el.Border.Style, borderWidth)
				if len(dashes) > 0 {
					dc.SetDash(dashes...)
				}
			}
			
			dc.Stroke()
		} else {
			dc.ClearPath()
		}
	} else if !hasFill {
		// Neither fill nor stroke is applied, clear the path
		dc.ClearPath()
	}
}
