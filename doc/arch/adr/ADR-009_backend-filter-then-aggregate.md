---
date: 2026-03-30
status: Accepted
supersedes: ADR-001
---

# ADR-009 — Backend filter-then-aggregate with temp file storage

## Context

The original architecture (ADR-001) computed a fixed set of report variants
up-front (all hosts × all/success/error status), stored the full
`MultiHostReport` in the browser, and applied all further filtering client-side.
This worked well for the initial host and status dimensions but broke down when
more filter dimensions were added (date range, country, browser, OS, page):
cross-dimensional accuracy was impossible without storing an exponentially
large per-dimension breakdown in the response, and the client-side
re-aggregation code grew complex and bug-prone.

The question was whether re-running the analysis on every filter change would
be too slow. Benchmarking showed that Go can stream-parse and aggregate a
100 MB JSONL file in well under a second, making per-request re-analysis
practical even for moderately large logs.

## Decision

All filter conditions are applied **before** aggregation, in a single streaming
pass on the backend. The frontend sends filter parameters as query-string
arguments to `GET /api/analyze` and receives a single, fully aggregated
`AnalysisResult`. No client-side re-aggregation is performed.

To allow re-analysis of uploaded files without re-uploading, `POST /api/upload`
now saves the file to the OS temp directory under a random hex ID and returns
that ID to the frontend. Subsequent `GET /api/analyze?file=<id>&...` calls
re-read the same temp file with different filter params. Local server logs are
read from disk on every call; no caching is needed there.

**Key design points:**

- `FilterParams` struct carries all filter dimensions: host, date range (start,
  end), country, browser, OS, page, status. All conditions are ANDed.
- Hosts available for the host dropdown are collected during the same pass from
  entries that satisfy every active filter *except* the host filter. This keeps
  the dropdown meaningful under any combination of other filters.
- The response type simplifies from `MultiHostReport → FullReport → Report` to
  `AnalysisResult { FileID, Hosts, Report }`.
- `DayBreakdown` / `DailyBreakdown` are removed; the simpler `DailyTraffic
  []DayCount` is restored.
- Temp files are not explicitly cleaned up; the OS temp directory lifecycle
  applies. This is acceptable for single-user and small-team deployments.

## Consequences

**Positive:**
- Every filter dimension (including cross-dimensional combinations like
  Browser + Country) is accurate by construction.
- The frontend shrinks significantly: no `applyDateFilter`, no per-day maps,
  no client-side merging or re-aggregation.
- The JSON response is smaller (no `DailyBreakdown` arrays).
- Adding a new filter dimension requires only: one field in `FilterParams`, one
  condition in `passesNonHostFilters`, one query param in the handler, and one
  state variable + event listener in the frontend.

**Negative / trade-offs:**
- Every filter change triggers a backend round-trip and a full re-parse of the
  log file. For files above ~200 MB this may take 1–3 seconds per change.
- Uploaded files are stored in the OS temp directory for the lifetime of the
  process (no explicit cleanup).
- The stateless guarantee of ADR-001 is partially relaxed: uploaded files
  persist on disk between requests (though no session state is kept in memory).
