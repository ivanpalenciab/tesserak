package crm

import (
	"context"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
)

var (
	errTipoCampanaNotFound  = errors.New("tipo de campaña no encontrado")
	errCampanaNotFound      = errors.New("campaña no encontrada")
	errCampanaLeadNotFound  = errors.New("asociación campaña-lead no encontrada")
	errCampanaGastoNotFound = errors.New("gasto de campaña no encontrado")
)

// =====================================================================
// Tipos de campaña
// =====================================================================

func (r *Repository) CreateTipoCampana(ctx context.Context, t *TipoCampana) error {
	return r.db.WithContext(ctx).Create(t).Error
}

func (r *Repository) ListTiposCampana(ctx context.Context, empresaID int32, activo *bool) ([]TipoCampana, error) {
	q := r.db.WithContext(ctx).Where("empresa_id = ?", empresaID)
	if activo != nil {
		q = q.Where("activo = ?", *activo)
	}
	var items []TipoCampana
	err := q.Order("nombre ASC").Find(&items).Error
	return items, err
}

func (r *Repository) FindTipoCampanaByID(ctx context.Context, empresaID int32, id int32) (*TipoCampana, error) {
	var t TipoCampana
	err := r.db.WithContext(ctx).
		Where("empresa_id = ? AND id = ?", empresaID, id).
		First(&t).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errTipoCampanaNotFound
	}
	return &t, err
}

func (r *Repository) UpdateTipoCampana(ctx context.Context, t *TipoCampana, cambios map[string]any) error {
	if len(cambios) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Model(t).Updates(cambios).Error
}

func (r *Repository) SoftDeleteTipoCampana(ctx context.Context, empresaID int32, id int32) error {
	res := r.db.WithContext(ctx).
		Where("empresa_id = ? AND id = ?", empresaID, id).
		Delete(&TipoCampana{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errTipoCampanaNotFound
	}
	return nil
}

// =====================================================================
// Campañas
// =====================================================================

func (r *Repository) CreateCampana(ctx context.Context, c *Campana) error {
	return r.db.WithContext(ctx).Create(c).Error
}

func (r *Repository) FindCampanaByID(ctx context.Context, empresaID int32, id int64) (*Campana, error) {
	var c Campana
	err := r.db.WithContext(ctx).
		Where("empresa_id = ? AND id = ?", empresaID, id).
		First(&c).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errCampanaNotFound
	}
	return &c, err
}

func (r *Repository) ListCampanas(ctx context.Context, empresaID int32, f ListarCampanasQuery, fechaDesde, fechaHasta *time.Time) ([]Campana, int64, error) {
	q := r.db.WithContext(ctx).Model(&Campana{}).Where("empresa_id = ?", empresaID)

	if f.TipoCampanaID > 0 {
		q = q.Where("tipo_campana_id = ?", f.TipoCampanaID)
	}
	if f.OrigenID > 0 {
		q = q.Where("origen_id = ?", f.OrigenID)
	}
	if f.Estado != "" {
		q = q.Where("estado = ?", f.Estado)
	}
	if fechaDesde != nil {
		q = q.Where("fecha_inicio >= ?", *fechaDesde)
	}
	if fechaHasta != nil {
		q = q.Where("fecha_inicio <= ?", *fechaHasta)
	}
	if s := strings.TrimSpace(f.Q); s != "" {
		like := "%" + escapeLike(s) + "%"
		q = q.Where("nombre ILIKE ? OR descripcion ILIKE ? OR objetivo ILIKE ?", like, like, like)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var items []Campana
	err := q.Order("fecha_inicio DESC, id DESC").
		Offset((f.Page - 1) * f.PageSize).
		Limit(f.PageSize).
		Find(&items).Error
	return items, total, err
}

func (r *Repository) UpdateCampana(ctx context.Context, c *Campana, cambios map[string]any) error {
	if len(cambios) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Model(c).Updates(cambios).Error
}

func (r *Repository) CountLeadsCampana(ctx context.Context, empresaID int32, campanaID int64) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).
		Model(&CampanaLead{}).
		Where("empresa_id = ? AND campana_id = ?", empresaID, campanaID).
		Count(&total).Error
	return total, err
}

