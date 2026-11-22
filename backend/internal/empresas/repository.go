package empresas

import (
	"context"
	"errors"
	"strings"

	"gorm.io/gorm"
)

var errNotFound = errors.New("empresa no encontrada")

// Repository encapsula el acceso a datos de empresas con GORM.
type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// WithTx ejecuta fn dentro de una transacción.
func (r *Repository) WithTx(ctx context.Context, fn func(tx *Repository) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(&Repository{db: tx})
	})
}

func (r *Repository) Create(ctx context.Context, e *Empresa) error {
	return r.db.WithContext(ctx).Create(e).Error
}

// FindByID busca una empresa no eliminada (GORM excluye deleted_at automáticamente).
func (r *Repository) FindByID(ctx context.Context, id int32) (*Empresa, error) {
	var e Empresa
	err := r.db.WithContext(ctx).First(&e, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errNotFound
	}
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *Repository) List(ctx context.Context, f ListarEmpresasQuery) ([]Empresa, int64, error) {
	q := r.db.WithContext(ctx).Model(&Empresa{})

	if f.Activo != nil {
		q = q.Where("activo = ?", *f.Activo)
	}
	if s := strings.TrimSpace(f.Q); s != "" {
		like := "%" + escapeLike(s) + "%"
		q = q.Where(
			"nombre_legal ILIKE ? OR razon_social ILIKE ? OR identificador_fiscal ILIKE ? OR email ILIKE ?",
			like, like, like, like,
		)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var items []Empresa
	err := q.Order("id ASC").
		Offset((f.Page - 1) * f.PageSize).
		Limit(f.PageSize).
		Find(&items).Error
	return items, total, err
}

// Update actualiza solo las columnas indicadas. Se usa un map para que
// valores "cero" (ej. activo=false) también se guarden.
func (r *Repository) Update(ctx context.Context, e *Empresa, cambios map[string]any) error {
	if len(cambios) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Model(e).Updates(cambios).Error
}

// SoftDelete marca deleted_at = now().
func (r *Repository) SoftDelete(ctx context.Context, id int32) error {
	res := r.db.WithContext(ctx).Delete(&Empresa{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errNotFound
	}
	return nil
}

func escapeLike(s string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(s)
}
