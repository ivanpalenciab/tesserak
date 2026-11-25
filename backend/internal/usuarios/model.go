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
	SucursalID    int32          `gorm:"column:sucursal_id;not null"`
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
