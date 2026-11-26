package terceros

import "time"

type CrearRolTerceroRequest struct {
	Codigo      string  `json:"codigo"      binding:"required,min=2,max=30"`
	NombreRol   string  `json:"nombre_rol"  binding:"required,min=2,max=80"`
	Descripcion *string `json:"descripcion" binding:"omitnil,max=255"`
	Activo      *bool   `json:"activo"`
}

type ActualizarRolTerceroRequest struct {
	Codigo      *string `json:"codigo"      binding:"omitnil,min=2,max=30"`
	NombreRol   *string `json:"nombre_rol"  binding:"omitnil,min=2,max=80"`
	Descripcion *string `json:"descripcion" binding:"omitnil,max=255"`
	Activo      *bool   `json:"activo"`
}

type ListarRolesTerceroQuery struct {
	Activo *bool  `form:"activo"`
	Q      string `form:"q" binding:"omitempty,max=100"`
}

type RolTerceroResponse struct {
	ID          int32     `json:"id"`
	Codigo      string    `json:"codigo"`
	NombreRol   string    `json:"nombre_rol"`
	Descripcion *string   `json:"descripcion"`
	Activo      bool      `json:"activo"`
	EsSistema   bool      `json:"es_sistema"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ListaRolesTerceroResponse struct {
	Data []RolTerceroResponse `json:"data"`
}

func toRolResponse(r *RolTercero) RolTerceroResponse {
	return RolTerceroResponse{
		ID:          r.ID,
		Codigo:      r.Codigo,
		NombreRol:   r.NombreRol,
		Descripcion: r.Descripcion,
		Activo:      r.Activo,
		EsSistema:   r.EsSistema,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}
}
