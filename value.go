// Copyright (c) 2026, Anton Bespalov. Licensed under the MIT License.
package csirender

import (
	"fmt"
	"github.com/fogleman/gg"
	"math"
)

func renderValueCard(dc *gg.Context, rc *RenderContext, el *ValueCardElement) {
	valIntf := rc.data.Values[el.Source]
	valFloat := 0.0
	valStr := ""
	isString := false

	if valIntf != nil {
		if s, ok := valIntf.(string); ok {
			valStr = s
			isString = true
		} else {
			switch v := valIntf.(type) {
			case float64:
				valFloat = v
			case int:
				valFloat = float64(v)
			case int64:
				valFloat = float64(v)
			}
		}
	}

	base := cleanBase(el.Source)
	label := ""
	if lblVal, ok := rc.data.Values[base+".label"]; ok {
		label = fmt.Sprintf("%v", lblVal)
	}
	if label == "" {
		label = el.Label.Template
	}

	unit := ""
	if el.Unit.Template != "" {
		unit = rc.resolvePlaceholders(el.Unit.Template)
	}
	if unit == "" && el.Unit.Parameter != "" {
		if val, ok := rc.data.Values[el.Unit.Parameter]; ok {
			unit = fmt.Sprintf("%v", val)
		}
	}
	if unit == "" {
		if val, ok := rc.data.Values[base+".unit"]; ok {
			unit = fmt.Sprintf("%v", val)
		}
	}

	if !isString {
		fmtFmt := "%.0f"
		if el.Format != "" {
			fmtFmt = el.Format
		}
		valStr = fmt.Sprintf(fmtFmt, valFloat)
	}

	rectX := el.Position.X
	rectY := el.Position.Y
	rectW := el.Size.W
	rectH := el.Size.H

	var valXMin, valYMin, valXMax, valYMax float64
	var hasVal bool
	if el.Value.Font.Face != "" && valStr != "" {
		hasVal = true
		vFace := rc.getFontFace(el.Value.Font.Face)
		valXMin, valYMin, valXMax, valYMax = getStringBounds(vFace, valStr)
	}

	var labelXMin, labelYMin, labelXMax, labelYMax float64
	var hasLabel bool
	if el.Label.Font.Face != "" && label != "" {
		hasLabel = true
		lFace := rc.getFontFace(el.Label.Font.Face)
		labelXMin, labelYMin, labelXMax, labelYMax = getStringBounds(lFace, label)
	}

	var unitXMin, unitYMin, unitXMax, unitYMax float64
	var hasUnit bool
	if el.Unit.Font.Face != "" && unit != "" && el.Unit.Align.P != "none" {
		hasUnit = true
		uFace := rc.getFontFace(el.Unit.Font.Face)
		unitXMin, unitYMin, unitXMax, unitYMax = getStringBounds(uFace, unit)
	}

	// Alignment of label
	var labelX, labelY float64
	if hasLabel {
		switch el.Label.Align.H {
		case "left":
			labelX = rectX - labelXMin
		case "right":
			labelX = rectX + rectW - labelXMax
		default:
			labelX = rectX + rectW/2 - (labelXMin+labelXMax)/2
		}
		labelX += el.Label.Margin.L - el.Label.Margin.R

		switch el.Label.Align.V {
		case "top":
			labelY = rectY - labelYMin
		case "bottom":
			labelY = rectY + rectH - labelYMax
		default:
			labelY = rectY + rectH/2 - (labelYMin+labelYMax)/2
		}
		labelY += el.Label.Margin.T - el.Label.Margin.B
	}

	// Value and Unit layout alignment
	unitPos := el.Unit.Align.P
	if unitPos == "" {
		unitPos = "right"
	}

	var valX, valY float64
	var unitX, unitY float64
	var rowXMin, rowXMax float64

	if hasVal {
		rowXMin = valXMin
		rowXMax = valXMax

		if hasUnit {
			if unitPos == "bottom" {
				valCenter := (valXMin + valXMax) / 2
				unitCenter := (unitXMin + unitXMax) / 2
				unitX = valCenter - unitCenter
				gapV := el.Unit.Align.S
				unitY = valYMax + gapV - unitYMin

				rowXMin = math.Min(valXMin, unitX+unitXMin)
				rowXMax = math.Max(valXMax, unitX+unitXMax)
			} else {
				gapH := el.Unit.Align.S
				unitX = valXMax + gapH - unitXMin

				uy := 0.0
				fracV, isFracV := parseAlignFraction(el.Unit.Align.V)
				if isFracV {
					valHeight := valYMax - valYMin
					uy = valYMax - valHeight*fracV
				} else {
					switch el.Unit.Align.V {
					case "top":
						uy = valYMin - unitYMin
					case "bottom":
						uy = valYMax - unitYMax
					case "idxUp":
						uy = valYMin
					case "idxDown":
						uy = valYMax
					default:
						uy = (valYMin+valYMax)/2 - (unitYMin+unitYMax)/2
					}
				}
				unitY = uy
				rowXMax = unitX + unitXMax
			}
		}
	} else if hasUnit {
		rowXMin = unitXMin
		rowXMax = unitXMax
		unitX = 0
		unitY = 0
	}

	if hasVal || hasUnit {
		switch el.Value.Align.H {
		case "left":
			valX = rectX - rowXMin
		case "right":
			valX = rectX + rectW - rowXMax
		default:
			valX = rectX + rectW/2 - (rowXMin+rowXMax)/2
		}
		valX += el.Value.Margin.L - el.Value.Margin.R

		if hasVal {
			switch el.Value.Align.V {
			case "top":
				valY = rectY - valYMin
			case "bottom":
				valY = rectY + rectH - valYMax
			default:
				valY = rectY + rectH/2 - (valYMin+valYMax)/2
			}
			valY += el.Value.Margin.T - el.Value.Margin.B
		} else if hasUnit {
			switch el.Value.Align.V {
			case "top":
				valY = rectY - unitYMin
			case "bottom":
				valY = rectY + rectH - unitYMax
			default:
				valY = rectY + rectH/2 - (unitYMin+unitYMax)/2
			}
			valY += el.Value.Margin.T - el.Value.Margin.B
		}
	}

	// Render border background and stroke
	var bX, bY, bW, bH, bR float64
	padding := 0.0
	if el.Border != nil {
		padding = el.Border.Padding
		bR = el.Radius
		if bR <= 0 {
			bR = 8.0
		}

		if len(el.Border.Affects) == 0 {
			bX = rectX - padding
			bY = rectY - padding
			bW = rectW + padding*2
			bH = rectH + padding*2
		} else {
			// Calculate bounds of elements inside the border
			minX, minY := math.MaxFloat64, math.MaxFloat64
			maxX, maxY := -math.MaxFloat64, -math.MaxFloat64
			hasAffected := false

			valLeft := valX + valXMin
			valRight := valX + valXMax
			valTop := valY + valYMin
			valBottom := valY + valYMax

			unitLeft := valX + unitX + unitXMin
			unitRight := valX + unitX + unitXMax
			unitTop := valY + unitY + unitYMin
			unitBottom := valY + unitY + unitYMax

			labelLeft := labelX + labelXMin
			labelRight := labelX + labelXMax
			labelTop := labelY + labelYMin
			labelBottom := labelY + labelYMax

			for _, affected := range el.Border.Affects {
				switch affected {
				case "value":
					if hasVal {
						minX = math.Min(minX, valLeft)
						maxX = math.Max(maxX, valRight)
						minY = math.Min(minY, valTop)
						maxY = math.Max(maxY, valBottom)
						hasAffected = true
					}
					if hasUnit {
						minX = math.Min(minX, unitLeft)
						maxX = math.Max(maxX, unitRight)
						minY = math.Min(minY, unitTop)
						maxY = math.Max(maxY, unitBottom)
						hasAffected = true
					}
				case "label":
					if hasLabel {
						minX = math.Min(minX, labelLeft)
						maxX = math.Max(maxX, labelRight)
						minY = math.Min(minY, labelTop)
						maxY = math.Max(maxY, labelBottom)
						hasAffected = true
					}
				case "unit":
					if hasUnit {
						minX = math.Min(minX, unitLeft)
						maxX = math.Max(maxX, unitRight)
						minY = math.Min(minY, unitTop)
						maxY = math.Max(maxY, unitBottom)
						hasAffected = true
					}
				}
			}

			if hasAffected {
				bX = minX - padding
				bY = minY - padding
				bW = (maxX - minX) + padding*2
				bH = (maxY - minY) + padding*2
			} else {
				bX = rectX - padding
				bY = rectY - padding
				bW = rectW + padding*2
				bH = rectH + padding*2
			}
		}

		bT := getThreshold(rc.data.Values, el.Border.Thresholds)
		bg := rc.parseColor(bT.Bg)
		if bg == "" {
			bg = rc.parseColor(el.Border.Bg)
		}
		fg := rc.parseColor(bT.Color)
		if fg == "" {
			fg = rc.parseColor(el.Border.Color)
		}

		if bg != "" {
			dc.SetHexColor(bg)
			dc.DrawRoundedRectangle(bX, bY, bW, bH, bR)
			dc.Fill()
		}
		var borderT float64
		if el.Border.T != nil {
			borderT = *el.Border.T
		}
		
		if fg != "" && borderT > 0 {
			dc.SetHexColor(fg)
			dc.SetLineWidth(borderT)
			dc.DrawRoundedRectangle(bX, bY, bW, bH, bR)
			
			dashes := parseLineStyle(el.Border.Style, borderT)
			if len(dashes) > 0 {
				dc.SetDash(dashes...)
			}
			
			dc.Stroke()
			
			if len(dashes) > 0 {
				dc.SetDash() // Reset dash for subsequent drawing
			}
		}
	}

	// Render components texts on top of background
	if hasLabel {
		lblT := getThreshold(rc.data.Values, el.Label.Thresholds)
		lblFg := rc.parseColor(lblT.Fg)
		if lblFg == "" {
			lblFg = rc.parseColor(el.Label.Font.Fg)
		}
		if lblFg == "" {
			lblFg = "#000000"
		}

		rc.loadFont(el.Label.Font.Face)
		dc.SetHexColor(lblFg)
		dc.DrawString(label, math.Round(labelX), math.Round(labelY))
	}

	if hasVal {
		valT := getThreshold(rc.data.Values, el.Value.Thresholds)
		valFg := rc.parseColor(valT.Fg)
		if valFg == "" {
			valFg = rc.parseColor(el.Value.Font.Fg)
		}
		if valFg == "" {
			valFg = "#000000"
		}

		rc.loadFont(el.Value.Font.Face)
		dc.SetHexColor(valFg)
		dc.DrawString(valStr, math.Round(valX), math.Round(valY))
	}

	if hasUnit {
		uT := getThreshold(rc.data.Values, el.Unit.Thresholds)
		unitFg := rc.parseColor(uT.Fg)
		if unitFg == "" {
			unitFg = rc.parseColor(el.Unit.Font.Fg)
		}
		if unitFg == "" {
			unitFg = "#000000"
		}

		rc.loadFont(el.Unit.Font.Face)
		dc.SetHexColor(unitFg)
		dc.DrawString(unit, math.Round(valX+unitX), math.Round(valY+unitY))
	}

}
