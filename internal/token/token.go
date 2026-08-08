package token

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"os"

	"github.com/golang-jwt/jwt/v5"
)

// claims that the auth service vouches for with its signature.
// field names must match the auth service's Claims exactly.
type Claims struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}


// Manager should only verify tokens with a public key, and never hold a private
// signing key, to avoid this service forging new tokens
type Manager struct {
	publicKey *rsa.PublicKey
}

// loads the public key used to verify tokens
func New(publicPath string) (*Manager, error) {
	publicPEM, err := os.ReadFile(publicPath)
	if err != nil {
		return nil, fmt.Errorf("read public key: %w", err)
	}
	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(publicPEM)
	if err != nil {
		return nil, fmt.Errorf("parse public key: %w", err)
	}

	return &Manager{publicKey: publicKey}, nil
}

func (m *Manager) Verify(tokenString string) (*Claims, error) {
	parsed, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		// we check if the signing method is the correct one before verifying,
		// to prevent algorithm confusion attacks
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return m.publicKey, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := parsed.Claims.(*Claims)
	if !ok || !parsed.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}
