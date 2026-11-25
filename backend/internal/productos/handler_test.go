package productos

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

const testSecret = "test-secret-productos"

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

func TestAutenticacionProductos(t *testing.T) {
	r := newTestRouter()

	for _, tc := range []struct {
		metodo, ruta string
	}{
		{http.MethodGet, "/api/v1/productos/tipos-articulo"},
		{http.MethodGet, "/api/v1/productos/metodos-valoracion"},
		{http.MethodGet, "/api/v1/productos/modalidades-servicio"},
		{http.MethodGet, "/api/v1/productos/grupos-impuesto"},
		{http.MethodGet, "/api/v1/productos/unidades-medida"},
		{http.MethodPost, "/api/v1/productos"},
		{http.MethodGet, "/api/v1/productos"},
		{http.MethodGet, "/api/v1/productos/1"},
		{http.MethodPatch, "/api/v1/productos/1"},
		{http.MethodDelete, "/api/v1/productos/1"},
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

func TestValidacionesProductos(t *testing.T) {
	r := newTestRouter()
	token := makeToken(t, 1)

	casos := []struct {
		nombre, metodo, ruta, body string
		esperado                   int
		contiene                   string
	}{
		{"crear sin campos", http.MethodPost, "/api/v1/productos", `{}`, 400, `"validacion"`},
		{"crear sin codigo", http.MethodPost, "/api/v1/productos",
			`{"nombre":"Laptop","tipo_articulo_id":1,"unidad_medida_id":1,"categoria_id":1,"grupo_impuesto_id":1}`, 400, `"field":"codigo"`},
		{"crear sin tipo_articulo_id", http.MethodPost, "/api/v1/productos",
			`{"codigo":"SKU-1","nombre":"Laptop","unidad_medida_id":1,"categoria_id":1,"grupo_impuesto_id":1}`, 400, `"field":"tipo_articulo_id"`},
		{"crear sin unidad_medida_id", http.MethodPost, "/api/v1/productos",
			`{"codigo":"SKU-1","nombre":"Laptop","tipo_articulo_id":1,"categoria_id":1,"grupo_impuesto_id":1}`, 400, `"field":"unidad_medida_id"`},
		{"crear sin categoria_id", http.MethodPost, "/api/v1/productos",
			`{"codigo":"SKU-1","nombre":"Laptop","tipo_articulo_id":1,"unidad_medida_id":1,"grupo_impuesto_id":1}`, 400, `"field":"categoria_id"`},
		{"crear sin grupo_impuesto_id", http.MethodPost, "/api/v1/productos",
			`{"codigo":"SKU-1","nombre":"Laptop","tipo_articulo_id":1,"unidad_medida_id":1,"categoria_id":1}`, 400, `"field":"grupo_impuesto_id"`},
		{"servicio duracion_estimada <= 0", http.MethodPost, "/api/v1/productos",
			`{"codigo":"SRV-1","nombre":"Consultoría","tipo_articulo_id":2,"unidad_medida_id":1,"categoria_id":1,"grupo_impuesto_id":1,"servicio":{"modalidad_id":1,"duracion_estimada":0}}`,
			400, `"field":"duracion_estimada"`},
		{"crear categoria sin nombre", http.MethodPost, "/api/v1/productos/categorias", `{}`, 400, `"field":"nombre"`},
		{"json malformado", http.MethodPost, "/api/v1/productos", `{mal`, 400, `"solicitud_invalida"`},
		{"obtener id inválido", http.MethodGet, "/api/v1/productos/abc", ``, 400, `"id_invalido"`},
		{"patch nombre vacío", http.MethodPatch, "/api/v1/productos/1", `{"nombre":""}`, 400, `"field":"nombre"`},
		{"listar page_size > 100", http.MethodGet, "/api/v1/productos?page_size=200", ``, 400, `"field":"page_size"`},
		{"delete id negativo", http.MethodDelete, "/api/v1/productos/-3", ``, 400, `"id_invalido"`},
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

func TestMigracionArticulosEmbebida(t *testing.T) {
	upBytes, err := fs.ReadFile(migrations.FS, "000009_refactor_articulos_productos_servicios.up.sql")
	if err != nil {
		t.Fatalf("no se encontró 000009_refactor_articulos_productos_servicios.up.sql: %v", err)
	}
	upSQL := string(upBytes)
	for _, tabla := range []string{"tipos_articulo", "metodos_valoracion", "modalidades_servicio", "articulos", "productos", "servicios"} {
		if !strings.Contains(upSQL, "CREATE TABLE "+tabla) {
			t.Fatalf("la migración 000009 up no crea la tabla %s", tabla)
		}
	}

	downBytes, err := fs.ReadFile(migrations.FS, "000009_refactor_articulos_productos_servicios.down.sql")
	if err != nil {
		t.Fatalf("no se encontró 000009_refactor_articulos_productos_servicios.down.sql: %v", err)
	}
	if !strings.Contains(string(downBytes), "DROP TABLE IF EXISTS articulos") {
		t.Fatal("la migración 000009 down no elimina articulos")
	}
}
