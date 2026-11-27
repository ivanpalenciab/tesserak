package lealtad

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"tesserak/backend/internal/auth"
	"tesserak/backend/internal/httpx"
)

const (
	TipoReglaPorMonto      = "POR_MONTO"
	TipoReglaBonoFijo      = "BONO_FIJO"
	TipoReglaMultiplicador = "MULTIPLICADOR"
)

var (
	errReglaNotFound        = errors.New("regla de lealtad no encontrada")
	ErrReglaNoEncontrada    = httpx.NewError(http.StatusNotFound, "regla_no_encontrada", "la regla de lealtad no existe")
	ErrSinPermisoRegla      = httpx.NewError(http.StatusForbidden, "sin_permiso_regla", "no tiene permisos para administrar esta regla de lealtad")
	ErrRequiereAdhesion     = httpx.NewError(http.StatusForbidden, "requiere_adhesion", "tu empresa debe estar adherida activamente a este plan para crear reglas propias")
)

type Regla struct {
	ID             int32          `gorm:"column:id;primaryKey;autoIncrement"`
	ProgramaID     int32          `gorm:"column:programa_id;not null"`
	EmpresaID      *int32         `gorm:"column:empresa_id"`
	Nombre         string         `gorm:"column:nombre;not null"`
	TipoRegla      string         `gorm:"column:tipo_regla;not null"`
	PuntosPorMonto float64        `gorm:"column:puntos_por_monto;not null"`
	MontoBase      float64        `gorm:"column:monto_base;not null"`
	CompraMinima   float64        `gorm:"column:compra_minima;not null"`
	MonedaID       int32          `gorm:"column:moneda_id;not null"`
	Prioridad      int32          `gorm:"column:prioridad;not null"`
	FechaInicio    time.Time      `gorm:"column:fecha_inicio;not null"`
	FechaFin       *time.Time     `gorm:"column:fecha_fin"`
	Activo         bool           `gorm:"column:activo;not null"`
	CreatedAt      time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt      time.Time      `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt      gorm.DeletedAt `gorm:"column:deleted_at"`

	Moneda  *MonedaRef  `gorm:"foreignKey:MonedaID;references:ID"`
	Empresa *EmpresaRef `gorm:"foreignKey:EmpresaID;references:ID"`
}

func (Regla) TableName() string { return "lealtad_reglas" }

type ReglaResponse struct {
	ID             int32      `json:"id"`
	ProgramaID     int32      `json:"programa_id"`
	EmpresaID      *int32     `json:"empresa_id,omitempty"`
	EmpresaNombre  string     `json:"empresa_nombre,omitempty"`
	EsGlobal       bool       `json:"es_global"`
	Nombre         string     `json:"nombre"`
	TipoRegla      string     `json:"tipo_regla"`
	PuntosPorMonto float64    `json:"puntos_por_monto"`
	MontoBase      float64    `json:"monto_base"`
	CompraMinima   float64    `json:"compra_minima"`
	MonedaID       int32      `json:"moneda_id"`
	MonedaCodigo   string     `json:"moneda_codigo,omitempty"`
	MonedaSimbolo  string     `json:"moneda_simbolo,omitempty"`
	Prioridad      int32      `json:"prioridad"`
	FechaInicio    time.Time  `json:"fecha_inicio"`
	FechaFin       *time.Time `json:"fecha_fin,omitempty"`
	Activo         bool       `json:"activo"`
	PuedeEditar    bool       `json:"puede_editar"`
	PuedeEliminar  bool       `json:"puede_eliminar"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type ListarReglasResult struct {
	Data              []ReglaResponse `json:"data"`
	PuedeCrearGlobal  bool            `json:"puede_crear_global"`
	PuedeCrearEmpresa bool            `json:"puede_crear_empresa"`
}

type CrearReglaRequest struct {
	Nombre         string     `json:"nombre"           binding:"required,min=2,max=100"`
	EsGlobal       bool       `json:"es_global"`
	TipoRegla      string     `json:"tipo_regla"       binding:"omitempty,oneof=POR_MONTO BONO_FIJO MULTIPLICADOR"`
	PuntosPorMonto float64    `json:"puntos_por_monto" binding:"required,gt=0"`
	MontoBase      *float64   `json:"monto_base"       binding:"omitnil,gt=0"`
	CompraMinima   *float64   `json:"compra_minima"    binding:"omitnil,gte=0"`
	MonedaID       *int32     `json:"moneda_id"        binding:"omitnil,gt=0"`
	Prioridad      *int32     `json:"prioridad"        binding:"omitnil,gte=1"`
	FechaInicio    *time.Time `json:"fecha_inicio"`
	FechaFin       *time.Time `json:"fecha_fin"`
	Activo         *bool      `json:"activo"`
}

type ActualizarReglaRequest struct {
	Nombre          *string    `json:"nombre"            binding:"omitnil,min=2,max=100"`
	TipoRegla       *string    `json:"tipo_regla"        binding:"omitnil,oneof=POR_MONTO BONO_FIJO MULTIPLICADOR"`
	PuntosPorMonto  *float64   `json:"puntos_por_monto"  binding:"omitnil,gt=0"`
	MontoBase       *float64   `json:"monto_base"        binding:"omitnil,gt=0"`
	CompraMinima    *float64   `json:"compra_minima"     binding:"omitnil,gte=0"`
	MonedaID        *int32     `json:"moneda_id"         binding:"omitnil,gt=0"`
	Prioridad       *int32     `json:"prioridad"         binding:"omitnil,gte=1"`
	FechaInicio     *time.Time `json:"fecha_inicio"`
	FechaFin        *time.Time `json:"fecha_fin"`
	LimpiarFechaFin bool       `json:"limpiar_fecha_fin"`
	Activo          *bool      `json:"activo"`
}

func toReglaResponse(r Regla, puedeEditar, puedeEliminar bool) ReglaResponse {
	resp := ReglaResponse{
		ID:             r.ID,
		ProgramaID:     r.ProgramaID,
		EmpresaID:      r.EmpresaID,
		EsGlobal:       r.EmpresaID == nil,
		Nombre:         r.Nombre,
		TipoRegla:      r.TipoRegla,
		PuntosPorMonto: r.PuntosPorMonto,
		MontoBase:      r.MontoBase,
		CompraMinima:   r.CompraMinima,
		MonedaID:       r.MonedaID,
		Prioridad:      r.Prioridad,
		FechaInicio:    r.FechaInicio,
		FechaFin:       r.FechaFin,
		Activo:         r.Activo,
		PuedeEditar:    puedeEditar,
		PuedeEliminar:  puedeEliminar,
		CreatedAt:      r.CreatedAt,
		UpdatedAt:      r.UpdatedAt,
	}
	if r.Moneda != nil {
		resp.MonedaCodigo = r.Moneda.CodigoISO
		resp.MonedaSimbolo = r.Moneda.Simbolo
	}
	if r.Empresa != nil {
		if strings.TrimSpace(r.Empresa.RazonSocial) != "" {
			resp.EmpresaNombre = r.Empresa.RazonSocial
		} else {
			resp.EmpresaNombre = r.Empresa.NombreLegal
		}
	}
	return resp
}

// =====================================================================
// Repository
// =====================================================================

func (r *Repository) ListReglas(ctx context.Context, programaID int32, empresaID int32, includeAllCompanies bool) ([]Regla, error) {
	q := r.db.WithContext(ctx).
		Model(&Regla{}).
		Preload("Moneda").
		Preload("Empresa").
		Where("programa_id = ?", programaID)

	if !includeAllCompanies {
		q = q.Where("empresa_id IS NULL OR empresa_id = ?", empresaID)
	}

	var items []Regla
	err := q.
		Order("CASE WHEN empresa_id IS NULL THEN 0 ELSE 1 END, prioridad ASC, id ASC").
		Find(&items).Error
	return items, err
}

func (r *Repository) FindReglaByID(ctx context.Context, programaID, reglaID int32) (*Regla, error) {
	var item Regla
	err := r.db.WithContext(ctx).
		Preload("Moneda").
		Preload("Empresa").
		Where("id = ? AND programa_id = ?", reglaID, programaID).
		First(&item).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errReglaNotFound
		}
		return nil, err
	}
	return &item, nil
}

