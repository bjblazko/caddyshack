# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.5.0] - 2026-10-04

### Added

- About dialog: version, who makes CaddyShack, links to the product page on huepattl.de, the GitHub repository and huepattl.de, what happens to your data, license and no-warranty statement
- "Licenses and thanks" page (`/licenses.html`) listing every built-in project with its use, license, website and license text, generated from `static/licenses/credits.json`; replaces `/licenses.txt`
- `/api/health` reports the version, stamped into release binaries and container images at build time
- `-logdir` flag for the directory of server-side logs (default `/var/log/caddy`)
- Empty state when no log is loaded; "Unknown" country with a hint when no GeoIP database is configured

### Changed

- All exclusions are on by default in the dashboard: static files and images are now hidden too, like monitors and bots, so the first view shows page traffic. Untick them to include those requests; the API still includes everything unless asked
- Redesigned interface following huepattl-rams-design: neutral surfaces with light and dark mode, IBM Plex Sans/Mono (self-hosted, OFL), flat panels, one primary action ("Upload log file"); colors in charts carry meaning
- Browsers and operating systems shown as ranked bars instead of donut charts; each entity keeps a fixed, color-vision-checked color, non-browsers are grey; status codes in green (2xx) and red (4xx/5xx) with their code as label
- One filter summary sentence replaces the filter hint badges on every panel; view tabs are links (`#statistics`, `#events`); errors appear on the page instead of browser dialogs; a status line replaces the loading overlay
- World map without graticule; size legend moved off land
- CI workflow on every push to `main` and every pull request (gofmt, go vet, go test, govulncheck, JavaScript syntax); releases are only built and published when it passes

### Fixed

- `/api/logs` returned `null` instead of `[]` for an empty log directory, which broke the page start

## [0.4.1] - 2026-10-04

### Security

- Build with Go 1.27.1: the downloadable release binaries up to v0.4.0 were built with Go 1.25.0 (no longer supported) and contained 45 known standard-library vulnerabilities, among them in `net/http`, `net/url` (query parsing) and `crypto/tls`; `govulncheck` now reports none. Container images were not affected to the same degree: they used the latest Go 1.25 patch release. Release builds use the `toolchain` version in `go.mod`, building from source needs Go 1.26+

### Changed

- macOS binaries require macOS 13 Ventura or later (Go 1.27)
- Release workflow uses current major versions of all actions (checkout v7, setup-go v7, action-gh-release v3, Docker actions v4/v6/v7)

## [0.4.0] - 2026-10-04

### Added

- Pre-anonymized logs: when the `User-Agent` header was stripped but `request.browser` / `request.os` are present, these names are used for the Browsers and OS charts and the client kind
- Footer with "IP geolocation by DB-IP" credit (required by the DB-IP Lite CC BY 4.0 license, shown when country data is resolved), license and no-warranty note, Caddy trademark note, and a link to `/licenses.txt` listing the license texts of all bundled third-party components (D3.js, topojson-client, world-atlas, Go runtime)
- README sections on privacy, licenses and trademarks
- Client kinds: every request is classified by User-Agent as Browser, Monitor (Uptime Kuma, UptimeRobot, …), Bot (crawlers, link previews, feed readers) or Script (curl, HTTP libraries, scanners); new "Exclude: Monitors" and "Exclude: Bots" checkboxes and `ignore_monitors` / `ignore_bots` query params, also accepted by `POST /api/upload`
- `-upload-ttl` flag (default `1h`): uploaded log files are deleted once they have not been used for this long; every analysis of an upload resets the timer, and the upload directory is emptied on startup (ADR-010)
- `/api/events` documented in the API spec

### Changed

- Container image is now based on `scratch` instead of `alpine`: only the static binary, its LICENSE and an empty `/tmp`; no shell inside the container anymore
- Monitors and bots are hidden by default in the dashboard, so totals are lower than before; untick the checkboxes to include them. The API still includes them unless asked otherwise
- Non-browser clients appear under their kind in the Browsers chart and the Browser filter (`Monitor`, `Bot`, `Script`); `curl` is now `Script`, UptimeRobot and Pingdom are `Monitor` instead of `Bot`
- Top Referrers shows external referrers only (renamed "Top External Referrers"): referrers pointing to any site found in the loaded log are internal navigation and left out; sites are derived from the log, nothing is configured
- Empty tables show "No data for the current filters." instead of a bare header
- `/api/events` caps `limit` at 200 instead of resetting values above 200 to 100, matching the documented behaviour
- An expired upload answers `404` with a message asking to upload the file again

### Fixed

- The same site appeared twice in the Site dropdown (`example.com` and `example.com:443`) with its traffic split between both; hosts are now grouped case-insensitively and without the default ports `:443`/`:80`
- Frontend assets are now embedded in the binary as ADR-002 intended; they were served from a `static/` directory next to the binary, so release binaries (shipped without that directory) showed no UI
- Pie chart legend swatches and the "World map data not available" message were unstyled because the CSP blocks inline `style` attributes
- `/api/events` with a negative `offset` crashed the request; it is now treated as 0
- Ranked lists (top pages, browsers, visitors, countries, …) ordered entries with equal counts randomly, so results could change between identical requests; ties are now ordered by name
- A failed or truncated upload could go unnoticed and stayed on disk; it is now reported and deleted

### Security

- Uploaded log files contain unanonymized IP addresses and were never deleted; they are now removed after `-upload-ttl` of inactivity (see Added)

- Added `Content-Security-Policy` response header at the application level: `default-src 'none'; script-src 'self'; style-src 'self'; img-src 'self'; font-src 'self'; connect-src 'self'; frame-ancestors 'none'; base-uri 'self'; form-action 'self'` — no `'unsafe-inline'` required since the app uses no inline scripts or styles

## [0.3.0] - 2026-04-02

### Added

- Static file exclusion filter: "Static files" checkbox hides JS, CSS, fonts, `robots.txt`, and `sitemap.xml` requests from all statistics and the Single Events view
- Image exclusion filter: "Images" checkbox hides PNG, JPG, SVG, WebP, ICO, and other image requests from all statistics and the Single Events view
- Search filter: glob-pattern text input matches URI, source IP, and HTTP Referer header; `*` spans any number of characters; no wildcard means literal exact match; empty input is ignored; positioned between HTTP Method and the Exclude checkboxes in the filter bar
- Referrer column in the Single Events table: shows the HTTP Referer header value, truncated with a tooltip for long URLs (same style as the URI column)
- Top Referrers panel in the Statistics tab: lists the 10 most frequent Referer header values with request counts

## [0.2.0] - 2026-03-30

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

[Unreleased]: https://github.com/bjblazko/caddyshack/compare/v0.5.0...HEAD
[0.5.0]: https://github.com/bjblazko/caddyshack/compare/v0.4.1...v0.5.0
[0.4.1]: https://github.com/bjblazko/caddyshack/compare/v0.4.0...v0.4.1
[0.4.0]: https://github.com/bjblazko/caddyshack/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/bjblazko/caddyshack/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/bjblazko/caddyshack/compare/v0.1.1...v0.2.0
[0.1.1]: https://github.com/bjblazko/caddyshack/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/bjblazko/caddyshack/releases/tag/v0.1.0