func (r *Repository) SumGastosPorMonedaCampana(ctx context.Context, empresaID int32, campanaID int64) ([]GastoPorMonedaResumen, error) {
	var res []GastoPorMonedaResumen
	err := r.db.WithContext(ctx).
		Table("crm_campana_gastos").
		Select("moneda_id, COALESCE(SUM(valor), 0)::float8 AS total_gastado, COUNT(*) AS cantidad_gastos").
		Where("empresa_id = ? AND campana_id = ? AND deleted_at IS NULL", empresaID, campanaID).
		Group("moneda_id").
		Order("moneda_id ASC").
		Scan(&res).Error
	if res == nil {
		res = []GastoPorMonedaResumen{}
	}
	return res, err
}

func (r *Repository) FindLeadByIDAndEmpresa(ctx context.Context, empresaID int32, leadID int64) (*Lead, error) {
	var l Lead
	err := r.db.WithContext(ctx).
		Where("empresa_id = ? AND id = ?", empresaID, leadID).
		First(&l).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errLeadNotFound
	}
	return &l, err
}

func (r *Repository) OrigenPerteneceAEmpresa(ctx context.Context, empresaID int32, origenID int32) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Table("crm_origen_lead").
		Where("id = ? AND empresa_id = ? AND deleted_at IS NULL", origenID, empresaID).
		Count(&count).Error
	return count > 0, err
}

func (r *Repository) MonedaExiste(ctx context.Context, monedaID int32) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Table("monedas").
		Where("id = ?", monedaID).
		Count(&count).Error
	return count > 0, err
}

// =====================================================================
// Asociación Campaña - Lead
// =====================================================================

func (r *Repository) CreateCampanaLead(ctx context.Context, cl *CampanaLead) error {
	return r.db.WithContext(ctx).Create(cl).Error
}

