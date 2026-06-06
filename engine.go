// Copyright (c) 2026, Anton Bespalov. Licensed under the MIT License.
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

// DataProvider defines an interface for lazy evaluation of telemetry values, charts, and thresholds.
type DataProvider interface {
	GetValue(key string) (interface{}, bool)
	GetChart(source string) ([]float64, bool)
	ResolveThreshold(condition string, chartSource string) (float64, bool)
	EnrichTelemetry(source string, value float64) map[string]interface{}
}

// RenderData holds flattened telemetry and pre-sampled charts
type RenderData struct {
	Values map[string]interface{}
	Charts map[string][]float64
}

// MapDataProvider is a compatibility wrapper that implements DataProvider
// using a static RenderData map and engine callbacks.
type MapDataProvider struct {
	Data     RenderData
	Resolver ValueResolver
	Enricher TelemetryEnricher
}

func (p *MapDataProvider) GetValue(key string) (interface{}, bool) {
	val, ok := p.Data.Values[key]
	return val, ok
}

func (p *MapDataProvider) GetChart(source string) ([]float64, bool) {
	chart, ok := p.Data.Charts[source]
	return chart, ok
}

func (p *MapDataProvider) ResolveThreshold(condition string, chartSource string) (float64, bool) {
	if p.Resolver != nil {
		return p.Resolver(condition, chartSource)
	}
	return 0, false
}

func (p *MapDataProvider) EnrichTelemetry(source string, value float64) map[string]interface{} {
	if p.Enricher != nil {
		return p.Enricher(source, value)
	}
	return nil
}

// RenderContext holds state for drawing elements on the GG context.
type RenderContext struct {
	dc        *gg.Context
	cfg       *LayoutConfig
	provider  DataProvider
	fontCache map[string]font.Face
	engine    *Engine
}

// Engine encapsulates the renderer and registered plugins.
type Engine struct {
	plugins   map[string]CustomRenderer
	Resolver  ValueResolver     // Deprecated: use DataProvider instead
	Enricher  TelemetryEnricher // Deprecated: use DataProvider instead
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
