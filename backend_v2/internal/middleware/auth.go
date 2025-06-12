package middleware

import (
	"github.com/gin-gonic/gin"

	"backend_v2/internal/config"
)

// AuthMiddleware creates an authentication middleware
func AuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip auth for health check
		if c.Request.URL.Path == "/health" || c.Request.URL.Path == "/api/health" {
			c.Next()
			return
		}

		// // Get authorization header
		// authHeader := c.GetHeader("Authorization")
		// if authHeader == "" {
		// 	c.JSON(http.StatusUnauthorized, models.ErrorResponse("Missing authorization header"))
		// 	c.Abort()
		// 	return
		// }

		// // Check Bearer token format
		// parts := strings.Split(authHeader, " ")
		// if len(parts) != 2 || parts[0] != "Bearer" {
		// 	c.JSON(http.StatusUnauthorized, models.ErrorResponse("Invalid authorization header format"))
		// 	c.Abort()
		// 	return
		// }

		// token := parts[1]

		// // Validate API key
		// if cfg.APIKey != "" && token != cfg.APIKey {
		// 	c.JSON(http.StatusUnauthorized, models.ErrorResponse("Invalid API key"))
		// 	c.Abort()
		// 	return
		// }

		c.Next()
	}
}
