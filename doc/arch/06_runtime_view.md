# 6. Runtime View

## Scenario 1: Browser Upload and Initial Analysis

The operator drops a log file onto the dashboard. The file is saved to the OS temp directory and immediately analyzed with no active filters.

```mermaid
sequenceDiagram
    participant B as Browser
    participant H as handler.Upload
    participant FS as OS temp dir
    participant A as analyzer.Analyze
    participant LP as logparser / geoip / anonymize / useragent

    B->>H: POST /api/upload (multipart, logfile)
    Note over H: MaxBytesReader(500 MB)<br/>ParseMultipartForm()<br/>extract file reader
    H->>FS: write <hex_id>.jsonl
    FS-->>H: file path
    H->>FS: re-open file
    H->>A: Analyze(reader, FilterParams{})
    loop for each line
        A->>LP: ParseStream(reader)
        LP-->>A: LogEntry
        Note over A: 1. extract client_ip / remote_ip<br/>2. geoip.Lookup(original ip)<br/>3. anonymize.IP(original ip)<br/>4. useragent.Parse(UA header)<br/>5. apply AND filter conditions<br/>6. collect host / increment counters
    end
    Note over A: sort + trim counters
    A-->>H: AnalysisResult{FileID, Hosts, Report}
    Note over H: JSON encode
    H-->>B: 200 OK (JSON with file_id)
    Note over B: render dashboard<br/>store file_id
```

---

## Scenario 2: Filter Change (Backend Re-Analysis)

After the initial upload, every filter change (host, status, date range, country, browser, OS, page) triggers a full backend re-analysis of the same temp file with the new `FilterParams`.

```mermaid
sequenceDiagram
    participant B as Browser
    participant H as handler.Analyze
    participant FS as OS temp dir
    participant A as analyzer.Analyze

    B->>H: GET /api/analyze?file=<id>&host=x&start=2026-01-01&...
    Note over H: parse FilterParams from query string<br/>validate file_id (no path chars)
    H->>FS: open <id>.jsonl
    FS-->>H: io.Reader
    H->>A: Analyze(reader, FilterParams{Host, StartDate, ...})
    Note over A: single pass: filter entries (AND),<br/>collect hosts, aggregate counters
    A-->>H: AnalysisResult{Hosts, Report}
    H-->>B: 200 OK (JSON, no file_id)
    Note over B: renderDashboard(result.report)<br/>repopulate dimension dropdowns
```

---

## Scenario 3: Server-Side Log Analysis

The operator selects a log file from the server's `/var/log/caddy` directory. On every filter change the file is re-read from disk (no temp copy needed).

```mermaid
sequenceDiagram
    participant B as Browser
    participant H as handler.Logs / Analyze
    participant FS as Filesystem

    B->>H: GET /api/logs
    H->>FS: ReadDir(/var/log/caddy)
    FS-->>H: []LogFileInfo
    H-->>B: 200 OK (JSON array)

    B->>H: GET /api/analyze?name=access.json&host=x&...
    Note over H: validate name (no path separator)<br/>Open(/var/log/caddy/access.json)
    H->>FS: open file
    FS-->>H: io.Reader
    Note over H: → analyzer.Analyze(reader, FilterParams)
    H-->>B: 200 OK (AnalysisResult, no file_id)
```

---

## Startup

```mermaid
flowchart TD
    A["main()"] --> B["Parse CLI flags\n-addr, -geodb"]
    B --> C["geoip.Load(geodb path)"]
    C --> D{CSV found?}
    D -->|Yes| E["IP ranges loaded into memory"]
    D -->|No| F["Log warning\ncontinue without GeoIP data"]
    E --> G["Register routes on http.ServeMux"]
    F --> G
    G --> H["http.ListenAndServe(addr)"]
```
