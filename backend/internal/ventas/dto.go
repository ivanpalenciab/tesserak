package ventas

import (
	"strings"
	"time"
)

const dateLayout = "2006-01-02"

type VentaDetalleInput struct {
	ArticuloID     int32    `json:"articulo_id"     binding:"required,gt=0"`
	Descripcion    *string  `json:"descripcion"     binding:"omitnil,max=255"`
	Cantidad       float64  `json:"cantidad"        binding:"required,gt=0"`
	PrecioUnitario *float64 `json:"precio_unitario" binding:"required,gte=0"`
	Impuesto       *float64 `json:"impuesto"        binding:"omitnil,gte=0"`
}

type CrearVentaRequest struct {
	SucursalID       *int32              `json:"sucursal_id"       binding:"omitnil,gt=0"`
	UbicacionID      *int32              `json:"ubicacion_id"      binding:"omitnil,gt=0"`
	ClienteID        *int64              `json:"cliente_id"        binding:"omitnil,gt=0"`
	Consecutivo      string              `json:"consecutivo"       binding:"omitempty,max=50"`
	MonedaID         *int32              `json:"moneda_id"         binding:"omitnil,gt=0"`
	Estado           *string             `json:"estado"            binding:"omitnil,oneof=BORRADOR REGISTRADA PAGADA ANULADA"`
	FechaVenta       string              `json:"fecha_venta"       binding:"required"`
	FechaVencimiento *string             `json:"fecha_vencimiento"`
	Observaciones    *string             `json:"observaciones"     binding:"omitnil,max=5000"`
	Detalles         []VentaDetalleInput `json:"detalles"          binding:"required,min=1,dive"`
}

type ActualizarVentaRequest struct {
	SucursalID              *int32               `json:"sucursal_id"               binding:"omitnil,gt=0"`
	UbicacionID             *int32               `json:"ubicacion_id"              binding:"omitnil,gt=0"`
	ClienteID               *int64               `json:"cliente_id"                binding:"omitnil,gt=0"`
	Consecutivo             *string              `json:"consecutivo"               binding:"omitnil,min=1,max=50"`
	MonedaID                *int32               `json:"moneda_id"                 binding:"omitnil,gt=0"`
	Estado                  *string              `json:"estado"                    binding:"omitnil,oneof=BORRADOR REGISTRADA PAGADA ANULADA"`
	FechaVenta              *string              `json:"fecha_venta"`
	FechaVencimiento        *string              `json:"fecha_vencimiento"`
	LimpiarFechaVencimiento bool                 `json:"limpiar_fecha_vencimiento"`
	Observaciones           *string              `json:"observaciones"             binding:"omitnil,max=5000"`
	Detalles                *[]VentaDetalleInput `json:"detalles"                  binding:"omitnil,min=1,dive"`
}

type ListarVentasQuery struct {
	SucursalID       int32   `form:"sucursal_id"  binding:"omitempty,gt=0"`
	SucursalIDsScope []int32 `form:"-"`
	UbicacionID      int32   `form:"ubicacion_id" binding:"omitempty,gt=0"`
	ClienteID   int64  `form:"cliente_id"   binding:"omitempty,gt=0"`
	Estado      string `form:"estado"       binding:"omitempty,oneof=BORRADOR REGISTRADA PAGADA ANULADA"`
	FechaDesde  string `form:"fecha_desde"`
	FechaHasta  string `form:"fecha_hasta"`
	Q           string `form:"q"            binding:"omitempty,max=100"`
	Page        int    `form:"page"         binding:"omitempty,gte=1"`
	PageSize    int    `form:"page_size"    binding:"omitempty,gte=1,lte=100"`
}

type VentaDetalleResponse struct {
	ID             int64   `json:"id"`
	ArticuloID     int32   `json:"articulo_id"`
	ArticuloCodigo string  `json:"articulo_codigo,omitempty"`
	ArticuloNombre string  `json:"articulo_nombre,omitempty"`
	Descripcion    *string `json:"descripcion"`
	Cantidad       float64 `json:"cantidad"`
	PrecioUnitario float64 `json:"precio_unitario"`
	Subtotal       float64 `json:"subtotal"`
	Impuesto       float64 `json:"impuesto"`
	Total          float64 `json:"total"`
}

