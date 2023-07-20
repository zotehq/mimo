package main

import (
	"fmt"
	"log"
	"time"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	"github.com/aelpxy/krofi/handlers"
	"github.com/aelpxy/krofi/middlewares"
	"github.com/aelpxy/krofi/utils"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.Use(middlewares.CustomHeaders())

	r.GET("/health", handlers.HealthStats)
	r.GET("/image/resize", handlers.ResizeImage)

	cleanupInterval := time.Minute * 30

	go func() {
		for range time.Tick(cleanupInterval) {
			utils.PurgeCache()
			log.Println("Cache purged from server")
		}
	}()

	err := r.Run(":8080")
	if err != nil {
		fmt.Println(err)
	}
}
