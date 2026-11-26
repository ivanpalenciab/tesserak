package compras

import (
	"context"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var errNotFound = errors.New("compra no encontrada")

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

func (r *Repository) DB() *gorm.DB {
	return r.db
}

func (r *Repository) ArticulosBelongToEmpresa(ctx context.Context, empresaID int32, articuloIDs []int32) (bool, error) {
	if len(articuloIDs) == 0 {
		return true, nil
	}
	uniq := make(map[int32]struct{}, len(articuloIDs))
	ids := make([]int32, 0, len(articuloIDs))
	for _, id := range articuloIDs {
		if _, ok := uniq[id]; !ok {
			uniq[id] = struct{}{}
			ids = append(ids, id)
		}
	}

	var count int64
	err := r.db.WithContext(ctx).
		Table("articulos").
		Where("empresa_id = ? AND deleted_at IS NULL AND id IN ?", empresaID, ids).
		Count(&count).Error
	return int(count) == len(ids), err
}

func (r *Repository) CreateCompra(ctx context.Context, c *Compra) error {
	return r.db.WithContext(ctx).Omit(clause.Associations).Create(c).Error
}

func (r *Repository) CreateDetalles(ctx context.Context, detalles []CompraDetalle) error {
	if len(detalles) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Omit(clause.Associations).Create(&detalles).Error
}

func (r *Repository) DeleteDetallesByCompra(ctx context.Context, compraID int64) error {
	return r.db.WithContext(ctx).Where("compra_id = ?", compraID).Delete(&CompraDetalle{}).Error
}

func (r *Repository) FindByID(ctx context.Context, empresaID int32, id int64) (*Compra, error) {
	var c Compra
	err := r.db.WithContext(ctx).
		Preload("Moneda").
		Preload("Sucursal").
		Preload("Usuario").
		Preload("Detalles", func(db *gorm.DB) *gorm.DB {
			return db.Order("id ASC")
		}).
		Preload("Detalles.Articulo").
		Where("empresa_id = ? AND id = ?", empresaID, id).
		First(&c).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errNotFound
	}
	return &c, err
}

func (r *Repository) List(ctx context.Context, empresaID int32, f ListarComprasQuery, fechaDesde, fechaHasta *time.Time) ([]Compra, int64, error) {
	q := r.db.WithContext(ctx).Model(&Compra{}).Where("empresa_id = ?", empresaID)

	if f.SucursalID > 0 {
		q = q.Where("sucursal_id = ?", f.SucursalID)
	}
	if f.Estado != "" {
		q = q.Where("estado = ?", f.Estado)
	}
	if f.MonedaID > 0 {
		q = q.Where("moneda_id = ?", f.MonedaID)
	}
	if fechaDesde != nil {
		q = q.Where("fecha_compra >= ?", *fechaDesde)
	}
	if fechaHasta != nil {
		q = q.Where("fecha_compra < ?", fechaHasta.Add(24*time.Hour))
	}
	if s := strings.TrimSpace(f.Q); s != "" {
		like := "%" + escapeLike(s) + "%"
		q = q.Where(
			"consecutivo ILIKE ? OR numero_factura_proveedor ILIKE ? OR proveedor_nombre ILIKE ? OR observaciones ILIKE ?",
			like, like, like, like,
		)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var items []Compra
	err := q.Preload("Moneda").
		Preload("Sucursal").
		Preload("Usuario").
		Order("fecha_compra DESC, id DESC").
		Offset((f.Page - 1) * f.PageSize).
		Limit(f.PageSize).
		Find(&items).Error
	return items, total, err
}

func (r *Repository) UpdateCompra(ctx context.Context, empresaID int32, id int64, cambios map[string]any) error {
	if len(cambios) == 0 {
		return nil
	}
	res := r.db.WithContext(ctx).
		Model(&Compra{}).
		Where("empresa_id = ? AND id = ?", empresaID, id).
		Updates(cambios)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errNotFound
	}
	return nil
}

// SoftDelete marca deleted_at = now() en com_compras manteniendo aislamiento por empresa_id.
func (r *Repository) SoftDelete(ctx context.Context, empresaID int32, id int64) error {
	res := r.db.WithContext(ctx).
		Where("empresa_id = ? AND id = ?", empresaID, id).
		Delete(&Compra{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errNotFound
	}
	return nil
}

func escapeLike(s string) string {
	r := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return r.Replace(s)
}
