package logparser

import (
	"bufio"
	"encoding/json"
	"io"
)

type TLSInfo struct {
	Resumed    bool   `json:"resumed"`
	Version    int    `json:"version"`
	CipherSuite int  `json:"cipher_suite"`
	Proto      string `json:"proto"`
	ServerName string `json:"server_name"`
}

type Request struct {
	RemoteIP string              `json:"remote_ip"`
	ClientIP string              `json:"client_ip"`
	Proto    string              `json:"proto"`
	Method   string              `json:"method"`
	Host     string              `json:"host"`
	URI      string              `json:"uri"`
	Headers  map[string][]string `json:"headers"`
	TLS      *TLSInfo            `json:"tls"`
}

type LogEntry struct {
	Level     string  `json:"level"`
	Timestamp float64 `json:"ts"`
	Logger    string  `json:"logger"`
	Msg       string  `json:"msg"`
	Request   Request `json:"request"`
	BytesRead int64   `json:"bytes_read"`
	UserID    string  `json:"user_id"`
	Duration  float64 `json:"duration"`
	Size      int64   `json:"size"`
	Status    int     `json:"status"`
}

// ParseStream reads JSONL log data line by line and calls fn for each parsed entry.
// Malformed lines are silently skipped.
func ParseStream(r io.Reader, fn func(LogEntry)) {
	scanner := bufio.NewScanner(r)
	buf := make([]byte, 0, 1024*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var entry LogEntry
		if err := json.Unmarshal(line, &entry); err != nil {
			continue
		}
		fn(entry)
	}
}
