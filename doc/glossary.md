# CaddyShack Glossary

Terms and concepts used throughout the CaddyShack codebase and documentation.

---

## Core Concepts

**Access Log**
HTTP server log produced by Caddy containing per-request metadata: timestamp, client IP, URI, method, status code, response size, and duration. CaddyShack's primary input.

**JSONL (JSON Lines)**
Log format used by Caddy: one JSON object per line. CaddyShack parses these line-by-line using a streaming scanner.

**Report**
The aggregated metrics for all entries that pass the active filters: summary cards, top-N tables, daily traffic and chart data. Computed fresh on every request.

**AnalysisResult**
The response of `/api/upload` and `/api/analyze`: `{ file_id?, hosts, report }`. `hosts` lists the canonical hosts selectable under all other active filters.

**FilterParams**
The set of filter conditions sent as query parameters (host, date range, status, country, browser, OS, page, method, search, static/image exclusion). All conditions are ANDed and applied before aggregation (ADR-009).

**Stateless**
Design principle: no database and no sessions. Analysis results are never stored. The only data kept between requests are uploaded log files, held temporarily on disk until the Upload TTL expires (ADR-009, ADR-010).

**Streaming**
Design principle: logs are parsed line-by-line with `bufio.Scanner`. Memory usage grows with the number of unique values (IPs, URIs), not the total number of log lines.

**Request-Scoped Analysis**
Every filter change re-reads the log file and recomputes the report within one HTTP request. Parsed entries and counters are discarded once the response is sent.

**Upload ID (`file_id`)**
Random 32-character hex ID under which an uploaded log file is stored in `$TMPDIR/caddyshack`. Passed as `file=<id>` to `/api/analyze` and `/api/events`.

**Upload TTL**
Idle time (`-upload-ttl`, default 1h) after which an uploaded log file is deleted. Every analysis of the upload resets the timer; the directory is emptied on startup (ADR-010).

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

**D3 Projection (geoNaturalEarth1)**
The D3.js map projection that converts geographic coordinates (lat/lon) to SVG pixel coordinates, fitting the world to the container size.

---

## HTTP & Protocol

**Status Code**
Standard HTTP response code. CaddyShack groups them into classes: 2xx (success), 3xx (redirect), 4xx (client error), 5xx (server error).

**Status Filter**
Segmentation of traffic into *All* (every request), *Success (2xx)* and *Errors (4xx–5xx)*. Sent as `status=success|error` and applied on the backend before aggregation.

**User-Agent**
HTTP request header identifying the client software. Parsed by `useragent.go` to extract a browser name and OS name via ordered string matching.

**Client Kind**
Classification of a request by its User-Agent: *Browser*, *Monitor* (uptime and health checks such as Uptime Kuma), *Bot* (crawlers, link previews, feed readers) or *Script* (curl, HTTP libraries, scanners, empty User-Agent). For non-browsers the kind is shown as the browser name. Built-in rules, nothing configured.

**Bot Detection**
Part of the Client Kind classification: user-agents containing `bot`, `spider`, `crawl` or a known crawler, preview or feed-reader token (case-insensitive) are "Bot" — unless a monitor rule matches first (UptimeRobot, Pingdom).

**Monitor / Bot Exclusion**
Filters `ignore_monitors=1` and `ignore_bots=1`. The UI enables both by default (like the static files and images exclusions), so statistics show page requests from people and scripts; unticking shows monitors or bots again. Scripts have no exclusion so scanner probes stay visible.

**ALPN (Application-Layer Protocol Negotiation)**
TLS extension that negotiates the HTTP protocol version. Appears as `h2` (HTTP/2) or `http/1.1` in the `tls.proto` log field.

**Payload Limit**
Maximum upload size enforced server-side via `MaxBytesReader`: 500 MB. Prevents excessive memory use during analysis.

---

## Asset Filtering

**Static Assets**
Non-page requests excluded from the "Top Pages" table. Identified by path prefix (`/css/`, `/js/`, `/img/`, `/fonts/`, `/api`) or file extension (`.css`, `.js`, `.png`, `.jpg`, `.svg`, `.ico`, `.woff`, `.woff2`, `.ttf`). Not to be confused with the Static Files and Images exclusion filters, which use their own lists.

**Static Files / Images Exclusion**
Filters (`ignore_static=1`, `ignore_images=1`) that remove requests from all statistics and the Single Events view; on by default in the UI, off unless requested in the API. Static files: JS, CSS, source maps, fonts, `robots.txt`, `sitemap.xml`, and paths under `/css/`, `/js/`, `/fonts/`. Images: PNG, JPG/JPEG, GIF, SVG, WebP, ICO, BMP, AVIF, and paths under `/img/`, `/images/`. The query string is ignored when matching extensions.

**Glob Search**
Search filter (`search=`) matched case-insensitively against URI, client IP and Referer. `*` matches any number of characters; without `*` the value must match exactly.

