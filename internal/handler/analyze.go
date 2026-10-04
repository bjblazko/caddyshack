package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/bjblazko/caddyshack/internal/analyzer"
)

// Analyze handles GET /api/analyze. It accepts either file=<id> (uploaded temp
// file) or name=<filename> (server-side log), plus optional filter params:
// host, start (YYYY-MM-DD), end (YYYY-MM-DD), country, browser, os, page, status.
func Analyze(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	f, filePath, ok := openLogFile(w, q)
	if !ok {
		return
	}
	defer func() { _ = f.Close() }()
	params := filterParams(q)

	log.Printf("Analyzing %s (host=%q start=%q end=%q country=%q browser=%q os=%q page=%q status=%q method=%q)",
		filePath, params.Host, params.StartDate, params.EndDate,
		params.Country, params.Browser, params.OS, params.Page, params.Status, params.Method)

	result := analyzer.Analyze(f, params)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}
