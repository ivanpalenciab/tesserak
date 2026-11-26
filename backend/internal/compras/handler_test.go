package compras

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

const testSecret = "test-secret-compras"

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

func TestAutenticacionCompras(t *testing.T) {
	r := newTestRouter()

	for _, tc := range []struct {
		metodo, ruta string
	}{
		{http.MethodPost, "/api/v1/compras"},
		{http.MethodGet, "/api/v1/compras"},
		{http.MethodGet, "/api/v1/compras/1"},
		{http.MethodPatch, "/api/v1/compras/1"},
		{http.MethodDelete, "/api/v1/compras/1"},
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

func TestValidacionesCompras(t *testing.T) {
	r := newTestRouter()
	token := makeToken(t, 1)

	casos := []struct {
		nombre, metodo, ruta, body string
		esperado                   int
		contiene                   string
	}{
		{"crear sin campos", http.MethodPost, "/api/v1/compras", `{}`, 400, `"validacion"`},
		{"crear sin detalles", http.MethodPost, "/api/v1/compras",
			`{"consecutivo":"OC-001","fecha_compra":"2026-01-10","detalles":[]}`, 400, `"field":"detalles"`},
		{"crear detalle cantidad <= 0", http.MethodPost, "/api/v1/compras",
			`{"consecutivo":"OC-001","fecha_compra":"2026-01-10","detalles":[{"articulo_id":1,"cantidad":0,"costo_unitario":100}]}`,
			400, `"field":"cantidad"`},
		{"crear detalle costo negativo", http.MethodPost, "/api/v1/compras",
			`{"consecutivo":"OC-001","fecha_compra":"2026-01-10","detalles":[{"articulo_id":1,"cantidad":2,"costo_unitario":-5}]}`,
			400, `"field":"costo_unitario"`},
		{"crear estado inválido", http.MethodPost, "/api/v1/compras",
			`{"consecutivo":"OC-001","fecha_compra":"2026-01-10","estado":"DESCONOCIDO","detalles":[{"articulo_id":1,"cantidad":1,"costo_unitario":100}]}`,
			400, `"field":"estado"`},
		{"crear fecha inválida", http.MethodPost, "/api/v1/compras",
			`{"consecutivo":"OC-001","fecha_compra":"10/01/2026","detalles":[{"articulo_id":1,"cantidad":1,"costo_unitario":100}]}`,
			400, `"fecha_invalida"`},
		{"crear vencimiento anterior a compra", http.MethodPost, "/api/v1/compras",
			`{"consecutivo":"OC-001","fecha_compra":"2026-05-10","fecha_vencimiento":"2026-05-01","detalles":[{"articulo_id":1,"cantidad":1,"costo_unitario":100}]}`,
			400, `"fechas_invalidas"`},
		{"listar estado inválido", http.MethodGet, "/api/v1/compras?estado=INVALIDO", ``, 400, `"field":"estado"`},
		{"listar fecha_hasta < fecha_desde", http.MethodGet, "/api/v1/compras?fecha_desde=2026-05-10&fecha_hasta=2026-05-01", ``, 400, `"fechas_invalidas"`},
		{"listar page_size > 100", http.MethodGet, "/api/v1/compras?page_size=200", ``, 400, `"field":"page_size"`},
		{"obtener id inválido", http.MethodGet, "/api/v1/compras/abc", ``, 400, `"id_invalido"`},
		{"actualizar consecutivo vacío", http.MethodPatch, "/api/v1/compras/1", `{"consecutivo":""}`, 400, `"field":"consecutivo"`},
		{"actualizar detalles vacío", http.MethodPatch, "/api/v1/compras/1", `{"detalles":[]}`, 400, `"field":"detalles"`},
		{"eliminar id negativo", http.MethodDelete, "/api/v1/compras/-1", ``, 400, `"id_invalido"`},
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

func TestBuildDetallesTotales(t *testing.T) {
	c1, c2 := 100.5, 50.0
	imp1 := 19.095
	detalles, subtotal, totalIVA, total := buildDetalles([]CompraDetalleInput{
		{ArticuloID: 1, Cantidad: 2, CostoUnitario: &c1, Impuesto: &imp1},
		{ArticuloID: 2, Cantidad: 3, CostoUnitario: &c2},
	})

	if len(detalles) != 2 {
		t.Fatalf("detalles = %d, esperado 2", len(detalles))
	}
	if subtotal != 351.0 {
		t.Fatalf("subtotal = %v, esperado 351", subtotal)
	}
	if totalIVA != 19.095 {
		t.Fatalf("totalIVA = %v, esperado 19.095", totalIVA)
	}
	if total != 370.095 {
		t.Fatalf("total = %v, esperado 370.095", total)
	}
}

func TestMigracionComprasEmbebida(t *testing.T) {
	upBytes, err := fs.ReadFile(migrations.FS, "000012_init_compras.up.sql")
	if err != nil {
		t.Fatalf("no se encontró 000012_init_compras.up.sql: %v", err)
	}
	upSQL := string(upBytes)
	for _, tabla := range []string{"com_compras", "com_compras_detalle"} {
		if !strings.Contains(upSQL, "CREATE TABLE "+tabla) {
			t.Fatalf("la migración 000012 up no crea la tabla %s", tabla)
		}
	}
	if !strings.Contains(upSQL, "deleted_at") {
		t.Fatal("la migración 000012 up no define deleted_at para soft delete")
	}

	downBytes, err := fs.ReadFile(migrations.FS, "000012_init_compras.down.sql")
	if err != nil {
		t.Fatalf("no se encontró 000012_init_compras.down.sql: %v", err)
	}
	downSQL := string(downBytes)
	for _, tabla := range []string{"com_compras_detalle", "com_compras"} {
		if !strings.Contains(downSQL, "DROP TABLE IF EXISTS "+tabla) {
			t.Fatalf("la migración 000012 down no elimina %s", tabla)
		}
	}
}
