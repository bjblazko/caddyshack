# ADR-005: IP Anonymization Unconditionally Applied

**Date:** 2026-03-29
**Status:** Accepted

## Context

Caddy access logs contain real visitor IP addresses. Displaying these in a browser-based dashboard creates a privacy risk: the IP data could be captured in browser history, developer tools, local storage, or browser extensions.

GDPR and similar regulations treat IP addresses as personal data. An analytics tool that exposes raw IPs in its API response requires the operator to implement appropriate safeguards around that API.

Two design options were considered:
1. Expose raw IPs and let operators configure anonymization.
2. Anonymize IPs unconditionally before they leave the server.

## Decision

IP addresses are always anonymized in `internal/anonymize` before being included in any API response. There is no flag or configuration to disable this.

- IPv4: last octet zeroed (`93.184.216.34` → `93.184.216.0`)
- IPv6: only first 3 groups retained (`2a01:4f8:c17::1` → `2a01:4f8:c17::`)

GeoIP lookup is performed on the original IP before anonymization. Only the resulting country code is stored; the original IP is discarded.

## Consequences

**Positive:**
- Operators cannot accidentally expose raw IPs through misconfiguration
- Dashboard output is GDPR-compliant by construction
- Simplifies the privacy story: no settings, no opt-in/opt-out
- Raw IPs never appear in browser developer tools, network captures, or localStorage

**Negative:**
- Operators who want to see exact IPs for security or abuse investigation cannot use CaddyShack for that purpose — they must consult the raw log file directly
- Two visitors from the same /24 subnet appear under the same anonymized IP, inflating the "unique visitors" count marginally
