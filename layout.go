// Copyright (c) 2026, Anton Bespalov. Licensed under the MIT License.
package csirender

import (
	"encoding/json"

	"gopkg.in/yaml.v3"
)

// Position defines the 2D coordinate of a visual element on the screen (in pixels).
type Position struct {
	X float64 `yaml:"x" json:"x"`
	Y float64 `yaml:"y" json:"y"`
}

// Size defines the dimensions of a visual container (in pixels).
type Size struct {
	W float64 `yaml:"w" json:"w"`
	H float64 `yaml:"h" json:"h"`
}

// FontStyle groups font family reference and foreground/background colors.
type FontStyle struct {
	Face string `yaml:"face" json:"face"`
	Fg   string `yaml:"fg" json:"fg"`
	Bg   string `yaml:"bg" json:"bg"`
}

// Align defines vertical, horizontal, and relative alignments, along with spacers.
type Align struct {
	H string  `yaml:"h" json:"h"` // Horizontal alignment: left, center, right
	V string  `yaml:"v" json:"v"` // Vertical alignment: top, center, bottom, under_center, under
	P string  `yaml:"p" json:"p"` // Placement relative to other elements: right, left, bottom, none
	S float64 `yaml:"s" json:"s"` // Spacing/gap size in pixels
}

// Margin defines external spacing around components.
type Margin struct {
	T float64 `yaml:"t" json:"t"` // Top margin
	B float64 `yaml:"b" json:"b"` // Bottom margin
	L float64 `yaml:"l" json:"l"` // Left margin
	R float64 `yaml:"r" json:"r"` // Right margin
}

// LineConfig defines shared visual properties for lines.
type LineConfig struct {
	Color string   `yaml:"color,omitempty" json:"color,omitempty"` // Stroke color
	T     *float64 `yaml:"t,omitempty" json:"t,omitempty"`         // Line thickness
	Style string   `yaml:"style,omitempty" json:"style,omitempty"` // Line style: solid, dashed, dotted, dashdot, or custom
}

// ThresholdLine configures a horizontal threshold line drawn on charts.
type ThresholdLine struct {
	Condition  string  `yaml:"condition,omitempty" json:"condition,omitempty"` // Expression to trigger the line
	Value      float64 `yaml:"value,omitempty" json:"value,omitempty"`         // Static threshold value
	LineConfig `yaml:",inline" json:",inline"`
}

// Threshold defines a condition-based styling override.
type Threshold struct {
	Condition string   `yaml:"condition,omitempty" json:"condition,omitempty"` // Comparison expression (e.g. "aqi.us.zone.index >= 3")
	Label     string   `yaml:"label,omitempty" json:"label,omitempty"`         // Custom text label on match
	Bg        string   `yaml:"bg,omitempty" json:"bg,omitempty"`               // Custom background color on match
	Fg        string   `yaml:"fg,omitempty" json:"fg,omitempty"`               // Custom foreground/text color on match
	Color     string   `yaml:"color,omitempty" json:"color,omitempty"`         // Hex color code reference
	Path      string   `yaml:"path,omitempty" json:"path,omitempty"`           // Image path override
	Opacity   *float64 `yaml:"opacity,omitempty" json:"opacity,omitempty"`     // Opacity override
}

// BorderConfig controls the border line and background of card components.
type BorderConfig struct {
	Bg         string      `yaml:"bg,omitempty" json:"bg,omitempty"`                 // Background fill color
	Padding    float64     `yaml:"padding,omitempty" json:"padding,omitempty"`       // Padding inside the border
	Affects    []string    `yaml:"affects,omitempty" json:"affects,omitempty"`       // Elements enclosed by the border
	Thresholds []Threshold `yaml:"thresholds,omitempty" json:"thresholds,omitempty"` // Dynamic styling overrides
	LineConfig `yaml:",inline" json:",inline"`
}

// ComponentConfig configures sub-components inside a complex widget like a value card.
type ComponentConfig struct {
	Font       FontStyle     `yaml:"font,omitempty" json:"font,omitempty"`             // Font settings
	Align      Align         `yaml:"align,omitempty" json:"align,omitempty"`           // Alignment settings
	Margin     Margin        `yaml:"margin,omitempty" json:"margin,omitempty"`         // Outer spacing
	Thresholds []Threshold   `yaml:"thresholds,omitempty" json:"thresholds,omitempty"` // Conditional styling overrides
	Border     *BorderConfig `yaml:"border,omitempty" json:"border,omitempty"`         // Optional border settings
	Template   string        `yaml:"template,omitempty" json:"template,omitempty"`     // Visual formatting template
	Parameter  string        `yaml:"parameter,omitempty" json:"parameter,omitempty"`   // Custom unit reference
}

