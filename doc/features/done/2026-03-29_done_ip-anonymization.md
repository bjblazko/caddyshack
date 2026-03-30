---
date: 2026-03-29
status: done
---

# GDPR-Compliant IP Anonymization

All IP addresses are truncated before inclusion in the API response or UI.

## Rules (`internal/anonymize`)

| IP version | Method | Example |
|------------|--------|---------|
| IPv4 | Zero the last octet | `93.184.216.34` → `93.184.216.0` |
| IPv6 | Keep first 3 groups, zero the rest | `2a01:4f8:c17::1` → `2a01:4f8:c17::` |

## Processing Order

1. GeoIP lookup is performed on the **original** IP
2. IP is then anonymized
3. Only the anonymized IP appears in the `Report` (top visitors, unique IP count)

Raw IP addresses never leave the server process.
