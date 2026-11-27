package inventario

import (
	"strings"
	"time"
)

const dateLayout = "2006-01-02"

// =====================================================================
// Ubicaciones
// =====================================================================

type CrearUbicacionRequest struct {
	SucursalID  *int32  `json:"sucursal_id" binding:"omitnil,gt=0"`
	Codigo      string  `json:"codigo"      binding:"required,min=1,max=40"`
	Nombre      string  `json:"nombre"      binding:"required,min=1,max=120"`
	Descripcion *string `json:"descripcion" binding:"omitnil,max=255"`
	Activo      *bool   `json:"activo"`
}

type SucursalResponse struct {
	ID     int32  `json:"id"`
	Nombre string `json:"nombre"`
	Activo bool   `json:"activo"`
}

type ActualizarUbicacionRequest struct {
	SucursalID  *int32  `json:"sucursal_id" binding:"omitnil,gt=0"`
	Codigo      *string `json:"codigo"      binding:"omitnil,min=1,max=40"`
	Nombre      *string `json:"nombre"      binding:"omitnil,min=1,max=120"`
	Descripcion *string `json:"descripcion" binding:"omitnil,max=255"`
	Activo      *bool   `json:"activo"`
}

type ListarUbicacionesQuery struct {
	SucursalID int32  `form:"sucursal_id" binding:"omitempty,gt=0"`
	Activo     *bool  `form:"activo"`
	Q          string `form:"q"           binding:"omitempty,max=100"`
}

