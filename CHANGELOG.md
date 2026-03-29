# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0] - 2026-03-29

### Added

- Drag-and-drop upload of Caddy JSONL access log files
- Summary cards: total requests, unique IPs, data transferred, average response time
- World map with proportional bubbles showing geographic request distribution
- Browser and OS detection with horizontal bar charts
- Daily traffic trend chart
- HTTP status code breakdown (2xx, 3xx, 4xx, 5xx)
- Top pages listing (excluding static assets)
- Top visitors with anonymized IPs and country attribution
- Multi-host log support with per-host and aggregate views
- Success/error traffic segmentation (2xx vs 4xx+)
- GDPR-compliant IP anonymization (IPv4 last-octet zeroing, IPv6 prefix truncation)
- GeoIP country resolution via optional DB-IP Lite CSV database
- Server-side log file discovery at GET /api/logs
- Health check endpoint at GET /api/health
- Vanilla HTML/CSS/JavaScript frontend with no CDN dependencies
- D3.js and TopoJSON served locally for offline use
- Single-binary deployment with configurable listen address and GeoIP path

[Unreleased]: https://github.com/bjblazko/caddyshack/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/bjblazko/caddyshack/releases/tag/v0.1.0
