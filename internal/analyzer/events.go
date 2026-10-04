package analyzer

import (
	"io"
	"sort"
	"time"

	"github.com/bjblazko/caddyshack/internal/anonymize"
)

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
	Referer     string  `json:"referer"`
}

const (
	defaultEventsLimit = 100
	maxEventsLimit     = 200
)

// EventsResult is the paginated response from ListEvents.
type EventsResult struct {
	Total  int          `json:"total"`
	Offset int          `json:"offset"`
	Limit  int          `json:"limit"`
	Events []EventEntry `json:"events"`
}

// ListEvents streams log data from r, applies all FilterParams with AND logic,
// enriches each matching entry (GeoIP, UA, anonymized IP), sorts most-recent-first,
// and returns a paginated slice. A non-positive limit defaults to 100 and limit
// is capped at 200; a negative offset is treated as 0.
func ListEvents(r io.Reader, params FilterParams, offset, limit int) *EventsResult {
	switch {
	case limit <= 0:
		limit = defaultEventsLimit
	case limit > maxEventsLimit:
		limit = maxEventsLimit
	}
	offset = max(offset, 0)

	// Convert while streaming: holding full log entries (headers included)
	// would cost far more memory than the compact EventEntry.
	type timedEvent struct {
		ts    float64
		entry EventEntry
	}
	var all []timedEvent

	parseEvents(r, func(e logEvent) {
		if params.matchesHost(e.Request.Host) && params.matchesExceptHost(e) {
			all = append(all, timedEvent{ts: e.Timestamp, entry: toEventEntry(e)})
		}
	})

	sort.SliceStable(all, func(i, j int) bool {
		return all[i].ts > all[j].ts
	})

	start := min(offset, len(all))
	end := min(offset+limit, len(all))
	events := make([]EventEntry, 0, end-start)
	for _, t := range all[start:end] {
		events = append(events, t.entry)
	}
	return &EventsResult{Total: len(all), Offset: offset, Limit: limit, Events: events}
}

func toEventEntry(e logEvent) EventEntry {
	return EventEntry{
		Timestamp:   e.Time.UTC().Format(time.RFC3339Nano),
		Method:      e.Request.Method,
		Host:        e.Request.Host,
		URI:         e.Request.URI,
		Status:      e.Status,
		Size:        e.Size,
		DurationMs:  e.Duration * 1000,
		IP:          anonymize.IP(e.ClientIP),
		Country:     e.CountryCode,
		CountryName: e.CountryName,
		Browser:     e.Browser,
		OS:          e.OS,
		Referer:     e.Referer,
	}
}
