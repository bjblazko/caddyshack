# CaddyShack — arc42 Architecture Documentation

This directory contains the software architecture documentation for CaddyShack, structured according to the [arc42 template](https://arc42.org/).

## Table of Contents

| # | Section | Summary |
|---|---------|---------|
| 1 | [Introduction & Goals](01_introduction_goals.md) | Purpose, key requirements, quality goals, stakeholders |
| 2 | [Constraints](02_constraints.md) | Technical and organizational constraints |
| 3 | [Context & Scope](03_context_scope.md) | System boundaries, external interfaces |
| 4 | [Solution Strategy](04_solution_strategy.md) | Core architectural decisions and their rationale |
| 5 | [Building Block View](05_building_block_view.md) | Static decomposition into packages |
| 6 | [Runtime View](06_runtime_view.md) | Key interaction scenarios |
| 7 | [Deployment View](07_deployment_view.md) | Infrastructure and deployment options |
| 8 | [Crosscutting Concepts](08_crosscutting_concepts.md) | Privacy, streaming model, error handling |
| 9 | [Architecture Decisions](09_architecture_decisions.md) | Index of ADRs |
| 10 | [Quality Requirements](10_quality_requirements.md) | Quality tree and scenarios |
| 11 | [Risks & Technical Debt](11_risks_technical_debt.md) | Known risks and open items |
| 12 | [Glossary](12_glossary.md) | Domain terms |

## Architecture Decision Records

Individual ADRs are in [`adr/`](adr/).

## Related Documents

- [`../specs/`](../specs/) — current-state specs grouped by concern (parsing, analysis, security, ui, api, deployment)
- [`../glossary.md`](../glossary.md) — full domain glossary
- [`../features/`](../features/) — feature tracking (todo / in-progress / done)
