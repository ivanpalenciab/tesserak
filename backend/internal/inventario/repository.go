package inventario

import (
	"context"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	errUbicacionNotFound = errors.New("ubicación no encontrada")
	errTipoMovNotFound   = errors.New("tipo de movimiento no encontrado")
	errDocumentoNotFound = errors.New("documento de inventario no encontrado")
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

// =====================================================================
// Sucursales
// =====================================================================

func (r *Repository) ListSucursales(ctx context.Context, empresaID int32) ([]SucursalRef, error) {
	var items []SucursalRef
	err := r.db.WithContext(ctx).
		Where("empresa_id = ? AND deleted_at IS NULL AND activo = TRUE", empresaID).
		Order("nombre ASC, id ASC").
		Find(&items).Error
	return items, err
}

// =====================================================================
// Ubicaciones
// =====================================================================

func (r *Repository) CreateUbicacion(ctx context.Context, u *Ubicacion) error {
	return r.db.WithContext(ctx).Omit(clause.Associations).Create(u).Error
}

func (r *Repository) FindUbicacionByID(ctx context.Context, empresaID, id int32) (*Ubicacion, error) {
	var u Ubicacion
	err := r.db.WithContext(ctx).
		Preload("Sucursal").
		Where("empresa_id = ? AND id = ?", empresaID, id).
		First(&u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errUbicacionNotFound
	}
	return &u, err
}

func (r *Repository) ListUbicaciones(ctx context.Context, empresaID int32, f ListarUbicacionesQuery) ([]Ubicacion, error) {
	q := r.db.WithContext(ctx).Model(&Ubicacion{}).Where("empresa_id = ?", empresaID)
	if f.SucursalID > 0 {
		q = q.Where("sucursal_id = ?", f.SucursalID)
	}
	if f.Activo != nil {
		q = q.Where("activo = ?", *f.Activo)
	}
	if s := strings.TrimSpace(f.Q); s != "" {
		like := "%" + escapeLike(s) + "%"
		q = q.Where("codigo ILIKE ? OR nombre ILIKE ? OR descripcion ILIKE ?", like, like, like)
	}
	var items []Ubicacion
	err := q.Preload("Sucursal").Order("nombre ASC, id ASC").Find(&items).Error
	return items, err
}

func (r *Repository) UpdateUbicacion(ctx context.Context, empresaID, id int32, cambios map[string]any) error {
	if len(cambios) == 0 {
		return nil
	}
	res := r.db.WithContext(ctx).
		Model(&Ubicacion{}).
		Where("empresa_id = ? AND id = ?", empresaID, id).
		Updates(cambios)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errUbicacionNotFound
	}
	return nil
}

func (r *Repository) SoftDeleteUbicacion(ctx context.Context, empresaID, id int32) error {
	res := r.db.WithContext(ctx).
		Where("empresa_id = ? AND id = ?", empresaID, id).
		Delete(&Ubicacion{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errUbicacionNotFound
	}
	return nil
}

// =====================================================================
// Tipos de movimiento
// =====================================================================

func (r *Repository) ListTiposMovimiento(ctx context.Context) ([]TipoMovimientoInventario, error) {
	var items []TipoMovimientoInventario
	err := r.db.WithContext(ctx).
		Where("activo = TRUE").
		Order("id ASC").
		Find(&items).Error
	return items, err
}

func (r *Repository) FindTipoMovimientoByID(ctx context.Context, id int32) (*TipoMovimientoInventario, error) {
	var t TipoMovimientoInventario
	err := r.db.WithContext(ctx).Where("id = ? AND activo = TRUE", id).First(&t).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errTipoMovNotFound
	}
	return &t, err
}

// =====================================================================
// Validaciones de pertenencia multi-tenant
// =====================================================================

func (r *Repository) ArticulosBelongToEmpresa(ctx context.Context, empresaID int32, ids []int32) (bool, error) {
	uniq := deduplicateInt32(ids)
	if len(uniq) == 0 {
		return true, nil
	}
	var count int64
	err := r.db.WithContext(ctx).
		Table("articulos").
		Where("empresa_id = ? AND deleted_at IS NULL AND id IN ?", empresaID, uniq).
		Count(&count).Error
	return int(count) == len(uniq), err
}

func (r *Repository) UbicacionesBelongToEmpresa(ctx context.Context, empresaID int32, ids []int32) (bool, error) {
	uniq := deduplicateInt32(ids)
	if len(uniq) == 0 {
		return true, nil
	}
	var count int64
	err := r.db.WithContext(ctx).
		Table("inv_ubicaciones").
		Where("empresa_id = ? AND deleted_at IS NULL AND activo = TRUE AND id IN ?", empresaID, uniq).
		Count(&count).Error
	return int(count) == len(uniq), err
}

func (r *Repository) TerceroBelongsToEmpresa(ctx context.Context, empresaID int32, terceroID int64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Table("tercero_maestro").
		Where("empresa_id = ? AND deleted_at IS NULL AND id = ?", empresaID, terceroID).
		Count(&count).Error
	return count > 0, err
}

// =====================================================================
// Stock materializado (con bloqueo FOR UPDATE para actualizaciones atómicas)
// =====================================================================

func (r *Repository) FindStockForUpdate(ctx context.Context, empresaID, articuloID, ubicacionID int32) (*InventarioStock, error) {
	var s InventarioStock
	err := r.db.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("empresa_id = ? AND articulo_id = ? AND ubicacion_id = ?", empresaID, articuloID, ubicacionID).
		First(&s).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &s, err
}

