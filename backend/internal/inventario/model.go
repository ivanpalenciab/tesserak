package inventario

import (
	"time"

	"gorm.io/gorm"
)

const (
	NaturalezaEntrada  = "ENTRADA"
	NaturalezaSalida   = "SALIDA"
	NaturalezaTraslado = "TRASLADO"

	EstadoAplicado = "APLICADO"
	EstadoAnulado  = "ANULADO"
)

type SucursalRef struct {
	ID        int32  `gorm:"column:id;primaryKey"`
	EmpresaID int32  `gorm:"column:empresa_id"`
	Nombre    string `gorm:"column:nombre"`
	Activo    bool   `gorm:"column:activo"`
}

func (SucursalRef) TableName() string { return "sucursales" }

type ArticuloRef struct {
	ID        int32  `gorm:"column:id;primaryKey"`
	EmpresaID int32  `gorm:"column:empresa_id"`
	Codigo    string `gorm:"column:codigo"`
	Nombre    string `gorm:"column:nombre"`
}

func (ArticuloRef) TableName() string { return "articulos" }

type UsuarioRef struct {
	ID        int64  `gorm:"column:id;primaryKey"`
	Nombres   string `gorm:"column:nombres"`
	Apellidos string `gorm:"column:apellidos"`
}

func (UsuarioRef) TableName() string { return "usuarios" }

type TerceroRef struct {
	ID          int64  `gorm:"column:id;primaryKey"`
	EmpresaID   int32  `gorm:"column:empresa_id"`
	RazonSocial string `gorm:"column:razon_social"`
}

func (TerceroRef) TableName() string { return "tercero_maestro" }

type Ubicacion struct {
	ID          int32          `gorm:"column:id;primaryKey"`
	EmpresaID   int32          `gorm:"column:empresa_id;not null"`
	SucursalID  int32          `gorm:"column:sucursal_id;not null"`
	Codigo      string         `gorm:"column:codigo;not null"`
	Nombre      string         `gorm:"column:nombre;not null"`
	Descripcion *string        `gorm:"column:descripcion"`
	Activo      bool           `gorm:"column:activo;not null"`
	CreatedAt   time.Time      `gorm:"column:created_at"`
	UpdatedAt   time.Time      `gorm:"column:updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at;index"`

	Sucursal *SucursalRef `gorm:"foreignKey:SucursalID;references:ID"`
}

func (Ubicacion) TableName() string { return "inv_ubicaciones" }

type TipoMovimientoInventario struct {
	ID          int32   `gorm:"column:id;primaryKey"`
	Codigo      string  `gorm:"column:codigo;not null"`
	Nombre      string  `gorm:"column:nombre;not null"`
	Naturaleza  string  `gorm:"column:naturaleza;not null"`
	Descripcion *string `gorm:"column:descripcion"`
	Activo      bool    `gorm:"column:activo;not null"`
}

func (TipoMovimientoInventario) TableName() string { return "inv_tipos_movimiento" }

type DocumentoInventario struct {
	ID                  int64     `gorm:"column:id;primaryKey"`
	EmpresaID           int32     `gorm:"column:empresa_id;not null"`
	SucursalID          int32     `gorm:"column:sucursal_id;not null"`
	TipoMovimientoID    int32     `gorm:"column:tipo_movimiento_id;not null"`
	Consecutivo         string    `gorm:"column:consecutivo;not null"`
	Fecha               time.Time `gorm:"column:fecha;not null"`
	TerceroID           *int64    `gorm:"column:tercero_id"`
	Motivo              string    `gorm:"column:motivo;not null"`
	Observaciones       *string   `gorm:"column:observaciones"`
	DocumentoOrigenTipo *string   `gorm:"column:documento_origen_tipo"`
	DocumentoOrigenID   *int64    `gorm:"column:documento_origen_id"`
	UsuarioID           int64     `gorm:"column:usuario_id;not null"`
	Estado              string    `gorm:"column:estado;not null"`
	CreatedAt           time.Time `gorm:"column:created_at"`
	UpdatedAt           time.Time `gorm:"column:updated_at"`

	Sucursal       *SucursalRef                  `gorm:"foreignKey:SucursalID;references:ID"`
	TipoMovimiento *TipoMovimientoInventario     `gorm:"foreignKey:TipoMovimientoID;references:ID"`
	Tercero        *TerceroRef                   `gorm:"foreignKey:TerceroID;references:ID"`
	Usuario        *UsuarioRef                   `gorm:"foreignKey:UsuarioID;references:ID"`
	Detalles       []MovimientoInventarioDetalle `gorm:"foreignKey:DocumentoID;references:ID"`
}

func (DocumentoInventario) TableName() string { return "inv_documentos" }

type MovimientoInventarioDetalle struct {
	ID                 int64     `gorm:"column:id;primaryKey"`
	DocumentoID        int64     `gorm:"column:documento_id;not null"`
	ArticuloID         int32     `gorm:"column:articulo_id;not null"`
	UbicacionOrigenID  *int32    `gorm:"column:ubicacion_origen_id"`
	UbicacionDestinoID *int32    `gorm:"column:ubicacion_destino_id"`
	Cantidad           float64   `gorm:"column:cantidad;not null"`
	CostoUnitario      float64   `gorm:"column:costo_unitario;not null"`
	CreatedAt          time.Time `gorm:"column:created_at"`

	Articulo         *ArticuloRef `gorm:"foreignKey:ArticuloID;references:ID"`
	UbicacionOrigen  *Ubicacion   `gorm:"foreignKey:UbicacionOrigenID;references:ID"`
	UbicacionDestino *Ubicacion   `gorm:"foreignKey:UbicacionDestinoID;references:ID"`
}

func (MovimientoInventarioDetalle) TableName() string { return "inv_movimientos_detalle" }

type InventarioStock struct {
	ID            int64     `gorm:"column:id;primaryKey"`
	EmpresaID     int32     `gorm:"column:empresa_id;not null"`
	ArticuloID    int32     `gorm:"column:articulo_id;not null"`
	UbicacionID   int32     `gorm:"column:ubicacion_id;not null"`
	Cantidad      float64   `gorm:"column:cantidad;not null"`
	CostoPromedio float64   `gorm:"column:costo_promedio;not null"`
	UpdatedAt     time.Time `gorm:"column:updated_at"`

	Articulo  *ArticuloRef `gorm:"foreignKey:ArticuloID;references:ID"`
	Ubicacion *Ubicacion   `gorm:"foreignKey:UbicacionID;references:ID"`
}

func (InventarioStock) TableName() string { return "inv_stock" }
