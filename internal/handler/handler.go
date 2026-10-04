// Package handler implements the CaddyShack HTTP API.
package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/bjblazko/caddyshack/internal/analyzer"
)

const maxUploadSize = 500 * 1024 * 1024 // 500 MB

// Upload handles POST /api/upload. It saves the file to a temp directory so it
// can be re-analyzed on each filter change without re-uploading.
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
	defer func() { _ = file.Close() }()

	log.Printf("Saving uploaded file: %s (%d bytes)", header.Filename, header.Size)

	fileID, err := saveTempFile(file)
	if err != nil {
		log.Printf("Error saving upload: %v", err)
		http.Error(w, "Failed to store file", http.StatusInternalServerError)
		return
	}

	saved, err := os.Open(filepath.Join(tempDir, fileID+".jsonl"))
	if err != nil {
		log.Printf("Error reopening saved file: %v", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
	defer func() { _ = saved.Close() }()

	log.Printf("Analyzing uploaded file: %s", header.Filename)
	// Same filter query params as /api/analyze, so the first view already
	// honours the UI's default exclusions.
	result := analyzer.Analyze(saved, filterParams(r.URL.Query()))
	result.FileID = fileID

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}

// version is the release this binary was built from, shown in the About
// dialog; "dev" for local builds.
var version = "dev"

// SetVersion sets the version reported by Health (set via -ldflags in main).
func SetVersion(v string) { version = v }

// Health handles GET /api/health.
func Health(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{"status": "ok", "version": version}); err != nil {
		log.Printf("Error encoding health response: %v", err)
	}
}
