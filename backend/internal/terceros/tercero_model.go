package terceros

import (
	"time"

	"gorm.io/gorm"
)

type PaisRef struct {
	ID        int32  `gorm:"column:id;primaryKey"`
	Nombre    string `gorm:"column:nombre"`
	IsoAlpha2 string `gorm:"column:iso_alpha_2"`
}

func (PaisRef) TableName() string { return "paises" }

type TerceroRol struct {
	ID        int64     `gorm:"column:id;primaryKey"`
	TerceroID int64     `gorm:"column:tercero_id;not null"`
	RolID     int32     `gorm:"column:rol_id;not null"`
	CreatedAt time.Time `gorm:"column:created_at"`

	Rol *RolTercero `gorm:"foreignKey:RolID;references:ID"`
}

func (TerceroRol) TableName() string { return "tercero_empresa_roles" }

type TerceroMaestro struct {
	ID                   int64          `gorm:"column:id;primaryKey"`
	EmpresaID            int32          `gorm:"column:empresa_id;not null"`
	TipoIdentificacion   string         `gorm:"column:tipo_identificacion;not null"`
	NumeroIdentificacion string         `gorm:"column:numero_identificacion;not null"`
	RazonSocial          string         `gorm:"column:razon_social;not null"`
	PaisID               *int32         `gorm:"column:pais_id"`
	NaturalezaJuridica   string         `gorm:"column:naturaleza_juridica;not null"`
	Email                *string        `gorm:"column:email"`
	Telefono             *string        `gorm:"column:telefono"`
	Direccion            *string        `gorm:"column:direccion"`
	Activo               bool           `gorm:"column:activo;not null"`
	CreatedAt            time.Time      `gorm:"column:created_at"`
	UpdatedAt            time.Time      `gorm:"column:updated_at"`
	DeletedAt            gorm.DeletedAt `gorm:"column:deleted_at;index"`

	Pais  *PaisRef     `gorm:"foreignKey:PaisID;references:ID"`
	Roles []TerceroRol `gorm:"foreignKey:TerceroID;references:ID"`
}

func (TerceroMaestro) TableName() string { return "tercero_maestro" }
