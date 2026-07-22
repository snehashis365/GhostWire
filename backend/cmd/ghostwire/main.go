package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"ghostwire/backend/internal/db"
	"ghostwire/backend/internal/janitor"
	"ghostwire/backend/internal/server"
	"ghostwire/backend/internal/ws"
)

func main() {
	dbPath := env("GHOSTWIRE_DB", "ghostwire.db")
	staticDir := env("GHOSTWIRE_STATIC", "./public")
	addr := env("GHOSTWIRE_ADDR", ":8080")
	database, err := db.Open(dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()
	hub := ws.NewHub(database)
	go hub.Run()
	go janitor.New(database, 60*time.Second).Run(context.Background())
	log.Println("ghostwire listening on", addr)
	log.Fatal(http.ListenAndServe(addr, server.New(database, hub, staticDir).Handler()))
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
