package compras

import (
	"strings"
	"time"
)

const dateLayout = "2006-01-02"

type CompraDetalleInput struct {
	ArticuloID    int32    `json:"articulo_id"    binding:"required,gt=0"`
	Descripcion   *string  `json:"descripcion"    binding:"omitnil,max=255"`
	Cantidad      float64  `json:"cantidad"       binding:"required,gt=0"`
	CostoUnitario *float64 `json:"costo_unitario" binding:"required,gte=0"`
	Impuesto      *float64 `json:"impuesto"       binding:"omitnil,gte=0"`
}

type CrearCompraRequest struct {
	SucursalID             *int32               `json:"sucursal_id"              binding:"omitnil,gt=0"`
	TerceroEmpresaRolID    *int64               `json:"tercero_empresa_rol_id"   binding:"omitnil,gt=0"`
	ProveedorNombre        *string              `json:"proveedor_nombre"         binding:"omitnil,max=150"`
	Consecutivo            string               `json:"consecutivo"              binding:"omitempty,max=50"`
	NumeroFacturaProveedor *string              `json:"numero_factura_proveedor" binding:"omitnil,max=50"`
	MonedaID               *int32               `json:"moneda_id"                binding:"omitnil,gt=0"`
	Estado                 *string              `json:"estado"                   binding:"omitnil,oneof=BORRADOR REGISTRADA PAGADA ANULADA"`
	FechaCompra            string               `json:"fecha_compra"             binding:"required"`
	FechaVencimiento       *string              `json:"fecha_vencimiento"`
	Observaciones          *string              `json:"observaciones"            binding:"omitnil,max=5000"`
	Detalles               []CompraDetalleInput `json:"detalles"                 binding:"required,min=1,dive"`
}

type ActualizarCompraRequest struct {
	SucursalID              *int32                `json:"sucursal_id"               binding:"omitnil,gt=0"`
	TerceroEmpresaRolID     *int64                `json:"tercero_empresa_rol_id"    binding:"omitnil,gt=0"`
	ProveedorNombre         *string               `json:"proveedor_nombre"          binding:"omitnil,max=150"`
	Consecutivo             *string               `json:"consecutivo"               binding:"omitnil,min=1,max=50"`
	NumeroFacturaProveedor  *string               `json:"numero_factura_proveedor"  binding:"omitnil,max=50"`
	MonedaID                *int32                `json:"moneda_id"                 binding:"omitnil,gt=0"`
	Estado                  *string               `json:"estado"                    binding:"omitnil,oneof=BORRADOR REGISTRADA PAGADA ANULADA"`
	FechaCompra             *string               `json:"fecha_compra"`
	FechaVencimiento        *string               `json:"fecha_vencimiento"`
	LimpiarFechaVencimiento bool                  `json:"limpiar_fecha_vencimiento"`
	Observaciones           *string               `json:"observaciones"             binding:"omitnil,max=5000"`
	Detalles                *[]CompraDetalleInput `json:"detalles"                  binding:"omitnil,min=1,dive"`
}

type ListarComprasQuery struct {
	SucursalID       int32   `form:"sucursal_id" binding:"omitempty,gt=0"`
	SucursalIDsScope []int32 `form:"-"`
	Estado           string  `form:"estado"      binding:"omitempty,oneof=BORRADOR REGISTRADA PAGADA ANULADA"`
	MonedaID   int32  `form:"moneda_id"   binding:"omitempty,gt=0"`
	FechaDesde string `form:"fecha_desde"`
	FechaHasta string `form:"fecha_hasta"`
	Q          string `form:"q"           binding:"omitempty,max=100"`
	Page       int    `form:"page"        binding:"omitempty,gte=1"`
	PageSize   int    `form:"page_size"   binding:"omitempty,gte=1,lte=100"`
}

type CompraDetalleResponse struct {
	ID             int64   `json:"id"`
	CompraID       int64   `json:"compra_id"`
	ArticuloID     int32   `json:"articulo_id"`
	ArticuloCodigo string  `json:"articulo_codigo,omitempty"`
	ArticuloNombre string  `json:"articulo_nombre,omitempty"`
	Descripcion    *string `json:"descripcion"`
	Cantidad       float64 `json:"cantidad"`
	CostoUnitario  float64 `json:"costo_unitario"`
	Subtotal       float64 `json:"subtotal"`
	Impuesto       float64 `json:"impuesto"`
	Total          float64 `json:"total"`
}

