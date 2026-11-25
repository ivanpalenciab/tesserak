package precios

import (
	"strings"
	"time"
)

const dateLayout = "2006-01-02"

// =====================================================================
// DTOs de listas_precios (Cabecera)
// =====================================================================

type CrearListaPrecioRequest struct {
	Nombre        string  `json:"nombre"         binding:"required,min=1,max=150"`
	MonedaID      int32   `json:"moneda_id"      binding:"required,gt=0"`
	TipoLista     string  `json:"tipo_lista"     binding:"required,oneof=GENERAL MAYORISTA MINORISTA ESPECIAL"`
	VigenciaDesde string  `json:"vigencia_desde" binding:"required"`
	VigenciaHasta *string `json:"vigencia_hasta"`
	Estado        *bool   `json:"estado"`
}

type ActualizarListaPrecioRequest struct {
	Nombre               *string `json:"nombre"                 binding:"omitnil,min=1,max=150"`
	MonedaID             *int32  `json:"moneda_id"              binding:"omitnil,gt=0"`
	TipoLista            *string `json:"tipo_lista"             binding:"omitnil,oneof=GENERAL MAYORISTA MINORISTA ESPECIAL"`
	VigenciaDesde        *string `json:"vigencia_desde"`
	VigenciaHasta        *string `json:"vigencia_hasta"`
	LimpiarVigenciaHasta bool    `json:"limpiar_vigencia_hasta"`
	Estado               *bool   `json:"estado"`
}

type ListarListasPreciosQuery struct {
	TipoLista string `form:"tipo_lista" binding:"omitempty,oneof=GENERAL MAYORISTA MINORISTA ESPECIAL"`
	MonedaID  int32  `form:"moneda_id"  binding:"omitempty,gt=0"`
	Estado    *bool  `form:"estado"`
	Q         string `form:"q"          binding:"omitempty,max=100"`
	Page      int    `form:"page"       binding:"omitempty,gte=1"`
	PageSize  int    `form:"page_size"  binding:"omitempty,gte=1,lte=100"`
}

