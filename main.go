package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/ethpalser/chirpy/internal/database"
	"github.com/joho/godotenv"
)

type apiConfig struct {
	fileserverHits int
	database       database.DB
	jwtSecret      string
	polkaApiKey    string
}

func main() {
	godotenv.Load()
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	jwtSecret := os.Getenv("JWT_SECRET")
	dbSource := os.Getenv("DB_SOURCE")
	polkaApiKey := os.Getenv("POLKA_API_KEY")

	db, err := database.NewDB(dbSource)
	if err != nil {
		log.Fatal(err)
	}

	dbg := flag.Bool("debug", false, "Enable debug mode")
	flag.Parse()
	if dbg != nil && *dbg {
		dbErr := db.ResetDB()
		if dbErr != nil {
			log.Fatal(dbErr)
		}
	}

	apiCfg := apiConfig{
		fileserverHits: 0,
		database:       *db,
		jwtSecret:      jwtSecret,
		polkaApiKey:    polkaApiKey,
	}

	// Create a multiplexer that can handle HTTP requests for a server at its endpoints
	mux := http.NewServeMux()
	handler := apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir("."))))
	mux.Handle("/app/*", handler)
	// General APIs
	mux.HandleFunc("GET /api/healthz", health)
	mux.HandleFunc("GET /api/reset", apiCfg.handlerMetricsReset)
	// User APIs
	mux.HandleFunc("POST /api/users", apiCfg.handlerUsersCreate)
	mux.HandleFunc("PUT /api/users", apiCfg.handlerUsersUpdate)
	mux.HandleFunc("POST /api/login", apiCfg.handlerLogin)
	// Chirp APIs
	mux.HandleFunc("GET /api/chirps", apiCfg.handlerChirpsGetAll)
	mux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.handlerChirpsGetOne)
	mux.HandleFunc("POST /api/chirps", apiCfg.handlerChirpsCreate)
	mux.HandleFunc("DELETE /api/chirps/{chirpID}", apiCfg.handlerChirpsDelete)
	// Token APIs
	mux.HandleFunc("POST /api/refresh", apiCfg.handlerTokenRefresh)
	mux.HandleFunc("POST /api/revoke", apiCfg.handlerTokenRevoke)
	// Admin APIs
	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerMetrics)
	// Webhooks
	mux.HandleFunc("POST /api/polka/webhooks", apiCfg.webhookPolka)
	// Update the multiplexer to accept CORS data
	corsMux := middlewareCors(mux)
	// Setup a server that uses the new multiplexer
	server := &http.Server{
		Addr:    "localhost:8080",
		Handler: corsMux,
	}
	log.Fatal(server.ListenAndServe())
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
