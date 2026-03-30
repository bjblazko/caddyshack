---
date: 2026-03-29
status: done
---

# Server-Side Log File Discovery

Allows the UI to list and analyze log files that already exist on the server without requiring a manual upload.

## Endpoints

**`GET /api/logs`** — lists available log files from `/var/log/caddy`. Returns a JSON array of `LogFileInfo` objects (name, size, modification time).

**`GET /api/analyze-local?name=<filename>`** — analyzes a named server-side log file. The `name` parameter must be a bare filename (no path traversal). Returns the same `MultiHostReport` as `/api/upload`.

## Use Case

Useful when CaddyShack runs on the same host as Caddy and the operator wants to analyze logs without downloading and re-uploading them.
