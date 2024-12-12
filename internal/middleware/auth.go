package middleware

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/yigit-demirko/go-ledger/internal/auth"
	"github.com/yigit-demirko/go-ledger/internal/models"
)

// AuthMiddleware checks if the user is logged in
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// look for the auth token in headers
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// check if token format is correct
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		// make sure token is valid
		claims, err := auth.ValidateToken(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// save user info for later use
		c.Set("user", claims)
		c.Next()
	}
}

// RequireRole makes sure user has permission to do something
func RequireRole(role models.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		// get user info we saved earlier
		claims, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			c.Abort()
			return
		}

		userClaims, ok := claims.(*auth.Claims)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user claims"})
			c.Abort()
			return
		}

		// check if user can do this
		if userClaims.Role != role && userClaims.Role != models.RoleAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireOwnershipOrAdmin checks if users are accessing their own stuff
func RequireOwnershipOrAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		// get user info we saved earlier
		claims, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			c.Abort()
			return
		}

		userClaims, ok := claims.(*auth.Claims)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user claims"})
			c.Abort()
			return
		}

		// admins can do anything
		if userClaims.Role == models.RoleAdmin {
			c.Next()
			return
		}

		// get the ID of user being accessed
		requestedUserIDStr := c.Param("id")
		if requestedUserIDStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "User ID required"})
			c.Abort()
			return
		}

		// convert ID to number for comparison
		requestedUserID, err := strconv.ParseInt(requestedUserIDStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
			c.Abort()
			return
		}

		// make sure users only access their own stuff
		if requestedUserID != userClaims.UserID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			c.Abort()
			return
		}

		c.Next()
	}
} 