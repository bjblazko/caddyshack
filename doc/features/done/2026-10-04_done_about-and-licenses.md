---
date: 2026-10-04
status: done
---

# About and Licenses and Thanks

An About dialog says who makes CaddyShack and links the product page, the
GitHub repository and huepattl.de. A "Licenses and thanks" page lists every
project CaddyShack is built on, with license, website and license text —
the pattern Unterlumen uses, now also described in the huepattl-legal-check
skill.

## Implementation Notes

- `static/licenses/credits.json` + one license text per project, embedded
- `main_test.go` keeps the list complete (shipped files, embedded texts)
- Support links only where verified; none of the current projects has one
- Version stamped via `-ldflags -X main.version`, reported by `/api/health`
