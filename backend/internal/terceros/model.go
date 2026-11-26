package terceros

import (
	"time"

	"gorm.io/gorm"
)

const (
	RolCodigoCliente   = "CLIENTE"
	RolCodigoProveedor = "PROVEEDOR"
	RolCodigoEmpleado  = "EMPLEADO"
)

type RolTercero struct {
	ID          int32          `gorm:"column:id;primaryKey"`
	Codigo      string         `gorm:"column:codigo;not null"`
	NombreRol   string         `gorm:"column:nombre_rol;not null"`
	Descripcion *string        `gorm:"column:descripcion"`
	Activo      bool           `gorm:"column:activo;not null"`
	EsSistema   bool           `gorm:"column:es_sistema;not null"`
	CreatedAt   time.Time      `gorm:"column:created_at"`
	UpdatedAt   time.Time      `gorm:"column:updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (RolTercero) TableName() string { return "roles_tercero" }
