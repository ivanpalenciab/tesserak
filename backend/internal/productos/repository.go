package productos

import (
	"context"
	"errors"
	"strings"

	"gorm.io/gorm"
)

var errNotFound = errors.New("artículo no encontrado")

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

func (r *Repository) FindTipoArticuloByID(ctx context.Context, id int32) (*TipoArticulo, error) {
	var t TipoArticulo
	err := r.db.WithContext(ctx).First(&t, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errTipoArticuloNotFound
	}
	return &t, err
}

func (r *Repository) CreateArticulo(ctx context.Context, a *Articulo) error {
	return r.db.WithContext(ctx).Create(a).Error
}

func (r *Repository) CreateProductoExt(ctx context.Context, ext *ProductoExtension) error {
	return r.db.WithContext(ctx).Create(ext).Error
}

func (r *Repository) CreateServicioExt(ctx context.Context, ext *ServicioExtension) error {
	return r.db.WithContext(ctx).Create(ext).Error
}

func (r *Repository) FindByID(ctx context.Context, empresaID, id int32) (*Articulo, error) {
	var a Articulo
	err := r.db.WithContext(ctx).
		Preload("TipoArticulo").
		Preload("Producto").
		Preload("Servicio").
		Where("empresa_id = ? AND id = ?", empresaID, id).
		First(&a).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errNotFound
	}
	return &a, err
}

func (r *Repository) List(ctx context.Context, empresaID int32, f ListarProductosQuery) ([]Articulo, int64, error) {
	q := r.db.WithContext(ctx).Model(&Articulo{}).Where("empresa_id = ?", empresaID)

	if f.TipoArticuloID > 0 {
		q = q.Where("tipo_articulo_id = ?", f.TipoArticuloID)
	}
	if f.Estado != nil {
		q = q.Where("estado = ?", *f.Estado)
	}
	if f.CategoriaID > 0 {
		q = q.Where("categoria_id = ?", f.CategoriaID)
	}
	if f.MarcaID > 0 {
		q = q.Where("marca_id = ?", f.MarcaID)
	}
	if f.UnidadMedidaID > 0 {
		q = q.Where("unidad_medida_id = ?", f.UnidadMedidaID)
	}
	if s := strings.TrimSpace(f.Q); s != "" {
		like := "%" + escapeLike(s) + "%"
		q = q.Where("codigo ILIKE ? OR nombre ILIKE ? OR descripcion ILIKE ?", like, like, like)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var items []Articulo
	err := q.Preload("TipoArticulo").
		Preload("Producto").
		Preload("Servicio").
		Order("nombre ASC, id DESC").
		Offset((f.Page - 1) * f.PageSize).
		Limit(f.PageSize).
		Find(&items).Error
	return items, total, err
}

func (r *Repository) UpdateArticulo(ctx context.Context, a *Articulo, cambios map[string]any) error {
	if len(cambios) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Model(a).Updates(cambios).Error
}

func (r *Repository) UpdateProductoExt(ctx context.Context, articuloID int32, cambios map[string]any) error {
	if len(cambios) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Model(&ProductoExtension{}).Where("articulo_id = ?", articuloID).Updates(cambios).Error
}

func (r *Repository) UpdateServicioExt(ctx context.Context, articuloID int32, cambios map[string]any) error {
	if len(cambios) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Model(&ServicioExtension{}).Where("articulo_id = ?", articuloID).Updates(cambios).Error
}

// SoftDelete marca deleted_at = now() en articulos manteniendo aislamiento por empresa_id.
func (r *Repository) SoftDelete(ctx context.Context, empresaID, id int32) error {
	res := r.db.WithContext(ctx).
		Where("empresa_id = ? AND id = ?", empresaID, id).
		Delete(&Articulo{})
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
