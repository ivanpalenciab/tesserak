package usuarios

import (
	"strings"
	"time"

	"tesserak/backend/internal/auth"
)

// CrearUsuarioRequest es el cuerpo de POST /usuarios.
//   - Si la empresa aún no tiene usuarios, se permite la creación inicial y se asigna
//     automáticamente el rol ADMINISTRADOR.
//   - Si la empresa ya tiene usuarios, solo un ADMINISTRADOR autenticado puede crear nuevos.
//   - Si el rol asignado es ADMINISTRADOR, las sucursales son opcionales (tiene acceso a todas).
//   - Si no es ADMINISTRADOR, debe indicarse al menos una sucursal (sucursal_ids o sucursal_id).
type CrearUsuarioRequest struct {
	EmpresaID     *int32  `json:"empresa_id"     binding:"omitnil,gt=0"`
	RolID         *int32  `json:"rol_id"         binding:"omitnil,gt=0"`
	RolIDs        []int32 `json:"rol_ids"        binding:"omitempty,dive,gt=0"`
	SucursalID    *int32  `json:"sucursal_id"    binding:"omitnil,gt=0"`
	SucursalIDs   []int32 `json:"sucursal_ids"   binding:"omitempty,dive,gt=0"`
	NombreUsuario string  `json:"nombre_usuario" binding:"required,min=3,max=50"`
	Nombres       string  `json:"nombres"        binding:"required,min=1,max=100"`
	Apellidos     string  `json:"apellidos"      binding:"required,min=1,max=100"`
	Email         string  `json:"email"          binding:"required,email,max=255"`
	Password      string  `json:"password"       binding:"required,min=8,max=72"`
	Activo        *bool   `json:"activo"`
}

// ActualizarUsuarioRequest es el cuerpo de PATCH /usuarios/:id.
// Solo se actualizan los campos enviados (no nulos).
type ActualizarUsuarioRequest struct {
	RolID         *int32   `json:"rol_id"         binding:"omitnil,gt=0"`
	RolIDs        *[]int32 `json:"rol_ids"        binding:"omitnil,min=1,dive,gt=0"`
	SucursalID    *int32   `json:"sucursal_id"    binding:"omitnil,gt=0"`
	SucursalIDs   *[]int32 `json:"sucursal_ids"   binding:"omitnil,dive,gt=0"`
	NombreUsuario *string  `json:"nombre_usuario" binding:"omitnil,min=3,max=50"`
	Nombres       *string  `json:"nombres"        binding:"omitnil,min=1,max=100"`
	Apellidos     *string  `json:"apellidos"      binding:"omitnil,min=1,max=100"`
	Email         *string  `json:"email"          binding:"omitnil,email,max=255"`
	Password      *string  `json:"password"       binding:"omitnil,min=8,max=72"`
	Activo        *bool    `json:"activo"`
}

// ActualizarPerfilRequest es el cuerpo de PATCH /usuarios/mi-perfil.
// Permite al usuario autenticado editar sus datos personales y/o cambiar su contraseña
// verificando previamente su contraseña actual.
type ActualizarPerfilRequest struct {
	NombreUsuario  *string `json:"nombre_usuario"  binding:"omitnil,min=3,max=50"`
	Nombres        *string `json:"nombres"         binding:"omitnil,min=1,max=100"`
	Apellidos      *string `json:"apellidos"       binding:"omitnil,min=1,max=100"`
	Email          *string `json:"email"           binding:"omitnil,email,max=255"`
	PasswordActual *string `json:"password_actual" binding:"omitnil,min=1,max=72"`
	PasswordNueva  *string `json:"password_nueva"  binding:"omitnil,min=8,max=72"`
}

