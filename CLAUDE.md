# CaddyShack — Claude Instructions

## Glossary

Always consult `doc/glossary.md` when working with domain terms. If a new term is introduced during implementation, add it to the glossary.

## Feature Tracking

All features are tracked in `doc/features/` with three state folders:

- `doc/features/todo/` — planned, not started
- `doc/features/in-progress/` — currently being developed
- `doc/features/done/` — shipped

### File Convention

Filename: `YYYY-MM-DD_<status>_<slug>.md`
Example: `2026-03-29_done_ip-anonymization.md`

When a feature changes state, rename the file (update date and status prefix) and move it to the matching folder.

Each feature file must begin with this frontmatter:

```
---
date: YYYY-MM-DD
status: todo | in-progress | done
---
```

- When planning a new feature: create the file in `todo/`
- When work starts: move to `in-progress/`, update prefix date and status
- When done: move to `done/`, update prefix, then update the relevant spec(s) in `doc/specs/`

## Architecture & ADRs

Before making any structural or design change, consult:

- `doc/arch/` — arc42 architecture documentation
- `doc/arch/adr/` — Architecture Decision Records

Check whether an ADR already covers the area you are changing. If a decision is being revisited or a new architectural choice is made, create a new ADR in `doc/arch/adr/` before or alongside the implementation.

## Specs

`doc/specs/` holds the authoritative current-state description of the software, grouped by concern. Consult the relevant spec(s) before making changes to understand the current design. Update the relevant spec after each feature ships. Specs describe *what is*, not history.

| File | Covers |
|------|--------|
| `parsing.md` | JSONL log format, logparser package, Caddy configuration |
| `analysis.md` | Aggregation engine, metrics, filtering rules, report structure |
| `security.md` | IP anonymization, GeoIP, privacy model |
| `ui.md` | Frontend layout, charts, world map, color scheme |
| `api.md` | HTTP endpoints, request/response contracts |
| `deployment.md` | Single binary, CLI flags, Docker, health check |

## Diagrams

Always use Mermaid for diagrams. Never use ASCII box-drawing characters. Wrap diagrams in fenced code blocks with the `mermaid` language tag:

~~~
```mermaid
graph TD
    ...
```
~~~

Choose the diagram type that best fits the content:

| Content | Mermaid type |
|---------|-------------|
| Component / package relationships | `graph TD` or `graph LR` |
| Request / interaction flows | `sequenceDiagram` |
| Process / decision flows | `flowchart TD` |
| Deployment topology | `graph TB` |
| Hierarchical quality or feature trees | `mindmap` |
