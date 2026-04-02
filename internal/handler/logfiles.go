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

// Analyze handles GET /api/analyze. It accepts either file=<id> (uploaded temp
// file) or name=<filename> (server-side log), plus optional filter params:
// host, start (YYYY-MM-DD), end (YYYY-MM-DD), country, browser, os, page, status.
func Analyze(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	fileID := q.Get("file")
	localName := q.Get("name")

	var filePath string
	switch {
	case fileID != "":
		// Uploaded temp file — validate ID is a plain hex string (no path chars).
		if strings.ContainsAny(fileID, "/\\..") {
			http.Error(w, "Invalid file id", http.StatusBadRequest)
			return
		}
		filePath = filepath.Join(tempDir, fileID+".jsonl")
	case localName != "":
		// Server-side log — guard against path traversal.
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

	log.Printf("Analyzing %s (host=%q start=%q end=%q country=%q browser=%q os=%q page=%q status=%q method=%q)",
		filePath, params.Host, params.StartDate, params.EndDate,
		params.Country, params.Browser, params.OS, params.Page, params.Status, params.Method)

	result := analyzer.Analyze(f, params)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}
