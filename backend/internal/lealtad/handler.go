package lealtad

import (
	"net/http"
	"strconv"
	"strings"

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

// RegisterRoutes registra las rutas de administración de planes de lealtad bajo /lealtad.
// Acceso exclusivo para usuarios con rol de sistema ADMINISTRADOR (ADMINISTRADOR_SUCURSAL
// y cualquier otro rol quedan bloqueados con 403).
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	g := rg.Group("/lealtad")
	if h.authMiddleware != nil {
		g.Use(h.authMiddleware, auth.RequireAdmin())
	} else {
		g.Use(auth.RequireAdmin())
	}

	g.GET("/monedas", h.listarMonedas)

	programas := g.Group("/programas")
	programas.GET("", h.listarProgramas)
	programas.POST("", h.crearPrograma)
	programas.GET("/:id", h.obtenerPrograma)
	programas.PATCH("/:id", h.actualizarPrograma)
	programas.DELETE("/:id", h.eliminarPrograma)
	programas.POST("/:id/adherir", h.adherirEmpresa)
	programas.PATCH("/:id/adhesion", h.actualizarAdhesion)
	programas.DELETE("/:id/adherir", h.desadherirEmpresa)
	programas.GET("/:id/empresas", h.listarEmpresasAdheridas)
	programas.GET("/:id/reglas", h.listarReglas)
	programas.POST("/:id/reglas", h.crearRegla)
	programas.PATCH("/:id/reglas/:reglaId", h.actualizarRegla)
	programas.DELETE("/:id/reglas/:reglaId", h.eliminarRegla)
	programas.GET("/:id/conciliacion", h.obtenerConciliacion)
	programas.POST("/:id/procesar-vencimientos", h.procesarVencimientos)

	// Consulta de planes adheridos y saldo de puntos por cliente (accesible desde Ventas y Operación)
	ven := rg.Group("/lealtad/operacion")
	if h.authMiddleware != nil {
		ven.Use(h.authMiddleware, auth.RequirePermission((*auth.Claims).CanAccessVentas))
	} else {
		ven.Use(auth.RequirePermission((*auth.Claims).CanAccessVentas))
	}
	ven.GET("/programas", h.listarProgramasOperacion)
	ven.GET("/cuenta-cliente/:terceroId", h.consultarCuentaCliente)

	// Rutas operativas de Lealtad (miembros y cuentas): accesibles para usuarios con acceso operativo
	op := rg.Group("/lealtad")
	if h.authMiddleware != nil {
		op.Use(h.authMiddleware, auth.RequirePermission((*auth.Claims).CanManageOperacion))
	} else {
		op.Use(auth.RequirePermission((*auth.Claims).CanManageOperacion))
	}

	op.GET("/operacion/paises", h.listarPaises)
	op.GET("/operacion/clientes", h.listarClientesDisponibles)

	miembros := op.Group("/miembros")
	miembros.GET("/buscar", h.buscarMiembroPorDocumento)
	miembros.POST("", h.registrarMiembro)
	miembros.POST("/desde-tercero", h.inscribirDesdeTercero)
	miembros.PATCH("/:id", h.actualizarMiembro)

	cuentas := op.Group("/cuentas")
	cuentas.GET("", h.listarCuentas)
	cuentas.POST("", h.abrirCuenta)
	cuentas.PATCH("/:id/estado", h.actualizarEstadoCuenta)
	cuentas.GET("/:id/transacciones", h.listarTransaccionesCuenta)
	cuentas.POST("/:id/acumular", h.acumularPuntosCuenta)
	cuentas.POST("/:id/redimir", h.redimirPuntosCuenta)
}

func (h *Handler) listarMonedas(c *gin.Context) {
	resp, err := h.svc.ListarMonedas(c.Request.Context())
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}

