---
date: 2026-03-29
status: done
---

# Multi-Host Log Support

A single Caddy log file can contain requests to multiple hostnames. CaddyShack produces per-host reports as well as an aggregate view.

## Implementation

- `analyzer.Analyze` groups entries by `request.host`
- Returns a `MultiHostReport`: `{ hosts: string[], by_host: { [host]: FullReport } }`
- The special key `"__all__"` holds the aggregate across all hosts

## UI

- Host dropdown appears when more than one host is detected
- Default selection: "All Sites" (aggregate)
- Switching hosts re-renders all dashboard sections with per-host data without a new upload
