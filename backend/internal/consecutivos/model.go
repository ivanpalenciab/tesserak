package consecutivos

import "time"

const (
	TipoCompra = "COMPRA"
	TipoVenta  = "VENTA"

	FrecuenciaNunca   = "NUNCA"
	FrecuenciaAnual   = "ANUAL"
	FrecuenciaMensual = "MENSUAL"
)

type TipoDocumento struct {
	ID             int32     `gorm:"column:id;primaryKey"`
	Codigo         string    `gorm:"column:codigo;not null"`
	Nombre         string    `gorm:"column:nombre;not null"`
	Descripcion    *string   `gorm:"column:descripcion"`
	PrefijoDefecto string    `gorm:"column:prefijo_defecto;not null"`
	Activo         bool      `gorm:"column:activo;not null"`
	CreatedAt      time.Time `gorm:"column:created_at"`
	UpdatedAt      time.Time `gorm:"column:updated_at"`
}

func (TipoDocumento) TableName() string { return "tipos_documento" }

type SecuenciaDocumento struct {
	ID                 int64          `gorm:"column:id;primaryKey"`
	EmpresaID          int32          `gorm:"column:empresa_id;not null"`
	TipoDocumentoID    int32          `gorm:"column:tipo_documento_id;not null"`
	Prefijo            string         `gorm:"column:prefijo;not null"`
	SiguienteNumero    int64          `gorm:"column:siguiente_numero;not null"`
	LongitudRelleno    int16          `gorm:"column:longitud_relleno;not null"`
	FrecuenciaReinicio string         `gorm:"column:frecuencia_reinicio;not null"`
	ClavePeriodo       *string        `gorm:"column:clave_periodo"`
	Activo             bool           `gorm:"column:activo;not null"`
	CreatedAt          time.Time      `gorm:"column:created_at"`
	UpdatedAt          time.Time      `gorm:"column:updated_at"`
	TipoDocumento      *TipoDocumento `gorm:"foreignKey:TipoDocumentoID"`
}

func (SecuenciaDocumento) TableName() string { return "secuencias_documento" }
