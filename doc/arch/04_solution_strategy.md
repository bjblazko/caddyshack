# 4. Solution Strategy

The following table maps each top-priority quality goal to the architectural approach chosen to achieve it.

| Quality Goal | Approach | ADR |
|---|---|---|
| Privacy | Anonymize IPs server-side before the response is sent; perform GeoIP on the original IP, never expose it | [ADR-005](adr/ADR-005_ip-anonymization-by-default.md) |
| Deployment simplicity | Single binary with embedded assets; no database; no configuration file required | [ADR-002](adr/ADR-002_single-binary-embedded-assets.md) |
| Memory efficiency | Stream log file line-by-line; accumulate only counters, not raw entries | [ADR-008](adr/ADR-008_single-pass-streaming-analysis.md) |
| Offline operation | Vendor all JS dependencies into the binary; no CDN or external fetch at runtime | [ADR-002](adr/ADR-002_single-binary-embedded-assets.md) |
| Maintainability | Go standard library only; vanilla JS; no framework abstractions to learn | [ADR-003](adr/ADR-003_vanilla-js-no-framework.md), [ADR-004](adr/ADR-004_go-standard-library-only.md) |

## Fundamental Decisions

1. **Stateless, request-scoped analysis.** Each upload request is self-contained. There is no server state between requests. This eliminates the entire persistence layer and makes the system trivially deployable and horizontally scalable. See [ADR-001](adr/ADR-001_stateless-request-scoped-analysis.md).

2. **Single-pass streaming.** The log file is scanned line-by-line. All dimensions (IPs, browsers, countries, etc.) are aggregated into maps in one pass. No second pass, no intermediate storage. Memory grows with unique values, not log size. See [ADR-008](adr/ADR-008_single-pass-streaming-analysis.md).

3. **Privacy by default, unconditionally.** IP anonymization is not a setting — it is always applied. This removes any risk of misconfiguration exposing personal data. See [ADR-005](adr/ADR-005_ip-anonymization-by-default.md).

4. **GeoIP is optional.** The GeoIP database is loaded at startup from a configurable path. If absent, the application runs without country data. This avoids bundling a third-party dataset and keeps the binary licence-clean. See [ADR-006](adr/ADR-006_optional-geoip-graceful-degradation.md).

5. **No external dependencies (Go or JS).** The Go backend uses only `net/http` and the standard library. The frontend uses only vanilla JS plus D3.js (vendored). This minimizes supply-chain risk and ensures the project builds with just `go build`. See [ADR-003](adr/ADR-003_vanilla-js-no-framework.md) and [ADR-004](adr/ADR-004_go-standard-library-only.md).
