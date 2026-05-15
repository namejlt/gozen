// Package config provides a unified, pluggable configuration system.
//
// Design:
//
//	Manager  ── orchestrates one or more Source instances
//	Source   ── reads raw bytes (file, nacos, env, http, …)
//	Codec    ── decodes/encodes raw bytes (yaml, json, toml, …)
//
// Watcher / hot-reload is supported via Source.Watch() — a Source can push
// updates to the Manager, which then re-decodes and notifies subscribers.
//
// Typical bootstrap:
//
//	mgr := config.NewManager()
//	mgr.Add(config.NewFileSource("./configs", config.WithCodec(&config.YAMLCodec{})))
//	mgr.Load("app", &appCfg)
//	mgr.Load("db", &dbCfg)
//	go mgr.Watch(ctx)
package config

import (
	"sync"
)

// Manager is the central configuration orchestrator.  It loads named
// configuration blocks from one or more Sources.
type Manager struct {
	mu      sync.RWMutex
	sources []Source
	codecs  map[string]Codec // format name → codec
}

// NewManager creates an empty Manager.
func NewManager() *Manager {
	return &Manager{codecs: map[string]Codec{
		"yaml": &YAMLCodec{},
		"json": &JSONCodec{},
	}}
}

// AddSource registers a configuration source.
func (m *Manager) AddSource(s Source) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sources = append(m.sources, s)
}

// Load reads the named configuration block into v.
// It iterates sources in registration order — the first source that
// contains "name" wins.
func (m *Manager) Load(name string, v any) error {
	m.mu.RLock()
	srcs := make([]Source, len(m.sources))
	copy(srcs, m.sources)
	m.mu.RUnlock()

	for _, s := range srcs {
		data, err := s.Read(name)
		if err != nil {
			continue
		}
		c := m.codecFor(s.Format())
		if c == nil {
			continue
		}
		return c.Unmarshal(data, v)
	}
	return &ErrNotFound{Name: name}
}

// Watch starts watching all sources for changes.  When a change is
// detected the callback is invoked with the config name.
func (m *Manager) Watch(onChange func(name string)) {
	for _, s := range m.sources {
		go func(src Source) {
			ch := src.Watch()
			if ch == nil {
				return
			}
			for name := range ch {
				onChange(name)
			}
		}(s)
	}
}

func (m *Manager) codecFor(format string) Codec {
	if c, ok := m.codecs[format]; ok {
		return c
	}
	return nil
}

// ErrNotFound is returned when no source contains the requested config.
type ErrNotFound struct{ Name string }

func (e *ErrNotFound) Error() string { return "config not found: " + e.Name }
