package crm

import (
	"time"

	"gorm.io/gorm"
)

// Constantes válidas para temperatura de lead, tipo y estado de actividad.
const (
	TempFrio     = "FRIO"
	TempTibio    = "TIBIO"
	TempCaliente = "CALIENTE"

	ActTipoLlamada  = "LLAMADA"
	ActTipoCorreo   = "CORREO"
	ActTipoReunion  = "REUNION"
	ActTipoWhatsapp = "WHATSAPP"
	ActTipoTarea    = "TAREA"
	ActTipoNota     = "NOTA"

	ActEstadoPendiente  = "PENDIENTE"
	ActEstadoCompletada = "COMPLETADA"
	ActEstadoCancelada  = "CANCELADA"
)

type OrigenLead struct {
	ID        int32          `gorm:"column:id;primaryKey"`
	EmpresaID int32          `gorm:"column:empresa_id;not null"`
	Nombre    string         `gorm:"column:nombre;not null"`
	Activo    bool           `gorm:"column:activo;not null"`
	CreatedAt time.Time      `gorm:"column:created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (OrigenLead) TableName() string { return "crm_origen_lead" }

type EstadoLead struct {
	ID            int32          `gorm:"column:id;primaryKey"`
	EmpresaID     int32          `gorm:"column:empresa_id;not null"`
	Nombre        string         `gorm:"column:nombre;not null"`
	Orden         int32          `gorm:"column:orden;not null"`
	EsEstadoFinal bool           `gorm:"column:es_estado_final;not null"`
	CreatedAt     time.Time      `gorm:"column:created_at"`
	UpdatedAt     time.Time      `gorm:"column:updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (EstadoLead) TableName() string { return "crm_estado_lead" }

type MotivoDescarte struct {
	ID          int32          `gorm:"column:id;primaryKey"`
	EmpresaID   int32          `gorm:"column:empresa_id;not null"`
	Descripcion string         `gorm:"column:descripcion;not null"`
	Activo      bool           `gorm:"column:activo;not null"`
	CreatedAt   time.Time      `gorm:"column:created_at"`
	UpdatedAt   time.Time      `gorm:"column:updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (MotivoDescarte) TableName() string { return "crm_motivo_descarte" }

type Lead struct {
	ID                             int64          `gorm:"column:id;primaryKey"`
	EmpresaID                      int32          `gorm:"column:empresa_id;not null"`
	SucursalID                     *int32         `gorm:"column:sucursal_id"`
	OrigenID                       int32          `gorm:"column:origen_id;not null"`
	EstadoID                       int32          `gorm:"column:estado_id;not null"`
	MotivoDescarteID               *int32         `gorm:"column:motivo_descarte_id"`
	UsuarioAsignadoID              *int64         `gorm:"column:usuario_asignado_id"`
	Nombre                         string         `gorm:"column:nombre;not null"`
	Apellido                       string         `gorm:"column:apellido;not null"`
	NombreEmpresa                  *string        `gorm:"column:nombre_empresa"`
	Cargo                          *string        `gorm:"column:cargo"`
	Email                          *string        `gorm:"column:email"`
	Telefono                       *string        `gorm:"column:telefono"`
	UsuarioFacebook                *string        `gorm:"column:usuario_facebook"`
	Temperatura                    string         `gorm:"column:temperatura;not null"`
	Score                          int32          `gorm:"column:score;not null"`
	ValorEstimado                  float64        `gorm:"column:valor_estimado;not null"`
	MonedaID                       *int32         `gorm:"column:moneda_id"`
	Convertido                     bool           `gorm:"column:convertido;not null"`
	ConvertidoTerceroEmpresaRolID  *int64         `gorm:"column:convertido_tercero_empresa_rol_id"`
	FechaConversion                *time.Time     `gorm:"column:fecha_conversion"`
	CreatedBy                      int64          `gorm:"column:created_by;not null"`
	CreatedAt                      time.Time      `gorm:"column:created_at"`
	UpdatedAt                      time.Time      `gorm:"column:updated_at"`
	DeletedAt                      gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (Lead) TableName() string { return "crm_leads" }

type LeadActividad struct {
	ID               int64          `gorm:"column:id;primaryKey"`
	LeadID           int64          `gorm:"column:lead_id;not null"`
	UsuarioID        int64          `gorm:"column:usuario_id;not null"`
	TipoActividad    string         `gorm:"column:tipo_actividad;not null"`
	FechaActividad   time.Time      `gorm:"column:fecha_actividad;not null"`
	FechaVencimiento *time.Time     `gorm:"column:fecha_vencimiento"`
	Estado           string         `gorm:"column:estado;not null"`
	Asunto           string         `gorm:"column:asunto;not null"`
	Observaciones    *string        `gorm:"column:observaciones"`
	CreatedAt        time.Time      `gorm:"column:created_at"`
	UpdatedAt        time.Time      `gorm:"column:updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (LeadActividad) TableName() string { return "crm_lead_actividades" }

type LeadInteresProducto struct {
	ID                int64     `gorm:"column:id;primaryKey"`
	LeadID            int64     `gorm:"column:lead_id;not null"`
	EmpresaProductoID *int32    `gorm:"column:empresa_producto_id"`
	CategoriaID       *int32    `gorm:"column:categoria_id"`
	CantidadEstimada  float64   `gorm:"column:cantidad_estimada;not null"`
	Notas             *string   `gorm:"column:notas"`
	CreatedAt         time.Time `gorm:"column:created_at"`
}

func (LeadInteresProducto) TableName() string { return "crm_lead_interes_productos" }
