package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type HealthCheck struct {
	Status     string `json:"status"`
	ResponseMs int64  `json:"response_ms"`
}

func HealthStats(c *gin.Context) {
	start := time.Now()

	hc := HealthCheck{Status: "OK"}
	hc.ResponseMs = time.Since(start).Milliseconds()

	c.JSON(http.StatusOK, hc)
}
