package crm

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"tesserak/backend/internal/auth"
	"tesserak/backend/internal/httpx"
)

func (h *Handler) registerCampanaRoutes(g *gin.RouterGroup) {
	secured := g.Group("")
	if h.authMiddleware != nil {
		secured.Use(h.authMiddleware)
	}

	// Catálogo de tipos de campaña por empresa
	tipos := secured.Group("/tipos-campana")
	tipos.POST("", h.crearTipoCampana)
	tipos.GET("", h.listarTiposCampana)
	tipos.GET("/:id", h.obtenerTipoCampana)
	tipos.PATCH("/:id", h.actualizarTipoCampana)
	tipos.DELETE("/:id", h.eliminarTipoCampana)

	// Campañas de marketing
	campanas := secured.Group("/campanas")
	campanas.POST("", h.crearCampana)
	campanas.GET("", h.listarCampanas)
	campanas.GET("/:id", h.obtenerCampana)
	campanas.PATCH("/:id", h.actualizarCampana)
	campanas.PATCH("/:id/estado", h.cambiarEstadoCampana)
	campanas.POST("/:id/activar", h.activarCampana)
	campanas.POST("/:id/pausar", h.pausarCampana)
	campanas.POST("/:id/finalizar", h.finalizarCampana)
	campanas.GET("/:id/indicadores", h.indicadoresCampana)

	// Relación campaña - leads
	campanas.POST("/:id/leads", h.asociarLeadACampana)
	campanas.GET("/:id/leads", h.listarLeadsDeCampana)
	campanas.DELETE("/:id/leads/:leadId", h.desasociarLeadDeCampana)

	// Relación inversa: campañas asociadas a un lead
	secured.GET("/leads/:id/campanas", h.listarCampanasDeLead)

	// Gastos de campaña
	campanas.POST("/:id/gastos", h.crearGastoCampana)
	campanas.GET("/:id/gastos", h.listarGastosCampana)

	gastos := secured.Group("/gastos")
	gastos.GET("/:id", h.obtenerGastoCampana)
	gastos.PATCH("/:id", h.actualizarGastoCampana)
	gastos.DELETE("/:id", h.eliminarGastoCampana)
}

func requireClaims(c *gin.Context) (*auth.Claims, bool) {
	claims, ok := auth.ClaimsFromContext(c)
	if !ok {
		httpx.RespondError(c, auth.ErrNoAutenticado)
		return nil, false
	}
	return claims, true
}

// =====================================================================
// Tipos de campaña
// =====================================================================

func (h *Handler) crearTipoCampana(c *gin.Context) {
	claims, ok := requireClaims(c)
	if !ok {
		return
	}
	var req CrearTipoCampanaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondError(c, httpx.BindError(err))
		return
	}
	resp, err := h.svc.CrearTipoCampana(c.Request.Context(), claims.EmpresaID, req)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *Handler) listarTiposCampana(c *gin.Context) {
	claims, ok := requireClaims(c)
	if !ok {
		return
	}
	var q ListarTiposCampanaQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		httpx.RespondError(c, httpx.BindError(err))
		return
	}
	resp, err := h.svc.ListarTiposCampana(c.Request.Context(), claims.EmpresaID, q)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}

func (h *Handler) obtenerTipoCampana(c *gin.Context) {
	claims, ok := requireClaims(c)
	if !ok {
		return
	}
	id, err := parseID32(c, "id")
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	resp, err := h.svc.ObtenerTipoCampana(c.Request.Context(), claims.EmpresaID, id)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) actualizarTipoCampana(c *gin.Context) {
	claims, ok := requireClaims(c)
	if !ok {
		return
	}
	id, err := parseID32(c, "id")
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	var req ActualizarTipoCampanaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondError(c, httpx.BindError(err))
		return
	}
	resp, err := h.svc.ActualizarTipoCampana(c.Request.Context(), claims.EmpresaID, id, req)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) eliminarTipoCampana(c *gin.Context) {
	claims, ok := requireClaims(c)
	if !ok {
		return
	}
	id, err := parseID32(c, "id")
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	if err := h.svc.EliminarTipoCampana(c.Request.Context(), claims.EmpresaID, id); err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// =====================================================================
// Campañas
// =====================================================================

func (h *Handler) crearCampana(c *gin.Context) {
	claims, ok := requireClaims(c)
	if !ok {
		return
	}
	var req CrearCampanaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondError(c, httpx.BindError(err))
		return
	}
	resp, err := h.svc.CrearCampana(c.Request.Context(), claims.EmpresaID, claims.Sub, req)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *Handler) listarCampanas(c *gin.Context) {
	claims, ok := requireClaims(c)
	if !ok {
		return
	}
	var q ListarCampanasQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		httpx.RespondError(c, httpx.BindError(err))
		return
	}
	resp, err := h.svc.ListarCampanas(c.Request.Context(), claims.EmpresaID, q)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) obtenerCampana(c *gin.Context) {
	claims, ok := requireClaims(c)
	if !ok {
		return
	}
	id, err := parseID64(c, "id")
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	resp, err := h.svc.ObtenerCampana(c.Request.Context(), claims.EmpresaID, id)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) actualizarCampana(c *gin.Context) {
	claims, ok := requireClaims(c)
	if !ok {
		return
	}
	id, err := parseID64(c, "id")
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	var req ActualizarCampanaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondError(c, httpx.BindError(err))
		return
	}
	resp, err := h.svc.ActualizarCampana(c.Request.Context(), claims.EmpresaID, id, req)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) cambiarEstadoCampana(c *gin.Context) {
	claims, ok := requireClaims(c)
	if !ok {
		return
	}
	id, err := parseID64(c, "id")
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	var req CambiarEstadoCampanaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondError(c, httpx.BindError(err))
		return
	}
	resp, err := h.svc.CambiarEstadoCampana(c.Request.Context(), claims.EmpresaID, id, req.Estado)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) activarCampana(c *gin.Context) {
	h.cambiarEstadoAccion(c, CampanaEstadoActiva)
}