func (r *Repository) CreateRegla(ctx context.Context, regla *Regla) error {
	return r.db.WithContext(ctx).Create(regla).Error
}

func (r *Repository) UpdateRegla(ctx context.Context, programaID, reglaID int32, updates map[string]any) error {
	res := r.db.WithContext(ctx).
		Model(&Regla{}).
		Where("id = ? AND programa_id = ? AND deleted_at IS NULL", reglaID, programaID).
		Updates(updates)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errReglaNotFound
	}
	return nil
}

func (r *Repository) DeleteRegla(ctx context.Context, programaID, reglaID int32) error {
	now := time.Now().UTC()
	res := r.db.WithContext(ctx).
		Model(&Regla{}).
		Where("id = ? AND programa_id = ? AND deleted_at IS NULL", reglaID, programaID).
		Updates(map[string]any{
			"activo":     false,
			"deleted_at": now,
			"updated_at": now,
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errReglaNotFound
	}
	return nil
}

// =====================================================================
// Service
// =====================================================================

func (s *Service) canManageRegla(claims *auth.Claims, canManageProg bool, r Regla) bool {
	if claims == nil || !claims.IsAdmin() {
		return false
	}
	if r.EmpresaID == nil {
		return canManageProg
	}
	return *r.EmpresaID == claims.EmpresaID || canManageProg
}

func (s *Service) ListarReglas(ctx context.Context, claims *auth.Claims, programaID int32) (*ListarReglasResult, error) {
	p, err := s.repo.FindProgramaByID(ctx, programaID)
	if err != nil {
		if errors.Is(err, errProgramaNotFound) {
			return nil, ErrProgramaNoEncontrado
		}
		return nil, err
	}
	if p.Alcance == AlcanceLocal && (p.EmpresaPropietariaID == nil || *p.EmpresaPropietariaID != claims.EmpresaID) {
		return nil, ErrProgramaNoEncontrado
	}

	isGlobalAdmin, err := s.isPlatformLoyaltyAdmin(ctx, claims)
	if err != nil {
		return nil, err
	}
	canManageProg := s.canManagePrograma(claims, isGlobalAdmin, *p)

	adh, err := s.repo.FindAdhesion(ctx, p.ID, claims.EmpresaID)
	if err != nil && !errors.Is(err, errAdhesionNotFound) {
		return nil, err
	}
	isActiveAdhered := adh != nil && adh.Activo
	isOwner := p.EmpresaPropietariaID != nil && *p.EmpresaPropietariaID == claims.EmpresaID

	rows, err := s.repo.ListReglas(ctx, p.ID, claims.EmpresaID, canManageProg)
	if err != nil {
		return nil, err
	}

	out := make([]ReglaResponse, 0, len(rows))
	for _, r := range rows {
		canEdit := s.canManageRegla(claims, canManageProg, r)
		out = append(out, toReglaResponse(r, canEdit, canEdit))
	}

	return &ListarReglasResult{
		Data:              out,
		PuedeCrearGlobal:  canManageProg,
		PuedeCrearEmpresa: isActiveAdhered || isOwner,
	}, nil
}

func (s *Service) CrearRegla(ctx context.Context, claims *auth.Claims, programaID int32, req CrearReglaRequest) (*ReglaResponse, error) {
	p, err := s.repo.FindProgramaByID(ctx, programaID)
	if err != nil {
		if errors.Is(err, errProgramaNotFound) {
			return nil, ErrProgramaNoEncontrado
		}
		return nil, err
	}
	if p.Alcance == AlcanceLocal && (p.EmpresaPropietariaID == nil || *p.EmpresaPropietariaID != claims.EmpresaID) {
		return nil, ErrProgramaNoEncontrado
	}

	isGlobalAdmin, err := s.isPlatformLoyaltyAdmin(ctx, claims)
	if err != nil {
		return nil, err
	}
	canManageProg := s.canManagePrograma(claims, isGlobalAdmin, *p)

	var empresaID *int32
	if req.EsGlobal {
		if !canManageProg {
			if p.Alcance == AlcanceGlobal {
				return nil, ErrSinPermisoPlanGlobal
			}
			return nil, ErrSinPermisoPlanEmpresa
		}
	} else {
		isOwner := p.EmpresaPropietariaID != nil && *p.EmpresaPropietariaID == claims.EmpresaID
		if !isOwner {
			adh, err := s.repo.FindAdhesion(ctx, p.ID, claims.EmpresaID)
			if err != nil && !errors.Is(err, errAdhesionNotFound) {
				return nil, err
			}
			if adh == nil || !adh.Activo {
				return nil, ErrRequiereAdhesion
			}
		}
		empID := claims.EmpresaID
		empresaID = &empID
	}

	nombre := strings.TrimSpace(req.Nombre)
	if nombre == "" {
		return nil, httpx.NewError(http.StatusBadRequest, "validacion", "el nombre de la regla es obligatorio")
	}

	tipoRegla := strings.ToUpper(strings.TrimSpace(req.TipoRegla))
	if tipoRegla == "" {
		tipoRegla = TipoReglaPorMonto
	}

	montoBase := 1000.0
	if tipoRegla != TipoReglaPorMonto {
		montoBase = 1.0
	}
	if req.MontoBase != nil {
		montoBase = *req.MontoBase
	}

	compraMinima := 0.0
	if req.CompraMinima != nil {
		compraMinima = *req.CompraMinima
	}

	monedaID := p.MonedaReferenciaID
	if req.MonedaID != nil {
		okMoneda, err := s.repo.ExistsMoneda(ctx, *req.MonedaID)
		if err != nil {
			return nil, err
		}
		if !okMoneda {
			return nil, ErrMonedaNoEncontrada
		}
		monedaID = *req.MonedaID
	}

	prioridad := int32(100)
	if req.Prioridad != nil {
		prioridad = *req.Prioridad
	}

	fechaInicio := time.Now().UTC()
	if req.FechaInicio != nil {
		fechaInicio = req.FechaInicio.UTC()
	}
	if req.FechaFin != nil && req.FechaFin.UTC().Before(fechaInicio) {
		return nil, ErrVigenciaInvalida
	}

	activo := true
	if req.Activo != nil {
		activo = *req.Activo
	}

	regla := &Regla{
		ProgramaID:     p.ID,
		EmpresaID:      empresaID,
		Nombre:         nombre,
		TipoRegla:      tipoRegla,
		PuntosPorMonto: req.PuntosPorMonto,
		MontoBase:      montoBase,
		CompraMinima:   compraMinima,
		MonedaID:       monedaID,
		Prioridad:      prioridad,
		FechaInicio:    fechaInicio,
		FechaFin:       req.FechaFin,
		Activo:         activo,
	}

	if err := s.repo.CreateRegla(ctx, regla); err != nil {
		return nil, err
	}

	creada, err := s.repo.FindReglaByID(ctx, p.ID, regla.ID)
	if err != nil {
		return nil, err
	}
	resp := toReglaResponse(*creada, true, true)
	return &resp, nil
}

func (s *Service) ActualizarRegla(ctx context.Context, claims *auth.Claims, programaID, reglaID int32, req ActualizarReglaRequest) (*ReglaResponse, error) {
	p, err := s.repo.FindProgramaByID(ctx, programaID)
	if err != nil {
		if errors.Is(err, errProgramaNotFound) {
			return nil, ErrProgramaNoEncontrado
		}
		return nil, err
	}
	r, err := s.repo.FindReglaByID(ctx, programaID, reglaID)
	if err != nil {
		if errors.Is(err, errReglaNotFound) {
			return nil, ErrReglaNoEncontrada
		}
		return nil, err
	}

	isGlobalAdmin, err := s.isPlatformLoyaltyAdmin(ctx, claims)
	if err != nil {
		return nil, err
	}
	canManageProg := s.canManagePrograma(claims, isGlobalAdmin, *p)
	if !s.canManageRegla(claims, canManageProg, *r) {
		return nil, ErrSinPermisoRegla
	}

	updates := map[string]any{
		"updated_at": time.Now().UTC(),
	}
	if req.Nombre != nil {
		nombre := strings.TrimSpace(*req.Nombre)
		if nombre != "" {
			updates["nombre"] = nombre
		}
	}
	if req.TipoRegla != nil {
		updates["tipo_regla"] = strings.ToUpper(strings.TrimSpace(*req.TipoRegla))
	}
	if req.PuntosPorMonto != nil {
		updates["puntos_por_monto"] = *req.PuntosPorMonto
	}
	if req.MontoBase != nil {
		updates["monto_base"] = *req.MontoBase
	}
	if req.CompraMinima != nil {
		updates["compra_minima"] = *req.CompraMinima
	}
	if req.MonedaID != nil {
		okMoneda, err := s.repo.ExistsMoneda(ctx, *req.MonedaID)
		if err != nil {
			return nil, err
		}
		if !okMoneda {
			return nil, ErrMonedaNoEncontrada
		}
		updates["moneda_id"] = *req.MonedaID
	}
	if req.Prioridad != nil {
		updates["prioridad"] = *req.Prioridad
	}
	if req.Activo != nil {
		updates["activo"] = *req.Activo
	}

	fechaInicio := r.FechaInicio
	if req.FechaInicio != nil {
		fechaInicio = req.FechaInicio.UTC()
		updates["fecha_inicio"] = fechaInicio
	}
	if req.LimpiarFechaFin {
		updates["fecha_fin"] = nil
	} else if req.FechaFin != nil {
		ff := req.FechaFin.UTC()
		if ff.Before(fechaInicio) {
			return nil, ErrVigenciaInvalida
		}
		updates["fecha_fin"] = ff
	} else if r.FechaFin != nil && r.FechaFin.Before(fechaInicio) {
		return nil, ErrVigenciaInvalida
	}

	if err := s.repo.UpdateRegla(ctx, programaID, reglaID, updates); err != nil {
		if errors.Is(err, errReglaNotFound) {
			return nil, ErrReglaNoEncontrada
		}
		return nil, err
	}
	actualizada, err := s.repo.FindReglaByID(ctx, programaID, reglaID)
	if err != nil {
		return nil, err
	}
	resp := toReglaResponse(*actualizada, true, true)
	return &resp, nil
}

func (s *Service) EliminarRegla(ctx context.Context, claims *auth.Claims, programaID, reglaID int32) error {
	p, err := s.repo.FindProgramaByID(ctx, programaID)
	if err != nil {
		if errors.Is(err, errProgramaNotFound) {
			return ErrProgramaNoEncontrado
		}
		return err
	}
	r, err := s.repo.FindReglaByID(ctx, programaID, reglaID)
	if err != nil {
		if errors.Is(err, errReglaNotFound) {
			return ErrReglaNoEncontrada
		}
		return err
	}

	isGlobalAdmin, err := s.isPlatformLoyaltyAdmin(ctx, claims)
	if err != nil {
		return err
	}
	canManageProg := s.canManagePrograma(claims, isGlobalAdmin, *p)
	if !s.canManageRegla(claims, canManageProg, *r) {
		return ErrSinPermisoRegla
	}

	if err := s.repo.DeleteRegla(ctx, programaID, reglaID); err != nil {
		if errors.Is(err, errReglaNotFound) {
			return ErrReglaNoEncontrada
		}
		return err
	}
	return nil
}

// =====================================================================
// Handlers
// =====================================================================

func (h *Handler) listarReglas(c *gin.Context) {
	claims, ok := auth.ClaimsFromContext(c)
	if !ok {
		httpx.RespondError(c, auth.ErrNoAutenticado)
		return
	}
	programaID, err := parseID32(c)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	resp, err := h.svc.ListarReglas(c.Request.Context(), claims, programaID)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) crearRegla(c *gin.Context) {
	claims, ok := auth.ClaimsFromContext(c)
	if !ok {
		httpx.RespondError(c, auth.ErrNoAutenticado)
		return
	}
	programaID, err := parseID32(c)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	var req CrearReglaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondError(c, httpx.BindError(err))
		return
	}
	resp, err := h.svc.CrearRegla(c.Request.Context(), claims, programaID, req)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *Handler) actualizarRegla(c *gin.Context) {
	claims, ok := auth.ClaimsFromContext(c)
	if !ok {
		httpx.RespondError(c, auth.ErrNoAutenticado)
		return
	}
	programaID, err := parseID32(c)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	reglaID, err := parseParamID32(c, "reglaId")
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	var req ActualizarReglaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondError(c, httpx.BindError(err))
		return
	}
	resp, err := h.svc.ActualizarRegla(c.Request.Context(), claims, programaID, reglaID, req)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) eliminarRegla(c *gin.Context) {
	claims, ok := auth.ClaimsFromContext(c)
	if !ok {
		httpx.RespondError(c, auth.ErrNoAutenticado)
		return
	}
	programaID, err := parseID32(c)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	reglaID, err := parseParamID32(c, "reglaId")
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	if err := h.svc.EliminarRegla(c.Request.Context(), claims, programaID, reglaID); err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func parseParamID32(c *gin.Context, param string) (int32, error) {
	raw := c.Param(param)
	v, err := strconv.ParseInt(raw, 10, 32)
	if err != nil || v <= 0 {
		return 0, httpx.NewError(http.StatusBadRequest, "id_invalido", "el identificador debe ser un entero positivo")
	}
	return int32(v), nil
}
