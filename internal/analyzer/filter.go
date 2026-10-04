package analyzer

import (
	"strings"
)

// FilterParams holds all filter criteria applied before aggregation.
// All conditions are ANDed. Empty/zero values mean "no filter" for that dimension.
type FilterParams struct {
	Host         string // virtual host, "" = all
	StartDate    string // "YYYY-MM-DD", "" = unbounded
	EndDate      string // "YYYY-MM-DD", "" = unbounded
	Country      string // country name, "" = all
	Browser      string // browser name, "" = all
	OS           string // OS name, "" = all
	Page         string // exact URI, "" = all
	Status       string // "success" | "error", "" = all
	Method       string // HTTP method e.g. "GET", "" = all
	IgnoreStatic bool   // exclude JS, CSS, fonts, robots.txt, etc.
	IgnoreImages bool   // exclude PNG, JPG, SVG, ICO, etc.
	Search       string // glob pattern matched against URI, client IP, and Referer; "" = no filter
}

// matchesExceptHost reports whether e satisfies every active filter except the
// host filter.
func (p FilterParams) matchesExceptHost(e logEvent) bool {
	return p.matchesDateRange(e.Day) &&
		p.matchesStatus(e.Status) &&
		p.matchesDimensions(e) &&
		p.matchesResourceType(e.Request.URI) &&
		p.matchesSearch(e)
}

func (p FilterParams) matchesHost(host string) bool {
	return p.Host == "" || host == p.Host
}

func (p FilterParams) matchesDateRange(day string) bool {
	return (p.StartDate == "" || day >= p.StartDate) &&
		(p.EndDate == "" || day <= p.EndDate)
}

func (p FilterParams) matchesStatus(status int) bool {
	switch p.Status {
	case "success":
		return status >= 200 && status < 300
	case "error":
		return status >= 400
	}
	return true
}

// matchesDimensions checks the exact-match filters on single dimensions.
func (p FilterParams) matchesDimensions(e logEvent) bool {
	return matchesExact(p.Browser, e.Browser) &&
		matchesExact(p.OS, e.OS) &&
		matchesExact(p.Country, e.CountryName) &&
		matchesExact(p.Page, e.Request.URI) &&
		matchesExact(p.Method, e.Request.Method)
}

func (p FilterParams) matchesResourceType(uri string) bool {
	return (!p.IgnoreStatic || !isStaticResource(uri)) &&
		(!p.IgnoreImages || !isImageResource(uri))
}

func (p FilterParams) matchesSearch(e logEvent) bool {
	return p.Search == "" ||
		matchGlob(p.Search, e.Request.URI) ||
		matchGlob(p.Search, e.ClientIP) ||
		matchGlob(p.Search, e.Referer)
}

// matchesExact treats an empty filter value as "no filter".
func matchesExact(filter, value string) bool {
	return filter == "" || value == filter
}

// matchGlob matches pattern against text case-insensitively using * as a
// multi-character wildcard. No * means an exact (literal) match is required.
// Multiple * are supported; a leading/trailing non-empty segment is anchored
// to the start/end of text.
func matchGlob(pattern, text string) bool {
	p := strings.ToLower(pattern)
	t := strings.ToLower(text)
	if !strings.Contains(p, "*") {
		return p == t
	}
	parts := strings.Split(p, "*")
	// Leading non-empty segment must match as a prefix.
	if parts[0] != "" {
		if !strings.HasPrefix(t, parts[0]) {
			return false
		}
		t = t[len(parts[0]):]
	}
	// Trailing non-empty segment must match as a suffix.
	last := parts[len(parts)-1]
	if last != "" {
		if !strings.HasSuffix(t, last) {
			return false
		}
		t = t[:len(t)-len(last)]
	}
	// Middle segments must appear in order within the remaining text.
	for _, seg := range parts[1 : len(parts)-1] {
		if seg == "" {
			continue
		}
		idx := strings.Index(t, seg)
		if idx == -1 {
			return false
		}
		t = t[idx+len(seg):]
	}
	return true
}
