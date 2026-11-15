package main

import (
	"log"
	"net/http"
	"time"

	"github.com/Dima-Vasilenko/task-manager/internal/domain"
)

func main() {
	t := domain.CreateTasks()
	app := application{
		tasks:  t,
		nextID: len(t) + 1,
	}

	srv := &http.Server{
		Addr:              ":8080",
		Handler:           app.route(),
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
