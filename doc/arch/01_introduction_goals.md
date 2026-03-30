# 1. Introduction and Goals

## What Is CaddyShack?

CaddyShack is a self-hosted web analytics tool for [Caddy](https://caddyserver.com/) web server operators. It parses Caddy's structured JSONL access logs and renders an interactive dashboard showing traffic patterns, geographic distribution, browser/OS breakdown, and privacy-compliant visitor statistics — without sending data to any third party.

## Key Functional Requirements

| ID | Requirement |
|----|-------------|
| F-01 | Accept a Caddy JSONL access log file via browser upload or server-side path |
| F-02 | Parse and aggregate log entries in a single streaming pass |
| F-03 | Produce per-host and aggregate reports when the log contains multiple virtual hosts |
| F-04 | Segment traffic into All / Success (2xx) / Error (4xx+) views |
| F-05 | Display geographic request distribution via a world map with GeoIP resolution |
| F-06 | Anonymize all IP addresses before returning data to the client |
| F-07 | Detect browser and operating system from User-Agent strings |
| F-08 | Require no installation, configuration file, or external service to run |

## Quality Goals

Listed in descending priority:

| Priority | Quality Goal | Motivation |
|----------|-------------|------------|
| 1 | **Privacy** | IP addresses must not leave the server in raw form; GDPR compliance is non-negotiable |
| 2 | **Simplicity of deployment** | Single binary, no database, no runtime dependencies — must run with `./caddyshack` |
| 3 | **Memory efficiency** | Must handle large log files without loading them fully into memory |
| 4 | **Offline operation** | No CDN calls; works in air-gapped or firewalled environments |
| 5 | **Maintainability** | No framework magic; standard-library code that any Go developer can read |

## Stakeholders

| Role | Concern |
|------|---------|
| Caddy server operator | Wants traffic insight without sending logs to a third-party SaaS |
| Privacy-conscious site owner | Needs GDPR-compliant analytics |
| Self-hoster / homelab user | Wants zero-dependency, single-binary deployment |
| Open-source contributor | Needs readable, framework-free Go and JS code to contribute to |
