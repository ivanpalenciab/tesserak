package usuarios

import "time"

// CrearUsuarioRequest es el cuerpo de POST /usuarios.
// La empresa se deduce de la sucursal; el cliente no la envía.
type CrearUsuarioRequest struct {
	SucursalID    int32  `json:"sucursal_id"    binding:"required,gt=0"`
	NombreUsuario string `json:"nombre_usuario" binding:"required,min=3,max=50"`
	Nombres       string `json:"nombres"        binding:"required,min=1,max=100"`
	Apellidos     string `json:"apellidos"      binding:"required,min=1,max=100"`
	Email         string `json:"email"          binding:"required,email,max=255"`
	Password      string `json:"password"       binding:"required,min=8,max=72"`
	Activo        *bool  `json:"activo"`
}

// ActualizarUsuarioRequest es el cuerpo de PATCH /usuarios/:id.
// Solo se actualizan los campos enviados (no nulos).
type ActualizarUsuarioRequest struct {
	SucursalID    *int32  `json:"sucursal_id"    binding:"omitnil,gt=0"`
	NombreUsuario *string `json:"nombre_usuario" binding:"omitnil,min=3,max=50"`
	Nombres       *string `json:"nombres"        binding:"omitnil,min=1,max=100"`
	Apellidos     *string `json:"apellidos"      binding:"omitnil,min=1,max=100"`
	Email         *string `json:"email"          binding:"omitnil,email,max=255"`
	Password      *string `json:"password"       binding:"omitnil,min=8,max=72"`
	Activo        *bool   `json:"activo"`
}

// ListarUsuariosQuery son los filtros de GET /usuarios.
type ListarUsuariosQuery struct {
	EmpresaID  int32  `form:"empresa_id"  binding:"omitempty,gt=0"`
	SucursalID int32  `form:"sucursal_id" binding:"omitempty,gt=0"`
	Activo     *bool  `form:"activo"`
	Q          string `form:"q"           binding:"omitempty,max=100"`
	Page       int    `form:"page"        binding:"omitempty,gte=1"`
	PageSize   int    `form:"page_size"   binding:"omitempty,gte=1,lte=100"`
}

// UsuarioResponse es la representación pública de un usuario (sin password_hash).
type UsuarioResponse struct {
	ID            int64      `json:"id"`
	Key           string     `json:"key"`
	EmpresaID     int32      `json:"empresa_id"`
	SucursalID    int32      `json:"sucursal_id"`
	NombreUsuario string     `json:"nombre_usuario"`
	Nombres       string     `json:"nombres"`
	Apellidos     string     `json:"apellidos"`
	Email         string     `json:"email"`
	Activo        bool       `json:"activo"`
	UltimoLoginAt *time.Time `json:"ultimo_login_at"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type ListaUsuariosResponse struct {
	Data     []UsuarioResponse `json:"data"`
	Page     int               `json:"page"`
	PageSize int               `json:"page_size"`
	Total    int64             `json:"total"`
}

func toResponse(u *Usuario) UsuarioResponse {
	return UsuarioResponse{
		ID:            u.ID,
		Key:           u.Key,
		EmpresaID:     u.EmpresaID,
		SucursalID:    u.SucursalID,
		NombreUsuario: u.NombreUsuario,
		Nombres:       u.Nombres,
		Apellidos:     u.Apellidos,
		Email:         u.Email,
		Activo:        u.Activo,
		UltimoLoginAt: u.UltimoLoginAt,
		CreatedAt:     u.CreatedAt,
		UpdatedAt:     u.UpdatedAt,
	}
}
