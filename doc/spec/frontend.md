# CaddyShack Frontend Specification

## Technology

- Vanilla HTML5, CSS3, JavaScript (ES2020+)
- Canvas 2D API for bar charts
- D3.js v7 + topojson-client v3 for the world map (served locally from `/vendor/`)
- Natural Earth 110m TopoJSON for country boundaries (served from `/data/countries-110m.json`)
- No build step, no bundler, no npm

## Dashboard Layout

The UI is a single-page dashboard (`index.html`). All sections are hidden until a log file is uploaded.

### Sections (top to bottom)

1. **Header** — logo, title ("CaddyShack"), subtitle, drag-and-drop upload zone
2. **Summary Cards** (4-column grid) — total requests, unique IPs, data transferred, avg response time
3. **World Map + Countries** (2-column) — D3 bubble map on the left, country breakdown table on the right
4. **Browsers + Operating Systems** (2-column) — horizontal bar charts rendered on `<canvas>`
5. **Daily Traffic** (full-width) — vertical bar chart showing requests per day
6. **Status Codes + Top Pages** (2-column) — bar chart and table
7. **Top Visitors** (full-width) — table with anonymized IPs, country, request count

### Responsive Behavior

- Desktop (>768px): 2-column grids, 4-column card row
- Mobile (<=768px): single-column layout, 2-column card row

## JavaScript Modules

### `app.js`

Main application orchestrator.

- Sets up drag-and-drop and file input event listeners on the upload zone
- On file selection: creates `FormData`, POSTs to `/api/upload`, shows loading overlay
- On response: calls render functions from `charts.js` and `map.js`, populates tables via DOM manipulation
- Utility: `formatBytes()` for human-readable byte sizes

### `charts.js`

Exposes a `Charts` namespace with two functions:

- `Charts.renderBarChart(canvasId, labels, values, total, color)` — horizontal bar chart with percentage labels. Handles DPR scaling for crisp rendering on Retina displays.
- `Charts.renderVerticalBarChart(canvasId, labels, values)` — vertical bar chart for daily traffic. Y-axis gridlines, rotated date labels.

### `map.js`

Exposes a `WorldMap` namespace:

- `WorldMap.render(containerId, countries)` — fetches `/data/countries-110m.json`, draws country outlines with D3's Natural Earth projection, overlays proportional bubbles at country centroids.
- Bubble sizing uses `d3.scaleSqrt()` for area-proportional representation.
- Hover tooltips show country name and request count.
- Legend with reference bubble sizes in the bottom-right corner.
- Color scheme: semi-transparent green fill (`rgba(74, 140, 63, 0.55)`), dark green stroke.

## Color Scheme

| Token | Hex | Usage |
|-------|-----|-------|
| `--green-dark` | `#2d5a27` | Header background, headings, card value text |
| `--green-mid` | `#4a8c3f` | Card top border, table header border, browser chart bars |
| `--green-light` | `#8bc34a` | Upload zone highlight, file label, status chart bars |
| `--green-pale` | `#e8f5e9` | Table row hover |
| `--bg` | `#f5f5f5` | Page background |
| `--card-bg` | `#ffffff` | Cards and panels |

## External Dependencies (vendored)

| File | Version | Size | Source |
|------|---------|------|--------|
| `vendor/d3.min.js` | 7.x | ~280 KB | D3.js |
| `vendor/topojson-client.min.js` | 3.x | ~7 KB | topojson-client |
| `data/countries-110m.json` | 2.x | ~108 KB | world-atlas (Natural Earth) |

All served from the local `static/` directory. No CDN requests.
