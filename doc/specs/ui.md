# Spec: UI & Visuals

## Design System

The UI follows **huepattl-rams-design** (Dieter Rams' principles applied to
software): neutral warm surfaces, color only as a signal, flat surfaces
separated by hairlines, one type family, one spacing scale.

- `static/css/tokens.css` — the design system's reference tokens, copied
  unchanged. Light and dark mode follow the OS (`prefers-color-scheme`).
- `static/css/style.css` — layout and components; uses only `var(--…)`
  values from the tokens, plus the chart colors below.
- Fonts: IBM Plex Sans (400/500/600) for UI and text, IBM Plex Mono (400/500)
  for numbers, IPs and timestamps, with tabular numerals. IBM's own Latin-1
  WOFF2 subsets, self-hosted in `static/fonts/` with `OFL.txt`.

**Primary job of the screen:** see what traffic a Caddy log contains — how
much, from where, for what — narrowed by filters. The only element in the
accent color is **Upload log file**.

### Deviations from huepattl-rams-design

| Deviation | Reason |
|---|---|
| Charts use categorical colors (`--series-1` … `--series-7`) | Browsers and operating systems must be told apart at a glance; neutral or single-hue bars were not distinguishable. Slots come from the validated dataviz reference palette (light and dark checked with its validator against `--bg`; all hard gates pass). Used only inside charts, always with a visible label. Slot 8 (red) is omitted so red keeps meaning "needs attention". |

### Signal colors in charts

| Color | Meaning |
|---|---|
| `--series-N` per entity | Fixed per browser/OS name (Chrome blue, Firefox orange, Safari aqua, Brave yellow, Opera magenta, Vivaldi green, Edge violet; Windows blue, macOS aqua, Linux yellow, ChromeOS magenta, Android green, iOS violet), so a filter that changes the ranking never repaints an entity |
| `--fg-3` (grey) | Not a browser: Bot, Monitor, Script, Other |
| `--series-1` (blue) | Magnitude: daily traffic bars, map bubbles |
| `--confirm` | 2xx status codes |
| `--warning` | 4xx and 5xx status codes (always next to the code as label) |

## Technology

- Vanilla HTML5, CSS3, JavaScript (ES2020+) — no framework, no build step, no npm
- Canvas 2D API for bar charts (browsers, operating systems, status codes, daily traffic)
- D3.js v7 for the world map only (served locally, no CDN)
- topojson-client v3 for country boundary rendering (served locally, no CDN)
- Natural Earth 110m TopoJSON for country boundaries

## Layout

Single-page application (`index.html`).

1. **Header** — logo, name, "Log file" select (server logs and the uploaded
   file), **About** (quiet button), **Upload log file** (primary button), a
   status line ("Analyzing …") while a request runs. A file can also be dropped
   anywhere on the page.
2. **Message** — errors in plain words below the header (`role="alert"`),
   replacing browser `alert()` dialogs.
3. **Empty state** — "No log file loaded" with one sentence on what to do,
   shown when no server log is readable and nothing was uploaded.
4. **Filters** — labels above controls, all 36 px high; status as a segment
   group (`aria-pressed`), see table below.
5. **Filter summary** — one sentence for all views, e.g. "Showing example.com ·
   success (2xx), without static files, images, monitors and bots."
6. **View tabs** — links `#statistics` and `#events` (back button and deep
   links work), current view marked with `aria-current="page"`.
7. **Statistics** — key figures (requests, unique IPs, transferred, average
   response; 4 columns, 2 on narrow screens), then panels in two columns:
   world map + countries, browsers + operating systems, daily traffic (full
   width), status codes + top URIs, top visitors + top external referrers.
8. **Single events** — filtered log entries newest first, loaded 100 at a time
   while scrolling; rows with status ≥ 400 have a warning tint and the status in
   `--warning-ink`.
9. **Footer** — see below.

Panels are flat with a 1 px `--line` border and `--radius-md`; nothing has a
shadow except the floating tooltip.

### States

| State | Shown as |
|---|---|
| No file | Empty state |
| Loading | Status line in the header names the file; dashboard at 50 % opacity |
| Error | Message bar: what failed and what to do |
| Empty table | One muted line: "No data for the current filters." (referrers: "No external referrers for the current filters.") |
| Empty chart | The same sentence in place of the canvas |
| No GeoIP database | Countries read "Unknown"; a note under the map names the `-geodb` flag |

### Footer

Always visible, muted small text: "no telemetry, no external requests", the
license with the no-warranty statement, a link to **Licenses and thanks**
(`/licenses.html`), and the
Caddy trademark note. When any country is resolved, it starts with "IP
geolocation by DB-IP" linking to db-ip.com, as the DB-IP Lite license
(CC BY 4.0) requires.

### About dialog

Native `<dialog>` (`#about`), opened by **About** or the URL `#about`; Esc,
Close and a click on the backdrop close it. Content: name and version (from
`/api/health`), what CaddyShack does, *Made by* Timo Böwing with links to the
product page on huepattl.de, the GitHub repository and huepattl.de, *Your
data* (no outbound requests, uploads deleted after the idle time, IPs
shortened), *License* (Apache-2.0, no warranty, link to Licenses and thanks).
The only element with a shadow besides the tooltip; backdrop `--scrim`.

