---
date: 2026-03-29
status: done
---

# World Map with Geographic Distribution

D3.js bubble map showing request volume by country of origin.

## Implementation

- Country boundaries rendered using Natural Earth 110m TopoJSON (`/data/countries-110m.json`)
- Projection: `d3.geoNaturalEarth1()` fitted to the container size
- Graticule (lat/lon grid) drawn as a reference layer
- One bubble per country at the country's centroid; radius scaled with `d3.scaleSqrt()` for area-proportional representation
- Hover tooltip shows country name and request count, positioned via `clientX/Y`
- Legend with reference bubble sizes in the bottom-right corner
- Color: semi-transparent green fill (`rgba(74, 140, 63, 0.55)`), dark green stroke

## Data Source

Country codes from GeoIP lookup. Countries with `??` code are excluded from the map. Country centroids are hardcoded in `map.js`.

## Offline

D3.js and TopoJSON client served from `/vendor/`; no CDN calls.
