package crm

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"tesserak/backend/internal/httpx"
)

type Handler struct {
	svc            *Service
	authMiddleware gin.HandlerFunc
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) WithAuthMiddleware(mw gin.HandlerFunc) *Handler {
	h.authMiddleware = mw
	return h
}

// RegisterRoutes registra las rutas del módulo CRM bajo /crm.
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	g := rg.Group("/crm")

	// Control / Dashboard
	g.GET("/resumen", h.resumen)

	// Catálogos por empresa
	origenes := g.Group("/origenes")
	origenes.POST("", h.crearOrigen)
	origenes.GET("", h.listarOrigenes)
	origenes.GET("/:id", h.obtenerOrigen)
	origenes.PATCH("/:id", h.actualizarOrigen)
	origenes.DELETE("/:id", h.eliminarOrigen)

	estados := g.Group("/estados")
	estados.POST("", h.crearEstado)
	estados.GET("", h.listarEstados)
	estados.GET("/:id", h.obtenerEstado)
	estados.PATCH("/:id", h.actualizarEstado)
	estados.DELETE("/:id", h.eliminarEstado)

	motivos := g.Group("/motivos-descarte")
	motivos.POST("", h.crearMotivo)
	motivos.GET("", h.listarMotivos)
	motivos.GET("/:id", h.obtenerMotivo)
	motivos.PATCH("/:id", h.actualizarMotivo)
	motivos.DELETE("/:id", h.eliminarMotivo)

	h.registerEtapaOportunidadRoutes(g)

	// Leads y acciones del pipeline
	leads := g.Group("/leads")
	leads.POST("", h.crearLead)
	leads.GET("", h.listarLeads)
	leads.GET("/:id", h.obtenerLead)
	leads.PATCH("/:id", h.actualizarLead)
	leads.PATCH("/:id/estado", h.cambiarEstadoLead)
	leads.POST("/:id/convertir", h.convertirLead)
	leads.DELETE("/:id", h.eliminarLead)

	// Actividades del lead
	leads.POST("/:id/actividades", h.crearActividad)
	leads.GET("/:id/actividades", h.listarActividadesLead)

	// Intereses en productos del lead
	leads.POST("/:id/intereses", h.agregarInteres)
	leads.GET("/:id/intereses", h.listarIntereses)
	leads.DELETE("/:id/intereses/:interesId", h.eliminarInteres)

	// Agenda global de actividades
	actividades := g.Group("/actividades")
	actividades.GET("", h.listarActividadesGlobal)
	actividades.PATCH("/:id", h.actualizarActividad)
	actividades.DELETE("/:id", h.eliminarActividad)

	// Campañas de marketing, atribución de leads y gastos (autenticados por JWT)
	h.registerCampanaRoutes(g)
}

// =====================================================================
// Resumen
// =====================================================================

func (h *Handler) resumen(c *gin.Context) {
	var q ResumenCRMQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		httpx.RespondError(c, httpx.BindError(err))
		return
	}
	resp, err := h.svc.Resumen(c.Request.Context(), q.EmpresaID)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// =====================================================================
// Orígenes
// =====================================================================

func (h *Handler) crearOrigen(c *gin.Context) {
	var req CrearOrigenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondError(c, httpx.BindError(err))
		return
	}
	resp, err := h.svc.CrearOrigen(c.Request.Context(), req)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *Handler) listarOrigenes(c *gin.Context) {
	var q ListarCatalogoQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		httpx.RespondError(c, httpx.BindError(err))
		return
	}
	resp, err := h.svc.ListarOrigenes(c.Request.Context(), q)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}

func (h *Handler) obtenerOrigen(c *gin.Context) {
	id, err := parseID32(c, "id")
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	resp, err := h.svc.ObtenerOrigen(c.Request.Context(), id)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) actualizarOrigen(c *gin.Context) {
	id, err := parseID32(c, "id")
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	var req ActualizarOrigenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondError(c, httpx.BindError(err))
		return
	}
	resp, err := h.svc.ActualizarOrigen(c.Request.Context(), id, req)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) eliminarOrigen(c *gin.Context) {
	id, err := parseID32(c, "id")
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	if err := h.svc.EliminarOrigen(c.Request.Context(), id); err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// =====================================================================
// Estados
// =====================================================================

func (h *Handler) crearEstado(c *gin.Context) {
	var req CrearEstadoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondError(c, httpx.BindError(err))
		return
	}
	resp, err := h.svc.CrearEstado(c.Request.Context(), req)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *Handler) listarEstados(c *gin.Context) {
	var q ListarCatalogoQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		httpx.RespondError(c, httpx.BindError(err))
		return
	}
	resp, err := h.svc.ListarEstados(c.Request.Context(), q.EmpresaID)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}

func (h *Handler) obtenerEstado(c *gin.Context) {
	id, err := parseID32(c, "id")
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	resp, err := h.svc.ObtenerEstado(c.Request.Context(), id)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) actualizarEstado(c *gin.Context) {
	id, err := parseID32(c, "id")
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	var req ActualizarEstadoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondError(c, httpx.BindError(err))
		return
	}
	resp, err := h.svc.ActualizarEstado(c.Request.Context(), id, req)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) eliminarEstado(c *gin.Context) {
	id, err := parseID32(c, "id")
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	if err := h.svc.EliminarEstado(c.Request.Context(), id); err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// =====================================================================
// Motivos de descarte
// =====================================================================

func (h *Handler) crearMotivo(c *gin.Context) {
	var req CrearMotivoDescarteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondError(c, httpx.BindError(err))
		return
	}
	resp, err := h.svc.CrearMotivo(c.Request.Context(), req)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *Handler) listarMotivos(c *gin.Context) {
	var q ListarCatalogoQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		httpx.RespondError(c, httpx.BindError(err))
		return
	}
	resp, err := h.svc.ListarMotivos(c.Request.Context(), q)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}

