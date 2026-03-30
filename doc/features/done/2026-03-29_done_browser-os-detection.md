---
date: 2026-03-29
status: done
---

# Browser & OS Detection

User-Agent string parsing to identify browser and operating system for each log entry.

## Detection Logic (`internal/useragent`)

Order-dependent string matching — specific patterns checked before generic ones.

**OS order** (must precede more general checks):
1. `iPhone`/`iPad` → iOS (must come before macOS: iPhone UAs contain "Mac OS X")
2. `Windows` → Windows
3. `Macintosh`/`Mac OS X` → macOS
4. `Android` → Android
5. `CrOS` → ChromeOS (must come before Linux)
6. `Linux` → Linux
7. fallback → Other

**Browser order:**
1. `curl/` → curl
2. `bot`/`spider`/`crawl` (case-insensitive) → Bot
3. `Edg/` → Edge
4. `OPR/`/`Opera` → Opera
5. `Vivaldi/` → Vivaldi
6. `Brave` → Brave
7. `Chrome/` + `Safari/` → Chrome
8. `Safari/` without `Chrome/` → Safari
9. `Firefox/` → Firefox
10. fallback → Other

## Visualization

Horizontal bar charts rendered on `<canvas>` via `Charts.renderBarChart()`. Shows top 10 for each dimension, sorted by count descending.
