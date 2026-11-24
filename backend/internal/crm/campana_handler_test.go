package crm

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

const testJWTSecret = "test-secret-campanas"

func newCampanaTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	NewHandler(NewService(nil)).
		WithAuthMiddleware(auth.RequireAuth(testJWTSecret)).
		RegisterRoutes(r.Group("/api/v1"))
	return r
}

func makeTestToken(t *testing.T, empresaID int32, userID int64) string {
	t.Helper()
	now := time.Now().UTC()
	tok, err := auth.SignHS256(auth.Claims{
		Sub:        userID,
		EmpresaID:  empresaID,
		SucursalID: 1,
		Email:      "test@tesserak.local",
		Iat:        now.Unix(),
		Exp:        now.Add(time.Hour).Unix(),
	}, testJWTSecret)
	if err != nil {
		t.Fatalf("error firmando token: %v", err)
	}
	return tok
}

func TestAutenticacionRutasCampanas(t *testing.T) {
	r := newCampanaTestRouter()

	rutas := []struct {
		metodo string
		ruta   string
	}{
		{http.MethodGet, "/api/v1/crm/tipos-campana"},
		{http.MethodPost, "/api/v1/crm/tipos-campana"},
		{http.MethodGet, "/api/v1/crm/campanas"},
		{http.MethodPost, "/api/v1/crm/campanas"},
		{http.MethodGet, "/api/v1/crm/campanas/1"},
		{http.MethodPatch, "/api/v1/crm/campanas/1"},
		{http.MethodPatch, "/api/v1/crm/campanas/1/estado"},
		{http.MethodPost, "/api/v1/crm/campanas/1/activar"},
		{http.MethodPost, "/api/v1/crm/campanas/1/pausar"},
		{http.MethodPost, "/api/v1/crm/campanas/1/finalizar"},
		{http.MethodGet, "/api/v1/crm/campanas/1/indicadores"},
		{http.MethodPost, "/api/v1/crm/campanas/1/leads"},
		{http.MethodGet, "/api/v1/crm/campanas/1/leads"},
		{http.MethodDelete, "/api/v1/crm/campanas/1/leads/1"},
		{http.MethodGet, "/api/v1/crm/leads/1/campanas"},
		{http.MethodPost, "/api/v1/crm/campanas/1/gastos"},
		{http.MethodGet, "/api/v1/crm/campanas/1/gastos"},
		{http.MethodGet, "/api/v1/crm/gastos/1"},
		{http.MethodPatch, "/api/v1/crm/gastos/1"},
		{http.MethodDelete, "/api/v1/crm/gastos/1"},
	}

	for _, tc := range rutas {
		t.Run("sin token "+tc.metodo+" "+tc.ruta, func(t *testing.T) {
			req := httptest.NewRequest(tc.metodo, tc.ruta, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, esperado 401; body = %s", w.Code, w.Body.String())
			}
		})
	}

	t.Run("token expirado rechazado", func(t *testing.T) {
		expired, _ := auth.SignHS256(auth.Claims{
			Sub:       1,
			EmpresaID: 1,
			Iat:       time.Now().Add(-2 * time.Hour).Unix(),
			Exp:       time.Now().Add(-1 * time.Hour).Unix(),
		}, testJWTSecret)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/crm/campanas", nil)
		req.Header.Set("Authorization", "Bearer "+expired)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, esperado 401", w.Code)
		}
	})
}

