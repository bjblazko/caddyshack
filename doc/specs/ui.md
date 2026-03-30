# Spec: UI & Visuals

## Technology

- Vanilla HTML5, CSS3, JavaScript (ES2020+) — no framework, no build step, no npm
- Canvas 2D API for bar charts
- D3.js v7 + topojson-client v3 for the world map (served locally, no CDN)
- Natural Earth 110m TopoJSON for country boundaries

## Dashboard Layout

Single-page application (`index.html`). All sections hidden until a log file is loaded.

### Sections (top to bottom)

1. **Header** — logo, title, subtitle, drag-and-drop upload zone
2. **Filter bar** — host dropdown, Success/Error/All toggle, date range inputs, Country/Browser/OS/Page dropdowns
3. **Summary Cards** (4-column desktop / 2-column mobile) — total requests, unique IPs, data transferred, avg response time
4. **World Map + Countries** (2-column) — D3 bubble map left, country table right
5. **Browsers + Operating Systems** (2-column) — horizontal bar charts on `<canvas>`
6. **Daily Traffic** (full-width) — vertical bar chart on `<canvas>`
7. **Status Codes + Top Pages** (2-column) — bar chart and table
8. **Top Visitors** (full-width) — table with anonymized IPs, country, count

### Responsive Breakpoint

| Viewport | Layout |
|----------|--------|
| > 768px | 2-column grids, 4-column card row |
| ≤ 768px | Single-column, 2-column card row |

## Filter Bar

All filter controls are rendered in a single `.filter-bar` row above the summary cards.

| Control | Element | State variable | Effect |
|---------|---------|----------------|--------|
| Host | `<select id="host-select">` | `currentHost` | Scopes to one virtual host; `""` = all hosts |
| Status | Toggle buttons (All / Success / Error) | `currentStatus` | `""` / `"success"` / `"error"` |
| Start date | `<input type="date" id="date-start">` | `currentDateStart` | Inclusive lower bound (`YYYY-MM-DD`) |
| End date | `<input type="date" id="date-end">` | `currentDateEnd` | Inclusive upper bound (`YYYY-MM-DD`) |
| Date clear | Button `#date-clear` | — | Resets both date inputs |
| Country | `<select id="country-filter" class="dim-filter-select">` | `currentCountry` | Country name exact match |
| Browser | `<select id="browser-filter" class="dim-filter-select">` | `currentBrowser` | Browser name exact match |
| OS | `<select id="os-filter" class="dim-filter-select">` | `currentOS` | OS name exact match |
| Page | `<select id="page-filter" class="dim-filter-select">` | `currentPage` | URI exact match |

Every filter change triggers `doFetch()`, which sends all active filter params to `GET /api/analyze` and re-renders the entire dashboard from the backend response. Date inputs use a 400ms debounce (`scheduleFetch`) to avoid rapid requests while typing.

### Dimension Dropdown Repopulation

After each `GET /api/analyze` response, `populateDimensionDropdowns(report)` rebuilds the four dimension selects from the returned report's `countries`, `browsers`, `operating_systems`, and `top_pages` arrays. These arrays already reflect the current host, status, and date filters, so the dimension dropdowns never show stale or out-of-range values. If the previously selected value is no longer present, the dropdown resets to "All" and the state variable is cleared.

## Filter Hints

Every dashboard panel has a `<div class="filter-hint">` element. After each render, `updateFilterHints()` inspects the active filter state and writes a pill badge per panel:

- **Active filter** → green pill (`.filter-tag--active`) showing the active filter value
- **No filter** → muted italic text (`.filter-tag--none`) reading "all data"

## JavaScript Modules

### `app.js`

Main orchestrator. Handles:
- Drag-and-drop and file input events
- `FormData` construction and `POST /api/upload`; stores returned `file_id` as `fileRef`
- `GET /api/analyze` with all active filter params via `buildQuery()` / `doFetch()`
- 400ms debounce (`scheduleFetch`) for date inputs
- Loading overlay during fetch
- DOM population of tables and section visibility via `renderDashboard(data)`
- Host dropdown population (`populateHostDropdown`) and dimension dropdown population (`populateDimensionDropdowns`)
- Filter hint rendering (`updateFilterHints`)
- Full filter reset on new file load (`resetFilters`)

No client-side re-aggregation is performed. The backend always returns a fully aggregated `AnalysisResult` for the current filter combination.

### `charts.js` — `Charts` namespace

- `Charts.renderBarChart(canvasId, labels, values, total, color)` — horizontal bar chart with percentage labels; DPR-scaled for Retina displays
- `Charts.renderVerticalBarChart(canvasId, labels, values)` — vertical bar chart with Y-axis gridlines and rotated date labels; DPR-scaled

### `map.js` — `WorldMap` namespace

- `WorldMap.render(containerId, countries)` — fetches `/data/countries-110m.json`, renders country outlines with D3 Natural Earth projection, overlays proportional bubbles at country centroids
- Bubble sizing: `d3.scaleSqrt()` for area-proportional representation
- Hover tooltip: country name + request count, positioned at `clientX/Y`
- Legend with reference bubble sizes (bottom-right)
- Graticule (lat/lon grid) as a reference layer

## Color Scheme

| CSS Variable | Hex | Usage |
|---|---|---|
| `--green-dark` | `#2d5a27` | Header background, headings, card value text |
| `--green-mid` | `#4a8c3f` | Card top border, table header border, browser chart bars |
| `--green-light` | `#8bc34a` | Upload zone highlight, file label, status chart bars |
| `--green-pale` | `#e8f5e9` | Table row hover, active filter tag background |
| `--bg` | `#f5f5f5` | Page background |
| `--card-bg` | `#ffffff` | Cards and panels |

Map bubble fill: `rgba(74, 140, 63, 0.55)` with dark green stroke.

## File Upload

- Drag-and-drop onto the upload zone or click to open the file picker
- File converted to `FormData` (field name: `logfile`) and POSTed as `multipart/form-data`
- On success the returned `file_id` is stored; subsequent filter changes re-analyze the same temp file via `GET /api/analyze?file=<id>`
- Loading overlay shown during analysis; hidden on response

## Vendored Assets

All served locally from `static/` — no external requests at runtime.

| File | Version |
|------|---------|
| `vendor/d3.min.js` | 7.x |
| `vendor/topojson-client.min.js` | 3.x |
| `data/countries-110m.json` | Natural Earth 110m |
