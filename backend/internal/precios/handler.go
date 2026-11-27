package precios

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
	g := rg.Group("/listas-precios")
	if h.authMiddleware != nil {
		g.Use(h.authMiddleware)
	}

	// Cabecera: listas_precios
	g.POST("", auth.RequirePermission((*auth.Claims).CanManageOperacion), h.crearLista)
	g.GET("", h.listarListas)
	g.GET("/:id", h.obtenerLista)
	g.PATCH("/:id", auth.RequirePermission((*auth.Claims).CanManageOperacion), h.actualizarLista)
	g.DELETE("/:id", auth.RequirePermission((*auth.Claims).CanManageOperacion), h.eliminarLista)

	// Detalle: lista_precios_detalle
	g.POST("/:id/detalles", auth.RequirePermission((*auth.Claims).CanManageOperacion), h.crearDetalle)
	g.GET("/:id/detalles", h.listarDetalles)
	g.GET("/:id/detalles/:detalleId", h.obtenerDetalle)
	g.PATCH("/:id/detalles/:detalleId", auth.RequirePermission((*auth.Claims).CanManageOperacion), h.actualizarDetalle)
	g.DELETE("/:id/detalles/:detalleId", auth.RequirePermission((*auth.Claims).CanManageOperacion), h.eliminarDetalle)
}

func (h *Handler) crearLista(c *gin.Context) {
	claims, ok := requireClaims(c)
	if !ok {
		return
	}
	var req CrearListaPrecioRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondError(c, httpx.BindError(err))
		return
	}
	resp, err := h.svc.CrearLista(c.Request.Context(), claims.EmpresaID, req)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *Handler) listarListas(c *gin.Context) {
	claims, ok := requireClaims(c)
	if !ok {
		return
	}
	var q ListarListasPreciosQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		httpx.RespondError(c, httpx.BindError(err))
		return
	}
	resp, err := h.svc.ListarListas(c.Request.Context(), claims.EmpresaID, q)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) obtenerLista(c *gin.Context) {
	claims, ok := requireClaims(c)
	if !ok {
		return
	}
	id, err := parseID32(c, "id")
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	resp, err := h.svc.ObtenerLista(c.Request.Context(), claims.EmpresaID, id)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) actualizarLista(c *gin.Context) {
	claims, ok := requireClaims(c)
	if !ok {
		return
	}
	id, err := parseID32(c, "id")
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	var req ActualizarListaPrecioRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondError(c, httpx.BindError(err))
		return
	}
	resp, err := h.svc.ActualizarLista(c.Request.Context(), claims.EmpresaID, id, req)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) eliminarLista(c *gin.Context) {
	claims, ok := requireClaims(c)
	if !ok {
		return
	}
	id, err := parseID32(c, "id")
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	if err := h.svc.EliminarLista(c.Request.Context(), claims.EmpresaID, id); err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) crearDetalle(c *gin.Context) {
	claims, ok := requireClaims(c)
	if !ok {
		return
	}
	listaID, err := parseID32(c, "id")
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	var req CrearDetallePrecioRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondError(c, httpx.BindError(err))
		return
	}
	resp, err := h.svc.CrearDetalle(c.Request.Context(), claims.EmpresaID, listaID, req)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *Handler) listarDetalles(c *gin.Context) {
	claims, ok := requireClaims(c)
	if !ok {
		return
	}
	listaID, err := parseID32(c, "id")
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	var q ListarDetallesPrecioQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		httpx.RespondError(c, httpx.BindError(err))
		return
	}
	resp, err := h.svc.ListarDetalles(c.Request.Context(), claims.EmpresaID, listaID, q)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) obtenerDetalle(c *gin.Context) {
	claims, ok := requireClaims(c)
	if !ok {
		return
	}
	listaID, err := parseID32(c, "id")
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	detalleID, err := parseID64(c, "detalleId")
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	resp, err := h.svc.ObtenerDetalle(c.Request.Context(), claims.EmpresaID, listaID, detalleID)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) actualizarDetalle(c *gin.Context) {
	claims, ok := requireClaims(c)
	if !ok {
		return
	}
	listaID, err := parseID32(c, "id")
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	detalleID, err := parseID64(c, "detalleId")
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	var req ActualizarDetallePrecioRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondError(c, httpx.BindError(err))
		return
	}
	resp, err := h.svc.ActualizarDetalle(c.Request.Context(), claims.EmpresaID, listaID, detalleID, req)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) eliminarDetalle(c *gin.Context) {
	claims, ok := requireClaims(c)
	if !ok {
		return
	}
	listaID, err := parseID32(c, "id")
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	detalleID, err := parseID64(c, "detalleId")
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	if err := h.svc.EliminarDetalle(c.Request.Context(), claims.EmpresaID, listaID, detalleID); err != nil {
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
