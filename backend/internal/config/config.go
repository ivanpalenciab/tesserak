// Package config carga la configuración de la aplicación desde variables de entorno.
package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port        string
	GinMode     string
	DatabaseURL string
}

func Load() (Config, error) {
	cfg := Config{
		Port:        getEnv("PORT", "8080"),
		GinMode:     getEnv("GIN_MODE", "debug"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
	}
	if cfg.DatabaseURL == "" {
		return cfg, fmt.Errorf("la variable de entorno DATABASE_URL es obligatoria " +
			"(ej: postgres://tesserak:tesserak@localhost:5432/tesserak?sslmode=disable)")
	}
	return cfg, nil
}

func getEnv(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return def
}
