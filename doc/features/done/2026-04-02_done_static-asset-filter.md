---
date: 2026-04-02
status: done
---

# Static Asset & Image Exclusion Filters

Two checkboxes in the filter bar let users exclude static file requests (JS, CSS, fonts, robots.txt) and/or image requests (PNG, JPG, SVG, etc.) from both the Statistics aggregation and the Single Events view.

## Motivation

Traffic logs contain many requests for static assets that are often uninteresting for analytics purposes. Hiding them makes it easier to focus on meaningful content requests.

## Implementation Notes

- Two new `FilterParams` fields: `IgnoreStatic` and `IgnoreImages`
- Two new URI classifiers: `isStaticResource()` and `isImageResource()`, applied in `passesNonHostFilters()`
- Query params: `ignore_static=1` and `ignore_images=1` on both `/api/analyze` and `/api/events`
- Frontend: two checkboxes in the filter bar; state variables `ignoreStatic` / `ignoreImages`
