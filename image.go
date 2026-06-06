// Copyright (c) 2026, Anton Bespalov. Licensed under the MIT License.
package csirender

import (
	"image"
	"image/color"
	"image/draw"
	"math"

	"github.com/fogleman/gg"
	xdraw "golang.org/x/image/draw"
	"log"
)

func renderImage(dc *gg.Context, rc *RenderContext, el *ImageElement) {
	// Determine the base path and opacity
	path := el.Path
	opacity := 1.0

	// 1. Global default opacity
	if rc.cfg.Defaults.Opacity != nil {
		opacity = *rc.cfg.Defaults.Opacity
	}
	// 2. Local element opacity
	if el.Opacity != nil {
		opacity = *el.Opacity
	}

	// 3. Evaluate thresholds
	if len(el.Thresholds) > 0 {
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

		for _, th := range el.Thresholds {
			if evalCondition(th.Condition, localTel) {
				if th.Path != "" {
					path = th.Path
				}
				if th.Opacity != nil {
					opacity = *th.Opacity
				}
				break
			}
		}
	}

	if path == "" {
		return // Nothing to draw
	}

	// Load image from cache or disk
	srcImg, err := rc.engine.GetImage(path)
	if err != nil {
		log.Printf("csirender: failed to load image path=%s err=%v", path, err)
		return
	}

	srcBounds := srcImg.Bounds()
	srcW := float64(srcBounds.Dx())
	srcH := float64(srcBounds.Dy())

	if srcW == 0 || srcH == 0 || el.Size.W == 0 || el.Size.H == 0 {
		return
	}

	// Calculate scaling and placement
	var drawW, drawH float64
	mode := el.Mode
	if mode == "" {
		mode = "fit" // default mode
	}

	switch mode {
	case "stretch":
		drawW = el.Size.W
		drawH = el.Size.H
	case "center":
		drawW = srcW
		drawH = srcH
	case "fit":
		fallthrough
	default:
		scaleX := el.Size.W / srcW
		scaleY := el.Size.H / srcH
		scale := math.Min(scaleX, scaleY)
		drawW = srcW * scale
		drawH = srcH * scale
	}

	// Calculate alignment
	var drawX, drawY float64

	switch el.Align.H {
	case "left":
		drawX = el.Position.X
	case "right":
		drawX = el.Position.X + el.Size.W - drawW
	case "center":
		fallthrough
	default:
		drawX = el.Position.X + (el.Size.W-drawW)/2.0
	}

	switch el.Align.V {
	case "top":
		drawY = el.Position.Y
	case "bottom":
		drawY = el.Position.Y + el.Size.H - drawH
	case "center":
		fallthrough
	default:
		drawY = el.Position.Y + (el.Size.H-drawH)/2.0
	}

	// Dest rect
	rect := image.Rect(
		int(math.Round(drawX)),
		int(math.Round(drawY)),
		int(math.Round(drawX+drawW)),
		int(math.Round(drawY+drawH)),
	)

	// Scale image using high quality ApproxBiLinear
	scaledImg := image.NewRGBA(image.Rect(0, 0, rect.Dx(), rect.Dy()))
	xdraw.ApproxBiLinear.Scale(scaledImg, scaledImg.Bounds(), srcImg, srcBounds, xdraw.Over, nil)

	// Draw on canvas
	canvas := dc.Image().(*image.RGBA)

	if opacity >= 1.0 {
		// Draw opaque (preserving alpha of the image itself)
		draw.Draw(canvas, rect, scaledImg, image.Point{0, 0}, draw.Over)
	} else if opacity > 0.0 {
		// Draw with global alpha mask
		mask := image.NewUniform(color.Alpha{A: uint8(opacity * 255.0)})
		draw.DrawMask(canvas, rect, scaledImg, image.Point{0, 0}, mask, image.Point{0, 0}, draw.Over)
	}
}
