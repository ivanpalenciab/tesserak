package crm

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func newTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	// Service nil: las pruebas de validación fallan antes de llegar al servicio,
	// excepto AgregarInteres sin referencia que se valida antes de tocar el repo.
	NewHandler(NewService(nil)).RegisterRoutes(r.Group("/api/v1"))
	return r
}

func TestValidacionesCRM(t *testing.T) {
	r := newTestRouter()

	casos := []struct {
		nombre, metodo, ruta, body string
		esperado                   int
		contiene                   string
	}{
		// Resumen
		{"resumen sin empresa_id", http.MethodGet, "/api/v1/crm/resumen", ``, 400, `"field":"empresa_id"`},

		// Catálogos
		{"origen sin campos", http.MethodPost, "/api/v1/crm/origenes", `{}`, 400, `"validacion"`},
		{"origenes sin empresa_id", http.MethodGet, "/api/v1/crm/origenes", ``, 400, `"field":"empresa_id"`},
		{"estado orden negativo", http.MethodPost, "/api/v1/crm/estados", `{"empresa_id":1,"nombre":"Nuevo","orden":-1}`, 400, `"field":"orden"`},
		{"motivo sin descripcion", http.MethodPost, "/api/v1/crm/motivos-descarte", `{"empresa_id":1}`, 400, `"field":"descripcion"`},

		// Leads
		{"lead sin campos", http.MethodPost, "/api/v1/crm/leads", `{}`, 400, `"validacion"`},
		{"lead temperatura inválida", http.MethodPost, "/api/v1/crm/leads",
			`{"empresa_id":1,"origen_id":1,"estado_id":1,"nombre":"Ana","apellido":"Gomez","temperatura":"ARDIENDO","created_by":1}`,
			400, `"field":"temperatura"`},
		{"lead score > 100", http.MethodPost, "/api/v1/crm/leads",
			`{"empresa_id":1,"origen_id":1,"estado_id":1,"nombre":"Ana","apellido":"Gomez","score":150,"created_by":1}`,
			400, `"field":"score"`},
		{"lead valor negativo", http.MethodPost, "/api/v1/crm/leads",
			`{"empresa_id":1,"origen_id":1,"estado_id":1,"nombre":"Ana","apellido":"Gomez","valor_estimado":-10,"created_by":1}`,
			400, `"field":"valor_estimado"`},
		{"listar leads sin empresa_id", http.MethodGet, "/api/v1/crm/leads", ``, 400, `"field":"empresa_id"`},
		{"cambiar estado sin estado_id", http.MethodPatch, "/api/v1/crm/leads/1/estado", `{}`, 400, `"field":"estado_id"`},
		{"lead id inválido", http.MethodGet, "/api/v1/crm/leads/xyz", ``, 400, `"id_invalido"`},

		// Actividades
		{"actividad tipo inválido", http.MethodPost, "/api/v1/crm/leads/1/actividades",
			`{"usuario_id":1,"tipo_actividad":"PALOMA","asunto":"Llamar"}`, 400, `"field":"tipo_actividad"`},
		{"actividad estado inválido", http.MethodPatch, "/api/v1/crm/actividades/1",
			`{"estado":"EN_PROCESO"}`, 400, `"field":"estado"`},
		{"actividades page_size > 100", http.MethodGet, "/api/v1/crm/actividades?page_size=200", ``, 400, `"field":"page_size"`},

		// Intereses
		{"interes sin referencia", http.MethodPost, "/api/v1/crm/leads/1/intereses", `{"cantidad_estimada":2}`, 400, `"interes_sin_referencia"`},
		{"interes cantidad <= 0", http.MethodPost, "/api/v1/crm/leads/1/intereses", `{"notas":"Plan pro","cantidad_estimada":0}`, 400, `"field":"cantidad_estimada"`},
	}

	for _, tc := range casos {
		t.Run(tc.nombre, func(t *testing.T) {
			req := httptest.NewRequest(tc.metodo, tc.ruta, strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
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
