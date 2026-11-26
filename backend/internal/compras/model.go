package compras

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

type MonedaRef struct {
	ID           int32  `gorm:"column:id;primaryKey"`
	CodigoISO    string `gorm:"column:codigo_iso"`
	NombreMoneda string `gorm:"column:nombre_moneda"`
	Simbolo      string `gorm:"column:simbolo"`
}

func (MonedaRef) TableName() string { return "monedas" }

type SucursalRef struct {
	ID        int32  `gorm:"column:id;primaryKey"`
	EmpresaID int32  `gorm:"column:empresa_id"`
	Nombre    string `gorm:"column:nombre"`
}

func (SucursalRef) TableName() string { return "sucursales" }

type UsuarioRef struct {
	ID        int64  `gorm:"column:id;primaryKey"`
	Nombres   string `gorm:"column:nombres"`
	Apellidos string `gorm:"column:apellidos"`
}

func (UsuarioRef) TableName() string { return "usuarios" }

type ArticuloRef struct {
	ID        int32  `gorm:"column:id;primaryKey"`
	EmpresaID int32  `gorm:"column:empresa_id"`
	Codigo    string `gorm:"column:codigo"`
	Nombre    string `gorm:"column:nombre"`
}

func (ArticuloRef) TableName() string { return "articulos" }

type Compra struct {
	ID                     int64          `gorm:"column:id;primaryKey"`
	EmpresaID              int32          `gorm:"column:empresa_id;not null"`
	SucursalID             int32          `gorm:"column:sucursal_id;not null"`
	TerceroEmpresaRolID    *int64         `gorm:"column:tercero_empresa_rol_id"`
	ProveedorNombre        *string        `gorm:"column:proveedor_nombre"`
	Consecutivo            string         `gorm:"column:consecutivo;not null"`
	NumeroFacturaProveedor *string        `gorm:"column:numero_factura_proveedor"`
	MonedaID               *int32         `gorm:"column:moneda_id"`
	Estado                 string         `gorm:"column:estado;not null"`
	FechaCompra            time.Time      `gorm:"column:fecha_compra;not null"`
	FechaVencimiento       *time.Time     `gorm:"column:fecha_vencimiento"`
	Subtotal               float64        `gorm:"column:subtotal;not null"`
	TotalIVA               float64        `gorm:"column:total_iva;not null"`
	Total                  float64        `gorm:"column:total;not null"`
	Observaciones          *string        `gorm:"column:observaciones"`
	UsuarioID              int64          `gorm:"column:usuario_id;not null"`
	CreatedAt              time.Time      `gorm:"column:created_at"`
	UpdatedAt              time.Time      `gorm:"column:updated_at"`
	DeletedAt              gorm.DeletedAt `gorm:"column:deleted_at;index"`

	Moneda   *MonedaRef      `gorm:"foreignKey:MonedaID;references:ID"`
	Sucursal *SucursalRef    `gorm:"foreignKey:SucursalID;references:ID"`
	Usuario  *UsuarioRef     `gorm:"foreignKey:UsuarioID;references:ID"`
	Detalles []CompraDetalle `gorm:"foreignKey:CompraID;references:ID"`
}

func (Compra) TableName() string { return "com_compras" }

type CompraDetalle struct {
	ID            int64   `gorm:"column:id;primaryKey"`
	CompraID      int64   `gorm:"column:compra_id;not null"`
	ArticuloID    int32   `gorm:"column:articulo_id;not null"`
	Descripcion   *string `gorm:"column:descripcion"`
	Cantidad      float64 `gorm:"column:cantidad;not null"`
	CostoUnitario float64 `gorm:"column:costo_unitario;not null"`
	Subtotal      float64 `gorm:"column:subtotal;not null"`
	Impuesto      float64 `gorm:"column:impuesto;not null"`
	Total         float64 `gorm:"column:total;not null"`

	Articulo *ArticuloRef `gorm:"foreignKey:ArticuloID;references:ID"`
}

func (CompraDetalle) TableName() string { return "com_compras_detalle" }
