---
date: 2026-03-29
status: done
---

# Docker & GHCR Support

Containerized deployment via a multi-stage Dockerfile and Docker Compose, with images published to GitHub Container Registry.

## Dockerfile

Multi-stage build:
1. **Builder stage** — compiles the Go binary
2. **Runtime stage** — copies only the binary into a minimal base image

## Docker Compose

`compose.yml` defines the service with port mapping and optional volume mount for the GeoIP CSV.

## GHCR Publishing

GitHub Actions release workflow builds and pushes multi-arch images (`linux/amd64`, `linux/arm64`) to `ghcr.io/bjblazko/caddyshack` on every tagged release.

## Usage

```sh
docker run -p 8080:8080 ghcr.io/bjblazko/caddyshack:latest
```
