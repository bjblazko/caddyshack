---
date: 2026-10-04
status: done
---

# Client Types: Hide Monitors and Bots

Every request is classified by its User-Agent as **Browser**, **Monitor**,
**Bot** or **Script**. Monitors and bots are excluded by default via two
"Exclude" checkboxes; unticking them shows that traffic again.

## Motivation

Uptime monitors poll around the clock and crawlers, link previews and feed
readers fetch pages no person reads. Both inflated the statistics, and most
of them were counted as a normal browser or "Other". Probing requests (PHP,
`/wp-admin`, `.env` …) are of interest and must stay visible.

## Implementation Notes

- Classification in `internal/useragent` from built-in, case-insensitive
  substring rules; nothing to configure. Order: Monitor, Bot, Script, Browser.
- Monitor: Uptime Kuma, UptimeRobot, Pingdom, StatusCake, Better Stack, …
- Bot: search engines, AI and SEO crawlers, link previews, feed readers
- Script: curl, wget, HTTP libraries, scanners (zgrab, Nuclei, …), empty UA.
  Scripts have no exclude checkbox, so scanner probes stay visible.
- For non-browser clients the browser name is the client type, so the
  Browsers chart and the Single Events "Browser" column show Monitor / Bot /
  Script instead of "Other".
- API: `ignore_monitors=1`, `ignore_bots=1` on `/api/upload`, `/api/analyze`
  and `/api/events`. Without them nothing is excluded; the UI sends both by default.
- Heuristic: a client that claims to be a browser is counted as one.
