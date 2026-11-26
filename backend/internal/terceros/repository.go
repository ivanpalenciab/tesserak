package terceros

import (
	"context"
	"errors"
	"strings"

	"gorm.io/gorm"
)

var errRolNotFound = errors.New("rol de tercero no encontrado")

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateRol(ctx context.Context, rol *RolTercero) error {
	return r.db.WithContext(ctx).Create(rol).Error
}

func (r *Repository) FindRolByID(ctx context.Context, id int32) (*RolTercero, error) {
	var rol RolTercero
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&rol).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errRolNotFound
	}
	return &rol, err
}

func (r *Repository) ListRoles(ctx context.Context, f ListarRolesTerceroQuery) ([]RolTercero, error) {
	q := r.db.WithContext(ctx).Model(&RolTercero{})

	if f.Activo != nil {
		q = q.Where("activo = ?", *f.Activo)
	}
	if s := strings.TrimSpace(f.Q); s != "" {
		like := "%" + escapeLike(s) + "%"
		q = q.Where("codigo ILIKE ? OR nombre_rol ILIKE ? OR descripcion ILIKE ?", like, like, like)
	}

	var items []RolTercero
	err := q.Order("es_sistema DESC, nombre_rol ASC, id ASC").Find(&items).Error
	return items, err
}

func (r *Repository) UpdateRol(ctx context.Context, id int32, cambios map[string]any) error {
	if len(cambios) == 0 {
		return nil
	}
	res := r.db.WithContext(ctx).Model(&RolTercero{}).Where("id = ?", id).Updates(cambios)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errRolNotFound
	}
	return nil
}

func (r *Repository) SoftDeleteRol(ctx context.Context, id int32) error {
	res := r.db.WithContext(ctx).Where("id = ?", id).Delete(&RolTercero{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errRolNotFound
	}
	return nil
}

func escapeLike(s string) string {
	r := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return r.Replace(s)
}
