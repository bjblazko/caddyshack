# Caddy Log Format Reference

CaddyShack consumes Caddy v2 JSON access logs. Each line is a self-contained JSON object (JSONL format).

## Enabling JSON Logging in Caddy

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

Multiple site blocks can log to the same file. The `request.host` field distinguishes them.

## Schema

### Top-Level Fields

| Field | Type | Description |
|-------|------|-------------|
| `level` | string | Always `"info"` for access logs |
| `ts` | float | Unix timestamp with sub-second precision |
| `logger` | string | Logger name, e.g. `"http.log.access.log0"` |
| `msg` | string | Always `"handled request"` |
| `request` | object | Incoming HTTP request details |
| `bytes_read` | int | Bytes read from request body |
| `user_id` | string | Authenticated user (empty for anonymous) |
| `duration` | float | Request handling time in **seconds** |
| `size` | int | Response body size in bytes |
| `status` | int | HTTP response status code |
| `resp_headers` | object | Response headers as `{ "Name": ["value"] }` |

### `request` Object

| Field | Type | Description |
|-------|------|-------------|
| `remote_ip` | string | IP of the direct TCP connection |
| `remote_port` | string | Source port |
| `client_ip` | string | Resolved client IP (respects `trusted_proxies`) |
| `proto` | string | e.g. `"HTTP/2.0"`, `"HTTP/1.1"` |
| `method` | string | `GET`, `POST`, `HEAD`, etc. |
| `host` | string | Requested hostname |
| `uri` | string | Request URI including query string |
| `headers` | object | Request headers as `{ "Name": ["value"] }` |
| `tls` | object | TLS connection details |

### `tls` Object

| Field | Type | Description |
|-------|------|-------------|
| `resumed` | bool | Whether TLS session was resumed |
| `version` | int | IANA TLS version (772 = TLS 1.3, 771 = TLS 1.2) |
| `cipher_suite` | int | IANA cipher suite number |
| `proto` | string | ALPN negotiated protocol (`"h2"`, `"http/1.1"`) |
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
| Hostname | `request.host` |
