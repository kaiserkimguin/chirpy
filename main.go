package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
)

func main() {
	cfg := &apiConfig{}
	ServeMux := http.NewServeMux()
	ServeMux.Handle("/app/", cfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))
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
	cfg.fileserverHits.Store(0)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("request-counter reset"))
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
		w.WriteHeader(500)
		return
	}
	// if request was successful: test body for length and respond accordingly
	type returnValid struct {
		Valid bool `json:"valid"`
	}

	type returnError struct {
		Error string `json:"error"`
	}

	if len(cB.Body) < 1 {
		respondWithError(w, 400, "chirp cannot be empty")
	} else if len(cB.Body) > 140 {
		respondWithError(w, 400, "chirp is too long")
	} else {
		rV := returnValid{Valid: true}
		respondWithJSON(w, 200, rV)
	}
}
