---
date: 2026-03-30
status: done
---

# Filter-then-Aggregate Dashboard Filters

Complete dashboard filtering: date range, host, status, country, browser, OS, and page. All filters are applied on the backend with AND logic in a single streaming pass before aggregation.

## Scope

This feature covers three sequential phases of work, all shipped together:

### Phase 1: Date Range Filter

Native `<input type="date">` controls (start + end) in the filter bar. Single-day selection by setting start = end. Clear button resets to full-period view.

### Phase 2: Dimension Dropdown Filters

Four additional dropdown filters: Country, Browser, OS, Page. Options are populated from the current backend response (already scoped to the active date range, host, and status filters).

### Phase 3: Backend Filter-then-Aggregate (ADR-009)

All filter dimensions moved to the backend. `analyzer.Analyze(r, FilterParams)` applies all conditions (host, date range, country, browser, OS, page, status) with AND logic in a single streaming pass, then aggregates. The frontend no longer performs any client-side re-aggregation.

Uploaded files are saved to `os.TempDir()/caddyshack/<hex_id>.jsonl`; subsequent filter changes re-analyze the same temp file via `GET /api/analyze?file=<id>&...` without re-uploading.

`POST /api/upload` returns `AnalysisResult` with `file_id`. `GET /api/analyze` accepts all filter params as query-string arguments and returns `AnalysisResult` without `file_id`.

`MultiHostReport`, `FullReport`, `DailyBreakdown`, and `DayBreakdown` types are removed. Replaced by `AnalysisResult { FileID, Hosts, Report }` and `FilterParams`.

## Acceptance criteria

- All 8 filter controls (host, status, start date, end date, country, browser, OS, page) are visible in the filter bar
- Every filter change triggers a backend re-analysis; all panels reflect the combined filter state
- Dimension dropdowns only show values present in the current response; auto-reset if the selected value disappears
- Filter hints per panel indicate whether each panel is filtered or shows all data
- No client-side re-aggregation; no external frontend dependencies added
- Uploaded file is re-analyzed on filter changes without re-uploading
