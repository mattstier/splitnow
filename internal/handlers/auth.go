package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"splitnow/internal/token"
)

func RequireAuth(tm *token.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		tokenString, ok := strings.CutPrefix(header, "Bearer ")
		if !ok || tokenString == "" {
			// same as .JSON, but ensures that it breaks the chain, even if no return is called 
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		claims, err := tm.Verify(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		// make the caller's identity available to the handler
		c.Set("user", claims)

		// needed to continue to the next handler
		c.Next()
	}
}
