---
date: 2026-03-29
status: done
---

# Dashboard Summary Cards

Four top-level metric cards shown at the top of the dashboard after a log file is loaded.

## Metrics

| Card | Source field | Notes |
|------|-------------|-------|
| Total Requests | count of parsed entries | — |
| Unique IPs | distinct anonymized IPs | after anonymization |
| Data Transferred | sum of `size` fields | formatted as human-readable bytes |
| Avg Response Time | `(sum of duration) / count × 1000` | result in milliseconds |

## Layout

4-column grid on desktop (>768px), 2-column on mobile. Each card has a top border in `--green-mid` and the value text in `--green-dark`.