type UbicacionResponse struct {
	ID             int32     `json:"id"`
	EmpresaID      int32     `json:"empresa_id"`
	SucursalID     int32     `json:"sucursal_id"`
	SucursalNombre string    `json:"sucursal_nombre,omitempty"`
	Codigo         string    `json:"codigo"`
	Nombre         string    `json:"nombre"`
	Descripcion    *string   `json:"descripcion"`
	Activo         bool      `json:"activo"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// =====================================================================
// Tipos de movimiento
// =====================================================================

type TipoMovimientoResponse struct {
	ID          int32   `json:"id"`
	Codigo      string  `json:"codigo"`
	Nombre      string  `json:"nombre"`
	Naturaleza  string  `json:"naturaleza"`
	Descripcion *string `json:"descripcion"`
	Activo      bool    `json:"activo"`
}

// =====================================================================
// Stock por ubicación
// =====================================================================

type ListarStockQuery struct {
	UbicacionID      int32   `form:"ubicacion_id"   binding:"omitempty,gt=0"`
	ArticuloID       int32   `form:"articulo_id"    binding:"omitempty,gt=0"`
	SucursalIDsScope []int32 `form:"-"`
	SoloConStock     *bool   `form:"solo_con_stock"`
	Q                string  `form:"q"              binding:"omitempty,max=100"`
	Page             int     `form:"page"           binding:"omitempty,gte=1"`
	PageSize         int     `form:"page_size"      binding:"omitempty,gte=1,lte=100"`
}

type StockItemResponse struct {
	ID              int64     `json:"id"`
	EmpresaID       int32     `json:"empresa_id"`
	ArticuloID      int32     `json:"articulo_id"`
	ArticuloCodigo  string    `json:"articulo_codigo,omitempty"`
	ArticuloNombre  string    `json:"articulo_nombre,omitempty"`
	UbicacionID     int32     `json:"ubicacion_id"`
	UbicacionCodigo string    `json:"ubicacion_codigo,omitempty"`
	UbicacionNombre string    `json:"ubicacion_nombre,omitempty"`
	Cantidad        float64   `json:"cantidad"`
	CostoPromedio   float64   `json:"costo_promedio"`
	ValorTotal      float64   `json:"valor_total"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type ListaStockResponse struct {
	Data     []StockItemResponse `json:"data"`
	Page     int                 `json:"page"`
	PageSize int                 `json:"page_size"`
	Total    int64               `json:"total"`
}

// =====================================================================
// Documentos y movimientos de inventario (ajustes, pérdidas, averías, devoluciones, traslados)
// =====================================================================

type MovimientoLineaInput struct {
	ArticuloID         int32    `json:"articulo_id"          binding:"required,gt=0"`
	UbicacionOrigenID  *int32   `json:"ubicacion_origen_id"  binding:"omitnil,gt=0"`
	UbicacionDestinoID *int32   `json:"ubicacion_destino_id" binding:"omitnil,gt=0"`
	Cantidad           float64  `json:"cantidad"             binding:"required,gt=0"`
	CostoUnitario      *float64 `json:"costo_unitario"       binding:"omitnil,gte=0"`
}

type CrearMovimientoRequest struct {
	SucursalID          *int32                 `json:"sucursal_id"           binding:"omitnil,gt=0"`
	TipoMovimientoID    int32                  `json:"tipo_movimiento_id"    binding:"required,gt=0"`
	Fecha               *string                `json:"fecha"`
	TerceroID           *int64                 `json:"tercero_id"            binding:"omitnil,gt=0"`
	Motivo              string                 `json:"motivo"                binding:"required,min=2,max=255"`
	Observaciones       *string                `json:"observaciones"         binding:"omitnil,max=5000"`
	DocumentoOrigenTipo *string                `json:"documento_origen_tipo" binding:"omitnil,max=40"`
	DocumentoOrigenID   *int64                 `json:"documento_origen_id"   binding:"omitnil,gt=0"`
	Detalles            []MovimientoLineaInput `json:"detalles"              binding:"required,min=1,dive"`
}

type ListarMovimientosQuery struct {
	TipoMovimientoID int32   `form:"tipo_movimiento_id" binding:"omitempty,gt=0"`
	UbicacionID      int32   `form:"ubicacion_id"       binding:"omitempty,gt=0"`
	ArticuloID       int32   `form:"articulo_id"        binding:"omitempty,gt=0"`
	SucursalIDsScope []int32 `form:"-"`
	FechaDesde       string  `form:"fecha_desde"`
	FechaHasta       string  `form:"fecha_hasta"`
	Q                string  `form:"q"                  binding:"omitempty,max=100"`
	Page             int     `form:"page"               binding:"omitempty,gte=1"`
	PageSize         int     `form:"page_size"          binding:"omitempty,gte=1,lte=100"`
}

type MovimientoDetalleResponse struct {
	ID                     int64   `json:"id"`
	ArticuloID             int32   `json:"articulo_id"`
	ArticuloCodigo         string  `json:"articulo_codigo,omitempty"`
	ArticuloNombre         string  `json:"articulo_nombre,omitempty"`
	UbicacionOrigenID      *int32  `json:"ubicacion_origen_id"`
	UbicacionOrigenNombre  string  `json:"ubicacion_origen_nombre,omitempty"`
	UbicacionDestinoID     *int32  `json:"ubicacion_destino_id"`
	UbicacionDestinoNombre string  `json:"ubicacion_destino_nombre,omitempty"`
	Cantidad               float64 `json:"cantidad"`
	CostoUnitario          float64 `json:"costo_unitario"`
}

type DocumentoInventarioResponse struct {
	ID                       int64                       `json:"id"`
	EmpresaID                int32                       `json:"empresa_id"`
	SucursalID               int32                       `json:"sucursal_id"`
	SucursalNombre           string                      `json:"sucursal_nombre,omitempty"`
	TipoMovimientoID         int32                       `json:"tipo_movimiento_id"`
	TipoMovimientoCodigo     string                      `json:"tipo_movimiento_codigo,omitempty"`
	TipoMovimientoNombre     string                      `json:"tipo_movimiento_nombre,omitempty"`
	TipoMovimientoNaturaleza string                      `json:"tipo_movimiento_naturaleza,omitempty"`
	Consecutivo              string                      `json:"consecutivo"`
	Fecha                    time.Time                   `json:"fecha"`
	TerceroID                *int64                      `json:"tercero_id"`
	TerceroNombre            string                      `json:"tercero_nombre,omitempty"`
	Motivo                   string                      `json:"motivo"`
	Observaciones            *string                     `json:"observaciones"`
	DocumentoOrigenTipo      *string                     `json:"documento_origen_tipo"`
	DocumentoOrigenID        *int64                      `json:"documento_origen_id"`
	UsuarioID                int64                       `json:"usuario_id"`
	UsuarioNombre            string                      `json:"usuario_nombre,omitempty"`
	Estado                   string                      `json:"estado"`
	Detalles                 []MovimientoDetalleResponse `json:"detalles,omitempty"`
	CreatedAt                time.Time                   `json:"created_at"`
}

type ListaMovimientosResponse struct {
	Data     []DocumentoInventarioResponse `json:"data"`
	Page     int                           `json:"page"`
	PageSize int                           `json:"page_size"`
	Total    int64                         `json:"total"`
}

func toUbicacionResponse(u *Ubicacion) UbicacionResponse {
	resp := UbicacionResponse{
		ID:          u.ID,
		EmpresaID:   u.EmpresaID,
		SucursalID:  u.SucursalID,
		Codigo:      u.Codigo,
		Nombre:      u.Nombre,
		Descripcion: u.Descripcion,
		Activo:      u.Activo,
		CreatedAt:   u.CreatedAt,
		UpdatedAt:   u.UpdatedAt,
	}
	if u.Sucursal != nil {
		resp.SucursalNombre = u.Sucursal.Nombre
	}
	return resp
}

func toTipoMovimientoResponse(t *TipoMovimientoInventario) TipoMovimientoResponse {
	return TipoMovimientoResponse{
		ID:          t.ID,
		Codigo:      t.Codigo,
		Nombre:      t.Nombre,
		Naturaleza:  t.Naturaleza,
		Descripcion: t.Descripcion,
		Activo:      t.Activo,
	}
}

func toStockResponse(s *InventarioStock) StockItemResponse {
	resp := StockItemResponse{
		ID:            s.ID,
		EmpresaID:     s.EmpresaID,
		ArticuloID:    s.ArticuloID,
		UbicacionID:   s.UbicacionID,
		Cantidad:      s.Cantidad,
		CostoPromedio: s.CostoPromedio,
		ValorTotal:    round4(s.Cantidad * s.CostoPromedio),
		UpdatedAt:     s.UpdatedAt,
	}
	if s.Articulo != nil {
		resp.ArticuloCodigo = s.Articulo.Codigo
		resp.ArticuloNombre = s.Articulo.Nombre
	}
	if s.Ubicacion != nil {
		resp.UbicacionCodigo = s.Ubicacion.Codigo
		resp.UbicacionNombre = s.Ubicacion.Nombre
	}
	return resp
}

func toDocumentoResponse(d *DocumentoInventario, includeDetalles bool) DocumentoInventarioResponse {
	resp := DocumentoInventarioResponse{
		ID:                  d.ID,
		EmpresaID:           d.EmpresaID,
		SucursalID:          d.SucursalID,
		TipoMovimientoID:    d.TipoMovimientoID,
		Consecutivo:         d.Consecutivo,
		Fecha:               d.Fecha,
		TerceroID:           d.TerceroID,
		Motivo:              d.Motivo,
		Observaciones:       d.Observaciones,
		DocumentoOrigenTipo: d.DocumentoOrigenTipo,
		DocumentoOrigenID:   d.DocumentoOrigenID,
		UsuarioID:           d.UsuarioID,
		Estado:              d.Estado,
		CreatedAt:           d.CreatedAt,
	}
	if d.Sucursal != nil {
		resp.SucursalNombre = d.Sucursal.Nombre
	}
	if d.TipoMovimiento != nil {
		resp.TipoMovimientoCodigo = d.TipoMovimiento.Codigo
		resp.TipoMovimientoNombre = d.TipoMovimiento.Nombre
		resp.TipoMovimientoNaturaleza = d.TipoMovimiento.Naturaleza
	}
	if d.Tercero != nil {
		resp.TerceroNombre = d.Tercero.RazonSocial
	}
	if d.Usuario != nil {
		resp.UsuarioNombre = strings.TrimSpace(d.Usuario.Nombres + " " + d.Usuario.Apellidos)
	}
	if includeDetalles {
		detalles := make([]MovimientoDetalleResponse, 0, len(d.Detalles))
		for i := range d.Detalles {
			m := &d.Detalles[i]
			dr := MovimientoDetalleResponse{
				ID:                 m.ID,
				ArticuloID:         m.ArticuloID,
				UbicacionOrigenID:  m.UbicacionOrigenID,
				UbicacionDestinoID: m.UbicacionDestinoID,
				Cantidad:           m.Cantidad,
				CostoUnitario:      m.CostoUnitario,
			}
			if m.Articulo != nil {
				dr.ArticuloCodigo = m.Articulo.Codigo
				dr.ArticuloNombre = m.Articulo.Nombre
			}
			if m.UbicacionOrigen != nil {
				dr.UbicacionOrigenNombre = m.UbicacionOrigen.Nombre
			}
			if m.UbicacionDestino != nil {
				dr.UbicacionDestinoNombre = m.UbicacionDestino.Nombre
			}
			detalles = append(detalles, dr)
		}
		resp.Detalles = detalles
	}
	return resp
}
