# CaddyShack Architecture

## Overview

CaddyShack is a single-binary Go web application with an embedded static frontend. It follows a stateless request/response model: the user uploads a Caddy JSONL log file, the backend parses and aggregates it in a single streaming pass, and returns a JSON report that the frontend renders as a dashboard.

```
┌──────────────────────────────────────────────────┐
│  Browser (Vanilla HTML/JS/CSS)                   │
│                                                  │
│  ┌──────────┐  ┌────────┐  ┌────────┐  ┌──────┐│
│  │ app.js   │  │charts.js│  │ map.js │  │D3.js ││
│  │ upload & │  │Canvas 2D│  │D3 geo  │  │local ││
│  │ render   │  │bar chart│  │bubbles │  │vendor││
│  └────┬─────┘  └────────┘  └────────┘  └──────┘│
│       │                                          │
│       │ POST /api/upload (multipart/form-data)   │
└───────┼──────────────────────────────────────────┘
        │
        ▼
┌──────────────────────────────────────────────────┐
│  Go Backend (net/http)                           │
│                                                  │
│  main.go                                         │
│    ├─ handler.Upload                             │
│    │    └─ analyzer.Analyze(io.Reader)           │
│    │         ├─ logparser.ParseStream (line-by-  │
│    │         │    line JSONL streaming)           │
│    │         ├─ useragent.Parse (browser/OS)     │
│    │         ├─ geoip.Lookup (country from IP)   │
│    │         ├─ anonymize.IP (GDPR compliance)   │
│    │         └─ map counters → Report struct     │
│    │                                             │
│    ├─ handler.Health                             │
│    └─ http.FileServer (static/)                  │
└──────────────────────────────────────────────────┘
```

## Design Principles

1. **Stateless** — no database, no sessions, no stored state. Each upload produces a fresh analysis. This keeps deployment trivial (single binary + static files).

2. **Streaming** — log files are parsed line-by-line via `bufio.Scanner`. Memory usage is proportional to the number of unique values (IPs, URIs, user agents), not the total number of log lines.

3. **No frameworks** — the Go backend uses only `net/http` from the standard library. The frontend uses vanilla HTML, CSS, and JavaScript. D3.js is the only external dependency and is served locally.

4. **Privacy by default** — IPs are anonymized before they appear in the API response. GeoIP lookups happen on the original IP during analysis, but only the anonymized IP reaches the client.

## Package Dependency Graph

```
main
 ├─ handler
 │   └─ analyzer
 │       ├─ logparser
 │       ├─ useragent
 │       ├─ anonymize
 │       └─ geoip
 └─ geoip (Load on startup)
```

No circular dependencies. Each internal package has a single responsibility.

## Data Flow

1. User drops/selects a `.jsonl` file in the browser
2. `app.js` creates a `FormData` with the file and POSTs to `/api/upload`
3. `handler.Upload` wraps the body in `MaxBytesReader` (500 MB limit), extracts the file from multipart form
4. `analyzer.Analyze` receives an `io.Reader` and streams through entries:
   - `logparser.ParseStream` reads one JSON line at a time
   - For each entry: extract UA → `useragent.Parse`, look up IP → `geoip.Lookup`, anonymize → `anonymize.IP`
   - Increment map-based counters for all dimensions
5. After the stream completes, counters are sorted/trimmed into the `Report` struct
6. `Report` is serialized as JSON and returned to the client
7. `app.js` distributes the data to `charts.js` (Canvas bar charts) and `map.js` (D3 bubble map)

## File Serving

Static files are served from the `static/` directory via `http.FileServer`. The URL space:

| Path | Source |
|------|--------|
| `/` | `static/index.html` |
| `/css/*` | `static/css/` |
| `/js/*` | `static/js/` |
| `/img/*` | `static/img/` |
| `/vendor/*` | `static/vendor/` (D3.js, topojson-client) |
| `/data/*` | `static/data/` (countries-110m.json) |
| `/api/upload` | `handler.Upload` |
| `/api/health` | `handler.Health` |
