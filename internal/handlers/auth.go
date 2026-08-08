package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"splitnow/internal/token"
)

// TokenExtractor enables injecting the a strategy depending on
// the protocol of the endpoint
type TokenExtractor func(c *gin.Context) (string, bool)

// token extraction strategy for REST endpoint(s), extracts token from the header
func ExtractFromHeader(c *gin.Context) (string, bool) {
	return strings.CutPrefix(c.GetHeader("Authorization"), "Bearer ")
}

// token extraction strategy for WS endpoint(s), extracts token from the query param 
func ExtractFromQuery(c *gin.Context) (string, bool) {
	token := c.Query("token")
	return token, token != ""
}

func RequireAuth(tm *token.Manager, extractToken TokenExtractor) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, ok := extractToken(c) 
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
