// Copyright (c) 2026, Anton Bespalov
package csirender

import (
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"sync"

	"github.com/fogleman/gg"
	"golang.org/x/image/font"
	_ "golang.org/x/image/tiff"
)

// ValueResolver defines a callback to resolve condition strings to concrete chart values.
type ValueResolver func(condition string, chartSource string) (float64, bool)

// TelemetryEnricher is an optional callback that injects derived key-value pairs
// into the local telemetry map when evaluating per-bar or per-label thresholds.
// It receives the chart source name and the current bar/label value, and returns
// any additional keys that should be visible in the condition expression.
type TelemetryEnricher func(source string, value float64) map[string]interface{}

// CustomRenderer allows the parent application to draw domain-specific elements.
type CustomRenderer interface {
	RenderElement(dc *gg.Context, rc *RenderContext, el Element) error
}

// RenderData holds flattened telemetry and pre-sampled charts
type RenderData struct {
	Values map[string]interface{}
	Charts map[string][]float64
}

// RenderContext holds state for drawing elements on the GG context.
type RenderContext struct {
	dc        *gg.Context
	cfg       *LayoutConfig
	data      RenderData
	fontCache map[string]font.Face
	Resolver  ValueResolver
	Enricher  TelemetryEnricher
	engine    *Engine
}

// Engine encapsulates the renderer and registered plugins.
type Engine struct {
	plugins   map[string]CustomRenderer
	Resolver  ValueResolver
	Enricher  TelemetryEnricher
	imgCache  map[string]image.Image
	cacheLock sync.RWMutex
}

func New() *Engine {
	return &Engine{
		plugins:  make(map[string]CustomRenderer),
		imgCache: make(map[string]image.Image),
	}
}

// RegisterRenderer registers a plugin to handle a specific element type.
func (e *Engine) RegisterRenderer(typ string, renderer CustomRenderer) {
	e.plugins[typ] = renderer
}

// GetImage loads an image from disk with caching.
func (e *Engine) GetImage(path string) (image.Image, error) {
	e.cacheLock.RLock()
	img, ok := e.imgCache[path]
	e.cacheLock.RUnlock()
	if ok {
		return img, nil
	}

	e.cacheLock.Lock()
	defer e.cacheLock.Unlock()
	// Double-check after acquiring write lock
	img, ok = e.imgCache[path]
	if ok {
		return img, nil
	}

	img, err := gg.LoadImage(path)
	if err != nil {
		return nil, err
	}

	e.imgCache[path] = img
	return img, nil
}
