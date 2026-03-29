package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/bjblazko/caddyshack/internal/analyzer"
)

const logDir = "/var/log/caddy"

type LogFileInfo struct {
	Name     string    `json:"name"`
	Size     int64     `json:"size"`
	Modified time.Time `json:"modified"`
}

// LogFiles handles GET /api/logs — returns available Caddy log files sorted newest first.
func LogFiles(w http.ResponseWriter, r *http.Request) {
	entries, err := os.ReadDir(logDir)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("[]"))
		return
	}

	var files []LogFileInfo
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(name, ".json") && !strings.HasSuffix(name, ".jsonl") && !strings.HasSuffix(name, ".log") {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		files = append(files, LogFileInfo{
			Name:     name,
			Size:     info.Size(),
			Modified: info.ModTime(),
		})
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].Modified.After(files[j].Modified)
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(files)
}

// AnalyzeLocal handles GET /api/analyze-local?name=<filename> for server-side log files.
func AnalyzeLocal(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" || strings.Contains(name, "/") || strings.Contains(name, "..") {
		http.Error(w, "Invalid filename", http.StatusBadRequest)
		return
	}

	path := filepath.Join(logDir, name)
	file, err := os.Open(path)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}
	defer file.Close()

	log.Printf("Analyzing local log file: %s", path)

	report := analyzer.Analyze(file)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(report); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}
