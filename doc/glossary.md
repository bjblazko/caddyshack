# CaddyShack Glossary

Terms and concepts used throughout the CaddyShack codebase and documentation.

---

## Core Concepts

**Access Log**
HTTP server log produced by Caddy containing per-request metadata: timestamp, client IP, URI, method, status code, response size, and duration. CaddyShack's primary input.

**JSONL (JSON Lines)**
Log format used by Caddy: one JSON object per line. CaddyShack parses these line-by-line using a streaming scanner.

**Report**
The JSON object returned by the API after analysis. Contains summary cards, top-N tables, daily traffic data, and chart data for a given host and status filter.

**FullReport**
A wrapper around three `Report` objects: one for all traffic, one for 2xx (success), one for 4xx+ (errors). Enables client-side status filtering without re-analysis.

**MultiHostReport**
The root API response structure. Contains a list of hostnames and a `by_host` map (host → FullReport) covering both per-host and aggregate ("All Sites") views.

**Stateless**
Design principle: no database, no sessions, no disk writes. All parsing and aggregation happens in memory within a single HTTP request and is discarded after the response is sent.

**Streaming**
Design principle: logs are parsed line-by-line with `bufio.Scanner`. Memory usage grows with the number of unique values (IPs, URIs), not the total number of log lines.

**Request-Scoped Analysis**
All data lives only for the duration of one HTTP request. There is no shared state between requests.

---

## Data Structures

**LogEntry**
Top-level Go struct representing one parsed Caddy log line. Fields: `ts`, `status`, `size`, `duration`, `request`, `user_id`, `bytes_read`.

**Request Object**
Nested struct inside `LogEntry`. Fields: `client_ip`, `remote_ip`, `uri`, `method`, `host`, `proto`, `headers`, `tls`.

**TLSInfo**
TLS connection metadata embedded in the request object: cipher suite, protocol version, server name, and whether the session was resumed.

**NameCount**
General-purpose tuple `{ "name": string, "count": int }`. Used for browsers, operating systems, status codes, and top pages.

**DayCount**
Daily aggregation tuple `{ "date": "YYYY-MM-DD", "count": int }`. Used for the traffic-over-time chart.

**VisitorInfo**
Per-visitor record `{ "ip": string, "count": int, "country": "XX", "country_name": string }`. The `ip` field contains an anonymized address.

**CountryCount**
Per-country record `{ "code": "XX", "name": string, "count": int }`. Used for the country table and world map.

---

## IP & Privacy

**IP Anonymization**
GDPR-oriented IP truncation applied before any data leaves the server. For IPv4, the last octet is zeroed (`93.184.216.34` → `93.184.216.0`). For IPv6, only the first three groups are kept (`2a01:4f8:c17:1::` → `2a01:4f8:c17::`). Implemented in `anonymize.go`.

**ClientIP vs RemoteIP**
`client_ip` is the resolved visitor IP after Caddy applies trusted-proxy rules (e.g. reading `X-Forwarded-For`). `remote_ip` is the raw TCP peer address. CaddyShack prefers `client_ip` and falls back to `remote_ip`.

**GeoIP Lookup**
Country-level geolocation performed on the original (non-anonymized) IP at analysis time using the DB-IP Lite CSV database. Only the country code is included in the response — the original IP is never sent to the client.

**Privacy by Default**
Design principle: IPs are anonymized before any response, GeoIP is resolved server-side, and no raw IP addresses or sensitive identifiers appear in the output.

---

## Geographic Visualization

**GeoIP Database (DB-IP Lite)**
A free CSV file mapping IPv4 ranges to ISO 3166-1 alpha-2 country codes. Format: `start_ip,end_ip,country_code`. Loaded at startup and searched via binary search on uint32-converted IPs.

**Country Code (ISO 3166-1 alpha-2)**
Two-letter country identifier (e.g. `US`, `DE`, `GB`). Returns `??` for unrecognized or IPv6 addresses not covered by the database.

**Country Centroid**
Approximate `(longitude, latitude)` of a country's geographic center, used to position bubbles on the world map.

**TopoJSON**
A compact topology-based variant of GeoJSON used for country boundary data. Sourced from the Natural Earth 110m dataset and served locally under `/data/`.

**Natural Earth 110m**
Free, public-domain geographic dataset providing country boundaries at 110-metre scale resolution. Used as the base map.

**Bubble Map / Proportional Bubbles**
The world map visualization: circles sized by request count using `d3.scaleSqrt()` so that area is proportional to count. Hovering a bubble shows a tooltip with the country name and count.

**Graticule**
The latitude/longitude grid lines drawn on the world map for geographic reference.

**D3 Projection (geoNaturalEarth1)**
The D3.js map projection that converts geographic coordinates (lat/lon) to SVG pixel coordinates, fitting the world to the container size.

---

## HTTP & Protocol

**Status Code**
Standard HTTP response code. CaddyShack groups them into classes: 2xx (success), 3xx (redirect), 4xx (client error), 5xx (server error).

**Status Filter**
UI-level segmentation of traffic into three views: *All* (every request), *Success (2xx)*, and *Errors (4xx–5xx)*. Each corresponds to one of the three reports in a `FullReport`.

