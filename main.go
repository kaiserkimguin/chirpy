package main

import (
	"log"
	"net/http"
)

func main() {
	ServeMux := http.NewServeMux()
	ServeMux.Handle("/", http.FileServer(http.Dir(".")))
	server := &http.Server{
		Addr:    ":8080",
		Handler: ServeMux,
	}
	log.Fatal(server.ListenAndServe())
}
