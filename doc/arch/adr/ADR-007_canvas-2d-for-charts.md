# ADR-007: Canvas 2D API for Bar Charts — No Charting Library

**Date:** 2026-03-29
**Status:** Accepted

## Context

The dashboard requires bar charts for browsers, OS, status codes, and daily traffic. Options:

1. **Use a charting library** (Chart.js, Highcharts, etc.) — convenient but adds 50–200 KB of JavaScript and an external dependency.
2. **Use SVG directly** — good for scalability, but more verbose than Canvas for simple bar charts.
3. **Use the Canvas 2D API** — low-level but complete; no external dependency; crisp output with DPR scaling.

D3.js is already vendored for the world map. It could also produce bar charts via SVG, but using D3 for charts would couple the chart module to the map module unnecessarily.

## Decision

Bar charts are rendered using the HTML5 Canvas 2D API directly in `charts.js`. Two functions are exposed via the `Charts` namespace:

- `Charts.renderBarChart` — horizontal bar chart with percentage labels
- `Charts.renderVerticalBarChart` — vertical bar chart for daily traffic with gridlines and rotated date labels

Both functions scale the canvas by `window.devicePixelRatio` for crisp Retina/HiDPI rendering.

## Consequences

**Positive:**
- No additional JS library; `charts.js` is a self-contained ~200-line file
- Full control over layout, colors, and typography
- DPR scaling produces crisp output on all display densities
- No licence or supply-chain considerations

**Negative:**
- More code than calling `new Chart(ctx, config)` — bar chart rendering logic is manual
- Accessibility is limited: Canvas elements are bitmaps with no inherent semantic structure (no ARIA roles, no text selection)
- Animations require manual frame management if ever desired