// BaseElement contains fields common to all visual elements.
type BaseElement struct {
	Type       string            `yaml:"type" json:"type"`                                 // Element type
	Position   Position          `yaml:"position" json:"position"`                         // Offset position on canvas
	Size       Size              `yaml:"size" json:"size"`                                 // Layout box dimensions
	Source     string            `yaml:"source,omitempty" json:"source,omitempty"`         // Data mapping path (e.g. "pm25.value")
	Parameters map[string]string `yaml:"parameters,omitempty" json:"parameters,omitempty"` // Arbitrary key-value parameters
}

func (b *BaseElement) GetType() string       { return b.Type }
func (b *BaseElement) GetPosition() Position { return b.Position }
func (b *BaseElement) GetSize() Size         { return b.Size }

// TextElement represents a text component.
type TextElement struct {
	BaseElement `yaml:",inline" json:",inline"`
	Font        FontStyle `yaml:"font,omitempty" json:"font,omitempty"`
	Align       Align     `yaml:"align,omitempty" json:"align,omitempty"`
	Color       string    `yaml:"color,omitempty" json:"color,omitempty"`
	Template    string    `yaml:"template,omitempty" json:"template,omitempty"`
	FitText     *bool     `yaml:"fit_text,omitempty" json:"fit_text,omitempty"`
}

// ValueCardElement represents a complex value displaying component.
type ValueCardElement struct {
	BaseElement `yaml:",inline" json:",inline"`
	Value       ComponentConfig `yaml:"value,omitempty" json:"value,omitempty"`
	Label       ComponentConfig `yaml:"label,omitempty" json:"label,omitempty"`
	Unit        ComponentConfig `yaml:"unit,omitempty" json:"unit,omitempty"`
	Border      *BorderConfig   `yaml:"border,omitempty" json:"border,omitempty"`
	Format      string          `yaml:"format,omitempty" json:"format,omitempty"`
	Padding     float64         `yaml:"padding,omitempty" json:"padding,omitempty"`
	Radius      float64         `yaml:"radius,omitempty" json:"radius,omitempty"`
}

// LineElement renders a simple line or separator.
type LineElement struct {
	BaseElement `yaml:",inline" json:",inline"`
	LineConfig  `yaml:",inline" json:",inline"`
	L           *float64 `yaml:"l,omitempty" json:"l,omitempty"`
	Angle       *float64 `yaml:"angle,omitempty" json:"angle,omitempty"`
}

// ChartElement represents a graph or chart.
type ChartElement struct {
	BaseElement    `yaml:",inline" json:",inline"`
	Style          string                 `yaml:"style,omitempty" json:"style,omitempty"`
	Duration       string                 `yaml:"duration,omitempty" json:"duration,omitempty"`
	Points         int                    `yaml:"points,omitempty" json:"points,omitempty"`
	Axis           map[string]interface{} `yaml:"axis,omitempty" json:"axis,omitempty"`
	ThresholdLines []ThresholdLine        `yaml:"threshold_lines,omitempty" json:"threshold_lines,omitempty"`
	Thresholds     []Threshold            `yaml:"thresholds,omitempty" json:"thresholds,omitempty"`
	ChartType      string                 `yaml:"chart_type,omitempty" json:"chart_type,omitempty"` // Obsolete
}

// ImageElement represents a static or dynamic image.
type ImageElement struct {
	BaseElement `yaml:",inline" json:",inline"`
	Path        string      `yaml:"path,omitempty" json:"path,omitempty"`
	Mode        string      `yaml:"mode,omitempty" json:"mode,omitempty"`
	Opacity     *float64    `yaml:"opacity,omitempty" json:"opacity,omitempty"`
	Align       Align       `yaml:"align,omitempty" json:"align,omitempty"`
	Thresholds  []Threshold `yaml:"thresholds,omitempty" json:"thresholds,omitempty"`
}

// ShapeElement represents a geometric primitive (rectangle, circle, polygon).
type ShapeElement struct {
	BaseElement  `yaml:",inline" json:",inline"`
	Kind         string       `yaml:"kind,omitempty" json:"kind,omitempty"`                   // "rect", "circle", "ellipse", "polygon"
	Color        string        `yaml:"color,omitempty" json:"color,omitempty"`                 // Fill color
	Border       *BorderConfig `yaml:"border,omitempty" json:"border,omitempty"`               // Stroke/border settings
	CornerRadius float64      `yaml:"corner_radius,omitempty" json:"corner_radius,omitempty"` // For "rect"
	Sides        int          `yaml:"sides,omitempty" json:"sides,omitempty"`                 // For "polygon"
	Rotation     float64      `yaml:"rotation,omitempty" json:"rotation,omitempty"`           // Static rotation in degrees
	RotationExpr string       `yaml:"rotation_expr,omitempty" json:"rotation_expr,omitempty"` // Dynamic rotation expression/source
	Thresholds   []Threshold  `yaml:"thresholds,omitempty" json:"thresholds,omitempty"`
}

