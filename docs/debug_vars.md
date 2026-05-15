# Debug Variables

The HTTP server exposes runtime debugging information at `/debug/vars`
(provided by `expvar.Handler()`).

## Endpoints

| Path | Description |
|------|-------------|
| `/debug/vars` | JSON-encoded runtime metrics |

## Metric Variables

- `cmdline` — Command-line arguments
- `memstats` — Go runtime memory statistics
- `memstats.BySize` — Memory allocation by size class
- `memstats.GCSys` — GC metadata size

## Access

```bash
curl http://localhost:8080/debug/vars | jq .
```

The prefix can be configured via `app.yaml`:

```yaml
debug_vars_prefix: "admin"
# Now accessible at /admin/debug/vars
```
