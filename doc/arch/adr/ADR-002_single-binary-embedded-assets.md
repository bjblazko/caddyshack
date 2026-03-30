# ADR-002: Single Binary with Embedded Static Assets

**Date:** 2026-03-29
**Status:** Accepted

## Context

Web applications typically ship the backend and frontend as separate artifacts: the binary plus a directory of static files that must be co-located and served correctly. This complicates deployment, particularly for self-hosters who expect to download one thing and run it.

The frontend consists of HTML, CSS, JavaScript, vendored D3.js + TopoJSON, and a country boundary JSON file (~400 KB total).

## Decision

All static assets are embedded into the Go binary at compile time using Go's `embed` package. The `http.FileServer` handler serves them from the in-process filesystem. The only external file the operator may need to provide is the optional GeoIP CSV.

Vendored JS libraries (D3.js, topojson-client) and geographic data are checked into the repository under `static/vendor/` and `static/data/` and embedded along with the application HTML/CSS/JS.

## Consequences

**Positive:**
- Deployment is a single file copy — no directory structure to manage
- No risk of a missing or mis-versioned frontend asset
- Works fully offline; no CDN or internet access needed
- Binary is self-contained and portable across environments

**Negative:**
- Binary size is larger than a backend-only binary (~400 KB of vendored assets)
- Updating a frontend asset requires recompiling and redistributing the binary
- Vendored JS libraries must be manually updated (no npm integration)
