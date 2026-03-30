---
date: 2026-03-29
status: done
---

# HTTP Status Code Tracking

Counts and visualizes HTTP response codes from the log.

## Implementation

- All status codes are counted as `NameCount` entries (name = code as string)
- Sorted ascending by code
- No limit — all distinct codes in the log are included
- Rendered as a horizontal bar chart in the dashboard (2-column row alongside Top Pages)

## Grouping (UI labels only)

Codes are displayed individually; grouping into 2xx/3xx/4xx/5xx classes is handled by the traffic segmentation filter, not this feature.
