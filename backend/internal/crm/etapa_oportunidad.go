package crm

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"tesserak/backend/internal/httpx"
)

var (
	errEtapaOportunidadNotFound = errors.New("etapa de oportunidad no encontrada")

	ErrEtapaOportunidadNoEncontrada = httpx.NewError(http.StatusNotFound, "etapa_oportunidad_no_encontrada", "la etapa de oportunidad no existe")
	ErrEtapaGanadaYPerdida          = httpx.NewError(http.StatusBadRequest, "etapa_cierre_invalida", "una etapa no puede ser ganada y perdida al mismo tiempo")
)

type EtapaOportunidad struct {
	ID              int32          `gorm:"column:id;primaryKey"`
	EmpresaID       int32          `gorm:"column:empresa_id;not null"`
	Nombre          string         `gorm:"column:nombre;not null"`
	Orden           int32          `gorm:"column:orden;not null"`
	ProbabilidadPct int32          `gorm:"column:probabilidad_pct;not null"`
	EsGanada        bool           `gorm:"column:es_ganada;not null"`
	EsPerdida       bool           `gorm:"column:es_perdida;not null"`
	Activo          bool           `gorm:"column:activo;not null"`
	CreatedAt       time.Time      `gorm:"column:created_at"`
	UpdatedAt       time.Time      `gorm:"column:updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (EtapaOportunidad) TableName() string { return "crm_etapa_oportunidad" }

type CrearEtapaOportunidadRequest struct {
	EmpresaID       int32  `json:"empresa_id"       binding:"required,gt=0"`
	Nombre          string `json:"nombre"           binding:"required,min=1,max=80"`
	Orden           *int32 `json:"orden"            binding:"omitnil,gte=0"`
	ProbabilidadPct *int32 `json:"probabilidad_pct" binding:"omitnil,gte=0,lte=100"`
	EsGanada        *bool  `json:"es_ganada"`
	EsPerdida       *bool  `json:"es_perdida"`
	Activo          *bool  `json:"activo"`
}

type ActualizarEtapaOportunidadRequest struct {
	Nombre          *string `json:"nombre"           binding:"omitnil,min=1,max=80"`
	Orden           *int32  `json:"orden"            binding:"omitnil,gte=0"`
	ProbabilidadPct *int32  `json:"probabilidad_pct" binding:"omitnil,gte=0,lte=100"`
	EsGanada        *bool   `json:"es_ganada"`
	EsPerdida       *bool   `json:"es_perdida"`
	Activo          *bool   `json:"activo"`
}

type EtapaOportunidadResponse struct {
	ID              int32     `json:"id"`
	EmpresaID       int32     `json:"empresa_id"`
	Nombre          string    `json:"nombre"`
	Orden           int32     `json:"orden"`
	ProbabilidadPct int32     `json:"probabilidad_pct"`
	EsGanada        bool      `json:"es_ganada"`
	EsPerdida       bool      `json:"es_perdida"`
	Activo          bool      `json:"activo"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func toEtapaOportunidadResponse(e *EtapaOportunidad) EtapaOportunidadResponse {
	return EtapaOportunidadResponse{
		ID:              e.ID,
		EmpresaID:       e.EmpresaID,
		Nombre:          e.Nombre,
		Orden:           e.Orden,
		ProbabilidadPct: e.ProbabilidadPct,
		EsGanada:        e.EsGanada,
		EsPerdida:       e.EsPerdida,
		Activo:          e.Activo,
		CreatedAt:       e.CreatedAt,
		UpdatedAt:       e.UpdatedAt,
	}
}

// Repository

func (r *Repository) CreateEtapaOportunidad(ctx context.Context, e *EtapaOportunidad) error {
	return r.db.WithContext(ctx).Create(e).Error
}

func (r *Repository) ListEtapasOportunidad(ctx context.Context, q ListarCatalogoQuery) ([]EtapaOportunidad, error) {
	db := r.db.WithContext(ctx).Where("empresa_id = ?", q.EmpresaID)
	if q.Activo != nil {
		db = db.Where("activo = ?", *q.Activo)
	}
	var items []EtapaOportunidad
	err := db.Order("orden ASC, id ASC").Find(&items).Error
	return items, err
}

func (r *Repository) FindEtapaOportunidadByID(ctx context.Context, id int32) (*EtapaOportunidad, error) {
	var e EtapaOportunidad
	err := r.db.WithContext(ctx).First(&e, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errEtapaOportunidadNotFound
	}
	return &e, err
}

func (r *Repository) UpdateEtapaOportunidad(ctx context.Context, e *EtapaOportunidad, cambios map[string]any) error {
	if len(cambios) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Model(e).Updates(cambios).Error
}

func (r *Repository) SoftDeleteEtapaOportunidad(ctx context.Context, id int32) error {
	res := r.db.WithContext(ctx).Delete(&EtapaOportunidad{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errEtapaOportunidadNotFound
	}
	return nil
}

// Service

func (s *Service) CrearEtapaOportunidad(ctx context.Context, req CrearEtapaOportunidadRequest) (*EtapaOportunidadResponse, error) {
	esGanada := req.EsGanada != nil && *req.EsGanada
	esPerdida := req.EsPerdida != nil && *req.EsPerdida
	if esGanada && esPerdida {
		return nil, ErrEtapaGanadaYPerdida
	}
	orden := int32(1)
	if req.Orden != nil {
		orden = *req.Orden
	}
	prob := int32(0)
	if req.ProbabilidadPct != nil {
		prob = *req.ProbabilidadPct
	}
	activo := true
	if req.Activo != nil {
		activo = *req.Activo
	}

	e := &EtapaOportunidad{
		EmpresaID:       req.EmpresaID,
		Nombre:          strings.TrimSpace(req.Nombre),
		Orden:           orden,
		ProbabilidadPct: prob,
		EsGanada:        esGanada,
		EsPerdida:       esPerdida,
		Activo:          activo,
	}
	if err := s.repo.CreateEtapaOportunidad(ctx, e); err != nil {
		return nil, mapEtapaOportunidadError(err)
	}
	resp := toEtapaOportunidadResponse(e)
	return &resp, nil
}

func (s *Service) ListarEtapasOportunidad(ctx context.Context, q ListarCatalogoQuery) ([]EtapaOportunidadResponse, error) {
	items, err := s.repo.ListEtapasOportunidad(ctx, q)
	if err != nil {
		return nil, err
	}
	out := make([]EtapaOportunidadResponse, 0, len(items))
	for i := range items {
		out = append(out, toEtapaOportunidadResponse(&items[i]))
	}
	return out, nil
}

func (s *Service) ObtenerEtapaOportunidad(ctx context.Context, id int32) (*EtapaOportunidadResponse, error) {
	e, err := s.repo.FindEtapaOportunidadByID(ctx, id)
	if err != nil {
		return nil, mapEtapaOportunidadError(err)
	}
	resp := toEtapaOportunidadResponse(e)
	return &resp, nil
}

func (s *Service) ActualizarEtapaOportunidad(ctx context.Context, id int32, req ActualizarEtapaOportunidadRequest) (*EtapaOportunidadResponse, error) {
	var res *EtapaOportunidad
	err := s.repo.WithTx(ctx, func(tx *Repository) error {
		e, err := tx.FindEtapaOportunidadByID(ctx, id)
		if err != nil {
			return err
		}

		effGanada := e.EsGanada
		if req.EsGanada != nil {
			effGanada = *req.EsGanada
		}
		effPerdida := e.EsPerdida
		if req.EsPerdida != nil {
			effPerdida = *req.EsPerdida
		}
		if effGanada && effPerdida {
			return ErrEtapaGanadaYPerdida
		}

		cambios := map[string]any{}
		if req.Nombre != nil {
			cambios["nombre"] = strings.TrimSpace(*req.Nombre)
		}
		if req.Orden != nil {
			cambios["orden"] = *req.Orden
		}
		if req.ProbabilidadPct != nil {
			cambios["probabilidad_pct"] = *req.ProbabilidadPct
		}
		if req.EsGanada != nil {
			cambios["es_ganada"] = *req.EsGanada
		}
		if req.EsPerdida != nil {
			cambios["es_perdida"] = *req.EsPerdida
		}
		if req.Activo != nil {
			cambios["activo"] = *req.Activo
		}
		if err := tx.UpdateEtapaOportunidad(ctx, e, cambios); err != nil {
			return err
		}
		res, err = tx.FindEtapaOportunidadByID(ctx, id)
		return err
	})
	if err != nil {
		return nil, mapEtapaOportunidadError(err)
	}
	resp := toEtapaOportunidadResponse(res)
	return &resp, nil
}

func (s *Service) EliminarEtapaOportunidad(ctx context.Context, id int32) error {
	return mapEtapaOportunidadError(s.repo.SoftDeleteEtapaOportunidad(ctx, id))
}

func mapEtapaOportunidadError(err error) error {
	if errors.Is(err, errEtapaOportunidadNotFound) {
		return ErrEtapaOportunidadNoEncontrada
	}
	return mapError(err)
}

// Handlers

func (h *Handler) registerEtapaOportunidadRoutes(g *gin.RouterGroup) {
	etapas := g.Group("/etapas-oportunidad")
	etapas.POST("", h.crearEtapaOportunidad)
	etapas.GET("", h.listarEtapasOportunidad)
	etapas.GET("/:id", h.obtenerEtapaOportunidad)
	etapas.PATCH("/:id", h.actualizarEtapaOportunidad)
	etapas.DELETE("/:id", h.eliminarEtapaOportunidad)
}

func (h *Handler) crearEtapaOportunidad(c *gin.Context) {
	var req CrearEtapaOportunidadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondError(c, httpx.BindError(err))
		return
	}
	resp, err := h.svc.CrearEtapaOportunidad(c.Request.Context(), req)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *Handler) listarEtapasOportunidad(c *gin.Context) {
	var q ListarCatalogoQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		httpx.RespondError(c, httpx.BindError(err))
		return
	}
	resp, err := h.svc.ListarEtapasOportunidad(c.Request.Context(), q)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}

func (h *Handler) obtenerEtapaOportunidad(c *gin.Context) {
	id, err := parseID32(c, "id")
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	resp, err := h.svc.ObtenerEtapaOportunidad(c.Request.Context(), id)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) actualizarEtapaOportunidad(c *gin.Context) {
	id, err := parseID32(c, "id")
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	var req ActualizarEtapaOportunidadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondError(c, httpx.BindError(err))
		return
	}
	resp, err := h.svc.ActualizarEtapaOportunidad(c.Request.Context(), id, req)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) eliminarEtapaOportunidad(c *gin.Context) {
	id, err := parseID32(c, "id")
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	if err := h.svc.EliminarEtapaOportunidad(c.Request.Context(), id); err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
