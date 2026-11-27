package crm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"tesserak/backend/internal/auth"
	"tesserak/backend/internal/database"
)

// TestIntegracionCampanas ejecuta el flujo completo contra PostgreSQL cuando
// TEST_DATABASE_URL está configurada. Si no hay BD de pruebas disponible, omite la prueba.
func TestIntegracionCampanas(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL no configurada; omitiendo prueba de integración con PostgreSQL")
	}

	db, err := database.Connect(dsn, false)
	if err != nil {
		t.Fatalf("conectando a PostgreSQL de pruebas: %v", err)
	}
	if err := database.Migrate(db); err != nil {
		t.Fatalf("aplicando migraciones: %v", err)
	}

	// Usamos una transacción aislada o limpieza al final para no ensuciar la BD.
	tx := db.Begin()
	if tx.Error != nil {
		t.Fatalf("iniciando tx: %v", tx.Error)
	}
	defer tx.Rollback()

	ctx := setupFixturesIntegracion(t, tx)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	NewHandler(NewService(NewRepository(tx))).
		WithAuthMiddleware(auth.RequireAuth(testJWTSecret)).
		RegisterRoutes(r.Group("/api/v1"))

	tokA := makeTestToken(t, ctx.empresaA, ctx.usuarioA)
	tokB := makeTestToken(t, ctx.empresaB, ctx.usuarioB)

	// 1. Crear tipo de campaña en empresa A
	w := doJSON(r, http.MethodPost, "/api/v1/crm/tipos-campana", tokA, map[string]any{
		"nombre":      "Lanzamiento VIP",
		"descripcion": "Campañas exclusivas",
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("crear tipo = %d; body = %s", w.Code, w.Body.String())
	}
	var tipoA TipoCampanaResponse
	mustDecode(t, w.Body.Bytes(), &tipoA)

	// 2. Crear campaña en empresa A
	w = doJSON(r, http.MethodPost, "/api/v1/crm/campanas", tokA, map[string]any{
		"tipo_campana_id": tipoA.ID,
		"origen_id":       ctx.origenA,
		"moneda_id":       ctx.monedaUSD,
		"nombre":          "Black Friday 2026",
		"descripcion":     "Campaña principal Q4",
		"objetivo":        "Captar 200 leads calificados",
		"fecha_inicio":    "2026-11-01",
		"fecha_fin":       "2026-11-30",
		"presupuesto":     5000.0,
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("crear campaña = %d; body = %s", w.Code, w.Body.String())
	}
	var campA CampanaResponse
	mustDecode(t, w.Body.Bytes(), &campA)

	// 3. Aislamiento multiempresa: empresa B no puede consultar ni actualizar campaña A
	w = doJSON(r, http.MethodGet, fmt.Sprintf("/api/v1/crm/campanas/%d", campA.ID), tokB, nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("aislamiento GET campaña = %d, esperado 404", w.Code)
	}
	w = doJSON(r, http.MethodPatch, fmt.Sprintf("/api/v1/crm/campanas/%d", campA.ID), tokB, map[string]any{
		"nombre": "Hackeada",
	})
	if w.Code != http.StatusNotFound {
		t.Fatalf("aislamiento PATCH campaña = %d, esperado 404", w.Code)
	}

	// 4. Actualizar campaña A y validar fecha_fin < fecha_inicio
	w = doJSON(r, http.MethodPatch, fmt.Sprintf("/api/v1/crm/campanas/%d", campA.ID), tokA, map[string]any{
		"fecha_fin": "2026-10-15",
	})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("fecha_fin inválida = %d, esperado 400", w.Code)
	}
	w = doJSON(r, http.MethodPatch, fmt.Sprintf("/api/v1/crm/campanas/%d", campA.ID), tokA, map[string]any{
		"presupuesto": 6000.0,
	})
	if w.Code != http.StatusOK {
		t.Fatalf("actualizar campaña = %d; body = %s", w.Code, w.Body.String())
	}

	// 5. Activar campaña A antes de asociar leads
	w = doJSON(r, http.MethodPost, fmt.Sprintf("/api/v1/crm/campanas/%d/activar", campA.ID), tokA, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("activar campaña = %d; body = %s", w.Code, w.Body.String())
	}

	// 6. Asociar lead A a campaña A con UTM y first_touch
	w = doJSON(r, http.MethodPost, fmt.Sprintf("/api/v1/crm/campanas/%d/leads", campA.ID), tokA, map[string]any{
		"lead_id":         ctx.leadA,
		"tipo_atribucion": "first_touch",
		"utm_source":      "facebook",
		"utm_medium":      "cpc",
		"utm_campaign":    "bf_2026",
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("asociar lead = %d; body = %s", w.Code, w.Body.String())
	}

	// 7. Prevención de asociación duplicada
	w = doJSON(r, http.MethodPost, fmt.Sprintf("/api/v1/crm/campanas/%d/leads", campA.ID), tokA, map[string]any{
		"lead_id":         ctx.leadA,
		"tipo_atribucion": "influencia",
	})
	if w.Code != http.StatusConflict {
		t.Fatalf("asociación duplicada = %d, esperado 409", w.Code)
	}

	// 8. Prevención de relación entre empresas diferentes (lead de empresa B en campaña de empresa A)
	w = doJSON(r, http.MethodPost, fmt.Sprintf("/api/v1/crm/campanas/%d/leads", campA.ID), tokA, map[string]any{
		"lead_id": ctx.leadB,
	})
	if w.Code != http.StatusNotFound {
		t.Fatalf("lead de otra empresa = %d, esperado 404", w.Code)
	}

	// 9. Consultar leads desde campaña y campañas desde lead
	w = doJSON(r, http.MethodGet, fmt.Sprintf("/api/v1/crm/campanas/%d/leads", campA.ID), tokA, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("listar leads de campaña = %d; body = %s", w.Code, w.Body.String())
	}
	var listaLeads ListaLeadsCampanaResponse
	mustDecode(t, w.Body.Bytes(), &listaLeads)
	if listaLeads.Total != 1 || listaLeads.Data[0].LeadID != ctx.leadA {
		t.Fatalf("esperado 1 lead (%d), obtenido %+v", ctx.leadA, listaLeads)
	}

	w = doJSON(r, http.MethodGet, fmt.Sprintf("/api/v1/crm/leads/%d/campanas", ctx.leadA), tokA, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("listar campañas de lead = %d; body = %s", w.Code, w.Body.String())
	}

	// 10. Registrar gasto en moneda correcta y rechazar moneda diferente
	w = doJSON(r, http.MethodPost, fmt.Sprintf("/api/v1/crm/campanas/%d/gastos", campA.ID), tokA, map[string]any{
		"fecha":     "2026-11-05",
		"concepto":  "Pauta Meta Ads",
		"valor":     1250.50,
		"moneda_id": ctx.monedaCOP,
	})
	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("moneda incompatible = %d, esperado 422", w.Code)
	}

	w = doJSON(r, http.MethodPost, fmt.Sprintf("/api/v1/crm/campanas/%d/gastos", campA.ID), tokA, map[string]any{
		"fecha":           "2026-11-05",
		"concepto":        "Pauta Meta Ads",
		"categoria_gasto": "PAUTA_DIGITAL",
		"valor":           1250.50,
		"moneda_id":       ctx.monedaUSD,
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("crear gasto = %d; body = %s", w.Code, w.Body.String())
	}

	// 11. Consultar indicadores de campaña
	w = doJSON(r, http.MethodGet, fmt.Sprintf("/api/v1/crm/campanas/%d/indicadores", campA.ID), tokA, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("indicadores = %d; body = %s", w.Code, w.Body.String())
	}
	var ind IndicadoresCampanaResponse
	mustDecode(t, w.Body.Bytes(), &ind)
	if ind.TotalLeads != 1 || ind.GastoAcumulado == nil || *ind.GastoAcumulado != 1250.50 {
		t.Fatalf("indicadores inesperados: %+v", ind)
	}
	if ind.PresupuestoDisponible == nil || *ind.PresupuestoDisponible != 4749.50 {
		t.Fatalf("presupuesto disponible inesperado: %+v", ind)
	}
}

type integracionFixtures struct {
	empresaA, empresaB   int32
	usuarioA, usuarioB   int64
	origenA, estadoA     int32
	origenB, estadoB     int32
	leadA, leadB         int64
	monedaUSD, monedaCOP int32
}

func setupFixturesIntegracion(t *testing.T, tx *gorm.DB) integracionFixtures {
	t.Helper()
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	var f integracionFixtures

	mustExecScan(t, tx, `INSERT INTO empresas (nombre_legal, razon_social, identificador_fiscal) VALUES ($1, $1, $2) RETURNING id`,
		&f.empresaA, "Emp A "+suffix, "NIT-A-"+suffix)
	mustExecScan(t, tx, `INSERT INTO empresas (nombre_legal, razon_social, identificador_fiscal) VALUES ($1, $1, $2) RETURNING id`,
		&f.empresaB, "Emp B "+suffix, "NIT-B-"+suffix)

	var sucA, sucB int32
	mustExecScan(t, tx, `INSERT INTO sucursales (empresa_id, nombre) VALUES ($1, 'Principal') RETURNING id`, &sucA, f.empresaA)
	mustExecScan(t, tx, `INSERT INTO sucursales (empresa_id, nombre) VALUES ($1, 'Principal') RETURNING id`, &sucB, f.empresaB)

	mustExecScan(t, tx, `INSERT INTO usuarios (empresa_id, nombre_usuario, nombres, apellidos, email, password_hash)
		VALUES ($1, $2, 'Ana', 'A', $3, 'hash') RETURNING id`, &f.usuarioA, f.empresaA, "usera_"+suffix, "a_"+suffix+"@test.local")
	mustExecScan(t, tx, `INSERT INTO usuarios (empresa_id, nombre_usuario, nombres, apellidos, email, password_hash)
		VALUES ($1, $2, 'Bob', 'B', $3, 'hash') RETURNING id`, &f.usuarioB, f.empresaB, "userb_"+suffix, "b_"+suffix+"@test.local")

	if err := tx.Exec(`INSERT INTO usuario_sucursales (usuario_id, sucursal_id, empresa_id, es_principal) VALUES ($1, $2, $3, TRUE), ($4, $5, $6, TRUE)`,
		f.usuarioA, sucA, f.empresaA, f.usuarioB, sucB, f.empresaB).Error; err != nil {
		t.Fatalf("insert usuario_sucursales: %v", err)
	}

	mustExecScan(t, tx, `INSERT INTO monedas (codigo_iso, nombre_moneda, simbolo) VALUES ('U'||substr($1,1,2), 'Test USD', '$') RETURNING id`, &f.monedaUSD, suffix)
	mustExecScan(t, tx, `INSERT INTO monedas (codigo_iso, nombre_moneda, simbolo) VALUES ('C'||substr($1,1,2), 'Test COP', '$') RETURNING id`, &f.monedaCOP, suffix)

	mustExecScan(t, tx, `INSERT INTO crm_origen_lead (empresa_id, nombre) VALUES ($1, 'Web') RETURNING id`, &f.origenA, f.empresaA)
	mustExecScan(t, tx, `INSERT INTO crm_estado_lead (empresa_id, nombre, orden) VALUES ($1, 'Nuevo', 1) RETURNING id`, &f.estadoA, f.empresaA)
	mustExecScan(t, tx, `INSERT INTO crm_origen_lead (empresa_id, nombre) VALUES ($1, 'Web') RETURNING id`, &f.origenB, f.empresaB)
	mustExecScan(t, tx, `INSERT INTO crm_estado_lead (empresa_id, nombre, orden) VALUES ($1, 'Nuevo', 1) RETURNING id`, &f.estadoB, f.empresaB)

	mustExecScan(t, tx, `INSERT INTO crm_leads (empresa_id, origen_id, estado_id, nombre, apellido, created_by)
		VALUES ($1, $2, $3, 'Lead', 'A', $4) RETURNING id`, &f.leadA, f.empresaA, f.origenA, f.estadoA, f.usuarioA)
	mustExecScan(t, tx, `INSERT INTO crm_leads (empresa_id, origen_id, estado_id, nombre, apellido, created_by)
		VALUES ($1, $2, $3, 'Lead', 'B', $4) RETURNING id`, &f.leadB, f.empresaB, f.origenB, f.estadoB, f.usuarioB)

	return f
}

func mustExecScan(t *testing.T, tx *gorm.DB, query string, dest any, args ...any) {
	t.Helper()
	if err := tx.Raw(query, args...).Scan(dest).Error; err != nil {
		t.Fatalf("query failed (%s): %v", query, err)
	}
}

func doJSON(r *gin.Engine, method, path, token string, body any) *httptest.ResponseRecorder {
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func mustDecode(t *testing.T, data []byte, dest any) {
	t.Helper()
	if err := json.Unmarshal(data, dest); err != nil {
		t.Fatalf("error decodificando JSON (%s): %v", string(data), err)
	}
}
