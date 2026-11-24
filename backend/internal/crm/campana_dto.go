package crm

import "time"

const dateLayout = "2006-01-02"

// =====================================================================
// Tipos de campaña
// =====================================================================

type CrearTipoCampanaRequest struct {
	Nombre      string  `json:"nombre"      binding:"required,min=1,max=80"`
	Descripcion *string `json:"descripcion" binding:"omitnil,max=255"`
	Activo      *bool   `json:"activo"`
}

type ActualizarTipoCampanaRequest struct {
	Nombre      *string `json:"nombre"      binding:"omitnil,min=1,max=80"`
	Descripcion *string `json:"descripcion" binding:"omitnil,max=255"`
	Activo      *bool   `json:"activo"`
}

type ListarTiposCampanaQuery struct {
	Activo *bool `form:"activo"`
}

type TipoCampanaResponse struct {
	ID          int32     `json:"id"`
	EmpresaID   int32     `json:"empresa_id"`
	Nombre      string    `json:"nombre"`
	Descripcion *string   `json:"descripcion"`
	Activo      bool      `json:"activo"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// =====================================================================
// Campañas
// =====================================================================

type CrearCampanaRequest struct {
	TipoCampanaID int32    `json:"tipo_campana_id" binding:"required,gt=0"`
	OrigenID      *int32   `json:"origen_id"       binding:"omitnil,gt=0"`
	MonedaID      *int32   `json:"moneda_id"       binding:"omitnil,gt=0"`
	Nombre        string   `json:"nombre"          binding:"required,min=1,max=150"`
	Descripcion   *string  `json:"descripcion"     binding:"omitnil,max=5000"`
	Objetivo      *string  `json:"objetivo"        binding:"omitnil,max=5000"`
	FechaInicio   string   `json:"fecha_inicio"    binding:"required"`
	FechaFin      *string  `json:"fecha_fin"`
	Presupuesto   *float64 `json:"presupuesto"     binding:"omitnil,gte=0"`
	Estado        *string  `json:"estado"          binding:"omitnil,oneof=BORRADOR PROGRAMADA ACTIVA PAUSADA FINALIZADA"`
}

type ActualizarCampanaRequest struct {
	TipoCampanaID *int32   `json:"tipo_campana_id" binding:"omitnil,gt=0"`
	OrigenID      *int32   `json:"origen_id"       binding:"omitnil,gt=0"`
	MonedaID      *int32   `json:"moneda_id"       binding:"omitnil,gt=0"`
	Nombre        *string  `json:"nombre"          binding:"omitnil,min=1,max=150"`
	Descripcion   *string  `json:"descripcion"     binding:"omitnil,max=5000"`
	Objetivo      *string  `json:"objetivo"        binding:"omitnil,max=5000"`
	FechaInicio   *string  `json:"fecha_inicio"`
	FechaFin      *string  `json:"fecha_fin"`
	Presupuesto   *float64 `json:"presupuesto"     binding:"omitnil,gte=0"`
}

type CambiarEstadoCampanaRequest struct {
	Estado string `json:"estado" binding:"required,oneof=BORRADOR PROGRAMADA ACTIVA PAUSADA FINALIZADA"`
}

type ListarCampanasQuery struct {
	TipoCampanaID int32  `form:"tipo_campana_id" binding:"omitempty,gt=0"`
	OrigenID      int32  `form:"origen_id"       binding:"omitempty,gt=0"`
	Estado        string `form:"estado"          binding:"omitempty,oneof=BORRADOR PROGRAMADA ACTIVA PAUSADA FINALIZADA"`
	FechaDesde    string `form:"fecha_desde"`
	FechaHasta    string `form:"fecha_hasta"`
	Q             string `form:"q"               binding:"omitempty,max=100"`
	Page          int    `form:"page"            binding:"omitempty,gte=1"`
	PageSize      int    `form:"page_size"       binding:"omitempty,gte=1,lte=100"`
}

type CampanaResponse struct {
	ID            int64     `json:"id"`
	EmpresaID     int32     `json:"empresa_id"`
	TipoCampanaID int32     `json:"tipo_campana_id"`
	OrigenID      *int32    `json:"origen_id"`
	MonedaID      *int32    `json:"moneda_id"`
	Nombre        string    `json:"nombre"`
	Descripcion   *string   `json:"descripcion"`
	Objetivo      *string   `json:"objetivo"`
	FechaInicio   string    `json:"fecha_inicio"`
	FechaFin      *string   `json:"fecha_fin"`
	Presupuesto   *float64  `json:"presupuesto"`
	Estado        string    `json:"estado"`
	CreatedBy     int64     `json:"created_by"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type ListaCampanasResponse struct {
	Data     []CampanaResponse `json:"data"`
	Page     int               `json:"page"`
	PageSize int               `json:"page_size"`
	Total    int64             `json:"total"`
}

type GastoPorMonedaResumen struct {
	MonedaID       int32   `json:"moneda_id"`
	TotalGastado   float64 `json:"total_gastado"`
	CantidadGastos int64   `json:"cantidad_gastos"`
}

type IndicadoresCampanaResponse struct {
	CampanaID             int64                   `json:"campana_id"`
	EmpresaID             int32                   `json:"empresa_id"`
	Nombre                string                  `json:"nombre"`
	Estado                string                  `json:"estado"`
	MonedaID              *int32                  `json:"moneda_id"`
	Presupuesto           *float64                `json:"presupuesto"`
	TotalLeads            int64                   `json:"total_leads"`
	GastoAcumulado        *float64                `json:"gasto_acumulado"`
	PresupuestoDisponible *float64                `json:"presupuesto_disponible"`
	GastosPorMoneda       []GastoPorMonedaResumen `json:"gastos_por_moneda"`
}

// =====================================================================
// Asociación Campaña - Lead
// =====================================================================

type AsociarLeadCampanaRequest struct {
	LeadID             int64      `json:"lead_id"              binding:"required,gt=0"`
	FechaAsociacion    *time.Time `json:"fecha_asociacion"`
	TipoAtribucion     *string    `json:"tipo_atribucion"      binding:"omitnil,oneof=first_touch last_touch influencia"`
	UTMSource          *string    `json:"utm_source"           binding:"omitnil,max=255"`
	UTMMedium          *string    `json:"utm_medium"           binding:"omitnil,max=255"`
	UTMCampaign        *string    `json:"utm_campaign"         binding:"omitnil,max=255"`
	UTMContent         *string    `json:"utm_content"          binding:"omitnil,max=255"`
	UTMTerm            *string    `json:"utm_term"             binding:"omitnil,max=255"`
	LandingPage        *string    `json:"landing_page"         binding:"omitnil,max=2000"`
	ReferrerURL        *string    `json:"referrer_url"         binding:"omitnil,max=2000"`
	ExternalCampaignID *string    `json:"external_campaign_id" binding:"omitnil,max=150"`
	ExternalAdID       *string    `json:"external_ad_id"       binding:"omitnil,max=150"`
}

type ListarLeadsCampanaQuery struct {
	TipoAtribucion string `form:"tipo_atribucion" binding:"omitempty,oneof=first_touch last_touch influencia"`
	Page           int    `form:"page"            binding:"omitempty,gte=1"`
	PageSize       int    `form:"page_size"       binding:"omitempty,gte=1,lte=100"`
}

type CampanaLeadResponse struct {
	ID                 int64     `json:"id"`
	EmpresaID          int32     `json:"empresa_id"`
	CampanaID          int64     `json:"campana_id"`
	LeadID             int64     `json:"lead_id"`
	FechaAsociacion    time.Time `json:"fecha_asociacion"`
	TipoAtribucion     *string   `json:"tipo_atribucion"`
	UTMSource          *string   `json:"utm_source"`
	UTMMedium          *string   `json:"utm_medium"`
	UTMCampaign        *string   `json:"utm_campaign"`
	UTMContent         *string   `json:"utm_content"`
	UTMTerm            *string   `json:"utm_term"`
	LandingPage        *string   `json:"landing_page"`
	ReferrerURL        *string   `json:"referrer_url"`
	ExternalCampaignID *string   `json:"external_campaign_id"`
	ExternalAdID       *string   `json:"external_ad_id"`
	CreatedAt          time.Time `json:"created_at"`
}

type LeadAsociadoCampanaResponse struct {
	CampanaLeadResponse
	LeadNombre        string  `json:"lead_nombre"`
	LeadApellido      string  `json:"lead_apellido"`
	LeadEmail         *string `json:"lead_email"`
	LeadTelefono      *string `json:"lead_telefono"`
	LeadNombreEmpresa *string `json:"lead_nombre_empresa"`
	LeadEstadoID      int32   `json:"lead_estado_id"`
	LeadConvertido    bool    `json:"lead_convertido"`
}

type ListaLeadsCampanaResponse struct {
	Data     []LeadAsociadoCampanaResponse `json:"data"`
	Page     int                           `json:"page"`
	PageSize int                           `json:"page_size"`
	Total    int64                         `json:"total"`
}

type CampanaAsociadaLeadResponse struct {
	CampanaLeadResponse
	CampanaNombre        string `json:"campana_nombre"`
	CampanaEstado        string `json:"campana_estado"`
	CampanaTipoID        int32  `json:"campana_tipo_id"`
	CampanaFechaInicio   string `json:"campana_fecha_inicio"`
}

// =====================================================================
// Gastos de campaña
// =====================================================================

type CrearGastoCampanaRequest struct {
	Fecha                        string  `json:"fecha"                            binding:"required"`
	Concepto                     string  `json:"concepto"                         binding:"required,min=1,max=150"`
	CategoriaGasto               *string `json:"categoria_gasto"                  binding:"omitnil,max=50"`
	Valor                        float64 `json:"valor"                            binding:"gte=0"`
	MonedaID                     int32   `json:"moneda_id"                        binding:"required,gt=0"`
	ProveedorTerceroEmpresaRolID *int64  `json:"proveedor_tercero_empresa_rol_id" binding:"omitnil,gt=0"`
}

type ActualizarGastoCampanaRequest struct {
	Fecha                        *string  `json:"fecha"`
	Concepto                     *string  `json:"concepto"                         binding:"omitnil,min=1,max=150"`
	CategoriaGasto               *string  `json:"categoria_gasto"                  binding:"omitnil,max=50"`
	Valor                        *float64 `json:"valor"                            binding:"omitnil,gte=0"`
	MonedaID                     *int32   `json:"moneda_id"                        binding:"omitnil,gt=0"`
	ProveedorTerceroEmpresaRolID *int64   `json:"proveedor_tercero_empresa_rol_id" binding:"omitnil,gt=0"`
}

type ListarGastosCampanaQuery struct {
	FechaDesde     string `form:"fecha_desde"`
	FechaHasta     string `form:"fecha_hasta"`
	CategoriaGasto string `form:"categoria_gasto" binding:"omitempty,max=50"`
	Page           int    `form:"page"            binding:"omitempty,gte=1"`
	PageSize       int    `form:"page_size"       binding:"omitempty,gte=1,lte=100"`
}

type GastoCampanaResponse struct {
	ID                           int64     `json:"id"`
	EmpresaID                    int32     `json:"empresa_id"`
	CampanaID                    int64     `json:"campana_id"`
	Fecha                        string    `json:"fecha"`
	Concepto                     string    `json:"concepto"`
	CategoriaGasto               *string   `json:"categoria_gasto"`
	Valor                        float64   `json:"valor"`
	MonedaID                     int32     `json:"moneda_id"`
	ProveedorTerceroEmpresaRolID *int64    `json:"proveedor_tercero_empresa_rol_id"`
	CreatedBy                    int64     `json:"created_by"`
	CreatedAt                    time.Time `json:"created_at"`
	UpdatedAt                    time.Time `json:"updated_at"`
}

type ListaGastosCampanaResponse struct {
	Data     []GastoCampanaResponse `json:"data"`
	Page     int                    `json:"page"`
	PageSize int                    `json:"page_size"`
	Total    int64                  `json:"total"`
}

// =====================================================================
// Mappers
// =====================================================================

func toTipoCampanaResponse(t *TipoCampana) TipoCampanaResponse {
	return TipoCampanaResponse{
		ID:          t.ID,
		EmpresaID:   t.EmpresaID,
		Nombre:      t.Nombre,
		Descripcion: t.Descripcion,
		Activo:      t.Activo,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
}

func toCampanaResponse(c *Campana) CampanaResponse {
	var fechaFin *string
	if c.FechaFin != nil {
		s := c.FechaFin.Format(dateLayout)
		fechaFin = &s
	}
	return CampanaResponse{
		ID:            c.ID,
		EmpresaID:     c.EmpresaID,
		TipoCampanaID: c.TipoCampanaID,
		OrigenID:      c.OrigenID,
		MonedaID:      c.MonedaID,
		Nombre:        c.Nombre,
		Descripcion:   c.Descripcion,
		Objetivo:      c.Objetivo,
		FechaInicio:   c.FechaInicio.Format(dateLayout),
		FechaFin:      fechaFin,
		Presupuesto:   c.Presupuesto,
		Estado:        c.Estado,
		CreatedBy:     c.CreatedBy,
		CreatedAt:     c.CreatedAt,
		UpdatedAt:     c.UpdatedAt,
	}
}

func toCampanaLeadResponse(cl *CampanaLead) CampanaLeadResponse {
	return CampanaLeadResponse{
		ID:                 cl.ID,
		EmpresaID:          cl.EmpresaID,
		CampanaID:          cl.CampanaID,
		LeadID:             cl.LeadID,
		FechaAsociacion:    cl.FechaAsociacion,
		TipoAtribucion:     cl.TipoAtribucion,
		UTMSource:          cl.UTMSource,
		UTMMedium:          cl.UTMMedium,
		UTMCampaign:        cl.UTMCampaign,
		UTMContent:         cl.UTMContent,
		UTMTerm:            cl.UTMTerm,
		LandingPage:        cl.LandingPage,
		ReferrerURL:        cl.ReferrerURL,
		ExternalCampaignID: cl.ExternalCampaignID,
		ExternalAdID:       cl.ExternalAdID,
		CreatedAt:          cl.CreatedAt,
	}
}

func toGastoCampanaResponse(g *CampanaGasto) GastoCampanaResponse {
	return GastoCampanaResponse{
		ID:                           g.ID,
		EmpresaID:                    g.EmpresaID,
		CampanaID:                    g.CampanaID,
		Fecha:                        g.Fecha.Format(dateLayout),
		Concepto:                     g.Concepto,
		CategoriaGasto:               g.CategoriaGasto,
		Valor:                        g.Valor,
		MonedaID:                     g.MonedaID,
		ProveedorTerceroEmpresaRolID: g.ProveedorTerceroEmpresaRolID,
		CreatedBy:                    g.CreatedBy,
		CreatedAt:                    g.CreatedAt,
		UpdatedAt:                    g.UpdatedAt,
	}
}
