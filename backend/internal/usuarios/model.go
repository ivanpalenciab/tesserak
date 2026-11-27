package usuarios

import (
	"time"

	"gorm.io/gorm"
)

// Usuario mapea la tabla "usuarios".
//
// El esquema lo define la migración SQL (migrations/000001_init_usuarios.up.sql);
// no se usa AutoMigrate porque necesitamos índices únicos parciales y
// FK compuestas que GORM no genera correctamente.
type Usuario struct {
	ID            int64          `gorm:"column:id;primaryKey"`
	Key           string         `gorm:"column:key;type:uuid;not null;default:gen_random_uuid()"`
	EmpresaID     int32          `gorm:"column:empresa_id;not null"`
	NombreUsuario string         `gorm:"column:nombre_usuario;not null"`
	Nombres       string         `gorm:"column:nombres;not null"`
	Apellidos     string         `gorm:"column:apellidos;not null"`
	Email         string         `gorm:"column:email;not null"`
	PasswordHash  string         `gorm:"column:password_hash;not null"`
	Activo        bool           `gorm:"column:activo;not null"`
	UltimoLoginAt *time.Time     `gorm:"column:ultimo_login_at"`
	CreatedAt     time.Time      `gorm:"column:created_at"`
	UpdatedAt     time.Time      `gorm:"column:updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"column:deleted_at;index"` // soft delete
}

func (Usuario) TableName() string { return "usuarios" }

// UsuarioSucursal mapea la tabla intermedia "usuario_sucursales" (relación N:M).
type UsuarioSucursal struct {
	UsuarioID   int64     `gorm:"column:usuario_id;primaryKey"`
	SucursalID  int32     `gorm:"column:sucursal_id;primaryKey"`
	EmpresaID   int32     `gorm:"column:empresa_id;not null"`
	EsPrincipal bool      `gorm:"column:es_principal;not null;default:false"`
	CreatedAt   time.Time `gorm:"column:created_at"`
}

func (UsuarioSucursal) TableName() string { return "usuario_sucursales" }

// RolSistema mapea la tabla "roles_sistema".
type RolSistema struct {
	ID          int32     `gorm:"column:id;primaryKey"`
	Codigo      string    `gorm:"column:codigo;not null"`
	Nombre      string    `gorm:"column:nombre;not null"`
	Descripcion *string   `gorm:"column:descripcion"`
	EsSistema   bool      `gorm:"column:es_sistema;not null"`
	Activo      bool      `gorm:"column:activo;not null"`
	CreatedAt   time.Time `gorm:"column:created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
}

func (RolSistema) TableName() string { return "roles_sistema" }

// UsuarioRol mapea la tabla puente N:M "usuario_roles".
type UsuarioRol struct {
	UsuarioID int64     `gorm:"column:usuario_id;primaryKey"`
	RolID     int32     `gorm:"column:rol_id;primaryKey"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

func (UsuarioRol) TableName() string { return "usuario_roles" }

// SucursalResumen es una proyección ligera de "sucursales" para el selector en Usuarios.
type SucursalResumen struct {
	ID        int32   `gorm:"column:id"        json:"id"`
	Nombre    string  `gorm:"column:nombre"    json:"nombre"`
	Direccion *string `gorm:"column:direccion" json:"direccion"`
	Activo    bool    `gorm:"column:activo"    json:"activo"`
}
