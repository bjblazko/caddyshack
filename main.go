// CaddyShack is a self-hosted analytics dashboard for Caddy access logs.
package main

import (
	"embed"
	"flag"
	"io/fs"
	"log"
	"net/http"
	"time"

	"github.com/bjblazko/caddyshack/internal/geoip"
	"github.com/bjblazko/caddyshack/internal/handler"
)

// staticFiles holds the frontend, embedded so the binary runs on its own (ADR-002).
//
//go:embed static
var staticFiles embed.FS

// version is set at build time: go build -ldflags "-X main.version=v1.2.3".
var version = "dev"

const csp = "default-src 'none'; script-src 'self'; style-src 'self'; img-src 'self'; font-src 'self'; connect-src 'self'; frame-ancestors 'none'; base-uri 'self'; form-action 'self'"

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy", csp)
		next.ServeHTTP(w, r)
	})
}

func main() {
	addr := flag.String("addr", ":8080", "listen address")
	geodb := flag.String("geodb", "./data/dbip-country-lite.csv", "path to DB-IP country CSV")
	logDir := flag.String("logdir", "/var/log/caddy", "directory of server-side Caddy logs")
	uploadTTL := flag.Duration("upload-ttl", time.Hour, "delete uploaded logs not used for this long")
	flag.Parse()
	if *uploadTTL <= 0 {
		log.Fatal("-upload-ttl must be positive")
	}

	geoip.Load(*geodb)
	handler.SetLogDir(*logDir)
	handler.SetVersion(version)
	handler.StartUploadCleanup(*uploadTTL)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/upload", handler.Upload)
	mux.HandleFunc("GET /api/logs", handler.LogFiles)
	mux.HandleFunc("GET /api/analyze", handler.Analyze)
	mux.HandleFunc("GET /api/events", handler.Events)
	mux.HandleFunc("GET /api/health", handler.Health)
	static, err := fs.Sub(staticFiles, "static")
	if err != nil {
		log.Fatal(err)
	}
	mux.Handle("/", http.FileServer(http.FS(static)))

	log.Printf("CaddyShack listening on %s", *addr)
	log.Fatal(http.ListenAndServe(*addr, securityHeaders(mux)))
}
