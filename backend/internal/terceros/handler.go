package terceros

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"tesserak/backend/internal/auth"
	"tesserak/backend/internal/httpx"
)

type Handler struct {
	svc            *Service
	authMiddleware gin.HandlerFunc
}

func NewHandler(svc *Service, authMiddleware gin.HandlerFunc) *Handler {
	return &Handler{svc: svc, authMiddleware: authMiddleware}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	g := rg.Group("/terceros")
	if h.authMiddleware != nil {
		g.Use(h.authMiddleware)
	}

	roles := g.Group("/roles")
	roles.POST("", auth.RequirePermission((*auth.Claims).CanManageTercerosFull), h.crearRol)
	roles.GET("", h.listarRoles)
	roles.GET("/:id", h.obtenerRol)
	roles.PATCH("/:id", auth.RequirePermission((*auth.Claims).CanManageTercerosFull), h.actualizarRol)
	roles.DELETE("/:id", auth.RequirePermission((*auth.Claims).CanManageTercerosFull), h.eliminarRol)

	// Tercero maestro (aislado por empresa_id, con soft delete)
	g.POST("", auth.RequirePermission((*auth.Claims).CanManageTercerosCliente), h.crearTercero)
	g.GET("", h.listarTerceros)
	g.GET("/:id", h.obtenerTercero)
	g.PATCH("/:id", auth.RequirePermission((*auth.Claims).CanManageTercerosCliente), h.actualizarTercero)
	g.DELETE("/:id", auth.RequirePermission((*auth.Claims).CanManageTercerosFull), h.eliminarTercero)
}

func (h *Handler) crearRol(c *gin.Context) {
	if _, ok := requireClaims(c); !ok {
		return
	}
	var req CrearRolTerceroRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondError(c, httpx.BindError(err))
		return
	}
	resp, err := h.svc.CrearRol(c.Request.Context(), req)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *Handler) listarRoles(c *gin.Context) {
	if _, ok := requireClaims(c); !ok {
		return
	}
	var q ListarRolesTerceroQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		httpx.RespondError(c, httpx.BindError(err))
		return
	}
	resp, err := h.svc.ListarRoles(c.Request.Context(), q)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) obtenerRol(c *gin.Context) {
	if _, ok := requireClaims(c); !ok {
		return
	}
	id, err := parseID32(c, "id")
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	resp, err := h.svc.ObtenerRol(c.Request.Context(), id)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) actualizarRol(c *gin.Context) {
	if _, ok := requireClaims(c); !ok {
		return
	}
	id, err := parseID32(c, "id")
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	var req ActualizarRolTerceroRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondError(c, httpx.BindError(err))
		return
	}
	resp, err := h.svc.ActualizarRol(c.Request.Context(), id, req)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) eliminarRol(c *gin.Context) {
	if _, ok := requireClaims(c); !ok {
		return
	}
	id, err := parseID32(c, "id")
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	if err := h.svc.EliminarRol(c.Request.Context(), id); err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func requireClaims(c *gin.Context) (*auth.Claims, bool) {
	claims, ok := auth.ClaimsFromContext(c)
	if !ok {
		httpx.RespondError(c, auth.ErrNoAutenticado)
		return nil, false
	}
	return claims, true
}

func (h *Handler) crearTercero(c *gin.Context) {
	claims, ok := requireClaims(c)
	if !ok {
		return
	}
	var req CrearTerceroRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondError(c, httpx.BindError(err))
		return
	}
	resp, err := h.svc.CrearTercero(c.Request.Context(), claims.EmpresaID, req, claims.IsOnlyTercerosCliente())
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *Handler) listarTerceros(c *gin.Context) {
	claims, ok := requireClaims(c)
	if !ok {
		return
	}
	var q ListarTercerosQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		httpx.RespondError(c, httpx.BindError(err))
		return
	}
	resp, err := h.svc.ListarTerceros(c.Request.Context(), claims.EmpresaID, q)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) obtenerTercero(c *gin.Context) {
	claims, ok := requireClaims(c)
	if !ok {
		return
	}
	id, err := parseID64(c, "id")
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	resp, err := h.svc.ObtenerTercero(c.Request.Context(), claims.EmpresaID, id)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) actualizarTercero(c *gin.Context) {
	claims, ok := requireClaims(c)
	if !ok {
		return
	}
	id, err := parseID64(c, "id")
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	var req ActualizarTerceroRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondError(c, httpx.BindError(err))
		return
	}
	resp, err := h.svc.ActualizarTercero(c.Request.Context(), claims.EmpresaID, id, req, claims.IsOnlyTercerosCliente())
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) eliminarTercero(c *gin.Context) {
	claims, ok := requireClaims(c)
	if !ok {
		return
	}
	id, err := parseID64(c, "id")
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	if err := h.svc.EliminarTercero(c.Request.Context(), claims.EmpresaID, id); err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func parseID32(c *gin.Context, param string) (int32, error) {
	v, err := strconv.ParseInt(c.Param(param), 10, 32)
	if err != nil || v <= 0 {
		return 0, httpx.NewError(http.StatusBadRequest, "id_invalido", "el parámetro "+param+" debe ser un entero positivo")
	}
	return int32(v), nil
}

func parseID64(c *gin.Context, param string) (int64, error) {
	v, err := strconv.ParseInt(c.Param(param), 10, 64)
	if err != nil || v <= 0 {
		return 0, httpx.NewError(http.StatusBadRequest, "id_invalido", "el parámetro "+param+" debe ser un entero positivo")
	}
	return v, nil
}
