package consecutivos

import (
	"io/fs"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"tesserak/backend/internal/auth"
	"tesserak/backend/migrations"
)

const testSecret = "test-secret-consecutivos"

func newTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	NewHandler(NewService(nil), auth.RequireAuth(testSecret)).RegisterRoutes(r.Group("/api/v1"))
	return r
}

func makeToken(t *testing.T, empresaID int32) string {
	t.Helper()
	now := time.Now().UTC()
	tok, err := auth.SignHS256(auth.Claims{
		Sub:        1,
		EmpresaID:  empresaID,
		SucursalID: 1,
		Email:      "test@tesserak.local",
		Iat:        now.Unix(),
		Exp:        now.Add(time.Hour).Unix(),
	}, testSecret)
	if err != nil {
		t.Fatalf("error firmando token: %v", err)
	}
	return tok
}

func TestAutenticacionConsecutivos(t *testing.T) {
	r := newTestRouter()

	for _, tc := range []struct {
		metodo, ruta string
	}{
		{http.MethodGet, "/api/v1/consecutivos/tipos-documento"},
		{http.MethodPost, "/api/v1/consecutivos/tipos-documento"},
		{http.MethodGet, "/api/v1/consecutivos/secuencias"},
		{http.MethodPost, "/api/v1/consecutivos/secuencias"},
		{http.MethodPatch, "/api/v1/consecutivos/secuencias/1"},
	} {
		t.Run("sin token "+tc.metodo+" "+tc.ruta, func(t *testing.T) {
			req := httptest.NewRequest(tc.metodo, tc.ruta, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, esperado 401; body = %s", w.Code, w.Body.String())
			}
		})
	}
}

func TestValidacionesConsecutivos(t *testing.T) {
	r := newTestRouter()
	token := makeToken(t, 1)

	casos := []struct {
		nombre, metodo, ruta, body string
		esperado                   int
		contiene                   string
	}{
		{"crear tipo sin campos", http.MethodPost, "/api/v1/consecutivos/tipos-documento", `{}`, 400, `"validacion"`},
		{"crear secuencia sin campos", http.MethodPost, "/api/v1/consecutivos/secuencias", `{}`, 400, `"validacion"`},
		{"crear secuencia frecuencia inválida", http.MethodPost, "/api/v1/consecutivos/secuencias",
			`{"tipo_documento_id":1,"prefijo":"COM","frecuencia_reinicio":"DIARIA"}`, 400, `"field":"frecuencia_reinicio"`},
		{"crear secuencia relleno > 20", http.MethodPost, "/api/v1/consecutivos/secuencias",
			`{"tipo_documento_id":1,"prefijo":"COM","longitud_relleno":25}`, 400, `"field":"longitud_relleno"`},
		{"actualizar secuencia id inválido", http.MethodPatch, "/api/v1/consecutivos/secuencias/abc", `{"prefijo":"OC"}`, 400, `"id_invalido"`},
		{"actualizar secuencia prefijo vacío", http.MethodPatch, "/api/v1/consecutivos/secuencias/1", `{"prefijo":""}`, 400, `"field":"prefijo"`},
		{"actualizar secuencia siguiente < 1", http.MethodPatch, "/api/v1/consecutivos/secuencias/1", `{"siguiente_numero":0}`, 400, `"field":"siguiente_numero"`},
	}

	for _, tc := range casos {
		t.Run(tc.nombre, func(t *testing.T) {
			req := httptest.NewRequest(tc.metodo, tc.ruta, strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+token)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != tc.esperado {
				t.Fatalf("status = %d, esperado %d; body = %s", w.Code, tc.esperado, w.Body.String())
			}
			if !strings.Contains(w.Body.String(), tc.contiene) {
				t.Fatalf("body %s no contiene %s", w.Body.String(), tc.contiene)
			}
		})
	}
}

func TestFormatearConsecutivo(t *testing.T) {
	ref := time.Date(2026, 11, 26, 12, 0, 0, 0, time.UTC)

	if got := FormatearConsecutivo("COM", calcularClavePeriodo(FrecuenciaNunca, ref), 6, 1); got != "COM-000001" {
		t.Fatalf("NUNCA = %q, esperado COM-000001", got)
	}
	if got := FormatearConsecutivo("VEN", calcularClavePeriodo(FrecuenciaAnual, ref), 6, 42); got != "VEN-2026-000042" {
		t.Fatalf("ANUAL = %q, esperado VEN-2026-000042", got)
	}
	if got := FormatearConsecutivo("COT", calcularClavePeriodo(FrecuenciaMensual, ref), 4, 7); got != "COT-2026-11-0007" {
		t.Fatalf("MENSUAL = %q, esperado COT-2026-11-0007", got)
	}
}

func TestMigracionSecuenciasEmbebida(t *testing.T) {
	upBytes, err := fs.ReadFile(migrations.FS, "000016_init_secuencias_documento.up.sql")
	if err != nil {
		t.Fatalf("no se encontró 000016_init_secuencias_documento.up.sql: %v", err)
	}
	upSQL := string(upBytes)
	for _, needle := range []string{
		"CREATE TABLE tipos_documento",
		"CREATE TABLE secuencias_documento",
		"UNIQUE (empresa_id, tipo_documento_id)",
		"'COMPRA'",
		"'VENTA'",
		"'COM'",
		"'VEN'",
	} {
		if !strings.Contains(upSQL, needle) {
			t.Fatalf("la migración 000016 up no contiene %q", needle)
		}
	}

	downBytes, err := fs.ReadFile(migrations.FS, "000016_init_secuencias_documento.down.sql")
	if err != nil {
		t.Fatalf("no se encontró 000016_init_secuencias_documento.down.sql: %v", err)
	}
	downSQL := string(downBytes)
	for _, tabla := range []string{"secuencias_documento", "tipos_documento"} {
		if !strings.Contains(downSQL, "DROP TABLE IF EXISTS "+tabla) {
			t.Fatalf("la migración 000016 down no elimina %s", tabla)
		}
	}
}
