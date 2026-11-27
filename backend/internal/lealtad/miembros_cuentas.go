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
	"gorm.io/gorm/clause"

	"tesserak/backend/internal/auth"
	"tesserak/backend/internal/httpx"
)

const (
	EstadoCuentaActiva    = "ACTIVA"
	EstadoCuentaBloqueada = "BLOQUEADA"
	EstadoCuentaCerrada   = "CERRADA"
)

var (
	errMiembroNotFound      = errors.New("miembro de lealtad no encontrado")
	errCuentaNotFound       = errors.New("cuenta de lealtad no encontrada")
	errTerceroNotFound      = errors.New("cliente no encontrado")
	ErrMiembroNoEncontrado  = httpx.NewError(http.StatusNotFound, "miembro_no_encontrado", "el miembro de lealtad no existe")
	ErrCuentaNoEncontrada   = httpx.NewError(http.StatusNotFound, "cuenta_no_encontrada", "la cuenta de lealtad no existe o no pertenece a un plan adherido")
	ErrCuentaDuplicada      = httpx.NewError(http.StatusConflict, "cuenta_duplicada", "el miembro ya tiene una cuenta registrada en este plan de lealtad")
	ErrPaisNoEncontrado     = httpx.NewError(http.StatusBadRequest, "pais_invalido", "el país indicado no existe")
	ErrEmpresaNoAdherida    = httpx.NewError(http.StatusForbidden, "empresa_no_adherida", "tu empresa no está adherida activamente a este plan de lealtad")
	ErrTerceroNoEncontrado  = httpx.NewError(http.StatusNotFound, "tercero_no_encontrado", "el cliente indicado no existe en tu empresa")
)

type PaisRef struct {
	ID        int32  `gorm:"column:id;primaryKey"`
	Nombre    string `gorm:"column:nombre"`
	ISOAlpha2 string `gorm:"column:iso_alpha_2"`
}

func (PaisRef) TableName() string { return "paises" }

type TerceroClienteRef struct {
	ID                   int64   `gorm:"column:id;primaryKey"`
	EmpresaID            int32   `gorm:"column:empresa_id"`
	PaisID               *int32  `gorm:"column:pais_id"`
	TipoIdentificacion   string  `gorm:"column:tipo_identificacion"`
	NumeroIdentificacion string  `gorm:"column:numero_identificacion"`
	RazonSocial          string  `gorm:"column:razon_social"`
	Email                *string `gorm:"column:email"`
	Telefono             *string `gorm:"column:telefono"`
}

func (TerceroClienteRef) TableName() string { return "tercero_maestro" }

type Miembro struct {
	ID                   int64     `gorm:"column:id;primaryKey;autoIncrement"`
	PaisID               int32     `gorm:"column:pais_id;not null"`
	TipoIdentificacion   string    `gorm:"column:tipo_identificacion;not null"`
	NumeroIdentificacion string    `gorm:"column:numero_identificacion;not null"`
	NombreCompleto       string    `gorm:"column:nombre_completo;not null"`
	Email                *string   `gorm:"column:email"`
	Telefono             *string   `gorm:"column:telefono"`
	CreatedAt            time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt            time.Time `gorm:"column:updated_at;autoUpdateTime"`

	Pais *PaisRef `gorm:"foreignKey:PaisID;references:ID"`
}

func (Miembro) TableName() string { return "lealtad_miembros" }

