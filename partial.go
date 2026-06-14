// Copyright (c) 2026, Anton Bespalov. Licensed under the MIT License.
package csirender

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
)

type rect struct {
	minX, minY, maxX, maxY int
}

func rectsOverlap(r1, r2 rect) bool {
	return r1.minX < r2.maxX && r1.maxX > r2.minX && r1.minY < r2.maxY && r1.maxY > r2.minY
}

func mergeRects(r1, r2 rect) rect {
	return rect{
		minX: min(r1.minX, r2.minX),
		minY: min(r1.minY, r2.minY),
		maxX: max(r1.maxX, r2.maxX),
		maxY: max(r1.maxY, r2.maxY),
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// EncodePartialImage encodes an image using the epd_pr (partial refresh) format.
// It compares oldImg and newImg to find bounding boxes of differences on an 8x8 grid.
// If oldImg is nil, it encodes the entire newImg as a single tile.
func EncodePartialImage(oldImg, newImg image.Image, format string, mapping map[string][]int, palette map[string]string) ([]byte, error) {
	if mapping == nil {
		return nil, fmt.Errorf("missing 'mapping' in output configuration for epd_pr format")
	}

	bounds := newImg.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	var components []rect

	if oldImg == nil || oldImg.Bounds() != bounds {
		// Full refresh
		components = append(components, rect{minX: 0, minY: 0, maxX: width, maxY: height})
	} else {
		// Partial refresh: find dirty 8x8 blocks
		gridW := (width + 7) / 8
		gridH := (height + 7) / 8

		dirty := make([][]bool, gridH)
		for i := range dirty {
			dirty[i] = make([]bool, gridW)
		}

		for gy := 0; gy < gridH; gy++ {
			for gx := 0; gx < gridW; gx++ {
				isDirty := false
				for y := gy * 8; y < (gy+1)*8 && y < height; y++ {
					for x := gx * 8; x < (gx+1)*8 && x < width; x++ {
						c1 := oldImg.At(x, y)
						c2 := newImg.At(x, y)
						r1, g1, b1, _ := c1.RGBA()
						r2, g2, b2, _ := c2.RGBA()
						if r1 != r2 || g1 != g2 || b1 != b2 {
							isDirty = true
							break
						}
					}
					if isDirty {
						break
					}
				}
				dirty[gy][gx] = isDirty
			}
		}

		// Find connected components
		visited := make([][]bool, gridH)
		for i := range visited {
			visited[i] = make([]bool, gridW)
		}

		for gy := 0; gy < gridH; gy++ {
			for gx := 0; gx < gridW; gx++ {
				if dirty[gy][gx] && !visited[gy][gx] {
					q := [][2]int{{gx, gy}}
					visited[gy][gx] = true

					minX, maxX := gx, gx
					minY, maxY := gy, gy

					for len(q) > 0 {
						curr := q[0]
						q = q[1:]
						cx, cy := curr[0], curr[1]

						if cx < minX {
							minX = cx
						}
						if cx > maxX {
							maxX = cx
						}
						if cy < minY {
							minY = cy
						}
						if cy > maxY {
							maxY = cy
						}

						for dy := -1; dy <= 1; dy++ {
							for dx := -1; dx <= 1; dx++ {
								nx, ny := cx+dx, cy+dy
								if nx >= 0 && nx < gridW && ny >= 0 && ny < gridH {
									if dirty[ny][nx] && !visited[ny][nx] {
										visited[ny][nx] = true
										q = append(q, [2]int{nx, ny})
									}
								}
							}
						}
					}

					rc := rect{
						minX: minX * 8,
						minY: minY * 8,
						maxX: (maxX + 1) * 8,
						maxY: (maxY + 1) * 8,
					}
					if rc.maxX > width {
						rc.maxX = width
					}
					if rc.maxY > height {
						rc.maxY = height
					}
					components = append(components, rc)
				}
			}
		}

		// Merge overlapping bounding boxes
		for {
			merged := false
			for i := 0; i < len(components); i++ {
				for j := i + 1; j < len(components); j++ {
					if rectsOverlap(components[i], components[j]) {
						components[i] = mergeRects(components[i], components[j])
						components = append(components[:j], components[j+1:]...)
						merged = true
						break
					}
				}
				if merged {
					break
				}
			}
			if !merged {
				break
			}
		}
	}

	// Pack tiles
	var buf bytes.Buffer
	// Header: Number of tiles (uint16, little-endian)
	if err := binary.Write(&buf, binary.LittleEndian, uint16(len(components))); err != nil {
		return nil, err
	}

	// Prepare palette mapping cache (same as PackEPDRaw)
	var mappedColors []color.RGBA
	var mappedBits [][]int
	var defaultBits []int

	for hexOrKey, bitValues := range mapping {
		if hexOrKey == "default" {
			defaultBits = bitValues
			continue
		}
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
		defaultBits = []int{0}
	}

	totalBits := 0
	for _, bits := range mapping {
		if len(bits) > totalBits {
			totalBits = len(bits)
		}
	}
	bytesCount := totalBits

	// Pack each tile
	for _, rc := range components {
		tileW := rc.maxX - rc.minX
		tileH := rc.maxY - rc.minY

		if err := binary.Write(&buf, binary.LittleEndian, uint16(rc.minX)); err != nil {
			return nil, err
		}
		if err := binary.Write(&buf, binary.LittleEndian, uint16(rc.minY)); err != nil {
			return nil, err
		}
		if err := binary.Write(&buf, binary.LittleEndian, uint16(tileW)); err != nil {
			return nil, err
		}
		if err := binary.Write(&buf, binary.LittleEndian, uint16(tileH)); err != nil {
			return nil, err
		}

		currentBytes := make([]byte, bytesCount)
		pixelIndex := 0

		for y := rc.minY; y < rc.maxY; y++ {
			for x := rc.minX; x < rc.maxX; x++ {
				p := newImg.At(x, y)
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
						buf.WriteByte(currentBytes[i])
						currentBytes[i] = 0
					}
				}
			}
		}

		if pixelIndex%8 != 0 {
			for i := 0; i < bytesCount; i++ {
				buf.WriteByte(currentBytes[i])
			}
		}
	}

	return buf.Bytes(), nil
}