### Licenses and thanks

`/licenses.html` renders `static/licenses/credits.json` (`js/licenses.js`): an
intro with CaddyShack's own license and the no-warranty sentence, then
*Built into CaddyShack* and *Data*, each entry with name and version, what it
does here, its license, and links *Website* and *License text* (and *Support
the project* where one exists — none of the current projects has one). The
Caddy trademark note closes the page.

### Responsive

| Viewport | Layout |
|----------|--------|
| > 960 px | Two-column panels, four key figures in a row |
| ≤ 960 px | One column of panels |
| ≤ 640 px | Full-width fields stacked, two key figures per row, 16 px gutter; no horizontal scrolling |

Charts redraw on window resize and when the color scheme changes.

## Filters

| Control | Element | State variable | Effect |
|---------|---------|----------------|--------|
| Site | `<select id="host-select">` | `currentHost` | One canonical host; `""` = all |
| Status | Segment group `.status-btn` | `currentStatus` | `all` / `success` / `error` |
| Date range | `#date-start`, `#date-end`, clear button `#date-clear` | `currentDateStart`, `currentDateEnd` | Inclusive bounds (`YYYY-MM-DD`) |
| Country | `<select id="country-filter">` | `currentCountry` | Country name exact match |
| Browser | `<select id="browser-filter">` | `currentBrowser` | Browser name exact match |
| Operating system | `<select id="os-filter">` | `currentOS` | OS name exact match |
| URI | `<select id="page-filter">` | `currentPage` | URI exact match |
| Method | `<select id="method-filter">` | `currentMethod` | HTTP method exact match |
| Search | `<input id="search-filter">` | `currentSearch` | Glob over URI, IP, referrer (400 ms debounce); placeholder shows examples |
| Exclude: Static files | `#ignore-static` | `ignoreStatic` | **On** by default; re-set on every new file |
| Exclude: Images | `#ignore-images` | `ignoreImages` | **On** by default; re-set on every new file |
| Exclude: Monitors | `#ignore-monitors` | `ignoreMonitors` | **On** by default; re-set on every new file |
| Exclude: Bots | `#ignore-bots` | `ignoreBots` | **On** by default; re-set on every new file |

Scripts (curl, scanners …) deliberately have no exclusion so probing requests
stay visible. The upload request carries the same filter params, so the first
view already honours the defaults.

Every filter change triggers `doFetch()`, which sends all active filter params
to `GET /api/analyze` and re-renders the dashboard from the backend response.

### Dimension Dropdown Repopulation

After each response, `populateDimensionDropdowns(report)` rebuilds the
dimension selects from the report's `countries`, `browsers`,
`operating_systems`, `top_pages` and `methods`. If the previously selected
value is no longer present, the select resets to "All" and the state variable
is cleared.

## JavaScript Modules

### `app.js`

File loading (upload, drop, server log select), filter state and query
building, `GET /api/analyze` and `/api/events`, rendering of key figures and
tables (`renderTable` right-aligns columns whose header has class `num`),
charts via `renderCharts`, the filter summary, view switching by URL hash,
status line and messages. No client-side re-aggregation.

### `charts.js` — `Charts` and `Tooltip`

- `Charts.renderBarChart(canvasId, labels, values, total, colorFor)` — ranked
  horizontal bars: label, 12 px bar with 4 px rounded end on a `--bg-2` track,
  count and share in mono; `colorFor(label)` picks the bar color
- `Charts.renderVerticalBarChart(canvasId, labels, values)` — requests per day,
  recessive grid, date labels thinned to avoid collisions, hover tooltip per day
- `Charts.entityColor`, `Charts.statusColor` — color rules above
- Colors are read from CSS custom properties at draw time (`Charts.token`)
- `Tooltip.show(event, text)` / `Tooltip.hide()` — one shared floating tooltip

### `map.js` — `WorldMap`

- Natural Earth projection; land in `--bg-2`, borders and outline in `--line`
- Bubbles at country centroids, area-proportional (`d3.scaleSqrt`), blue with a
  surface-colored ring so overlapping bubbles stay separable; hover tooltip
- Size legend bottom left (over the Pacific)

## File Upload

- "Upload log file" opens the picker; dropping a file anywhere also uploads it
- Sent as `multipart/form-data` (field `logfile`) to `POST /api/upload` with
  the current filter params; the returned `file_id` is used for later requests
- An expired upload shows the server's message asking to upload again

## Vendored Assets

All served locally from `static/` — no external requests at runtime; license
texts linked from `/licenses.html`.

| File | Version |
|------|---------|
| `vendor/d3.min.js` | 7.9.0 |
| `vendor/topojson-client.min.js` | 3.1.0 |
| `data/countries-110m.json` | world-atlas 2.0.2 (Natural Earth 110m) |
| `fonts/IBMPlex*-Latin1.woff2` | IBM Plex Sans 1.1.0, IBM Plex Mono 2.5.0 |
