# ADR-008: Single-Pass Streaming Log Analysis

**Date:** 2026-03-29
**Status:** Accepted

## Context

Log files can be large — tens or hundreds of megabytes containing millions of entries. Approaches to analysis:

1. **Load entire file into memory, then analyze** — simple to implement but memory usage scales linearly with file size. A 500 MB log would require at least 500 MB of RAM plus additional memory for parsed structures.
2. **Multi-pass streaming** — read through the file multiple times, each pass computing different metrics. Memory-efficient per pass but requires either a seekable stream (not always available) or writing a temp file.
3. **Single-pass streaming with map accumulators** — read each line once; maintain in-memory counter maps that are updated per entry. Memory scales with unique values, not file size.

## Decision

`analyzer.Analyze` performs a single forward pass through the log stream using `logparser.ParseStream` (which uses `bufio.Scanner`). All metrics — status codes, browsers, OS, IPs, countries, daily traffic, page counts — are accumulated simultaneously into map-based counters during this single pass.

After the stream is exhausted, counters are sorted, trimmed to their Top-N limits, and assembled into the `Report` struct.

## Consequences

**Positive:**
- Memory usage is proportional to the number of **unique values** (IPs, URIs, user-agents), not the total number of log lines. A 500 MB log with 10,000 unique IPs uses far less than 500 MB of RAM.
- No temp files; no seeking; works with any `io.Reader` (upload stream, local file, or future pipe)
- Analysis latency is O(n) in number of lines; no repeated passes

**Negative:**
- All counters must be maintained simultaneously, which increases code complexity compared to a simpler sequential approach
- A log with an extremely large number of unique IPs (e.g., a DDoS log) could still consume significant memory — bounded by unique IP count, not file size, but potentially large
- No ability to short-circuit early (e.g., stop after the first N entries for a preview)
