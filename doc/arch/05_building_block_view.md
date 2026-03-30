# 5. Building Block View

## Level 1 — System Decomposition

```mermaid
graph TB
    subgraph Browser["Browser — Vanilla HTML/JS/CSS"]
        appjs["app.js\nupload & render"]
        chartsjs["charts.js\nCanvas 2D bar charts"]
        mapjs["map.js\nD3 geo bubbles"]
    end

    subgraph Backend["Go Backend — net/http"]
        handler["handler\nUpload · Health · Logs · AnalyzeLocal"]
        analyzer["analyzer.Analyze"]
        fileserver["http.FileServer\nstatic/ (embedded)"]
    end

    appjs -->|"POST /api/upload\nmultipart/form-data"| handler
    appjs -->|"GET /api/logs\nGET /api/analyze-local"| handler
    appjs -->|"GET /css, /js, /data …"| fileserver
    handler --> analyzer
```

## Level 2 — Go Package Decomposition

### Package Dependency Graph

```mermaid
graph TD
    main --> handler["internal/handler"]
    main -->|"Load at startup"| geoip_s["internal/geoip"]
    handler --> analyzer["internal/analyzer"]
    analyzer --> logparser["internal/logparser"]
    analyzer --> useragent["internal/useragent"]
    analyzer --> anonymize["internal/anonymize"]
    analyzer --> geoip["internal/geoip"]
```

No circular dependencies. Each package has a single responsibility.

### Package Responsibilities

| Package | Responsibility |
|---------|---------------|
| `main` | Entry point: parse CLI flags, load GeoIP database, register routes, start HTTP server |
| `internal/handler` | HTTP request/response boundary: parse multipart upload, enforce size limits, JSON-encode responses |
| `internal/analyzer` | Core aggregation engine: orchestrate a single-pass stream, build and trim all counter maps, produce `MultiHostReport` |
| `internal/logparser` | JSONL deserialization: read line-by-line with `bufio.Scanner`, decode JSON, expose `LogEntry` structs |
| `internal/useragent` | User-Agent string parsing: ordered string matching to detect browser and OS names |
| `internal/anonymize` | IP anonymization: zero last IPv4 octet; truncate IPv6 to first 3 groups |
| `internal/geoip` | GeoIP lookup: load DB-IP Lite CSV into sorted uint32 slices; binary-search lookup; country name resolution |

### Frontend Modules

| Module | Responsibility |
|--------|---------------|
| `app.js` | Main orchestrator: file upload, API call, DOM population, host/filter state |
| `charts.js` (`Charts` namespace) | Canvas 2D horizontal and vertical bar charts with DPR scaling |
| `map.js` (`WorldMap` namespace) | D3.js Natural Earth bubble map with proportional sizing and hover tooltips |

### Key Data Types

| Type | Owner | Description |
|------|-------|-------------|
| `LogEntry` | logparser | Deserialized Caddy log line |
| `MultiHostReport` | analyzer | Root response: hosts list + per-host FullReports |
| `FullReport` | analyzer | Three `Report` objects: All, Success, Error |
| `Report` | analyzer | Aggregated metrics for one host + one status filter |
| `NameCount` | analyzer | Generic `{name, count}` tuple |
| `DayCount` | analyzer | `{date, count}` for daily traffic |
| `VisitorInfo` | analyzer | `{ip, count, country, country_name}` for top visitors |
| `CountryCount` | analyzer | `{code, name, count}` for geographic breakdown |
