package terceros

import (
	"context"
	"errors"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var errTerceroNotFound = errors.New("tercero no encontrado")

func (r *Repository) WithTx(ctx context.Context, fn func(tx *Repository) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(&Repository{db: tx})
	})
}

func (r *Repository) RolesActivosExist(ctx context.Context, rolIDs []int32) (bool, error) {
	if len(rolIDs) == 0 {
		return false, nil
	}
	var count int64
	err := r.db.WithContext(ctx).
		Model(&RolTercero{}).
		Where("activo = TRUE AND id IN ?", rolIDs).
		Count(&count).Error
	return int(count) == len(rolIDs), err
}

func (r *Repository) CreateTercero(ctx context.Context, t *TerceroMaestro) error {
	return r.db.WithContext(ctx).Omit(clause.Associations).Create(t).Error
}

func (r *Repository) ReplaceTerceroRoles(ctx context.Context, terceroID int64, rolIDs []int32) error {
	if err := r.db.WithContext(ctx).Where("tercero_id = ?", terceroID).Delete(&TerceroRol{}).Error; err != nil {
		return err
	}
	if len(rolIDs) == 0 {
		return nil
	}
	asignaciones := make([]TerceroRol, 0, len(rolIDs))
	for _, rid := range rolIDs {
		asignaciones = append(asignaciones, TerceroRol{
			TerceroID: terceroID,
			RolID:     rid,
		})
	}
	return r.db.WithContext(ctx).Omit(clause.Associations).Create(&asignaciones).Error
}

func (r *Repository) FindTerceroByID(ctx context.Context, empresaID int32, id int64) (*TerceroMaestro, error) {
	var t TerceroMaestro
	err := r.db.WithContext(ctx).
		Preload("Pais").
		Preload("Roles", func(db *gorm.DB) *gorm.DB {
			return db.Order("rol_id ASC")
		}).
		Preload("Roles.Rol").
		Where("empresa_id = ? AND id = ?", empresaID, id).
		First(&t).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errTerceroNotFound
	}
	return &t, err
}

func (r *Repository) ListTerceros(ctx context.Context, empresaID int32, f ListarTercerosQuery) ([]TerceroMaestro, int64, error) {
	q := r.db.WithContext(ctx).Model(&TerceroMaestro{}).Where("tercero_maestro.empresa_id = ?", empresaID)

	if f.TipoIdentificacion != "" {
		q = q.Where("tercero_maestro.tipo_identificacion = ?", f.TipoIdentificacion)
	}
	if f.NaturalezaJuridica != "" {
		q = q.Where("tercero_maestro.naturaleza_juridica = ?", f.NaturalezaJuridica)
	}
	if f.PaisID > 0 {
		q = q.Where("tercero_maestro.pais_id = ?", f.PaisID)
	}
	if f.RolID > 0 {
		q = q.Where(
			"EXISTS (SELECT 1 FROM tercero_empresa_roles ter WHERE ter.tercero_id = tercero_maestro.id AND ter.rol_id = ?)",
			f.RolID,
		)
	}
	if f.Activo != nil {
		q = q.Where("tercero_maestro.activo = ?", *f.Activo)
	}
	if s := strings.TrimSpace(f.Q); s != "" {
		like := "%" + escapeLike(s) + "%"
		q = q.Where(
			"tercero_maestro.numero_identificacion ILIKE ? OR tercero_maestro.razon_social ILIKE ? OR tercero_maestro.email ILIKE ? OR tercero_maestro.telefono ILIKE ?",
			like, like, like, like,
		)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var items []TerceroMaestro
	err := q.Preload("Pais").
		Preload("Roles", func(db *gorm.DB) *gorm.DB {
			return db.Order("rol_id ASC")
		}).
		Preload("Roles.Rol").
		Order("tercero_maestro.razon_social ASC, tercero_maestro.id DESC").
		Offset((f.Page - 1) * f.PageSize).
		Limit(f.PageSize).
		Find(&items).Error
	return items, total, err
}

func (r *Repository) UpdateTercero(ctx context.Context, empresaID int32, id int64, cambios map[string]any) error {
	if len(cambios) == 0 {
		return nil
	}
	res := r.db.WithContext(ctx).
		Model(&TerceroMaestro{}).
		Where("empresa_id = ? AND id = ?", empresaID, id).
		Updates(cambios)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errTerceroNotFound
	}
	return nil
}

// SoftDeleteTercero marca deleted_at = now() manteniendo aislamiento por empresa_id.
func (r *Repository) SoftDeleteTercero(ctx context.Context, empresaID int32, id int64) error {
	res := r.db.WithContext(ctx).
		Where("empresa_id = ? AND id = ?", empresaID, id).
		Delete(&TerceroMaestro{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errTerceroNotFound
	}
	return nil
}