func (r *Repository) UpsertStock(ctx context.Context, s *InventarioStock) error {
	if s.ID == 0 {
		return r.db.WithContext(ctx).Omit(clause.Associations).Create(s).Error
	}
	return r.db.WithContext(ctx).
		Model(&InventarioStock{}).
		Where("id = ?", s.ID).
		Updates(map[string]any{
			"cantidad":       s.Cantidad,
			"costo_promedio": s.CostoPromedio,
			"updated_at":     time.Now().UTC(),
		}).Error
}

func (r *Repository) ListStock(ctx context.Context, empresaID int32, f ListarStockQuery) ([]InventarioStock, int64, error) {
	q := r.db.WithContext(ctx).
		Model(&InventarioStock{}).
		Joins("JOIN articulos a ON a.id = inv_stock.articulo_id AND a.deleted_at IS NULL").
		Joins("JOIN inv_ubicaciones u ON u.id = inv_stock.ubicacion_id AND u.deleted_at IS NULL").
		Where("inv_stock.empresa_id = ?", empresaID)

	if f.UbicacionID > 0 {
		q = q.Where("inv_stock.ubicacion_id = ?", f.UbicacionID)
	}
	if len(f.SucursalIDsScope) > 0 {
		q = q.Where("u.sucursal_id IN ?", f.SucursalIDsScope)
	}
	if f.ArticuloID > 0 {
		q = q.Where("inv_stock.articulo_id = ?", f.ArticuloID)
	}
	if f.SoloConStock != nil && *f.SoloConStock {
		q = q.Where("inv_stock.cantidad > 0")
	}
	if s := strings.TrimSpace(f.Q); s != "" {
		like := "%" + escapeLike(s) + "%"
		q = q.Where("a.codigo ILIKE ? OR a.nombre ILIKE ? OR u.codigo ILIKE ? OR u.nombre ILIKE ?", like, like, like, like)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var items []InventarioStock
	err := q.Preload("Articulo").
		Preload("Ubicacion").
		Order("u.nombre ASC, a.nombre ASC").
		Offset((f.Page - 1) * f.PageSize).
		Limit(f.PageSize).
		Find(&items).Error
	return items, total, err
}

// =====================================================================
// Documentos y movimientos (Kardex)
// =====================================================================

func (r *Repository) CreateDocumento(ctx context.Context, d *DocumentoInventario) error {
	return r.db.WithContext(ctx).Omit(clause.Associations).Create(d).Error
}

func (r *Repository) CreateDetalles(ctx context.Context, detalles []MovimientoInventarioDetalle) error {
	if len(detalles) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Omit(clause.Associations).Create(&detalles).Error
}

func (r *Repository) FindDocumentoByID(ctx context.Context, empresaID int32, id int64) (*DocumentoInventario, error) {
	var d DocumentoInventario
	err := r.db.WithContext(ctx).
		Preload("Sucursal").
		Preload("TipoMovimiento").
		Preload("Tercero").
		Preload("Usuario").
		Preload("Detalles", func(db *gorm.DB) *gorm.DB {
			return db.Order("id ASC")
		}).
		Preload("Detalles.Articulo").
		Preload("Detalles.UbicacionOrigen").
		Preload("Detalles.UbicacionDestino").
		Where("empresa_id = ? AND id = ?", empresaID, id).
		First(&d).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errDocumentoNotFound
	}
	return &d, err
}

func (r *Repository) ListDocumentos(ctx context.Context, empresaID int32, f ListarMovimientosQuery, desde, hasta *time.Time) ([]DocumentoInventario, int64, error) {
	q := r.db.WithContext(ctx).Model(&DocumentoInventario{}).Where("inv_documentos.empresa_id = ?", empresaID)

	if f.TipoMovimientoID > 0 {
		q = q.Where("inv_documentos.tipo_movimiento_id = ?", f.TipoMovimientoID)
	}
	if len(f.SucursalIDsScope) > 0 {
		q = q.Where("inv_documentos.sucursal_id IN ?", f.SucursalIDsScope)
	}
	if f.UbicacionID > 0 {
		q = q.Where(
			"EXISTS (SELECT 1 FROM inv_movimientos_detalle m WHERE m.documento_id = inv_documentos.id AND (m.ubicacion_origen_id = ? OR m.ubicacion_destino_id = ?))",
			f.UbicacionID, f.UbicacionID,
		)
	}
	if f.ArticuloID > 0 {
		q = q.Where(
			"EXISTS (SELECT 1 FROM inv_movimientos_detalle m WHERE m.documento_id = inv_documentos.id AND m.articulo_id = ?)",
			f.ArticuloID,
		)
	}
	if desde != nil {
		q = q.Where("inv_documentos.fecha >= ?", *desde)
	}
	if hasta != nil {
		q = q.Where("inv_documentos.fecha < ?", hasta.Add(24*time.Hour))
	}
	if s := strings.TrimSpace(f.Q); s != "" {
		like := "%" + escapeLike(s) + "%"
		q = q.Where("inv_documentos.consecutivo ILIKE ? OR inv_documentos.motivo ILIKE ? OR inv_documentos.observaciones ILIKE ?", like, like, like)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var items []DocumentoInventario
	err := q.Preload("Sucursal").
		Preload("TipoMovimiento").
		Preload("Tercero").
		Preload("Usuario").
		Order("inv_documentos.fecha DESC, inv_documentos.id DESC").
		Offset((f.Page - 1) * f.PageSize).
		Limit(f.PageSize).
		Find(&items).Error
	return items, total, err
}

func deduplicateInt32(ids []int32) []int32 {
	seen := make(map[int32]struct{}, len(ids))
	out := make([]int32, 0, len(ids))
	for _, id := range ids {
		if _, ok := seen[id]; !ok {
			seen[id] = struct{}{}
			out = append(out, id)
		}
	}
	return out
}

func escapeLike(s string) string {
	r := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return r.Replace(s)
}
