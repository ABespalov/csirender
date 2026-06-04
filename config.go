// Copyright (c) 2026, Anton Bespalov
//
// Package csirender provides an engine for declarative rendering of visual dashboards.
package csirender

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

// Parser handles loading and caching of generic configuration structures.
// It allows business applications to define their own config struct that embeds LayoutConfig,
// parse it efficiently, and cache the parsed result to avoid expensive re-parsing.
type Parser[T any] struct {
	mu    sync.RWMutex
	cache map[string]cachedConfig[T]
}

type cachedConfig[T any] struct {
	modTime time.Time
	config  T
}

// NewParser creates a new caching configuration parser.
func NewParser[T any]() *Parser[T] {
	return &Parser[T]{
		cache: make(map[string]cachedConfig[T]),
	}
}

// Parse loads a configuration file. It uses the file extension to determine
// whether to use JSON or YAML parser. It caches the parsed object based on
// the file's modification time.
// Note: The returned configuration is a cached instance and should be treated as read-only.
func (p *Parser[T]) Parse(path string) (T, error) {
	var zero T

	info, err := os.Stat(path)
	if err != nil {
		return zero, fmt.Errorf("stat config file: %w", err)
	}

	p.mu.RLock()
	cached, ok := p.cache[path]
	p.mu.RUnlock()

	if ok && cached.modTime.Equal(info.ModTime()) {
		return cached.config, nil
	}

	bytes, err := os.ReadFile(path)
	if err != nil {
		return zero, fmt.Errorf("read config file: %w", err)
	}

	var cfg T
	ext := strings.ToLower(filepath.Ext(path))
	if ext == ".json" {
		if err := json.Unmarshal(bytes, &cfg); err != nil {
			return zero, fmt.Errorf("parse JSON: %w", err)
		}
	} else {
		if err := yaml.Unmarshal(bytes, &cfg); err != nil {
			return zero, fmt.Errorf("parse YAML: %w", err)
		}
	}

	p.mu.Lock()
	p.cache[path] = cachedConfig[T]{
		modTime: info.ModTime(),
		config:  cfg,
	}
	p.mu.Unlock()

	return cfg, nil
}
