# 8. Crosscutting Concepts

## Privacy and IP Anonymization

Applies to: `internal/anonymize`, `internal/geoip`, `internal/analyzer`

IP addresses are anonymized unconditionally before being stored in any counter or included in any response. There is no configuration option to disable this.

| IP version | Rule | Example |
|------------|------|---------|
| IPv4 | Zero the last octet | `93.184.216.34` → `93.184.216.0` |
| IPv6 | Keep first 3 groups, zero the rest | `2a01:4f8:c17::1` → `2a01:4f8:c17::` |

Processing order per log entry:
1. GeoIP lookup on the **original** IP → country code stored
2. IP anonymized → anonymized form stored in all subsequent counters
3. Raw IP discarded

See [ADR-005](adr/ADR-005_ip-anonymization-by-default.md).

---

## Streaming Memory Model

Applies to: `internal/logparser`, `internal/analyzer`

Log files are read line-by-line via `bufio.Scanner` (1 MB line buffer). At no point is the entire file held in memory. Memory consumption is proportional to the number of distinct values encountered (unique IPs, unique URIs, unique user-agents), not to the number of log lines.

The 500 MB `MaxBytesReader` limit prevents memory exhaustion from oversized uploads, but in practice even very large log files produce bounded counter maps.

---

## Error Handling

Applies to: `internal/logparser`, `internal/handler`

**Malformed log lines** — silently skipped by `logparser.ParseStream`. The scanner continues to the next line. This handles partial writes, truncated files, and non-JSON lines (e.g., log rotation markers) without crashing.

**Missing GeoIP database** — `geoip.Load` logs a warning and returns without populating the lookup table. All subsequent `geoip.Lookup` calls return `"??"`. The application continues normally.

**Upload errors** — `handler.Upload` returns `400 Bad Request` with a plain-text error message for: missing `logfile` field, invalid multipart form, file exceeding 500 MB.

---

## Offline-First Frontend

Applies to: `static/vendor/`, `static/data/`, `static/js/`

All runtime assets are served from the embedded binary:
- D3.js v7 and topojson-client v3 from `/vendor/`
- Natural Earth 110m country boundaries from `/data/countries-110m.json`

The dashboard makes zero external HTTP requests after the initial page load. This ensures it works in firewalled environments and avoids third-party data leakage.

---

## DPR-Aware Canvas Rendering

Applies to: `static/js/charts.js`

Canvas 2D elements are scaled by `window.devicePixelRatio` at render time and their CSS size set to the logical size. This produces crisp output on Retina / HiDPI displays without requiring any image assets or SVG.

---

## Single Responsibility per Package

Each Go package in `internal/` has exactly one job and no knowledge of HTTP or JSON. `handler` is the only layer that touches `net/http`. This keeps packages unit-testable in isolation and prevents business logic from leaking into the HTTP layer.

---

## Stateless Request Lifecycle

Every HTTP request is fully self-contained:
- No global mutable state is written during request handling
- The only shared state is the GeoIP lookup table (read-only after startup)
- All counters and maps created during `analyzer.Analyze` are local to that call and garbage-collected after the response is written

This means CaddyShack can be safely run behind a load balancer with multiple instances without any coordination.
