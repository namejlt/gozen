package config

import (
	"os"
	"path/filepath"
	"time"
)

// FileSource reads YAML / JSON configuration files from a local directory.
// Each file is named "<config-name>.<ext>" (e.g. "app.yaml", "db.json").
type FileSource struct {
	dir       string
	ext       string   // "yaml" or "json"
	watchCh   chan string
	modTimes  map[string]time.Time
	interval  time.Duration
}

// FileSourceOption is a functional option for FileSource.
type FileSourceOption func(*FileSource)

// WithExt overrides the file extension (default "yaml").
func WithExt(ext string) FileSourceOption {
	return func(fs *FileSource) { fs.ext = ext }
}

// WithWatchInterval sets the poll interval for file changes (default 5m).
func WithWatchInterval(d time.Duration) FileSourceOption {
	return func(fs *FileSource) { fs.interval = d }
}

// NewFileSource creates a FileSource backed by the given directory.
func NewFileSource(dir string, opts ...FileSourceOption) *FileSource {
	fs := &FileSource{
		dir:      dir,
		ext:      "yaml",
		modTimes: make(map[string]time.Time),
		interval: 5 * time.Minute,
	}
	for _, o := range opts {
		o(fs)
	}
	return fs
}

// Read implements Source.
func (fs *FileSource) Read(name string) ([]byte, error) {
	path := filepath.Join(fs.dir, name+"."+fs.ext)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if info, err := os.Stat(path); err == nil {
		fs.modTimes[name] = info.ModTime()
	}
	return data, nil
}

// Format implements Source.
func (fs *FileSource) Format() string {
	if fs.ext == "json" {
		return "json"
	}
	return "yaml"
}

// Watch implements Source.  It polls the directory every interval.
func (fs *FileSource) Watch() <-chan string {
	if fs.watchCh != nil {
		return fs.watchCh
	}
	fs.watchCh = make(chan string, 8)
	go func() {
		t := time.NewTicker(fs.interval)
		defer t.Stop()
		for range t.C {
			entries, err := os.ReadDir(fs.dir)
			if err != nil {
				continue
			}
			for _, e := range entries {
				if e.IsDir() {
					continue
				}
				ext := filepath.Ext(e.Name())
				if ext != "."+fs.ext {
					continue
				}
				name := e.Name()[:len(e.Name())-len(ext)]
				info, err := e.Info()
				if err != nil {
					continue
				}
				prev, ok := fs.modTimes[name]
				if !ok || info.ModTime().After(prev) {
					// Re-read to update modTime and signal change.
					if _, err := fs.Read(name); err == nil {
						fs.watchCh <- name
					}
				}
			}
		}
	}()
	return fs.watchCh
}
