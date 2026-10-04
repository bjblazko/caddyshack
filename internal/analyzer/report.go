package analyzer

import "github.com/bjblazko/caddyshack/internal/anonymize"

// reportBuilder accumulates the counters of a Report one event at a time.
type reportBuilder struct {
	includeErrorPages bool

	totalRequests int
	totalBytes    int64
	totalDuration float64

	statusCodes map[int]int
	pages       map[string]int
	browsers    map[string]int
	oses        map[string]int
	ips         map[string]int // keyed by anonymized IP
	daily       map[string]int
	countries   map[string]int
	methods     map[string]int
	referrers   map[string]int
}

// newReportBuilder creates an empty builder. Top pages count only non-error
// responses unless includeErrorPages is set (used for error-scoped queries).
func newReportBuilder(includeErrorPages bool) *reportBuilder {
	return &reportBuilder{
		includeErrorPages: includeErrorPages,
		statusCodes:       make(map[int]int),
		pages:             make(map[string]int),
		browsers:          make(map[string]int),
		oses:              make(map[string]int),
		ips:               make(map[string]int),
		daily:             make(map[string]int),
		countries:         make(map[string]int),
		methods:           make(map[string]int),
		referrers:         make(map[string]int),
	}
}

func (b *reportBuilder) add(e logEvent) {
	b.totalRequests++
	b.totalBytes += e.Size
	b.totalDuration += e.Duration

	b.ips[anonymize.IP(e.ClientIP)]++
	b.browsers[e.Browser]++
	b.oses[e.OS]++
	b.statusCodes[e.Status]++
	b.daily[e.Day]++
	b.countries[e.CountryCode]++
	if e.Request.Method != "" {
		b.methods[e.Request.Method]++
	}
	if e.Referer != "" {
		b.referrers[e.Referer]++
	}
	if !isAsset(e.Request.URI) && (b.includeErrorPages || e.Status < 400) {
		b.pages[e.Request.URI]++
	}
}

// build assembles the Report. Referrers whose host is one of siteHosts are
// internal navigation and left out of TopReferrers.
func (b *reportBuilder) build(siteHosts map[string]bool) *Report {
	var avgMs float64
	if b.totalRequests > 0 {
		avgMs = (b.totalDuration / float64(b.totalRequests)) * 1000
	}
	return &Report{
		TotalRequests:    b.totalRequests,
		UniqueIPs:        len(b.ips),
		TotalBytes:       b.totalBytes,
		AvgResponseMs:    avgMs,
		StatusCodes:      sortedIntNameCounts(b.statusCodes),
		TopPages:         topN(b.pages, 15),
		Browsers:         topN(b.browsers, 10),
		OperatingSystems: topN(b.oses, 10),
		DailyTraffic:     sortedDays(b.daily),
		TopVisitors:      topVisitors(b.ips, 10),
		Countries:        topCountries(b.countries, 15),
		Methods:          topN(b.methods, 20),
		TopReferrers:     topN(externalReferrers(b.referrers, siteHosts), 10),
	}
}

func externalReferrers(referrers map[string]int, siteHosts map[string]bool) map[string]int {
	external := make(map[string]int, len(referrers))
	for ref, count := range referrers {
		if !siteHosts[referrerHost(ref)] {
			external[ref] = count
		}
	}
	return external
}
