// Package config carga la configuración de la aplicación desde variables de entorno
// y, si existe, desde un archivo .env local.
package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type Config struct {
	Port        string
	GinMode     string
	DatabaseURL string
	JWTSecret   string
	CORSOrigin  string
}

func Load() (Config, error) {
	// Carga .env desde el directorio actual (ej. backend/.env si se corre en backend/)
	// o desde backend/.env si se ejecuta desde la raíz del repositorio.
	_ = loadDotEnv(".env", "backend/.env")

	cfg := Config{
		Port:        getEnv("PORT", "8080"),
		GinMode:     getEnv("GIN_MODE", "debug"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
		JWTSecret:   getEnv("JWT_SECRET", "tesserak-dev-secret-cambiar-en-produccion"),
		CORSOrigin:  getEnv("CORS_ORIGIN", "http://localhost:4200"),
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

// loadDotEnv intenta leer el primer archivo .env encontrado en paths y carga sus
// pares CLAVE=VALOR al entorno sin sobrescribir variables ya definidas.
func loadDotEnv(paths ...string) error {
	for _, path := range paths {
		f, err := os.Open(path)
		if err != nil {
			continue
		}
		defer f.Close()

		scanner := bufio.NewScanner(f)
		firstLine := true
		for scanner.Scan() {
			line := scanner.Text()
			if firstLine {
				// Elimina posible BOM UTF-8 generado por editores en Windows.
				line = strings.TrimPrefix(line, "\uFEFF")
				firstLine = false
			}
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			line = strings.TrimPrefix(line, "export ")

			key, val, ok := strings.Cut(line, "=")
			if !ok {
				continue
			}
			key = strings.TrimSpace(key)
			val = strings.TrimSpace(val)
			if key == "" {
				continue
			}
			if len(val) >= 2 && ((val[0] == '"' && val[len(val)-1] == '"') || (val[0] == '\'' && val[len(val)-1] == '\'')) {
				val = val[1 : len(val)-1]
			}

			// Las variables ya definidas en el entorno tienen prioridad sobre .env.
			if existing, exists := os.LookupEnv(key); !exists || existing == "" {
				_ = os.Setenv(key, val)
			}
		}
		return scanner.Err()
	}
	return nil
}
