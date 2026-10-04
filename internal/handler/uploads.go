package handler

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"path/filepath"
)

var tempDir = filepath.Join(os.TempDir(), "caddyshack")

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
