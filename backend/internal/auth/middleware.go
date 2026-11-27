package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"tesserak/backend/internal/httpx"
)

const contextClaimsKey = "auth_claims"

var (
	ErrNoAutenticado       = httpx.NewError(http.StatusUnauthorized, "no_autenticado", "token de autenticación requerido o inválido")
	ErrSoloAdministrador   = httpx.NewError(http.StatusForbidden, "solo_administrador", "esta acción requiere el rol de sistema ADMINISTRADOR")
	ErrSoloAdminOSucursal   = httpx.NewError(http.StatusForbidden, "permiso_denegado", "esta acción requiere rol ADMINISTRADOR o ADMINISTRADOR_SUCURSAL")
	ErrSucursalNoAutorizada   = httpx.NewError(http.StatusForbidden, "sucursal_no_autorizada", "no tiene permisos para operar sobre una o más de las sucursales indicadas")
	ErrPermisoDenegado        = httpx.NewError(http.StatusForbidden, "permiso_denegado", "no tiene permisos para realizar esta acción con sus roles actuales")
	ErrPermisoGestorVentas    = httpx.NewError(http.StatusForbidden, "permiso_denegado", "el rol Gestor de ventas no tiene permisos para realizar esta acción")
	ErrPermisoAsesorComercial = httpx.NewError(http.StatusForbidden, "permiso_denegado", "el rol Asesor comercial no tiene permisos para realizar esta acción")
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

// OptionalAuth verifica el JWT si el header Authorization viene presente y almacena
// los claims en el contexto sin rechazar peticiones anónimas.
func OptionalAuth(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := strings.TrimSpace(c.GetHeader("Authorization"))
		if strings.HasPrefix(strings.ToLower(header), "bearer ") {
			token := strings.TrimSpace(header[7:])
			if token != "" {
				if claims, err := VerifyHS256(token, jwtSecret); err == nil && claims.EmpresaID > 0 && claims.Sub > 0 {
					c.Set(contextClaimsKey, claims)
				}
			}
		}
		c.Next()
	}
}

// RequireAdmin exige que el usuario autenticado posea el rol de sistema ADMINISTRADOR.
func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, ok := ClaimsFromContext(c)
		if !ok {
			httpx.RespondError(c, ErrNoAutenticado)
			return
		}
		if !claims.IsAdmin() {
			httpx.RespondError(c, ErrSoloAdministrador)
			return
		}
		c.Next()
	}
}

// RequireAdminOrBranchAdmin exige rol ADMINISTRADOR o ADMINISTRADOR_SUCURSAL.
func RequireAdminOrBranchAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, ok := ClaimsFromContext(c)
		if !ok {
			httpx.RespondError(c, ErrNoAutenticado)
			return
		}
		if !claims.CanManageUsers() {
			httpx.RespondError(c, ErrSoloAdminOSucursal)
			return
		}
		c.Next()
	}
}

// RequirePermission verifica una capacidad aditiva sobre los claims del usuario autenticado.
func RequirePermission(check func(*Claims) bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, ok := ClaimsFromContext(c)
		if ok && !check(claims) {
			httpx.RespondError(c, ErrPermisoDenegado)
			return
		}
		c.Next()
	}
}

// RejectGestorVentas impide que un usuario restringido al rol GESTOR_VENTAS
// ejecute la ruta (usado para bloqueos de módulos y eliminaciones no permitidas).
func RejectGestorVentas() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, ok := ClaimsFromContext(c)
		if ok && claims.IsOnlyGestorVentas() {
			httpx.RespondError(c, ErrPermisoGestorVentas)
			return
		}
		c.Next()
	}
}

// RejectAsesorComercial impide que un usuario restringido al rol ASESOR_COMERCIAL
// ejecute la ruta (módulos fuera de CRM/creación de ventas).
func RejectAsesorComercial() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, ok := ClaimsFromContext(c)
		if ok && claims.IsOnlyAsesorComercial() {
			httpx.RespondError(c, ErrPermisoAsesorComercial)
			return
		}
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
