package ventas

import (
	"time"

	"gorm.io/gorm"
)

const (
	EstadoBorrador   = "BORRADOR"
	EstadoRegistrada = "REGISTRADA"
	EstadoPagada     = "PAGADA"
	EstadoAnulada    = "ANULADA"
)

type SucursalRef struct {
	ID     int32  `gorm:"column:id;primaryKey"`
	Nombre string `gorm:"column:nombre"`
}

func (SucursalRef) TableName() string { return "sucursales" }

type UbicacionRef struct {
	ID     int32  `gorm:"column:id;primaryKey"`
	Codigo string `gorm:"column:codigo"`
	Nombre string `gorm:"column:nombre"`
}

func (UbicacionRef) TableName() string { return "inv_ubicaciones" }

type ClienteRef struct {
	ID                   int64  `gorm:"column:id;primaryKey"`
	TipoIdentificacion   string `gorm:"column:tipo_identificacion"`
	NumeroIdentificacion string `gorm:"column:numero_identificacion"`
	RazonSocial          string `gorm:"column:razon_social"`
}

func (ClienteRef) TableName() string { return "tercero_maestro" }

type MonedaRef struct {
	ID     int32  `gorm:"column:id;primaryKey"`
	Codigo string `gorm:"column:codigo"`
	Nombre string `gorm:"column:nombre"`
}

func (MonedaRef) TableName() string { return "monedas" }

type UsuarioRef struct {
	ID        int64  `gorm:"column:id;primaryKey"`
	Nombres   string `gorm:"column:nombres"`
	Apellidos string `gorm:"column:apellidos"`
}

func (UsuarioRef) TableName() string { return "usuarios" }

type ArticuloRef struct {
	ID     int32  `gorm:"column:id;primaryKey"`
	Codigo string `gorm:"column:codigo"`
	Nombre string `gorm:"column:nombre"`
}

func (ArticuloRef) TableName() string { return "articulos" }

type Venta struct {
	ID               int64          `gorm:"column:id;primaryKey"`
	EmpresaID        int32          `gorm:"column:empresa_id;not null"`
	SucursalID       int32          `gorm:"column:sucursal_id;not null"`
	UbicacionID      int32          `gorm:"column:ubicacion_id;not null"`
	ClienteID        int64          `gorm:"column:cliente_id;not null"`
	Consecutivo      string         `gorm:"column:consecutivo;not null"`
	MonedaID         *int32         `gorm:"column:moneda_id"`
	Estado           string         `gorm:"column:estado;not null"`
	FechaVenta       time.Time      `gorm:"column:fecha_venta;not null"`
	FechaVencimiento *time.Time     `gorm:"column:fecha_vencimiento"`
	Subtotal         float64        `gorm:"column:subtotal;not null"`
	TotalIVA         float64        `gorm:"column:total_iva;not null"`
	PuntosRedimidos  int64          `gorm:"column:puntos_redimidos;not null"`
	DescuentoLealtad float64        `gorm:"column:descuento_lealtad;not null"`
	Total            float64        `gorm:"column:total;not null"`
	Observaciones    *string        `gorm:"column:observaciones"`
	UsuarioID        int64          `gorm:"column:usuario_id;not null"`
	CreatedAt        time.Time      `gorm:"column:created_at"`
	UpdatedAt        time.Time      `gorm:"column:updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"column:deleted_at;index"`

	Sucursal  *SucursalRef   `gorm:"foreignKey:SucursalID;references:ID"`
	Ubicacion *UbicacionRef  `gorm:"foreignKey:UbicacionID;references:ID"`
	Cliente   *ClienteRef    `gorm:"foreignKey:ClienteID;references:ID"`
	Moneda    *MonedaRef     `gorm:"foreignKey:MonedaID;references:ID"`
	Usuario   *UsuarioRef    `gorm:"foreignKey:UsuarioID;references:ID"`
	Detalles  []VentaDetalle `gorm:"foreignKey:VentaID;references:ID"`
}

func (Venta) TableName() string { return "ven_ventas" }

type VentaDetalle struct {
	ID             int64   `gorm:"column:id;primaryKey"`
	VentaID        int64   `gorm:"column:venta_id;not null"`
	ArticuloID     int32   `gorm:"column:articulo_id;not null"`
	Descripcion    *string `gorm:"column:descripcion"`
	Cantidad       float64 `gorm:"column:cantidad;not null"`
	PrecioUnitario float64 `gorm:"column:precio_unitario;not null"`
	Subtotal       float64 `gorm:"column:subtotal;not null"`
	Impuesto       float64 `gorm:"column:impuesto;not null"`
	Total          float64 `gorm:"column:total;not null"`

	Articulo *ArticuloRef `gorm:"foreignKey:ArticuloID;references:ID"`
}

func (VentaDetalle) TableName() string { return "ven_ventas_detalle" }