type ListaPrecioResponse struct {
	ID            int32     `json:"id"`
	EmpresaID     int32     `json:"empresa_id"`
	Nombre        string    `json:"nombre"`
	MonedaID      int32     `json:"moneda_id"`
	MonedaCodigo  string    `json:"moneda_codigo,omitempty"`
	MonedaSimbolo string    `json:"moneda_simbolo,omitempty"`
	TipoLista     string    `json:"tipo_lista"`
	VigenciaDesde string    `json:"vigencia_desde"`
	VigenciaHasta *string   `json:"vigencia_hasta"`
	Estado        bool      `json:"estado"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type ListasPreciosPaginadasResponse struct {
	Data     []ListaPrecioResponse `json:"data"`
	Page     int                   `json:"page"`
	PageSize int                   `json:"page_size"`
	Total    int64                 `json:"total"`
}

// =====================================================================
// DTOs de lista_precios_detalle (Detalle)
// =====================================================================

type CrearDetallePrecioRequest struct {
	ArticuloID     int32    `json:"articulo_id"      binding:"required,gt=0"`
	Precio         *float64 `json:"precio"           binding:"required,gte=0"`
	UnidadMedidaID int32    `json:"unidad_medida_id" binding:"required,gt=0"`
	CantidadMinima *float64 `json:"cantidad_minima"  binding:"omitnil,gt=0"`
	VigenciaDesde  *string  `json:"vigencia_desde"`
	VigenciaHasta  *string  `json:"vigencia_hasta"`
}

type ActualizarDetallePrecioRequest struct {
	ArticuloID           *int32   `json:"articulo_id"            binding:"omitnil,gt=0"`
	Precio               *float64 `json:"precio"                 binding:"omitnil,gte=0"`
	UnidadMedidaID       *int32   `json:"unidad_medida_id"       binding:"omitnil,gt=0"`
	CantidadMinima       *float64 `json:"cantidad_minima"        binding:"omitnil,gt=0"`
	VigenciaDesde        *string  `json:"vigencia_desde"`
	VigenciaHasta        *string  `json:"vigencia_hasta"`
	LimpiarVigenciaDesde bool     `json:"limpiar_vigencia_desde"`
	LimpiarVigenciaHasta bool     `json:"limpiar_vigencia_hasta"`
}

type ListarDetallesPrecioQuery struct {
	ArticuloID     int32  `form:"articulo_id"      binding:"omitempty,gt=0"`
	UnidadMedidaID int32  `form:"unidad_medida_id" binding:"omitempty,gt=0"`
	Q              string `form:"q"                binding:"omitempty,max=100"`
	Page           int    `form:"page"             binding:"omitempty,gte=1"`
	PageSize       int    `form:"page_size"        binding:"omitempty,gte=1,lte=100"`
}

type DetallePrecioResponse struct {
	ID                 int64     `json:"id"`
	ListaPrecioID      int32     `json:"lista_precio_id"`
	ArticuloID         int32     `json:"articulo_id"`
	ArticuloCodigo     string    `json:"articulo_codigo,omitempty"`
	ArticuloNombre     string    `json:"articulo_nombre,omitempty"`
	Precio             float64   `json:"precio"`
	UnidadMedidaID     int32     `json:"unidad_medida_id"`
	UnidadMedidaNombre string    `json:"unidad_medida_nombre,omitempty"`
	CantidadMinima     float64   `json:"cantidad_minima"`
	VigenciaDesde      *string   `json:"vigencia_desde"`
	VigenciaHasta      *string   `json:"vigencia_hasta"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type DetallesPrecioPaginadosResponse struct {
	Data     []DetallePrecioResponse `json:"data"`
	Page     int                     `json:"page"`
	PageSize int                     `json:"page_size"`
	Total    int64                   `json:"total"`
}

func toListaResponse(lp *ListaPrecio) ListaPrecioResponse {
	resp := ListaPrecioResponse{
		ID:            lp.ID,
		EmpresaID:     lp.EmpresaID,
		Nombre:        lp.Nombre,
		MonedaID:      lp.MonedaID,
		TipoLista:     lp.TipoLista,
		VigenciaDesde: lp.VigenciaDesde.Format(dateLayout),
		VigenciaHasta: formatDatePtr(lp.VigenciaHasta),
		Estado:        lp.Estado,
		CreatedAt:     lp.CreatedAt,
		UpdatedAt:     lp.UpdatedAt,
	}
	if lp.Moneda != nil {
		resp.MonedaCodigo = strings.TrimSpace(lp.Moneda.CodigoISO)
		resp.MonedaSimbolo = lp.Moneda.Simbolo
	}
	return resp
}

func toDetalleResponse(d *ListaPrecioDetalle) DetallePrecioResponse {
	resp := DetallePrecioResponse{
		ID:             d.ID,
		ListaPrecioID:  d.ListaPrecioID,
		ArticuloID:     d.ArticuloID,
		Precio:         d.Precio,
		UnidadMedidaID: d.UnidadMedidaID,
		CantidadMinima: d.CantidadMinima,
		VigenciaDesde:  formatDatePtr(d.VigenciaDesde),
		VigenciaHasta:  formatDatePtr(d.VigenciaHasta),
		CreatedAt:      d.CreatedAt,
		UpdatedAt:      d.UpdatedAt,
	}
	if d.Articulo != nil {
		resp.ArticuloCodigo = d.Articulo.Codigo
		resp.ArticuloNombre = d.Articulo.Nombre
	}
	if d.UnidadMedida != nil {
		resp.UnidadMedidaNombre = d.UnidadMedida.Nombre
	}
	return resp
}

func formatDatePtr(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.Format(dateLayout)
	return &s
}
