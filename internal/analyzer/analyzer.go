package analyzer

import (
	"io"
	"sort"
	"time"

	"github.com/bjblazko/caddyshack/internal/anonymize"
	"github.com/bjblazko/caddyshack/internal/geoip"
	"github.com/bjblazko/caddyshack/internal/logparser"
	"github.com/bjblazko/caddyshack/internal/useragent"
	"strings"
)

var assetPrefixes = []string{"/css/", "/js/", "/img/", "/fonts/", "/api"}
var assetExtensions = []string{".css", ".js", ".png", ".jpg", ".svg", ".ttf", ".woff", ".woff2", ".ico"}

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

type FullReport struct {
	All     *Report `json:"all"`
	Success *Report `json:"success"`
	Error   *Report `json:"error"`
}

type MultiHostReport struct {
	Hosts  []string                `json:"hosts"`
	ByHost map[string]*FullReport  `json:"by_host"`
}

// Analyze reads JSONL log data and produces reports split by virtual host and status filter.
func Analyze(r io.Reader) *MultiHostReport {
	var entries []logparser.LogEntry
	hostSeen := make(map[string]bool)
	logparser.ParseStream(r, func(entry logparser.LogEntry) {
		entries = append(entries, entry)
		host := entry.Request.Host
		if host != "" {
			hostSeen[host] = true
		}
	})

	hosts := make([]string, 0, len(hostSeen))
	for h := range hostSeen {
		hosts = append(hosts, h)
	}
	sort.Strings(hosts)

	byHost := make(map[string]*FullReport, len(hosts)+1)
	byHost["all"] = fullReportFor(entries, nil)
	for _, host := range hosts {
		h := host
		byHost[h] = fullReportFor(entries, func(e logparser.LogEntry) bool { return e.Request.Host == h })
	}

	return &MultiHostReport{Hosts: hosts, ByHost: byHost}
}

func fullReportFor(entries []logparser.LogEntry, hostFilter func(logparser.LogEntry) bool) *FullReport {
	var subset []logparser.LogEntry
	if hostFilter == nil {
		subset = entries
	} else {
		for _, e := range entries {
			if hostFilter(e) {
				subset = append(subset, e)
			}
		}
	}
	return &FullReport{
		All:     analyzeFiltered(subset, nil, false),
		Success: analyzeFiltered(subset, func(e logparser.LogEntry) bool { return e.Status >= 200 && e.Status < 300 }, false),
		Error:   analyzeFiltered(subset, func(e logparser.LogEntry) bool { return e.Status >= 400 }, true),
	}
}

// analyzeFiltered produces a report from log entries, optionally filtered by a predicate.
// isErrorFilter indicates whether to count URIs with status >= 400 (errors) or < 400 (success).
func analyzeFiltered(entries []logparser.LogEntry, filter func(logparser.LogEntry) bool, isErrorFilter bool) *Report {
	statusCodes := make(map[int]int)
	pages := make(map[string]int)
	browsers := make(map[string]int)
	oses := make(map[string]int)
	ips := make(map[string]int)
	daily := make(map[string]int)
	countryCounts := make(map[string]int)

	var totalRequests int
	var totalBytes int64
	var totalDuration float64

	for _, entry := range entries {
		if filter != nil && !filter(entry) {
			continue
		}

		totalRequests++
		totalBytes += entry.Size
		totalDuration += entry.Duration

		req := entry.Request
		clientIP := req.ClientIP
		if clientIP == "" {
			clientIP = req.RemoteIP
		}

		// GeoIP on original IP before anonymization
		countryCode := geoip.Lookup(clientIP)
		countryCounts[countryCode]++

		// Anonymize IP for storage/display
		anonIP := anonymize.IP(clientIP)
		ips[anonIP]++

		// User-Agent
		ua := ""
		if uaList, ok := req.Headers["User-Agent"]; ok && len(uaList) > 0 {
			ua = uaList[0]
		}
		browser, osName := useragent.Parse(ua)
		browsers[browser]++
		oses[osName]++

		// Status
		statusCodes[entry.Status]++

		// Daily
		t := time.Unix(int64(entry.Timestamp), int64((entry.Timestamp-float64(int64(entry.Timestamp)))*1e9))
		daily[t.UTC().Format("2006-01-02")]++

		// Pages: for error filter, include status >= 400; otherwise, only status < 400
		uri := req.URI
		if !isAsset(uri) {
			if isErrorFilter && entry.Status >= 400 {
				pages[uri]++
			} else if !isErrorFilter && entry.Status < 400 {
				pages[uri]++
			}
		}
	}

	var avgMs float64
	if totalRequests > 0 {
		avgMs = (totalDuration / float64(totalRequests)) * 1000
	}

	return &Report{
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
	}
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
		result[i] = NameCount{Name: strings.TrimLeft(strings.Replace(string(rune('0'+k/100))+"xx", "0xx", "", 1), ""), Count: m[k]}
	}
	// Actually, let's use the actual status code as string
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
