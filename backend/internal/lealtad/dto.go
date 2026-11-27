package lealtad

import (
	"strings"
	"time"
)

type MonedaResponse struct {
	ID           int32  `json:"id"`
	CodigoISO    string `json:"codigo_iso"`
	NombreMoneda string `json:"nombre_moneda"`
	Simbolo      string `json:"simbolo"`
}

type AdhesionResponse struct {
	ID                  int64      `json:"id"`
	ProgramaID          int32      `json:"programa_id"`
	EmpresaID           int32      `json:"empresa_id"`
	EmpresaNombre       string     `json:"empresa_nombre,omitempty"`
	EmpresaNIT          string     `json:"empresa_nit,omitempty"`
	RolEmpresa          string     `json:"rol_empresa"`
	PermiteAcumular     bool       `json:"permite_acumular"`
	PermiteRedimir      bool       `json:"permite_redimir"`
	EsProgramaPrincipal bool       `json:"es_programa_principal"`
	FechaAdhesion       time.Time  `json:"fecha_adhesion"`
	FechaBaja           *time.Time `json:"fecha_baja,omitempty"`
	Activo              bool       `json:"activo"`
}

type ProgramaResponse struct {
	ID                           int32             `json:"id"`
	Codigo                       string            `json:"codigo"`
	Nombre                       string            `json:"nombre"`
	Descripcion                  *string           `json:"descripcion,omitempty"`
	Alcance                      string            `json:"alcance"`
	EmpresaPropietariaID         *int32            `json:"empresa_propietaria_id,omitempty"`
	EmpresaPropietariaNombre     string            `json:"empresa_propietaria_nombre,omitempty"`
	UsuarioAdministradorID       *int64            `json:"usuario_administrador_id,omitempty"`
	UsuarioAdministradorUsername string            `json:"usuario_administrador_username,omitempty"`
	UsuarioAdministradorNombre   string            `json:"usuario_administrador_nombre,omitempty"`
	EsSistema                    bool              `json:"es_sistema"`
	MonedaReferenciaID           int32             `json:"moneda_referencia_id"`
	MonedaCodigo                 string            `json:"moneda_codigo,omitempty"`
	MonedaNombre                 string            `json:"moneda_nombre,omitempty"`
	MonedaSimbolo                string            `json:"moneda_simbolo,omitempty"`
	ValorPuntoRedencion          float64           `json:"valor_punto_redencion"`
	DiasVencimiento              *int32            `json:"dias_vencimiento,omitempty"`
	Estado                       string            `json:"estado"`
	FechaInicio                  time.Time         `json:"fecha_inicio"`
	FechaFin                     *time.Time        `json:"fecha_fin,omitempty"`
	EmpresasAdheridasCount       int64             `json:"empresas_adheridas_count"`
	Adhesion                     *AdhesionResponse `json:"adhesion,omitempty"`
	PuedeEditar                  bool              `json:"puede_editar"`
	PuedeEliminar                bool              `json:"puede_eliminar"`
	PuedeVerEmpresas             bool              `json:"puede_ver_empresas"`
	PuedeAdherir                 bool              `json:"puede_adherir"`
	CreatedAt                    time.Time         `json:"created_at"`
	UpdatedAt                    time.Time         `json:"updated_at"`
}

type ProgramaFilter struct {
	Search        string
	Alcance       string
	Estado        string
	SoloAdheridos bool
}

type CrearProgramaRequest struct {
	Codigo              string     `json:"codigo"                binding:"required,min=2,max=30"`
	Nombre              string     `json:"nombre"                binding:"required,min=2,max=100"`
	Descripcion         *string    `json:"descripcion"           binding:"omitnil,max=1000"`
	Alcance             string     `json:"alcance"               binding:"omitempty,oneof=LOCAL COALICION GLOBAL"`
	MonedaReferenciaID  int32      `json:"moneda_referencia_id"  binding:"required,gt=0"`
	ValorPuntoRedencion float64    `json:"valor_punto_redencion" binding:"required,gt=0"`
	DiasVencimiento     *int32     `json:"dias_vencimiento"      binding:"omitnil,gt=0"`
	Estado              string     `json:"estado"                binding:"omitempty,oneof=ACTIVO INACTIVO PAUSADO"`
	FechaInicio         *time.Time `json:"fecha_inicio"`
	FechaFin            *time.Time `json:"fecha_fin"`
}

