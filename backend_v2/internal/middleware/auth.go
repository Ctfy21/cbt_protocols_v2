package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"backend_v2/internal/models"
	"backend_v2/internal/services"
)

// AuthMiddleware creates an authentication middleware
func AuthMiddleware(authService *services.AuthService, apiTokenService *services.APITokenService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip auth for certain paths
		path := c.Request.URL.Path
		if isPublicPath(path) {
			c.Next()
			return
		}

		// Get authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse("Missing authorization header"))
			c.Abort()
			return
		}

		// Check Bearer token format
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse("Invalid authorization header format"))
			c.Abort()
			return
		}

		token := parts[1]

		var user *models.User
		var err error

		// Check if it's an API token (starts with "svc_")
		if strings.HasPrefix(token, "svc_") {
			// Validate API token
			var apiToken *models.APIToken
			user, apiToken, err = apiTokenService.ValidateAPIToken(token)
			if err != nil {
				c.JSON(http.StatusUnauthorized, models.ErrorResponse(err.Error()))
				c.Abort()
				return
			}
			// Store API token info in context
			c.Set("api_token", apiToken)
			c.Set("auth_type", "api_token")
		} else {
			// Validate JWT token
			user, err = authService.ValidateToken(token)
			if err != nil {
				c.JSON(http.StatusUnauthorized, models.ErrorResponse(err.Error()))
				c.Abort()
				return
			}
			c.Set("auth_type", "jwt")
		}

		// Store user in context
		c.Set("user", user)
		c.Set("user_id", user.ID.Hex())
		c.Set("user_role", string(user.Role))

		c.Next()
	}
}

// RequireRole creates a middleware that requires specific roles
func RequireRole(roles ...models.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRoleStr, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusForbidden, models.ErrorResponse("Access denied"))
			c.Abort()
			return
		}

		userRole := models.UserRole(userRoleStr.(string))

		// Check if user has one of the required roles
		hasRole := false
		for _, role := range roles {
			if userRole == role {
				hasRole = true
				break
			}
		}

		if !hasRole {
			c.JSON(http.StatusForbidden, models.ErrorResponse("Insufficient permissions"))
			c.Abort()
			return
		}

		c.Next()
	}
}

// OptionalAuth creates an optional authentication middleware
func OptionalAuth(authService *services.AuthService, apiTokenService *services.APITokenService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		// Check Bearer token format
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.Next()
			return
		}

		token := parts[1]

		var user *models.User
		var err error

		// Check if it's an API token (starts with "svc_")
		if strings.HasPrefix(token, "svc_") {
			// Validate API token
			var apiToken *models.APIToken
			user, apiToken, err = apiTokenService.ValidateAPIToken(token)
			if err == nil && user != nil {
				// Store API token info in context
				c.Set("api_token", apiToken)
				c.Set("auth_type", "api_token")
			}
		} else {
			// Validate JWT token
			user, err = authService.ValidateToken(token)
			if err == nil && user != nil {
				c.Set("auth_type", "jwt")
			}
		}

		if err == nil && user != nil {
			// Store user in context if valid
			c.Set("user", user)
			c.Set("user_id", user.ID.Hex())
			c.Set("user_role", string(user.Role))
		}

		c.Next()
	}
}

// isPublicPath checks if a path is public (doesn't require authentication)
func isPublicPath(path string) bool {
	publicPaths := []string{
		"/health",
		"/api/health",
		"/auth/login",
		"/auth/register",
		"/auth/refresh",
	}

	for _, publicPath := range publicPaths {
		if path == publicPath {
			return true
		}
	}

	return false
}
