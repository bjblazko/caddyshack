package handler

import (
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/bjblazko/caddyshack/internal/analyzer"
)

// openLogFile opens the log selected by the query: file=<id> (uploaded temp
// file) or name=<filename> (server-side log). On failure it writes the HTTP
// error and returns ok=false.
func openLogFile(w http.ResponseWriter, q url.Values) (f *os.File, path string, ok bool) {
	path, msg := logFilePath(q)
	if msg != "" {
		http.Error(w, msg, http.StatusBadRequest)
		return nil, "", false
	}
	isUpload := q.Get("file") != ""
	f, err := os.Open(path)
	switch {
	case err != nil && isUpload:
		http.Error(w, "Upload expired or not found, please upload the file again", http.StatusNotFound)
		return nil, "", false
	case err != nil:
		http.Error(w, "File not found", http.StatusNotFound)
		return nil, "", false
	case isUpload:
		touchUpload(path)
	}
	return f, path, true
}

// logFilePath resolves the query to a file path, guarding against path
// traversal. A non-empty msg describes why the query is invalid.
func logFilePath(q url.Values) (path, msg string) {
	fileID := q.Get("file")
	localName := q.Get("name")
	switch {
	case fileID != "":
		// Uploaded temp file — the ID must not contain path characters.
		if strings.ContainsAny(fileID, "/\\..") {
			return "", "Invalid file id"
		}
		return filepath.Join(tempDir, fileID+".jsonl"), ""
	case localName != "":
		if strings.Contains(localName, "/") || strings.Contains(localName, "..") {
			return "", "Invalid filename"
		}
		return filepath.Join(logDir, localName), ""
	default:
		return "", "Provide file or name query param"
	}
}

func filterParams(q url.Values) analyzer.FilterParams {
	return analyzer.FilterParams{
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
}