// GenericElement is a fallback for custom plugin elements.
type GenericElement struct {
	BaseElement `yaml:",inline" json:",inline"`
	Raw         map[string]interface{} `yaml:",inline" json:",inline"`
}

// Element is a generic node representing a visual item inside a layout.
type Element interface {
	GetType() string
	GetPosition() Position
	GetSize() Size
}

// ElementWrapper is a helper type for unmarshaling elements into their concrete types.
type ElementWrapper struct {
	Element Element
}

func (w *ElementWrapper) UnmarshalYAML(node *yaml.Node) error {
	var raw struct {
		Type string `yaml:"type" json:"type"`
	}
	if err := node.Decode(&raw); err != nil {
		return err
	}

	var el Element
	switch raw.Type {
	case "text":
		el = &TextElement{}
	case "value":
		el = &ValueCardElement{}
	case "line":
		el = &LineElement{}
	case "chart":
		el = &ChartElement{}
	case "image":
		el = &ImageElement{}
	case "shape":
		el = &ShapeElement{}
	default:
		el = &GenericElement{}
	}

	if err := node.Decode(el); err != nil {
		return err
	}
	w.Element = el
	return nil
}

// UnmarshalJSON implements the json.Unmarshaler interface for ElementWrapper.
// It dynamically detects the "type" field and unmarshals the JSON into the appropriate concrete struct.
func (w *ElementWrapper) UnmarshalJSON(data []byte) error {
	var raw struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	var el Element
	switch raw.Type {
	case "text":
		el = &TextElement{}
	case "value":
		el = &ValueCardElement{}
	case "line":
		el = &LineElement{}
	case "chart":
		el = &ChartElement{}
	case "image":
		el = &ImageElement{}
	case "shape":
		el = &ShapeElement{}
	default:
		el = &GenericElement{}
	}

	if err := json.Unmarshal(data, el); err != nil {
		return err
	}
	w.Element = el
	return nil
}

// FontConfig defines the TrueType font file and basic properties.
type FontConfig struct {
	File    string  `yaml:"file" json:"file"`       // Path to TrueType font file
	Size    float64 `yaml:"size" json:"size"`       // Default size of font
	Hinting string  `yaml:"hinting" json:"hinting"` // Hinting mode: none, vertical, full
}

// ScreenConfig defines the canvas dimensions and palette.
type ScreenConfig struct {
	Width        int               `yaml:"width" json:"width"`                 // Canvas width in pixels
	Height       int               `yaml:"height" json:"height"`               // Canvas height in pixels
	AntiAliasing bool              `yaml:"anti_aliasing" json:"anti_aliasing"` // Enable/disable anti-aliasing
	Palette      map[string]string `yaml:"palette" json:"palette"`             // Logical palette color mappings
}

// ErrorConfig defines how errors are rendered.
type ErrorConfig struct {
	Margin struct {
		X int `yaml:"x" json:"x"`
		Y int `yaml:"y" json:"y"`
	} `yaml:"margin" json:"margin"`
	Size int `yaml:"size" json:"size"`
}

// OutputConfig configures serialization parameters.
type OutputConfig struct {
	Format     string           `yaml:"format" json:"format"`         // Serialization format: epd_raw, png, bmp
	Background string           `yaml:"background" json:"background"` // Base fill color
	Mapping    map[string][]int `yaml:"mapping" json:"mapping"`       // Bit packing map for EPD outputs
}

// LayoutDefaults sets fallback values for visual properties.
type LayoutDefaults struct {
	FitText *bool    `yaml:"fit_text,omitempty" json:"fit_text,omitempty"`
	Align   Align    `yaml:"align,omitempty" json:"align,omitempty"`
	Opacity *float64 `yaml:"opacity,omitempty" json:"opacity,omitempty"`
}

// LayoutConfig represents the pure visual rendering configuration.
type LayoutConfig struct {
	Screen   ScreenConfig          `yaml:"screen" json:"screen"`
	Error    ErrorConfig           `yaml:"error" json:"error"`
	Defaults LayoutDefaults        `yaml:"defaults,omitempty" json:"defaults,omitempty"`
	Output   OutputConfig          `yaml:"output" json:"output"`
	Fonts    map[string]FontConfig `yaml:"fonts" json:"fonts"`
	Layout   []ElementWrapper      `yaml:"layout" json:"layout"` // Sequence of elements to draw
}