func TestValidacionesCampanas(t *testing.T) {
	r := newCampanaTestRouter()
	token := makeTestToken(t, 1, 10)

	casos := []struct {
		nombre, metodo, ruta, body string
		esperado                   int
		contiene                   string
	}{
		// Tipos de campaña
		{"tipo sin nombre", http.MethodPost, "/api/v1/crm/tipos-campana", `{}`, 400, `"field":"nombre"`},

		// Campañas
		{"campana sin campos", http.MethodPost, "/api/v1/crm/campanas", `{}`, 400, `"validacion"`},
		{"campana fecha_inicio inválida", http.MethodPost, "/api/v1/crm/campanas",
			`{"tipo_campana_id":1,"nombre":"Promo Q4","fecha_inicio":"24-11-2026"}`, 400, `"fecha_invalida"`},
		{"campana fecha_fin anterior a fecha_inicio", http.MethodPost, "/api/v1/crm/campanas",
			`{"tipo_campana_id":1,"nombre":"Promo Q4","fecha_inicio":"2026-11-20","fecha_fin":"2026-11-10"}`, 400, `"rango_fechas_invalido"`},
		{"campana presupuesto sin moneda", http.MethodPost, "/api/v1/crm/campanas",
			`{"tipo_campana_id":1,"nombre":"Promo Q4","fecha_inicio":"2026-11-20","presupuesto":1000}`, 400, `"moneda_requerida"`},
		{"campana presupuesto negativo", http.MethodPost, "/api/v1/crm/campanas",
			`{"tipo_campana_id":1,"nombre":"Promo Q4","fecha_inicio":"2026-11-20","presupuesto":-50,"moneda_id":1}`, 400, `"field":"presupuesto"`},
		{"campana estado inválido", http.MethodPatch, "/api/v1/crm/campanas/1/estado",
			`{"estado":"CANCELADA"}`, 400, `"field":"estado"`},
		{"campana id inválido", http.MethodGet, "/api/v1/crm/campanas/abc", ``, 400, `"id_invalido"`},
		{"listar campanas rango de fechas inválido", http.MethodGet, "/api/v1/crm/campanas?fecha_desde=2026-12-10&fecha_hasta=2026-12-01", ``, 400, `"rango_fechas_invalido"`},

		// Asociación campaña - lead
		{"asociar sin lead_id", http.MethodPost, "/api/v1/crm/campanas/1/leads", `{}`, 400, `"field":"lead_id"`},
		{"asociar tipo_atribucion inválido", http.MethodPost, "/api/v1/crm/campanas/1/leads",
			`{"lead_id":1,"tipo_atribucion":"multi_touch"}`, 400, `"field":"tipo_atribucion"`},

		// Gastos
		{"gasto sin campos", http.MethodPost, "/api/v1/crm/campanas/1/gastos", `{}`, 400, `"validacion"`},
		{"gasto valor negativo", http.MethodPost, "/api/v1/crm/campanas/1/gastos",
			`{"fecha":"2026-11-20","concepto":"Ads","valor":-100,"moneda_id":1}`, 400, `"field":"valor"`},
		{"gasto fecha inválida", http.MethodPost, "/api/v1/crm/campanas/1/gastos",
			`{"fecha":"2026/11/20","concepto":"Ads","valor":100,"moneda_id":1}`, 400, `"fecha_invalida"`},
		{"actualizar gasto valor negativo", http.MethodPatch, "/api/v1/crm/gastos/1",
			`{"valor":-1}`, 400, `"field":"valor"`},
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

func TestTransicionesEstadoCampana(t *testing.T) {
	if !esTransicionEstadoValida(CampanaEstadoBorrador, CampanaEstadoActiva) {
		t.Fatal("BORRADOR -> ACTIVA debe ser válida")
	}
	if !esTransicionEstadoValida(CampanaEstadoActiva, CampanaEstadoPausada) {
		t.Fatal("ACTIVA -> PAUSADA debe ser válida")
	}
	if !esTransicionEstadoValida(CampanaEstadoPausada, CampanaEstadoActiva) {
		t.Fatal("PAUSADA -> ACTIVA debe ser válida")
	}
	if !esTransicionEstadoValida(CampanaEstadoActiva, CampanaEstadoFinalizada) {
		t.Fatal("ACTIVA -> FINALIZADA debe ser válida")
	}
	if esTransicionEstadoValida(CampanaEstadoFinalizada, CampanaEstadoActiva) {
		t.Fatal("FINALIZADA -> ACTIVA NO debe ser válida")
	}
	if esTransicionEstadoValida(CampanaEstadoActiva, CampanaEstadoBorrador) {
		t.Fatal("ACTIVA -> BORRADOR NO debe ser válida")
	}

	if campanaHabilitadaParaLeads(CampanaEstadoBorrador) {
		t.Fatal("BORRADOR no debe habilitar asociación de leads")
	}
	if campanaHabilitadaParaLeads(CampanaEstadoFinalizada) {
		t.Fatal("FINALIZADA no debe habilitar asociación de leads")
	}
	if !campanaHabilitadaParaLeads(CampanaEstadoActiva) {
		t.Fatal("ACTIVA debe habilitar asociación de leads")
	}
}

func TestMigracionCampanasEmbebida(t *testing.T) {
	upBytes, err := fs.ReadFile(migrations.FS, "000005_init_crm_campanas.up.sql")
	if err != nil {
		t.Fatalf("no se encontró 000005_init_crm_campanas.up.sql: %v", err)
	}
	upSQL := string(upBytes)
	for _, tabla := range []string{"crm_tipos_campana", "crm_campanas", "crm_campana_leads", "crm_campana_gastos"} {
		if !strings.Contains(upSQL, "CREATE TABLE "+tabla) {
			t.Fatalf("la migración up no crea la tabla %s", tabla)
		}
	}

	downBytes, err := fs.ReadFile(migrations.FS, "000005_init_crm_campanas.down.sql")
	if err != nil {
		t.Fatalf("no se encontró 000005_init_crm_campanas.down.sql: %v", err)
	}
	downSQL := string(downBytes)
	for _, tabla := range []string{"crm_campana_gastos", "crm_campana_leads", "crm_campanas", "crm_tipos_campana"} {
		if !strings.Contains(downSQL, "DROP TABLE IF EXISTS "+tabla) {
			t.Fatalf("la migración down no elimina la tabla %s", tabla)
		}
	}
}
