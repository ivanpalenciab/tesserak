package crm

import (
	"time"

	"gorm.io/gorm"
)

// Estados válidos de una campaña de marketing.
const (
	CampanaEstadoBorrador   = "BORRADOR"
	CampanaEstadoProgramada = "PROGRAMADA"
	CampanaEstadoActiva     = "ACTIVA"
	CampanaEstadoPausada    = "PAUSADA"
	CampanaEstadoFinalizada = "FINALIZADA"

	AtribucionFirstTouch = "first_touch"
	AtribucionLastTouch  = "last_touch"
	AtribucionInfluencia = "influencia"
)

type TipoCampana struct {
	ID          int32          `gorm:"column:id;primaryKey"`
	EmpresaID   int32          `gorm:"column:empresa_id;not null"`
	Nombre      string         `gorm:"column:nombre;not null"`
	Descripcion *string        `gorm:"column:descripcion"`
	Activo      bool           `gorm:"column:activo;not null"`
	CreatedAt   time.Time      `gorm:"column:created_at"`
	UpdatedAt   time.Time      `gorm:"column:updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (TipoCampana) TableName() string { return "crm_tipos_campana" }

type Campana struct {
	ID            int64          `gorm:"column:id;primaryKey"`
	EmpresaID     int32          `gorm:"column:empresa_id;not null"`
	TipoCampanaID int32          `gorm:"column:tipo_campana_id;not null"`
	OrigenID      *int32         `gorm:"column:origen_id"`
	MonedaID      *int32         `gorm:"column:moneda_id"`
	Nombre        string         `gorm:"column:nombre;not null"`
	Descripcion   *string        `gorm:"column:descripcion"`
	Objetivo      *string        `gorm:"column:objetivo"`
	FechaInicio   time.Time      `gorm:"column:fecha_inicio;type:date;not null"`
	FechaFin      *time.Time     `gorm:"column:fecha_fin;type:date"`
	Presupuesto   *float64       `gorm:"column:presupuesto"`
	Estado        string         `gorm:"column:estado;not null"`
	CreatedBy     int64          `gorm:"column:created_by;not null"`
	CreatedAt     time.Time      `gorm:"column:created_at"`
	UpdatedAt     time.Time      `gorm:"column:updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (Campana) TableName() string { return "crm_campanas" }

type CampanaLead struct {
	ID                 int64     `gorm:"column:id;primaryKey"`
	EmpresaID          int32     `gorm:"column:empresa_id;not null"`
	CampanaID          int64     `gorm:"column:campana_id;not null"`
	LeadID             int64     `gorm:"column:lead_id;not null"`
	FechaAsociacion    time.Time `gorm:"column:fecha_asociacion;not null"`
	TipoAtribucion     *string   `gorm:"column:tipo_atribucion"`
	UTMSource          *string   `gorm:"column:utm_source"`
	UTMMedium          *string   `gorm:"column:utm_medium"`
	UTMCampaign        *string   `gorm:"column:utm_campaign"`
	UTMContent         *string   `gorm:"column:utm_content"`
	UTMTerm            *string   `gorm:"column:utm_term"`
	LandingPage        *string   `gorm:"column:landing_page"`
	ReferrerURL        *string   `gorm:"column:referrer_url"`
	ExternalCampaignID *string   `gorm:"column:external_campaign_id"`
	ExternalAdID       *string   `gorm:"column:external_ad_id"`
	CreatedAt          time.Time `gorm:"column:created_at"`
}

func (CampanaLead) TableName() string { return "crm_campana_leads" }

type CampanaGasto struct {
	ID                           int64          `gorm:"column:id;primaryKey"`
	EmpresaID                    int32          `gorm:"column:empresa_id;not null"`
	CampanaID                    int64          `gorm:"column:campana_id;not null"`
	Fecha                        time.Time      `gorm:"column:fecha;type:date;not null"`
	Concepto                     string         `gorm:"column:concepto;not null"`
	CategoriaGasto               *string        `gorm:"column:categoria_gasto"`
	Valor                        float64        `gorm:"column:valor;not null"`
	MonedaID                     int32          `gorm:"column:moneda_id;not null"`
	ProveedorTerceroEmpresaRolID *int64         `gorm:"column:proveedor_tercero_empresa_rol_id"`
	CreatedBy                    int64          `gorm:"column:created_by;not null"`
	CreatedAt                    time.Time      `gorm:"column:created_at"`
	UpdatedAt                    time.Time      `gorm:"column:updated_at"`
	DeletedAt                    gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (CampanaGasto) TableName() string { return "crm_campana_gastos" }
