// Copyright (c) 2026, Anton Bespalov. Licensed under the MIT License.
package csirender

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/fogleman/gg"
)

func renderChart(dc *gg.Context, rc *RenderContext, el *ChartElement) {
	// buildLocalTel creates a per-bar/per-label telemetry snapshot overriding the
	// chart source key with the given value, and injecting any domain-derived keys
	// (e.g. zone.index) via the optional TelemetryEnricher callback.
	buildLocalTel := func(v float64) map[string]interface{} {
		local := make(map[string]interface{}, len(rc.data.Values)+8)
		for k, val := range rc.data.Values {
			local[k] = val
		}
		local[el.Source] = v
		local[cleanBase(el.Source)] = v
		if rc.Enricher != nil {
			for k, val := range rc.Enricher(el.Source, v) {
				local[k] = val
			}
		}
		return local
	}

	numPoints := el.Points
	if numPoints == 0 {
		numPoints = 48
	}

	var data []float64
	if d, ok := rc.data.Charts[el.Source]; ok {
		data = d
	} else {
		data = make([]float64, numPoints)
	}

	// Determine scale options
	scaleAuto := false
	scaleMin := "0"
	scaleMax := "auto"

	if yAxisMap, ok := el.Axis["y"].(map[string]interface{}); ok {
		// Backwards compatibility for auto_scale
		if autosc, ok := yAxisMap["auto_scale"]; ok {
			if b, ok := autosc.(bool); ok {
				scaleAuto = b
				if b {
					scaleMin = "0"
					scaleMax = "auto"
				} else {
					scaleMin = "0"
					scaleMax = "150"
				}
			}
		}

		// Support new scale hierarchy
		if scaleMap, ok := yAxisMap["scale"].(map[string]interface{}); ok {
			if autoVal, ok := scaleMap["auto"].(bool); ok {
				scaleAuto = autoVal
			}
			if minVal, ok := scaleMap["min"]; ok {
				scaleMin = fmt.Sprintf("%v", minVal)
			}
			if maxVal, ok := scaleMap["max"]; ok {
				scaleMax = fmt.Sprintf("%v", maxVal)
			}
		}
	}

	// Calculate data bounds
	dataMin := 0.0
	dataMax := 0.0
	if len(data) > 0 {
		dataMin = data[0]
		dataMax = data[0]
		for _, v := range data {
			if v < dataMin {
				dataMin = v
			}
			if v > dataMax {
				dataMax = v
			}
		}
	}

	yMin := 0.0
	if strings.ToLower(scaleMin) == "auto" {
		if scaleAuto {
			yMin = dataMin - math.Abs(dataMin)*0.10
		}
	} else {
		if val, err := strconv.ParseFloat(scaleMin, 64); err == nil {
			yMin = val
		}
	}

	yMax := 150.0
	if strings.ToLower(scaleMax) == "auto" {
		if scaleAuto {
			yMax = dataMax + math.Abs(dataMax)*0.10
			if yMax < yMin+50.0 {
				yMax = yMin + 50.0
			}
		}
	} else {
		if val, err := strconv.ParseFloat(scaleMax, 64); err == nil {
			yMax = val
		}
	}

	// Safety feature to prevent clipping
	if dataMax > yMax {
		yMax = dataMax + math.Abs(dataMax)*0.10
	}

	yMin = math.Floor(yMin)
	yMax = math.Ceil(yMax)
	yRange := yMax - yMin
	if yRange <= 0 {
		yRange = 1.0
	}

	// Draw chart columns
	barW := el.Size.W / float64(numPoints)
	for i, v := range data {
		barH := ((v - yMin) / yRange) * el.Size.H
		if barH < 0 {
			barH = 0
		}
		bx1 := el.Position.X + float64(i)*barW
		bx2 := el.Position.X + float64(i+1)*barW
		by := el.Position.Y + el.Size.H - barH

		// Evaluate bar color using per-bar telemetry (with enriched zone.index).
		// Falls back to global threshold only when no per-bar thresholds are defined.
		c := ""
		if len(el.Thresholds) > 0 {
			localTel := buildLocalTel(v)
			for _, th := range el.Thresholds {
				if evalCondition(th.Condition, localTel) {
					c = rc.parseColor(th.Bg)
					if c == "" {
						c = rc.parseColor(th.Color)
					}
					break
				}
			}
		} else {
			// No per-bar thresholds: use global threshold as a uniform chart color
			t := getThreshold(rc.data.Values, el.Thresholds)
			c = rc.parseColor(t.Bg)
			if c == "" {
				c = rc.parseColor(t.Color)
			}
		}
		if c == "" {
			c = "#000000"
		}

		dc.SetHexColor(c)
		dc.DrawRectangle(bx1, by, bx2-bx1, barH)
		dc.Fill()
	}

	// Draw horizontal threshold guidelines
	for _, tl := range el.ThresholdLines {
		tlValue := tl.Value
		if tlValue == 0 && rc.Resolver != nil {
			if val, ok := rc.Resolver(tl.Condition, el.Source); ok {
				tlValue = val
			}
		}

		if tlValue == 0 {
			continue
		}

		ly := math.Round(el.Position.Y + el.Size.H - ((tlValue-yMin)/yRange)*el.Size.H)
		tlColorHex := rc.parseColor(tl.Color)
		if tlColorHex == "" {
			tlColorHex = "#000000"
		}

		lineWidth := 2.0
		if tl.T != nil {
			lineWidth = *tl.T
		}

		dashW, gapW := parseDashStyle(tl.Style)
		totalPeriod := dashW + gapW

		for cx := el.Position.X; cx < el.Position.X+el.Size.W; cx += totalPeriod {
			cw := dashW
			if cx+cw > el.Position.X+el.Size.W {
				cw = (el.Position.X + el.Size.W) - cx
			}
			if cw <= 0 {
				break
			}

			barIndex := int((cx - el.Position.X) / barW)
			if barIndex < 0 {
				barIndex = 0
			} else if barIndex >= len(data) {
				barIndex = len(data) - 1
			}
			v := data[barIndex]
			barH := ((v - yMin) / yRange) * el.Size.H
			if barH < 0 {
				barH = 0
			}
			by := el.Position.Y + el.Size.H - barH

			dotColor := tlColorHex
			if ly >= by {
				barColorHex := ""
				for _, th := range el.Thresholds {
					if evalCondition(th.Condition, buildLocalTel(v)) {
						barColorHex = rc.parseColor(th.Bg)
						break
					}
				}
				if barColorHex == "" {
					barColorHex = "#000000"
				}
				if dotColor == barColorHex {
					dotColor = "#FFFFFF"
				}
			}

			dc.SetHexColor(dotColor)
			dc.DrawRectangle(cx, ly-lineWidth/2, cw, lineWidth)
			dc.Fill()
		}
	}

	// Render Axis & Notches
	showYAxis := true
	showYLabels := true
	yStep := 0.0
	var zoneStyles []map[string]string
	yAxisFont := "tiny"
	yAxisFontFg := "black"
	yAxisW := 2.0
	yAxisFg := "black"
	yNotchesVisible := true
	yNotchesFg := "black"
	yNotchesT := 2.0 // thickness
	yNotchesL := 6.0 // length
	yLabelsCount := 0
	yLabelsFormat := "%.0f"

	if yAxisMap, ok := el.Axis["y"].(map[string]interface{}); ok {
		if show, ok := yAxisMap["show"].(bool); ok {
			showYAxis = show
		}
		yStep = getFloatVal(yAxisMap, "min_step", 0.0)
		yAxisW = getFloatVal(yAxisMap, "t", getFloatVal(yAxisMap, "w", yAxisW))
		yAxisFg = getStringVal(yAxisMap, "fg", yAxisFg)

		if labelsMap, ok := yAxisMap["labels"].(map[string]interface{}); ok {
			if show, ok := labelsMap["show"].(bool); ok {
				showYLabels = show
			}
			if cVal, ok := labelsMap["count"].(int); ok {
				yLabelsCount = cVal
			}
			yLabelsFormat = getStringVal(labelsMap, "format", yLabelsFormat)

			if fVal, ok := labelsMap["font"]; ok {
				if face, ok := fVal.(string); ok && face != "" {
					yAxisFont = face
				} else if fMap, ok := fVal.(map[string]interface{}); ok {
					yAxisFont = getStringVal(fMap, "face", yAxisFont)
					yAxisFontFg = getStringVal(fMap, "fg", yAxisFontFg)
				}
			}

			if thsRaw, ok := labelsMap["thresholds"].([]interface{}); ok {
				for _, th := range thsRaw {
					if tmap, ok := th.(map[string]interface{}); ok {
						zstyle := make(map[string]string)
						if cond, ok := tmap["condition"].(string); ok {
							zstyle["condition"] = cond
						}
						if bg, ok := tmap["bg"].(string); ok {
							zstyle["bg"] = bg
						}
						if fg, ok := tmap["fg"].(string); ok {
							zstyle["fg"] = fg
						}
						zoneStyles = append(zoneStyles, zstyle)
					}
				}
			}
		}

		if notchesMap, ok := yAxisMap["notches"].(map[string]interface{}); ok {
			if show, ok := notchesMap["show"].(bool); ok {
				yNotchesVisible = show
			}
			yNotchesFg = getStringVal(notchesMap, "fg", yNotchesFg)
			yNotchesT = getFloatVal(notchesMap, "t", getFloatVal(notchesMap, "h", yNotchesT))
			yNotchesL = getFloatVal(notchesMap, "l", getFloatVal(notchesMap, "w", yNotchesL))
		}
	}

	showXAxis := true
	showXLabels := true
	xAxisW := 2.0
	xAxisFg := "black"
	xAxisFont := "tiny"
	xAxisFontFg := "black"
	notchesVisible := true
	notchesFg := "black"
	notchesT := 2.0 // thickness
	notchesL := 6.0 // length
	labelsCount := 5
	labelsFormat := "15:04"

	if xAxisMap, ok := el.Axis["x"].(map[string]interface{}); ok {
		if show, ok := xAxisMap["show"].(bool); ok {
			showXAxis = show
		}
		xAxisW = getFloatVal(xAxisMap, "t", getFloatVal(xAxisMap, "h", xAxisW))
		xAxisFg = getStringVal(xAxisMap, "fg", xAxisFg)

		if labelsMap, ok := xAxisMap["labels"].(map[string]interface{}); ok {
			if show, ok := labelsMap["show"].(bool); ok {
				showXLabels = show
			}
			if cVal, ok := labelsMap["count"].(int); ok {
				labelsCount = cVal
			}
			labelsFormat = getStringVal(labelsMap, "format", labelsFormat)
			if fVal, ok := labelsMap["font"]; ok {
				if face, ok := fVal.(string); ok && face != "" {
					xAxisFont = face
				} else if fMap, ok := fVal.(map[string]interface{}); ok {
					xAxisFont = getStringVal(fMap, "face", xAxisFont)
					xAxisFontFg = getStringVal(fMap, "fg", xAxisFontFg)
				}
			}
		}
		if notchesMap, ok := xAxisMap["notches"].(map[string]interface{}); ok {
			if show, ok := notchesMap["show"].(bool); ok {
				notchesVisible = show
			}
			notchesFg = getStringVal(notchesMap, "fg", notchesFg)
			notchesT = getFloatVal(notchesMap, "t", getFloatVal(notchesMap, "w", notchesT))
			notchesL = getFloatVal(notchesMap, "l", getFloatVal(notchesMap, "h", notchesL))
		}
	}

	if showXAxis {
		dc.SetLineWidth(xAxisW)
		dc.SetHexColor(rc.parseColor(xAxisFg))
		dc.DrawLine(el.Position.X, el.Position.Y+el.Size.H, el.Position.X+el.Size.W, el.Position.Y+el.Size.H)
		dc.Stroke()
	}

	if showYAxis {
		dc.SetLineWidth(yAxisW)
		dc.SetHexColor(rc.parseColor(yAxisFg))
		dc.DrawLine(el.Position.X, el.Position.Y, el.Position.X, el.Position.Y+el.Size.H)
		dc.Stroke()
	}

	if el.Axis != nil && el.Axis["y"] != nil {
		yFace := rc.getFontFace(yAxisFont)
		var yVals []float64
		if yStep > 0 {
			// min_step is used as the exact label step; the config author controls spacing.
			for val := yMin; val <= yMax; val += yStep {
				yVals = append(yVals, val)
			}
		} else if yLabelsCount > 1 {
			for i := 0; i < yLabelsCount; i++ {
				yVals = append(yVals, yMin+float64(i)*(yRange/float64(yLabelsCount-1)))
			}
		} else {
			for val := yMin; val <= yMax; val += 50.0 {
				yVals = append(yVals, val)
			}
		}

		for _, val := range yVals {
			ly := el.Position.Y + el.Size.H - ((val-yMin)/yRange)*el.Size.H
			bg := ""
			fg := yAxisFontFg
			localTel := buildLocalTel(val)

			for _, zs := range zoneStyles {
				if evalCondition(zs["condition"], localTel) {
					bg = zs["bg"]
					fg = zs["fg"]
					break
				}
			}

			if yNotchesVisible {
				dc.SetLineWidth(yNotchesT)
				dc.SetHexColor(rc.parseColor(yNotchesFg))
				dc.DrawLine(el.Position.X-yNotchesL, ly, el.Position.X, ly)
				dc.Stroke()
			}

			if showYLabels {
				textStr := fmt.Sprintf(yLabelsFormat, val)
				bxMin, byMin, bxMax, byMax := getStringBounds(yFace, textStr)
				w := bxMax - bxMin
				h := byMax - byMin

				pad := 2.0 * yAxisW
				radius := 2.0 * yAxisW
				gapY := 2 * yNotchesT // gap between notch end and label
				valX := el.Position.X - yNotchesL - gapY - bxMax
				valY := ly - (byMin+byMax)/2

				if bg != "" && bg != "transparent" {
					dc.SetHexColor(rc.parseColor(bg))
					dc.DrawRoundedRectangle(el.Position.X-yNotchesL-gapY-w-pad, ly-h/2-pad, w+pad*2, h+pad*2, radius)
					dc.Fill()
				}

				dc.SetHexColor(rc.parseColor(fg))
				rc.loadFont(yAxisFont)
				dc.DrawString(textStr, math.Round(valX), math.Round(valY))
			}
		}
	}

	if el.Axis != nil && el.Axis["x"] != nil {
		labels := make([]string, labelsCount)
		duration := 24 * time.Hour
		if el.Duration != "" {
			if d, err := time.ParseDuration(el.Duration); err == nil {
				duration = d
			}
		}

		endTime := time.Now()
		startTime := endTime.Add(-duration)
		labelInterval := duration / time.Duration(labelsCount-1)

		for i := 0; i < labelsCount; i++ {
			t := startTime.Add(labelInterval * time.Duration(i))
			labels[i] = t.Local().Format(labelsFormat)
		}

		step := el.Size.W / float64(len(labels)-1)
		for i, lbl := range labels {
			tickX := el.Position.X + float64(i)*step
			if notchesVisible {
				dc.SetLineWidth(notchesT)
				dc.SetHexColor(rc.parseColor(notchesFg))
				dc.DrawLine(tickX, el.Position.Y+el.Size.H, tickX, el.Position.Y+el.Size.H+notchesL)
				dc.Stroke()
			}
			if showXLabels {
				rc.loadFont(xAxisFont)
				dc.SetHexColor(rc.parseColor(xAxisFontFg))
				w, h := dc.MeasureString(lbl)
				lx := tickX - 0.5*w
				gapX := 2 * notchesT // gap between notch end and label
				ly := el.Position.Y + el.Size.H + notchesL + gapX + (1.0-0.5)*h
				dc.DrawString(lbl, math.Round(lx), math.Round(ly))
			}
		}
	}
}
