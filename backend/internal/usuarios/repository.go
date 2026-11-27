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
	errRolNotFound      = errors.New("rol de sistema no encontrado")
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

// EmpresasDeSucursales devuelve el mapa sucursal_id -> empresa_id para las sucursales activas encontradas.
func (r *Repository) EmpresasDeSucursales(ctx context.Context, sucursalIDs []int32) (map[int32]int32, error) {
	type row struct {
		ID        int32 `gorm:"column:id"`
		EmpresaID int32 `gorm:"column:empresa_id"`
	}
	var rows []row
	if err := r.db.WithContext(ctx).
		Table("sucursales").
		Select("id, empresa_id").
		Where("id IN ? AND deleted_at IS NULL", sucursalIDs).
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	out := make(map[int32]int32, len(rows))
	for _, rw := range rows {
		out[rw.ID] = rw.EmpresaID
	}
	return out, nil
}

// SucursalesPorUsuario devuelve los IDs de sucursales asignadas a un usuario,
// ordenados con la sucursal principal primero.
func (r *Repository) SucursalesPorUsuario(ctx context.Context, usuarioID int64) ([]int32, error) {
	var ids []int32
	err := r.db.WithContext(ctx).
		Table("usuario_sucursales").
		Select("sucursal_id").
		Where("usuario_id = ?", usuarioID).
		Order("es_principal DESC, sucursal_id ASC").
		Pluck("sucursal_id", &ids).Error
	return ids, err
}

// SucursalesPorUsuarios devuelve el mapa usuario_id -> []sucursal_id (principal primero).
func (r *Repository) SucursalesPorUsuarios(ctx context.Context, usuarioIDs []int64) (map[int64][]int32, error) {
	out := make(map[int64][]int32, len(usuarioIDs))
	if len(usuarioIDs) == 0 {
		return out, nil
	}
	var rels []UsuarioSucursal
	err := r.db.WithContext(ctx).
		Where("usuario_id IN ?", usuarioIDs).
		Order("usuario_id ASC, es_principal DESC, sucursal_id ASC").
		Find(&rels).Error
	if err != nil {
		return nil, err
	}
	for _, rel := range rels {
		out[rel.UsuarioID] = append(out[rel.UsuarioID], rel.SucursalID)
	}
	return out, nil
}

// ReemplazarSucursales reemplaza las sucursales asignadas a un usuario en usuario_sucursales.
// La primera sucursal de la lista se marca como es_principal = TRUE.
func (r *Repository) ReemplazarSucursales(ctx context.Context, usuarioID int64, empresaID int32, sucursalIDs []int32) error {
	if err := r.db.WithContext(ctx).
		Where("usuario_id = ?", usuarioID).
		Delete(&UsuarioSucursal{}).Error; err != nil {
		return err
	}
	if len(sucursalIDs) == 0 {
		return nil
	}
	rels := make([]UsuarioSucursal, 0, len(sucursalIDs))
	for i, sid := range sucursalIDs {
		rels = append(rels, UsuarioSucursal{
			UsuarioID:   usuarioID,
			SucursalID:  sid,
			EmpresaID:   empresaID,
			EsPrincipal: i == 0,
		})
	}
	return r.db.WithContext(ctx).Create(&rels).Error
}

// ContarUsuariosEmpresa devuelve cuántos usuarios no eliminados existen en la empresa.
func (r *Repository) ContarUsuariosEmpresa(ctx context.Context, empresaID int32) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).
		Model(&Usuario{}).
		Where("empresa_id = ?", empresaID).
		Count(&total).Error
	return total, err
}

// ListarSucursalesEmpresa devuelve las sucursales activas de una empresa.
func (r *Repository) ListarSucursalesEmpresa(ctx context.Context, empresaID int32) ([]SucursalResumen, error) {
	var items []SucursalResumen
	err := r.db.WithContext(ctx).
		Table("sucursales").
		Select("id, nombre, direccion, activo").
		Where("empresa_id = ? AND activo = TRUE AND deleted_at IS NULL", empresaID).
		Order("id ASC").
		Scan(&items).Error
	return items, err
}

// ListarSucursalesPorIDs devuelve las sucursales activas de una empresa filtradas por IDs.
func (r *Repository) ListarSucursalesPorIDs(ctx context.Context, empresaID int32, ids []int32) ([]SucursalResumen, error) {
	if len(ids) == 0 {
		return []SucursalResumen{}, nil
	}
	var items []SucursalResumen
	err := r.db.WithContext(ctx).
		Table("sucursales").
		Select("id, nombre, direccion, activo").
		Where("empresa_id = ? AND id IN ? AND activo = TRUE AND deleted_at IS NULL", empresaID, ids).
		Order("id ASC").
		Scan(&items).Error
	return items, err
}

