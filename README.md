# csirender

`csirender` is an extensible, declarative 2D rendering engine written in Go. It enables you to define dynamic dashboard layouts in YAML or JSON, and render them to standard image formats (PNG, BMP) or EPD raw data streams.

## Features

- **Declarative Configuration**: Define visual elements (Text, Charts, Value Cards, Separators, Images, Shapes) directly in YAML or JSON.
- **Dynamic Styling**: Built-in support for conditional rendering (Thresholds) based on telemetry values or expressions. This allows changing colors, background, image paths, and opacity dynamically.
- **Image Support**: Render static or dynamic images (PNG, JPEG, GIF, TIFF) with scaling modes (fit, stretch, center) and transparency.
- **Shapes & Primitives**: Draw rectangles, circles, ellipses, and polygons with dynamic rotations bound to telemetry (e.g. for compass arrows).
- **Extensible Architecture**: Implement the `CustomRenderer` interface to draw your own domain-specific components.
- **Optimized Loading**: Generic `Parser[T]` seamlessly loads and caches configurations, watching file modification times to prevent unnecessary re-parsing.

## Basic Usage

### 1. Define a Configuration

`csirender` is designed to be embedded in your business application. You can parse the config directly or use the caching parser.
See the [`res/`](res/) directory for full configuration examples and client-server implementation code.

```go
package main

import (
    "github.com/ABespalov/csirender"
    "fmt"
)

// AppConfig embeds csirender.LayoutConfig and adds application-specific fields
type AppConfig struct {
    csirender.LayoutConfig `yaml:",inline" json:",inline"`
    MyBusinessLogic string `yaml:"my_logic" json:"my_logic"`
}

var parser = csirender.NewParser[*AppConfig]()

func main() {
    cfg, err := parser.Parse("layout.yaml")
    if err != nil {
        panic(err)
    }
    
    fmt.Println("Screen width:", cfg.Screen.Width)
    fmt.Println("My logic:", cfg.MyBusinessLogic)
}
```

### 2. Render the Layout

Create an `Engine`, feed it your telemetry data, and call `Render`:

```go
engine := csirender.New()

data := csirender.RenderData{
    Values: map[string]interface{}{
        "temperature": 22.5,
    },
    Charts: make(map[string][]float64),
}

img, err := engine.Render(&cfg.LayoutConfig, data)
if err != nil {
    panic(err)
}

// Save or display the image
```

## Plugin System

If you need a custom element type (e.g., a "gauge" or a "progress_bar"), you can register a custom renderer:

```go
type MyGaugeRenderer struct {}

func (r *MyGaugeRenderer) RenderElement(dc *gg.Context, rc *csirender.RenderContext, el csirender.Element) error {
    genericEl := el.(*csirender.GenericElement)
    // Read arbitrary parameters from genericEl.Raw
    // Draw the gauge using dc (gg.Context)
    return nil
}

// Registration
engine.RegisterRenderer("gauge", &MyGaugeRenderer{})
```

## License

Copyright (c) 2026, Anton Bespalov.
