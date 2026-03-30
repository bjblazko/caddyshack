---
date: 2026-03-29
status: done
---

# Top Visitors Table

Lists the most active source IPs with anonymized addresses and country attribution.

## Data Per Visitor

| Field | Description |
|-------|-------------|
| `ip` | Anonymized IP address |
| `count` | Total requests from this IP |
| `country` | ISO 3166-1 alpha-2 country code |
| `country_name` | Human-readable country name |

## Rules

- GeoIP lookup performed on the **original** IP before anonymization
- Top 10 visitors by request count, sorted descending
- IPs anonymized before inclusion in the response (last IPv4 octet zeroed; IPv6 truncated to first 3 groups)
