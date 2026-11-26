package inventario

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

const testSecret = "test-secret-inventario"

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

func TestAutenticacionInventario(t *testing.T) {
	r := newTestRouter()

	for _, tc := range []struct {
		metodo, ruta string
	}{
		{http.MethodGet, "/api/v1/inventario/tipos-movimiento"},
		{http.MethodGet, "/api/v1/inventario/ubicaciones"},
		{http.MethodPost, "/api/v1/inventario/ubicaciones"},
		{http.MethodGet, "/api/v1/inventario/stock"},
		{http.MethodGet, "/api/v1/inventario/movimientos"},
		{http.MethodPost, "/api/v1/inventario/movimientos"},
		{http.MethodGet, "/api/v1/inventario/movimientos/1"},
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

func TestValidacionesInventario(t *testing.T) {
	r := newTestRouter()
	token := makeToken(t, 1)

	casos := []struct {
		nombre, metodo, ruta, body string
		esperado                   int
		contiene                   string
	}{
		{"crear ubicación sin campos", http.MethodPost, "/api/v1/inventario/ubicaciones", `{}`, 400, `"validacion"`},
		{"actualizar ubicación id inválido", http.MethodPatch, "/api/v1/inventario/ubicaciones/abc", `{"nombre":"Bodega 2"}`, 400, `"id_invalido"`},
		{"eliminar ubicación id negativo", http.MethodDelete, "/api/v1/inventario/ubicaciones/-1", ``, 400, `"id_invalido"`},
		{"listar stock page_size > 100", http.MethodGet, "/api/v1/inventario/stock?page_size=200", ``, 400, `"field":"page_size"`},
		{"crear movimiento sin campos", http.MethodPost, "/api/v1/inventario/movimientos", `{}`, 400, `"validacion"`},
		{"crear movimiento sin detalles", http.MethodPost, "/api/v1/inventario/movimientos",
			`{"tipo_movimiento_id":1,"motivo":"Ajuste físico","detalles":[]}`, 400, `"field":"detalles"`},
		{"crear movimiento cantidad <= 0", http.MethodPost, "/api/v1/inventario/movimientos",
			`{"tipo_movimiento_id":1,"motivo":"Pérdida","detalles":[{"articulo_id":1,"ubicacion_origen_id":1,"cantidad":0}]}`, 400, `"field":"cantidad"`},
		{"crear movimiento fecha inválida", http.MethodPost, "/api/v1/inventario/movimientos",
			`{"tipo_movimiento_id":1,"motivo":"Avería","fecha":"26/11/2026","detalles":[{"articulo_id":1,"ubicacion_origen_id":1,"cantidad":2}]}`, 400, `"fecha_invalida"`},
		{"listar movimientos rango fechas inválido", http.MethodGet, "/api/v1/inventario/movimientos?fecha_desde=2026-11-20&fecha_hasta=2026-11-01", ``, 400, `"fechas_invalidas"`},
		{"obtener movimiento id inválido", http.MethodGet, "/api/v1/inventario/movimientos/xyz", ``, 400, `"id_invalido"`},
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

func TestMigracionInventarioEmbebida(t *testing.T) {
	upBytes, err := fs.ReadFile(migrations.FS, "000017_init_inventario.up.sql")
	if err != nil {
		t.Fatalf("no se encontró 000017_init_inventario.up.sql: %v", err)
	}
	upSQL := string(upBytes)
	for _, frag := range []string{
		"CREATE TABLE inv_ubicaciones",
		"CREATE TABLE inv_tipos_movimiento",
		"CREATE TABLE inv_documentos",
		"CREATE TABLE inv_movimientos_detalle",
		"CREATE TABLE inv_stock",
		"'PERDIDA'",
		"'AVERIA'",
		"'DEVOLUCION_PROVEEDOR'",
		"'DEVOLUCION_CLIENTE'",
		"'AJUSTE_POSITIVO'",
		"'AJUSTE_NEGATIVO'",
		"'TRASLADO'",
	} {
		if !strings.Contains(upSQL, frag) {
			t.Fatalf("la migración 000017 up no contiene %s", frag)
		}
	}
}
