package usuarios

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"tesserak/backend/internal/httpx"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes registra las rutas del recurso /usuarios.
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	g := rg.Group("/usuarios")
	g.POST("", h.crear)
	g.GET("", h.listar)
	g.GET("/:id", h.obtener)
	g.PATCH("/:id", h.actualizar)
	g.DELETE("/:id", h.eliminar)
}

// POST /usuarios
func (h *Handler) crear(c *gin.Context) {
	var req CrearUsuarioRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondError(c, httpx.BindError(err))
		return
	}
	resp, err := h.svc.Crear(c.Request.Context(), req)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

// GET /usuarios?empresa_id=&sucursal_id=&activo=&q=&page=&page_size=
func (h *Handler) listar(c *gin.Context) {
	var q ListarUsuariosQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		httpx.RespondError(c, httpx.BindError(err))
		return
	}
	resp, err := h.svc.Listar(c.Request.Context(), q)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// GET /usuarios/:id
func (h *Handler) obtener(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	resp, err := h.svc.Obtener(c.Request.Context(), id)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// PATCH /usuarios/:id
func (h *Handler) actualizar(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	var req ActualizarUsuarioRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondError(c, httpx.BindError(err))
		return
	}
	resp, err := h.svc.Actualizar(c.Request.Context(), id, req)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// DELETE /usuarios/:id  (soft delete)
func (h *Handler) eliminar(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	if err := h.svc.Eliminar(c.Request.Context(), id); err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

var errIDInvalido = errors.New("id inválido")

func parseID(c *gin.Context) (int64, error) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		return 0, httpx.NewError(http.StatusBadRequest, "id_invalido", errIDInvalido.Error())
	}
	return id, nil
}
