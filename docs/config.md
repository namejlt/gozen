# Configuration Guide

The `config/` package provides a unified, pluggable configuration system.

## Architecture

```
  ┌─────────────────┐     ┌──────────────────┐     ┌──────────────────┐
  │  FileSource     │     │  NacosSource     │     │  CustomSource    │
  │  (local YAML)   │     │  (remote centre) │     │  (env/etcd/…)    │
  └────────┬────────┘     └────────┬─────────┘     └────────┬─────────┘
           │                      │                          │
           └──────────────────────┼──────────────────────────┘
                                  ▼
                    ┌─────────────────────────┐
                    │     Manager             │
                    │  - AddSource(Source)    │
                    │  - Load(name, &v)       │
                    │  - Watch(callback)      │
                    └─────────────────────────┘
```

## Usage

```go
import "github.com/namejlt/gozen/config"

func main() {
    mgr := config.NewManager()

    // 1. Add source(s)
    mgr.AddSource(config.NewFileSource("./configs"))

    // 2. Load config into typed struct
    var appCfg config.AppConfig
    if err := mgr.Load("app", &appCfg); err != nil {
        panic(err)
    }
    fmt.Println(appCfg.Env) // "dev"

    // 3. Watch for hot-reload
    mgr.Watch(func(name string) {
        log.L().Infow("config changed", "name", name)
    })
}
```

## Source Types

### FileSource

Reads YAML/JSON files from a local directory. Each file is named
`<name>.yaml` (or `<name>.json`):

```
configs/
├── app.yaml
├── db.yaml
├── cache.yaml
└── tracer.yaml
```

Options:

```go
fs := config.NewFileSource("./configs",
    config.WithExt("yaml"),       // file extension (default yaml)
    config.WithWatchInterval(5*time.Minute),  // poll interval
)
```

### NacosSource

Reads configuration from a Nacos config centre and writes to a local
directory for caching:

```go
ns, _ := config.NewNacosSource(config.NacosConfig{
    Addr:        "127.0.0.1",
    Port:        8848,
    NamespaceID: "public",
    Group:       "DEFAULT_GROUP",
    DataIDs:     []string{"app", "db"},
}, "./configs")
ns.PullAll()
mgr.AddSource(ns)
```

## Built-in Config Types

All standard configuration structures are defined in `config/types.go`:

| Type | Description | YAML Example |
|------|-------------|-------------|
| `config.AppConfig` | Application-level settings | `env: dev`, `Host: ":8080"` |
| `config.DBConfig` | Database connections | mysql write/read, redis, mongo |
| `config.CacheConfig` | Redis cache | address, pool, cluster |
| `config.TracerConfig` | SkyWalking tracing | service name, reporter, sampler |
| `config.LogConfig` | Logger bootstrap | path, level, rotation |
| `config.ESConfig` | Elasticsearch | address, transport, timeout |
| `config.ProjectConfig` | Framework bootstrap | mode (local/nacos), format |

## Custom Config

For application-specific config, define your own struct and load it:

```go
type MyConfig struct {
    FeatureFlag bool   `yaml:"feature_flag"`
    MaxRetries  int    `yaml:"max_retries"`
    Endpoint    string `yaml:"endpoint"`
}

var myCfg MyConfig
mgr.Load("myapp", &myCfg)
```

## Config File Format

```yaml
# app.yaml
env: dev
host: ":8080"
docs: swagger
configs:
  sign_switch: "1"
  app_secret_key: "my-secret"

# db.yaml
mysql:
  db_name: myapp
  pool:
    pool_min_cap: 5
    pool_max_cap: 20
  write:
    address: 127.0.0.1
    port: 3306
    user: root
    password: root
```
