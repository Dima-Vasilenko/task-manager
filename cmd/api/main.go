package main

import (
	"log"
	"net/http"
	"time"
)

func main() {
	app := NewApplication()

	srv := &http.Server{
		Addr:              ":8080",
		Handler:           app.Route(),
		IdleTimeout:       time.Minute,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      30 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
	}

	err := srv.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
