package config

import "os"

type Config struct {
	DBURL string
	Port  string
}

func Load() Config {
	return Config{
		DBURL: getEnv("DB_URL", "postgres:///splitnow_auth?host=/var/run/postgresql"),
		Port:  getEnv("PORT", "8081"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
