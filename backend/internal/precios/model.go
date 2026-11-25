package precios

import (
	"time"

	"gorm.io/gorm"
)

const (
	TipoListaGeneral   = "GENERAL"
	TipoListaMayorista = "MAYORISTA"
	TipoListaMinorista = "MINORISTA"
	TipoListaEspecial  = "ESPECIAL"
)

type MonedaRef struct {
	ID           int32  `gorm:"column:id;primaryKey"`
	CodigoISO    string `gorm:"column:codigo_iso"`
	NombreMoneda string `gorm:"column:nombre_moneda"`
	Simbolo      string `gorm:"column:simbolo"`
}

func (MonedaRef) TableName() string { return "monedas" }

type ArticuloRef struct {
	ID        int32  `gorm:"column:id;primaryKey"`
	EmpresaID int32  `gorm:"column:empresa_id"`
	Codigo    string `gorm:"column:codigo"`
	Nombre    string `gorm:"column:nombre"`
}

func (ArticuloRef) TableName() string { return "articulos" }

type UnidadMedidaRef struct {
	ID          int32   `gorm:"column:id;primaryKey"`
	Nombre      string  `gorm:"column:nombre"`
	Descripcion *string `gorm:"column:descripcion"`
}

func (UnidadMedidaRef) TableName() string { return "unidades_medida" }

type ListaPrecio struct {
	ID            int32          `gorm:"column:id;primaryKey"`
	EmpresaID     int32          `gorm:"column:empresa_id;not null"`
	Nombre        string         `gorm:"column:nombre;not null"`
	MonedaID      int32          `gorm:"column:moneda_id;not null"`
	TipoLista     string         `gorm:"column:tipo_lista;not null"`
	VigenciaDesde time.Time      `gorm:"column:vigencia_desde;type:date;not null"`
	VigenciaHasta *time.Time     `gorm:"column:vigencia_hasta;type:date"`
	Estado        bool           `gorm:"column:estado;not null"`
	CreatedAt     time.Time      `gorm:"column:created_at"`
	UpdatedAt     time.Time      `gorm:"column:updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"column:deleted_at;index"`

	Moneda *MonedaRef `gorm:"foreignKey:MonedaID;references:ID"`
}

func (ListaPrecio) TableName() string { return "listas_precios" }

type ListaPrecioDetalle struct {
	ID             int64      `gorm:"column:id;primaryKey"`
	ListaPrecioID  int32      `gorm:"column:lista_precio_id;not null"`
	ArticuloID     int32      `gorm:"column:articulo_id;not null"`
	Precio         float64    `gorm:"column:precio;not null"`
	UnidadMedidaID int32      `gorm:"column:unidad_medida_id;not null"`
	CantidadMinima float64    `gorm:"column:cantidad_minima;not null"`
	VigenciaDesde  *time.Time `gorm:"column:vigencia_desde;type:date"`
	VigenciaHasta  *time.Time `gorm:"column:vigencia_hasta;type:date"`
	CreatedAt      time.Time  `gorm:"column:created_at"`
	UpdatedAt      time.Time  `gorm:"column:updated_at"`

	Articulo     *ArticuloRef     `gorm:"foreignKey:ArticuloID;references:ID"`
	UnidadMedida *UnidadMedidaRef `gorm:"foreignKey:UnidadMedidaID;references:ID"`
}

func (ListaPrecioDetalle) TableName() string { return "lista_precios_detalle" }
