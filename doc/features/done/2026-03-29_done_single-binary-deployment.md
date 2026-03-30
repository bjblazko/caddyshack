---
date: 2026-03-29
status: done
---

# Single-Binary Deployment

CaddyShack ships as a single compiled Go binary with all frontend assets embedded.

## Implementation

- Static files (HTML, CSS, JS, vendor libraries, geographic data) embedded at compile time via Go's `embed` package
- Served from the `static/` directory tree via `http.FileServer`
- No external runtime dependencies except the optional GeoIP CSV

## CLI Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-addr` | `:8080` | Listen address |
| `-geodb` | `./data/dbip-country-lite.csv` | Path to DB-IP Lite CSV |

## Health Check

`GET /api/health` returns `{"status":"ok"}`. Suitable for container readiness probes and uptime monitors.

## Go Version

Requires Go 1.22+ (uses method-based routing in `http.ServeMux`).
