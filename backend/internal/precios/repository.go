package precios

import (
	"context"
	"errors"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	errListaNotFound   = errors.New("lista de precios no encontrada")
	errDetalleNotFound = errors.New("detalle de lista de precios no encontrado")
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// =====================================================================
// listas_precios
// =====================================================================

func (r *Repository) CreateLista(ctx context.Context, lp *ListaPrecio) error {
	return r.db.WithContext(ctx).Omit(clause.Associations).Create(lp).Error
}

func (r *Repository) FindListaByID(ctx context.Context, empresaID, id int32) (*ListaPrecio, error) {
	var lp ListaPrecio
	err := r.db.WithContext(ctx).
		Preload("Moneda").
		Where("empresa_id = ? AND id = ?", empresaID, id).
		First(&lp).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errListaNotFound
	}
	return &lp, err
}

func (r *Repository) ListListas(ctx context.Context, empresaID int32, f ListarListasPreciosQuery) ([]ListaPrecio, int64, error) {
	q := r.db.WithContext(ctx).Model(&ListaPrecio{}).Where("empresa_id = ?", empresaID)

	if f.TipoLista != "" {
		q = q.Where("tipo_lista = ?", f.TipoLista)
	}
	if f.MonedaID > 0 {
		q = q.Where("moneda_id = ?", f.MonedaID)
	}
	if f.Estado != nil {
		q = q.Where("estado = ?", *f.Estado)
	}
	if s := strings.TrimSpace(f.Q); s != "" {
		like := "%" + escapeLike(s) + "%"
		q = q.Where("nombre ILIKE ?", like)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var items []ListaPrecio
	err := q.Preload("Moneda").
		Order("nombre ASC, id DESC").
		Offset((f.Page - 1) * f.PageSize).
		Limit(f.PageSize).
		Find(&items).Error
	return items, total, err
}

func (r *Repository) UpdateLista(ctx context.Context, empresaID, id int32, cambios map[string]any) error {
	if len(cambios) == 0 {
		return nil
	}
	res := r.db.WithContext(ctx).
		Model(&ListaPrecio{}).
		Where("empresa_id = ? AND id = ?", empresaID, id).
		Updates(cambios)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errListaNotFound
	}
	return nil
}

func (r *Repository) SoftDeleteLista(ctx context.Context, empresaID, id int32) error {
	res := r.db.WithContext(ctx).
		Where("empresa_id = ? AND id = ?", empresaID, id).
		Delete(&ListaPrecio{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errListaNotFound
	}
	return nil
}

// =====================================================================
// lista_precios_detalle
// =====================================================================

func (r *Repository) ArticuloBelongsToEmpresa(ctx context.Context, empresaID, articuloID int32) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Table("articulos").
		Where("id = ? AND empresa_id = ? AND deleted_at IS NULL", articuloID, empresaID).
		Count(&count).Error
	return count > 0, err
}

func (r *Repository) CreateDetalle(ctx context.Context, d *ListaPrecioDetalle) error {
	return r.db.WithContext(ctx).Omit(clause.Associations).Create(d).Error
}

func (r *Repository) FindDetalleByID(ctx context.Context, listaPrecioID int32, detalleID int64) (*ListaPrecioDetalle, error) {
	var d ListaPrecioDetalle
	err := r.db.WithContext(ctx).
		Preload("Articulo").
		Preload("UnidadMedida").
		Where("lista_precio_id = ? AND id = ?", listaPrecioID, detalleID).
		First(&d).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errDetalleNotFound
	}
	return &d, err
}

func (r *Repository) ListDetalles(ctx context.Context, listaPrecioID int32, f ListarDetallesPrecioQuery) ([]ListaPrecioDetalle, int64, error) {
	q := r.db.WithContext(ctx).
		Model(&ListaPrecioDetalle{}).
		Where("lista_precios_detalle.lista_precio_id = ?", listaPrecioID)

	if f.ArticuloID > 0 {
		q = q.Where("lista_precios_detalle.articulo_id = ?", f.ArticuloID)
	}
	if f.UnidadMedidaID > 0 {
		q = q.Where("lista_precios_detalle.unidad_medida_id = ?", f.UnidadMedidaID)
	}
	if s := strings.TrimSpace(f.Q); s != "" {
		like := "%" + escapeLike(s) + "%"
		q = q.Joins("JOIN articulos a ON a.id = lista_precios_detalle.articulo_id").
			Where("a.codigo ILIKE ? OR a.nombre ILIKE ?", like, like)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var items []ListaPrecioDetalle
	err := q.Preload("Articulo").
		Preload("UnidadMedida").
		Order("lista_precios_detalle.articulo_id ASC, lista_precios_detalle.cantidad_minima ASC, lista_precios_detalle.id DESC").
		Offset((f.Page - 1) * f.PageSize).
		Limit(f.PageSize).
		Find(&items).Error
	return items, total, err
}

func (r *Repository) UpdateDetalle(ctx context.Context, listaPrecioID int32, detalleID int64, cambios map[string]any) error {
	if len(cambios) == 0 {
		return nil
	}
	res := r.db.WithContext(ctx).
		Model(&ListaPrecioDetalle{}).
		Where("lista_precio_id = ? AND id = ?", listaPrecioID, detalleID).
		Updates(cambios)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errDetalleNotFound
	}
	return nil
}

func (r *Repository) DeleteDetalle(ctx context.Context, listaPrecioID int32, detalleID int64) error {
	res := r.db.WithContext(ctx).
		Where("lista_precio_id = ? AND id = ?", listaPrecioID, detalleID).
		Delete(&ListaPrecioDetalle{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errDetalleNotFound
	}
	return nil
}

func escapeLike(s string) string {
	r := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return r.Replace(s)
}
