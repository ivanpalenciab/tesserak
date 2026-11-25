package productos

import (
	"time"

	"gorm.io/gorm"
)

const (
	TipoCodigoProducto = "PRODUCTO"
	TipoCodigoServicio = "SERVICIO"
)

type TipoArticulo struct {
	ID     int32  `gorm:"column:id;primaryKey"`
	Codigo string `gorm:"column:codigo;not null"`
	Nombre string `gorm:"column:nombre;not null"`
}

func (TipoArticulo) TableName() string { return "tipos_articulo" }

type MetodoValoracion struct {
	ID     int32  `gorm:"column:id;primaryKey"`
	Codigo string `gorm:"column:codigo;not null"`
	Nombre string `gorm:"column:nombre;not null"`
}

func (MetodoValoracion) TableName() string { return "metodos_valoracion" }

type ModalidadServicio struct {
	ID     int32  `gorm:"column:id;primaryKey"`
	Codigo string `gorm:"column:codigo;not null"`
	Nombre string `gorm:"column:nombre;not null"`
}

func (ModalidadServicio) TableName() string { return "modalidades_servicio" }

type GrupoImpuesto struct {
	ID        int32          `gorm:"column:id;primaryKey"`
	EmpresaID int32          `gorm:"column:empresa_id;not null"`
	Nombre    string         `gorm:"column:nombre;not null"`
	Tarifa    float64        `gorm:"column:tarifa;not null"`
	Activo    bool           `gorm:"column:activo;not null"`
	CreatedAt time.Time      `gorm:"column:created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (GrupoImpuesto) TableName() string { return "grupos_impuesto" }

type Articulo struct {
	ID              int32          `gorm:"column:id;primaryKey"`
	EmpresaID       int32          `gorm:"column:empresa_id;not null"`
	Codigo          string         `gorm:"column:codigo;not null"`
	Nombre          string         `gorm:"column:nombre;not null"`
	Descripcion     *string        `gorm:"column:descripcion"`
	TipoArticuloID  int32          `gorm:"column:tipo_articulo_id;not null"`
	UnidadMedidaID  int32          `gorm:"column:unidad_medida_id;not null"`
	CategoriaID     int32          `gorm:"column:categoria_id;not null"`
	MarcaID         *int32         `gorm:"column:marca_id"`
	GrupoImpuestoID int32          `gorm:"column:grupo_impuesto_id;not null"`
	Estado          bool           `gorm:"column:estado;not null"`
	CreatedAt       time.Time      `gorm:"column:created_at"`
	UpdatedAt       time.Time      `gorm:"column:updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"column:deleted_at;index"`

	TipoArticulo *TipoArticulo      `gorm:"foreignKey:TipoArticuloID;references:ID"`
	Producto     *ProductoExtension `gorm:"foreignKey:ArticuloID;references:ID"`
	Servicio     *ServicioExtension `gorm:"foreignKey:ArticuloID;references:ID"`
}

func (Articulo) TableName() string { return "articulos" }

type ProductoExtension struct {
	ArticuloID         int32  `gorm:"column:articulo_id;primaryKey"`
	ControlInventario  bool   `gorm:"column:control_inventario;not null"`
	MetodoValoracionID *int32 `gorm:"column:metodo_valoracion_id"`
	PermiteCompra      bool   `gorm:"column:permite_compra;not null"`
	PermiteVenta       bool   `gorm:"column:permite_venta;not null"`
}

func (ProductoExtension) TableName() string { return "productos" }

type ServicioExtension struct {
	ArticuloID           int32    `gorm:"column:articulo_id;primaryKey"`
	ModalidadID          int32    `gorm:"column:modalidad_id;not null"`
	DuracionEstimada     *float64 `gorm:"column:duracion_estimada"`
	RequiereProgramacion bool     `gorm:"column:requiere_programacion;not null"`
	ControlConsumos      bool     `gorm:"column:control_consumos;not null"`
}

func (ServicioExtension) TableName() string { return "servicios" }