func (r *Repository) DeleteCampanaLead(ctx context.Context, empresaID int32, campanaID, leadID int64) error {
	res := r.db.WithContext(ctx).
		Where("empresa_id = ? AND campana_id = ? AND lead_id = ?", empresaID, campanaID, leadID).
		Delete(&CampanaLead{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errCampanaLeadNotFound
	}
	return nil
}

type leadAsociadoRow struct {
	CampanaLead
	LeadNombre        string  `gorm:"column:lead_nombre"`
	LeadApellido      string  `gorm:"column:lead_apellido"`
	LeadEmail         *string `gorm:"column:lead_email"`
	LeadTelefono      *string `gorm:"column:lead_telefono"`
	LeadNombreEmpresa *string `gorm:"column:lead_nombre_empresa"`
	LeadEstadoID      int32   `gorm:"column:lead_estado_id"`
	LeadConvertido    bool    `gorm:"column:lead_convertido"`
}

func (r *Repository) ListLeadsByCampana(ctx context.Context, empresaID int32, campanaID int64, f ListarLeadsCampanaQuery) ([]LeadAsociadoCampanaResponse, int64, error) {
	base := r.db.WithContext(ctx).
		Table("crm_campana_leads cl").
		Joins("JOIN crm_leads l ON l.id = cl.lead_id AND l.empresa_id = cl.empresa_id AND l.deleted_at IS NULL").
		Where("cl.empresa_id = ? AND cl.campana_id = ?", empresaID, campanaID)

	if f.TipoAtribucion != "" {
		base = base.Where("cl.tipo_atribucion = ?", f.TipoAtribucion)
	}

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []leadAsociadoRow
	err := base.Select(`
			cl.*,
			l.nombre AS lead_nombre,
			l.apellido AS lead_apellido,
			l.email AS lead_email,
			l.telefono AS lead_telefono,
			l.nombre_empresa AS lead_nombre_empresa,
			l.estado_id AS lead_estado_id,
			l.convertido AS lead_convertido
		`).
		Order("cl.fecha_asociacion DESC, cl.id DESC").
		Offset((f.Page - 1) * f.PageSize).
		Limit(f.PageSize).
		Scan(&rows).Error
	if err != nil {
		return nil, 0, err
	}

	out := make([]LeadAsociadoCampanaResponse, 0, len(rows))
	for i := range rows {
		out = append(out, LeadAsociadoCampanaResponse{
			CampanaLeadResponse: toCampanaLeadResponse(&rows[i].CampanaLead),
			LeadNombre:          rows[i].LeadNombre,
			LeadApellido:        rows[i].LeadApellido,
			LeadEmail:           rows[i].LeadEmail,
			LeadTelefono:        rows[i].LeadTelefono,
			LeadNombreEmpresa:   rows[i].LeadNombreEmpresa,
			LeadEstadoID:        rows[i].LeadEstadoID,
			LeadConvertido:      rows[i].LeadConvertido,
		})
	}
	return out, total, nil
}

type campanaAsociadaRow struct {
	CampanaLead
	CampanaNombre      string    `gorm:"column:campana_nombre"`
	CampanaEstado      string    `gorm:"column:campana_estado"`
	CampanaTipoID      int32     `gorm:"column:campana_tipo_id"`
	CampanaFechaInicio time.Time `gorm:"column:campana_fecha_inicio"`
}

func (r *Repository) ListCampanasByLead(ctx context.Context, empresaID int32, leadID int64) ([]CampanaAsociadaLeadResponse, error) {
	var rows []campanaAsociadaRow
	err := r.db.WithContext(ctx).
		Table("crm_campana_leads cl").
		Joins("JOIN crm_campanas c ON c.id = cl.campana_id AND c.empresa_id = cl.empresa_id AND c.deleted_at IS NULL").
		Where("cl.empresa_id = ? AND cl.lead_id = ?", empresaID, leadID).
		Select(`
			cl.*,
			c.nombre AS campana_nombre,
			c.estado AS campana_estado,
			c.tipo_campana_id AS campana_tipo_id,
			c.fecha_inicio AS campana_fecha_inicio
		`).
		Order("cl.fecha_asociacion DESC, cl.id DESC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	out := make([]CampanaAsociadaLeadResponse, 0, len(rows))
	for i := range rows {
		out = append(out, CampanaAsociadaLeadResponse{
			CampanaLeadResponse: toCampanaLeadResponse(&rows[i].CampanaLead),
			CampanaNombre:       rows[i].CampanaNombre,
			CampanaEstado:       rows[i].CampanaEstado,
			CampanaTipoID:       rows[i].CampanaTipoID,
			CampanaFechaInicio:  rows[i].CampanaFechaInicio.Format(dateLayout),
		})
	}
	return out, nil
}

// =====================================================================
// Gastos de campaña
// =====================================================================

func (r *Repository) CreateCampanaGasto(ctx context.Context, g *CampanaGasto) error {
	return r.db.WithContext(ctx).Create(g).Error
}

func (r *Repository) FindCampanaGastoByID(ctx context.Context, empresaID int32, id int64) (*CampanaGasto, error) {
	var g CampanaGasto
	err := r.db.WithContext(ctx).
		Where("empresa_id = ? AND id = ?", empresaID, id).
		First(&g).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errCampanaGastoNotFound
	}
	return &g, err
}

func (r *Repository) ListGastosByCampana(ctx context.Context, empresaID int32, campanaID int64, f ListarGastosCampanaQuery, fechaDesde, fechaHasta *time.Time) ([]CampanaGasto, int64, error) {
	q := r.db.WithContext(ctx).
		Model(&CampanaGasto{}).
		Where("empresa_id = ? AND campana_id = ?", empresaID, campanaID)

	if fechaDesde != nil {
		q = q.Where("fecha >= ?", *fechaDesde)
	}
	if fechaHasta != nil {
		q = q.Where("fecha <= ?", *fechaHasta)
	}
	if s := strings.TrimSpace(f.CategoriaGasto); s != "" {
		q = q.Where("categoria_gasto ILIKE ?", s)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var items []CampanaGasto
	err := q.Order("fecha DESC, id DESC").
		Offset((f.Page - 1) * f.PageSize).
		Limit(f.PageSize).
		Find(&items).Error
	return items, total, err
}

func (r *Repository) UpdateCampanaGasto(ctx context.Context, g *CampanaGasto, cambios map[string]any) error {
	if len(cambios) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Model(g).Updates(cambios).Error
}

func (r *Repository) SoftDeleteCampanaGasto(ctx context.Context, empresaID int32, id int64) error {
	res := r.db.WithContext(ctx).
		Where("empresa_id = ? AND id = ?", empresaID, id).
		Delete(&CampanaGasto{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errCampanaGastoNotFound
	}
	return nil
}
