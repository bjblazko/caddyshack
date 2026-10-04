package analyzer

import (
	"io"
	"time"

	"github.com/bjblazko/caddyshack/internal/geoip"
	"github.com/bjblazko/caddyshack/internal/logparser"
	"github.com/bjblazko/caddyshack/internal/useragent"
)

// logEvent is a parsed log entry enriched with the derived dimensions that
// filtering and aggregation work on. ClientIP is the raw address; anonymize it
// before it leaves the package.
type logEvent struct {
	logparser.LogEntry
	Time        time.Time
	Day         string // "YYYY-MM-DD" in UTC
	ClientIP    string
	Kind        useragent.Kind
	Browser     string
	OS          string
	Referer     string
	CountryCode string
	CountryName string
}

func enrich(entry logparser.LogEntry) logEvent {
	entry.Request.Host = canonicalHost(entry.Request.Host)
	req := entry.Request
	clientIP := req.ClientIP
	if clientIP == "" {
		clientIP = req.RemoteIP
	}
	t := time.Unix(int64(entry.Timestamp), int64((entry.Timestamp-float64(int64(entry.Timestamp)))*1e9))
	kind, browser, osName := clientOf(req)
	countryCode := geoip.Lookup(clientIP)

	return logEvent{
		LogEntry:    entry,
		Time:        t,
		Day:         t.UTC().Format("2006-01-02"),
		ClientIP:    clientIP,
		Kind:        kind,
		Browser:     browser,
		OS:          osName,
		Referer:     firstHeader(req.Headers, "Referer"),
		CountryCode: countryCode,
		CountryName: geoip.CountryName(countryCode),
	}
}

// clientOf classifies the client from its User-Agent, or, in pre-anonymized
// logs without one, from the browser and OS names the anonymizer kept.
func clientOf(req logparser.Request) (useragent.Kind, string, string) {
	ua := firstHeader(req.Headers, "User-Agent")
	if ua == "" && req.Browser != "" {
		return useragent.FromNames(req.Browser, req.OS)
	}
	return useragent.Parse(ua)
}

func firstHeader(headers map[string][]string, name string) string {
	if values := headers[name]; len(values) > 0 {
		return values[0]
	}
	return ""
}

// parseEvents streams r and calls fn with every enriched entry.
func parseEvents(r io.Reader, fn func(logEvent)) {
	logparser.ParseStream(r, func(entry logparser.LogEntry) {
		fn(enrich(entry))
	})
}