type CompraResponse struct {
	ID                     int64                   `json:"id"`
	EmpresaID              int32                   `json:"empresa_id"`
	SucursalID             int32                   `json:"sucursal_id"`
	SucursalNombre         string                  `json:"sucursal_nombre,omitempty"`
	TerceroEmpresaRolID    *int64                  `json:"tercero_empresa_rol_id"`
	ProveedorNombre        *string                 `json:"proveedor_nombre"`
	Consecutivo            string                  `json:"consecutivo"`
	NumeroFacturaProveedor *string                 `json:"numero_factura_proveedor"`
	MonedaID               *int32                  `json:"moneda_id"`
	MonedaCodigo           string                  `json:"moneda_codigo,omitempty"`
	MonedaSimbolo          string                  `json:"moneda_simbolo,omitempty"`
	Estado                 string                  `json:"estado"`
	FechaCompra            time.Time               `json:"fecha_compra"`
	FechaVencimiento       *time.Time              `json:"fecha_vencimiento"`
	Subtotal               float64                 `json:"subtotal"`
	TotalIVA               float64                 `json:"total_iva"`
	Total                  float64                 `json:"total"`
	Observaciones          *string                 `json:"observaciones"`
	UsuarioID              int64                   `json:"usuario_id"`
	UsuarioNombre          string                  `json:"usuario_nombre,omitempty"`
	Detalles               []CompraDetalleResponse `json:"detalles,omitempty"`
	CreatedAt              time.Time               `json:"created_at"`
	UpdatedAt              time.Time               `json:"updated_at"`
}

type ListaComprasResponse struct {
	Data     []CompraResponse `json:"data"`
	Page     int              `json:"page"`
	PageSize int              `json:"page_size"`
	Total    int64            `json:"total"`
}

func toResponse(c *Compra, includeDetalles bool) CompraResponse {
	resp := CompraResponse{
		ID:                     c.ID,
		EmpresaID:              c.EmpresaID,
		SucursalID:             c.SucursalID,
		TerceroEmpresaRolID:    c.TerceroEmpresaRolID,
		ProveedorNombre:        c.ProveedorNombre,
		Consecutivo:            c.Consecutivo,
		NumeroFacturaProveedor: c.NumeroFacturaProveedor,
		MonedaID:               c.MonedaID,
		Estado:                 c.Estado,
		FechaCompra:            c.FechaCompra,
		FechaVencimiento:       c.FechaVencimiento,
		Subtotal:               c.Subtotal,
		TotalIVA:               c.TotalIVA,
		Total:                  c.Total,
		Observaciones:          c.Observaciones,
		UsuarioID:              c.UsuarioID,
		CreatedAt:              c.CreatedAt,
		UpdatedAt:              c.UpdatedAt,
	}
	if c.Sucursal != nil {
		resp.SucursalNombre = c.Sucursal.Nombre
	}
	if c.Moneda != nil {
		resp.MonedaCodigo = strings.TrimSpace(c.Moneda.CodigoISO)
		resp.MonedaSimbolo = c.Moneda.Simbolo
	}
	if c.Usuario != nil {
		resp.UsuarioNombre = strings.TrimSpace(c.Usuario.Nombres + " " + c.Usuario.Apellidos)
	}
	if includeDetalles {
		detalles := make([]CompraDetalleResponse, 0, len(c.Detalles))
		for i := range c.Detalles {
			d := &c.Detalles[i]
			dr := CompraDetalleResponse{
				ID:            d.ID,
				CompraID:      d.CompraID,
				ArticuloID:    d.ArticuloID,
				Descripcion:   d.Descripcion,
				Cantidad:      d.Cantidad,
				CostoUnitario: d.CostoUnitario,
				Subtotal:      d.Subtotal,
				Impuesto:      d.Impuesto,
				Total:         d.Total,
			}
			if d.Articulo != nil {
				dr.ArticuloCodigo = d.Articulo.Codigo
				dr.ArticuloNombre = d.Articulo.Nombre
			}
			detalles = append(detalles, dr)
		}
		resp.Detalles = detalles
	}
	return resp
}
