package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/StortM/Structura/internal/api"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	router := api.NewRouter()

	staticDir := filepath.Join(".", "web", "static")
	fs := http.FileServer(http.Dir(staticDir))
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs))

	log.Printf("Server starting on port %s", port)
	err := http.ListenAndServe(":"+port, router)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
