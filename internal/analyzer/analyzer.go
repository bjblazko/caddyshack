// Package analyzer filters and aggregates Caddy access log entries.
package analyzer

import (
	"io"
	"sort"
)

// NameCount is one row of a ranked or sorted breakdown.
type NameCount struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

// DayCount is the number of requests on one UTC day.
type DayCount struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

// VisitorInfo is one row of the top visitors list; IP is anonymized.
type VisitorInfo struct {
	IP          string `json:"ip"`
	Count       int    `json:"count"`
	Country     string `json:"country"`
	CountryName string `json:"country_name"`
}

// CountryCount is the number of requests from one country.
type CountryCount struct {
	Code  string `json:"code"`
	Name  string `json:"name"`
	Count int    `json:"count"`
}

// Report holds all aggregated metrics for the entries that passed the filters.
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
	Methods          []NameCount    `json:"methods"`
	TopReferrers     []NameCount    `json:"top_referrers"`
}

// AnalysisResult is the top-level response from Analyze.
type AnalysisResult struct {
	FileID string   `json:"file_id,omitempty"`
	Hosts  []string `json:"hosts"`
	Report *Report  `json:"report"`
}

// Analyze streams log data from r, applies all FilterParams conditions with AND
// logic, and returns an aggregated Report. Hosts are collected from entries that
// pass all filters except the host filter, so the host list always reflects what
// is selectable given the other active filters.
func Analyze(r io.Reader, params FilterParams) *AnalysisResult {
	hostSeen := make(map[string]bool)
	report := newReportBuilder(params.Status == "error")

	parseEvents(r, func(e logEvent) {
		if !params.matchesExceptHost(e) {
			return
		}
		// Collected before the host filter so the host dropdown stays
		// meaningful under all other filters.
		if e.Request.Host != "" {
			hostSeen[e.Request.Host] = true
		}
		if params.matchesHost(e.Request.Host) {
			report.add(e)
		}
	})

	return &AnalysisResult{
		Hosts:  sortedKeys(hostSeen),
		Report: report.build(),
	}
}

func sortedKeys(set map[string]bool) []string {
	keys := make([]string, 0, len(set))
	for k := range set {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
