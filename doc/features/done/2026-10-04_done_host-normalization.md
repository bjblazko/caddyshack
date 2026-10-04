---
date: 2026-10-04
status: done
---

# Host Normalization

Requests for the same site are grouped under one host even when the `Host`
header carries the scheme's default port or different letter case, e.g.
`huepattl.de` and `huepattl.de:443` appear as one entry in the Site dropdown.

## Motivation

Caddy logs the `Host` header verbatim. Some clients (notably HTTP/2 and
HTTP/3 clients) send the authority with an explicit `:443`, so every site
appeared twice in the Site dropdown and its traffic was split between both.

## Implementation Notes

- Canonical host = lower-case host name without the default ports `:443` and `:80`
- Non-default ports (e.g. `:8443`) stay separate sites
- Applied once per entry during enrichment, so the host list, the host filter,
  the statistics and the Single Events view all use the canonical form
- The `host` query parameter is normalized the same way, so links with `:443` keep working