func (h *Handler) pausarCampana(c *gin.Context) {
	h.cambiarEstadoAccion(c, CampanaEstadoPausada)
}

func (h *Handler) finalizarCampana(c *gin.Context) {
	h.cambiarEstadoAccion(c, CampanaEstadoFinalizada)
}

func (h *Handler) cambiarEstadoAccion(c *gin.Context, estado string) {
	claims, ok := requireClaims(c)
	if !ok {
		return
	}
	id, err := parseID64(c, "id")
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	resp, err := h.svc.CambiarEstadoCampana(c.Request.Context(), claims.EmpresaID, id, estado)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) indicadoresCampana(c *gin.Context) {
	claims, ok := requireClaims(c)
	if !ok {
		return
	}
	id, err := parseID64(c, "id")
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	resp, err := h.svc.ObtenerIndicadoresCampana(c.Request.Context(), claims.EmpresaID, id)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// =====================================================================
// Asociación Campaña - Lead
// =====================================================================

func (h *Handler) asociarLeadACampana(c *gin.Context) {
	claims, ok := requireClaims(c)
	if !ok {
		return
	}
	campanaID, err := parseID64(c, "id")
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	var req AsociarLeadCampanaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondError(c, httpx.BindError(err))
		return
	}
	resp, err := h.svc.AsociarLeadACampana(c.Request.Context(), claims.EmpresaID, campanaID, req)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *Handler) listarLeadsDeCampana(c *gin.Context) {
	claims, ok := requireClaims(c)
	if !ok {
		return
	}
	campanaID, err := parseID64(c, "id")
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	var q ListarLeadsCampanaQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		httpx.RespondError(c, httpx.BindError(err))
		return
	}
	resp, err := h.svc.ListarLeadsDeCampana(c.Request.Context(), claims.EmpresaID, campanaID, q)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) listarCampanasDeLead(c *gin.Context) {
	claims, ok := requireClaims(c)
	if !ok {
		return
	}
	leadID, err := parseID64(c, "id")
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	resp, err := h.svc.ListarCampanasDeLead(c.Request.Context(), claims.EmpresaID, leadID)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}

func (h *Handler) desasociarLeadDeCampana(c *gin.Context) {
	claims, ok := requireClaims(c)
	if !ok {
		return
	}
	campanaID, err := parseID64(c, "id")
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	leadID, err := parseID64(c, "leadId")
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	if err := h.svc.DesasociarLeadDeCampana(c.Request.Context(), claims.EmpresaID, campanaID, leadID); err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// =====================================================================
// Gastos de campaña
// =====================================================================

func (h *Handler) crearGastoCampana(c *gin.Context) {
	claims, ok := requireClaims(c)
	if !ok {
		return
	}
	campanaID, err := parseID64(c, "id")
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	var req CrearGastoCampanaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondError(c, httpx.BindError(err))
		return
	}
	resp, err := h.svc.CrearGastoCampana(c.Request.Context(), claims.EmpresaID, claims.Sub, campanaID, req)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *Handler) listarGastosCampana(c *gin.Context) {
	claims, ok := requireClaims(c)
	if !ok {
		return
	}
	campanaID, err := parseID64(c, "id")
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	var q ListarGastosCampanaQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		httpx.RespondError(c, httpx.BindError(err))
		return
	}
	resp, err := h.svc.ListarGastosCampana(c.Request.Context(), claims.EmpresaID, campanaID, q)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) obtenerGastoCampana(c *gin.Context) {
	claims, ok := requireClaims(c)
	if !ok {
		return
	}
	id, err := parseID64(c, "id")
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	resp, err := h.svc.ObtenerGastoCampana(c.Request.Context(), claims.EmpresaID, id)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) actualizarGastoCampana(c *gin.Context) {
	claims, ok := requireClaims(c)
	if !ok {
		return
	}
	id, err := parseID64(c, "id")
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	var req ActualizarGastoCampanaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondError(c, httpx.BindError(err))
		return
	}
	resp, err := h.svc.ActualizarGastoCampana(c.Request.Context(), claims.EmpresaID, id, req)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) eliminarGastoCampana(c *gin.Context) {
	claims, ok := requireClaims(c)
	if !ok {
		return
	}
	id, err := parseID64(c, "id")
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	if err := h.svc.EliminarGastoCampana(c.Request.Context(), claims.EmpresaID, id); err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
