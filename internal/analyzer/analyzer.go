package analyzer

import (
	"io"
	"sort"
	"strings"
	"time"

	"github.com/bjblazko/caddyshack/internal/anonymize"
	"github.com/bjblazko/caddyshack/internal/geoip"
	"github.com/bjblazko/caddyshack/internal/logparser"
	"github.com/bjblazko/caddyshack/internal/useragent"
)

var assetPrefixes = []string{"/css/", "/js/", "/img/", "/fonts/", "/api"}
var assetExtensions = []string{".css", ".js", ".png", ".jpg", ".svg", ".ttf", ".woff", ".woff2", ".ico"}

// FilterParams holds all filter criteria applied before aggregation.
// All conditions are ANDed. Empty/zero values mean "no filter" for that dimension.
type FilterParams struct {
	Host      string // virtual host, "" = all
	StartDate string // "YYYY-MM-DD", "" = unbounded
	EndDate   string // "YYYY-MM-DD", "" = unbounded
	Country   string // country name, "" = all
	Browser   string // browser name, "" = all
	OS        string // OS name, "" = all
	Page      string // exact URI, "" = all
	Status    string // "success" | "error", "" = all
}

type NameCount struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

type DayCount struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

type VisitorInfo struct {
	IP          string `json:"ip"`
	Count       int    `json:"count"`
	Country     string `json:"country"`
	CountryName string `json:"country_name"`
}

type CountryCount struct {
	Code  string `json:"code"`
	Name  string `json:"name"`
	Count int    `json:"count"`
}

type Report struct {
	TotalRequests    int            `json:"total_requests"`
	UniqueIPs        int            `json:"unique_ips"`
	TotalBytes       int64          `json:"total_bytes"`
	AvgResponseMs    float64        `json:"avg_response_ms"`
	StatusCodes      []NameCount    `json:"status_codes"`
	TopPages         []NameCount    `json:"top_pages"`
	Browsers         []NameCount    `json:"browsers"`
	OperatingSystems []NameCount    `json:"operating_systems"`
	DailyTraffic     []DayCount     `json:"daily_traffic"`
	TopVisitors      []VisitorInfo  `json:"top_visitors"`
	Countries        []CountryCount `json:"countries"`
}

// AnalysisResult is the top-level response from Analyze.
type AnalysisResult struct {
	FileID string   `json:"file_id,omitempty"`
	Hosts  []string `json:"hosts"`
	Report *Report  `json:"report"`
}

// EventEntry is a single enriched log entry for the Single Events view.
type EventEntry struct {
	Timestamp   string  `json:"ts"`
	Method      string  `json:"method"`
	Host        string  `json:"host"`
	URI         string  `json:"uri"`
	Status      int     `json:"status"`
	Size        int64   `json:"size"`
	DurationMs  float64 `json:"duration_ms"`
	IP          string  `json:"ip"`
	Country     string  `json:"country"`
	CountryName string  `json:"country_name"`
	Browser     string  `json:"browser"`
	OS          string  `json:"os"`
}

// EventsResult is the paginated response from ListEvents.
type EventsResult struct {
	Total  int          `json:"total"`
	Offset int          `json:"offset"`
	Limit  int          `json:"limit"`
	Events []EventEntry `json:"events"`
}

