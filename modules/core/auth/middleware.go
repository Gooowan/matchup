package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	coremodels "github.com/Gooowan/matchup/modules/core/gen"
)

const (
	AuthorizationHeader = "Authorization"
	BearerPrefix        = "Bearer "
	UserContextKey      = "user"
)

func JWTMiddleware(authService *AuthService) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		var tokenString string

		// First try to get token from Authorization header
		authHeader := c.GetHeader(AuthorizationHeader)
		if authHeader != "" && strings.HasPrefix(authHeader, BearerPrefix) {
			tokenString = strings.TrimPrefix(authHeader, BearerPrefix)
		} else {
			// Fallback to cookie
			cookie, err := c.Cookie("auth_token")
			if err != nil || cookie == "" {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing authentication token"})
				c.Abort()
				return
			}
			tokenString = cookie
		}

		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing token"})
			c.Abort()
			return
		}

		user, err := authService.ValidateJwtToken(c.Request.Context(), tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		// Store user in context
		c.Set(UserContextKey, user)
		c.Next()
	})
}

func GetUserFromContext(c *gin.Context) (*coremodels.User, bool) {
	user, exists := c.Get(UserContextKey)
	if !exists {
		return nil, false
	}

	userModel, ok := user.(*coremodels.User)
	return userModel, ok
}

func RequireAuth(authService *AuthService) gin.HandlerFunc {
	return JWTMiddleware(authService)
}

func RequireRole(authService *AuthService, requiredRole string) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		var tokenString string

		authHeader := c.GetHeader(AuthorizationHeader)
		if authHeader != "" && strings.HasPrefix(authHeader, BearerPrefix) {
			tokenString = strings.TrimPrefix(authHeader, BearerPrefix)
		} else {
			cookie, err := c.Cookie("auth_token")
			if err != nil || cookie == "" {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing authentication token"})
				c.Abort()
				return
			}
			tokenString = cookie
		}

		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing token"})
			c.Abort()
			return
		}

		user, err := authService.ValidateJwtToken(c.Request.Context(), tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		if user.Role != requiredRole {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			c.Abort()
			return
		}

		c.Set(UserContextKey, user)
		c.Next()
	})
}

func RequireAdmin(authService *AuthService) gin.HandlerFunc {
	return RequireRole(authService, "ADMIN")
}
