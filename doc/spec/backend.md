# CaddyShack Backend Specification

## Technology

- Go 1.22+ (required for method-based routing in `http.ServeMux`)
- Standard library only, no external Go dependencies

## Packages

### `main` (`main.go`)

Entry point. Parses CLI flags, loads GeoIP database, registers routes, starts HTTP server.

**CLI Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `-addr` | `:8080` | Listen address |
| `-geodb` | `./data/dbip-country-lite.csv` | Path to DB-IP Lite CSV |

**Routes:**

| Method | Path | Handler |
|--------|------|---------|
| `POST` | `/api/upload` | `handler.Upload` |
| `GET` | `/api/health` | `handler.Health` |
| `GET` | `/*` | `http.FileServer` (serves `static/`) |

---

### `internal/logparser`

Parses Caddy JSONL log data.

**Types:**

- `LogEntry` — top-level log entry with `Timestamp`, `Status`, `Size`, `Duration`, `Request`
- `Request` — HTTP request with `ClientIP`, `RemoteIP`, `URI`, `Method`, `Host`, `Proto`, `Headers`, `TLS`
- `TLSInfo` — TLS metadata (`Version`, `CipherSuite`, `Proto`, `ServerName`)

**Functions:**

- `ParseStream(r io.Reader, fn func(LogEntry))` — reads JSONL line-by-line with a 1 MB buffer. Calls `fn` for each successfully parsed entry. Malformed lines are silently skipped.

---

### `internal/useragent`

Extracts browser and OS names from User-Agent strings.

**Function:**

- `Parse(ua string) (browser, os string)` — returns detected browser and OS names.

**Detection Order** (order matters for correct classification):

OS detection (specific before generic — order matters):
1. `"iPhone"` or `"iPad"` → iOS *(must precede macOS: iPhone UAs contain "Mac OS X")*
2. `"Windows"` → Windows
3. `"Macintosh"` or `"Mac OS X"` → macOS
4. `"Android"` → Android
5. `"CrOS"` → ChromeOS *(must precede Linux: ChromeOS UAs contain "Linux")*
6. `"Linux"` → Linux
7. (fallback) → Other

Browser detection:
1. `"curl/"` → curl
2. `"bot"` / `"spider"` / `"crawl"` (case-insensitive) → Bot
3. `"Edg/"` → Edge
4. `"OPR/"` or `"Opera"` → Opera
5. `"Vivaldi/"` → Vivaldi
6. `"Brave"` → Brave
7. `"Chrome/"` + `"Safari/"` → Chrome
8. `"Safari/"` without `"Chrome/"` → Safari
9. `"Firefox/"` → Firefox
10. (fallback) → Other

---

### `internal/anonymize`

GDPR-compliant IP anonymization.

**Function:**

- `IP(ip string) string`

**Rules:**
- IPv4: zero the last octet (`93.184.216.34` → `93.184.216.0`)
- IPv6: keep first 3 groups, zero the rest (`2a01:4f8:c17::1` → `2a01:4f8:c17::`)

---

### `internal/geoip`

Country-level GeoIP lookups using the DB-IP Lite CSV database.

**Functions:**

- `Load(path string)` — reads CSV into sorted `uint32` slices for binary search. Only IPv4 ranges are loaded. Logs a warning if the file is missing; the app continues without country data.
- `Lookup(ip string) string` — returns ISO 3166-1 alpha-2 country code. Returns `"??"` for IPv6 addresses, unknown IPs, or if no database is loaded.
- `CountryName(code string) string` — returns display name for a country code (e.g. `"DE"` → `"Germany"`). Falls back to the code itself if unknown.

**GeoIP Database:**

The free [DB-IP Lite](https://db-ip.com/db/download/ip-to-country-lite) CSV. Format: `start_ip,end_ip,country_code`. Not shipped with the application.

---

### `internal/analyzer`

Core aggregation engine.

**Function:**

- `Analyze(r io.Reader) *Report` — streams log entries via `logparser.ParseStream`, builds map-based counters for all dimensions, then sorts and trims into the `Report` struct.

**Processing per entry:**
1. Extract `client_ip` (fallback to `remote_ip`)
2. GeoIP lookup on original IP
3. Anonymize IP
4. Parse User-Agent for browser and OS
5. Increment counters: status codes, browsers, OS, IPs, countries, daily, pages

**Page filtering:** URIs are counted as pages only if status < 400 and the URI does not match asset prefixes (`/css/`, `/js/`, `/img/`, `/fonts/`) or asset extensions (`.css`, `.js`, `.png`, `.jpg`, `.svg`, `.ttf`, `.woff`, `.woff2`, `.ico`).

**Report limits:**
- Top pages: 15
- Browsers: 10
- Operating systems: 10
- Top visitors: 10
- Countries: 15
- Daily traffic: all days (no limit)
- Status codes: all codes (sorted ascending)

---

### `internal/handler`

HTTP request handlers.

**`Upload`** — `POST /api/upload`
1. Wraps request body in `MaxBytesReader` (500 MB)
2. Parses multipart form (10 MB memory threshold)
3. Extracts `logfile` field
4. Passes file reader to `analyzer.Analyze`
5. JSON-encodes and returns the `Report`

**`Health`** — `GET /api/health`
Returns `{"status":"ok"}`.
