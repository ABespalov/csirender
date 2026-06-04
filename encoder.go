// Copyright (c) 2026, Anton Bespalov
package csirender

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"strings"

	"golang.org/x/image/bmp"
)

// PackEPDRaw takes an image.Image and packs its pixels into a byte array
// corresponding to the physical color indices of the target E-Ink display.
func PackEPDRaw(img image.Image, width, height int, mapping map[string][]int, palette map[string]string) []byte {
	// First pass: extract palette keys from mappings
	var mappedColors []color.RGBA
	var mappedBits [][]int
	var defaultBits []int

	for hexOrKey, bitValues := range mapping {
		if hexOrKey == "default" {
			defaultBits = bitValues
			continue
		}
		// Resolve through palette if necessary
		actualHex := hexOrKey
		if palette != nil {
			if val, ok := palette[hexOrKey]; ok {
				actualHex = val
			}
		}

		c, err := parseHexColor(actualHex)
		if err == nil {
			mappedColors = append(mappedColors, c)
			mappedBits = append(mappedBits, bitValues)
		}
	}

	if len(defaultBits) == 0 {
		defaultBits = []int{0} // fallback to 0 bit array
	}

	totalBits := 0
	for _, bits := range mapping {
		if len(bits) > totalBits {
			totalBits = len(bits)
		}
	}
	bytesCount := totalBits

	var raw []byte
	currentBytes := make([]byte, bytesCount)
	pixelIndex := 0

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			p := img.At(x, y)
			r, g, b, _ := p.RGBA()
			pxColor := color.RGBA{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), 255}

			bestDist := 1e9
			bestBits := defaultBits

			for i, mc := range mappedColors {
				d := colorDistance(pxColor, mc)
				if d < bestDist {
					bestDist = d
					bestBits = mappedBits[i]
				}
			}

			// Pack resolved bits into streams
			for planeIdx := 0; planeIdx < bytesCount; planeIdx++ {
				bitVal := 0
				if planeIdx < len(bestBits) {
					bitVal = bestBits[planeIdx]
				}

				if bitVal != 0 {
					currentBytes[planeIdx] |= (1 << (7 - (pixelIndex % 8)))
				}
			}

			pixelIndex++
			if pixelIndex%8 == 0 {
				for i := 0; i < bytesCount; i++ {
					raw = append(raw, currentBytes[i])
					currentBytes[i] = 0
				}
			}
		}
	}

	// Flush remaining bits
	if pixelIndex%8 != 0 {
		for i := 0; i < bytesCount; i++ {
			raw = append(raw, currentBytes[i])
		}
	}

	return raw
}

// EncodeImage wraps final formatting for the target output stream.
func EncodeImage(img image.Image, format string, mapping map[string][]int, palette map[string]string) ([]byte, error) {
	var buf bytes.Buffer
	format = strings.ToLower(strings.TrimSpace(format))

	switch format {
	case "png":
		if err := png.Encode(&buf, img); err != nil {
			return nil, fmt.Errorf("encoding PNG: %w", err)
		}
	case "bmp":
		if err := bmp.Encode(&buf, img); err != nil {
			return nil, fmt.Errorf("encoding BMP: %w", err)
		}
	case "epd_raw":
		if mapping == nil {
			return nil, fmt.Errorf("missing 'mapping' in output configuration for epd_raw format")
		}
		raw := PackEPDRaw(img, img.Bounds().Dx(), img.Bounds().Dy(), mapping, palette)
		return raw, nil
	default:
		// Default to PNG
		if err := png.Encode(&buf, img); err != nil {
			return nil, fmt.Errorf("encoding PNG: %w", err)
		}
	}

	return buf.Bytes(), nil
}
