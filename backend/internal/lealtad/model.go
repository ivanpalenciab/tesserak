// Package lealtad implementa la administración de planes/programas de lealtad
// y la adhesión de empresas (arquitectura híbrida: LOCAL, COALICION y GLOBAL).
package lealtad

import (
	"time"

	"gorm.io/gorm"
)

const (
	AlcanceLocal     = "LOCAL"
	AlcanceCoalicion = "COALICION"
	AlcanceGlobal    = "GLOBAL"

	EstadoActivo   = "ACTIVO"
	EstadoInactivo = "INACTIVO"
	EstadoPausado  = "PAUSADO"

	RolEmpresaPropietaria = "PROPIETARIA"
	RolEmpresaAliada      = "ALIADA"

	CodigoPlanDefecto  = "LEALTAD_TESSERAK"
	UsuarioAdminGlobal = "ipalenciab"
)

type MonedaRef struct {
	ID           int32  `gorm:"column:id;primaryKey"`
	CodigoISO    string `gorm:"column:codigo_iso"`
	NombreMoneda string `gorm:"column:nombre_moneda"`
	Simbolo      string `gorm:"column:simbolo"`
}

func (MonedaRef) TableName() string { return "monedas" }

type EmpresaRef struct {
	ID                  int32  `gorm:"column:id;primaryKey"`
	NombreLegal         string `gorm:"column:nombre_legal"`
	RazonSocial         string `gorm:"column:razon_social"`
	IdentificadorFiscal string `gorm:"column:identificador_fiscal"`
}

func (EmpresaRef) TableName() string { return "empresas" }

type UsuarioRef struct {
	ID            int64  `gorm:"column:id;primaryKey"`
	NombreUsuario string `gorm:"column:nombre_usuario"`
	Nombres       string `gorm:"column:nombres"`
	Apellidos     string `gorm:"column:apellidos"`
	Email         string `gorm:"column:email"`
}

func (UsuarioRef) TableName() string { return "usuarios" }

type Programa struct {
	ID                     int32          `gorm:"column:id;primaryKey;autoIncrement"`
	Codigo                 string         `gorm:"column:codigo;not null"`
	Nombre                 string         `gorm:"column:nombre;not null"`
	Descripcion            *string        `gorm:"column:descripcion"`
	Alcance                string         `gorm:"column:alcance;not null"`
	EmpresaPropietariaID   *int32         `gorm:"column:empresa_propietaria_id"`
	UsuarioAdministradorID *int64         `gorm:"column:usuario_administrador_id"`
	EsSistema              bool           `gorm:"column:es_sistema;not null"`
	MonedaReferenciaID     int32          `gorm:"column:moneda_referencia_id;not null"`
	ValorPuntoRedencion    float64        `gorm:"column:valor_punto_redencion;not null"`
	DiasVencimiento        *int32         `gorm:"column:dias_vencimiento"`
	Estado                 string         `gorm:"column:estado;not null"`
	FechaInicio            time.Time      `gorm:"column:fecha_inicio;not null"`
	FechaFin               *time.Time     `gorm:"column:fecha_fin"`
	CreatedAt              time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt              time.Time      `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt              gorm.DeletedAt `gorm:"column:deleted_at"`

	Moneda               *MonedaRef  `gorm:"foreignKey:MonedaReferenciaID;references:ID"`
	EmpresaPropietaria   *EmpresaRef `gorm:"foreignKey:EmpresaPropietariaID;references:ID"`
	UsuarioAdministrador *UsuarioRef `gorm:"foreignKey:UsuarioAdministradorID;references:ID"`
}

func (Programa) TableName() string { return "lealtad_programas" }

type ProgramaEmpresa struct {
	ID                  int64      `gorm:"column:id;primaryKey;autoIncrement"`
	ProgramaID          int32      `gorm:"column:programa_id;not null"`
	EmpresaID           int32      `gorm:"column:empresa_id;not null"`
	RolEmpresa          string     `gorm:"column:rol_empresa;not null"`
	PermiteAcumular     bool       `gorm:"column:permite_acumular;not null"`
	PermiteRedimir      bool       `gorm:"column:permite_redimir;not null"`
	EsProgramaPrincipal bool       `gorm:"column:es_programa_principal;not null"`
	FechaAdhesion       time.Time  `gorm:"column:fecha_adhesion;not null"`
	FechaBaja           *time.Time `gorm:"column:fecha_baja"`
	Activo              bool       `gorm:"column:activo;not null"`
	CreatedAt           time.Time  `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt           time.Time  `gorm:"column:updated_at;autoUpdateTime"`

	Empresa *EmpresaRef `gorm:"foreignKey:EmpresaID;references:ID"`
}

func (ProgramaEmpresa) TableName() string { return "lealtad_programa_empresas" }
