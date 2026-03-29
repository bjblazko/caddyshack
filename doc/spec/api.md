# CaddyShack API Specification

## Endpoints

### `POST /api/upload`

Upload a Caddy JSONL log file for analysis.

**Request**

- Content-Type: `multipart/form-data`
- Form field: `logfile` — the JSONL file
- Max size: 500 MB

**Response**

- Content-Type: `application/json`
- Status: `200 OK` on success, `400 Bad Request` on invalid input

**Response Body** — `Report` object:

```json
{
  "total_requests": 8,
  "unique_ips": 7,
  "total_bytes": 25967,
  "avg_response_ms": 1.51,
  "status_codes": [
    { "name": "200", "count": 6 },
    { "name": "404", "count": 2 }
  ],
  "top_pages": [
    { "name": "/", "count": 2 },
    { "name": "/blog/hello-world", "count": 1 }
  ],
  "browsers": [
    { "name": "Chrome", "count": 3 },
    { "name": "Firefox", "count": 1 }
  ],
  "operating_systems": [
    { "name": "macOS", "count": 3 },
    { "name": "Windows", "count": 1 }
  ],
  "daily_traffic": [
    { "date": "2026-03-23", "count": 8 }
  ],
  "top_visitors": [
    {
      "ip": "93.184.216.0",
      "count": 2,
      "country": "US",
      "country_name": "United States"
    }
  ],
  "countries": [
    { "code": "US", "name": "United States", "count": 4 },
    { "code": "DE", "name": "Germany", "count": 2 }
  ]
}
```

#### Report Fields

| Field | Type | Description |
|-------|------|-------------|
| `total_requests` | int | Total number of log entries processed |
| `unique_ips` | int | Count of distinct anonymized IPs |
| `total_bytes` | int | Sum of all response body sizes |
| `avg_response_ms` | float | Mean response time in milliseconds |
| `status_codes` | NameCount[] | HTTP status codes, sorted ascending |
| `top_pages` | NameCount[] | Top 15 URIs (excluding static assets, status < 400) |
| `browsers` | NameCount[] | Top 10 detected browsers, sorted by count descending |
| `operating_systems` | NameCount[] | Top 10 detected operating systems |
| `daily_traffic` | DayCount[] | Requests per day (UTC), sorted chronologically |
| `top_visitors` | VisitorInfo[] | Top 10 anonymized IPs with country info |
| `countries` | CountryCount[] | Top 15 countries by request count |

#### Nested Types

**NameCount**
```json
{ "name": "string", "count": 0 }
```

**DayCount**
```json
{ "date": "2006-01-02", "count": 0 }
```

**VisitorInfo**
```json
{ "ip": "string", "count": 0, "country": "XX", "country_name": "string" }
```

**CountryCount**
```json
{ "code": "XX", "name": "string", "count": 0 }
```

#### Error Responses

| Status | Condition |
|--------|-----------|
| `400 Bad Request` | Missing `logfile` field, invalid multipart form, or file exceeds 500 MB |

---

### `GET /api/health`

Health check endpoint.

**Response**

```json
{ "status": "ok" }
```

---

## Page Filtering Rules

Only URIs that satisfy all of the following are counted as "pages":

1. HTTP status code < 400
2. URI does **not** start with: `/css/`, `/js/`, `/img/`, `/fonts/`
3. URI does **not** end with: `.css`, `.js`, `.png`, `.jpg`, `.svg`, `.ttf`, `.woff`, `.woff2`, `.ico`
