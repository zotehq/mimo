package middlewares

import (
	"github.com/gin-gonic/gin"
)

func CustomHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Server", "Krofi")
		c.Next()
	}
}
