// CaddyShack is a self-hosted analytics dashboard for Caddy access logs.
package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/bjblazko/caddyshack/internal/geoip"
	"github.com/bjblazko/caddyshack/internal/handler"
)

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
	uploadTTL := flag.Duration("upload-ttl", time.Hour, "delete uploaded logs not used for this long")
	flag.Parse()
	if *uploadTTL <= 0 {
		log.Fatal("-upload-ttl must be positive")
	}

	geoip.Load(*geodb)
	handler.StartUploadCleanup(*uploadTTL)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/upload", handler.Upload)
	mux.HandleFunc("GET /api/logs", handler.LogFiles)
	mux.HandleFunc("GET /api/analyze", handler.Analyze)
	mux.HandleFunc("GET /api/events", handler.Events)
	mux.HandleFunc("GET /api/health", handler.Health)
	mux.Handle("/", http.FileServer(http.Dir("static")))

	log.Printf("CaddyShack listening on %s", *addr)
	log.Fatal(http.ListenAndServe(*addr, securityHeaders(mux)))
}
