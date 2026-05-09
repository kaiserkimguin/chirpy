package main

import (
	"log"
	"net/http"
)

func main() {
	ServeMux := http.NewServeMux()
	ServeMux.Handle("/app/", http.StripPrefix("/app", http.FileServer(http.Dir("."))))
	ServeMux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	server := &http.Server{
		Addr:    ":8080",
		Handler: ServeMux,
	}
	log.Fatal(server.ListenAndServe())
}
