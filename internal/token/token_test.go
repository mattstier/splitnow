package token

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestManager(t *testing.T) (*Manager, *rsa.PrivateKey) {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	publicDER, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	require.NoError(t, err)

	publicPath := filepath.Join(t.TempDir(), "public.pem")
	require.NoError(t, os.WriteFile(publicPath,
		pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: publicDER}), 0600))

	mgr, err := New(publicPath)
	require.NoError(t, err)

	return mgr, privateKey
}

func sign(t *testing.T, privateKey *rsa.PrivateKey, claims Claims) string {
	t.Helper()
	tokenString, err := jwt.NewWithClaims(jwt.SigningMethodRS256, claims).SignedString(privateKey)
	require.NoError(t, err)
	return tokenString
}

func TestVerifyValidToken(t *testing.T) {
	mgr, key := newTestManager(t)

	tokenString := sign(t, key, Claims{
		UserID:   42,
		Username: "matestier",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	})

	claims, err := mgr.Verify(tokenString)
	require.NoError(t, err)
	assert.Equal(t, 42, claims.UserID)
	assert.Equal(t, "matestier", claims.Username)
	assert.True(t, claims.ExpiresAt.After(time.Now()))
}

func TestVerifyExpiredToken(t *testing.T) {
	mgr, key := newTestManager(t)

	tokenString := sign(t, key, Claims{
		UserID:   1,
		Username: "foo",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
		},
	})

	_, err := mgr.Verify(tokenString)
	assert.Error(t, err)
}

func TestVerifyTamperedToken(t *testing.T) {
	mgr, key := newTestManager(t)

	tokenString := sign(t, key, Claims{UserID: 1, Username: "foo"})

	parts := strings.Split(tokenString, ".")
	require.Len(t, parts, 3)

	payload := []byte(parts[1])
	payload[0] ^= 0xFF
	parts[1] = string(payload)

	_, err := mgr.Verify(strings.Join(parts, "."))
	assert.Error(t, err)
}

func TestVerifyWrongKey(t *testing.T) {
	mgr, _ := newTestManager(t)
	_, otherKey := newTestManager(t)

	tokenString := sign(t, otherKey, Claims{UserID: 1, Username: "foo"})

	_, err := mgr.Verify(tokenString)
	assert.Error(t, err)
}

func TestVerifyAlgConfusion(t *testing.T) {
	mgr, _ := newTestManager(t)

	algNone := "eyJhbGciOiJub25lIn0." + base64URL(`{"user_id":1}`) + "."
	_, err := mgr.Verify(algNone)
	assert.Error(t, err)
}

func base64URL(s string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(s))
}