**Page Filtering**
Only non-asset URIs count as pages. Without a status filter or with the success filter, only responses with status < 400 count; with the error filter, error responses count as pages too.

**Top N**
Convention for limiting ranked results: top 15 pages and countries, top 10 browsers, OS, visitors and referrers, top 20 HTTP methods. Entries with equal counts are ordered by name.

**External Referrer**
A Referer whose host is not one of the sites in the loaded log. Only these appear in Top External Referrers; navigation within or between the operator's own sites is left out. Own sites are derived from the log, never configured.

**Single Events**
Tab listing individual filtered log entries newest first, paginated via `/api/events` (100 per page, at most 200). IPs are anonymized.

---

## Frontend Architecture

**Single-Page Dashboard**
All UI sections live in one `index.html` file. Sections are hidden until a log file is loaded; there is no client-side routing.

**Canonical Host**
The form under which requests to one site are grouped: the `Host` header in lower case without the default ports `:443` and `:80` (`huepattl.de:443` → `huepattl.de`). Non-default ports remain part of the host. Used for the host list, the host filter and the Single Events view.

**Host Dropdown**
UI control to switch between virtual hosts found in a multi-host log. Defaults to "All Sites" (aggregate view).

**Canvas 2D API**
HTML5 canvas used for all bar charts: ranked horizontal bars (browsers, operating systems, status codes) and vertical bars (daily traffic). Chosen for pixel-level control without a charting library dependency. Colors are read from the CSS tokens at draw time.

**DPR (Device Pixel Ratio)**
`window.devicePixelRatio` used to scale canvas rendering for high-DPI (Retina) displays, keeping charts crisp.

**Chart Namespace**
JavaScript module object (`Charts`) exposing `renderBarChart()`, `renderVerticalBarChart()` and the color rules `entityColor()` / `statusColor()`. A shared `Tooltip` serves charts and map.

**Entity Color**
Fixed chart color per browser or OS name (Chrome always blue, Windows always blue, …), so filters that change the ranking never repaint an entity. Non-browser clients (Bot, Monitor, Script, Other) stay grey.

**Filter Summary**
The one sentence under the filters that states what all views show, e.g. "Showing example.com, without monitors and bots." Replaced the per-panel filter hint badges.

**WorldMap Namespace**
JavaScript module object (`WorldMap`) exposing `render()`. Encapsulates all D3 map rendering logic.

**Tooltip**
Floating UI element that appears on hover (e.g. over a map bubble), positioned using `clientX`/`clientY` mouse coordinates.

**Offline-First**
D3.js, TopoJSON, and geographic data are served locally from `/vendor/` and `/data/`. No external CDN requests are made.

**Drag-and-Drop Upload**
A log file dropped anywhere on the page is uploaded, the same as choosing it with "Upload log file". It is sent as `FormData` to `/api/upload`.

**Multipart Form Data**
The HTTP encoding used for file uploads. The file field name is `logfile`. Parsed server-side with `r.ParseMultipartForm()`.

---

## API Endpoints

| Endpoint | Method | Purpose |
|---|---|---|
| `/api/upload` | POST | Store and analyze a log file (multipart, field: `logfile`); returns `file_id` |
| `/api/logs` | GET | List available server-side log files from `/var/log/caddy` |
| `/api/analyze` | GET | Analyze an upload (`file=<id>`) or server-side log (`name=<filename>`) with filters |
| `/api/events` | GET | Filtered single log entries, paginated (`offset`, `limit`) |
| `/api/health` | GET | Health check; returns `{"status":"ok"}` |

---

## Infrastructure & Deployment

**Single Binary**
The deployment model: one compiled Go executable with all static assets embedded. No installation steps, no configuration files, no external runtime dependencies (GeoIP CSV is optional).

**Embedded Static Assets**
Frontend files (HTML, CSS, JS, vendor libraries, geographic data) compiled into the binary via Go's `embed` package. Served from the `static/` directory tree.

**Multi-stage Dockerfile**
Docker build strategy with separate builder and runtime stages. The builder compiles a static Go binary; the runtime stage is `scratch` with only the binary, the project LICENSE and an empty `/tmp`.

**Licenses and Notices**
The page `/licenses.html` ("Licenses and thanks"), rendered from `static/licenses/credits.json`: every project CaddyShack is built on with its use, license, website and embedded license text, plus CaddyShack's own license, the no-warranty statement and the Caddy trademark note. Linked from the About dialog and the footer.

**About Dialog**
The dialog opened by "About" (or `#about`): version, who makes CaddyShack, links to the product page, GitHub and huepattl.de, what happens to your data, license.

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
A single Caddy instance (and thus a single log file) can serve multiple hostnames. CaddyShack groups log entries by their Canonical Host for per-host analysis.

---

## Testing & Development

**Sample Log Generator**
`testdata/generate.py` — a Python script that produces synthetic Caddy JSONL log files for development and testing.

**Test Data**
`testdata/sample-example.jsonl` — a pre-generated sample log file included in the repository for quick manual testing.
