---
date: 2026-03-29
status: done
---

# GeoIP Country Resolution

Maps source IP addresses to ISO 3166-1 alpha-2 country codes using the DB-IP Lite CSV database.

## Implementation (`internal/geoip`)

- `Load(path)` reads the CSV at startup into sorted `uint32` slices for binary search
- Only IPv4 ranges are loaded; IPv6 always returns `"??"`
- `Lookup(ip)` converts the IP to `uint32` and binary-searches the range table
- Returns `"??"` for unrecognized IPs or if no database is loaded (app continues without GeoIP data)
- `CountryName(code)` returns the display name (e.g. `"DE"` → `"Germany"`); falls back to the code itself

## Database

[DB-IP Lite](https://db-ip.com/db/download/ip-to-country-lite) free CSV. Format: `start_ip,end_ip,country_code`. Not shipped with the binary; path configurable via `-geodb` CLI flag.

## Privacy

Country code is the only GeoIP-derived field in the API response. The original IP is never sent to the client.
