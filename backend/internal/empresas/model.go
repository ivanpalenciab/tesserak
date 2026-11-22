package empresas

import (
	"time"

	"gorm.io/gorm"
)

// Empresa mapea la tabla "empresas".
type Empresa struct {
	ID                  int32          `gorm:"column:id;primaryKey"`
	NombreLegal         string         `gorm:"column:nombre_legal;not null"`
	RazonSocial         string         `gorm:"column:razon_social;not null"`
	IdentificadorFiscal string         `gorm:"column:identificador_fiscal;not null"`
	Email               *string        `gorm:"column:email"`
	Direccion           *string        `gorm:"column:direccion"`
	Logo                *string        `gorm:"column:logo"`
	Activo              bool           `gorm:"column:activo;not null"`
	CreatedAt           time.Time      `gorm:"column:created_at"`
	UpdatedAt           time.Time      `gorm:"column:updated_at"`
	DeletedAt           gorm.DeletedAt `gorm:"column:deleted_at;index"` // soft delete
}

func (Empresa) TableName() string { return "empresas" }
