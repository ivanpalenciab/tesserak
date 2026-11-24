package crm

import "time"

// =====================================================================
// Catálogos CRM: orígenes, estados y motivos de descarte
// =====================================================================

type CrearOrigenRequest struct {
	EmpresaID int32  `json:"empresa_id" binding:"required,gt=0"`
	Nombre    string `json:"nombre"     binding:"required,min=1,max=100"`
	Activo    *bool  `json:"activo"`
}

type ActualizarOrigenRequest struct {
	Nombre *string `json:"nombre" binding:"omitnil,min=1,max=100"`
	Activo *bool   `json:"activo"`
}

type ListarCatalogoQuery struct {
	EmpresaID int32 `form:"empresa_id" binding:"required,gt=0"`
	Activo    *bool `form:"activo"`
}

type OrigenResponse struct {
	ID        int32     `json:"id"`
	EmpresaID int32     `json:"empresa_id"`
	Nombre    string    `json:"nombre"`
	Activo    bool      `json:"activo"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CrearEstadoRequest struct {
	EmpresaID     int32  `json:"empresa_id"      binding:"required,gt=0"`
	Nombre        string `json:"nombre"          binding:"required,min=1,max=50"`
	Orden         *int32 `json:"orden"           binding:"omitnil,gte=0"`
	EsEstadoFinal *bool  `json:"es_estado_final"`
}

type ActualizarEstadoRequest struct {
	Nombre        *string `json:"nombre"          binding:"omitnil,min=1,max=50"`
	Orden         *int32  `json:"orden"           binding:"omitnil,gte=0"`
	EsEstadoFinal *bool   `json:"es_estado_final"`
}

type EstadoResponse struct {
	ID            int32     `json:"id"`
	EmpresaID     int32     `json:"empresa_id"`
	Nombre        string    `json:"nombre"`
	Orden         int32     `json:"orden"`
	EsEstadoFinal bool      `json:"es_estado_final"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type CrearMotivoDescarteRequest struct {
	EmpresaID   int32  `json:"empresa_id"  binding:"required,gt=0"`
	Descripcion string `json:"descripcion" binding:"required,min=1,max=150"`
	Activo      *bool  `json:"activo"`
}

type ActualizarMotivoDescarteRequest struct {
	Descripcion *string `json:"descripcion" binding:"omitnil,min=1,max=150"`
	Activo      *bool   `json:"activo"`
}

type MotivoDescarteResponse struct {
	ID          int32     `json:"id"`
	EmpresaID   int32     `json:"empresa_id"`
	Descripcion string    `json:"descripcion"`
	Activo      bool      `json:"activo"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// =====================================================================
// Leads
// =====================================================================

type CrearLeadRequest struct {
	EmpresaID         int32    `json:"empresa_id"          binding:"required,gt=0"`
	SucursalID        *int32   `json:"sucursal_id"         binding:"omitnil,gt=0"`
	OrigenID          int32    `json:"origen_id"           binding:"required,gt=0"`
	EstadoID          int32    `json:"estado_id"           binding:"required,gt=0"`
	MotivoDescarteID  *int32   `json:"motivo_descarte_id"  binding:"omitnil,gt=0"`
	UsuarioAsignadoID *int64   `json:"usuario_asignado_id" binding:"omitnil,gt=0"`
	Nombre            string   `json:"nombre"              binding:"required,min=1,max=100"`
	Apellido          string   `json:"apellido"            binding:"required,min=1,max=100"`
	NombreEmpresa     *string  `json:"nombre_empresa"      binding:"omitnil,max=150"`
	Cargo             *string  `json:"cargo"               binding:"omitnil,max=100"`
	Email             *string  `json:"email"               binding:"omitnil,email,max=255"`
	Telefono          *string  `json:"telefono"            binding:"omitnil,max=50"`
	UsuarioFacebook   *string  `json:"usuario_facebook"    binding:"omitnil,max=100"`
	Temperatura       *string  `json:"temperatura"         binding:"omitnil,oneof=FRIO TIBIO CALIENTE"`
	Score             *int32   `json:"score"               binding:"omitnil,gte=0,lte=100"`
	ValorEstimado     *float64 `json:"valor_estimado"      binding:"omitnil,gte=0"`
	MonedaID          *int32   `json:"moneda_id"           binding:"omitnil,gt=0"`
	CreatedBy         int64    `json:"created_by"          binding:"required,gt=0"`
}

type ActualizarLeadRequest struct {
	SucursalID        *int32   `json:"sucursal_id"         binding:"omitnil,gt=0"`
	OrigenID          *int32   `json:"origen_id"           binding:"omitnil,gt=0"`
	EstadoID          *int32   `json:"estado_id"           binding:"omitnil,gt=0"`
	MotivoDescarteID  *int32   `json:"motivo_descarte_id"  binding:"omitnil,gt=0"`
	UsuarioAsignadoID *int64   `json:"usuario_asignado_id" binding:"omitnil,gt=0"`
	Nombre            *string  `json:"nombre"              binding:"omitnil,min=1,max=100"`
	Apellido          *string  `json:"apellido"            binding:"omitnil,min=1,max=100"`
	NombreEmpresa     *string  `json:"nombre_empresa"      binding:"omitnil,max=150"`
	Cargo             *string  `json:"cargo"               binding:"omitnil,max=100"`
	Email             *string  `json:"email"               binding:"omitnil,email,max=255"`
	Telefono          *string  `json:"telefono"            binding:"omitnil,max=50"`
	UsuarioFacebook   *string  `json:"usuario_facebook"    binding:"omitnil,max=100"`
	Temperatura       *string  `json:"temperatura"         binding:"omitnil,oneof=FRIO TIBIO CALIENTE"`
	Score             *int32   `json:"score"               binding:"omitnil,gte=0,lte=100"`
	ValorEstimado     *float64 `json:"valor_estimado"      binding:"omitnil,gte=0"`
	MonedaID          *int32   `json:"moneda_id"           binding:"omitnil,gt=0"`
}

// CambiarEstadoLeadRequest mueve el lead a otra etapa del embudo.
type CambiarEstadoLeadRequest struct {
	EstadoID         int32  `json:"estado_id"          binding:"required,gt=0"`
	MotivoDescarteID *int32 `json:"motivo_descarte_id" binding:"omitnil,gt=0"`
}

// ConvertirLeadRequest marca un lead como convertido en cliente/tercero.
type ConvertirLeadRequest struct {
	TerceroEmpresaRolID *int64 `json:"convertido_tercero_empresa_rol_id" binding:"omitnil,gt=0"`
	EstadoID            *int32 `json:"estado_id"                         binding:"omitnil,gt=0"`
}

type ListarLeadsQuery struct {
	EmpresaID         int32  `form:"empresa_id"          binding:"required,gt=0"`
	SucursalID        int32  `form:"sucursal_id"         binding:"omitempty,gt=0"`
	OrigenID          int32  `form:"origen_id"           binding:"omitempty,gt=0"`
	EstadoID          int32  `form:"estado_id"           binding:"omitempty,gt=0"`
	UsuarioAsignadoID int64  `form:"usuario_asignado_id" binding:"omitempty,gt=0"`
	Temperatura       string `form:"temperatura"         binding:"omitempty,oneof=FRIO TIBIO CALIENTE"`
	Convertido        *bool  `form:"convertido"`
	Q                 string `form:"q"                   binding:"omitempty,max=100"`
	Page              int    `form:"page"                binding:"omitempty,gte=1"`
	PageSize          int    `form:"page_size"           binding:"omitempty,gte=1,lte=100"`
}

type LeadResponse struct {
	ID                            int64      `json:"id"`
	EmpresaID                     int32      `json:"empresa_id"`
	SucursalID                    *int32     `json:"sucursal_id"`
	OrigenID                      int32      `json:"origen_id"`
	EstadoID                      int32      `json:"estado_id"`
	MotivoDescarteID              *int32     `json:"motivo_descarte_id"`
	UsuarioAsignadoID             *int64     `json:"usuario_asignado_id"`
	Nombre                        string     `json:"nombre"`
	Apellido                      string     `json:"apellido"`
	NombreEmpresa                 *string    `json:"nombre_empresa"`
	Cargo                         *string    `json:"cargo"`
	Email                         *string    `json:"email"`
	Telefono                      *string    `json:"telefono"`
	UsuarioFacebook               *string    `json:"usuario_facebook"`
	Temperatura                   string     `json:"temperatura"`
	Score                         int32      `json:"score"`
	ValorEstimado                 float64    `json:"valor_estimado"`
	MonedaID                      *int32     `json:"moneda_id"`
	Convertido                    bool       `json:"convertido"`
	ConvertidoTerceroEmpresaRolID *int64     `json:"convertido_tercero_empresa_rol_id"`
	FechaConversion               *time.Time `json:"fecha_conversion"`
	CreatedBy                     int64      `json:"created_by"`
	CreatedAt                     time.Time  `json:"created_at"`
	UpdatedAt                     time.Time  `json:"updated_at"`
}

type ListaLeadsResponse struct {
	Data     []LeadResponse `json:"data"`
	Page     int            `json:"page"`
	PageSize int            `json:"page_size"`
	Total    int64          `json:"total"`
}

// =====================================================================
// Actividades de seguimiento
// =====================================================================

type CrearActividadRequest struct {
	UsuarioID        int64      `json:"usuario_id"        binding:"required,gt=0"`
	TipoActividad    string     `json:"tipo_actividad"    binding:"required,oneof=LLAMADA CORREO REUNION WHATSAPP TAREA NOTA"`
	FechaActividad   *time.Time `json:"fecha_actividad"`
	FechaVencimiento *time.Time `json:"fecha_vencimiento"`
	Estado           *string    `json:"estado"            binding:"omitnil,oneof=PENDIENTE COMPLETADA CANCELADA"`
	Asunto           string     `json:"asunto"            binding:"required,min=1,max=200"`
	Observaciones    *string    `json:"observaciones"     binding:"omitnil,max=5000"`
}

type ActualizarActividadRequest struct {
	TipoActividad    *string    `json:"tipo_actividad"    binding:"omitnil,oneof=LLAMADA CORREO REUNION WHATSAPP TAREA NOTA"`
	FechaActividad   *time.Time `json:"fecha_actividad"`
	FechaVencimiento *time.Time `json:"fecha_vencimiento"`
	Estado           *string    `json:"estado"            binding:"omitnil,oneof=PENDIENTE COMPLETADA CANCELADA"`
	Asunto           *string    `json:"asunto"            binding:"omitnil,min=1,max=200"`
	Observaciones    *string    `json:"observaciones"     binding:"omitnil,max=5000"`
}

type ListarActividadesQuery struct {
	EmpresaID     int32  `form:"empresa_id"     binding:"omitempty,gt=0"`
	UsuarioID     int64  `form:"usuario_id"     binding:"omitempty,gt=0"`
	TipoActividad string `form:"tipo_actividad" binding:"omitempty,oneof=LLAMADA CORREO REUNION WHATSAPP TAREA NOTA"`
	Estado        string `form:"estado"         binding:"omitempty,oneof=PENDIENTE COMPLETADA CANCELADA"`
	SoloVencidas  bool   `form:"solo_vencidas"`
	Page          int    `form:"page"           binding:"omitempty,gte=1"`
	PageSize      int    `form:"page_size"      binding:"omitempty,gte=1,lte=100"`
}

type ActividadResponse struct {
	ID               int64      `json:"id"`
	LeadID           int64      `json:"lead_id"`
	UsuarioID        int64      `json:"usuario_id"`
	TipoActividad    string     `json:"tipo_actividad"`
	FechaActividad   time.Time  `json:"fecha_actividad"`
	FechaVencimiento *time.Time `json:"fecha_vencimiento"`
	Estado           string     `json:"estado"`
	Asunto           string     `json:"asunto"`
	Observaciones    *string    `json:"observaciones"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

type ListaActividadesResponse struct {
	Data     []ActividadResponse `json:"data"`
	Page     int                 `json:"page"`
	PageSize int                 `json:"page_size"`
	Total    int64               `json:"total"`
}

// =====================================================================
// Interés en productos
// =====================================================================

type CrearInteresProductoRequest struct {
	EmpresaProductoID *int32   `json:"empresa_producto_id" binding:"omitnil,gt=0"`
	CategoriaID       *int32   `json:"categoria_id"        binding:"omitnil,gt=0"`
	CantidadEstimada  *float64 `json:"cantidad_estimada"   binding:"omitnil,gt=0"`
	Notas             *string  `json:"notas"               binding:"omitnil,max=255"`
}

type InteresProductoResponse struct {
	ID                int64     `json:"id"`
	LeadID            int64     `json:"lead_id"`
	EmpresaProductoID *int32    `json:"empresa_producto_id"`
	CategoriaID       *int32    `json:"categoria_id"`
	CantidadEstimada  float64   `json:"cantidad_estimada"`
	Notas             *string   `json:"notas"`
	CreatedAt         time.Time `json:"created_at"`
}

// =====================================================================
// Control / Resumen del CRM (Dashboard)
// =====================================================================

type ResumenCRMQuery struct {
	EmpresaID int32 `form:"empresa_id" binding:"required,gt=0"`
}

type EtapaPipelineResumen struct {
	EstadoID      int32   `json:"estado_id"`
	Nombre        string  `json:"nombre"`
	Orden         int32   `json:"orden"`
	EsEstadoFinal bool    `json:"es_estado_final"`
	Cantidad      int64   `json:"cantidad"`
	ValorEstimado float64 `json:"valor_estimado"`
}

type TemperaturaResumen struct {
	Temperatura   string  `json:"temperatura"`
	Cantidad      int64   `json:"cantidad"`
	ValorEstimado float64 `json:"valor_estimado"`
}

type ResumenCRMResponse struct {
	EmpresaID             int32                  `json:"empresa_id"`
	TotalLeads            int64                  `json:"total_leads"`
	LeadsConvertidos      int64                  `json:"leads_convertidos"`
	TasaConversionPct     float64                `json:"tasa_conversion_pct"`
	ValorEstimadoTotal    float64                `json:"valor_estimado_total"`
	ActividadesPendientes int64                  `json:"actividades_pendientes"`
	ActividadesVencidas   int64                  `json:"actividades_vencidas"`
	PorEtapa              []EtapaPipelineResumen `json:"por_etapa"`
	PorTemperatura        []TemperaturaResumen   `json:"por_temperatura"`
}

// =====================================================================
// Mappers
// =====================================================================

func toOrigenResponse(o *OrigenLead) OrigenResponse {
	return OrigenResponse{
		ID:        o.ID,
		EmpresaID: o.EmpresaID,
		Nombre:    o.Nombre,
		Activo:    o.Activo,
		CreatedAt: o.CreatedAt,
		UpdatedAt: o.UpdatedAt,
	}
}

func toEstadoResponse(e *EstadoLead) EstadoResponse {
	return EstadoResponse{
		ID:            e.ID,
		EmpresaID:     e.EmpresaID,
		Nombre:        e.Nombre,
		Orden:         e.Orden,
		EsEstadoFinal: e.EsEstadoFinal,
		CreatedAt:     e.CreatedAt,
		UpdatedAt:     e.UpdatedAt,
	}
}

func toMotivoResponse(m *MotivoDescarte) MotivoDescarteResponse {
	return MotivoDescarteResponse{
		ID:          m.ID,
		EmpresaID:   m.EmpresaID,
		Descripcion: m.Descripcion,
		Activo:      m.Activo,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

func toLeadResponse(l *Lead) LeadResponse {
	return LeadResponse{
		ID:                            l.ID,
		EmpresaID:                     l.EmpresaID,
		SucursalID:                    l.SucursalID,
		OrigenID:                      l.OrigenID,
		EstadoID:                      l.EstadoID,
		MotivoDescarteID:              l.MotivoDescarteID,
		UsuarioAsignadoID:             l.UsuarioAsignadoID,
		Nombre:                        l.Nombre,
		Apellido:                      l.Apellido,
		NombreEmpresa:                 l.NombreEmpresa,
		Cargo:                         l.Cargo,
		Email:                         l.Email,
		Telefono:                      l.Telefono,
		UsuarioFacebook:               l.UsuarioFacebook,
		Temperatura:                   l.Temperatura,
		Score:                         l.Score,
		ValorEstimado:                 l.ValorEstimado,
		MonedaID:                      l.MonedaID,
		Convertido:                    l.Convertido,
		ConvertidoTerceroEmpresaRolID: l.ConvertidoTerceroEmpresaRolID,
		FechaConversion:               l.FechaConversion,
		CreatedBy:                     l.CreatedBy,
		CreatedAt:                     l.CreatedAt,
		UpdatedAt:                     l.UpdatedAt,
	}
}

func toActividadResponse(a *LeadActividad) ActividadResponse {
	return ActividadResponse{
		ID:               a.ID,
		LeadID:           a.LeadID,
		UsuarioID:        a.UsuarioID,
		TipoActividad:    a.TipoActividad,
		FechaActividad:   a.FechaActividad,
		FechaVencimiento: a.FechaVencimiento,
		Estado:           a.Estado,
		Asunto:           a.Asunto,
		Observaciones:    a.Observaciones,
		CreatedAt:        a.CreatedAt,
		UpdatedAt:        a.UpdatedAt,
	}
}

func toInteresResponse(i *LeadInteresProducto) InteresProductoResponse {
	return InteresProductoResponse{
		ID:                i.ID,
		LeadID:            i.LeadID,
		EmpresaProductoID: i.EmpresaProductoID,
		CategoriaID:       i.CategoriaID,
		CantidadEstimada:  i.CantidadEstimada,
		Notas:             i.Notas,
		CreatedAt:         i.CreatedAt,
	}
}
