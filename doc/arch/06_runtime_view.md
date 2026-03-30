# 6. Runtime View

## Scenario 1: Browser Upload and Analysis

The primary use case: the operator drops a log file onto the dashboard.

```mermaid
sequenceDiagram
    participant B as Browser
    participant H as handler.Upload
    participant A as analyzer.Analyze
    participant LP as logparser / geoip / anonymize / useragent

    B->>H: POST /api/upload (multipart, logfile)
    Note over H: MaxBytesReader(500 MB)<br/>ParseMultipartForm()<br/>extract file reader
    H->>A: Analyze(reader)
    loop for each line
        A->>LP: ParseStream(reader)
        LP-->>A: LogEntry
        Note over A: 1. extract client_ip / remote_ip<br/>2. geoip.Lookup(original ip)<br/>3. anonymize.IP(original ip)<br/>4. useragent.Parse(UA header)<br/>5. increment all counters
    end
    Note over A: sort + trim counters
    A-->>H: MultiHostReport
    Note over H: JSON encode
    H-->>B: 200 OK (JSON)
    Note over B: render dashboard<br/>(app.js + charts.js + map.js)
```

**Memory lifecycle:** all parsed data (counters, maps) exists only within `analyzer.Analyze`. Once the JSON response is written, the memory is eligible for GC.

---

## Scenario 2: Server-Side Log Analysis

The operator selects a log file from the server's `/var/log/caddy` directory without uploading it.

```mermaid
sequenceDiagram
    participant B as Browser
    participant H as handler.Logs / AnalyzeLocal
    participant FS as Filesystem

    B->>H: GET /api/logs
    H->>FS: ReadDir(/var/log/caddy)
    FS-->>H: []LogFileInfo
    H-->>B: 200 OK (JSON array)

    B->>H: GET /api/analyze-local?name=access.json
    Note over H: validate name (no path separator)<br/>Open(file)
    H->>FS: Open(file)
    FS-->>H: io.Reader
    Note over H: → analyzer.Analyze(reader)
    H-->>B: 200 OK (MultiHostReport)
```

---

## Scenario 3: Host / Filter Switch (Client-Side Only)

After the initial analysis, switching host or traffic filter requires no server call.

```mermaid
flowchart TD
    A["User selects host or traffic filter"] --> B["app.js reads cached data\nreport = fullData.by_host[host][filter]"]
    B --> C["Charts.renderBarChart\n(browsers, OS, status codes)"]
    B --> D["Charts.renderVerticalBarChart\n(daily traffic)"]
    B --> E["WorldMap.render\n(country bubbles)"]
    B --> F["DOM table updates\n(pages, visitors, countries)"]
    C & D & E & F --> G["No HTTP request made"]
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
