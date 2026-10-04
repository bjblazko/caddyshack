# Spec: Log Parsing

## Input Format

CaddyShack consumes Caddy v2 JSON access logs in JSONL format (one JSON object per line). Each line is a self-contained log entry produced by Caddy's `format json` logging directive.

### Caddy Configuration

```caddyfile
example.com {
    log {
        output file /var/log/caddy/access.json {
            roll_size  50MiB
            roll_keep  7
            roll_keep_for 168h
        }
        format json
    }
}
```

Multiple site blocks may log to the same file; the `request.host` field distinguishes them.

## Log Entry Schema

### Top-Level Fields

| Field | Type | Used by CaddyShack |
|-------|------|-------------------|
| `ts` | float | Yes — Unix timestamp (seconds with sub-second precision) |
| `status` | int | Yes — HTTP response status code |
| `size` | int | Yes — Response body size in bytes |
| `duration` | float | Yes — Request handling time in seconds |
| `request` | object | Yes — see below |
| `bytes_read` | int | Parsed, not currently aggregated |
| `user_id` | string | Parsed, not currently aggregated |
| `level` | string | Ignored |
| `logger` | string | Ignored |
| `msg` | string | Ignored |
| `resp_headers` | object | Ignored |

### `request` Object

| Field | Type | Used by CaddyShack |
|-------|------|-------------------|
| `client_ip` | string | Yes — preferred client IP (respects `trusted_proxies`) |
| `remote_ip` | string | Yes — fallback when `client_ip` absent |
| `host` | string | Yes — virtual host grouping |
| `uri` | string | Yes — page tracking |
| `method` | string | Parsed, not currently aggregated |
| `proto` | string | Parsed, not currently aggregated |
| `headers` | object | Yes — `User-Agent` extracted for browser/OS detection |
| `tls` | object | Parsed, not currently aggregated |
| `remote_port` | string | Ignored |

### `tls` Object

| Field | Type | Notes |
|-------|------|-------|
| `resumed` | bool | — |
| `version` | int | IANA TLS version (772 = TLS 1.3, 771 = TLS 1.2) |
| `cipher_suite` | int | IANA cipher suite number |
| `proto` | string | ALPN negotiated protocol (`h2`, `http/1.1`) |
| `server_name` | string | SNI server name |
| `ech` | bool | Encrypted Client Hello used |

## Example Entry

```json
{
  "level": "info",
  "ts": 1774300000.123456,
  "logger": "http.log.access.log0",
  "msg": "handled request",
  "request": {
    "remote_ip": "93.184.216.34",
    "remote_port": "52431",
    "client_ip": "93.184.216.34",
    "proto": "HTTP/2.0",
    "method": "GET",
    "host": "example.com",
    "uri": "/blog/hello-world",
    "headers": {
      "User-Agent": ["Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) ..."],
      "Accept": ["text/html"]
    },
    "tls": {
      "resumed": false,
      "version": 772,
      "cipher_suite": 4865,
      "proto": "h2",
      "server_name": "example.com",
      "ech": false
    }
  },
  "bytes_read": 0,
  "user_id": "",
  "duration": 0.00312847,
  "size": 8421,
  "status": 200
}
```

## Fields Used by CaddyShack

| What | Source field |
|------|-------------|
| Timestamp | `ts` |
| Client IP | `request.client_ip` (fallback: `request.remote_ip`) |
| Request path | `request.uri` |
| HTTP method | `request.method` |
| Status code | `status` |
| Response size | `size` |
| Response time | `duration` (seconds → ms via ×1000) |
| User-Agent | `request.headers["User-Agent"][0]` |
| Referer | `request.headers["Referer"][0]` |
| Hostname | `request.host`, normalized to the canonical host (see `analysis.md`) |

## Package: `internal/logparser`

**`ParseStream(r io.Reader, fn func(LogEntry))`**

- Reads line-by-line using `bufio.Scanner` with a 1 MB line buffer
- Decodes each line as a `LogEntry` via `encoding/json`
- Malformed or incomplete lines are silently skipped
- Calls `fn` for each successfully parsed entry

### Go Types

```go
type LogEntry struct {
    Timestamp float64 `json:"ts"`
    Status    int     `json:"status"`
    Size      int     `json:"size"`
    Duration  float64 `json:"duration"`
    Request   Request `json:"request"`
}

type Request struct {
    ClientIP  string            `json:"client_ip"`
    RemoteIP  string            `json:"remote_ip"`
    URI       string            `json:"uri"`
    Method    string            `json:"method"`
    Host      string            `json:"host"`
    Proto     string            `json:"proto"`
    Headers   map[string][]string `json:"headers"`
    TLS       TLSInfo           `json:"tls"`
}

type TLSInfo struct {
    Version    int    `json:"version"`
    CipherSuite int   `json:"cipher_suite"`
    Proto      string `json:"proto"`
    ServerName string `json:"server_name"`
    Resumed    bool   `json:"resumed"`
}
```

## Upload Limits

- Maximum upload size: **500 MB**, enforced by `http.MaxBytesReader` before parsing begins
- Multipart form memory threshold: 10 MB (remainder spills to temp files, cleaned up by Go automatically)
