package ventas_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"tesserak/backend/internal/auth"
	"tesserak/backend/internal/ventas"
	"tesserak/backend/migrations"
)

const testSecret = "ventas-test-secret-0123456789"

func newTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	api := r.Group("/api/v1")
	ventas.NewHandler(ventas.NewService(nil), auth.RequireAuth(testSecret)).RegisterRoutes(api)
	return r
}

func testToken(t *testing.T) string {
	t.Helper()
	now := time.Now().UTC()
	tok, err := auth.SignHS256(auth.Claims{
		Sub:        1,
		EmpresaID:  1,
		SucursalID: 1,
		Email:      "tester@tesserak.local",
		Iat:        now.Unix(),
		Exp:        now.Add(time.Hour).Unix(),
	}, testSecret)
	if err != nil {
		t.Fatalf("SignHS256: %v", err)
	}
	return tok
}

func TestAutenticacionVentas(t *testing.T) {
	r := newTestRouter()
	rutas := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/api/v1/ventas"},
		{http.MethodPost, "/api/v1/ventas"},
		{http.MethodGet, "/api/v1/ventas/1"},
		{http.MethodPatch, "/api/v1/ventas/1"},
		{http.MethodDelete, "/api/v1/ventas/1"},
	}

	for _, tc := range rutas {
		t.Run("sin_token_"+tc.method+"_"+tc.path, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != http.StatusUnauthorized {
				t.Fatalf("esperado 401, obtenido %d (%s)", w.Code, w.Body.String())
			}
		})
	}
}

func TestValidacionesVentas(t *testing.T) {
	r := newTestRouter()
	tok := testToken(t)

	casos := []struct {
		name   string
		method string
		path   string
		body   string
		status int
	}{
		{"crear_sin_campos", http.MethodPost, "/api/v1/ventas", `{}`, http.StatusBadRequest},
		{"crear_sin_detalles", http.MethodPost, "/api/v1/ventas", `{"fecha_venta":"2026-11-26","detalles":[]}`, http.StatusBadRequest},
		{"crear_cantidad_cero", http.MethodPost, "/api/v1/ventas", `{"fecha_venta":"2026-11-26","detalles":[{"articulo_id":1,"cantidad":0,"precio_unitario":100}]}`, http.StatusBadRequest},
		{"crear_fecha_invalida", http.MethodPost, "/api/v1/ventas", `{"fecha_venta":"26/11/2026","detalles":[{"articulo_id":1,"cantidad":1,"precio_unitario":100}]}`, http.StatusBadRequest},
		{"crear_vencimiento_anterior", http.MethodPost, "/api/v1/ventas", `{"fecha_venta":"2026-11-26","fecha_vencimiento":"2026-11-20","detalles":[{"articulo_id":1,"cantidad":1,"precio_unitario":100}]}`, http.StatusBadRequest},
		{"listar_estado_invalido", http.MethodGet, "/api/v1/ventas?estado=OTRO", "", http.StatusBadRequest},
		{"obtener_id_invalido", http.MethodGet, "/api/v1/ventas/abc", "", http.StatusBadRequest},
		{"eliminar_id_negativo", http.MethodDelete, "/api/v1/ventas/-1", "", http.StatusBadRequest},
	}

	for _, tc := range casos {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, bytes.NewBufferString(tc.body))
			req.Header.Set("Authorization", "Bearer "+tok)
			if tc.body != "" {
				req.Header.Set("Content-Type", "application/json")
			}
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != tc.status {
				t.Fatalf("esperado %d, obtenido %d (%s)", tc.status, w.Code, w.Body.String())
			}
		})
	}
}

func TestMigracionVentasEmbebida(t *testing.T) {
	data, err := migrations.FS.ReadFile("000019_init_ventas.up.sql")
	if err != nil {
		t.Fatalf("no se encontró 000019_init_ventas.up.sql: %v", err)
	}
	sql := string(data)
	for _, needle := range []string{"ven_ventas", "ven_ventas_detalle", "Consumidor Final", "222222222222", "CLIENTE"} {
		if !strings.Contains(sql, needle) {
			t.Fatalf("la migración 000019_init_ventas.up.sql no contiene %q", needle)
		}
	}
}
