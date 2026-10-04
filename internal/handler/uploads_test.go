package handler

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestSweepUploadsDeletesOnlyIdleFiles(t *testing.T) {
	useTempDir(t)
	now := time.Now()
	files := map[string]time.Duration{"fresh.jsonl": 10 * time.Minute, "idle.jsonl": 2 * time.Hour}
	for name, age := range files {
		path := filepath.Join(tempDir, name)
		if err := os.WriteFile(path, nil, 0o600); err != nil {
			t.Fatal(err)
		}
		if err := os.Chtimes(path, now.Add(-age), now.Add(-age)); err != nil {
			t.Fatal(err)
		}
	}

	sweepUploads(now, time.Hour)

	if _, err := os.Stat(filepath.Join(tempDir, "fresh.jsonl")); err != nil {
		t.Errorf("fresh upload deleted: %v", err)
	}
	if _, err := os.Stat(filepath.Join(tempDir, "idle.jsonl")); !os.IsNotExist(err) {
		t.Errorf("idle upload kept: %v", err)
	}
}

func TestSweepUploadsWithoutDirectory(t *testing.T) {
	useTempDir(t)
	tempDir = filepath.Join(tempDir, "missing")
	sweepUploads(time.Now(), time.Hour) // must not panic or create the directory
	if _, err := os.Stat(tempDir); !os.IsNotExist(err) {
		t.Errorf("sweep created %s", tempDir)
	}
}

func TestAnalyzeRefreshesUploadTimestamp(t *testing.T) {
	useTempDir(t)
	id := upload(t).FileID
	path := filepath.Join(tempDir, id+".jsonl")
	old := time.Now().Add(-50 * time.Minute)
	if err := os.Chtimes(path, old, old); err != nil {
		t.Fatal(err)
	}

	for name, h := range map[string]http.HandlerFunc{"analyze": Analyze, "events": Events} {
		if err := os.Chtimes(path, old, old); err != nil {
			t.Fatal(err)
		}
		get(h, "/api/x?file="+id)
		info, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		if time.Since(info.ModTime()) > time.Minute {
			t.Errorf("%s did not refresh the upload timestamp", name)
		}
	}
}

func TestExpiredUploadMessage(t *testing.T) {
	useTempDir(t)
	rec := get(Analyze, "/api/analyze?file=deadbeef")
	if rec.Code != http.StatusNotFound || !strings.Contains(rec.Body.String(), "upload the file again") {
		t.Errorf("got %d %q", rec.Code, rec.Body)
	}
}
