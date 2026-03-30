# ADR-006: GeoIP as Optional Feature with Graceful Degradation

**Date:** 2026-03-29
**Status:** Accepted

## Context

Geographic analytics (country breakdown, world map) require a mapping from IP address to country. Several options exist:

1. **Bundle a GeoIP database in the binary** — convenient but adds 5–20 MB to binary size and introduces licence obligations. DB-IP Lite uses a Creative Commons Attribution licence that requires attribution in distributed products.
2. **Call an external GeoIP API at runtime** — breaks offline operation and introduces a network dependency.
3. **Ship the database separately, load at startup** — keeps the binary licence-clean and small; operators who want country data must supply the CSV.

## Decision

The GeoIP database (DB-IP Lite CSV) is not bundled with the binary. Its path is configurable via the `-geodb` CLI flag (default: `./data/dbip-country-lite.csv`).

`geoip.Load` is called at startup. If the file is absent or unreadable, a warning is logged and the in-memory lookup table is left empty. All subsequent `geoip.Lookup` calls return `"??"`. The application continues fully functional for all non-geographic analytics.

## Consequences

**Positive:**
- Binary is licence-clean: no third-party data embedded
- Binary stays small (~same size as without GeoIP)
- Operators who don't need country data have one fewer thing to configure
- Works fully offline: once loaded at startup, no network calls are made

**Negative:**
- First-time setup requires an extra manual step: download the CSV from DB-IP and point the flag at it
- Without the database, the world map is empty and all country fields show `??`
- Database does not auto-update; operators must manually refresh it periodically for accuracy