func (h *Handler) listarProgramas(c *gin.Context) {
	claims, ok := auth.ClaimsFromContext(c)
	if !ok {
		httpx.RespondError(c, auth.ErrNoAutenticado)
		return
	}
	filter := ProgramaFilter{
		Search:        c.Query("search"),
		Alcance:       c.Query("alcance"),
		Estado:        c.Query("estado"),
		SoloAdheridos: strings.EqualFold(c.Query("solo_adheridos"), "true"),
	}
	resp, err := h.svc.ListarProgramas(c.Request.Context(), claims, filter)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	esAdminGlobal, _ := h.svc.EsAdminGlobal(c.Request.Context(), claims)
	c.JSON(http.StatusOK, gin.H{
		"data":            resp,
		"es_admin_global": esAdminGlobal,
	})
}

func (h *Handler) obtenerPrograma(c *gin.Context) {
	claims, ok := auth.ClaimsFromContext(c)
	if !ok {
		httpx.RespondError(c, auth.ErrNoAutenticado)
		return
	}
	id, err := parseID32(c)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	resp, err := h.svc.ObtenerPrograma(c.Request.Context(), claims, id)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) crearPrograma(c *gin.Context) {
	claims, ok := auth.ClaimsFromContext(c)
	if !ok {
		httpx.RespondError(c, auth.ErrNoAutenticado)
		return
	}
	var req CrearProgramaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondError(c, httpx.BindError(err))
		return
	}
	resp, err := h.svc.CrearPrograma(c.Request.Context(), claims, req)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *Handler) actualizarPrograma(c *gin.Context) {
	claims, ok := auth.ClaimsFromContext(c)
	if !ok {
		httpx.RespondError(c, auth.ErrNoAutenticado)
		return
	}
	id, err := parseID32(c)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	var req ActualizarProgramaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondError(c, httpx.BindError(err))
		return
	}
	resp, err := h.svc.ActualizarPrograma(c.Request.Context(), claims, id, req)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) eliminarPrograma(c *gin.Context) {
	claims, ok := auth.ClaimsFromContext(c)
	if !ok {
		httpx.RespondError(c, auth.ErrNoAutenticado)
		return
	}
	id, err := parseID32(c)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	if err := h.svc.EliminarPrograma(c.Request.Context(), claims, id); err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) adherirEmpresa(c *gin.Context) {
	claims, ok := auth.ClaimsFromContext(c)
	if !ok {
		httpx.RespondError(c, auth.ErrNoAutenticado)
		return
	}
	id, err := parseID32(c)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	var req AdherirEmpresaRequest
	if c.Request.ContentLength > 0 {
		if err := c.ShouldBindJSON(&req); err != nil {
			httpx.RespondError(c, httpx.BindError(err))
			return
		}
	}
	resp, err := h.svc.AdherirEmpresa(c.Request.Context(), claims, id, req)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) actualizarAdhesion(c *gin.Context) {
	claims, ok := auth.ClaimsFromContext(c)
	if !ok {
		httpx.RespondError(c, auth.ErrNoAutenticado)
		return
	}
	id, err := parseID32(c)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	var req ActualizarAdhesionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondError(c, httpx.BindError(err))
		return
	}
	resp, err := h.svc.ActualizarAdhesion(c.Request.Context(), claims, id, req)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) desadherirEmpresa(c *gin.Context) {
	claims, ok := auth.ClaimsFromContext(c)
	if !ok {
		httpx.RespondError(c, auth.ErrNoAutenticado)
		return
	}
	id, err := parseID32(c)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	if err := h.svc.DesadherirEmpresa(c.Request.Context(), claims, id); err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) listarEmpresasAdheridas(c *gin.Context) {
	claims, ok := auth.ClaimsFromContext(c)
	if !ok {
		httpx.RespondError(c, auth.ErrNoAutenticado)
		return
	}
	id, err := parseID32(c)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	resp, err := h.svc.ListarEmpresasAdheridas(c.Request.Context(), claims, id)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}

func parseID32(c *gin.Context) (int32, error) {
	raw := c.Param("id")
	v, err := strconv.ParseInt(raw, 10, 32)
	if err != nil || v <= 0 {
		return 0, httpx.NewError(http.StatusBadRequest, "id_invalido", "el identificador debe ser un entero positivo")
	}
	return int32(v), nil
}
