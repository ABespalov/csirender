// Copyright (c) 2026, Anton Bespalov. Licensed under the MIT License.
package main

import (
	"bytes"
	"image/png"
	"log"
	"net/http"

	"github.com/ABespalov/csirender"
)

// AppConfig defines the structure for our application, embedding the layout config.
type AppConfig struct {
	csirender.LayoutConfig `yaml:",inline" json:",inline"`
}

var parser = csirender.NewParser[*AppConfig]()
var engine = csirender.New()

func main() {
	http.HandleFunc("/render", handleRender)

	log.Println("Server listening on :8080...")
	log.Println("Try visiting http://localhost:8080/render")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

func handleRender(w http.ResponseWriter, r *http.Request) {
	// 1. Load the layout configuration.
	// Since we use the csirender parser, this is heavily cached.
	// We go up one level since this runs in examples/server/
	cfg, err := parser.Parse("../assets/example.yaml")
	if err != nil {
		http.Error(w, "Failed to load config: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 2. Prepare telemetry data
	// In a real application, you would fetch this from sensors or a database.
	data := &csirender.MapDataProvider{
		Data: csirender.RenderData{
			Values: map[string]interface{}{
				"temperature":       28.5,
				"weather_condition": "sun",
			},
			Charts: map[string][]float64{
				"temperature_history": {20, 22, 25, 27, 28.5},
			},
		},
	}

	// 3. Render the dashboard
	img, err := engine.RenderWithProvider(&cfg.LayoutConfig, data)
	if err != nil {
		http.Error(w, "Failed to render: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 4. Encode as PNG and send to client
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		http.Error(w, "Failed to encode PNG: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.Write(buf.Bytes())
}
