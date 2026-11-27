package lealtad

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"tesserak/backend/internal/auth"
)

const testJWTSecret = "test-secret-lealtad"

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	api := r.Group("/api/v1")
	NewHandler(NewService(nil), auth.RequireAuth(testJWTSecret)).RegisterRoutes(api)
	return r
}

func makeToken(t *testing.T, roles []string, esAdmin, esAdminSucursal bool) string {
	t.Helper()
	now := time.Now().UTC()
	tok, err := auth.SignHS256(auth.Claims{
		Sub:             10,
		EmpresaID:       1,
		SucursalID:      1,
		Roles:           roles,
		EsAdministrador: esAdmin,
		EsAdminSucursal: esAdminSucursal,
		Email:           "test@tesserak.co",
		Iat:             now.Unix(),
		Exp:             now.Add(time.Hour).Unix(),
	}, testJWTSecret)
	if err != nil {
		t.Fatalf("error firmando token: %v", err)
	}
	return tok
}

func TestLealtadRechazaSinAutenticacion(t *testing.T) {
	r := setupTestRouter()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/lealtad/programas", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("esperaba 401, obtuve %d (%s)", w.Code, w.Body.String())
	}
}

func TestLealtadRechazaAdministradorSucursal(t *testing.T) {
	r := setupTestRouter()
	tok := makeToken(t, []string{auth.RolAdminSucursal}, false, true)

	cases := []struct {
		method string
		path   string
		body   string
	}{
		{http.MethodGet, "/api/v1/lealtad/programas", ""},
		{http.MethodPost, "/api/v1/lealtad/programas", `{"codigo":"TEST","nombre":"Test","moneda_referencia_id":1,"valor_punto_redencion":1}`},
		{http.MethodPost, "/api/v1/lealtad/programas/1/adherir", `{}`},
		{http.MethodDelete, "/api/v1/lealtad/programas/1", ""},
		{http.MethodGet, "/api/v1/lealtad/programas/1/reglas", ""},
		{http.MethodPost, "/api/v1/lealtad/programas/1/reglas", `{"nombre":"Regla","puntos_por_monto":1}`},
	}

	for _, tc := range cases {
		req := httptest.NewRequest(tc.method, tc.path, bytes.NewBufferString(tc.body))
		req.Header.Set("Authorization", "Bearer "+tok)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("%s %s: esperaba 403 para ADMINISTRADOR_SUCURSAL, obtuve %d (%s)", tc.method, tc.path, w.Code, w.Body.String())
		}
		var payload map[string]map[string]any
		_ = json.Unmarshal(w.Body.Bytes(), &payload)
		if payload["error"]["code"] != "solo_administrador" {
			t.Errorf("%s %s: esperaba código solo_administrador, obtuve %v", tc.method, tc.path, payload["error"]["code"])
		}
	}
}

func TestLealtadPermiteAdministradorYValidaPayload(t *testing.T) {
	r := setupTestRouter()
	tok := makeToken(t, []string{auth.RolAdministrador}, true, false)

	// Payload inválido (falta nombre y moneda) debe pasar el middleware RequireAdmin y retornar 400
	req := httptest.NewRequest(http.MethodPost, "/api/v1/lealtad/programas", bytes.NewBufferString(`{"codigo":"A"}`))
	req.Header.Set("Authorization", "Bearer "+tok)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("esperaba 400 de validación para ADMINISTRADOR, obtuve %d (%s)", w.Code, w.Body.String())
	}
}

func TestCanManageProgramaPermisosGlobalVsLocal(t *testing.T) {
	svc := NewService(nil)
	adminUserID := int64(42)
	otherAdminID := int64(99)
	empA := int32(1)
	empB := int32(2)

	planGlobal := Programa{
		ID:                     1,
		Codigo:                 CodigoPlanDefecto,
		Nombre:                 "Lealtad Tesserak",
		Alcance:                AlcanceGlobal,
		UsuarioAdministradorID: &adminUserID,
		EsSistema:              true,
		Estado:                 EstadoActivo,
	}
	planLocalEmpA := Programa{
		ID:                   2,
		Codigo:               "CLUB_A",
		Nombre:               "Club Empresa A",
		Alcance:              AlcanceLocal,
		EmpresaPropietariaID: &empA,
		Estado:               EstadoActivo,
	}

	claimsIvan := &auth.Claims{Sub: adminUserID, EmpresaID: empA, Roles: []string{auth.RolAdministrador}, EsAdministrador: true}
	claimsAdminEmpB := &auth.Claims{Sub: otherAdminID, EmpresaID: empB, Roles: []string{auth.RolAdministrador}, EsAdministrador: true}
	claimsBranchAdmin := &auth.Claims{Sub: 7, EmpresaID: empA, Roles: []string{auth.RolAdminSucursal}, EsAdminSucursal: true}

	if !svc.canManagePrograma(claimsIvan, true, planGlobal) {
		t.Errorf("ipalenciab debe poder administrar el plan global Lealtad Tesserak")
	}
	if svc.canManagePrograma(claimsAdminEmpB, false, planGlobal) {
		t.Errorf("un administrador de otra empresa NO debe poder editar el plan global Lealtad Tesserak")
	}
	if !svc.canAdherir(claimsAdminEmpB, planGlobal) {
		t.Errorf("un administrador de otra empresa SÍ debe poder adherirse al plan global Lealtad Tesserak")
	}
	if svc.canManagePrograma(claimsAdminEmpB, false, planLocalEmpA) {
		t.Errorf("un administrador de empresa B NO debe poder editar el plan local de empresa A")
	}
	if svc.canAdherir(claimsAdminEmpB, planLocalEmpA) {
		t.Errorf("un administrador de empresa B NO debe poder adherirse al plan local de empresa A")
	}
	if svc.canManagePrograma(claimsBranchAdmin, false, planLocalEmpA) || svc.canAdherir(claimsBranchAdmin, planGlobal) {
		t.Errorf("ADMINISTRADOR_SUCURSAL no debe tener permisos de gestión ni adhesión")
	}
}

