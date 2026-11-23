package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDotEnv(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	content := "\uFEFF# comentario\n" +
		"DATABASE_URL=\"postgres://user:pass@localhost:5432/db?sslmode=disable\"\n" +
		"PORT=9090\n"
	if err := os.WriteFile(envPath, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	t.Setenv("DATABASE_URL", "")
	t.Setenv("PORT", "")

	if err := loadDotEnv(envPath); err != nil {
		t.Fatalf("loadDotEnv falló: %v", err)
	}
	if got := os.Getenv("DATABASE_URL"); got != "postgres://user:pass@localhost:5432/db?sslmode=disable" {
		t.Fatalf("DATABASE_URL = %q", got)
	}
	if got := os.Getenv("PORT"); got != "9090" {
		t.Fatalf("PORT = %q", got)
	}
}
