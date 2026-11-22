package usuarios

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
		{"crear sin campos", http.MethodPost, "/api/v1/usuarios", `{}`, 400, `"validacion"`},
		{"crear email inválido", http.MethodPost, "/api/v1/usuarios",
			`{"sucursal_id":1,"nombre_usuario":"jperez","nombres":"J","apellidos":"P","email":"no-es-email","password":"12345678"}`,
			400, `"field":"email"`},
		{"crear password corto", http.MethodPost, "/api/v1/usuarios",
			`{"sucursal_id":1,"nombre_usuario":"jperez","nombres":"J","apellidos":"P","email":"a@b.co","password":"123"}`,
			400, `"field":"password"`},
		{"json malformado", http.MethodPost, "/api/v1/usuarios", `{`, 400, `"solicitud_invalida"`},
		{"id inválido", http.MethodGet, "/api/v1/usuarios/abc", ``, 400, `"id_invalido"`},
		{"patch nombres vacío", http.MethodPatch, "/api/v1/usuarios/1", `{"nombres":""}`, 400, `"field":"nombres"`},
		{"patch email inválido", http.MethodPatch, "/api/v1/usuarios/1", `{"email":"x"}`, 400, `"field":"email"`},
		{"listar page_size > 100", http.MethodGet, "/api/v1/usuarios?page_size=500", ``, 400, `"field":"page_size"`},
		{"delete id negativo", http.MethodDelete, "/api/v1/usuarios/-1", ``, 400, `"id_invalido"`},
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
