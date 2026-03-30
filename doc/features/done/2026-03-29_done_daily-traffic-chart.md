---
date: 2026-03-29
status: done
---

# Daily Traffic Chart

Vertical bar chart showing request count aggregated by UTC day.

## Implementation

- Aggregated in `analyzer.Analyze` by truncating the Unix timestamp to the UTC date (`YYYY-MM-DD`)
- All days in the log range are included; no limit on number of days
- Rendered via `Charts.renderVerticalBarChart()` on a `<canvas>` element
- Y-axis gridlines; date labels rotated for readability
- DPR-scaled for crisp rendering on Retina/HiDPI displays

## Data Shape

`DayCount[]` — `{ "date": "YYYY-MM-DD", "count": int }`, sorted chronologically.
