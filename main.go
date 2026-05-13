package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/google/uuid"
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
	ServeMux.HandleFunc("GET /api/healthz", handlerHealthz)
	ServeMux.HandleFunc("POST /api/chirps", cfg.handlerPostChirp)
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

func (cfg *apiConfig) handlerPostChirp(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body   string    `json:"body"`
		UserID uuid.UUID `json:"user_id"`
	}
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		w.WriteHeader(400)
		return
	}
	params.Body, err = getCleanedBody(params.Body)
	if err != nil {
		respondWithError(w, 400, err.Error())
	}
	chi, err := cfg.dbQueries.CreatePost(r.Context(), database.CreatePostParams{
		Body:   params.Body,
		UserID: params.UserID,
	})
	if err != nil {
		respondWithError(w, 400, err.Error())
	}
	chirpResponse := Chirp{
		ID:        chi.ID,
		CreatedAt: chi.CreatedAt,
		UpdatedAt: chi.UpdatedAt,
		Body:      chi.Body,
		UserID:    chi.UserID,
	}
	respondWithJSON(w, 201, chirpResponse)
}
