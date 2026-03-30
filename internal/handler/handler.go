package handler

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/bjblazko/caddyshack/internal/analyzer"
)

const maxUploadSize = 500 * 1024 * 1024 // 500 MB

var tempDir = filepath.Join(os.TempDir(), "caddyshack")

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
	defer file.Close()

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
	defer saved.Close()

	log.Printf("Analyzing uploaded file: %s", header.Filename)
	result := analyzer.Analyze(saved, analyzer.FilterParams{})
	result.FileID = fileID

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}

// Health handles GET /api/health.
func Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"ok"}`))
}

func saveTempFile(r io.Reader) (string, error) {
	if err := os.MkdirAll(tempDir, 0700); err != nil {
		return "", err
	}
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	id := hex.EncodeToString(b)
	f, err := os.Create(filepath.Join(tempDir, id+".jsonl"))
	if err != nil {
		return "", err
	}
	defer f.Close()
	if _, err := io.Copy(f, r); err != nil {
		return "", err
	}
	return id, nil
}
