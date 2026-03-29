package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/bjblazko/caddyshack/internal/analyzer"
)

const maxUploadSize = 500 * 1024 * 1024 // 500 MB

// Upload handles POST /api/upload with a JSONL log file.
func Upload(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "File too large or invalid form data", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("logfile")
	if err != nil {
		http.Error(w, "Missing logfile field", http.StatusBadRequest)
		return
	}
	defer file.Close()

	log.Printf("Analyzing uploaded file: %s (%d bytes)", header.Filename, header.Size)

	report := analyzer.Analyze(file)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(report); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}

// Health handles GET /api/health.
func Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"ok"}`))
}