// Analyze streams log data from r, applies all FilterParams conditions with AND
// logic, and returns an aggregated Report. Hosts are collected from entries that
// pass all filters except the host filter, so the host list always reflects what
// is selectable given the other active filters.
func Analyze(r io.Reader, params FilterParams) *AnalysisResult {
	statusCodes := make(map[int]int)
	pages := make(map[string]int)
	browsers := make(map[string]int)
	oses := make(map[string]int)
	ips := make(map[string]int)
	daily := make(map[string]int)
	countryCounts := make(map[string]int)
	hostSeen := make(map[string]bool)

	var totalRequests int
	var totalBytes int64
	var totalDuration float64

	logparser.ParseStream(r, func(entry logparser.LogEntry) {
		req := entry.Request

		clientIP := req.ClientIP
		if clientIP == "" {
			clientIP = req.RemoteIP
		}

		t := time.Unix(int64(entry.Timestamp), int64((entry.Timestamp-float64(int64(entry.Timestamp)))*1e9))
		day := t.UTC().Format("2006-01-02")

		ua := ""
		if uaList, ok := req.Headers["User-Agent"]; ok && len(uaList) > 0 {
			ua = uaList[0]
		}
		browser, osName := useragent.Parse(ua)

		countryCode := geoip.Lookup(clientIP)
		countryName := geoip.CountryName(countryCode)

		// Collect hosts from entries passing every filter except host,
		// so the host dropdown stays meaningful under all other filters.
		if passesNonHostFilters(day, entry.Status, browser, osName, countryName, req.URI, params) {
			if req.Host != "" {
				hostSeen[req.Host] = true
			}
		}

		// Skip entries that don't pass all filters (including host).
		if params.Host != "" && req.Host != params.Host {
			return
		}
		if !passesNonHostFilters(day, entry.Status, browser, osName, countryName, req.URI, params) {
			return
		}

		totalRequests++
		totalBytes += entry.Size
		totalDuration += entry.Duration

		anonIP := anonymize.IP(clientIP)
		ips[anonIP]++
		browsers[browser]++
		oses[osName]++
		statusCodes[entry.Status]++
		daily[day]++
		countryCounts[countryCode]++

		// Count non-asset pages: for error-scoped queries include error responses,
		// otherwise include only non-error responses (preserving historic behaviour).
		if !isAsset(req.URI) && (params.Status == "error" || entry.Status < 400) {
			pages[req.URI]++
		}
	})

	var avgMs float64
	if totalRequests > 0 {
		avgMs = (totalDuration / float64(totalRequests)) * 1000
	}

	hosts := make([]string, 0, len(hostSeen))
	for h := range hostSeen {
		hosts = append(hosts, h)
	}
	sort.Strings(hosts)

	return &AnalysisResult{
		Hosts: hosts,
		Report: &Report{
			TotalRequests:    totalRequests,
			UniqueIPs:        len(ips),
			TotalBytes:       totalBytes,
			AvgResponseMs:    avgMs,
			StatusCodes:      sortedIntNameCounts(statusCodes),
			TopPages:         topN(pages, 15),
			Browsers:         topN(browsers, 10),
			OperatingSystems: topN(oses, 10),
			DailyTraffic:     sortedDays(daily),
			TopVisitors:      topVisitors(ips, 10),
			Countries:        topCountries(countryCounts, 15),
		},
	}
}

// passesNonHostFilters returns true if the entry satisfies every active filter
// except the host filter.
func passesNonHostFilters(day string, status int, browser, osName, countryName, uri string, p FilterParams) bool {
	if p.StartDate != "" && day < p.StartDate {
		return false
	}
	if p.EndDate != "" && day > p.EndDate {
		return false
	}
	if p.Status == "success" && !(status >= 200 && status < 300) {
		return false
	}
	if p.Status == "error" && status < 400 {
		return false
	}
	if p.Browser != "" && browser != p.Browser {
		return false
	}
	if p.OS != "" && osName != p.OS {
		return false
	}
	if p.Country != "" && countryName != p.Country {
		return false
	}
	if p.Page != "" && uri != p.Page {
		return false
	}
	return true
}

func isAsset(uri string) bool {
	for _, p := range assetPrefixes {
		if strings.HasPrefix(uri, p) {
			return true
		}
	}
	for _, e := range assetExtensions {
		if strings.HasSuffix(uri, e) {
			return true
		}
	}
	return false
}

func topN(m map[string]int, n int) []NameCount {
	result := make([]NameCount, 0, len(m))
	for k, v := range m {
		result = append(result, NameCount{Name: k, Count: v})
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Count > result[j].Count
	})
	if len(result) > n {
		result = result[:n]
	}
	return result
}

