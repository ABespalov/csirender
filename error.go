// Copyright (c) 2026, Anton Bespalov. Licensed under the MIT License.
package csirender

import (
	"bytes"
	"image/color"
	"image/png"
	"strings"

	"github.com/fogleman/gg"
	"golang.org/x/image/font/opentype"
)

func wrapText(dc *gg.Context, text string, maxWidth float64) []string {
	var lines []string
	for _, paragraph := range strings.Split(text, "\n") {
		words := strings.Fields(paragraph)
		if len(words) == 0 {
			lines = append(lines, "")
			continue
		}
		var currentLine string
		for _, word := range words {
			wordWidth, _ := dc.MeasureString(word)
			if wordWidth > maxWidth {
				if currentLine != "" {
					lines = append(lines, currentLine)
					currentLine = ""
				}
				var currentPiece string
				for _, r := range word {
					charStr := string(r)
					w, _ := dc.MeasureString(currentPiece + charStr)
					if w > maxWidth && currentPiece != "" {
						lines = append(lines, currentPiece)
						currentPiece = charStr
					} else {
						currentPiece += charStr
					}
				}
				currentLine = currentPiece
			} else {
				testLine := currentLine
				if testLine != "" {
					testLine += " "
				}
				testLine += word
				w, _ := dc.MeasureString(testLine)
				if w > maxWidth && currentLine != "" {
					lines = append(lines, currentLine)
					currentLine = word
				} else {
					currentLine = testLine
				}
			}
		}
		if currentLine != "" {
			lines = append(lines, currentLine)
		}
	}
	return lines
}

// RenderErrorImage generates a safe fallback image containing error text.
func RenderErrorImage(width, height int, errMsg string, format string, mapping map[string][]int, palette map[string]string, errorFontSize, marginX, marginY int) []byte {
	if width <= 0 {
		width = 400
	}
	if height <= 0 {
		height = 300
	}
	if marginX < 0 {
		marginX = 20
	}
	if marginY < 0 {
		marginY = 40
	}
	if errorFontSize <= 0 {
		errorFontSize = 5 // Default fallback ID
	}

	fontFallback := getFallbackFont(errorFontSize)
	ptSize := fontFallback.PtSize
	if _, ok := embeddedFonts[errorFontSize]; !ok {
		ptSize = float64(errorFontSize)
	}

	dc := gg.NewContext(width, height)
	dc.SetColor(color.RGBA{255, 255, 255, 255})
	dc.Clear()
	dc.SetColor(color.RGBA{0, 0, 0, 255})

	face, _ := opentype.NewFace(fontFallback.Font, &opentype.FaceOptions{
		Size:    ptSize,
		DPI:     72,
		Hinting: 0,
	})
	dc.SetFontFace(face)

	lines := wrapText(dc, errMsg, float64(width-2*marginX))
	y := float64(marginY) + ptSize
	for _, line := range lines {
		dc.DrawString(line, float64(marginX), y)
		y += ptSize + 4
	}

	res, err := EncodeImage(dc.Image(), format, mapping, palette)
	if err != nil {
		// Extreme fallback: plain black/white raw bytes or PNG
		var buf bytes.Buffer
		_ = png.Encode(&buf, dc.Image())
		return buf.Bytes()
	}
	return res
}
