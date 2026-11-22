package empresas

import "time"

// CrearEmpresaRequest es el cuerpo de POST /empresas.
type CrearEmpresaRequest struct {
	NombreLegal         string  `json:"nombre_legal"         binding:"required,min=1,max=255"`
	RazonSocial         string  `json:"razon_social"         binding:"required,min=1,max=255"`
	IdentificadorFiscal string  `json:"identificador_fiscal" binding:"required,min=1,max=50"`
	Email               *string `json:"email"                binding:"omitnil,email,max=255"`
	Direccion           *string `json:"direccion"            binding:"omitnil,max=255"`
	Logo                *string `json:"logo"                 binding:"omitnil,max=500"`
	Activo              *bool   `json:"activo"`
}

// ActualizarEmpresaRequest es el cuerpo de PATCH /empresas/:id.
// Solo se actualizan los campos enviados (no nulos).
type ActualizarEmpresaRequest struct {
	NombreLegal         *string `json:"nombre_legal"         binding:"omitnil,min=1,max=255"`
	RazonSocial         *string `json:"razon_social"         binding:"omitnil,min=1,max=255"`
	IdentificadorFiscal *string `json:"identificador_fiscal" binding:"omitnil,min=1,max=50"`
	Email               *string `json:"email"                binding:"omitnil,email,max=255"`
	Direccion           *string `json:"direccion"            binding:"omitnil,max=255"`
	Logo                *string `json:"logo"                 binding:"omitnil,max=500"`
	Activo              *bool   `json:"activo"`
}

// ListarEmpresasQuery son los filtros de GET /empresas.
type ListarEmpresasQuery struct {
	Activo   *bool  `form:"activo"`
	Q        string `form:"q"         binding:"omitempty,max=100"`
	Page     int    `form:"page"      binding:"omitempty,gte=1"`
	PageSize int    `form:"page_size" binding:"omitempty,gte=1,lte=100"`
}

// EmpresaResponse es la representación pública de una empresa.
type EmpresaResponse struct {
	ID                  int32     `json:"id"`
	NombreLegal         string    `json:"nombre_legal"`
	RazonSocial         string    `json:"razon_social"`
	IdentificadorFiscal string    `json:"identificador_fiscal"`
	Email               *string   `json:"email"`
	Direccion           *string   `json:"direccion"`
	Logo                *string   `json:"logo"`
	Activo              bool      `json:"activo"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

type ListaEmpresasResponse struct {
	Data     []EmpresaResponse `json:"data"`
	Page     int               `json:"page"`
	PageSize int               `json:"page_size"`
	Total    int64             `json:"total"`
}

func toResponse(e *Empresa) EmpresaResponse {
	return EmpresaResponse{
		ID:                  e.ID,
		NombreLegal:         e.NombreLegal,
		RazonSocial:         e.RazonSocial,
		IdentificadorFiscal: e.IdentificadorFiscal,
		Email:               e.Email,
		Direccion:           e.Direccion,
		Logo:                e.Logo,
		Activo:              e.Activo,
		CreatedAt:           e.CreatedAt,
		UpdatedAt:           e.UpdatedAt,
	}
}