**User-Agent**
HTTP request header identifying the client software. Parsed by `useragent.go` to extract a browser name and OS name via ordered string matching.

**Bot Detection**
User-agents containing `bot`, `spider`, or `crawl` (case-insensitive) are classified as "Bot" rather than a named browser.

**ALPN (Application-Layer Protocol Negotiation)**
TLS extension that negotiates the HTTP protocol version. Appears as `h2` (HTTP/2) or `http/1.1` in the `tls.proto` log field.

**Payload Limit**
Maximum upload size enforced server-side via `MaxBytesReader`: 500 MB. Prevents excessive memory use during analysis.

---

## Asset Filtering

**Static Assets**
Non-page requests excluded from the "Top Pages" table. Identified by path prefix (`/css/`, `/js/`, `/img/`, `/fonts/`, `/api`) or file extension (`.css`, `.js`, `.png`, `.jpg`, `.svg`, `.ico`, `.woff`, `.woff2`, `.ttf`).

**Page Filtering**
For the success filter: only URIs with status < 400 and non-asset paths count as pages. For the error filter: URIs with status ≥ 400 count as pages regardless of extension.

**Top N**
Convention for limiting ranked results. Used throughout: top 15 pages, top 10 browsers/OS, top 10 visitors, top 15 countries.

---

## Frontend Architecture

**Single-Page Dashboard**
All UI sections live in one `index.html` file. Sections are hidden until a log file is loaded; there is no client-side routing.

**Host Dropdown**
UI control to switch between virtual hosts found in a multi-host log. Defaults to "All Sites" (aggregate view).

**Canvas 2D API**
HTML5 canvas used for rendering bar charts (horizontal and vertical). Chosen for pixel-level control without a charting library dependency.

**DPR (Device Pixel Ratio)**
`window.devicePixelRatio` used to scale canvas rendering for high-DPI (Retina) displays, keeping charts crisp.

**Chart Namespace**
JavaScript module object (`Charts`) exposing `renderBarChart()` and `renderVerticalBarChart()`. Encapsulates all canvas chart logic.

**WorldMap Namespace**
JavaScript module object (`WorldMap`) exposing `render()`. Encapsulates all D3 map rendering logic.

**Tooltip**
Floating UI element that appears on hover (e.g. over a map bubble), positioned using `clientX`/`clientY` mouse coordinates.

**Offline-First**
D3.js, TopoJSON, and geographic data are served locally from `/vendor/` and `/data/`. No external CDN requests are made.

**Drag-and-Drop Upload**
The primary file upload mechanism. A JSONL file is dropped onto the upload zone (or selected via file picker), converted to `FormData`, and POSTed to `/api/upload`.

**Multipart Form Data**
The HTTP encoding used for file uploads. The file field name is `logfile`. Parsed server-side with `r.ParseMultipartForm()`.

---

## API Endpoints

| Endpoint | Method | Purpose |
|---|---|---|
| `/api/upload` | POST | Upload and analyze a log file (multipart, field: `logfile`) |
| `/api/logs` | GET | List available server-side log files from `/var/log/caddy` |
| `/api/analyze-local` | GET | Analyze a server-side log file by name (`?name=<filename>`) |
| `/api/health` | GET | Health check; returns `{"status":"ok"}` |

---

## Infrastructure & Deployment

**Single Binary**
The deployment model: one compiled Go executable with all static assets embedded. No installation steps, no configuration files, no external runtime dependencies (GeoIP CSV is optional).

**Embedded Static Assets**
Frontend files (HTML, CSS, JS, vendor libraries, geographic data) compiled into the binary via Go's `embed` package. Served from the `static/` directory tree.

**Multi-stage Dockerfile**
Docker build strategy with separate builder and runtime stages. The builder compiles the Go binary; the runtime stage copies only the binary to keep the final image small.

**GHCR (GitHub Container Registry)**
Where CaddyShack Docker images are published. Images are built for `linux/amd64` and `linux/arm64`.

**Multi-arch Images**
Docker images supporting multiple CPU architectures (amd64 and arm64) from a single image reference.

**Health Check Endpoint**
`GET /api/health` returns `{"status":"ok"}`. Used for container readiness probes and uptime monitoring.

---

## Caddy Configuration Terms

**Caddyfile**
Caddy's native configuration format. Used to enable structured JSON access logging with the `format json` directive and to configure log output paths.

**Log Rolling / Log Rotation**
Caddy configuration directives (`roll_size`, `roll_keep`, `roll_keep_for`) that bound log file growth by size and age.

**Trusted Proxies**
Caddy directive specifying which upstream proxy IPs to trust for `X-Forwarded-For` header extraction, determining the correct `client_ip` value.

**Virtual Host / Multi-host**
A single Caddy instance (and thus a single log file) can serve multiple hostnames. CaddyShack groups log entries by the `host` field for per-host analysis.

---

## Testing & Development

**Sample Log Generator**
`testdata/generate.py` — a Python script that produces synthetic Caddy JSONL log files for development and testing.

**Test Data**
`testdata/sample-example.jsonl` — a pre-generated sample log file included in the repository for quick manual testing.
