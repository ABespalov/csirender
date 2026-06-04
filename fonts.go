// Copyright (c) 2026, Anton Bespalov. Licensed under the MIT License.
package csirender

import (
	_ "embed"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

//go:embed fonts/Tamzen5x9r.ttf
var tamzen5x9 []byte

//go:embed fonts/Tamzen6x12r.ttf
var tamzen6x12 []byte

//go:embed fonts/Tamzen7x13r.ttf
var tamzen7x13 []byte

//go:embed fonts/Tamzen7x14r.ttf
var tamzen7x14 []byte

//go:embed fonts/Tamzen8x15r.ttf
var tamzen8x15 []byte

//go:embed fonts/Tamzen8x16r.ttf
var tamzen8x16 []byte

//go:embed fonts/Tamzen10x20r.ttf
var tamzen10x20 []byte

type EmbeddedFont struct {
	Font   *opentype.Font
	PtSize float64
}

var embeddedFonts map[int]EmbeddedFont

func init() {
	embeddedFonts = make(map[int]EmbeddedFont)
	loadFontBytes(1, 9, tamzen5x9)
	loadFontBytes(2, 12, tamzen6x12)
	loadFontBytes(3, 13, tamzen7x13)
	loadFontBytes(4, 14, tamzen7x14)
	loadFontBytes(5, 15, tamzen8x15)
	loadFontBytes(7, 16, tamzen8x16)
	loadFontBytes(8, 20, tamzen10x20)
}

func loadFontBytes(id int, ptSize float64, b []byte) {
	f, err := opentype.Parse(b)
	if err != nil {
		panic("csirender: failed to parse embedded font")
	}
	embeddedFonts[id] = EmbeddedFont{Font: f, PtSize: ptSize}
}

// getFallbackFont returns the font object and its default point size.
func getFallbackFont(id int) EmbeddedFont {
	if emb, ok := embeddedFonts[id]; ok {
		return emb
	}
	return embeddedFonts[7]
}

// getFontFace loads a font face from the config or falls back to the embedded Tamzen font.
func (rc *RenderContext) getFontFace(fontName string) font.Face {
	if fontName == "" {
		return nil
	}
	if face, ok := rc.fontCache[fontName]; ok {
		return face
	}
	fontDef, ok := rc.cfg.Fonts[fontName]
	if !ok {
		return nil
	}

	// Try reading file relative to running directory
	fontBytes, err := os.ReadFile(fontDef.File)
	if err != nil {
		// Attempt resolution relative to executable directory
		exe, errExe := os.Executable()
		if errExe == nil {
			altPath := filepath.Join(filepath.Dir(exe), fontDef.File)
			fontBytes, err = os.ReadFile(altPath)
		}
	}

	if err != nil {
		// Gracefully fall back to embedded font face
		face, errFace := opentype.NewFace(getFallbackFont(7).Font, &opentype.FaceOptions{
			Size:    fontDef.Size,
			DPI:     72,
			Hinting: font.HintingFull,
		})
		if errFace == nil {
			rc.fontCache[fontName] = face
			return face
		}
		return nil
	}

	f, err := opentype.Parse(fontBytes)
	if err != nil {
		face, _ := opentype.NewFace(getFallbackFont(7).Font, &opentype.FaceOptions{
			Size:    fontDef.Size,
			DPI:     72,
			Hinting: font.HintingFull,
		})
		return face
	}

	hintStyle := font.HintingFull
	switch strings.ToLower(fontDef.Hinting) {
	case "none":
		hintStyle = font.HintingNone
	case "vertical":
		hintStyle = font.HintingVertical
	case "full":
		hintStyle = font.HintingFull
	}

	face, err := opentype.NewFace(f, &opentype.FaceOptions{
		Size:    fontDef.Size,
		DPI:     72,
		Hinting: hintStyle,
	})
	if err != nil {
		face, _ := opentype.NewFace(getFallbackFont(7).Font, &opentype.FaceOptions{
			Size:    fontDef.Size,
			DPI:     72,
			Hinting: font.HintingFull,
		})
		return face
	}

	rc.fontCache[fontName] = face
	return face
}

// loadFont applies the loaded font face to the GG context.
func (rc *RenderContext) loadFont(fontName string) {
	face := rc.getFontFace(fontName)
	if face != nil {
		rc.dc.SetFontFace(face)
	}
}

// getStringBounds calculates pixel width/height of a string.
func getStringBounds(face font.Face, s string) (xMin, yMin, xMax, yMax float64) {
	if face == nil || s == "" {
		return 0, 0, 0, 0
	}
	bounds, _ := font.BoundString(face, s)
	xMin = float64(bounds.Min.X) / 64.0
	yMin = float64(bounds.Min.Y) / 64.0
	xMax = float64(bounds.Max.X) / 64.0
	yMax = float64(bounds.Max.Y) / 64.0
	return
}
