# 12. Glossary

The full domain glossary is maintained in [`../glossary.md`](../glossary.md).

Key terms most relevant to the architecture:

| Term | Short Definition |
|------|-----------------|
| **JSONL** | JSON Lines — one JSON object per line; Caddy's native access log format |
| **Stateless** | No database, no sessions, no stored results; uploads are kept on disk only until the Upload TTL expires |
| **Streaming** | Log parsed line-by-line; memory proportional to unique values, not total lines |
| **AnalysisResult** | API response: `file_id` (upload only), selectable `hosts`, and the `Report` |
| **Report** | Aggregated metrics for the entries that pass all active filters |
| **FilterParams** | All filter conditions, ANDed and applied before aggregation |
| **Canonical Host** | Host in lower case without default ports `:443`/`:80`; groups one site's requests |
| **IP Anonymization** | IPv4 last-octet zeroed; IPv6 truncated to first 3 groups — always applied |
| **GeoIP** | Country-level resolution from the optional DB-IP Lite CSV |
| **Request-Scoped Analysis** | Every filter change re-reads the log and recomputes the report in one request |
| **Upload TTL** | Idle time after which an uploaded log file is deleted (`-upload-ttl`, default 1h) |
| **Embedded Assets** | Frontend files compiled into the binary via Go's `embed` package |
| **Single Binary** | One compiled executable; no installation or config files required |

See [`../glossary.md`](../glossary.md) for the complete reference.
