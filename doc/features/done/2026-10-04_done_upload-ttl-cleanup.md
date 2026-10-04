---
date: 2026-10-04
status: done
---

# Upload TTL Cleanup

Uploaded log files are deleted from the temp directory once they have not been
used for a configurable time (`-upload-ttl`, default 1h). Leftovers from a
previous run are removed on startup.

## Motivation

Uploads are stored in `$TMPDIR/caddyshack` so filters can re-analyze them
(ADR-009). They contain unanonymized visitor IPs and were never deleted, which
contradicts the privacy model in `doc/specs/security.md`.

## Implementation Notes

- See ADR-010
- New flag `-upload-ttl` (Go duration, default `1h`)
- Every `/api/analyze` and `/api/events` access to an upload refreshes its mtime
- A background sweeper deletes uploads whose mtime is older than the TTL
- The upload directory is emptied on startup
- An expired upload answers `404` with a message asking to upload again
- Failed or truncated uploads are deleted immediately
