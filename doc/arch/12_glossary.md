# 12. Glossary

The full domain glossary is maintained in [`../glossary.md`](../glossary.md).

Key terms most relevant to the architecture:

| Term | Short Definition |
|------|-----------------|
| **JSONL** | JSON Lines — one JSON object per line; Caddy's native access log format |
| **Stateless** | No server-side state between requests; each upload is self-contained |
| **Streaming** | Log parsed line-by-line; memory proportional to unique values, not total lines |
| **MultiHostReport** | Root API response: per-host and aggregate `FullReport` objects |
| **FullReport** | Three `Report` objects: All traffic, Success (2xx), Error (4xx+) |
| **Report** | Aggregated metrics for one host + one status filter |
| **IP Anonymization** | IPv4 last-octet zeroed; IPv6 truncated to first 3 groups — always applied |
| **GeoIP** | Country-level resolution from the optional DB-IP Lite CSV |
| **Request-Scoped Analysis** | All data lives only for the duration of a single HTTP request |
| **Embedded Assets** | Frontend files compiled into the binary via Go's `embed` package |
| **Single Binary** | One compiled executable; no installation or config files required |

See [`../glossary.md`](../glossary.md) for the complete reference.
