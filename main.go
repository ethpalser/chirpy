package main

import (
	"fmt"
	"net/http"
)

func main() {
	// Create a multiplexer that can handle HTTP requests for a server at its endpoints
	mux := http.NewServeMux()
	handler := http.StripPrefix("/app", http.FileServer(http.Dir(".")))
	apiCfg := apiConfig{}
	// Creates a Handler that handles HTTP requests at the path and returns system files
	mux.Handle("/app/*", apiCfg.middlewareMetricsInc(handler))
	mux.HandleFunc("/healthz", health)
	mux.HandleFunc("/metrics", apiCfg.hits)
	mux.HandleFunc("/reset", apiCfg.reset)
	// Update the multiplexer to accept CORS data
	corsMux := middlewareCors(mux)
	// Setup a server that uses the new multiplexer
	server := http.Server{Addr: "localhost:8080", Handler: corsMux}
	server.ListenAndServe()
}

func middlewareCors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func health(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

type apiConfig struct {
	fileserverHits int
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	cfg.fileserverHits += 1
	return next
}

func (cfg *apiConfig) hits(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hits: " + fmt.Sprint(cfg.fileserverHits)))
}

func (cfg *apiConfig) reset(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits = 0
	w.WriteHeader(http.StatusOK)
}
