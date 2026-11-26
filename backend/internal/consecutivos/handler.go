package consecutivos

import (
	"errors"
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

// RegisterRoutes registra las rutas de tipos de documento y secuencias bajo /consecutivos.
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	g := rg.Group("/consecutivos")
	if h.authMiddleware != nil {
		g.Use(h.authMiddleware)
	}

	tipos := g.Group("/tipos-documento")
	tipos.GET("", h.listarTipos)
	tipos.POST("", h.crearTipo)

	secuencias := g.Group("/secuencias")
	secuencias.GET("", h.listarSecuencias)
	secuencias.POST("", h.crearSecuencia)
	secuencias.PATCH("/:id", h.actualizarSecuencia)
}

func (h *Handler) listarTipos(c *gin.Context) {
	resp, err := h.svc.ListarTipos(c.Request.Context())
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}

func (h *Handler) crearTipo(c *gin.Context) {
	var req CrearTipoDocumentoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondError(c, httpx.BindError(err))
		return
	}
	resp, err := h.svc.CrearTipo(c.Request.Context(), req)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *Handler) listarSecuencias(c *gin.Context) {
	claims, ok := auth.ClaimsFromContext(c)
	if !ok {
		httpx.RespondError(c, auth.ErrNoAutenticado)
		return
	}
	resp, err := h.svc.ListarSecuencias(c.Request.Context(), claims.EmpresaID)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}

func (h *Handler) crearSecuencia(c *gin.Context) {
	claims, ok := auth.ClaimsFromContext(c)
	if !ok {
		httpx.RespondError(c, auth.ErrNoAutenticado)
		return
	}
	var req CrearSecuenciaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondError(c, httpx.BindError(err))
		return
	}
	resp, err := h.svc.CrearSecuencia(c.Request.Context(), claims.EmpresaID, req)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *Handler) actualizarSecuencia(c *gin.Context) {
	claims, ok := auth.ClaimsFromContext(c)
	if !ok {
		httpx.RespondError(c, auth.ErrNoAutenticado)
		return
	}
	id, err := parseID64(c)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	var req ActualizarSecuenciaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondError(c, httpx.BindError(err))
		return
	}
	resp, err := h.svc.ActualizarSecuencia(c.Request.Context(), claims.EmpresaID, id, req)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

var errIDInvalido = errors.New("id inválido")

func parseID64(c *gin.Context) (int64, error) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		return 0, httpx.NewError(http.StatusBadRequest, "id_invalido", errIDInvalido.Error())
	}
	return id, nil
}
