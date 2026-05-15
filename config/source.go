package config

import (
	"encoding/json"

	"gopkg.in/yaml.v3"
)

// Source reads configuration blocks identified by name (e.g. "app", "db").
// A Source may be backed by local files, a remote config center (Nacos,
// Consul, etc.), environment variables, or any other key-value store.
type Source interface {
	// Read returns the raw bytes for the named configuration.
	Read(name string) ([]byte, error)

	// Format returns the encoding format ("yaml", "json", "toml", …).
	Format() string

	// Watch returns a channel that emits the names of changed configs.
	// Return nil if the source does not support hot-reload.
	Watch() <-chan string
}

// Codec encodes / decodes configuration bytes.
type Codec interface {
	Marshal(v any) ([]byte, error)
	Unmarshal(data []byte, v any) error
}

// — built-in codecs ---------------------------------------------------------

// YAMLCodec implements Codec for YAML.
type YAMLCodec struct{}

func (YAMLCodec) Marshal(v any) ([]byte, error)  { return yaml.Marshal(v) }
func (YAMLCodec) Unmarshal(d []byte, v any) error { return yaml.Unmarshal(d, v) }

// JSONCodec implements Codec for JSON.
type JSONCodec struct{}

func (JSONCodec) Marshal(v any) ([]byte, error)  { return json.Marshal(v) }
func (JSONCodec) Unmarshal(d []byte, v any) error { return json.Unmarshal(d, v) }
