package usuarios

import (
	"regexp"
	"testing"
)

func TestNewUUIDv4(t *testing.T) {
	re := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

	u1, err := newUUIDv4()
	if err != nil {
		t.Fatalf("newUUIDv4() error: %v", err)
	}
	u2, err := newUUIDv4()
	if err != nil {
		t.Fatalf("newUUIDv4() error: %v", err)
	}

	if !re.MatchString(u1) {
		t.Fatalf("u1 = %q no cumple el formato UUID v4", u1)
	}
	if !re.MatchString(u2) {
		t.Fatalf("u2 = %q no cumple el formato UUID v4", u2)
	}
	if u1 == u2 {
		t.Fatalf("u1 y u2 no deben ser iguales: %q", u1)
	}
}
