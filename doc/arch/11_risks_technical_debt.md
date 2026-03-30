# 11. Risks and Technical Debt

## Risks

### R-01: No Authentication

**Description:** CaddyShack has no authentication mechanism. Anyone who can reach port 8080 can upload log files and view analytics.

**Impact:** On a publicly exposed server, an attacker could enumerate log file names via `/api/logs`, analyze server-side logs, or perform denial-of-service by uploading large files repeatedly.

**Mitigation:** Run CaddyShack on a non-public port and protect it with Caddy's `basicauth` directive or a firewall rule. Do not expose it directly to the internet.

---

### R-02: GeoIP Database Not Bundled

**Description:** The DB-IP Lite CSV is not included in the binary due to licence restrictions. Operators must download and configure it manually.

**Impact:** Without the database, all country fields return `??` and the world map shows no data. This degrades the user experience but does not affect other analytics.

**Mitigation:** README documents the download and `-geodb` flag clearly. The graceful degradation means the app remains fully usable without it.

---

### R-03: IPv6 GeoIP Not Supported

**Description:** The GeoIP lookup only handles IPv4. IPv6 addresses always return `??`.

**Impact:** Sites with significant IPv6 traffic will see an inflated `??` country entry.

**Mitigation:** Accept as a known limitation for now. A future ADR could evaluate extending `internal/geoip` to support IPv6 ranges.

---

### R-04: Memory Spike on Highly Diverse Logs

**Description:** Memory is bounded by unique values, not line count. A log with tens of millions of unique IPs (e.g., from a DDoS event) could consume significant RAM.

**Impact:** Potential OOM on resource-constrained hosts for adversarial input.

**Mitigation:** The 500 MB `MaxBytesReader` limits the file size. For a normal log file, the number of unique IPs is bounded by the number of real visitors. No mitigation beyond the upload limit exists today.

---

### R-05: Uploaded Files Persist in OS Temp Directory

**Description:** Files uploaded via `POST /api/upload` are saved to `os.TempDir()/caddyshack/<hex_id>.jsonl` and are never explicitly deleted by the application. They persist for the lifetime of the process (or until the OS clears the temp directory).

**Impact:** Disk space accumulates if many large files are uploaded. On shared systems, temp files from the current session remain readable by other processes with filesystem access.

**Mitigation:** This is acceptable for single-user and small-team deployments. The OS temp directory is typically cleaned on reboot. For long-running deployments, operators can add a periodic cleanup cron for `os.TempDir()/caddyshack/`. A future improvement could add a TTL-based cleanup goroutine at startup.

---

## Technical Debt

### TD-01: No Automated Tests

There are currently no Go unit tests or integration tests. The parser, anonymizer, GeoIP, and analyzer packages all have clear pure-function APIs that would be straightforward to test.

**Risk:** Regressions in core logic (filtering rules, anonymization, user-agent detection order) may go undetected.

**Suggested action:** Add table-driven tests for `internal/logparser`, `internal/anonymize`, `internal/useragent`, and `internal/analyzer`.

---

### TD-02: Hardcoded Asset Prefixes and Extensions

The asset filtering logic in `internal/analyzer` (prefixes: `/css/`, `/js/`, etc.; extensions: `.css`, `.js`, etc.) is hardcoded. Sites with non-standard asset paths will see static assets appear in "Top Pages".

**Suggested action:** Consider making the filter configurable via a flag or a simple config file.

---

### TD-03: Country Centroid Coordinates Hardcoded in `map.js`

The lat/lon centroid for each country used in the bubble map is a static lookup table in `map.js`. Adding or correcting a centroid requires a code change.

**Suggested action:** Low priority; Natural Earth centroids are stable. Accept as-is unless a correctness issue is reported.

---

### TD-04: No Streaming to Client

The analysis must complete before any response is sent. For very large log files, the browser shows a loading overlay with no progress indication. With the backend filter-then-aggregate model, every filter change re-triggers a full analysis pass.

**Suggested action:** Consider server-sent events or chunked JSON to report progress for large files. Low priority given the 500 MB upload cap and sub-second analysis time for typical log sizes.