type Cuenta struct {
	ID          int64     `gorm:"column:id;primaryKey;autoIncrement"`
	ProgramaID  int32     `gorm:"column:programa_id;not null"`
	MiembroID   int64     `gorm:"column:miembro_id;not null"`
	SaldoPuntos int64     `gorm:"column:saldo_puntos;not null"`
	Estado      string    `gorm:"column:estado;not null"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt   time.Time `gorm:"column:updated_at;autoUpdateTime"`

	Programa *Programa `gorm:"foreignKey:ProgramaID;references:ID"`
	Miembro  *Miembro  `gorm:"foreignKey:MiembroID;references:ID"`
}

func (Cuenta) TableName() string { return "lealtad_cuentas" }

// =====================================================================
// DTOs
// =====================================================================

type PaisResponse struct {
	ID        int32  `json:"id"`
	Nombre    string `json:"nombre"`
	ISOAlpha2 string `json:"iso_alpha_2"`
}

type ProgramaOperacionResponse struct {
	ID                  int32   `json:"id"`
	Codigo              string  `json:"codigo"`
	Nombre              string  `json:"nombre"`
	Alcance             string  `json:"alcance"`
	Estado              string  `json:"estado"`
	ValorPuntoRedencion float64 `json:"valor_punto_redencion"`
	MonedaCodigo        string  `json:"moneda_codigo,omitempty"`
	MonedaSimbolo       string  `json:"moneda_simbolo,omitempty"`
	EsProgramaPrincipal bool    `json:"es_programa_principal"`
	PermiteAcumular     bool    `json:"permite_acumular"`
	PermiteRedimir      bool    `json:"permite_redimir"`
}

type ClienteDisponibleResponse struct {
	ID                   int64   `json:"id"`
	PaisID               *int32  `json:"pais_id,omitempty"`
	TipoIdentificacion   string  `json:"tipo_identificacion"`
	NumeroIdentificacion string  `json:"numero_identificacion"`
	RazonSocial          string  `json:"razon_social"`
	Email                *string `json:"email,omitempty"`
	Telefono             *string `json:"telefono,omitempty"`
}

type MiembroResponse struct {
	ID                   int64     `json:"id"`
	PaisID               int32     `json:"pais_id"`
	PaisNombre           string    `json:"pais_nombre,omitempty"`
	PaisISO              string    `json:"pais_iso,omitempty"`
	TipoIdentificacion   string    `json:"tipo_identificacion"`
	NumeroIdentificacion string    `json:"numero_identificacion"`
	NombreCompleto       string    `json:"nombre_completo"`
	Email                *string   `json:"email,omitempty"`
	Telefono             *string   `json:"telefono,omitempty"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

type CuentaResponse struct {
	ID                   int64     `json:"id"`
	ProgramaID           int32     `json:"programa_id"`
	ProgramaCodigo       string    `json:"programa_codigo,omitempty"`
	ProgramaNombre       string    `json:"programa_nombre,omitempty"`
	ProgramaAlcance      string    `json:"programa_alcance,omitempty"`
	ValorPuntoRedencion  float64   `json:"valor_punto_redencion,omitempty"`
	MonedaCodigo         string    `json:"moneda_codigo,omitempty"`
	MonedaSimbolo        string    `json:"moneda_simbolo,omitempty"`
	ValorEstimadoSaldo   float64   `json:"valor_estimado_saldo"`
	MiembroID            int64     `json:"miembro_id"`
	MiembroNombre        string    `json:"miembro_nombre,omitempty"`
	MiembroTipoDoc       string    `json:"miembro_tipo_doc,omitempty"`
	MiembroNumDoc        string    `json:"miembro_num_doc,omitempty"`
	MiembroPaisISO       string    `json:"miembro_pais_iso,omitempty"`
	MiembroEmail         *string   `json:"miembro_email,omitempty"`
	MiembroTelefono      *string   `json:"miembro_telefono,omitempty"`
	SaldoPuntos          int64     `json:"saldo_puntos"`
	Estado               string    `json:"estado"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

type RegistrarMiembroRequest struct {
	PaisID               int32   `json:"pais_id"               binding:"required,gt=0"`
	TipoIdentificacion   string  `json:"tipo_identificacion"   binding:"required,oneof=NIT CC CE PASAPORTE RUT DNI OTRO"`
	NumeroIdentificacion string  `json:"numero_identificacion" binding:"required,min=2,max=50"`
	NombreCompleto       string  `json:"nombre_completo"       binding:"required,min=2,max=200"`
	Email                *string `json:"email"                 binding:"omitnil,email,max=255"`
	Telefono             *string `json:"telefono"              binding:"omitnil,max=50"`
	ProgramaID           int32   `json:"programa_id"           binding:"required,gt=0"`
}

type InscribirDesdeTerceroRequest struct {
	TerceroID  int64  `json:"tercero_id"  binding:"required,gt=0"`
	ProgramaID int32  `json:"programa_id" binding:"required,gt=0"`
	PaisID     *int32 `json:"pais_id"     binding:"omitnil,gt=0"`
}

type ActualizarMiembroRequest struct {
	NombreCompleto *string `json:"nombre_completo" binding:"omitnil,min=2,max=200"`
	Email          *string `json:"email"           binding:"omitnil,max=255"`
	Telefono       *string `json:"telefono"        binding:"omitnil,max=50"`
}

type AbrirCuentaRequest struct {
	ProgramaID int32 `json:"programa_id" binding:"required,gt=0"`
	MiembroID  int64 `json:"miembro_id"  binding:"required,gt=0"`
}

type ActualizarEstadoCuentaRequest struct {
	Estado string `json:"estado" binding:"required,oneof=ACTIVA BLOQUEADA CERRADA"`
}

type CuentaFilter struct {
	ProgramaID int32
	Estado     string
	Search     string
}

func toMiembroResponse(m Miembro) MiembroResponse {
	resp := MiembroResponse{
		ID:                   m.ID,
		PaisID:               m.PaisID,
		TipoIdentificacion:   m.TipoIdentificacion,
		NumeroIdentificacion: m.NumeroIdentificacion,
		NombreCompleto:       m.NombreCompleto,
		Email:                m.Email,
		Telefono:             m.Telefono,
		CreatedAt:            m.CreatedAt,
		UpdatedAt:            m.UpdatedAt,
	}
	if m.Pais != nil {
		resp.PaisNombre = m.Pais.Nombre
		resp.PaisISO = m.Pais.ISOAlpha2
	}
	return resp
}

func toCuentaResponse(c Cuenta) CuentaResponse {
	resp := CuentaResponse{
		ID:          c.ID,
		ProgramaID:  c.ProgramaID,
		MiembroID:   c.MiembroID,
		SaldoPuntos: c.SaldoPuntos,
		Estado:      c.Estado,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}
	if c.Programa != nil {
		resp.ProgramaCodigo = c.Programa.Codigo
		resp.ProgramaNombre = c.Programa.Nombre
		resp.ProgramaAlcance = c.Programa.Alcance
		resp.ValorPuntoRedencion = c.Programa.ValorPuntoRedencion
		resp.ValorEstimadoSaldo = float64(c.SaldoPuntos) * c.Programa.ValorPuntoRedencion
		if c.Programa.Moneda != nil {
			resp.MonedaCodigo = c.Programa.Moneda.CodigoISO
			resp.MonedaSimbolo = c.Programa.Moneda.Simbolo
		}
	}
	if c.Miembro != nil {
		resp.MiembroNombre = c.Miembro.NombreCompleto
		resp.MiembroTipoDoc = c.Miembro.TipoIdentificacion
		resp.MiembroNumDoc = c.Miembro.NumeroIdentificacion
		resp.MiembroEmail = c.Miembro.Email
		resp.MiembroTelefono = c.Miembro.Telefono
		if c.Miembro.Pais != nil {
			resp.MiembroPaisISO = c.Miembro.Pais.ISOAlpha2
		}
	}
	return resp
}

// =====================================================================
// Repository
// =====================================================================

func (r *Repository) ListPaises(ctx context.Context) ([]PaisRef, error) {
	var items []PaisRef
	err := r.db.WithContext(ctx).
		Order("CASE WHEN iso_alpha_2 = 'CO' THEN 0 ELSE 1 END, nombre ASC").
		Find(&items).Error
	return items, err
}

func (r *Repository) ExistsPais(ctx context.Context, paisID int32) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&PaisRef{}).Where("id = ?", paisID).Count(&count).Error
	return count > 0, err
}

func (r *Repository) GetDefaultPaisID(ctx context.Context) (int32, error) {
	var p PaisRef
	err := r.db.WithContext(ctx).
		Order("CASE WHEN iso_alpha_2 = 'CO' THEN 0 ELSE 1 END, id ASC").
		First(&p).Error
	return p.ID, err
}

func (r *Repository) ListProgramasAdheridosEmpresa(ctx context.Context, empresaID int32) ([]ProgramaOperacionResponse, error) {
	type row struct {
		ID                  int32
		Codigo              string
		Nombre              string
		Alcance             string
		Estado              string
		ValorPuntoRedencion float64
		MonedaCodigo        string
		MonedaSimbolo       string
		EsProgramaPrincipal bool
		PermiteAcumular     bool
		PermiteRedimir      bool
	}
	var rows []row
	err := r.db.WithContext(ctx).
		Table("lealtad_programa_empresas lpe").
		Select(`
			lp.id, lp.codigo, lp.nombre, lp.alcance, lp.estado, lp.valor_punto_redencion,
			m.codigo_iso AS moneda_codigo, m.simbolo AS moneda_simbolo,
			lpe.es_programa_principal, lpe.permite_acumular, lpe.permite_redimir
		`).
		Joins("JOIN lealtad_programas lp ON lp.id = lpe.programa_id AND lp.deleted_at IS NULL").
		Joins("JOIN monedas m ON m.id = lp.moneda_referencia_id").
		Where("lpe.empresa_id = ? AND lpe.activo = TRUE", empresaID).
		Order("lpe.es_programa_principal DESC, lp.es_sistema DESC, lp.nombre ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]ProgramaOperacionResponse, 0, len(rows))
	for _, x := range rows {
		out = append(out, ProgramaOperacionResponse{
			ID:                  x.ID,
			Codigo:              x.Codigo,
			Nombre:              x.Nombre,
			Alcance:             x.Alcance,
			Estado:              x.Estado,
			ValorPuntoRedencion: x.ValorPuntoRedencion,
			MonedaCodigo:        x.MonedaCodigo,
			MonedaSimbolo:       x.MonedaSimbolo,
			EsProgramaPrincipal: x.EsProgramaPrincipal,
			PermiteAcumular:     x.PermiteAcumular,
			PermiteRedimir:      x.PermiteRedimir,
		})
	}
	return out, nil
}

func (r *Repository) IsEmpresaAdheridaActiva(ctx context.Context, empresaID, programaID int32) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&ProgramaEmpresa{}).
		Where("empresa_id = ? AND programa_id = ? AND activo = TRUE", empresaID, programaID).
		Count(&count).Error
	return count > 0, err
}

func (r *Repository) ListClientesDisponibles(ctx context.Context, empresaID int32, search string) ([]TerceroClienteRef, error) {
	q := r.db.WithContext(ctx).
		Table("tercero_maestro tm").
		Select("tm.id, tm.empresa_id, tm.pais_id, tm.tipo_identificacion, tm.numero_identificacion, tm.razon_social, tm.email, tm.telefono").
		Joins("JOIN tercero_empresa_roles ter ON ter.tercero_id = tm.id").
		Joins("JOIN roles_tercero rt ON rt.id = ter.rol_id AND upper(rt.codigo) = 'CLIENTE' AND rt.deleted_at IS NULL").
		Where("tm.empresa_id = ? AND tm.activo = TRUE AND tm.deleted_at IS NULL", empresaID)

	if s := strings.TrimSpace(search); s != "" {
		like := "%" + strings.ToLower(s) + "%"
		q = q.Where(
			"lower(tm.razon_social) LIKE ? OR lower(tm.numero_identificacion) LIKE ? OR lower(COALESCE(tm.email, '')) LIKE ?",
			like, like, like,
		)
	}

	var items []TerceroClienteRef
	err := q.Order("tm.razon_social ASC").Limit(50).Scan(&items).Error
	return items, err
}

func (r *Repository) FindTerceroByID(ctx context.Context, empresaID int32, terceroID int64) (*TerceroClienteRef, error) {
	var t TerceroClienteRef
	err := r.db.WithContext(ctx).
		Where("id = ? AND empresa_id = ? AND deleted_at IS NULL", terceroID, empresaID).
		First(&t).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errTerceroNotFound
		}
		return nil, err
	}
	return &t, nil
}

func (r *Repository) FindMiembroByDocumento(ctx context.Context, paisID int32, tipoDoc, numDoc string) (*Miembro, error) {
	var m Miembro
	err := r.db.WithContext(ctx).
		Preload("Pais").
		Where("pais_id = ? AND upper(tipo_identificacion) = ? AND lower(trim(numero_identificacion)) = ?",
			paisID, strings.ToUpper(strings.TrimSpace(tipoDoc)), strings.ToLower(strings.TrimSpace(numDoc))).
		First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errMiembroNotFound
		}
		return nil, err
	}
	return &m, nil
}

func (r *Repository) FindMiembroByID(ctx context.Context, id int64) (*Miembro, error) {
	var m Miembro
	err := r.db.WithContext(ctx).Preload("Pais").Where("id = ?", id).First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errMiembroNotFound
		}
		return nil, err
	}
	return &m, nil
}

func (r *Repository) UpsertMiembroYAbrirCuenta(
	ctx context.Context,
	paisID int32,
	tipoDoc, numDoc, nombre string,
	email, telefono *string,
	programaID int32,
) (*Cuenta, error) {
	var cuenta Cuenta
	now := time.Now().UTC()

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var miembro Miembro
		err := tx.Where(
			"pais_id = ? AND upper(tipo_identificacion) = ? AND lower(trim(numero_identificacion)) = ?",
			paisID, strings.ToUpper(strings.TrimSpace(tipoDoc)), strings.ToLower(strings.TrimSpace(numDoc)),
		).First(&miembro).Error

		if errors.Is(err, gorm.ErrRecordNotFound) {
			miembro = Miembro{
				PaisID:               paisID,
				TipoIdentificacion:   strings.ToUpper(strings.TrimSpace(tipoDoc)),
				NumeroIdentificacion: strings.TrimSpace(numDoc),
				NombreCompleto:       strings.TrimSpace(nombre),
				Email:                email,
				Telefono:             telefono,
				CreatedAt:            now,
				UpdatedAt:            now,
			}
			if err := tx.Create(&miembro).Error; err != nil {
				return err
			}
		} else if err != nil {
			return err
		}

		// Crear cuenta si no existe (idempotente para coalición: si ya tenía cuenta en Lealtad Tesserak, la reutiliza)
		nueva := Cuenta{
			ProgramaID:  programaID,
			MiembroID:   miembro.ID,
			SaldoPuntos: 0,
			Estado:      EstadoCuentaActiva,
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		if err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "programa_id"}, {Name: "miembro_id"}},
			DoNothing: true,
		}).Create(&nueva).Error; err != nil {
			return err
		}

		return tx.
			Preload("Programa").
			Preload("Programa.Moneda").
			Preload("Miembro").
			Preload("Miembro.Pais").
			Where("programa_id = ? AND miembro_id = ?", programaID, miembro.ID).
			First(&cuenta).Error
	})
	if err != nil {
		return nil, err
	}
	return &cuenta, nil
}

func (r *Repository) UpdateMiembro(ctx context.Context, id int64, updates map[string]any) error {
	res := r.db.WithContext(ctx).Model(&Miembro{}).Where("id = ?", id).Updates(updates)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errMiembroNotFound
	}
	return nil
}

func (r *Repository) ListCuentasForEmpresa(ctx context.Context, empresaID int32, f CuentaFilter) ([]Cuenta, error) {
	q := r.db.WithContext(ctx).
		Model(&Cuenta{}).
		Preload("Programa").
		Preload("Programa.Moneda").
		Preload("Miembro").
		Preload("Miembro.Pais").
		Joins("JOIN lealtad_programa_empresas lpe ON lpe.programa_id = lealtad_cuentas.programa_id AND lpe.empresa_id = ? AND lpe.activo = TRUE", empresaID).
		Joins("JOIN lealtad_miembros lm ON lm.id = lealtad_cuentas.miembro_id")

	if f.ProgramaID > 0 {
		q = q.Where("lealtad_cuentas.programa_id = ?", f.ProgramaID)
	}
	if e := strings.ToUpper(strings.TrimSpace(f.Estado)); e != "" {
		q = q.Where("lealtad_cuentas.estado = ?", e)
	}
	if s := strings.TrimSpace(f.Search); s != "" {
		like := "%" + strings.ToLower(s) + "%"
		q = q.Where(
			"lower(lm.nombre_completo) LIKE ? OR lower(lm.numero_identificacion) LIKE ? OR lower(COALESCE(lm.email, '')) LIKE ?",
			like, like, like,
		)
	}

	var items []Cuenta
	err := q.Order("lealtad_cuentas.updated_at DESC, lealtad_cuentas.id DESC").Find(&items).Error
	return items, err
}

func (r *Repository) FindCuentaForEmpresa(ctx context.Context, empresaID int32, cuentaID int64) (*Cuenta, error) {
	var c Cuenta
	err := r.db.WithContext(ctx).
		Preload("Programa").
		Preload("Programa.Moneda").
		Preload("Miembro").
		Preload("Miembro.Pais").
		Joins("JOIN lealtad_programa_empresas lpe ON lpe.programa_id = lealtad_cuentas.programa_id AND lpe.empresa_id = ? AND lpe.activo = TRUE", empresaID).
		Where("lealtad_cuentas.id = ?", cuentaID).
		First(&c).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errCuentaNotFound
		}
		return nil, err
	}
	return &c, nil
}

func (r *Repository) CreateCuenta(ctx context.Context, programaID int32, miembroID int64) (*Cuenta, error) {
	now := time.Now().UTC()
	c := Cuenta{
		ProgramaID:  programaID,
		MiembroID:   miembroID,
		SaldoPuntos: 0,
		Estado:      EstadoCuentaActiva,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := r.db.WithContext(ctx).Create(&c).Error; err != nil {
		if isUniqueViolation(err) {
			return nil, ErrCuentaDuplicada
		}
		return nil, err
	}
	var out Cuenta
	err := r.db.WithContext(ctx).
		Preload("Programa").
		Preload("Programa.Moneda").
		Preload("Miembro").
		Preload("Miembro.Pais").
		Where("id = ?", c.ID).
		First(&out).Error
	return &out, err
}

func (r *Repository) UpdateCuentaEstado(ctx context.Context, cuentaID int64, estado string) error {
	res := r.db.WithContext(ctx).
		Model(&Cuenta{}).
		Where("id = ?", cuentaID).
		Updates(map[string]any{
			"estado":     estado,
			"updated_at": time.Now().UTC(),
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errCuentaNotFound
	}
	return nil
}

// =====================================================================
// Service
// =====================================================================

func (s *Service) ListarPaises(ctx context.Context) ([]PaisResponse, error) {
	items, err := s.repo.ListPaises(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]PaisResponse, 0, len(items))
	for _, p := range items {
		out = append(out, PaisResponse{ID: p.ID, Nombre: p.Nombre, ISOAlpha2: p.ISOAlpha2})
	}
	return out, nil
}

func (s *Service) ListarProgramasOperacion(ctx context.Context, claims *auth.Claims) ([]ProgramaOperacionResponse, error) {
	return s.repo.ListProgramasAdheridosEmpresa(ctx, claims.EmpresaID)
}

func (s *Service) ListarClientesDisponibles(ctx context.Context, claims *auth.Claims, search string) ([]ClienteDisponibleResponse, error) {
	items, err := s.repo.ListClientesDisponibles(ctx, claims.EmpresaID, search)
	if err != nil {
		return nil, err
	}
	out := make([]ClienteDisponibleResponse, 0, len(items))
	for _, t := range items {
		out = append(out, ClienteDisponibleResponse{
			ID:                   t.ID,
			PaisID:               t.PaisID,
			TipoIdentificacion:   t.TipoIdentificacion,
			NumeroIdentificacion: t.NumeroIdentificacion,
			RazonSocial:          t.RazonSocial,
			Email:                t.Email,
			Telefono:             t.Telefono,
		})
	}
	return out, nil
}

func (s *Service) BuscarMiembroPorDocumento(ctx context.Context, paisID int32, tipoDoc, numDoc string) (*MiembroResponse, error) {
	m, err := s.repo.FindMiembroByDocumento(ctx, paisID, tipoDoc, numDoc)
	if err != nil {
		if errors.Is(err, errMiembroNotFound) {
			return nil, ErrMiembroNoEncontrado
		}
		return nil, err
	}
	resp := toMiembroResponse(*m)
	return &resp, nil
}

func (s *Service) RegistrarMiembroYAbrirCuenta(ctx context.Context, claims *auth.Claims, req RegistrarMiembroRequest) (*CuentaResponse, error) {
	p, err := s.repo.FindProgramaByID(ctx, req.ProgramaID)
	if err != nil {
		if errors.Is(err, errProgramaNotFound) {
			return nil, ErrProgramaNoEncontrado
		}
		return nil, err
	}
	if p.Estado != EstadoActivo {
		return nil, ErrPlanNoDisponible
	}
	adherida, err := s.repo.IsEmpresaAdheridaActiva(ctx, claims.EmpresaID, p.ID)
	if err != nil {
		return nil, err
	}
	if !adherida {
		return nil, ErrEmpresaNoAdherida
	}

	okPais, err := s.repo.ExistsPais(ctx, req.PaisID)
	if err != nil {
		return nil, err
	}
	if !okPais {
		return nil, ErrPaisNoEncontrado
	}

	var email, telefono *string
	if req.Email != nil && strings.TrimSpace(*req.Email) != "" {
		v := strings.TrimSpace(*req.Email)
		email = &v
	}
	if req.Telefono != nil && strings.TrimSpace(*req.Telefono) != "" {
		v := strings.TrimSpace(*req.Telefono)
		telefono = &v
	}

	cuenta, err := s.repo.UpsertMiembroYAbrirCuenta(
		ctx,
		req.PaisID,
		req.TipoIdentificacion,
		req.NumeroIdentificacion,
		req.NombreCompleto,
		email,
		telefono,
		p.ID,
	)
	if err != nil {
		return nil, err
	}
	resp := toCuentaResponse(*cuenta)
	return &resp, nil
}

func (s *Service) InscribirDesdeTercero(ctx context.Context, claims *auth.Claims, req InscribirDesdeTerceroRequest) (*CuentaResponse, error) {
	p, err := s.repo.FindProgramaByID(ctx, req.ProgramaID)
	if err != nil {
		if errors.Is(err, errProgramaNotFound) {
			return nil, ErrProgramaNoEncontrado
		}
		return nil, err
	}
	if p.Estado != EstadoActivo {
		return nil, ErrPlanNoDisponible
	}
	adherida, err := s.repo.IsEmpresaAdheridaActiva(ctx, claims.EmpresaID, p.ID)
	if err != nil {
		return nil, err
	}
	if !adherida {
		return nil, ErrEmpresaNoAdherida
	}

	tercero, err := s.repo.FindTerceroByID(ctx, claims.EmpresaID, req.TerceroID)
	if err != nil {
		if errors.Is(err, errTerceroNotFound) {
			return nil, ErrTerceroNoEncontrado
		}
		return nil, err
	}

	var paisID int32
	if req.PaisID != nil && *req.PaisID > 0 {
		paisID = *req.PaisID
	} else if tercero.PaisID != nil && *tercero.PaisID > 0 {
		paisID = *tercero.PaisID
	} else {
		paisID, err = s.repo.GetDefaultPaisID(ctx)
		if err != nil {
			return nil, ErrPaisNoEncontrado
		}
	}

	cuenta, err := s.repo.UpsertMiembroYAbrirCuenta(
		ctx,
		paisID,
		tercero.TipoIdentificacion,
		tercero.NumeroIdentificacion,
		tercero.RazonSocial,
		tercero.Email,
		tercero.Telefono,
		p.ID,
	)
	if err != nil {
		return nil, err
	}
	resp := toCuentaResponse(*cuenta)
	return &resp, nil
}

func (s *Service) ActualizarMiembro(ctx context.Context, id int64, req ActualizarMiembroRequest) (*MiembroResponse, error) {
	updates := map[string]any{
		"updated_at": time.Now().UTC(),
	}
	if req.NombreCompleto != nil {
		if v := strings.TrimSpace(*req.NombreCompleto); v != "" {
			updates["nombre_completo"] = v
		}
	}
	if req.Email != nil {
		v := strings.TrimSpace(*req.Email)
		if v == "" {
			updates["email"] = nil
		} else {
			updates["email"] = v
		}
	}
	if req.Telefono != nil {
		v := strings.TrimSpace(*req.Telefono)
		if v == "" {
			updates["telefono"] = nil
		} else {
			updates["telefono"] = v
		}
	}
	if err := s.repo.UpdateMiembro(ctx, id, updates); err != nil {
		if errors.Is(err, errMiembroNotFound) {
			return nil, ErrMiembroNoEncontrado
		}
		return nil, err
	}
	m, err := s.repo.FindMiembroByID(ctx, id)
	if err != nil {
		return nil, err
	}
	resp := toMiembroResponse(*m)
	return &resp, nil
}

func (s *Service) ListarCuentas(ctx context.Context, claims *auth.Claims, f CuentaFilter) ([]CuentaResponse, error) {
	rows, err := s.repo.ListCuentasForEmpresa(ctx, claims.EmpresaID, f)
	if err != nil {
		return nil, err
	}
	out := make([]CuentaResponse, 0, len(rows))
	for _, row := range rows {
		out = append(out, toCuentaResponse(row))
	}
	return out, nil
}

func (s *Service) AbrirCuenta(ctx context.Context, claims *auth.Claims, req AbrirCuentaRequest) (*CuentaResponse, error) {
	p, err := s.repo.FindProgramaByID(ctx, req.ProgramaID)
	if err != nil {
		if errors.Is(err, errProgramaNotFound) {
			return nil, ErrProgramaNoEncontrado
		}
		return nil, err
	}
	if p.Estado != EstadoActivo {
		return nil, ErrPlanNoDisponible
	}
	adherida, err := s.repo.IsEmpresaAdheridaActiva(ctx, claims.EmpresaID, p.ID)
	if err != nil {
		return nil, err
	}
	if !adherida {
		return nil, ErrEmpresaNoAdherida
	}
	if _, err := s.repo.FindMiembroByID(ctx, req.MiembroID); err != nil {
		if errors.Is(err, errMiembroNotFound) {
			return nil, ErrMiembroNoEncontrado
		}
		return nil, err
	}
	c, err := s.repo.CreateCuenta(ctx, p.ID, req.MiembroID)
	if err != nil {
		return nil, err
	}
	resp := toCuentaResponse(*c)
	return &resp, nil
}

func (s *Service) ActualizarEstadoCuenta(ctx context.Context, claims *auth.Claims, cuentaID int64, req ActualizarEstadoCuentaRequest) (*CuentaResponse, error) {
	if _, err := s.repo.FindCuentaForEmpresa(ctx, claims.EmpresaID, cuentaID); err != nil {
		if errors.Is(err, errCuentaNotFound) {
			return nil, ErrCuentaNoEncontrada
		}
		return nil, err
	}
	if err := s.repo.UpdateCuentaEstado(ctx, cuentaID, strings.ToUpper(strings.TrimSpace(req.Estado))); err != nil {
		if errors.Is(err, errCuentaNotFound) {
			return nil, ErrCuentaNoEncontrada
		}
		return nil, err
	}
	updated, err := s.repo.FindCuentaForEmpresa(ctx, claims.EmpresaID, cuentaID)
	if err != nil {
		return nil, err
	}
	resp := toCuentaResponse(*updated)
	return &resp, nil
}

// =====================================================================
// Handlers (Operación)
// =====================================================================

func (h *Handler) listarPaises(c *gin.Context) {
	resp, err := h.svc.ListarPaises(c.Request.Context())
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}

func (h *Handler) listarProgramasOperacion(c *gin.Context) {
	claims, ok := auth.ClaimsFromContext(c)
	if !ok {
		httpx.RespondError(c, auth.ErrNoAutenticado)
		return
	}
	resp, err := h.svc.ListarProgramasOperacion(c.Request.Context(), claims)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}

func (h *Handler) listarClientesDisponibles(c *gin.Context) {
	claims, ok := auth.ClaimsFromContext(c)
	if !ok {
		httpx.RespondError(c, auth.ErrNoAutenticado)
		return
	}
	resp, err := h.svc.ListarClientesDisponibles(c.Request.Context(), claims, c.Query("search"))
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}

func (h *Handler) buscarMiembroPorDocumento(c *gin.Context) {
	paisIDRaw := c.Query("pais_id")
	paisID, err := strconv.ParseInt(paisIDRaw, 10, 32)
	if err != nil || paisID <= 0 {
		httpx.RespondError(c, httpx.NewError(http.StatusBadRequest, "validacion", "pais_id es obligatorio"))
		return
	}
	tipoDoc := strings.TrimSpace(c.Query("tipo_identificacion"))
	numDoc := strings.TrimSpace(c.Query("numero_identificacion"))
	if tipoDoc == "" || numDoc == "" {
		httpx.RespondError(c, httpx.NewError(http.StatusBadRequest, "validacion", "tipo_identificacion y numero_identificacion son obligatorios"))
		return
	}
	resp, err := h.svc.BuscarMiembroPorDocumento(c.Request.Context(), int32(paisID), tipoDoc, numDoc)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) registrarMiembro(c *gin.Context) {
	claims, ok := auth.ClaimsFromContext(c)
	if !ok {
		httpx.RespondError(c, auth.ErrNoAutenticado)
		return
	}
	var req RegistrarMiembroRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondError(c, httpx.BindError(err))
		return
	}
	resp, err := h.svc.RegistrarMiembroYAbrirCuenta(c.Request.Context(), claims, req)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *Handler) inscribirDesdeTercero(c *gin.Context) {
	claims, ok := auth.ClaimsFromContext(c)
	if !ok {
		httpx.RespondError(c, auth.ErrNoAutenticado)
		return
	}
	var req InscribirDesdeTerceroRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondError(c, httpx.BindError(err))
		return
	}
	resp, err := h.svc.InscribirDesdeTercero(c.Request.Context(), claims, req)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *Handler) actualizarMiembro(c *gin.Context) {
	id, err := parseID64Param(c, "id")
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	var req ActualizarMiembroRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondError(c, httpx.BindError(err))
		return
	}
	resp, err := h.svc.ActualizarMiembro(c.Request.Context(), id, req)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) listarCuentas(c *gin.Context) {
	claims, ok := auth.ClaimsFromContext(c)
	if !ok {
		httpx.RespondError(c, auth.ErrNoAutenticado)
		return
	}
	var progID int32
	if raw := strings.TrimSpace(c.Query("programa_id")); raw != "" {
		if v, err := strconv.ParseInt(raw, 10, 32); err == nil && v > 0 {
			progID = int32(v)
		}
	}
	resp, err := h.svc.ListarCuentas(c.Request.Context(), claims, CuentaFilter{
		ProgramaID: progID,
		Estado:     c.Query("estado"),
		Search:     c.Query("search"),
	})
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}

func (h *Handler) abrirCuenta(c *gin.Context) {
	claims, ok := auth.ClaimsFromContext(c)
	if !ok {
		httpx.RespondError(c, auth.ErrNoAutenticado)
		return
	}
	var req AbrirCuentaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondError(c, httpx.BindError(err))
		return
	}
	resp, err := h.svc.AbrirCuenta(c.Request.Context(), claims, req)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *Handler) actualizarEstadoCuenta(c *gin.Context) {
	claims, ok := auth.ClaimsFromContext(c)
	if !ok {
		httpx.RespondError(c, auth.ErrNoAutenticado)
		return
	}
	id, err := parseID64Param(c, "id")
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	var req ActualizarEstadoCuentaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondError(c, httpx.BindError(err))
		return
	}
	resp, err := h.svc.ActualizarEstadoCuenta(c.Request.Context(), claims, id, req)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func parseID64Param(c *gin.Context, param string) (int64, error) {
	raw := c.Param(param)
	v, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || v <= 0 {
		return 0, httpx.NewError(http.StatusBadRequest, "id_invalido", "el identificador debe ser un entero positivo")
	}
	return v, nil
}
