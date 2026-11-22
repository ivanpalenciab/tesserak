package empresas

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
	// Service nil: estas pruebas solo cubren validación, que falla antes de llegar al servicio.
	NewHandler(nil).RegisterRoutes(r.Group("/api/v1"))
	return r
}

func TestValidaciones(t *testing.T) {
	r := newTestRouter()

	casos := []struct {
		nombre, metodo, ruta, body string
		esperado                   int
		contiene                   string
	}{
		{"crear sin campos", http.MethodPost, "/api/v1/empresas", `{}`, 400, `"validacion"`},
		{"crear email inválido", http.MethodPost, "/api/v1/empresas",
			`{"nombre_legal":"Tesserak","razon_social":"Tesserak SAS","identificador_fiscal":"900123456-7","email":"no-es-email"}`,
			400, `"field":"email"`},
		{"crear sin identificador fiscal", http.MethodPost, "/api/v1/empresas",
			`{"nombre_legal":"Tesserak","razon_social":"Tesserak SAS"}`,
			400, `"field":"identificador_fiscal"`},
		{"json malformado", http.MethodPost, "/api/v1/empresas", `{`, 400, `"solicitud_invalida"`},
		{"id inválido", http.MethodGet, "/api/v1/empresas/abc", ``, 400, `"id_invalido"`},
		{"patch nombre_legal vacío", http.MethodPatch, "/api/v1/empresas/1", `{"nombre_legal":""}`, 400, `"field":"nombre_legal"`},
		{"patch email inválido", http.MethodPatch, "/api/v1/empresas/1", `{"email":"x"}`, 400, `"field":"email"`},
		{"listar page_size > 100", http.MethodGet, "/api/v1/empresas?page_size=500", ``, 400, `"field":"page_size"`},
		{"delete id negativo", http.MethodDelete, "/api/v1/empresas/-1", ``, 400, `"id_invalido"`},
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
