---
date: 2026-03-29
status: done
---

# Core Analysis Pipeline

Single-pass streaming analysis of Caddy JSONL access logs over HTTP.

## What It Does

- User drops or selects a `.jsonl` file in the browser
- Frontend POSTs it as `multipart/form-data` to `POST /api/upload`
- Backend wraps the body in `MaxBytesReader` (500 MB limit) and streams it line-by-line via `bufio.Scanner`
- `logparser.ParseStream` decodes one JSON object per line; malformed lines are silently skipped
- `analyzer.Analyze` runs a single pass: for each entry it resolves GeoIP, anonymizes the IP, parses the User-Agent, and increments all map-based counters
- After the stream ends, counters are sorted and trimmed into a `Report` struct, serialized as JSON, and returned

## Design Constraints

- Stateless: no database, no sessions, no disk writes beyond Go's multipart temp handling
- Memory scales with unique values (IPs, URIs), not total log lines
- Standard library only (`net/http`, `bufio`, `encoding/json`)
