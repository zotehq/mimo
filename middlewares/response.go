package middlewares

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

func ResponseTimeMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		responseTime := time.Since(start).Milliseconds()

		c.Next()

		c.Header("X-Response-Time", fmt.Sprintf("%d ms", responseTime))
	}
}