type ActualizarProgramaRequest struct {
	Codigo                 *string    `json:"codigo"                   binding:"omitnil,min=2,max=30"`
	Nombre                 *string    `json:"nombre"                   binding:"omitnil,min=2,max=100"`
	Descripcion            *string    `json:"descripcion"              binding:"omitnil,max=1000"`
	MonedaReferenciaID     *int32     `json:"moneda_referencia_id"     binding:"omitnil,gt=0"`
	ValorPuntoRedencion    *float64   `json:"valor_punto_redencion"    binding:"omitnil,gt=0"`
	DiasVencimiento        *int32     `json:"dias_vencimiento"         binding:"omitnil,gt=0"`
	LimpiarDiasVencimiento bool       `json:"limpiar_dias_vencimiento"`
	Estado                 *string    `json:"estado"                   binding:"omitnil,oneof=ACTIVO INACTIVO PAUSADO"`
	FechaInicio            *time.Time `json:"fecha_inicio"`
	FechaFin               *time.Time `json:"fecha_fin"`
	LimpiarFechaFin        bool       `json:"limpiar_fecha_fin"`
}

type AdherirEmpresaRequest struct {
	PermiteAcumular     *bool `json:"permite_acumular"`
	PermiteRedimir      *bool `json:"permite_redimir"`
	EsProgramaPrincipal *bool `json:"es_programa_principal"`
}

type ActualizarAdhesionRequest struct {
	PermiteAcumular     *bool `json:"permite_acumular"`
	PermiteRedimir      *bool `json:"permite_redimir"`
	EsProgramaPrincipal *bool `json:"es_programa_principal"`
}

func toAdhesionResponse(pe ProgramaEmpresa) AdhesionResponse {
	resp := AdhesionResponse{
		ID:                  pe.ID,
		ProgramaID:          pe.ProgramaID,
		EmpresaID:           pe.EmpresaID,
		RolEmpresa:          pe.RolEmpresa,
		PermiteAcumular:     pe.PermiteAcumular,
		PermiteRedimir:      pe.PermiteRedimir,
		EsProgramaPrincipal: pe.EsProgramaPrincipal,
		FechaAdhesion:       pe.FechaAdhesion,
		FechaBaja:           pe.FechaBaja,
		Activo:              pe.Activo,
	}
	if pe.Empresa != nil {
		if strings.TrimSpace(pe.Empresa.RazonSocial) != "" {
			resp.EmpresaNombre = pe.Empresa.RazonSocial
		} else {
			resp.EmpresaNombre = pe.Empresa.NombreLegal
		}
		resp.EmpresaNIT = pe.Empresa.IdentificadorFiscal
	}
	return resp
}

func toProgramaResponse(
	p Programa,
	adhesion *ProgramaEmpresa,
	adheridasCount int64,
	puedeEditar bool,
	puedeEliminar bool,
	puedeVerEmpresas bool,
	puedeAdherir bool,
) ProgramaResponse {
	resp := ProgramaResponse{
		ID:                     p.ID,
		Codigo:                 p.Codigo,
		Nombre:                 p.Nombre,
		Descripcion:            p.Descripcion,
		Alcance:                p.Alcance,
		EmpresaPropietariaID:   p.EmpresaPropietariaID,
		UsuarioAdministradorID: p.UsuarioAdministradorID,
		EsSistema:              p.EsSistema,
		MonedaReferenciaID:     p.MonedaReferenciaID,
		ValorPuntoRedencion:    p.ValorPuntoRedencion,
		DiasVencimiento:        p.DiasVencimiento,
		Estado:                 p.Estado,
		FechaInicio:            p.FechaInicio,
		FechaFin:               p.FechaFin,
		EmpresasAdheridasCount: adheridasCount,
		PuedeEditar:            puedeEditar,
		PuedeEliminar:          puedeEliminar,
		PuedeVerEmpresas:       puedeVerEmpresas,
		PuedeAdherir:           puedeAdherir,
		CreatedAt:              p.CreatedAt,
		UpdatedAt:              p.UpdatedAt,
	}
	if p.Moneda != nil {
		resp.MonedaCodigo = p.Moneda.CodigoISO
		resp.MonedaNombre = p.Moneda.NombreMoneda
		resp.MonedaSimbolo = p.Moneda.Simbolo
	}
	if p.EmpresaPropietaria != nil {
		if strings.TrimSpace(p.EmpresaPropietaria.RazonSocial) != "" {
			resp.EmpresaPropietariaNombre = p.EmpresaPropietaria.RazonSocial
		} else {
			resp.EmpresaPropietariaNombre = p.EmpresaPropietaria.NombreLegal
		}
	}
	if p.UsuarioAdministrador != nil {
		resp.UsuarioAdministradorUsername = p.UsuarioAdministrador.NombreUsuario
		resp.UsuarioAdministradorNombre = strings.TrimSpace(
			p.UsuarioAdministrador.Nombres + " " + p.UsuarioAdministrador.Apellidos,
		)
	}
	if adhesion != nil {
		adh := toAdhesionResponse(*adhesion)
		resp.Adhesion = &adh
	}
	return resp
}
