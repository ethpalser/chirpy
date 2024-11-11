package main

import _ "github.com/lib/pq"

import (
	"flag"
	"log"
	"net/http"
	"os"
	"database/sql"

	database "github.com/ethpalser/chirpy/internal/database"
	database2 "github.com/ethpalser/chirpy/internal/database/v2"
	"github.com/joho/godotenv"
)

type apiConfig struct {
	fileserverHits 	int
	database	database.DB
	dbQueries	*database2.Queries
	jwtSecret      	string
	polkaApiKey    	string
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	jwtSecret := os.Getenv("JWT_SECRET")
	dbSource := os.Getenv("DB_SOURCE")
	dbURL := os.Getenv("DB_URL")
	polkaApiKey := os.Getenv("POLKA_API_KEY")

	db, err := database.NewDB(dbSource)
	if err != nil {
		log.Fatal(err)
	}

	// for sqlc and goose update
	db2, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}
	dbQueries := database2.New(db2)

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
		dbQueries:	dbQueries,
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
	// Admin APIs
	mux.HandleFunc("POST /admin/reset", apiCfg.handlerReset)
	// User APIs
	//	mux.HandleFunc("POST /api/users", apiCfg.handlerUsersCreate)
	mux.HandleFunc("POST /api/users", apiCfg.handlerUsersCreateV2)
	mux.HandleFunc("PUT /api/users", apiCfg.handlerUsersUpdate)
	mux.HandleFunc("POST /api/login", apiCfg.handlerLogin)
	// Chirp APIs
	mux.HandleFunc("GET /api/chirps", apiCfg.handlerChirpsGetAll)
	mux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.handlerChirpsGetOne)
	//	mux.HandleFunc("POST /api/chirps", apiCfg.handlerChirpsCreate)
	mux.HandleFunc("POST /api/chirps", apiCfg.handlerChirpsCreateV2)
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
