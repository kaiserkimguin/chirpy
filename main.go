package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/joho/godotenv"
	"github.com/kaiserkimguin/chirpy/internal/database"
	_ "github.com/lib/pq"
)

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
		return
	}
	cfg := &apiConfig{}
	cfg.dbQueries = database.New(db)
	cfg.platform = os.Getenv("PLATFORM")
	ServeMux := http.NewServeMux()
	ServeMux.Handle("/app/", cfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))
	ServeMux.HandleFunc("POST /api/users", cfg.handlerApiUsers)
	ServeMux.HandleFunc("GET /admin/metrics/", cfg.handlerMetrics)
	ServeMux.HandleFunc("POST /admin/reset/", cfg.handlerReset)
	ServeMux.HandleFunc("POST /api/validate_chirp/", handlerValidateChirp)
	ServeMux.HandleFunc("GET /api/healthz", handlerHealthz)
	server := &http.Server{
		Addr:    ":8080",
		Handler: ServeMux,
	}
	log.Fatal(server.ListenAndServe())
}

type apiConfig struct {
	fileserverHits atomic.Int32
	dbQueries      *database.Queries
	platform       string
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func handlerHealthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (cfg *apiConfig) handlerMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>`, cfg.fileserverHits.Load())
}

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
	if cfg.platform != "dev" {
		respondWithError(w, 403, "Forbidden")
	}
	_, err := cfg.dbQueries.DeleteUsers(r.Context())
	if err != nil {
		respondWithError(w, 500, "unable to reset users")
	}
	cfg.fileserverHits.Store(0)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("request-counter & users reset"))
}

func handlerValidateChirp(w http.ResponseWriter, r *http.Request) {
	type chirpBody struct {
		Body string `json:"body"`
	}
	decoder := json.NewDecoder(r.Body)
	cB := chirpBody{}
	err := decoder.Decode(&cB)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		w.WriteHeader(400)
		return
	}
	// if request was successful: test body for length and respond accordingly
	type returnValid struct {
		Valid       bool   `json:"valid"`
		CleanedBody string `json:"cleaned_body"`
	}

	// clean the body before accepting the post.
	cB.Body = getCleanedBody(cB.Body)
	rV := returnValid{CleanedBody: cB.Body}

	if len(cB.Body) < 1 {
		respondWithError(w, 400, "chirp cannot be empty")
	} else if len(cB.Body) > 140 {
		respondWithError(w, 400, "chirp is too long")
	} else {
		rV.Valid = true
		respondWithJSON(w, 200, rV)
	}
}

func (cfg *apiConfig) handlerApiUsers(w http.ResponseWriter, r *http.Request) {
	// decode request body
	type parameters struct {
		Email string `json:"email"`
	}
	params := parameters{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 400, "bad request")
		return
	}
	// create and write response
	u, err := cfg.dbQueries.CreateUser(r.Context(), params.Email)
	if err != nil {
		respondWithError(w, 500, "unable to create user")
		return
	}
	jsonUser := User{
		ID:        u.ID,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
		Email:     u.Email,
	}
	respondWithJSON(w, 201, jsonUser)
}