// ListarUsuariosQuery son los filtros de GET /usuarios.
type ListarUsuariosQuery struct {
	EmpresaID        int32   `form:"empresa_id"  binding:"omitempty,gt=0"`
	SucursalID       int32   `form:"sucursal_id" binding:"omitempty,gt=0"`
	SucursalIDsScope []int32 `form:"-"`
	Activo           *bool   `form:"activo"`
	Q                string  `form:"q"           binding:"omitempty,max=100"`
	Page             int     `form:"page"        binding:"omitempty,gte=1"`
	PageSize         int     `form:"page_size"   binding:"omitempty,gte=1,lte=100"`
}

// RolSistemaResponse es la representación pública de un rol de sistema.
type RolSistemaResponse struct {
	ID          int32   `json:"id"`
	Codigo      string  `json:"codigo"`
	Nombre      string  `json:"nombre"`
	Descripcion *string `json:"descripcion"`
	EsSistema   bool    `json:"es_sistema"`
	Activo      bool    `json:"activo"`
}

// UsuarioResponse es la representación pública de un usuario (sin password_hash).
type UsuarioResponse struct {
	ID              int64                `json:"id"`
	Key             string               `json:"key"`
	EmpresaID       int32                `json:"empresa_id"`
	SucursalID      int32                `json:"sucursal_id"`
	SucursalIDs     []int32              `json:"sucursal_ids"`
	Roles           []RolSistemaResponse `json:"roles"`
	EsAdministrador bool                 `json:"es_administrador"`
	EsAdminSucursal bool                 `json:"es_admin_sucursal"`
	TodasSucursales bool                 `json:"todas_sucursales"`
	NombreUsuario   string               `json:"nombre_usuario"`
	Nombres         string               `json:"nombres"`
	Apellidos       string               `json:"apellidos"`
	Email           string               `json:"email"`
	Activo          bool                 `json:"activo"`
	UltimoLoginAt   *time.Time           `json:"ultimo_login_at"`
	CreatedAt       time.Time            `json:"created_at"`
	UpdatedAt       time.Time            `json:"updated_at"`
}

type ListaUsuariosResponse struct {
	Data     []UsuarioResponse `json:"data"`
	Page     int               `json:"page"`
	PageSize int               `json:"page_size"`
	Total    int64             `json:"total"`
}

func toRolResponse(r RolSistema) RolSistemaResponse {
	return RolSistemaResponse{
		ID:          r.ID,
		Codigo:      r.Codigo,
		Nombre:      r.Nombre,
		Descripcion: r.Descripcion,
		EsSistema:   r.EsSistema,
		Activo:      r.Activo,
	}
}

func toResponse(u *Usuario, sucursalIDs []int32, roles []RolSistema) UsuarioResponse {
	if sucursalIDs == nil {
		sucursalIDs = []int32{}
	}
	var principal int32
	if len(sucursalIDs) > 0 {
		principal = sucursalIDs[0]
	}

	rolesResp := make([]RolSistemaResponse, 0, len(roles))
	esAdmin := false
	esAdminSucursal := false
	for _, r := range roles {
		rolesResp = append(rolesResp, toRolResponse(r))
		if strings.EqualFold(r.Codigo, auth.RolAdministrador) {
			esAdmin = true
		}
		if strings.EqualFold(r.Codigo, auth.RolAdminSucursal) {
			esAdminSucursal = true
		}
	}

	return UsuarioResponse{
		ID:              u.ID,
		Key:             u.Key,
		EmpresaID:       u.EmpresaID,
		SucursalID:      principal,
		SucursalIDs:     sucursalIDs,
		Roles:           rolesResp,
		EsAdministrador: esAdmin,
		EsAdminSucursal: esAdminSucursal,
		TodasSucursales: esAdmin,
		NombreUsuario:   u.NombreUsuario,
		Nombres:         u.Nombres,
		Apellidos:       u.Apellidos,
		Email:           u.Email,
		Activo:          u.Activo,
		UltimoLoginAt:   u.UltimoLoginAt,
		CreatedAt:       u.CreatedAt,
		UpdatedAt:       u.UpdatedAt,
	}
}
