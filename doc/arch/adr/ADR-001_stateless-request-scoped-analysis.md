# ADR-001: Stateless, Request-Scoped Analysis

**Date:** 2026-03-29
**Status:** Superseded by ADR-009

## Context

A log analytics tool needs to store parsed data somewhere between when it is computed and when it is presented to the user. Options range from a full database to in-memory session state to fully request-scoped computation.

The project targets self-hosters who want minimal operational overhead. Adding a database (even SQLite) would require schema management, migrations, file permissions, and backup considerations. Adding session state would require a session store and complicate deployment behind load balancers.

## Decision

All log parsing and aggregation happens entirely within the scope of a single HTTP request. The `analyzer.Analyze` function receives an `io.Reader`, performs a streaming single-pass analysis, and returns a `Report` struct. No result is stored on the server. The moment the HTTP response is written, all parsed data is eligible for garbage collection.

The only shared server-side state is the GeoIP lookup table, which is loaded once at startup and thereafter read-only.

## Consequences

**Positive:**
- Zero persistence layer: no database, no schema, no migrations
- Trivially deployable: copy and run the binary
- Safe for horizontal scaling: no coordination between instances needed
- Privacy: analysis results are never stored and cannot be leaked from server storage

**Negative:**
- Re-uploading the same file repeats the full analysis (no caching)
- No support for comparing analyses across sessions or building trend history
- Very large files must be re-uploaded each time the operator wants a fresh view
