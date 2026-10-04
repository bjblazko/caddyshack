# Spec: Security & Privacy

## IP Anonymization

All IP addresses are truncated before inclusion in the API response. Implemented in `internal/anonymize`.

| IP version | Rule | Example |
|------------|------|---------|
| IPv4 | Zero the last octet | `93.184.216.34` → `93.184.216.0` |
| IPv6 | Keep first 3 groups, zero the rest | `2a01:4f8:c17::1` → `2a01:4f8:c17::` |

Processing order within a single request:
1. GeoIP lookup on the **original** IP
2. IP anonymized
3. Only the anonymized IP is stored in counters and included in the response

Raw IP addresses never leave the server process.

## GeoIP

Country-level resolution via the DB-IP Lite CSV (`internal/geoip`). Only the resulting country code (ISO 3166-1 alpha-2) appears in the response — never the original IP.

- IPv4 only; IPv6 always returns `"??"`
- Optional: if the CSV is absent, the app runs without country data (logs a warning, does not fail)
- Country code `"??"` is used for unresolved or IPv6 addresses

## Processing Model

- Uploaded log files (raw IPs included) are stored in `$TMPDIR/caddyshack` so filters can re-analyze them (ADR-009). They are deleted once unused for `-upload-ttl` (default 1h), on startup, and immediately if the upload fails (ADR-010)
- No database, no sessions; apart from uploads no state persists between requests
- All parsed data lives only in memory for the duration of a single HTTP request

## Upload Limit

`http.MaxBytesReader` enforces a 500 MB cap on uploaded files, preventing memory exhaustion from oversized inputs.

## Path Traversal Prevention

`GET /api/analyze` and `GET /api/events` accept only a bare filename in the `name` query parameter, and an upload ID in `file` that must not contain `/`, `\` or `.`. Path separators and `..` are rejected to prevent directory traversal.

## Recommendations for Operators

- Run CaddyShack on the same host as Caddy to avoid transmitting raw logs over the network
- If remote access is needed, upload pre-anonymized logs
- The GeoIP CSV contains only IP-range-to-country-code mappings; no PII
