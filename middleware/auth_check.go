package middleware

import (
	"go-transaction/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthCheck is a middleware function that validates the Authorization header in the incoming request.
// 
// It checks for the presence of the Authorization header, ensures it follows the "Bearer <token>" format,
// and validates the token using the `ValidateToken` function from the utils package.
// 
// If the Authorization header is missing or invalid, or the token is invalid, it responds with a 401 Unauthorized status.
// 
// If the token is valid, it allows the request to proceed by calling c.Next().
func AuthCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Retrieve the Authorization header from the request
		authHeader := c.GetHeader("Authorization")
		
		// Check if the Authorization header is missing
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is missing"})
			c.Abort()
			return
		}

		// Ensure the Authorization header follows the "Bearer <token>" format
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Authorization header format"})
			c.Abort()
			return
		}

		// Extract the token from the Authorization header
		token := authHeader[len("Bearer "):]
		
		// Validate the token using the utils.ValidateToken function
		_, err := utils.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Proceed to the next handler if the token is valid
		c.Next()
	}
}