// ListarRolesSistema devuelve el catálogo de roles de sistema activos.
func (r *Repository) ListarRolesSistema(ctx context.Context) ([]RolSistema, error) {
	var roles []RolSistema
	err := r.db.WithContext(ctx).
		Where("activo = TRUE").
		Order("id ASC").
		Find(&roles).Error
	return roles, err
}

// BuscarRolPorID obtiene un rol de sistema activo por su ID.
func (r *Repository) BuscarRolPorID(ctx context.Context, id int32) (*RolSistema, error) {
	var rol RolSistema
	err := r.db.WithContext(ctx).
		Where("id = ? AND activo = TRUE", id).
		First(&rol).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errRolNotFound
	}
	if err != nil {
		return nil, err
	}
	return &rol, nil
}

// BuscarRolesPorIDs obtiene los roles de sistema activos cuyos IDs estén en ids.
func (r *Repository) BuscarRolesPorIDs(ctx context.Context, ids []int32) ([]RolSistema, error) {
	if len(ids) == 0 {
		return []RolSistema{}, nil
	}
	var roles []RolSistema
	err := r.db.WithContext(ctx).
		Where("id IN ? AND activo = TRUE", ids).
		Order("id ASC").
		Find(&roles).Error
	if err != nil {
		return nil, err
	}
	if len(roles) != len(ids) {
		return nil, errRolNotFound
	}
	return roles, nil
}

// BuscarRolPorCodigo obtiene un rol de sistema activo por su código (case-insensitive).
func (r *Repository) BuscarRolPorCodigo(ctx context.Context, codigo string) (*RolSistema, error) {
	var rol RolSistema
	err := r.db.WithContext(ctx).
		Where("upper(codigo) = upper(?) AND activo = TRUE", codigo).
		First(&rol).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errRolNotFound
	}
	if err != nil {
		return nil, err
	}
	return &rol, nil
}

// RolesPorUsuario devuelve los roles de sistema activos asignados a un usuario.
func (r *Repository) RolesPorUsuario(ctx context.Context, usuarioID int64) ([]RolSistema, error) {
	var roles []RolSistema
	err := r.db.WithContext(ctx).
		Table("roles_sistema rs").
		Select("rs.*").
		Joins("JOIN usuario_roles ur ON ur.rol_id = rs.id").
		Where("ur.usuario_id = ? AND rs.activo = TRUE", usuarioID).
		Order("rs.id ASC").
		Scan(&roles).Error
	return roles, err
}

// RolesPorUsuarios devuelve el mapa usuario_id -> []RolSistema.
func (r *Repository) RolesPorUsuarios(ctx context.Context, usuarioIDs []int64) (map[int64][]RolSistema, error) {
	out := make(map[int64][]RolSistema, len(usuarioIDs))
	if len(usuarioIDs) == 0 {
		return out, nil
	}
	type row struct {
		UsuarioID int64 `gorm:"column:usuario_id"`
		RolSistema
	}
	var rows []row
	err := r.db.WithContext(ctx).
		Table("usuario_roles ur").
		Select("ur.usuario_id, rs.*").
		Joins("JOIN roles_sistema rs ON rs.id = ur.rol_id AND rs.activo = TRUE").
		Where("ur.usuario_id IN ?", usuarioIDs).
		Order("ur.usuario_id ASC, rs.id ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	for _, rw := range rows {
		out[rw.UsuarioID] = append(out[rw.UsuarioID], rw.RolSistema)
	}
	return out, nil
}

// ReemplazarRoles reemplaza los roles asignados a un usuario en usuario_roles.
func (r *Repository) ReemplazarRoles(ctx context.Context, usuarioID int64, rolIDs []int32) error {
	if err := r.db.WithContext(ctx).
		Where("usuario_id = ?", usuarioID).
		Delete(&UsuarioRol{}).Error; err != nil {
		return err
	}
	if len(rolIDs) == 0 {
		return nil
	}
	rels := make([]UsuarioRol, 0, len(rolIDs))
	for _, rid := range rolIDs {
		rels = append(rels, UsuarioRol{
			UsuarioID: usuarioID,
			RolID:     rid,
		})
	}
	return r.db.WithContext(ctx).Create(&rels).Error
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
		q = q.Where(
			"EXISTS (SELECT 1 FROM usuario_sucursales us WHERE us.usuario_id = usuarios.id AND us.sucursal_id = ?)",
			f.SucursalID,
		)
	}
	if len(f.SucursalIDsScope) > 0 {
		q = q.Where(
			"EXISTS (SELECT 1 FROM usuario_sucursales us WHERE us.usuario_id = usuarios.id AND us.sucursal_id IN ?)",
			f.SucursalIDsScope,
		)
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
