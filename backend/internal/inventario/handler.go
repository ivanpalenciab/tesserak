package inventario

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
	g := rg.Group("/inventario")
	if h.authMiddleware != nil {
		g.Use(h.authMiddleware)
	}

	g.GET("/sucursales", h.listarSucursales)
	g.GET("/tipos-movimiento", h.listarTiposMovimiento)

	g.POST("/ubicaciones", h.crearUbicacion)
	g.GET("/ubicaciones", h.listarUbicaciones)
	g.PATCH("/ubicaciones/:id", h.actualizarUbicacion)
	g.DELETE("/ubicaciones/:id", h.eliminarUbicacion)

	g.GET("/stock", h.listarStock)

	g.POST("/movimientos", h.crearMovimiento)
	g.GET("/movimientos", h.listarMovimientos)
	g.GET("/movimientos/:id", h.obtenerMovimiento)
}

func (h *Handler) listarSucursales(c *gin.Context) {
	claims, ok := requireClaims(c)
	if !ok {
		return
	}
	resp, err := h.svc.ListarSucursales(c.Request.Context(), claims.EmpresaID)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}

func (h *Handler) listarTiposMovimiento(c *gin.Context) {
	if _, ok := requireClaims(c); !ok {
		return
	}
	resp, err := h.svc.ListarTiposMovimiento(c.Request.Context())
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}

func (h *Handler) crearUbicacion(c *gin.Context) {
	claims, ok := requireClaims(c)
	if !ok {
		return
	}
	var req CrearUbicacionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondError(c, httpx.BindError(err))
		return
	}
	resp, err := h.svc.CrearUbicacion(c.Request.Context(), claims.EmpresaID, claims.SucursalID, req)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *Handler) listarUbicaciones(c *gin.Context) {
	claims, ok := requireClaims(c)
	if !ok {
		return
	}
	var q ListarUbicacionesQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		httpx.RespondError(c, httpx.BindError(err))
		return
	}
	resp, err := h.svc.ListarUbicaciones(c.Request.Context(), claims.EmpresaID, q)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}

func (h *Handler) actualizarUbicacion(c *gin.Context) {
	claims, ok := requireClaims(c)
	if !ok {
		return
	}
	id, err := parseID32(c, "id")
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	var req ActualizarUbicacionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondError(c, httpx.BindError(err))
		return
	}
	resp, err := h.svc.ActualizarUbicacion(c.Request.Context(), claims.EmpresaID, id, req)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) eliminarUbicacion(c *gin.Context) {
	claims, ok := requireClaims(c)
	if !ok {
		return
	}
	id, err := parseID32(c, "id")
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	if err := h.svc.EliminarUbicacion(c.Request.Context(), claims.EmpresaID, id); err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) listarStock(c *gin.Context) {
	claims, ok := requireClaims(c)
	if !ok {
		return
	}
	var q ListarStockQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		httpx.RespondError(c, httpx.BindError(err))
		return
	}
	resp, err := h.svc.ListarStock(c.Request.Context(), claims.EmpresaID, q)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) crearMovimiento(c *gin.Context) {
	claims, ok := requireClaims(c)
	if !ok {
		return
	}
	var req CrearMovimientoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondError(c, httpx.BindError(err))
		return
	}
	resp, err := h.svc.CrearMovimiento(c.Request.Context(), claims.EmpresaID, claims.SucursalID, claims.Sub, req)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *Handler) listarMovimientos(c *gin.Context) {
	claims, ok := requireClaims(c)
	if !ok {
		return
	}
	var q ListarMovimientosQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		httpx.RespondError(c, httpx.BindError(err))
		return
	}
	resp, err := h.svc.ListarMovimientos(c.Request.Context(), claims.EmpresaID, q)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) obtenerMovimiento(c *gin.Context) {
	claims, ok := requireClaims(c)
	if !ok {
		return
	}
	id, err := parseID64(c, "id")
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	resp, err := h.svc.ObtenerMovimiento(c.Request.Context(), claims.EmpresaID, id)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func requireClaims(c *gin.Context) (*auth.Claims, bool) {
	claims, ok := auth.ClaimsFromContext(c)
	if !ok {
		httpx.RespondError(c, auth.ErrNoAutenticado)
		return nil, false
	}
	return claims, true
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