func (h *Handler) obtenerMotivo(c *gin.Context) {
	id, err := parseID32(c, "id")
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	resp, err := h.svc.ObtenerMotivo(c.Request.Context(), id)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) actualizarMotivo(c *gin.Context) {
	id, err := parseID32(c, "id")
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	var req ActualizarMotivoDescarteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondError(c, httpx.BindError(err))
		return
	}
	resp, err := h.svc.ActualizarMotivo(c.Request.Context(), id, req)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) eliminarMotivo(c *gin.Context) {
	id, err := parseID32(c, "id")
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	if err := h.svc.EliminarMotivo(c.Request.Context(), id); err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// =====================================================================
// Leads
// =====================================================================

func (h *Handler) crearLead(c *gin.Context) {
	var req CrearLeadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondError(c, httpx.BindError(err))
		return
	}
	resp, err := h.svc.CrearLead(c.Request.Context(), req)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *Handler) listarLeads(c *gin.Context) {
	var q ListarLeadsQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		httpx.RespondError(c, httpx.BindError(err))
		return
	}
	resp, err := h.svc.ListarLeads(c.Request.Context(), q)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) obtenerLead(c *gin.Context) {
	id, err := parseID64(c, "id")
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	resp, err := h.svc.ObtenerLead(c.Request.Context(), id)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) actualizarLead(c *gin.Context) {
	id, err := parseID64(c, "id")
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	var req ActualizarLeadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondError(c, httpx.BindError(err))
		return
	}
	resp, err := h.svc.ActualizarLead(c.Request.Context(), id, req)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) cambiarEstadoLead(c *gin.Context) {
	id, err := parseID64(c, "id")
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	var req CambiarEstadoLeadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondError(c, httpx.BindError(err))
		return
	}
	resp, err := h.svc.CambiarEstadoLead(c.Request.Context(), id, req)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) convertirLead(c *gin.Context) {
	id, err := parseID64(c, "id")
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	var req ConvertirLeadRequest
	if c.Request.ContentLength > 0 {
		if err := c.ShouldBindJSON(&req); err != nil {
			httpx.RespondError(c, httpx.BindError(err))
			return
		}
	}
	resp, err := h.svc.ConvertirLead(c.Request.Context(), id, req)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) eliminarLead(c *gin.Context) {
	id, err := parseID64(c, "id")
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	if err := h.svc.EliminarLead(c.Request.Context(), id); err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// =====================================================================
// Actividades
// =====================================================================

func (h *Handler) crearActividad(c *gin.Context) {
	leadID, err := parseID64(c, "id")
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	var req CrearActividadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondError(c, httpx.BindError(err))
		return
	}
	resp, err := h.svc.CrearActividad(c.Request.Context(), leadID, req)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *Handler) listarActividadesLead(c *gin.Context) {
	leadID, err := parseID64(c, "id")
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	var q ListarActividadesQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		httpx.RespondError(c, httpx.BindError(err))
		return
	}
	resp, err := h.svc.ListarActividades(c.Request.Context(), &leadID, q)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) listarActividadesGlobal(c *gin.Context) {
	var q ListarActividadesQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		httpx.RespondError(c, httpx.BindError(err))
		return
	}
	resp, err := h.svc.ListarActividades(c.Request.Context(), nil, q)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) actualizarActividad(c *gin.Context) {
	id, err := parseID64(c, "id")
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	var req ActualizarActividadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondError(c, httpx.BindError(err))
		return
	}
	resp, err := h.svc.ActualizarActividad(c.Request.Context(), id, req)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) eliminarActividad(c *gin.Context) {
	id, err := parseID64(c, "id")
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	if err := h.svc.EliminarActividad(c.Request.Context(), id); err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// =====================================================================
// Intereses en productos
// =====================================================================

func (h *Handler) agregarInteres(c *gin.Context) {
	leadID, err := parseID64(c, "id")
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	var req CrearInteresProductoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondError(c, httpx.BindError(err))
		return
	}
	resp, err := h.svc.AgregarInteres(c.Request.Context(), leadID, req)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *Handler) listarIntereses(c *gin.Context) {
	leadID, err := parseID64(c, "id")
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	resp, err := h.svc.ListarIntereses(c.Request.Context(), leadID)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}

func (h *Handler) eliminarInteres(c *gin.Context) {
	leadID, err := parseID64(c, "id")
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	interesID, err := parseID64(c, "interesId")
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	if err := h.svc.EliminarInteres(c.Request.Context(), leadID, interesID); err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// =====================================================================
// Helpers de parsing
// =====================================================================

var errIDInvalido = errors.New("id inválido")

func parseID32(c *gin.Context, param string) (int32, error) {
	id, err := strconv.ParseInt(c.Param(param), 10, 32)
	if err != nil || id <= 0 {
		return 0, httpx.NewError(http.StatusBadRequest, "id_invalido", errIDInvalido.Error())
	}
	return int32(id), nil
}

func parseID64(c *gin.Context, param string) (int64, error) {
	id, err := strconv.ParseInt(c.Param(param), 10, 64)
	if err != nil || id <= 0 {
		return 0, httpx.NewError(http.StatusBadRequest, "id_invalido", errIDInvalido.Error())
	}
	return id, nil
}
