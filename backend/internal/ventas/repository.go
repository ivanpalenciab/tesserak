package ventas

import (
	"context"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	errNotFound          = errors.New("venta no encontrada")
	errUbicacionNotFound = errors.New("no existe una ubicación activa para descontar el inventario")
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) DB() *gorm.DB { return r.db }

func (r *Repository) WithTx(ctx context.Context, fn func(tx *Repository) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(&Repository{db: tx})
	})
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
	if err != nil {
		return false, err
	}
	return int(count) == len(ids), nil
}

// GetTarifasImpuestoArticulos devuelve la tarifa (%) de grupos_impuesto asociada a cada artículo.
func (r *Repository) GetTarifasImpuestoArticulos(ctx context.Context, empresaID int32, articuloIDs []int32) (map[int32]float64, error) {
	out := make(map[int32]float64, len(articuloIDs))
	if len(articuloIDs) == 0 {
		return out, nil
	}
	uniq := make(map[int32]struct{}, len(articuloIDs))
	ids := make([]int32, 0, len(articuloIDs))
	for _, id := range articuloIDs {
		if _, ok := uniq[id]; !ok {
			uniq[id] = struct{}{}
			ids = append(ids, id)
		}
	}
	type row struct {
		ArticuloID int32   `gorm:"column:articulo_id"`
		Tarifa     float64 `gorm:"column:tarifa"`
	}
	var rows []row
	err := r.db.WithContext(ctx).
		Table("articulos a").
		Select("a.id AS articulo_id, COALESCE(gi.tarifa, 0) AS tarifa").
		Joins("LEFT JOIN grupos_impuesto gi ON gi.id = a.grupo_impuesto_id AND gi.empresa_id = a.empresa_id").
		Where("a.empresa_id = ? AND a.deleted_at IS NULL AND a.id IN ?", empresaID, ids).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	for _, rw := range rows {
		out[rw.ArticuloID] = rw.Tarifa
	}
	return out, nil
}

func (r *Repository) ClienteBelongsToEmpresa(ctx context.Context, empresaID int32, clienteID int64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Table("tercero_maestro t").
		Joins("JOIN tercero_empresa_roles ter ON ter.tercero_id = t.id").
		Joins("JOIN roles_tercero r ON r.id = ter.rol_id AND r.codigo = 'CLIENTE'").
		Where("t.empresa_id = ? AND t.id = ? AND t.deleted_at IS NULL AND t.activo = TRUE", empresaID, clienteID).
		Count(&count).Error
	return count > 0, err
}

func (r *Repository) ResolveUbicacionDefault(ctx context.Context, empresaID, sucursalID int32) (int32, error) {
	var u UbicacionRef
	err := r.db.WithContext(ctx).
		Table("inv_ubicaciones").
		Where("empresa_id = ? AND sucursal_id = ? AND activo = TRUE AND deleted_at IS NULL", empresaID, sucursalID).
		Order("id ASC").
		First(&u).Error
	if err == nil {
		return u.ID, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, err
	}

	err = r.db.WithContext(ctx).
		Table("inv_ubicaciones").
		Where("empresa_id = ? AND activo = TRUE AND deleted_at IS NULL", empresaID).
		Order("id ASC").
		First(&u).Error
	if err == nil {
		return u.ID, nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, errUbicacionNotFound
	}
	return 0, err
}

func (r *Repository) CreateVenta(ctx context.Context, v *Venta) error {
	return r.db.WithContext(ctx).Omit(clause.Associations).Create(v).Error
}

func (r *Repository) CreateDetalles(ctx context.Context, detalles []VentaDetalle) error {
	if len(detalles) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Omit(clause.Associations).Create(&detalles).Error
}

func (r *Repository) DeleteDetallesByVenta(ctx context.Context, ventaID int64) error {
	return r.db.WithContext(ctx).Where("venta_id = ?", ventaID).Delete(&VentaDetalle{}).Error
}

func (r *Repository) FindByID(ctx context.Context, empresaID int32, id int64) (*Venta, error) {
	var v Venta
	err := r.db.WithContext(ctx).
		Preload("Sucursal").
		Preload("Ubicacion").
		Preload("Cliente").
		Preload("Moneda").
		Preload("Usuario").
		Preload("Detalles", func(db *gorm.DB) *gorm.DB {
			return db.Order("id ASC")
		}).
		Preload("Detalles.Articulo").
		Where("empresa_id = ? AND id = ?", empresaID, id).
		First(&v).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errNotFound
	}
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *Repository) List(ctx context.Context, empresaID int32, f ListarVentasQuery, fechaDesde, fechaHasta *time.Time) ([]Venta, int64, error) {
	q := r.db.WithContext(ctx).
		Model(&Venta{}).
		Joins("LEFT JOIN tercero_maestro cli ON cli.id = ven_ventas.cliente_id").
		Where("ven_ventas.empresa_id = ?", empresaID)

	if f.SucursalID > 0 {
		q = q.Where("ven_ventas.sucursal_id = ?", f.SucursalID)
	}
	if f.UbicacionID > 0 {
		q = q.Where("ven_ventas.ubicacion_id = ?", f.UbicacionID)
	}
	if f.ClienteID > 0 {
		q = q.Where("ven_ventas.cliente_id = ?", f.ClienteID)
	}
	if f.Estado != "" {
		q = q.Where("ven_ventas.estado = ?", f.Estado)
	}
	if fechaDesde != nil {
		q = q.Where("ven_ventas.fecha_venta >= ?", *fechaDesde)
	}
	if fechaHasta != nil {
		q = q.Where("ven_ventas.fecha_venta < ?", fechaHasta.Add(24*time.Hour))
	}
	if s := strings.TrimSpace(f.Q); s != "" {
		like := "%" + escapeLike(s) + "%"
		q = q.Where(
			"ven_ventas.consecutivo ILIKE ? OR ven_ventas.observaciones ILIKE ? OR cli.razon_social ILIKE ? OR cli.numero_identificacion ILIKE ?",
			like, like, like, like,
		)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var items []Venta
	err := q.Preload("Sucursal").
		Preload("Ubicacion").
		Preload("Cliente").
		Preload("Moneda").
		Preload("Usuario").
		Order("ven_ventas.fecha_venta DESC, ven_ventas.id DESC").
		Offset((f.Page - 1) * f.PageSize).
		Limit(f.PageSize).
		Find(&items).Error
	return items, total, err
}

func (r *Repository) UpdateVenta(ctx context.Context, empresaID int32, id int64, cambios map[string]any) error {
	if len(cambios) == 0 {
		return nil
	}
	res := r.db.WithContext(ctx).
		Model(&Venta{}).
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

func (r *Repository) SoftDelete(ctx context.Context, empresaID int32, id int64) error {
	res := r.db.WithContext(ctx).
		Where("empresa_id = ? AND id = ?", empresaID, id).
		Delete(&Venta{})
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
