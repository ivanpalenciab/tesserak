package productos

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
	g := rg.Group("/productos")
	if h.authMiddleware != nil {
		g.Use(h.authMiddleware)
	}
	h.registerCatalogoRoutes(g)
	g.POST("", auth.RequirePermission((*auth.Claims).CanManageOperacion), h.crear)
	g.GET("", h.listar)
	g.GET("/:id", h.obtener)
	g.PATCH("/:id", auth.RequirePermission((*auth.Claims).CanManageOperacion), h.actualizar)
	g.DELETE("/:id", auth.RequirePermission((*auth.Claims).CanManageOperacion), h.eliminar)
}

func (h *Handler) crear(c *gin.Context) {
	claims, ok := requireClaims(c)
	if !ok {
		return
	}
	var req CrearProductoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondError(c, httpx.BindError(err))
		return
	}
	resp, err := h.svc.Crear(c.Request.Context(), claims.EmpresaID, req)
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
	var q ListarProductosQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		httpx.RespondError(c, httpx.BindError(err))
		return
	}
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
	id, err := parseID32(c, "id")
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	resp, err := h.svc.Obtener(c.Request.Context(), claims.EmpresaID, id)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) actualizar(c *gin.Context) {
	claims, ok := requireClaims(c)
	if !ok {
		return
	}
	id, err := parseID32(c, "id")
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	var req ActualizarProductoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondError(c, httpx.BindError(err))
		return
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
	id, err := parseID32(c, "id")
	if err != nil {
		httpx.RespondError(c, err)
		return
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

func parseID32(c *gin.Context, param string) (int32, error) {
	v, err := strconv.ParseInt(c.Param(param), 10, 32)
	if err != nil || v <= 0 {
		return 0, httpx.NewError(http.StatusBadRequest, "id_invalido", "el parámetro "+param+" debe ser un entero positivo")
	}
	return int32(v), nil
}
