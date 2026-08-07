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

	dir := t.TempDir()
	privatePath := filepath.Join(dir, "private.pem")
	publicPath := filepath.Join(dir, "public.pem")

	privateDER, err := x509.MarshalPKCS8PrivateKey(privateKey)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(privatePath,
		pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privateDER}), 0600))

	publicDER, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(publicPath,
		pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: publicDER}), 0600))

	mgr, err := New(privatePath, publicPath)
	require.NoError(t, err)

	return mgr, privateKey
}

func TestIssueAndVerify(t *testing.T) {
	mgr, _ := newTestManager(t)

	tokenString, err := mgr.Issue(42, "matestier")
	require.NoError(t, err)

	claims, err := mgr.Verify(tokenString)
	require.NoError(t, err)
	assert.Equal(t, 42, claims.UserID)
	assert.Equal(t, "matestier", claims.Username)
	assert.True(t, claims.ExpiresAt.After(time.Now()))
}

func TestVerifyExpiredToken(t *testing.T) {
	mgr, privateKey := newTestManager(t)

	expired := jwt.NewWithClaims(jwt.SigningMethodRS256, Claims{
		UserID:   1,
		Username: "foo",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
		},
	})
	tokenString, err := expired.SignedString(privateKey)
	require.NoError(t, err)

	_, err = mgr.Verify(tokenString)
	assert.Error(t, err)
}

func TestVerifyTamperedToken(t *testing.T) {
	mgr, _ := newTestManager(t)

	tokenString, err := mgr.Issue(1, "foo")
	require.NoError(t, err)

	parts := strings.Split(tokenString, ".")
	require.Len(t, parts, 3)

	payload := []byte(parts[1])
	payload[0] ^= 0xFF
	parts[1] = string(payload)

	_, err = mgr.Verify(strings.Join(parts, "."))
	assert.Error(t, err)
}

func TestVerifyWrongKey(t *testing.T) {
	mgr, _ := newTestManager(t)
	otherMgr, _ := newTestManager(t)

	tokenString, err := otherMgr.Issue(1, "foo")
	require.NoError(t, err)

	_, err = mgr.Verify(tokenString)
	assert.Error(t, err)
}

func TestVerifyAlgConfusion(t *testing.T) {
	mgr, _ := newTestManager(t)

	algNone := "eyJhbGciOiJub25lIn0." + base64URL("{\"user_id\":1}") + "."
	_, err := mgr.Verify(algNone)
	assert.Error(t, err)
}

func base64URL(s string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(s))
}