func TestLealtadOperacionPermiteAdminSucursal(t *testing.T) {
	r := setupTestRouter()
	tok := makeToken(t, []string{auth.RolAdminSucursal}, false, true)

	// En rutas operativas (/lealtad/miembros), ADMINISTRADOR_SUCURSAL pasa el middleware y recibe 400 por payload inválido
	req := httptest.NewRequest(http.MethodPost, "/api/v1/lealtad/miembros", bytes.NewBufferString(`{"nombre_completo":"A"}`))
	req.Header.Set("Authorization", "Bearer "+tok)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("esperaba 400 de validación para ADMINISTRADOR_SUCURSAL en /lealtad/miembros, obtuve %d (%s)", w.Code, w.Body.String())
	}
}

func TestCalcularPuntosPorReglas(t *testing.T) {
	now := time.Now().UTC()
	empID := int32(3)

	reglaGlobal := Regla{
		ID:             1,
		EmpresaID:      nil,
		TipoRegla:      TipoReglaPorMonto,
		PuntosPorMonto: 1,
		MontoBase:      1000,
		CompraMinima:   0,
		Prioridad:      100,
		FechaInicio:    now.Add(-time.Hour),
		Activo:         true,
	}
	reglaEmpresa := Regla{
		ID:             2,
		EmpresaID:      &empID,
		TipoRegla:      TipoReglaPorMonto,
		PuntosPorMonto: 2,
		MontoBase:      1000,
		CompraMinima:   5000,
		Prioridad:      50,
		FechaInicio:    now.Add(-time.Hour),
		Activo:         true,
	}
	reglaMult := Regla{
		ID:             3,
		EmpresaID:      &empID,
		TipoRegla:      TipoReglaMultiplicador,
		PuntosPorMonto: 2,
		MontoBase:      1,
		CompraMinima:   10000,
		Prioridad:      10,
		FechaInicio:    now.Add(-time.Hour),
		Activo:         true,
	}

	// 1. Solo regla global: $15.500 -> 15 pts
	pts, rid := CalcularPuntosPorReglas(15500, []Regla{reglaGlobal}, now)
	if pts != 15 || rid == nil || *rid != 1 {
		t.Fatalf("esperaba 15 pts con regla 1, obtuve pts=%d rid=%v", pts, rid)
	}

	// 2. Override por empresa ($6.000 >= compra_minima 5000): 6 * 2 = 12 pts
	pts, rid = CalcularPuntosPorReglas(6000, []Regla{reglaEmpresa, reglaGlobal}, now)
	if pts != 12 || rid == nil || *rid != 2 {
		t.Fatalf("esperaba 12 pts con regla empresa 2, obtuve pts=%d rid=%v", pts, rid)
	}

	// 3. Con multiplicador x2 ($10.000): (10 * 2) * 2 = 40 pts
	pts, rid = CalcularPuntosPorReglas(10000, []Regla{reglaMult, reglaEmpresa, reglaGlobal}, now)
	if pts != 40 || rid == nil || *rid != 2 {
		t.Fatalf("esperaba 40 pts con multiplicador, obtuve pts=%d rid=%v", pts, rid)
	}

	// 4. Regla creada más tarde el mismo día debe aplicar a ventas con fecha YYYY-MM-DD (00:00:00Z)
	fechaVentaMedianoche := time.Date(2026, 11, 27, 0, 0, 0, 0, time.UTC)
	reglaMismoDia := reglaGlobal
	reglaMismoDia.FechaInicio = time.Date(2026, 11, 27, 20, 47, 54, 0, time.UTC)
	pts, rid = CalcularPuntosPorReglas(21950000, []Regla{reglaMismoDia}, fechaVentaMedianoche)
	if pts != 21950 || rid == nil || *rid != 1 {
		t.Fatalf("esperaba 21950 pts para regla del mismo día, obtuve pts=%d rid=%v", pts, rid)
	}
}
