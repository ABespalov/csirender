// Copyright (c) 2026, Anton Bespalov. Licensed under the MIT License.
package csirender

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// parseAlignFraction parses layout ratio fractions like "1/2" or static floats.
func parseAlignFraction(s string) (float64, bool) {
	if s == "" {
		return 0, false
	}
	parts := strings.Split(s, "/")
	if len(parts) == 2 {
		num, err1 := strconv.ParseFloat(parts[0], 64)
		den, err2 := strconv.ParseFloat(parts[1], 64)
		if err1 == nil && err2 == nil && den != 0 {
			return num / den, true
		}
	} else if len(parts) == 1 {
		num, err := strconv.ParseFloat(parts[0], 64)
		if err == nil {
			return num, true
		}
	}
	return 0, false
}

// parseDashStyle decodes style commands like "dashed.2.14" to dash/gap widths.
// Kept for backward compatibility if needed, but parseLineStyle is the modern equivalent.
func parseDashStyle(style string) (dashW, gapW float64) {
	dashW = 1.0
	gapW = 6.0
	if style == "" {
		return
	}
	parts := strings.Split(style, ".")
	if len(parts) >= 3 && parts[0] == "dashed" {
		if dw, err := strconv.ParseFloat(parts[1], 64); err == nil {
			dashW = dw
		}
		if gw, err := strconv.ParseFloat(parts[2], 64); err == nil {
			gapW = gw
		}
	}
	return
}

// parseLineStyle parses a line style string into a dash pattern array for gg.Context.
func parseLineStyle(style string, thickness float64) []float64 {
	if thickness <= 0 {
		thickness = 1.0
	}
	
	style = strings.ToLower(strings.TrimSpace(style))
	if style == "" || style == "solid" {
		return nil
	}
	if style == "dashed" {
		return []float64{thickness * 4, thickness * 4}
	}
	if style == "dotted" {
		return []float64{thickness, thickness * 2}
	}
	if style == "dashdot" {
		return []float64{thickness * 4, thickness * 2, thickness, thickness * 2}
	}
	
	// Custom patterns: dashed.10.5 or custom.10.5.2.5
	parts := strings.Split(style, ".")
	if len(parts) > 1 {
		var dashes []float64
		for _, p := range parts[1:] {
			if v, err := strconv.ParseFloat(p, 64); err == nil && v > 0 {
				dashes = append(dashes, v)
			}
		}
		if len(dashes) > 0 {
			// gg.SetDash requires an even number of dash values or handles odd properly,
			// but we just pass what the user provides.
			return dashes
		}
	}

	return nil
}

// cleanBase trims property suffixes to get the core variable name.
func cleanBase(s string) string {
	for _, suffix := range []string{".value", ".zone", ".label"} {
		if idx := strings.Index(s, suffix); idx != -1 {
			return s[:idx]
		}
	}
	return s
}

// resolvePlaceholders evaluates text template variables like "{temp.value.to_celsius%+.1f}".
func (rc *RenderContext) resolvePlaceholders(template string) string {
	now := time.Now()
	res := template

	re := regexp.MustCompile(`\{([^{}]+)\}`)
	matches := re.FindAllStringSubmatch(res, -1)
	for _, match := range matches {
		fullMatch := match[0]
		content := match[1]

		if strings.HasPrefix(content, "date:") {
			layout := strings.TrimPrefix(content, "date:")
			res = strings.ReplaceAll(res, fullMatch, now.Local().Format(layout))
			continue
		}
		if strings.HasPrefix(content, "time:") {
			layout := strings.TrimPrefix(content, "time:")
			res = strings.ReplaceAll(res, fullMatch, now.Local().Format(layout))
			continue
		}

		varName := content
		format := "%v"
		if idx := strings.Index(content, "%"); idx != -1 {
			varName = strings.TrimSpace(content[:idx])
			format = strings.TrimSpace(content[idx:])
		}

		if val, ok := rc.data.Values[varName]; ok {
			formatted := fmt.Sprintf(format, val)
			res = strings.ReplaceAll(res, fullMatch, formatted)
		}
	}
	return res
}

func getFloatVal(m map[string]interface{}, key string, def float64) float64 {
	if val, ok := m[key]; ok {
		switch v := val.(type) {
		case float64:
			return v
		case int:
			return float64(v)
		}
	}
	return def
}

func getStringVal(m map[string]interface{}, key string, def string) string {
	if val, ok := m[key]; ok {
		if s, ok := val.(string); ok {
			return s
		}
	}
	return def
}

// getThreshold resolves matching threshold configuration
func getThreshold(values map[string]interface{}, thresholds []Threshold) Threshold {
	for _, t := range thresholds {
		if evalCondition(t.Condition, values) {
			return t
		}
	}
	return Threshold{}
}

// evalCondition evaluates if condition matches current values
func evalCondition(cond string, values map[string]interface{}) bool {
	cond = strings.TrimSpace(cond)
	if cond == "" {
		return true
	}

	parts := strings.Split(cond, " ")
	if len(parts) != 3 {
		return false
	}

	varName := parts[0]
	op := parts[1]
	rightStr := parts[2]

	var leftVal float64
	val, ok := values[varName]
	if !ok {
		return false
	}

	switch v := val.(type) {
	case float64:
		leftVal = v
	case int:
		leftVal = float64(v)
	case string:
		if op == "==" {
			return v == strings.Trim(rightStr, "\"'")
		}
		if op == "!=" {
			return v != strings.Trim(rightStr, "\"'")
		}
		return false
	default:
		return false
	}

	rightVal, err := strconv.ParseFloat(rightStr, 64)
	if err != nil {
		return false
	}

	switch op {
	case "==":
		return leftVal == rightVal
	case "!=":
		return leftVal != rightVal
	case ">":
		return leftVal > rightVal
	case ">=":
		return leftVal >= rightVal
	case "<":
		return leftVal < rightVal
	case "<=":
		return leftVal <= rightVal
	}

	return false
}
