# 2. Constraints

## Technical Constraints

| Constraint | Detail |
|------------|--------|
| Go standard library only | No external Go module dependencies beyond the standard library. Keeps the dependency surface minimal and the build reproducible. |
| Go 1.22+ | Required for method-based route registration in `http.ServeMux` (`mux.HandleFunc("POST /api/upload", ...)`). |
| No frontend framework | Vanilla HTML5 / CSS3 / JavaScript only. No React, Vue, Angular, npm, or build step. |
| No CDN at runtime | All frontend assets (D3.js, TopoJSON client, country boundary data) must be served locally from the binary. |
| No database | No SQL, key-value store, or any form of server-side persistence. |
| Input format | Only Caddy v2 JSONL access logs are supported. Other log formats are out of scope. |
| GeoIP database not bundled | The DB-IP Lite CSV is not distributed with the binary (licence terms). Operators must supply it separately. |

## Organizational Constraints

| Constraint | Detail |
|------------|--------|
| Open source | MIT licence. All design decisions must be explainable and reproducible without proprietary tools. |
| Privacy by default | GDPR compliance is a hard requirement, not a configuration option. IP anonymization cannot be disabled. |
| Single maintainer (initially) | Architecture must stay simple enough for a single person to reason about and maintain. |

## Conventions

| Convention | Detail |
|------------|--------|
| Semantic versioning | `vMAJOR.MINOR.PATCH`. Patch = bug fixes; minor = new features; major = breaking API change. |
| Release automation | Tagging `vX.Y.Z` triggers GitHub Actions to build binaries for 5 platforms and push Docker images to GHCR. |
| Feature tracking | New features planned in `doc/features/todo/`, moved through `in-progress/` to `done/` on completion. |
