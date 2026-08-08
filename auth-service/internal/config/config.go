package config

import "os"

type Config struct {
	DBURL             string
	Port              string
	JWTPrivateKeyPath string
	JWTPublicKeyPath  string
}

func Load() Config {
	return Config{
		DBURL:             getEnv("DB_URL", "postgres:///splitnow_auth?host=/var/run/postgresql"),
		Port:              getEnv("PORT", "8081"),
		JWTPrivateKeyPath: getEnv("JWT_PRIVATE_KEY_PATH", "keys/jwt_private.pem"),
		JWTPublicKeyPath:  getEnv("JWT_PUBLIC_KEY_PATH", "keys/jwt_public.pem"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
