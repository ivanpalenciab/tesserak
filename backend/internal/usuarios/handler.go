package usuarios

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"tesserak/backend/internal/auth"
	"tesserak/backend/internal/httpx"
)

type Handler struct {
	svc       *Service
	jwtSecret string
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// WithAuth habilita la verificación JWT en los endpoints de /usuarios.
func (h *Handler) WithAuth(jwtSecret string) *Handler {
	h.jwtSecret = jwtSecret
	return h
}

// RegisterRoutes registra las rutas del recurso /usuarios.
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	g := rg.Group("/usuarios")

	var optAuth, reqAuth, reqAdminOrBranch gin.HandlerFunc
	if h.jwtSecret != "" {
		optAuth = auth.OptionalAuth(h.jwtSecret)
		reqAuth = auth.RequireAuth(h.jwtSecret)
		reqAdminOrBranch = auth.RequireAdminOrBranchAdmin()
	}

	// Catálogos auxiliares y perfil propio (accesible a cualquier usuario autenticado)
	if reqAuth != nil {
		g.GET("/roles", reqAuth, h.listarRoles)
		g.GET("/sucursales", reqAuth, h.listarSucursales)
		g.GET("/mi-perfil", reqAuth, h.obtenerMiPerfil)
		g.PATCH("/mi-perfil", reqAuth, h.actualizarMiPerfil)
	} else {
		g.GET("/roles", h.listarRoles)
		g.GET("/sucursales", h.listarSucursales)
		g.GET("/mi-perfil", h.obtenerMiPerfil)
		g.PATCH("/mi-perfil", h.actualizarMiPerfil)
	}

	// POST /usuarios usa OptionalAuth para permitir bootstrap del 1er usuario de una empresa;
	// a partir del 2do usuario el servicio exige JWT con rol ADMINISTRADOR o ADMINISTRADOR_SUCURSAL.
	if optAuth != nil {
		g.POST("", optAuth, h.crear)
		g.GET("", optAuth, h.listar)
		g.GET("/:id", optAuth, h.obtener)
		g.POST("/:id/regenerar-key", optAuth, h.regenerarKey)
	} else {
		g.POST("", h.crear)
		g.GET("", h.listar)
		g.GET("/:id", h.obtener)
		g.POST("/:id/regenerar-key", h.regenerarKey)
	}

	if reqAuth != nil && reqAdminOrBranch != nil {
		g.PATCH("/:id", reqAuth, reqAdminOrBranch, h.actualizar)
		g.DELETE("/:id", reqAuth, reqAdminOrBranch, h.eliminar)
	} else {
		g.PATCH("/:id", h.actualizar)
		g.DELETE("/:id", h.eliminar)
	}
}

// GET /usuarios/roles
func (h *Handler) listarRoles(c *gin.Context) {
	claims, _ := auth.ClaimsFromContext(c)
	resp, err := h.svc.ListarRoles(c.Request.Context(), claims)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}

// GET /usuarios/sucursales
func (h *Handler) listarSucursales(c *gin.Context) {
	claims, ok := auth.ClaimsFromContext(c)
	if !ok {
		httpx.RespondError(c, auth.ErrNoAutenticado)
		return
	}
	resp, err := h.svc.ListarSucursalesParaCaller(c.Request.Context(), claims)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}

// GET /usuarios/mi-perfil
func (h *Handler) obtenerMiPerfil(c *gin.Context) {
	claims, ok := auth.ClaimsFromContext(c)
	if !ok {
		httpx.RespondError(c, auth.ErrNoAutenticado)
		return
	}
	resp, err := h.svc.Obtener(c.Request.Context(), claims, claims.Sub)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// PATCH /usuarios/mi-perfil
func (h *Handler) actualizarMiPerfil(c *gin.Context) {
	claims, ok := auth.ClaimsFromContext(c)
	if !ok {
		httpx.RespondError(c, auth.ErrNoAutenticado)
		return
	}
	var req ActualizarPerfilRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondError(c, httpx.BindError(err))
		return
	}
	resp, err := h.svc.ActualizarMiPerfil(c.Request.Context(), claims, req)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// POST /usuarios
func (h *Handler) crear(c *gin.Context) {
	var req CrearUsuarioRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondError(c, httpx.BindError(err))
		return
	}
	claims, _ := auth.ClaimsFromContext(c)
	resp, err := h.svc.Crear(c.Request.Context(), claims, req)
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
	claims, _ := auth.ClaimsFromContext(c)
	resp, err := h.svc.Listar(c.Request.Context(), claims, q)
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
	claims, _ := auth.ClaimsFromContext(c)
	resp, err := h.svc.Obtener(c.Request.Context(), claims, id)
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
	claims, _ := auth.ClaimsFromContext(c)
	resp, err := h.svc.Actualizar(c.Request.Context(), claims, id, req)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// POST /usuarios/:id/regenerar-key
func (h *Handler) regenerarKey(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	claims, _ := auth.ClaimsFromContext(c)
	resp, err := h.svc.RegenerarKey(c.Request.Context(), claims, id)
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
	claims, _ := auth.ClaimsFromContext(c)
	if err := h.svc.Eliminar(c.Request.Context(), claims, id); err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func parseID(c *gin.Context) (int64, error) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		return 0, httpx.NewError(http.StatusBadRequest, "id_invalido", "el id debe ser un entero positivo")
	}
	return id, nil
}

// Asegura que errors siga usado si se requiere en validaciones futuras.
var _ = errors.Is
