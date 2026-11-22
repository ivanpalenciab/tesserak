package crm

import (
	"context"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
)

var (
	errOrigenNotFound    = errors.New("origen de lead no encontrado")
	errEstadoNotFound    = errors.New("estado de lead no encontrado")
	errMotivoNotFound    = errors.New("motivo de descarte no encontrado")
	errLeadNotFound      = errors.New("lead no encontrado")
	errActividadNotFound = errors.New("actividad no encontrada")
	errInteresNotFound   = errors.New("interés de producto no encontrado")
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) WithTx(ctx context.Context, fn func(tx *Repository) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(&Repository{db: tx})
	})
}

// UsuarioPerteneceAEmpresa verifica que un usuario activo exista y pertenezca a la empresa.
func (r *Repository) UsuarioPerteneceAEmpresa(ctx context.Context, usuarioID int64, empresaID int32) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Table("usuarios").
		Where("id = ? AND empresa_id = ? AND deleted_at IS NULL", usuarioID, empresaID).
		Count(&count).Error
	return count > 0, err
}

// =====================================================================
// Orígenes de lead
// =====================================================================

func (r *Repository) CreateOrigen(ctx context.Context, o *OrigenLead) error {
	return r.db.WithContext(ctx).Create(o).Error
}

func (r *Repository) FindOrigenByID(ctx context.Context, id int32) (*OrigenLead, error) {
	var o OrigenLead
	err := r.db.WithContext(ctx).First(&o, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errOrigenNotFound
	}
	return &o, err
}

func (r *Repository) ListOrigenes(ctx context.Context, q ListarCatalogoQuery) ([]OrigenLead, error) {
	tx := r.db.WithContext(ctx).Model(&OrigenLead{}).Where("empresa_id = ?", q.EmpresaID)
	if q.Activo != nil {
		tx = tx.Where("activo = ?", *q.Activo)
	}
	var items []OrigenLead
	err := tx.Order("nombre ASC").Find(&items).Error
	return items, err
}

func (r *Repository) UpdateOrigen(ctx context.Context, o *OrigenLead, cambios map[string]any) error {
	if len(cambios) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Model(o).Updates(cambios).Error
}

