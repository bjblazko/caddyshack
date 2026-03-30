---
date: 2026-03-29
status: done
---

# Top Pages Listing

Ranks the most-requested URIs, excluding static assets and error responses.

## Filtering Rules

A URI is counted as a "page" only if:
1. HTTP status code < 400
2. URI does not start with `/css/`, `/js/`, `/img/`, `/fonts/`, `/api`
3. URI does not end with `.css`, `.js`, `.png`, `.jpg`, `.svg`, `.ttf`, `.woff`, `.woff2`, `.ico`

For the error traffic filter (4xx+), condition 1 is inverted: only URIs with status ≥ 400 are counted.

## Output

Top 15 URIs by request count, sorted descending. Displayed as a table in the dashboard.
