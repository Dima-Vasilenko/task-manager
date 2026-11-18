package main

import (
	"context"
	"database/sql"
	"log"

	"net/http"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	connString := "postgres://task_manager_app:t1a2s3k4@localhost:5432/task_manager"

	db, err := sql.Open("pgx", connString)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		log.Fatal(err)
	}

	log.Println("Successfully connected to PostgreSQL")

	app := NewApplication(db)

	srv := &http.Server{
		Addr:              ":8080",
		Handler:           app.Route(),
		IdleTimeout:       time.Minute,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      30 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
	}

	err2 := srv.ListenAndServe()
	if err2 != nil {
		log.Fatal(err2)
	}
}
