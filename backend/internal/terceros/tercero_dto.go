package terceros

import "time"

type CrearTerceroRequest struct {
	TipoIdentificacion   string  `json:"tipo_identificacion"   binding:"required,oneof=NIT CC CE PASAPORTE RUT DNI OTRO"`
	NumeroIdentificacion string  `json:"numero_identificacion" binding:"required,min=1,max=50"`
	RazonSocial          string  `json:"razon_social"          binding:"required,min=1,max=200"`
	PaisID               *int32  `json:"pais_id"               binding:"omitnil,gt=0"`
	NaturalezaJuridica   string  `json:"naturaleza_juridica"   binding:"required,oneof=NATURAL JURIDICA"`
	Email                *string `json:"email"                 binding:"omitnil,email,max=255"`
	Telefono             *string `json:"telefono"              binding:"omitnil,max=50"`
	Direccion            *string `json:"direccion"             binding:"omitnil,max=255"`
	Activo               *bool   `json:"activo"`
	RolIDs               []int32 `json:"rol_ids"               binding:"required,min=1,dive,gt=0"`
}

type ActualizarTerceroRequest struct {
	TipoIdentificacion   *string  `json:"tipo_identificacion"   binding:"omitnil,oneof=NIT CC CE PASAPORTE RUT DNI OTRO"`
	NumeroIdentificacion *string  `json:"numero_identificacion" binding:"omitnil,min=1,max=50"`
	RazonSocial          *string  `json:"razon_social"          binding:"omitnil,min=1,max=200"`
	PaisID               *int32   `json:"pais_id"               binding:"omitnil,gt=0"`
	LimpiarPais          bool     `json:"limpiar_pais"`
	NaturalezaJuridica   *string  `json:"naturaleza_juridica"   binding:"omitnil,oneof=NATURAL JURIDICA"`
	Email                *string  `json:"email"                 binding:"omitnil,email,max=255"`
	Telefono             *string  `json:"telefono"              binding:"omitnil,max=50"`
	Direccion            *string  `json:"direccion"             binding:"omitnil,max=255"`
	Activo               *bool    `json:"activo"`
	RolIDs               *[]int32 `json:"rol_ids"               binding:"omitnil,min=1,dive,gt=0"`
}

type ListarTercerosQuery struct {
	TipoIdentificacion string `form:"tipo_identificacion" binding:"omitempty,oneof=NIT CC CE PASAPORTE RUT DNI OTRO"`
	NaturalezaJuridica string `form:"naturaleza_juridica" binding:"omitempty,oneof=NATURAL JURIDICA"`
	PaisID             int32  `form:"pais_id"             binding:"omitempty,gt=0"`
	RolID              int32  `form:"rol_id"              binding:"omitempty,gt=0"`
	Activo             *bool  `form:"activo"`
	Q                  string `form:"q"                   binding:"omitempty,max=100"`
	Page               int    `form:"page"                binding:"omitempty,gte=1"`
	PageSize           int    `form:"page_size"           binding:"omitempty,gte=1,lte=100"`
}

type RolAsignadoResponse struct {
	TerceroRolID int64  `json:"tercero_rol_id"`
	RolID        int32  `json:"rol_id"`
	Codigo       string `json:"codigo"`
	NombreRol    string `json:"nombre_rol"`
}

type TerceroResponse struct {
	ID                   int64                 `json:"id"`
	EmpresaID            int32                 `json:"empresa_id"`
	TipoIdentificacion   string                `json:"tipo_identificacion"`
	NumeroIdentificacion string                `json:"numero_identificacion"`
	RazonSocial          string                `json:"razon_social"`
	PaisID               *int32                `json:"pais_id"`
	PaisNombre           string                `json:"pais_nombre,omitempty"`
	NaturalezaJuridica   string                `json:"naturaleza_juridica"`
	Email                *string               `json:"email"`
	Telefono             *string               `json:"telefono"`
	Direccion            *string               `json:"direccion"`
	Activo               bool                  `json:"activo"`
	Roles                []RolAsignadoResponse `json:"roles"`
	CreatedAt            time.Time             `json:"created_at"`
	UpdatedAt            time.Time             `json:"updated_at"`
}

type ListaTercerosResponse struct {
	Data     []TerceroResponse `json:"data"`
	Page     int               `json:"page"`
	PageSize int               `json:"page_size"`
	Total    int64             `json:"total"`
}

func toTerceroResponse(t *TerceroMaestro) TerceroResponse {
	roles := make([]RolAsignadoResponse, 0, len(t.Roles))
	for i := range t.Roles {
		tr := &t.Roles[i]
		r := RolAsignadoResponse{
			TerceroRolID: tr.ID,
			RolID:        tr.RolID,
		}
		if tr.Rol != nil {
			r.Codigo = tr.Rol.Codigo
			r.NombreRol = tr.Rol.NombreRol
		}
		roles = append(roles, r)
	}

	resp := TerceroResponse{
		ID:                   t.ID,
		EmpresaID:            t.EmpresaID,
		TipoIdentificacion:   t.TipoIdentificacion,
		NumeroIdentificacion: t.NumeroIdentificacion,
		RazonSocial:          t.RazonSocial,
		PaisID:               t.PaisID,
		NaturalezaJuridica:   t.NaturalezaJuridica,
		Email:                t.Email,
		Telefono:             t.Telefono,
		Direccion:            t.Direccion,
		Activo:               t.Activo,
		Roles:                roles,
		CreatedAt:            t.CreatedAt,
		UpdatedAt:            t.UpdatedAt,
	}
	if t.Pais != nil {
		resp.PaisNombre = t.Pais.Nombre
	}
	return resp
}
