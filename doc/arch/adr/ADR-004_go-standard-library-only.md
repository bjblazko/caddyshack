# ADR-004: Go Standard Library Only — No External Modules

**Date:** 2026-03-29
**Status:** Accepted

## Context

Go's ecosystem has popular web frameworks (Gin, Echo, Chi) and utility libraries (gorilla/mux, zerolog, etc.) that reduce boilerplate. However, every external module adds transitive dependencies, increases `go.sum` complexity, and creates a surface area for supply-chain attacks or breaking changes.

CaddyShack's HTTP surface is small: four endpoints, multipart form parsing, JSON encoding, and static file serving — all of which are well-covered by the standard library.

## Decision

The Go backend uses only packages from the Go standard library. No external modules are added to `go.mod` beyond the Go standard library itself.

Go 1.22's enhanced `http.ServeMux` (method-qualified patterns like `"POST /api/upload"`) eliminates the main use case for a router library.

## Consequences

**Positive:**
- `go build ./...` works with no network access and no `go mod download`
- No transitive dependency vulnerabilities or supply-chain risk
- `go.mod` and `go.sum` stay minimal and easy to audit
- Code is idiomatic Go that any Go developer can read without learning a framework

**Negative:**
- Slightly more verbose handler code compared to a framework (no middleware chains, no parameter extraction helpers)
- Requires Go 1.22+ for method-qualified routing; older Go versions are not supported
- Some convenience features (structured logging, request context propagation) require manual implementation if needed in the future
