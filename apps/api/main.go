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

	log.Printf("LifeOps Go spike listening on %s", addr)
	if err := http.ListenAndServe(addr, NewServer(token)); err != nil {
		log.Fatal(err)
	}
}
