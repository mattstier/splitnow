package config

import "os"

type Config struct {
	DBURL            string
	RedisAddr        string
	Port             string
	JWTPublicKeyPath string
}

func Load() Config {
	return Config{
		DBURL:            getEnv("DB_URL", "postgres:///splitnow?host=/var/run/postgresql"),
		RedisAddr:        getEnv("REDIS_ADDR", "localhost:6379"),
		Port:             getEnv("PORT", "8080"),
		JWTPublicKeyPath: getEnv("JWT_PUBLIC_KEY_PATH", "keys/jwt_public.pem"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
