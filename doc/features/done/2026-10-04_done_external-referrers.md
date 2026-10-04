---
date: 2026-10-04
status: done
---

# External Referrers Only

The Top Referrers panel lists only external referrers. Referrers pointing to
any site found in the loaded log are internal navigation and are left out.

## Motivation

Most Referer headers come from visitors clicking between pages of the same
site (or between the operator's own sites). They pushed real traffic sources
out of the top 10.

## Implementation Notes

- "Own sites" are derived from the log itself: every canonical host that
  appears in the file, regardless of active filters. No host names are
  configured or hard-coded.
- A referrer is internal when the canonical host of its URL is one of them.
  Unparseable referrers count as external.
- Filtering happens before the top-10 cut, so the panel shows ten external
  referrers when there are that many.
- Single Events and the search filter still see every Referer value.
- Tables show an explicit empty state instead of a bare header.
