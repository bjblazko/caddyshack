---
date: 2026-03-29
status: done
---

# Success/Error Traffic Segmentation

Allows the user to filter the dashboard between all traffic, successful (2xx) requests, and error (4xx+) requests.

## Implementation

- `analyzer.Analyze` produces a `FullReport` containing three `Report` objects: `All`, `Success`, `Error`
- All three are computed in a single pass and returned together
- No re-upload or re-analysis needed to switch views

## Filter Definitions

| Filter | Included status codes |
|--------|-----------------------|
| All | All codes |
| Success (2xx) | 200–299 |
| Errors (4xx–5xx) | 400–599 |

## UI

Toggle buttons in the dashboard header. Selecting a filter swaps the active `Report` and re-renders all sections.
