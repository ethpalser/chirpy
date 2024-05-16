package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

type Chirp struct {
	Id      int    `json:"id"`
	Message string `json:"body"`
}

type User struct {
	Id       int    `json:"id"`
	Email    string `json:"email"`
	Password string `json:"password,omitempty"`
}

func main() {
	godotenv.Load()
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	secret := os.Getenv("JWT_SECRET")

	dbg := flag.Bool("debug", false, "Enable debug mode")
	flag.Parse()
	if *dbg {
		os.WriteFile(connectionString, []byte("{}"), 0777)
	}
	db, err := NewDB(connectionString)
	if err != nil {
		print(err)
		return
	}
	// Create a multiplexer that can handle HTTP requests for a server at its endpoints
	mux := http.NewServeMux()
	handler := http.StripPrefix("/app", http.FileServer(http.Dir(".")))
	apiCfg := apiConfig{
		fileserverHits: 0,
		jwtSecret:      secret,
	}
	// Creates a Handler that handles HTTP requests at the path and returns system files
	mux.Handle("/app/*", apiCfg.middlewareMetricsInc(handler))
	mux.HandleFunc("GET /api/healthz", health)
	mux.HandleFunc("GET /admin/metrics", apiCfg.hits)
	mux.HandleFunc("/api/reset", apiCfg.reset)
	mux.Handle("GET /api/chirps/{chirpID}", getOneChirp(*db))
	mux.Handle("GET /api/chirps", getAllChirps(*db))
	mux.Handle("POST /api/chirps", postChirp(*db))
	mux.Handle("POST /api/users", postUser(*db))
	mux.Handle("POST /api/login", login(*db))
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
	jwtSecret      string
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits++
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) hits(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	template := `<html>

<body>
	<h1>Welcome, Chirpy Admin</h1>
	<p>Chirpy has been visited %d times!</p>
</body>

</html>
`
	w.Write([]byte(fmt.Sprintf(template, cfg.fileserverHits)))
}

func (cfg *apiConfig) reset(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits = 0
	w.WriteHeader(http.StatusOK)
}

func responseWithError(w http.ResponseWriter, code int, msg string) {
	if code > 499 {
		log.Printf("Responding with 5xx error: %s", msg)
	}
	type returnErr struct {
		Error string `json:"error"`
	}
	responseWithJSON(w, code, returnErr{
		Error: msg,
	})
}

func responseWithJSON(w http.ResponseWriter, code int, body interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json, err := json.Marshal(body)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		w.Write([]byte(`{"error":"Something went wrong"}`))
		return
	}
	w.WriteHeader(code)
	w.Write(json)
}

func filterProfanity(msg string) string {
	profanity := map[string]bool{
		"kerfuffle": true,
		"sharbert":  true,
		"fornax":    true,
	}
	msg_split := strings.Split(msg, " ")

	lower := strings.ToLower(msg)
	lower_split := strings.Split(lower, " ")
	for index, word := range lower_split {
		_, ok := profanity[word]
		if ok {
			msg_split[index] = "****"
		}
	}
	return strings.Join(msg_split, " ")
}

func getOneChirp(db DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		enErr := db.ensureDB()
		if enErr != nil {
			responseWithError(w, 500, enErr.Error())
			return
		}
		pathVal := r.PathValue("chirpID")
		idVal, atoiErr := strconv.Atoi(pathVal)
		if atoiErr != nil {
			responseWithError(w, 500, atoiErr.Error())
			return
		}
		json, getErr := db.GetChirp(idVal)
		if getErr != nil {
			responseWithError(w, 404, getErr.Error())
			return
		}
		responseWithJSON(w, 200, json)
	})
}

func getAllChirps(db DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		enErr := db.ensureDB()
		if enErr != nil {
			responseWithError(w, 500, enErr.Error())
			return
		}
		json, getErr := db.GetChirps()
		if getErr != nil {
			responseWithError(w, 500, getErr.Error())
			return
		}
		responseWithJSON(w, 200, json)
	})
}

func postChirp(db DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Parse
		type parameters struct {
			Body string `json:"body"`
		}

		decoder := json.NewDecoder(r.Body)
		params := parameters{}
		err := decoder.Decode(&params)
		if err != nil {
			log.Printf("Error decoding parameters: %s", err)
			responseWithError(w, 500, "Something went wrong")
			return
		}

		if len(params.Body) > 140 {
			responseWithError(w, 400, "Chirp is too long")
			return
		}
		// Validate
		cleaned := filterProfanity(params.Body)
		// Save
		enErr := db.ensureDB()
		if enErr != nil {
			responseWithError(w, 500, enErr.Error())
			return
		}
		json, getErr := db.CreateChirp(cleaned)
		if getErr != nil {
			responseWithError(w, 500, getErr.Error())
			return
		}
		responseWithJSON(w, 201, json)
	})
}

func postUser(db DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Parse
		type parameters struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}

		decoder := json.NewDecoder(r.Body)
		params := parameters{}
		err := decoder.Decode(&params)

		if err != nil {
			log.Printf("Error decoding parameters: %s", err)
			responseWithError(w, 500, "Something went wrong")
			return
		}

		hashpass, bcErr := bcrypt.GenerateFromPassword([]byte(params.Password), 10)
		if bcErr != nil {
			log.Printf("Error hashing password.")
			responseWithError(w, 500, "Something went wrong")
			return
		}

		// Save
		enErr := db.ensureDB()
		if enErr != nil {
			responseWithError(w, 500, enErr.Error())
			return
		}
		json, getErr := db.CreateUser(params.Email, string(hashpass))
		if getErr != nil {
			responseWithError(w, 500, getErr.Error())
			return
		}
		responseWithJSON(w, 201, json)
	})
}

func login(db DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Parse
		type parameters struct {
			Email         string `json:"email"`
			Password      string `json:"password"`
			ExpireSeconds int    `json:"expires_in_seconds"`
		}

		decoder := json.NewDecoder(r.Body)
		params := parameters{}
		err := decoder.Decode(&params)

		if err != nil {
			log.Printf("Error decoding parameters: %s", err)
			responseWithError(w, 500, "Something went wrong")
			return
		}

		enErr := db.ensureDB()
		if enErr != nil {
			responseWithError(w, 500, enErr.Error())
			return
		}
		json, getErr := db.Login(params.Email, params.Password, params.ExpireSeconds)
		json.Password = "" // Work-around to exclude in response

		if getErr != nil {
			responseWithError(w, 401, getErr.Error())
			return
		}
		responseWithJSON(w, 200, json)
	})
}
