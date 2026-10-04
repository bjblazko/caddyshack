package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/bjblazko/caddyshack/internal/analyzer"
)

// Events handles GET /api/events. Accepts the same file/name and filter params
// as /api/analyze, plus offset and limit for pagination (default limit 100, max 200).
func Events(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	f, filePath, ok := openLogFile(w, q)
	if !ok {
		return
	}
	defer func() { _ = f.Close() }()
	params := filterParams(q)

	offset, _ := strconv.Atoi(q.Get("offset"))
	limit, _ := strconv.Atoi(q.Get("limit")) // analyzer.ListEvents applies default and cap

	log.Printf("Events %s offset=%d limit=%d (host=%q start=%q end=%q)",
		filePath, offset, limit, params.Host, params.StartDate, params.EndDate)

	result := analyzer.ListEvents(f, params, offset, limit)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		log.Printf("Error encoding events response: %v", err)
	}
}
