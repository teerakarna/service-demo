package main

import (
	"log"
	"os"

	"github.com/teerakarna/service-demo/internal/api"
	"github.com/teerakarna/service-demo/internal/store"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	s := store.NewMemoryStore()
	r := api.NewRouter(s)

	log.Printf("starting service-demo on :%s (env=%s)", port, os.Getenv("ENV_NAME"))
	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
