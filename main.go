package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	"github.com/aelpxy/krofi/handlers"
	"github.com/aelpxy/krofi/middlewares"
	"github.com/aelpxy/krofi/utils"
)

func main() {
	mux := http.NewServeMux()
	cleanupInterval := time.Minute * 30
	cleanupTicker := time.NewTicker(cleanupInterval)

	mux.HandleFunc("/health", handlers.HealthStats)
	mux.HandleFunc("/serve/image", handlers.ServeWebPImage)

	handler := middlewares.CustomHeaders(mux)

	server := &http.Server{
		Addr:    ":8080",
		Handler: handler,
	}

	go func() {
		for range cleanupTicker.C {
			utils.PurgeCache()
			log.Println("Cache purged from server")
		}
	}()

	log.Println("Listening on PORT 8080")
	err := server.ListenAndServe()
	if err != nil {
		fmt.Println(err)
	}
}
