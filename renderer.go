// Copyright (c) 2026, Anton Bespalov. Licensed under the MIT License.
package csirender

import (
	"fmt"
	"image"
	"image/color"

	"github.com/fogleman/gg"
	"golang.org/x/image/font"
)

// Render executes the full dashboard rendering pass and returns an image.
// Deprecated: use RenderWithProvider instead for lazy evaluation.
func (e *Engine) Render(cfg *LayoutConfig, data RenderData) (image.Image, error) {
	provider := &MapDataProvider{
		Data:     data,
		Resolver: e.Resolver,
		Enricher: e.Enricher,
	}
	return e.RenderWithProvider(cfg, provider)
}

// RenderWithProvider executes the full dashboard rendering pass using a lazy DataProvider.
func (e *Engine) RenderWithProvider(cfg *LayoutConfig, provider DataProvider) (image.Image, error) {
	if cfg.Screen.Width <= 0 || cfg.Screen.Height <= 0 {
		return nil, fmt.Errorf("invalid screen dimensions: %dx%d", cfg.Screen.Width, cfg.Screen.Height)
	}

	rc := &RenderContext{
		dc:        gg.NewContext(cfg.Screen.Width, cfg.Screen.Height),
		cfg:       cfg,
		provider:  provider,
		fontCache: make(map[string]font.Face),
		engine:    e,
	}

	// 1. Fill base canvas background
	if cfg.Output.Background != "" {
		if c, err := parseHexColor(rc.parseColor(cfg.Output.Background)); err == nil {
			rc.dc.SetColor(c)
		} else {
			rc.dc.SetColor(color.RGBA{255, 255, 255, 255})
		}
	} else {
		rc.dc.SetColor(color.RGBA{255, 255, 255, 255})
	}
	rc.dc.Clear()

	// 2. Iterate elements and render them
	for i := range cfg.Layout {
		el := cfg.Layout[i].Element

		// Check if a plugin handles this element type
		if plugin, ok := e.plugins[el.GetType()]; ok {
			err := plugin.RenderElement(rc.dc, rc, el)
			if err != nil {
				return nil, fmt.Errorf("plugin %s failed: %w", el.GetType(), err)
			}
			continue
		}

		// Save state before drawing element
		rc.dc.Push()

		// Built-in elements
		switch v := el.(type) {
		case *ValueCardElement:
			renderValueCard(rc.dc, rc, v)
		case *TextElement:
			rc.loadFont(v.Font.Face)
			renderText(rc.dc, rc, v)
		case *LineElement:
			renderLine(rc.dc, rc, v)
		case *ChartElement:
			renderChart(rc.dc, rc, v)
		case *ImageElement:
			renderImage(rc.dc, rc, v)
		case *ShapeElement:
			renderShape(rc.dc, rc, v)
		default:
			// Unknown elements are skipped
		}

		// Restore graphics context
		rc.dc.Pop()
	}

	return rc.dc.Image(), nil
}