type VentaResponse struct {
	ID                   int64                  `json:"id"`
	EmpresaID            int32                  `json:"empresa_id"`
	SucursalID           int32                  `json:"sucursal_id"`
	SucursalNombre       string                 `json:"sucursal_nombre,omitempty"`
	UbicacionID          int32                  `json:"ubicacion_id"`
	UbicacionNombre      string                 `json:"ubicacion_nombre,omitempty"`
	ClienteID            int64                  `json:"cliente_id"`
	ClienteNombre        string                 `json:"cliente_nombre,omitempty"`
	ClienteIdentificacion string                `json:"cliente_identificacion,omitempty"`
	Consecutivo          string                 `json:"consecutivo"`
	MonedaID             *int32                 `json:"moneda_id"`
	MonedaCodigo         string                 `json:"moneda_codigo,omitempty"`
	Estado               string                 `json:"estado"`
	FechaVenta           time.Time              `json:"fecha_venta"`
	FechaVencimiento     *time.Time             `json:"fecha_vencimiento"`
	Subtotal             float64                `json:"subtotal"`
	TotalIVA             float64                `json:"total_iva"`
	Total                float64                `json:"total"`
	Observaciones        *string                `json:"observaciones"`
	UsuarioID            int64                  `json:"usuario_id"`
	UsuarioNombre        string                 `json:"usuario_nombre,omitempty"`
	Detalles             []VentaDetalleResponse `json:"detalles,omitempty"`
	CreatedAt            time.Time              `json:"created_at"`
	UpdatedAt            time.Time              `json:"updated_at"`
}

type ListaVentasResponse struct {
	Data     []VentaResponse `json:"data"`
	Page     int             `json:"page"`
	PageSize int             `json:"page_size"`
	Total    int64           `json:"total"`
}

func toResponse(v *Venta, includeDetalles bool) VentaResponse {
	resp := VentaResponse{
		ID:               v.ID,
		EmpresaID:        v.EmpresaID,
		SucursalID:       v.SucursalID,
		UbicacionID:      v.UbicacionID,
		ClienteID:        v.ClienteID,
		Consecutivo:      v.Consecutivo,
		MonedaID:         v.MonedaID,
		Estado:           v.Estado,
		FechaVenta:       v.FechaVenta,
		FechaVencimiento: v.FechaVencimiento,
		Subtotal:         v.Subtotal,
		TotalIVA:         v.TotalIVA,
		Total:            v.Total,
		Observaciones:    v.Observaciones,
		UsuarioID:        v.UsuarioID,
		CreatedAt:        v.CreatedAt,
		UpdatedAt:        v.UpdatedAt,
	}
	if v.Sucursal != nil {
		resp.SucursalNombre = v.Sucursal.Nombre
	}
	if v.Ubicacion != nil {
		resp.UbicacionNombre = v.Ubicacion.Nombre
	}
	if v.Cliente != nil {
		resp.ClienteNombre = v.Cliente.RazonSocial
		resp.ClienteIdentificacion = strings.TrimSpace(v.Cliente.TipoIdentificacion + " " + v.Cliente.NumeroIdentificacion)
	}
	if v.Moneda != nil {
		resp.MonedaCodigo = v.Moneda.Codigo
	}
	if v.Usuario != nil {
		resp.UsuarioNombre = strings.TrimSpace(v.Usuario.Nombres + " " + v.Usuario.Apellidos)
	}
	if includeDetalles {
		detalles := make([]VentaDetalleResponse, 0, len(v.Detalles))
		for i := range v.Detalles {
			d := &v.Detalles[i]
			dr := VentaDetalleResponse{
				ID:             d.ID,
				ArticuloID:     d.ArticuloID,
				Descripcion:    d.Descripcion,
				Cantidad:       d.Cantidad,
				PrecioUnitario: d.PrecioUnitario,
				Subtotal:       d.Subtotal,
				Impuesto:       d.Impuesto,
				Total:          d.Total,
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
