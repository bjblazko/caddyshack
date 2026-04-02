# Spec: API

## Endpoints

### `POST /api/upload`

Upload a Caddy JSONL log file. The file is saved to the OS temp directory and a
`file_id` is returned. Use the `file_id` with `GET /api/analyze` to re-analyze
with filter parameters.

**Request**
- Content-Type: `multipart/form-data`
- Field: `logfile` — the JSONL file
- Max size: 500 MB

**Response** `200 OK` — `AnalysisResult` JSON (includes `file_id`)
**Error** `400 Bad Request` — missing field, invalid multipart, or file too large

---

### `GET /api/analyze`

Analyze a log file with optional filter parameters. All filter conditions are
ANDed before aggregation.

**File source** (exactly one required):
- `file=<id>` — ID returned by `POST /api/upload`
- `name=<filename>` — bare filename from `/var/log/caddy` (no path components)

**Filter parameters** (all optional):
| Param | Description |
|-------|-------------|
| `host` | Virtual host (exact match). Omit for all hosts. |
| `start` | Start date `YYYY-MM-DD` (inclusive). |
| `end` | End date `YYYY-MM-DD` (inclusive). |
| `country` | Country name (exact match, e.g. `Germany`). |
| `browser` | Browser name (exact match, e.g. `Chrome`). |
| `os` | OS name (exact match, e.g. `macOS`). |
| `page` | Exact URI (e.g. `/blog/post-1`). |
| `status` | `success` (2xx) or `error` (4xx+). Omit for all. |
| `method` | HTTP method exact match (e.g. `GET`, `POST`). Omit for all. |
| `ignore_static` | `1` to exclude JS, CSS, fonts, robots.txt, sitemap.xml requests. |
| `ignore_images` | `1` to exclude PNG, JPG, SVG, ICO, and other image requests. |

**Response** `200 OK` — `AnalysisResult` (no `file_id` in this response)
**Error** `400 Bad Request` — missing or invalid params
**Error** `404 Not Found` — file not found

---

### `GET /api/logs`

List log files available on the server (from `/var/log/caddy`).

**Response** `200 OK` — JSON array of `LogFileInfo`:

```json
[
  { "name": "access.json", "size": 1048576, "modified": "2026-03-29T12:00:00Z" }
]
```

---

### `GET /api/health`

Health check.

**Response** `200 OK`
```json
{ "status": "ok" }
```

---

## Response Shape: `AnalysisResult`

```json
{
  "file_id": "a3f1...",
  "hosts": ["example.com", "blog.example.com"],
  "report": { ...Report... }
}
```

`file_id` is only present in the `POST /api/upload` response.
`hosts` lists virtual hosts found in entries that satisfy all active filters
except the host filter, so the list reflects what can usefully be selected.

## `Report` Object

```json
{
  "total_requests": 8,
  "unique_ips": 7,
  "total_bytes": 25967,
  "avg_response_ms": 1.51,
  "status_codes": [{ "name": "200", "count": 6 }],
  "top_pages":    [{ "name": "/", "count": 2 }],
  "browsers":     [{ "name": "Chrome", "count": 3 }],
  "operating_systems": [{ "name": "macOS", "count": 3 }],
  "daily_traffic": [{ "date": "2026-03-29", "count": 8 }],
  "top_visitors":  [{ "ip": "93.184.216.0", "count": 2, "country": "US", "country_name": "United States" }],
  "countries":     [{ "code": "US", "name": "United States", "count": 4 }],
  "methods":       [{ "name": "GET", "count": 6 }]
}
```

### Field Reference

| Field | Type | Limit | Notes |
|-------|------|-------|-------|
| `total_requests` | int | — | Entries passing all filters |
| `unique_ips` | int | — | Distinct anonymized IPs |
| `total_bytes` | int | — | Sum of response sizes |
| `avg_response_ms` | float64 | — | Mean duration × 1000 |
| `status_codes` | NameCount[] | all | Sorted ascending |
| `top_pages` | NameCount[] | 15 | Sorted descending |
| `browsers` | NameCount[] | 10 | Sorted descending |
| `operating_systems` | NameCount[] | 10 | Sorted descending |
| `daily_traffic` | DayCount[] | all | Sorted chronologically |
| `top_visitors` | VisitorInfo[] | 10 | Anonymized IPs |
| `countries` | CountryCount[] | 15 | Sorted descending |
| `methods` | NameCount[] | 20 | Sorted descending |
