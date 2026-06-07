package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	token := os.Getenv("LIFEOPS_WEBHOOK_TOKEN")
	if token == "" {
		token = "dev-token"
	}
	addr := os.Getenv("LIFEOPS_ADDR")
	if addr == "" {
		addr = ":8080"
	}
	dbPath := os.Getenv("LIFEOPS_DB_PATH")
	if dbPath == "" {
		dbPath = "lifeops.db"
	}

	store, err := NewStore(dbPath)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer store.Close()

	log.Printf("LifeOps API listening on %s (db: %s)", addr, dbPath)
	if err := http.ListenAndServe(addr, NewServer(token, store)); err != nil {
		log.Fatal(err)
	}
}
