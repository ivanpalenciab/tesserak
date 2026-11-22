package usuarios

import (
	"context"
	"errors"
	"strings"

	"gorm.io/gorm"
)

var (
	errNotFound         = errors.New("usuario no encontrado")
	errSucursalNotFound = errors.New("sucursal no encontrada")
)

// Repository encapsula el acceso a datos de usuarios con GORM.
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

// EmpresaDeSucursal devuelve la empresa a la que pertenece una sucursal activa.
func (r *Repository) EmpresaDeSucursal(ctx context.Context, sucursalID int32) (int32, error) {
	var empresaID int32
	res := r.db.WithContext(ctx).
		Table("sucursales").
		Select("empresa_id").
		Where("id = ? AND deleted_at IS NULL", sucursalID).
		Limit(1).
		Scan(&empresaID)
	if res.Error != nil {
		return 0, res.Error
	}
	if res.RowsAffected == 0 {
		return 0, errSucursalNotFound
	}
	return empresaID, nil
}

func (r *Repository) Create(ctx context.Context, u *Usuario) error {
	return r.db.WithContext(ctx).Create(u).Error
}

// FindByID busca un usuario no eliminado (GORM excluye deleted_at automáticamente).
func (r *Repository) FindByID(ctx context.Context, id int64) (*Usuario, error) {
	var u Usuario
	err := r.db.WithContext(ctx).First(&u, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errNotFound
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *Repository) List(ctx context.Context, f ListarUsuariosQuery) ([]Usuario, int64, error) {
	q := r.db.WithContext(ctx).Model(&Usuario{})

	if f.EmpresaID > 0 {
		q = q.Where("empresa_id = ?", f.EmpresaID)
	}
	if f.SucursalID > 0 {
		q = q.Where("sucursal_id = ?", f.SucursalID)
	}
	if f.Activo != nil {
		q = q.Where("activo = ?", *f.Activo)
	}
	if s := strings.TrimSpace(f.Q); s != "" {
		like := "%" + escapeLike(s) + "%"
		q = q.Where(
			"nombres ILIKE ? OR apellidos ILIKE ? OR email ILIKE ? OR nombre_usuario ILIKE ?",
			like, like, like, like,
		)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var items []Usuario
	err := q.Order("id ASC").
		Offset((f.Page - 1) * f.PageSize).
		Limit(f.PageSize).
		Find(&items).Error
	return items, total, err
}

// Update actualiza solo las columnas indicadas. Se usa un map para que
// valores "cero" (ej. activo=false) también se guarden.
func (r *Repository) Update(ctx context.Context, u *Usuario, cambios map[string]any) error {
	if len(cambios) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Model(u).Updates(cambios).Error
}

// SoftDelete marca deleted_at = now(). GORM lo hace automáticamente
// porque el modelo tiene un campo gorm.DeletedAt.
func (r *Repository) SoftDelete(ctx context.Context, id int64) error {
	res := r.db.WithContext(ctx).Delete(&Usuario{}, id)
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
