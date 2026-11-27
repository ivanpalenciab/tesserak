package crm

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"tesserak/backend/internal/auth"
	"tesserak/backend/internal/httpx"
)

const (
	PrioridadBaja  = "BAJA"
	PrioridadMedia = "MEDIA"
	PrioridadAlta  = "ALTA"
)

var (
	errOportunidadNotFound = errors.New("oportunidad no encontrada")

	ErrOportunidadNoEncontrada = httpx.NewError(http.StatusNotFound, "oportunidad_no_encontrada", "la oportunidad no existe")
)

type Oportunidad struct {
	ID                  int64          `gorm:"column:id;primaryKey"`
	EmpresaID           int32          `gorm:"column:empresa_id;not null"`
	SucursalID          *int32         `gorm:"column:sucursal_id"`
	LeadID              *int64         `gorm:"column:lead_id"`
	CampanaID           *int64         `gorm:"column:campana_id"`
	EstadoOportunidadID int32          `gorm:"column:estado_oportunidad_id;not null"`
	MotivoDescarteID    *int32         `gorm:"column:motivo_descarte_id"`
	UsuarioAsignadoID   *int64         `gorm:"column:usuario_asignado_id"`
	Titulo              string         `gorm:"column:titulo;not null"`
	Descripcion         *string        `gorm:"column:descripcion"`
	NombreCliente       *string        `gorm:"column:nombre_cliente"`
	ValorEstimado       float64        `gorm:"column:valor_estimado;not null"`
	MonedaID            *int32         `gorm:"column:moneda_id"`
	ProbabilidadPct     int32          `gorm:"column:probabilidad_pct;not null"`
	FechaCierreEstimada *time.Time     `gorm:"column:fecha_cierre_estimada;type:date"`
	FechaCierreReal     *time.Time     `gorm:"column:fecha_cierre_real"`
	Prioridad           string         `gorm:"column:prioridad;not null"`
	CreatedBy           int64          `gorm:"column:created_by;not null"`
	CreatedAt           time.Time      `gorm:"column:created_at"`
	UpdatedAt           time.Time      `gorm:"column:updated_at"`
	DeletedAt           gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (Oportunidad) TableName() string { return "crm_oportunidades" }

// =====================================================================
// DTOs
// =====================================================================

type CrearOportunidadRequest struct {
	EmpresaID           int32    `json:"empresa_id"            binding:"required,gt=0"`
	SucursalID          *int32   `json:"sucursal_id"           binding:"omitnil,gt=0"`
	LeadID              *int64   `json:"lead_id"               binding:"omitnil,gt=0"`
	CampanaID           *int64   `json:"campana_id"            binding:"omitnil,gt=0"`
	EstadoOportunidadID int32    `json:"estado_oportunidad_id" binding:"required,gt=0"`
	MotivoDescarteID    *int32   `json:"motivo_descarte_id"    binding:"omitnil,gt=0"`
	UsuarioAsignadoID   *int64   `json:"usuario_asignado_id"   binding:"omitnil,gt=0"`
	Titulo              string   `json:"titulo"                binding:"required,min=1,max=150"`
	Descripcion         *string  `json:"descripcion"           binding:"omitnil,max=5000"`
	NombreCliente       *string  `json:"nombre_cliente"        binding:"omitnil,max=150"`
	ValorEstimado       *float64 `json:"valor_estimado"        binding:"omitnil,gte=0"`
	MonedaID            *int32   `json:"moneda_id"             binding:"omitnil,gt=0"`
	ProbabilidadPct     *int32   `json:"probabilidad_pct"      binding:"omitnil,gte=0,lte=100"`
	FechaCierreEstimada *string  `json:"fecha_cierre_estimada"`
	Prioridad           *string  `json:"prioridad"             binding:"omitnil,oneof=BAJA MEDIA ALTA"`
	CreatedBy           int64    `json:"created_by"            binding:"required,gt=0"`
}

type ActualizarOportunidadRequest struct {
	SucursalID          *int32   `json:"sucursal_id"           binding:"omitnil,gt=0"`
	LeadID              *int64   `json:"lead_id"               binding:"omitnil,gt=0"`
	CampanaID           *int64   `json:"campana_id"            binding:"omitnil,gt=0"`
	EstadoOportunidadID *int32   `json:"estado_oportunidad_id" binding:"omitnil,gt=0"`
	MotivoDescarteID    *int32   `json:"motivo_descarte_id"    binding:"omitnil,gt=0"`
	UsuarioAsignadoID   *int64   `json:"usuario_asignado_id"   binding:"omitnil,gt=0"`
	Titulo              *string  `json:"titulo"                binding:"omitnil,min=1,max=150"`
	Descripcion         *string  `json:"descripcion"           binding:"omitnil,max=5000"`
	NombreCliente       *string  `json:"nombre_cliente"        binding:"omitnil,max=150"`
	ValorEstimado       *float64 `json:"valor_estimado"        binding:"omitnil,gte=0"`
	MonedaID            *int32   `json:"moneda_id"             binding:"omitnil,gt=0"`
	ProbabilidadPct     *int32   `json:"probabilidad_pct"      binding:"omitnil,gte=0,lte=100"`
	FechaCierreEstimada *string  `json:"fecha_cierre_estimada"`
	Prioridad           *string  `json:"prioridad"             binding:"omitnil,oneof=BAJA MEDIA ALTA"`
}

type CambiarEstadoOportunidadRequest struct {
	EstadoOportunidadID int32  `json:"estado_oportunidad_id" binding:"required,gt=0"`
	MotivoDescarteID    *int32 `json:"motivo_descarte_id"    binding:"omitnil,gt=0"`
}

type ListarOportunidadesQuery struct {
	EmpresaID           int32   `form:"empresa_id"            binding:"required,gt=0"`
	SucursalIDsScope    []int32 `form:"-"`
	EstadoOportunidadID int32   `form:"estado_oportunidad_id" binding:"omitempty,gt=0"`
	LeadID              int64  `form:"lead_id"               binding:"omitempty,gt=0"`
	CampanaID           int64  `form:"campana_id"            binding:"omitempty,gt=0"`
	Prioridad           string `form:"prioridad"             binding:"omitempty,oneof=BAJA MEDIA ALTA"`
	Q                   string `form:"q"                     binding:"omitempty,max=100"`
	Page                int    `form:"page"                  binding:"omitempty,gte=1"`
	PageSize            int    `form:"page_size"             binding:"omitempty,gte=1,lte=100"`
}

type OportunidadResponse struct {
	ID                  int64      `json:"id"`
	EmpresaID           int32      `json:"empresa_id"`
	SucursalID          *int32     `json:"sucursal_id"`
	LeadID              *int64     `json:"lead_id"`
	CampanaID           *int64     `json:"campana_id"`
	EstadoOportunidadID int32      `json:"estado_oportunidad_id"`
	MotivoDescarteID    *int32     `json:"motivo_descarte_id"`
	UsuarioAsignadoID   *int64     `json:"usuario_asignado_id"`
	Titulo              string     `json:"titulo"`
	Descripcion         *string    `json:"descripcion"`
	NombreCliente       *string    `json:"nombre_cliente"`
	ValorEstimado       float64    `json:"valor_estimado"`
	MonedaID            *int32     `json:"moneda_id"`
	ProbabilidadPct     int32      `json:"probabilidad_pct"`
	FechaCierreEstimada *string    `json:"fecha_cierre_estimada"`
	FechaCierreReal     *time.Time `json:"fecha_cierre_real"`
	Prioridad           string     `json:"prioridad"`
	CreatedBy           int64      `json:"created_by"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

type ListaOportunidadesResponse struct {
	Data     []OportunidadResponse `json:"data"`
	Page     int                   `json:"page"`
	PageSize int                   `json:"page_size"`
	Total    int64                 `json:"total"`
}

func toOportunidadResponse(o *Oportunidad) OportunidadResponse {
	var fechaEst *string
	if o.FechaCierreEstimada != nil {
		s := o.FechaCierreEstimada.Format(dateLayout)
		fechaEst = &s
	}
	return OportunidadResponse{
		ID:                  o.ID,
		EmpresaID:           o.EmpresaID,
		SucursalID:          o.SucursalID,
		LeadID:              o.LeadID,
		CampanaID:           o.CampanaID,
		EstadoOportunidadID: o.EstadoOportunidadID,
		MotivoDescarteID:    o.MotivoDescarteID,
		UsuarioAsignadoID:   o.UsuarioAsignadoID,
		Titulo:              o.Titulo,
		Descripcion:         o.Descripcion,
		NombreCliente:       o.NombreCliente,
		ValorEstimado:       o.ValorEstimado,
		MonedaID:            o.MonedaID,
		ProbabilidadPct:     o.ProbabilidadPct,
		FechaCierreEstimada: fechaEst,
		FechaCierreReal:     o.FechaCierreReal,
		Prioridad:           o.Prioridad,
		CreatedBy:           o.CreatedBy,
		CreatedAt:           o.CreatedAt,
		UpdatedAt:           o.UpdatedAt,
	}
}

// =====================================================================
// Repository
// =====================================================================

func (r *Repository) CreateOportunidad(ctx context.Context, o *Oportunidad) error {
	return r.db.WithContext(ctx).Create(o).Error
}

func (r *Repository) FindOportunidadByID(ctx context.Context, id int64) (*Oportunidad, error) {
	var o Oportunidad
	err := r.db.WithContext(ctx).First(&o, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errOportunidadNotFound
	}
	return &o, err
}

func (r *Repository) ListOportunidades(ctx context.Context, f ListarOportunidadesQuery) ([]Oportunidad, int64, error) {
	q := r.db.WithContext(ctx).Model(&Oportunidad{}).Where("empresa_id = ?", f.EmpresaID)

	if len(f.SucursalIDsScope) > 0 {
		q = q.Where("(sucursal_id IS NULL OR sucursal_id IN ?)", f.SucursalIDsScope)
	}

	if f.EstadoOportunidadID > 0 {
		q = q.Where("estado_oportunidad_id = ?", f.EstadoOportunidadID)
	}
	if f.LeadID > 0 {
		q = q.Where("lead_id = ?", f.LeadID)
	}
	if f.CampanaID > 0 {
		q = q.Where("campana_id = ?", f.CampanaID)
	}
	if f.Prioridad != "" {
		q = q.Where("prioridad = ?", f.Prioridad)
	}
	if s := strings.TrimSpace(f.Q); s != "" {
		like := "%" + escapeLike(s) + "%"
		q = q.Where("titulo ILIKE ? OR nombre_cliente ILIKE ? OR descripcion ILIKE ?", like, like, like)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var items []Oportunidad
	err := q.Order("created_at DESC, id DESC").
		Offset((f.Page - 1) * f.PageSize).
		Limit(f.PageSize).
		Find(&items).Error
	return items, total, err
}

func (r *Repository) UpdateOportunidad(ctx context.Context, o *Oportunidad, cambios map[string]any) error {
	if len(cambios) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Model(o).Updates(cambios).Error
}

func (r *Repository) SoftDeleteOportunidad(ctx context.Context, id int64) error {
	res := r.db.WithContext(ctx).Delete(&Oportunidad{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errOportunidadNotFound
	}
	return nil
}

// =====================================================================
// Service
// =====================================================================

func (s *Service) CrearOportunidad(ctx context.Context, req CrearOportunidadRequest) (*OportunidadResponse, error) {
	var fechaEst *time.Time
	if req.FechaCierreEstimada != nil && strings.TrimSpace(*req.FechaCierreEstimada) != "" {
		f, err := parseDate(*req.FechaCierreEstimada)
		if err != nil {
			return nil, ErrFechaInvalida
		}
		fechaEst = &f
	}

	ok, err := s.repo.UsuarioPerteneceAEmpresa(ctx, req.CreatedBy, req.EmpresaID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrUsuarioOtraEmpresa
	}
	if req.UsuarioAsignadoID != nil {
		ok, err := s.repo.UsuarioPerteneceAEmpresa(ctx, *req.UsuarioAsignadoID, req.EmpresaID)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, ErrUsuarioOtraEmpresa
		}
	}

	etapa, err := s.repo.FindEtapaOportunidadByID(ctx, req.EstadoOportunidadID)
	if err != nil {
		return nil, mapEtapaOportunidadError(err)
	}
	if etapa.EmpresaID != req.EmpresaID {
		return nil, ErrEstadoOtraEmpresa
	}

	prob := etapa.ProbabilidadPct
	if req.ProbabilidadPct != nil {
		prob = *req.ProbabilidadPct
	}
	var valor float64
	if req.ValorEstimado != nil {
		valor = *req.ValorEstimado
	}
	prioridad := PrioridadMedia
	if req.Prioridad != nil {
		prioridad = *req.Prioridad
	}
	var fechaCierreReal *time.Time
	if etapa.EsGanada || etapa.EsPerdida {
		now := time.Now().UTC()
		fechaCierreReal = &now
	}

	o := &Oportunidad{
		EmpresaID:           req.EmpresaID,
		SucursalID:          req.SucursalID,
		LeadID:              req.LeadID,
		CampanaID:           req.CampanaID,
		EstadoOportunidadID: req.EstadoOportunidadID,
		MotivoDescarteID:    req.MotivoDescarteID,
		UsuarioAsignadoID:   req.UsuarioAsignadoID,
		Titulo:              strings.TrimSpace(req.Titulo),
		Descripcion:         trimPtr(req.Descripcion),
		NombreCliente:       trimPtr(req.NombreCliente),
		ValorEstimado:       valor,
		MonedaID:            req.MonedaID,
		ProbabilidadPct:     prob,
		FechaCierreEstimada: fechaEst,
		FechaCierreReal:     fechaCierreReal,
		Prioridad:           prioridad,
		CreatedBy:           req.CreatedBy,
	}
	if err := s.repo.CreateOportunidad(ctx, o); err != nil {
		return nil, mapOportunidadError(err)
	}
	resp := toOportunidadResponse(o)
	return &resp, nil
}

func (s *Service) ObtenerOportunidad(ctx context.Context, id int64) (*OportunidadResponse, error) {
	o, err := s.repo.FindOportunidadByID(ctx, id)
	if err != nil {
		return nil, mapOportunidadError(err)
	}
	resp := toOportunidadResponse(o)
	return &resp, nil
}

func (s *Service) ListarOportunidades(ctx context.Context, q ListarOportunidadesQuery) (*ListaOportunidadesResponse, error) {
	q.Page, q.PageSize = normalizarPaginacion(q.Page, q.PageSize)
	items, total, err := s.repo.ListOportunidades(ctx, q)
	if err != nil {
		return nil, err
	}
	data := make([]OportunidadResponse, 0, len(items))
	for i := range items {
		data = append(data, toOportunidadResponse(&items[i]))
	}
	return &ListaOportunidadesResponse{Data: data, Page: q.Page, PageSize: q.PageSize, Total: total}, nil
}

func (s *Service) ActualizarOportunidad(ctx context.Context, id int64, req ActualizarOportunidadRequest) (*OportunidadResponse, error) {
	var parsedFechaEst *time.Time
	clearFechaEst := false
	if req.FechaCierreEstimada != nil {
		if strings.TrimSpace(*req.FechaCierreEstimada) == "" {
			clearFechaEst = true
		} else {
			f, err := parseDate(*req.FechaCierreEstimada)
			if err != nil {
				return nil, ErrFechaInvalida
			}
			parsedFechaEst = &f
		}
	}

	var res *Oportunidad
	err := s.repo.WithTx(ctx, func(tx *Repository) error {
		o, err := tx.FindOportunidadByID(ctx, id)
		if err != nil {
			return err
		}

		cambios := map[string]any{}
		if req.UsuarioAsignadoID != nil {
			ok, err := tx.UsuarioPerteneceAEmpresa(ctx, *req.UsuarioAsignadoID, o.EmpresaID)
			if err != nil {
				return err
			}
			if !ok {
				return ErrUsuarioOtraEmpresa
			}
			cambios["usuario_asignado_id"] = *req.UsuarioAsignadoID
		}
		if req.SucursalID != nil {
			cambios["sucursal_id"] = *req.SucursalID
		}
		if req.LeadID != nil {
			cambios["lead_id"] = *req.LeadID
		}
		if req.CampanaID != nil {
			cambios["campana_id"] = *req.CampanaID
		}
		if req.EstadoOportunidadID != nil {
			etapa, err := tx.FindEtapaOportunidadByID(ctx, *req.EstadoOportunidadID)
			if err != nil {
				return err
			}
			if etapa.EmpresaID != o.EmpresaID {
				return ErrEstadoOtraEmpresa
			}
			cambios["estado_oportunidad_id"] = *req.EstadoOportunidadID
			if etapa.EsGanada || etapa.EsPerdida {
				now := time.Now().UTC()
				cambios["fecha_cierre_real"] = now
			} else {
				cambios["fecha_cierre_real"] = nil
			}
		}
		if req.MotivoDescarteID != nil {
			cambios["motivo_descarte_id"] = *req.MotivoDescarteID
		}
		if req.Titulo != nil {
			cambios["titulo"] = strings.TrimSpace(*req.Titulo)
		}
		if req.Descripcion != nil {
			cambios["descripcion"] = trimPtr(req.Descripcion)
		}
		if req.NombreCliente != nil {
			cambios["nombre_cliente"] = trimPtr(req.NombreCliente)
		}
		if req.ValorEstimado != nil {
			cambios["valor_estimado"] = *req.ValorEstimado
		}
		if req.MonedaID != nil {
			cambios["moneda_id"] = *req.MonedaID
		}
		if req.ProbabilidadPct != nil {
			cambios["probabilidad_pct"] = *req.ProbabilidadPct
		}
		if clearFechaEst {
			cambios["fecha_cierre_estimada"] = nil
		} else if parsedFechaEst != nil {
			cambios["fecha_cierre_estimada"] = *parsedFechaEst
		}
		if req.Prioridad != nil {
			cambios["prioridad"] = *req.Prioridad
		}

		if err := tx.UpdateOportunidad(ctx, o, cambios); err != nil {
			return err
		}
		res, err = tx.FindOportunidadByID(ctx, id)
		return err
	})
	if err != nil {
		return nil, mapOportunidadError(err)
	}
	resp := toOportunidadResponse(res)
	return &resp, nil
}

func (s *Service) CambiarEstadoOportunidad(ctx context.Context, id int64, req CambiarEstadoOportunidadRequest) (*OportunidadResponse, error) {
	var res *Oportunidad
	err := s.repo.WithTx(ctx, func(tx *Repository) error {
		o, err := tx.FindOportunidadByID(ctx, id)
		if err != nil {
			return err
		}
		etapa, err := tx.FindEtapaOportunidadByID(ctx, req.EstadoOportunidadID)
		if err != nil {
			return err
		}
		if etapa.EmpresaID != o.EmpresaID {
			return ErrEstadoOtraEmpresa
		}

		cambios := map[string]any{
			"estado_oportunidad_id": req.EstadoOportunidadID,
			"probabilidad_pct":      etapa.ProbabilidadPct,
		}
		if req.MotivoDescarteID != nil {
			cambios["motivo_descarte_id"] = *req.MotivoDescarteID
		} else if !etapa.EsPerdida {
			cambios["motivo_descarte_id"] = nil
		}
		if etapa.EsGanada || etapa.EsPerdida {
			now := time.Now().UTC()
			cambios["fecha_cierre_real"] = now
		} else {
			cambios["fecha_cierre_real"] = nil
		}

		if err := tx.UpdateOportunidad(ctx, o, cambios); err != nil {
			return err
		}
		res, err = tx.FindOportunidadByID(ctx, id)
		return err
	})
	if err != nil {
		return nil, mapOportunidadError(err)
	}
	resp := toOportunidadResponse(res)
	return &resp, nil
}

func (s *Service) EliminarOportunidad(ctx context.Context, id int64) error {
	return mapOportunidadError(s.repo.SoftDeleteOportunidad(ctx, id))
}

func mapOportunidadError(err error) error {
	if errors.Is(err, errOportunidadNotFound) {
		return ErrOportunidadNoEncontrada
	}
	if errors.Is(err, errEtapaOportunidadNotFound) {
		return ErrEtapaOportunidadNoEncontrada
	}
	return mapError(err)
}

// =====================================================================
// Handlers
// =====================================================================

func (h *Handler) registerOportunidadRoutes(g *gin.RouterGroup) {
	ops := g.Group("/oportunidades")
	ops.POST("", h.crearOportunidad)
	ops.GET("", h.listarOportunidades)
	ops.GET("/:id", h.obtenerOportunidad)
	ops.PATCH("/:id", h.actualizarOportunidad)
	ops.PATCH("/:id/estado", h.cambiarEstadoOportunidad)
	ops.DELETE("/:id", h.eliminarOportunidad)
}

func (h *Handler) crearOportunidad(c *gin.Context) {
	var req CrearOportunidadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondError(c, httpx.BindError(err))
		return
	}
	if claims, ok := auth.ClaimsFromContext(c); ok && req.SucursalID != nil && !claims.HasSucursal(*req.SucursalID) {
		httpx.RespondError(c, auth.ErrSucursalNoAutorizada)
		return
	}
	resp, err := h.svc.CrearOportunidad(c.Request.Context(), req)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *Handler) listarOportunidades(c *gin.Context) {
	var q ListarOportunidadesQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		httpx.RespondError(c, httpx.BindError(err))
		return
	}
	if claims, ok := auth.ClaimsFromContext(c); ok {
		q.SucursalIDsScope = claims.AllowedSucursalIDs()
	}
	resp, err := h.svc.ListarOportunidades(c.Request.Context(), q)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) obtenerOportunidad(c *gin.Context) {
	id, err := parseID64(c, "id")
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	resp, err := h.svc.ObtenerOportunidad(c.Request.Context(), id)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	if claims, ok := auth.ClaimsFromContext(c); ok && resp.SucursalID != nil && !claims.HasSucursal(*resp.SucursalID) {
		httpx.RespondError(c, auth.ErrSucursalNoAutorizada)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) actualizarOportunidad(c *gin.Context) {
	id, err := parseID64(c, "id")
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	var req ActualizarOportunidadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondError(c, httpx.BindError(err))
		return
	}
	resp, err := h.svc.ActualizarOportunidad(c.Request.Context(), id, req)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) cambiarEstadoOportunidad(c *gin.Context) {
	id, err := parseID64(c, "id")
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	var req CambiarEstadoOportunidadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondError(c, httpx.BindError(err))
		return
	}
	resp, err := h.svc.CambiarEstadoOportunidad(c.Request.Context(), id, req)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) eliminarOportunidad(c *gin.Context) {
	id, err := parseID64(c, "id")
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	if err := h.svc.EliminarOportunidad(c.Request.Context(), id); err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
