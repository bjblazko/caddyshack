package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/bjblazko/caddyshack/internal/analyzer"
)

// Events handles GET /api/events. Accepts the same file/name and filter params
// as /api/analyze, plus offset and limit for pagination (default limit 100, max 200).
func Events(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	fileID := q.Get("file")
	localName := q.Get("name")

	var filePath string
	switch {
	case fileID != "":
		if strings.ContainsAny(fileID, "/\\..") {
			http.Error(w, "Invalid file id", http.StatusBadRequest)
			return
		}
		filePath = filepath.Join(tempDir, fileID+".jsonl")
	case localName != "":
		if strings.Contains(localName, "/") || strings.Contains(localName, "..") {
			http.Error(w, "Invalid filename", http.StatusBadRequest)
			return
		}
		filePath = filepath.Join(logDir, localName)
	default:
		http.Error(w, "Provide file or name query param", http.StatusBadRequest)
		return
	}

	f, err := os.Open(filePath)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}
	defer f.Close()

	params := analyzer.FilterParams{
		Host:         q.Get("host"),
		StartDate:    q.Get("start"),
		EndDate:      q.Get("end"),
		Country:      q.Get("country"),
		Browser:      q.Get("browser"),
		OS:           q.Get("os"),
		Page:         q.Get("page"),
		Status:       q.Get("status"),
		Method:       q.Get("method"),
		IgnoreStatic: q.Get("ignore_static") == "1",
		IgnoreImages: q.Get("ignore_images") == "1",
		Search:       q.Get("search"),
	}

	offset, _ := strconv.Atoi(q.Get("offset"))
	limit, _ := strconv.Atoi(q.Get("limit"))
	if limit <= 0 {
		limit = 100
	}

	log.Printf("Events %s offset=%d limit=%d (host=%q start=%q end=%q)",
		filePath, offset, limit, params.Host, params.StartDate, params.EndDate)

	result := analyzer.ListEvents(f, params, offset, limit)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		log.Printf("Error encoding events response: %v", err)
	}
}
