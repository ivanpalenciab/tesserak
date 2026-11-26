package terceros

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

const testSecret = "test-secret-terceros"

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

func TestAutenticacionRolesTercero(t *testing.T) {
	r := newTestRouter()

	for _, tc := range []struct {
		metodo, ruta string
	}{
		{http.MethodPost, "/api/v1/terceros/roles"},
		{http.MethodGet, "/api/v1/terceros/roles"},
		{http.MethodGet, "/api/v1/terceros/roles/1"},
		{http.MethodPatch, "/api/v1/terceros/roles/1"},
		{http.MethodDelete, "/api/v1/terceros/roles/1"},
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

func TestValidacionesRolesTercero(t *testing.T) {
	r := newTestRouter()
	token := makeToken(t, 1)

	casos := []struct {
		nombre, metodo, ruta, body string
		esperado                   int
		contiene                   string
	}{
		{"crear sin campos", http.MethodPost, "/api/v1/terceros/roles", `{}`, 400, `"validacion"`},
		{"crear sin codigo", http.MethodPost, "/api/v1/terceros/roles", `{"nombre_rol":"Distribuidor"}`, 400, `"field":"codigo"`},
		{"crear sin nombre_rol", http.MethodPost, "/api/v1/terceros/roles", `{"codigo":"DISTRIBUIDOR"}`, 400, `"field":"nombre_rol"`},
		{"crear codigo muy corto", http.MethodPost, "/api/v1/terceros/roles", `{"codigo":"A","nombre_rol":"Aliado"}`, 400, `"field":"codigo"`},
		{"obtener id inválido", http.MethodGet, "/api/v1/terceros/roles/abc", ``, 400, `"id_invalido"`},
		{"actualizar nombre vacío", http.MethodPatch, "/api/v1/terceros/roles/1", `{"nombre_rol":""}`, 400, `"field":"nombre_rol"`},
		{"eliminar id negativo", http.MethodDelete, "/api/v1/terceros/roles/-5", ``, 400, `"id_invalido"`},
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

func TestMigracionRolesTerceroConSemilla(t *testing.T) {
	upBytes, err := fs.ReadFile(migrations.FS, "000013_init_roles_tercero.up.sql")
	if err != nil {
		t.Fatalf("no se encontró 000013_init_roles_tercero.up.sql: %v", err)
	}
	upSQL := string(upBytes)
	if !strings.Contains(upSQL, "CREATE TABLE roles_tercero") {
		t.Fatal("la migración 000013 up no crea roles_tercero")
	}
	for _, seed := range []string{"'CLIENTE'", "'PROVEEDOR'", "'EMPLEADO'"} {
		if !strings.Contains(upSQL, seed) {
			t.Fatalf("la migración 000013 up no contiene la semilla %s", seed)
		}
	}

	downBytes, err := fs.ReadFile(migrations.FS, "000013_init_roles_tercero.down.sql")
	if err != nil {
		t.Fatalf("no se encontró 000013_init_roles_tercero.down.sql: %v", err)
	}
	if !strings.Contains(string(downBytes), "DROP TABLE IF EXISTS roles_tercero") {
		t.Fatal("la migración 000013 down no elimina roles_tercero")
	}
}

func TestAutenticacionTerceros(t *testing.T) {
	r := newTestRouter()

	for _, tc := range []struct {
		metodo, ruta string
	}{
		{http.MethodPost, "/api/v1/terceros"},
		{http.MethodGet, "/api/v1/terceros"},
		{http.MethodGet, "/api/v1/terceros/1"},
		{http.MethodPatch, "/api/v1/terceros/1"},
		{http.MethodDelete, "/api/v1/terceros/1"},
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

func TestValidacionesTerceros(t *testing.T) {
	r := newTestRouter()
	token := makeToken(t, 1)

	casos := []struct {
		nombre, metodo, ruta, body string
		esperado                   int
		contiene                   string
	}{
		{"crear sin campos", http.MethodPost, "/api/v1/terceros", `{}`, 400, `"validacion"`},
		{"crear sin rol_ids", http.MethodPost, "/api/v1/terceros",
			`{"tipo_identificacion":"NIT","numero_identificacion":"900123","razon_social":"Acme SAS","naturaleza_juridica":"JURIDICA"}`,
			400, `"field":"rol_ids"`},
		{"crear rol_ids vacío", http.MethodPost, "/api/v1/terceros",
			`{"tipo_identificacion":"NIT","numero_identificacion":"900123","razon_social":"Acme SAS","naturaleza_juridica":"JURIDICA","rol_ids":[]}`,
			400, `"field":"rol_ids"`},
		{"crear tipo_identificacion inválido", http.MethodPost, "/api/v1/terceros",
			`{"tipo_identificacion":"XYZ","numero_identificacion":"900123","razon_social":"Acme SAS","naturaleza_juridica":"JURIDICA","rol_ids":[1]}`,
			400, `"field":"tipo_identificacion"`},
		{"crear naturaleza_juridica inválida", http.MethodPost, "/api/v1/terceros",
			`{"tipo_identificacion":"NIT","numero_identificacion":"900123","razon_social":"Acme SAS","naturaleza_juridica":"MIXTA","rol_ids":[1]}`,
			400, `"field":"naturaleza_juridica"`},
		{"crear email inválido", http.MethodPost, "/api/v1/terceros",
			`{"tipo_identificacion":"NIT","numero_identificacion":"900123","razon_social":"Acme SAS","naturaleza_juridica":"JURIDICA","email":"no-es-email","rol_ids":[1]}`,
			400, `"field":"email"`},
		{"listar naturaleza inválida", http.MethodGet, "/api/v1/terceros?naturaleza_juridica=OTRA", ``, 400, `"field":"naturaleza_juridica"`},
		{"listar page_size > 100", http.MethodGet, "/api/v1/terceros?page_size=200", ``, 400, `"field":"page_size"`},
		{"obtener id inválido", http.MethodGet, "/api/v1/terceros/abc", ``, 400, `"id_invalido"`},
		{"actualizar razon_social vacía", http.MethodPatch, "/api/v1/terceros/1", `{"razon_social":""}`, 400, `"field":"razon_social"`},
		{"eliminar id negativo", http.MethodDelete, "/api/v1/terceros/-3", ``, 400, `"id_invalido"`},
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

func TestMigracionTerceroMaestroEmbebida(t *testing.T) {
	upBytes, err := fs.ReadFile(migrations.FS, "000014_init_tercero_maestro.up.sql")
	if err != nil {
		t.Fatalf("no se encontró 000014_init_tercero_maestro.up.sql: %v", err)
	}
	upSQL := string(upBytes)
	for _, fragmento := range []string{"CREATE TABLE tercero_maestro", "REFERENCES empresas (id)", "deleted_at"} {
		if !strings.Contains(upSQL, fragmento) {
			t.Fatalf("la migración 000014 up no contiene %q", fragmento)
		}
	}

	downBytes, err := fs.ReadFile(migrations.FS, "000014_init_tercero_maestro.down.sql")
	if err != nil {
		t.Fatalf("no se encontró 000014_init_tercero_maestro.down.sql: %v", err)
	}
	if !strings.Contains(string(downBytes), "DROP TABLE IF EXISTS tercero_maestro") {
		t.Fatal("la migración 000014 down no elimina tercero_maestro")
	}

	upRoles, err := fs.ReadFile(migrations.FS, "000015_init_tercero_empresa_roles.up.sql")
	if err != nil {
		t.Fatalf("no se encontró 000015_init_tercero_empresa_roles.up.sql: %v", err)
	}
	if !strings.Contains(string(upRoles), "CREATE TABLE tercero_empresa_roles") {
		t.Fatal("la migración 000015 up no crea tercero_empresa_roles")
	}
}
