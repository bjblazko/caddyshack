package handler

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

var tempDir = filepath.Join(os.TempDir(), "caddyshack")

// StartUploadCleanup empties the upload directory and starts a background
// sweeper that deletes uploads not used for longer than ttl (see ADR-010).
func StartUploadCleanup(ttl time.Duration) {
	if err := os.RemoveAll(tempDir); err != nil {
		log.Printf("Error clearing upload directory: %v", err)
	}
	go func() {
		for now := range time.Tick(max(ttl/4, time.Second)) {
			sweepUploads(now, ttl)
		}
	}()
}

// sweepUploads deletes uploads whose last use is more than ttl before now.
func sweepUploads(now time.Time, ttl time.Duration) {
	entries, err := os.ReadDir(tempDir)
	if err != nil {
		return // directory not created yet: nothing uploaded
	}
	for _, e := range entries {
		info, err := e.Info()
		if err != nil || now.Sub(info.ModTime()) <= ttl {
			continue
		}
		if err := os.Remove(filepath.Join(tempDir, e.Name())); err != nil {
			log.Printf("Error deleting expired upload: %v", err)
		}
	}
}

// touchUpload marks an upload as used now, postponing its expiry.
func touchUpload(path string) {
	now := time.Now()
	if err := os.Chtimes(path, now, now); err != nil {
		log.Printf("Error refreshing upload timestamp: %v", err)
	}
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
	path := filepath.Join(tempDir, id+".jsonl")
	f, err := os.Create(path)
	if err != nil {
		return "", err
	}
	_, copyErr := io.Copy(f, r)
	// Close flushes the write; its error means the stored file may be incomplete.
	closeErr := f.Close()
	if err := errors.Join(copyErr, closeErr); err != nil {
		_ = os.Remove(path)
		return "", err
	}
	return id, nil
}