func (r *Repository) SoftDeleteOrigen(ctx context.Context, id int32) error {
	res := r.db.WithContext(ctx).Delete(&OrigenLead{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errOrigenNotFound
	}
	return nil
}

// =====================================================================
// Estados de lead (etapas del pipeline)
// =====================================================================

func (r *Repository) CreateEstado(ctx context.Context, e *EstadoLead) error {
	return r.db.WithContext(ctx).Create(e).Error
}

func (r *Repository) FindEstadoByID(ctx context.Context, id int32) (*EstadoLead, error) {
	var e EstadoLead
	err := r.db.WithContext(ctx).First(&e, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errEstadoNotFound
	}
	return &e, err
}

func (r *Repository) ListEstados(ctx context.Context, empresaID int32) ([]EstadoLead, error) {
	var items []EstadoLead
	err := r.db.WithContext(ctx).
		Where("empresa_id = ?", empresaID).
		Order("orden ASC, id ASC").
		Find(&items).Error
	return items, err
}

func (r *Repository) UpdateEstado(ctx context.Context, e *EstadoLead, cambios map[string]any) error {
	if len(cambios) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Model(e).Updates(cambios).Error
}

func (r *Repository) SoftDeleteEstado(ctx context.Context, id int32) error {
	res := r.db.WithContext(ctx).Delete(&EstadoLead{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errEstadoNotFound
	}
	return nil
}

// =====================================================================
// Motivos de descarte
// =====================================================================

func (r *Repository) CreateMotivo(ctx context.Context, m *MotivoDescarte) error {
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *Repository) FindMotivoByID(ctx context.Context, id int32) (*MotivoDescarte, error) {
	var m MotivoDescarte
	err := r.db.WithContext(ctx).First(&m, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errMotivoNotFound
	}
	return &m, err
}

func (r *Repository) ListMotivos(ctx context.Context, q ListarCatalogoQuery) ([]MotivoDescarte, error) {
	tx := r.db.WithContext(ctx).Model(&MotivoDescarte{}).Where("empresa_id = ?", q.EmpresaID)
	if q.Activo != nil {
		tx = tx.Where("activo = ?", *q.Activo)
	}
	var items []MotivoDescarte
	err := tx.Order("descripcion ASC").Find(&items).Error
	return items, err
}

func (r *Repository) UpdateMotivo(ctx context.Context, m *MotivoDescarte, cambios map[string]any) error {
	if len(cambios) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Model(m).Updates(cambios).Error
}

func (r *Repository) SoftDeleteMotivo(ctx context.Context, id int32) error {
	res := r.db.WithContext(ctx).Delete(&MotivoDescarte{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errMotivoNotFound
	}
	return nil
}

// =====================================================================
// Leads
// =====================================================================

func (r *Repository) CreateLead(ctx context.Context, l *Lead) error {
	return r.db.WithContext(ctx).Create(l).Error
}

func (r *Repository) FindLeadByID(ctx context.Context, id int64) (*Lead, error) {
	var l Lead
	err := r.db.WithContext(ctx).First(&l, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errLeadNotFound
	}
	return &l, err
}

func (r *Repository) ListLeads(ctx context.Context, f ListarLeadsQuery) ([]Lead, int64, error) {
	q := r.db.WithContext(ctx).Model(&Lead{}).Where("empresa_id = ?", f.EmpresaID)

	if f.SucursalID > 0 {
		q = q.Where("sucursal_id = ?", f.SucursalID)
	}
	if f.OrigenID > 0 {
		q = q.Where("origen_id = ?", f.OrigenID)
	}
	if f.EstadoID > 0 {
		q = q.Where("estado_id = ?", f.EstadoID)
	}
	if f.UsuarioAsignadoID > 0 {
		q = q.Where("usuario_asignado_id = ?", f.UsuarioAsignadoID)
	}
	if f.Temperatura != "" {
		q = q.Where("temperatura = ?", f.Temperatura)
	}
	if f.Convertido != nil {
		q = q.Where("convertido = ?", *f.Convertido)
	}
	if s := strings.TrimSpace(f.Q); s != "" {
		like := "%" + escapeLike(s) + "%"
		q = q.Where(
			"nombre ILIKE ? OR apellido ILIKE ? OR nombre_empresa ILIKE ? OR email ILIKE ? OR telefono ILIKE ?",
			like, like, like, like, like,
		)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var items []Lead
	err := q.Order("created_at DESC, id DESC").
		Offset((f.Page - 1) * f.PageSize).
		Limit(f.PageSize).
		Find(&items).Error
	return items, total, err
}

func (r *Repository) UpdateLead(ctx context.Context, l *Lead, cambios map[string]any) error {
	if len(cambios) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Model(l).Updates(cambios).Error
}

func (r *Repository) SoftDeleteLead(ctx context.Context, id int64) error {
	res := r.db.WithContext(ctx).Delete(&Lead{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errLeadNotFound
	}
	return nil
}

// =====================================================================
// Actividades
// =====================================================================

func (r *Repository) CreateActividad(ctx context.Context, a *LeadActividad) error {
	return r.db.WithContext(ctx).Create(a).Error
}

func (r *Repository) FindActividadByID(ctx context.Context, id int64) (*LeadActividad, error) {
	var a LeadActividad
	err := r.db.WithContext(ctx).First(&a, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errActividadNotFound
	}
	return &a, err
}

func (r *Repository) ListActividades(ctx context.Context, leadID *int64, f ListarActividadesQuery) ([]LeadActividad, int64, error) {
	q := r.db.WithContext(ctx).Model(&LeadActividad{})

	if leadID != nil {
		q = q.Where("crm_lead_actividades.lead_id = ?", *leadID)
	}
	if f.EmpresaID > 0 {
		q = q.Joins("JOIN crm_leads ON crm_leads.id = crm_lead_actividades.lead_id AND crm_leads.deleted_at IS NULL").
			Where("crm_leads.empresa_id = ?", f.EmpresaID)
	}
	if f.UsuarioID > 0 {
		q = q.Where("crm_lead_actividades.usuario_id = ?", f.UsuarioID)
	}
	if f.TipoActividad != "" {
		q = q.Where("crm_lead_actividades.tipo_actividad = ?", f.TipoActividad)
	}
	if f.Estado != "" {
		q = q.Where("crm_lead_actividades.estado = ?", f.Estado)
	}
	if f.SoloVencidas {
		q = q.Where(
			"crm_lead_actividades.estado = ? AND crm_lead_actividades.fecha_vencimiento IS NOT NULL AND crm_lead_actividades.fecha_vencimiento < ?",
			ActEstadoPendiente, time.Now().UTC(),
		)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var items []LeadActividad
	err := q.Order("crm_lead_actividades.fecha_vencimiento ASC NULLS LAST, crm_lead_actividades.fecha_actividad DESC").
		Offset((f.Page - 1) * f.PageSize).
		Limit(f.PageSize).
		Find(&items).Error
	return items, total, err
}

func (r *Repository) UpdateActividad(ctx context.Context, a *LeadActividad, cambios map[string]any) error {
	if len(cambios) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Model(a).Updates(cambios).Error
}

func (r *Repository) SoftDeleteActividad(ctx context.Context, id int64) error {
	res := r.db.WithContext(ctx).Delete(&LeadActividad{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errActividadNotFound
	}
	return nil
}

// =====================================================================
// Intereses en productos
// =====================================================================

func (r *Repository) CreateInteres(ctx context.Context, i *LeadInteresProducto) error {
	return r.db.WithContext(ctx).Create(i).Error
}

func (r *Repository) ListInteresesByLead(ctx context.Context, leadID int64) ([]LeadInteresProducto, error) {
	var items []LeadInteresProducto
	err := r.db.WithContext(ctx).
		Where("lead_id = ?", leadID).
		Order("id ASC").
		Find(&items).Error
	return items, err
}

func (r *Repository) DeleteInteres(ctx context.Context, leadID, interesID int64) error {
	res := r.db.WithContext(ctx).
		Where("id = ? AND lead_id = ?", interesID, leadID).
		Delete(&LeadInteresProducto{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errInteresNotFound
	}
	return nil
}

// =====================================================================
// Resumen / Control del pipeline
// =====================================================================

func (r *Repository) ResumenCRM(ctx context.Context, empresaID int32) (*ResumenCRMResponse, error) {
	res := &ResumenCRMResponse{
		EmpresaID:      empresaID,
		PorEtapa:       []EtapaPipelineResumen{},
		PorTemperatura: []TemperaturaResumen{},
	}

	// Totales de leads
	var totales struct {
		TotalLeads         int64
		LeadsConvertidos   int64
		ValorEstimadoTotal float64
	}
	err := r.db.WithContext(ctx).
		Table("crm_leads").
		Select(`
			COUNT(*) AS total_leads,
			COALESCE(SUM(CASE WHEN convertido THEN 1 ELSE 0 END), 0) AS leads_convertidos,
			COALESCE(SUM(valor_estimado), 0) AS valor_estimado_total
		`).
		Where("empresa_id = ? AND deleted_at IS NULL", empresaID).
		Scan(&totales).Error
	if err != nil {
		return nil, err
	}
	res.TotalLeads = totales.TotalLeads
	res.LeadsConvertidos = totales.LeadsConvertidos
	res.ValorEstimadoTotal = totales.ValorEstimadoTotal
	if totales.TotalLeads > 0 {
		res.TasaConversionPct = float64(totales.LeadsConvertidos) * 100.0 / float64(totales.TotalLeads)
	}

	// Desglose por etapa (incluye etapas con 0 leads)
	err = r.db.WithContext(ctx).
		Table("crm_estado_lead e").
		Select(`
			e.id AS estado_id,
			e.nombre,
			e.orden,
			e.es_estado_final,
			COUNT(l.id) AS cantidad,
			COALESCE(SUM(l.valor_estimado), 0) AS valor_estimado
		`).
		Joins("LEFT JOIN crm_leads l ON l.estado_id = e.id AND l.empresa_id = e.empresa_id AND l.deleted_at IS NULL").
		Where("e.empresa_id = ? AND e.deleted_at IS NULL", empresaID).
		Group("e.id, e.nombre, e.orden, e.es_estado_final").
		Order("e.orden ASC, e.id ASC").
		Scan(&res.PorEtapa).Error
	if err != nil {
		return nil, err
	}

	// Desglose por temperatura
	err = r.db.WithContext(ctx).
		Table("crm_leads").
		Select("temperatura, COUNT(*) AS cantidad, COALESCE(SUM(valor_estimado), 0) AS valor_estimado").
		Where("empresa_id = ? AND deleted_at IS NULL", empresaID).
		Group("temperatura").
		Order("temperatura ASC").
		Scan(&res.PorTemperatura).Error
	if err != nil {
		return nil, err
	}

	// Actividades pendientes y vencidas
	var acts struct {
		Pendientes int64
		Vencidas   int64
	}
	now := time.Now().UTC()
	err = r.db.WithContext(ctx).
		Table("crm_lead_actividades a").
		Select(`
			COUNT(*) AS pendientes,
			COALESCE(SUM(CASE WHEN a.fecha_vencimiento IS NOT NULL AND a.fecha_vencimiento < ? THEN 1 ELSE 0 END), 0) AS vencidas
		`, now).
		Joins("JOIN crm_leads l ON l.id = a.lead_id AND l.deleted_at IS NULL").
		Where("l.empresa_id = ? AND a.deleted_at IS NULL AND a.estado = ?", empresaID, ActEstadoPendiente).
		Scan(&acts).Error
	if err != nil {
		return nil, err
	}
	res.ActividadesPendientes = acts.Pendientes
	res.ActividadesVencidas = acts.Vencidas

	return res, nil
}

func escapeLike(s string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(s)
}
