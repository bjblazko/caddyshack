package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/bjblazko/caddyshack/internal/geoip"
	"github.com/bjblazko/caddyshack/internal/handler"
)

func main() {
	addr := flag.String("addr", ":8080", "listen address")
	geodb := flag.String("geodb", "./data/dbip-country-lite.csv", "path to DB-IP country CSV")
	flag.Parse()

	geoip.Load(*geodb)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/upload", handler.Upload)
	mux.HandleFunc("GET /api/logs", handler.LogFiles)
	mux.HandleFunc("GET /api/analyze", handler.Analyze)
	mux.HandleFunc("GET /api/health", handler.Health)
	mux.Handle("/", http.FileServer(http.Dir("static")))

	log.Printf("CaddyShack listening on %s", *addr)
	log.Fatal(http.ListenAndServe(*addr, mux))
}
