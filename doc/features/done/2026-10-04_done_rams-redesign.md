---
date: 2026-10-04
status: done
---

# Rams Redesign

The interface follows huepattl-rams-design: neutral surfaces with light and
dark mode, IBM Plex, flat panels, one primary action, ranked bars with fixed
entity colors instead of donuts, one filter summary instead of per-panel
badges, honest states (empty, loading, error, no GeoIP). See `doc/specs/ui.md`
for the design and its one recorded deviation (chart colors).

## Implementation Notes

- `static/css/tokens.css` copied unchanged from the design system
- Chart palette validated with the dataviz validator, light and dark
- All exclusions (static files, images, monitors, bots) on by default
- `-logdir` flag added for previews and non-default log locations
- Fixed: `/api/logs` returned `null` for an empty directory
