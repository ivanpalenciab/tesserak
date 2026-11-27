package compras

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
	g := rg.Group("/compras")
	if h.authMiddleware != nil {
		g.Use(h.authMiddleware)
	}
	g.POST("", auth.RequirePermission((*auth.Claims).CanAccessCompras), h.crear)
	g.GET("", auth.RequirePermission((*auth.Claims).CanAccessCompras), h.listar)
	g.GET("/:id", auth.RequirePermission((*auth.Claims).CanAccessCompras), h.obtener)
	g.PATCH("/:id", auth.RequirePermission((*auth.Claims).CanAccessCompras), h.actualizar)
	g.DELETE("/:id", auth.RequirePermission((*auth.Claims).CanDeleteCompras), h.eliminar)
}

func (h *Handler) crear(c *gin.Context) {
	claims, ok := requireClaims(c)
	if !ok {
		return
	}
	var req CrearCompraRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondError(c, httpx.BindError(err))
		return
	}
	targetSucursal := claims.SucursalID
	if req.SucursalID != nil {
		targetSucursal = *req.SucursalID
	}
	if !claims.HasSucursal(targetSucursal) {
		httpx.RespondError(c, auth.ErrSucursalNoAutorizada)
		return
	}
	resp, err := h.svc.Crear(c.Request.Context(), claims.EmpresaID, claims.SucursalID, claims.Sub, req)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *Handler) listar(c *gin.Context) {
	claims, ok := requireClaims(c)
	if !ok {
		return
	}
	var q ListarComprasQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		httpx.RespondError(c, httpx.BindError(err))
		return
	}
	q.SucursalIDsScope = claims.AllowedSucursalIDs()
	resp, err := h.svc.Listar(c.Request.Context(), claims.EmpresaID, q)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) obtener(c *gin.Context) {
	claims, ok := requireClaims(c)
	if !ok {
		return
	}
	id, err := parseID64(c, "id")
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	resp, err := h.svc.Obtener(c.Request.Context(), claims.EmpresaID, id)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	if !claims.HasSucursal(resp.SucursalID) {
		httpx.RespondError(c, auth.ErrSucursalNoAutorizada)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) actualizar(c *gin.Context) {
	claims, ok := requireClaims(c)
	if !ok {
		return
	}
	id, err := parseID64(c, "id")
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	var req ActualizarCompraRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondError(c, httpx.BindError(err))
		return
	}
	if !claims.IsAdmin() {
		if req.SucursalID != nil && !claims.HasSucursal(*req.SucursalID) {
			httpx.RespondError(c, auth.ErrSucursalNoAutorizada)
			return
		}
		actual, err := h.svc.Obtener(c.Request.Context(), claims.EmpresaID, id)
		if err != nil {
			httpx.RespondError(c, err)
			return
		}
		if !claims.HasSucursal(actual.SucursalID) {
			httpx.RespondError(c, auth.ErrSucursalNoAutorizada)
			return
		}
	}
	resp, err := h.svc.Actualizar(c.Request.Context(), claims.EmpresaID, id, req)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) eliminar(c *gin.Context) {
	claims, ok := requireClaims(c)
	if !ok {
		return
	}
	id, err := parseID64(c, "id")
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	if !claims.IsAdmin() {
		actual, err := h.svc.Obtener(c.Request.Context(), claims.EmpresaID, id)
		if err != nil {
			httpx.RespondError(c, err)
			return
		}
		if !claims.HasSucursal(actual.SucursalID) {
			httpx.RespondError(c, auth.ErrSucursalNoAutorizada)
			return
		}
	}
	if err := h.svc.Eliminar(c.Request.Context(), claims.EmpresaID, id); err != nil {
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

func parseID64(c *gin.Context, param string) (int64, error) {
	v, err := strconv.ParseInt(c.Param(param), 10, 64)
	if err != nil || v <= 0 {
		return 0, httpx.NewError(http.StatusBadRequest, "id_invalido", "el parámetro "+param+" debe ser un entero positivo")
	}
	return v, nil
}
