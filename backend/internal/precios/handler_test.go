package precios

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

const testSecret = "test-secret-precios"

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

func TestAutenticacionListasPrecios(t *testing.T) {
	r := newTestRouter()

	for _, tc := range []struct {
		metodo, ruta string
	}{
		{http.MethodPost, "/api/v1/listas-precios"},
		{http.MethodGet, "/api/v1/listas-precios"},
		{http.MethodGet, "/api/v1/listas-precios/1"},
		{http.MethodPatch, "/api/v1/listas-precios/1"},
		{http.MethodDelete, "/api/v1/listas-precios/1"},
		{http.MethodPost, "/api/v1/listas-precios/1/detalles"},
		{http.MethodGet, "/api/v1/listas-precios/1/detalles"},
		{http.MethodGet, "/api/v1/listas-precios/1/detalles/1"},
		{http.MethodPatch, "/api/v1/listas-precios/1/detalles/1"},
		{http.MethodDelete, "/api/v1/listas-precios/1/detalles/1"},
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

func TestValidacionesListasPrecios(t *testing.T) {
	r := newTestRouter()
	token := makeToken(t, 1)

	casos := []struct {
		nombre, metodo, ruta, body string
		esperado                   int
		contiene                   string
	}{
		{"crear lista sin campos", http.MethodPost, "/api/v1/listas-precios", `{}`, 400, `"validacion"`},
		{"crear lista sin nombre", http.MethodPost, "/api/v1/listas-precios",
			`{"moneda_id":1,"tipo_lista":"GENERAL","vigencia_desde":"2026-01-01"}`, 400, `"field":"nombre"`},
		{"crear lista tipo_lista inválido", http.MethodPost, "/api/v1/listas-precios",
			`{"nombre":"Lista 2026","moneda_id":1,"tipo_lista":"VIP","vigencia_desde":"2026-01-01"}`, 400, `"field":"tipo_lista"`},
		{"crear lista fecha inválida", http.MethodPost, "/api/v1/listas-precios",
			`{"nombre":"Lista 2026","moneda_id":1,"tipo_lista":"GENERAL","vigencia_desde":"01-01-2026"}`, 400, `"fecha_invalida"`},
		{"crear lista vigencia_hasta < vigencia_desde", http.MethodPost, "/api/v1/listas-precios",
			`{"nombre":"Lista 2026","moneda_id":1,"tipo_lista":"MAYORISTA","vigencia_desde":"2026-06-01","vigencia_hasta":"2026-05-01"}`, 400, `"vigencia_invalida"`},
		{"listar listas tipo_lista inválido", http.MethodGet, "/api/v1/listas-precios?tipo_lista=OTRO", ``, 400, `"field":"tipo_lista"`},
		{"listar listas page_size > 100", http.MethodGet, "/api/v1/listas-precios?page_size=200", ``, 400, `"field":"page_size"`},
		{"obtener lista id inválido", http.MethodGet, "/api/v1/listas-precios/abc", ``, 400, `"id_invalido"`},
		{"actualizar lista nombre vacío", http.MethodPatch, "/api/v1/listas-precios/1", `{"nombre":""}`, 400, `"field":"nombre"`},
		{"eliminar lista id negativo", http.MethodDelete, "/api/v1/listas-precios/-2", ``, 400, `"id_invalido"`},

		// Validaciones de detalle
		{"crear detalle sin campos", http.MethodPost, "/api/v1/listas-precios/1/detalles", `{}`, 400, `"validacion"`},
		{"crear detalle precio negativo", http.MethodPost, "/api/v1/listas-precios/1/detalles",
			`{"articulo_id":1,"precio":-10,"unidad_medida_id":1}`, 400, `"field":"precio"`},
		{"crear detalle cantidad_minima <= 0", http.MethodPost, "/api/v1/listas-precios/1/detalles",
			`{"articulo_id":1,"precio":1500,"unidad_medida_id":1,"cantidad_minima":0}`, 400, `"field":"cantidad_minima"`},
		{"crear detalle vigencia_hasta < vigencia_desde", http.MethodPost, "/api/v1/listas-precios/1/detalles",
			`{"articulo_id":1,"precio":1500,"unidad_medida_id":1,"vigencia_desde":"2026-10-01","vigencia_hasta":"2026-09-01"}`, 400, `"vigencia_invalida"`},
		{"crear detalle fecha inválida", http.MethodPost, "/api/v1/listas-precios/1/detalles",
			`{"articulo_id":1,"precio":1500,"unidad_medida_id":1,"vigencia_desde":"2026/10/01"}`, 400, `"fecha_invalida"`},
		{"obtener detalle id inválido", http.MethodGet, "/api/v1/listas-precios/1/detalles/abc", ``, 400, `"id_invalido"`},
		{"actualizar detalle precio negativo", http.MethodPatch, "/api/v1/listas-precios/1/detalles/1", `{"precio":-5}`, 400, `"field":"precio"`},
		{"eliminar detalle lista id inválido", http.MethodDelete, "/api/v1/listas-precios/0/detalles/1", ``, 400, `"id_invalido"`},
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

func TestMigracionListasPreciosEmbebida(t *testing.T) {
	upBytes, err := fs.ReadFile(migrations.FS, "000011_init_listas_precios.up.sql")
	if err != nil {
		t.Fatalf("no se encontró 000011_init_listas_precios.up.sql: %v", err)
	}
	upSQL := string(upBytes)
	for _, tabla := range []string{"listas_precios", "lista_precios_detalle"} {
		if !strings.Contains(upSQL, "CREATE TABLE "+tabla) {
			t.Fatalf("la migración 000011 up no crea la tabla %s", tabla)
		}
	}

	downBytes, err := fs.ReadFile(migrations.FS, "000011_init_listas_precios.down.sql")
	if err != nil {
		t.Fatalf("no se encontró 000011_init_listas_precios.down.sql: %v", err)
	}
	downSQL := string(downBytes)
	for _, tabla := range []string{"lista_precios_detalle", "listas_precios"} {
		if !strings.Contains(downSQL, "DROP TABLE IF EXISTS "+tabla) {
			t.Fatalf("la migración 000011 down no elimina %s", tabla)
		}
	}
}