func sortedIntNameCounts(m map[int]int) []NameCount {
	keys := make([]int, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	result := make([]NameCount, len(keys))
	for i, k := range keys {
		result[i] = NameCount{Name: statusString(k), Count: m[k]}
	}
	return result
}

func statusString(code int) string {
	s := ""
	s += string(rune('0' + code/100))
	s += string(rune('0' + (code/10)%10))
	s += string(rune('0' + code%10))
	return s
}

func sortedDays(m map[string]int) []DayCount {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	result := make([]DayCount, len(keys))
	for i, k := range keys {
		result[i] = DayCount{Date: k, Count: m[k]}
	}
	return result
}

func topVisitors(m map[string]int, n int) []VisitorInfo {
	type ipCount struct {
		ip    string
		count int
	}
	items := make([]ipCount, 0, len(m))
	for ip, count := range m {
		items = append(items, ipCount{ip, count})
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].count > items[j].count
	})
	if len(items) > n {
		items = items[:n]
	}
	result := make([]VisitorInfo, len(items))
	for i, item := range items {
		code := geoip.Lookup(item.ip)
		result[i] = VisitorInfo{
			IP:          item.ip,
			Count:       item.count,
			Country:     code,
			CountryName: geoip.CountryName(code),
		}
	}
	return result
}

// ListEvents streams log data from r, applies all FilterParams with AND logic,
// enriches each matching entry (GeoIP, UA, anonymized IP), sorts most-recent-first,
// and returns a paginated slice. limit is capped at 200.
func ListEvents(r io.Reader, params FilterParams, offset, limit int) *EventsResult {
	if limit <= 0 || limit > 200 {
		limit = 100
	}

	type rawEvent struct {
		ts    float64
		entry EventEntry
	}
	var all []rawEvent

	logparser.ParseStream(r, func(entry logparser.LogEntry) {
		req := entry.Request

		clientIP := req.ClientIP
		if clientIP == "" {
			clientIP = req.RemoteIP
		}

		t := time.Unix(int64(entry.Timestamp), int64((entry.Timestamp-float64(int64(entry.Timestamp)))*1e9))
		day := t.UTC().Format("2006-01-02")

		ua := ""
		if uaList, ok := req.Headers["User-Agent"]; ok && len(uaList) > 0 {
			ua = uaList[0]
		}
		browser, osName := useragent.Parse(ua)

		countryCode := geoip.Lookup(clientIP)
		countryName := geoip.CountryName(countryCode)

		if params.Host != "" && req.Host != params.Host {
			return
		}
		if !passesNonHostFilters(day, entry.Status, browser, osName, countryName, req.URI, params) {
			return
		}

		all = append(all, rawEvent{
			ts: entry.Timestamp,
			entry: EventEntry{
				Timestamp:   t.UTC().Format(time.RFC3339Nano),
				Method:      req.Method,
				Host:        req.Host,
				URI:         req.URI,
				Status:      entry.Status,
				Size:        entry.Size,
				DurationMs:  entry.Duration * 1000,
				IP:          anonymize.IP(clientIP),
				Country:     countryCode,
				CountryName: countryName,
				Browser:     browser,
				OS:          osName,
			},
		})
	})

	sort.Slice(all, func(i, j int) bool {
		return all[i].ts > all[j].ts
	})

	total := len(all)
	if offset >= total {
		return &EventsResult{Total: total, Offset: offset, Limit: limit, Events: []EventEntry{}}
	}
	end := offset + limit
	if end > total {
		end = total
	}
	events := make([]EventEntry, end-offset)
	for i, raw := range all[offset:end] {
		events[i] = raw.entry
	}
	return &EventsResult{Total: total, Offset: offset, Limit: limit, Events: events}
}

func topCountries(m map[string]int, n int) []CountryCount {
	type cc struct {
		code  string
		count int
	}
	items := make([]cc, 0, len(m))
	for code, count := range m {
		items = append(items, cc{code, count})
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].count > items[j].count
	})
	if len(items) > n {
		items = items[:n]
	}
	result := make([]CountryCount, len(items))
	for i, item := range items {
		result[i] = CountryCount{
			Code:  item.code,
			Name:  geoip.CountryName(item.code),
			Count: item.count,
		}
	}
	return result
}
