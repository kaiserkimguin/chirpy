package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("unable to marhal response: %s", err)
		w.WriteHeader(500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write([]byte(data))
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	type errorBody struct {
		Error string `json:"error"`
	}

	respondWithJSON(w, code, errorBody{Error: msg})
}
