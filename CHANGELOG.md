# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- HTTP Method filter: dropdown in the filter bar lets users scope the entire dashboard to a specific HTTP verb (GET, POST, HEAD, etc.); options auto-populate from methods present in the log; filter applies to both `/api/analyze` and `/api/events`
- Single Events tab: browse raw log entries (most recent first) with lazy loading — 100 events per page, more load automatically on scroll via IntersectionObserver
- `GET /api/events` endpoint: same filter params as `/api/analyze`, returns paginated `EventsResult` with enriched `EventEntry` rows (timestamp ISO 8601, method, host, URI, status, size, duration, anonymized IP, country, browser, OS)
- Per-panel filter hint badges showing all active filters (site, HTTP status range, date, country, browser, OS, page); panels with no active filter display "all data"
- Backend filter-then-aggregate architecture (ADR-009): all filter dimensions (host, date range, country, browser, OS, page, HTTP status) applied with AND logic in a single streaming pass before aggregation; replaces client-side re-aggregation
- Uploaded log files saved to OS temp directory under a random hex ID; subsequent filter changes re-analyze the same file via `GET /api/analyze?file=<id>` without re-uploading
- Date range filter: native `<input type="date">` controls (start/end) with clear button; filters all dashboard panels via backend
- Dimension dropdown filters: Country, Browser, OS, Page — options populated from the current backend response, auto-reset when the selected value disappears under a new filter combination
- `FilterParams` struct carrying all filter dimensions passed to `analyzer.Analyze`
- `AnalysisResult` response type (`file_id`, `hosts`, `report`) replacing `MultiHostReport`/`FullReport`
- `GET /api/analyze` endpoint accepting all filter params as query-string arguments
- Glossary of project terms and concepts at `doc/glossary.md`
- arc42 architecture documentation in `doc/arch/` (12 sections)
- 9 Architecture Decision Records in `doc/arch/adr/`
- Current-state specs in `doc/specs/` grouped by concern: parsing, analysis, security, ui, api, deployment
- Feature tracking structure in `doc/features/` with todo/in-progress/done folders
- `CLAUDE.md` with project conventions for glossary use, feature tracking, spec maintenance, ADR consultation, and Mermaid-only diagrams

### Changed

- Replace Browsers and Operating Systems bar charts with D3.js donut charts; slices use eight distinguishable shades of green with a compact legend showing name and percentage; fixes layout overflow at medium viewport widths where canvas value labels broke out of their panels
- Dashboard split into tabbed layout: Statistics (all existing panels) and Single events
- 4xx/5xx event rows highlighted in red in the Single Events table
- `POST /api/upload` now saves the file to temp storage and returns `file_id` alongside the initial analysis result
- Host dropdown options reflect only hosts present under the current non-host filter combination
- HTTP status filter group label renamed from "Filter" to "HTTP Status Range"
- Safari browser detection tightened: requires `Version/X.X` token to exclude HTTP clients with partial WebKit UA strings
- `GET /api/analyze-local` replaced by `GET /api/analyze` (accepts both `file=<id>` and `name=<filename>`)
- Replaced all ASCII box-drawing diagrams with Mermaid diagrams across `doc/arch/` and `doc/specs/`
- Replaced personal domain in test data with `example.com`; renamed `testdata/sample-huepattl.jsonl` to `testdata/sample-example.jsonl`

## [0.1.1] - 2026-03-29

### Added

- Docker support: multi-stage `Dockerfile` and `compose.yml` for container-based deployment
- GHCR publishing: GitHub Actions release workflow now builds and pushes multi-arch images (`linux/amd64`, `linux/arm64`) to `ghcr.io/bjblazko/caddyshack`
- Screenshots in README showing the dashboard UI

## [0.1.0] - 2026-03-29

### Added

- Drag-and-drop upload of Caddy JSONL access log files
- Summary cards: total requests, unique IPs, data transferred, average response time
- World map with proportional bubbles showing geographic request distribution
- Browser and OS detection with horizontal bar charts
- Daily traffic trend chart
- HTTP status code breakdown (2xx, 3xx, 4xx, 5xx)
- Top pages listing (excluding static assets)
- Top visitors with anonymized IPs and country attribution
- Multi-host log support with per-host and aggregate views
- Success/error traffic segmentation (2xx vs 4xx+)
- GDPR-compliant IP anonymization (IPv4 last-octet zeroing, IPv6 prefix truncation)
- GeoIP country resolution via optional DB-IP Lite CSV database
- Server-side log file discovery at GET /api/logs
- Health check endpoint at GET /api/health
- Vanilla HTML/CSS/JavaScript frontend with no CDN dependencies
- D3.js and TopoJSON served locally for offline use
- Single-binary deployment with configurable listen address and GeoIP path

[Unreleased]: https://github.com/bjblazko/caddyshack/compare/v0.1.1...HEAD
[0.1.1]: https://github.com/bjblazko/caddyshack/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/bjblazko/caddyshack/releases/tag/v0.1.0
