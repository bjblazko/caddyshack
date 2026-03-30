# Spec: API

## Endpoints

### `POST /api/upload`

Upload a Caddy JSONL log file for analysis.

**Request**
- Content-Type: `multipart/form-data`
- Field: `logfile` — the JSONL file
- Max size: 500 MB

**Response** `200 OK` — `MultiHostReport` JSON object
**Error** `400 Bad Request` — missing field, invalid multipart, or file too large

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

### `GET /api/analyze-local?name=<filename>`

Analyze a server-side log file by name. `name` must be a bare filename — no path components.

**Response** `200 OK` — `MultiHostReport` (same shape as `/api/upload`)
**Error** `400 Bad Request` — missing or invalid `name` parameter

---

### `GET /api/health`

Health check.

**Response** `200 OK`
```json
{ "status": "ok" }
```

---

## Response Shape: `MultiHostReport`

```json
{
  "hosts": ["example.com", "blog.example.com"],
  "by_host": {
    "__all__": { "all": Report, "success": Report, "error": Report },
    "example.com": { "all": Report, "success": Report, "error": Report }
  }
}
```

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
  "countries":     [{ "code": "US", "name": "United States", "count": 4 }]
}
```

### Field Reference

| Field | Type | Limit | Notes |
|-------|------|-------|-------|
| `total_requests` | int | — | All parsed entries |
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
