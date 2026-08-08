package handlers

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"splitnow/internal/token"
)

// generates a key pair in memory and returns a manager loaded with the public
// key, so the middleware can be tested without key files on disk
func newTestManager(t *testing.T) (*token.Manager, *rsa.PrivateKey) {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	publicDER, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	require.NoError(t, err)

	publicPath := filepath.Join(t.TempDir(), "public.pem")
	require.NoError(t, os.WriteFile(publicPath,
		pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: publicDER}), 0600))

	mgr, err := token.New(publicPath)
	require.NoError(t, err)

	return mgr, privateKey
}

// signs claims with the private key so the token verifies against the test manager
func sign(t *testing.T, privateKey *rsa.PrivateKey, claims token.Claims) string {
	t.Helper()
	tokenString, err := jwt.NewWithClaims(jwt.SigningMethodRS256, claims).SignedString(privateKey)
	require.NoError(t, err)
	return tokenString
}

// middleware must reject any request without a usable token in the header
func TestRequireAuthRejectsInvalidTokens(t *testing.T) {
	mgr, key := newTestManager(t)

	expired := sign(t, key, token.Claims{
		UserID: 1, Username: "foo",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
		},
	})

	tampered := sign(t, key, token.Claims{UserID: 1, Username: "foo"})
	parts := strings.Split(tampered, ".")
	payload := []byte(parts[1])
	payload[0] ^= 0xFF
	parts[1] = string(payload)
	tampered = strings.Join(parts, ".")

	tests := []struct {
		name   string
		header string
	}{
		{"missing header", ""},
		{"wrong scheme", "Token 123"},
		{"empty token", "Bearer "},
		{"garbage token", "Bearer garbage.not.a.jwt"},
		{"expired token", "Bearer " + expired},
		{"tampered token", "Bearer " + tampered},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.Default()
			router.GET("/probe",
				RequireAuth(mgr, ExtractFromHeader),
				func(c *gin.Context) { c.Status(http.StatusOK) })

			req := httptest.NewRequest(http.MethodGet, "/probe", nil)
			if tt.header != "" {
				req.Header.Set("Authorization", tt.header)
			}
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusUnauthorized, w.Code)
			assert.JSONEq(t, `{"error":"unauthorized"}`, w.Body.String())
		})
	}
}

// middleware lets a valid token through and exposes the claims to the handler
func TestRequireAuthAllowsValidToken(t *testing.T) {
	mgr, key := newTestManager(t)

	tokenString := sign(t, key, token.Claims{
		UserID:   42,
		Username: "matestier",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	})

	router := gin.Default()
	router.GET("/protected",
		RequireAuth(mgr, ExtractFromHeader),
		func(c *gin.Context) {
			claims := c.MustGet("user").(*token.Claims)
			assert.Equal(t, 42, claims.UserID)
			assert.Equal(t, "matestier", claims.Username)
			c.Status(http.StatusOK)
		})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// the WS flavor reads the token from the query param instead of the header
func TestRequireAuthQueryToken(t *testing.T) {
	mgr, key := newTestManager(t)

	tokenString := sign(t, key, token.Claims{
		UserID:   7,
		Username: "alice",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	})

	router := gin.Default()
	router.GET("/protected",
		RequireAuth(mgr, ExtractFromQuery),
		func(c *gin.Context) {
			claims := c.MustGet("user").(*token.Claims)
			assert.Equal(t, 7, claims.UserID)
			c.Status(http.StatusOK)
		})

	// no token in the query -> rejected
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// valid token in the query -> allowed
	req = httptest.NewRequest(http.MethodGet, "/protected?token="+tokenString, nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}
