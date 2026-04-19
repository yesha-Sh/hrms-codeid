package main

import (
	"log"
	"net/http"

	"hrms/backend/internal/config"
	"hrms/backend/internal/db"
	"hrms/backend/internal/server"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	gormDB, err := db.OpenGORM(cfg)
	if err != nil {
		log.Fatal(err)
	}

	srv, err := server.New(cfg, gormDB)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("hrms api listening on %s\n", cfg.Address())
	log.Fatal(http.ListenAndServe(cfg.Address(), srv.Router()))
}
