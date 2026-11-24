package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"tesserak/backend/internal/httpx"
)

const contextClaimsKey = "auth_claims"

var (
	ErrNoAutenticado = httpx.NewError(http.StatusUnauthorized, "no_autenticado", "token de autenticación requerido o inválido")
)

// RequireAuth valida el header Authorization: Bearer <jwt> usando HS256 y almacena
// los claims verificados en el contexto de Gin.
func RequireAuth(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := strings.TrimSpace(c.GetHeader("Authorization"))
		if !strings.HasPrefix(strings.ToLower(header), "bearer ") {
			httpx.RespondError(c, ErrNoAutenticado)
			return
		}
		token := strings.TrimSpace(header[7:])
		if token == "" {
			httpx.RespondError(c, ErrNoAutenticado)
			return
		}

		claims, err := VerifyHS256(token, jwtSecret)
		if err != nil || claims.EmpresaID <= 0 || claims.Sub <= 0 {
			httpx.RespondError(c, ErrNoAutenticado)
			return
		}

		c.Set(contextClaimsKey, claims)
		c.Next()
	}
}

// ClaimsFromContext obtiene los claims JWT autenticados del contexto de Gin.
func ClaimsFromContext(c *gin.Context) (*Claims, bool) {
	v, ok := c.Get(contextClaimsKey)
	if !ok {
		return nil, false
	}
	claims, ok := v.(*Claims)
	return claims, ok && claims != nil
}
