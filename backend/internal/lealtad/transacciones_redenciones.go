package lealtad

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"tesserak/backend/internal/auth"
	"tesserak/backend/internal/httpx"
)

const (
	TipoTxAcumulacion    = "ACUMULACION"
	TipoTxRedencion      = "REDENCION"
	TipoTxAjustePositivo = "AJUSTE_POSITIVO"
	TipoTxAjusteNegativo = "AJUSTE_NEGATIVO"
	TipoTxVencimiento    = "VENCIMIENTO"
	TipoTxReverso        = "REVERSO"

	EstadoRedencionAplicada = "APLICADA"
	EstadoRedencionAnulada  = "ANULADA"
)

var (
	ErrCuentaInactiva       = httpx.NewError(http.StatusConflict, "cuenta_inactiva", "la cuenta de lealtad no está en estado ACTIVA")
	ErrSaldoInsuficiente    = httpx.NewError(http.StatusConflict, "saldo_insuficiente", "la cuenta no tiene suficientes puntos disponibles para realizar esta operación")
	ErrAcumulacionNoPermitida = httpx.NewError(http.StatusForbidden, "acumulacion_no_permitida", "la adhesión de tu empresa en este plan no permite acumular puntos")
	ErrRedencionNoPermitida = httpx.NewError(http.StatusForbidden, "redencion_no_permitida", "la adhesión de tu empresa en este plan no permite redimir puntos")
	ErrSinReglaAplicable    = httpx.NewError(http.StatusUnprocessableEntity, "sin_regla_aplicable", "ninguna regla de acumulación activa otorga puntos para el monto indicado")
)

type Transaccion struct {
	ID               int64      `gorm:"column:id;primaryKey;autoIncrement"`
	CuentaID         int64      `gorm:"column:cuenta_id;not null"`
	ProgramaID       int32      `gorm:"column:programa_id;not null"`
	EmpresaID        int32      `gorm:"column:empresa_id;not null"`
	SucursalID       *int32     `gorm:"column:sucursal_id"`
	VentaID          *int64     `gorm:"column:venta_id"`
	ReglaID          *int32     `gorm:"column:regla_id"`
	TipoTransaccion  string     `gorm:"column:tipo_transaccion;not null"`
	Puntos           int64      `gorm:"column:puntos;not null"`
	SaldoPosterior   int64      `gorm:"column:saldo_posterior;not null"`
	FechaVencimiento *time.Time `gorm:"column:fecha_vencimiento"`
	Observaciones    *string    `gorm:"column:observaciones"`
	CreatedBy        int64      `gorm:"column:created_by;not null"`
	CreatedAt        time.Time  `gorm:"column:created_at;autoCreateTime"`

	Empresa *EmpresaRef `gorm:"foreignKey:EmpresaID;references:ID"`
	Regla   *Regla      `gorm:"foreignKey:ReglaID;references:ID"`
	Usuario *UsuarioRef `gorm:"foreignKey:CreatedBy;references:ID"`
}

func (Transaccion) TableName() string { return "lealtad_transacciones" }

type Redencion struct {
	ID              int64     `gorm:"column:id;primaryKey;autoIncrement"`
	CuentaID        int64     `gorm:"column:cuenta_id;not null"`
	TransaccionID   int64     `gorm:"column:transaccion_id;not null"`
	EmpresaID       int32     `gorm:"column:empresa_id;not null"`
	SucursalID      *int32    `gorm:"column:sucursal_id"`
	VentaID         *int64    `gorm:"column:venta_id"`
	PuntosRedimidos int64     `gorm:"column:puntos_redimidos;not null"`
	MontoDescontado float64   `gorm:"column:monto_descontado;not null"`
	MonedaID        int32     `gorm:"column:moneda_id;not null"`
	Estado          string    `gorm:"column:estado;not null"`
	Observaciones   *string   `gorm:"column:observaciones"`
	CreatedBy       int64     `gorm:"column:created_by;not null"`
	CreatedAt       time.Time `gorm:"column:created_at;autoCreateTime"`

	Empresa *EmpresaRef `gorm:"foreignKey:EmpresaID;references:ID"`
	Moneda  *MonedaRef  `gorm:"foreignKey:MonedaID;references:ID"`
	Cuenta  *Cuenta     `gorm:"foreignKey:CuentaID;references:ID"`
}

func (Redencion) TableName() string { return "lealtad_redenciones" }

// =====================================================================
// DTOs
// =====================================================================

type TransaccionResponse struct {
	ID               int64      `json:"id"`
	CuentaID         int64      `json:"cuenta_id"`
	ProgramaID       int32      `json:"programa_id"`
	EmpresaID        int32      `json:"empresa_id"`
	EmpresaNombre    string     `json:"empresa_nombre,omitempty"`
	SucursalID       *int32     `json:"sucursal_id,omitempty"`
	VentaID          *int64     `json:"venta_id,omitempty"`
	ReglaID          *int32     `json:"regla_id,omitempty"`
	ReglaNombre      string     `json:"regla_nombre,omitempty"`
	TipoTransaccion  string     `json:"tipo_transaccion"`
	Puntos           int64      `json:"puntos"`
	SaldoPosterior   int64      `json:"saldo_posterior"`
	FechaVencimiento *time.Time `json:"fecha_vencimiento,omitempty"`
	Observaciones    *string    `json:"observaciones,omitempty"`
	CreatedBy        int64      `json:"created_by"`
	CreatedByNombre  string     `json:"created_by_nombre,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
}

type RedencionResponse struct {
	ID              int64     `json:"id"`
	CuentaID        int64     `json:"cuenta_id"`
	TransaccionID   int64     `json:"transaccion_id"`
	EmpresaID       int32     `json:"empresa_id"`
	EmpresaNombre   string    `json:"empresa_nombre,omitempty"`
	SucursalID      *int32    `json:"sucursal_id,omitempty"`
	VentaID         *int64    `json:"venta_id,omitempty"`
	PuntosRedimidos int64     `json:"puntos_redimidos"`
	MontoDescontado float64   `json:"monto_descontado"`
	MonedaID        int32     `json:"moneda_id"`
	MonedaCodigo    string    `json:"moneda_codigo,omitempty"`
	MonedaSimbolo   string    `json:"moneda_simbolo,omitempty"`
	Estado          string    `json:"estado"`
	Observaciones   *string   `json:"observaciones,omitempty"`
	SaldoPosterior  int64     `json:"saldo_posterior,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
}

type AcumularPuntosRequest struct {
	MontoCompra   *float64 `json:"monto_compra"  binding:"omitnil,gt=0"`
	Puntos        *int64   `json:"puntos"`
	Observaciones *string  `json:"observaciones" binding:"omitnil,max=500"`
}

type RedimirPuntosRequest struct {
	PuntosRedimidos int64   `json:"puntos_redimidos" binding:"required,gt=0"`
	VentaID         *int64  `json:"venta_id"         binding:"omitnil,gt=0"`
	Observaciones   *string `json:"observaciones"    binding:"omitnil,max=500"`
}

type AcumularVentaParams struct {
	EmpresaID   int32
	SucursalID  int32
	ClienteID   int64
	UsuarioID   int64
	VentaID     int64
	Consecutivo string
	TotalVenta  float64
	Fecha       time.Time
}

func toTransaccionResponse(t Transaccion) TransaccionResponse {
	resp := TransaccionResponse{
		ID:               t.ID,
		CuentaID:         t.CuentaID,
		ProgramaID:       t.ProgramaID,
		EmpresaID:        t.EmpresaID,
		SucursalID:       t.SucursalID,
		VentaID:          t.VentaID,
		ReglaID:          t.ReglaID,
		TipoTransaccion:  t.TipoTransaccion,
		Puntos:           t.Puntos,
		SaldoPosterior:   t.SaldoPosterior,
		FechaVencimiento: t.FechaVencimiento,
		Observaciones:    t.Observaciones,
		CreatedBy:        t.CreatedBy,
		CreatedAt:        t.CreatedAt,
	}
	if t.Empresa != nil {
		if strings.TrimSpace(t.Empresa.RazonSocial) != "" {
			resp.EmpresaNombre = t.Empresa.RazonSocial
		} else {
			resp.EmpresaNombre = t.Empresa.NombreLegal
		}
	}
	if t.Regla != nil {
		resp.ReglaNombre = t.Regla.Nombre
	}
	if t.Usuario != nil {
		resp.CreatedByNombre = strings.TrimSpace(t.Usuario.Nombres + " " + t.Usuario.Apellidos)
		if resp.CreatedByNombre == "" {
			resp.CreatedByNombre = t.Usuario.NombreUsuario
		}
	}
	return resp
}

func toRedencionResponse(r Redencion, saldoPosterior int64) RedencionResponse {
	resp := RedencionResponse{
		ID:              r.ID,
		CuentaID:        r.CuentaID,
		TransaccionID:   r.TransaccionID,
		EmpresaID:       r.EmpresaID,
		SucursalID:      r.SucursalID,
		VentaID:         r.VentaID,
		PuntosRedimidos: r.PuntosRedimidos,
		MontoDescontado: r.MontoDescontado,
		MonedaID:        r.MonedaID,
		Estado:          r.Estado,
		Observaciones:   r.Observaciones,
		SaldoPosterior:  saldoPosterior,
		CreatedAt:       r.CreatedAt,
	}
	if r.Empresa != nil {
		if strings.TrimSpace(r.Empresa.RazonSocial) != "" {
			resp.EmpresaNombre = r.Empresa.RazonSocial
		} else {
			resp.EmpresaNombre = r.Empresa.NombreLegal
		}
	}
	if r.Moneda != nil {
		resp.MonedaCodigo = r.Moneda.CodigoISO
		resp.MonedaSimbolo = r.Moneda.Simbolo
	}
	return resp
}

// =====================================================================
// Motor de cálculo de puntos según lealtad_reglas
// =====================================================================

// CalcularPuntosPorReglas evalúa las reglas activas y vigentes para un monto de compra.
// Prioriza reglas específicas de la empresa (EmpresaID != nil) sobre reglas globales del programa.
func CalcularPuntosPorReglas(monto float64, reglas []Regla, ahora time.Time) (int64, *int32) {
	if monto <= 0 || len(reglas) == 0 {
		return 0, nil
	}

	diaAhora := ahora.UTC().Truncate(24 * time.Hour)
	var empReglas, globalReglas []Regla
	for _, r := range reglas {
		if !r.Activo || r.DeletedAt.Valid {
			continue
		}
		diaInicio := r.FechaInicio.UTC().Truncate(24 * time.Hour)
		if diaInicio.After(diaAhora) {
			continue
		}
		if r.FechaFin != nil {
			diaFin := r.FechaFin.UTC().Truncate(24 * time.Hour)
			if diaFin.Before(diaAhora) {
				continue
			}
		}
		if monto < r.CompraMinima {
			continue
		}
		if r.EmpresaID != nil {
			empReglas = append(empReglas, r)
		} else {
			globalReglas = append(globalReglas, r)
		}
	}

	candidatas := empReglas
	if len(candidatas) == 0 {
		candidatas = globalReglas
	}
	if len(candidatas) == 0 {
		return 0, nil
	}

	var basePuntos float64
	var reglaUsadaID *int32
	for _, r := range candidatas {
		switch r.TipoRegla {
		case TipoReglaPorMonto:
			if r.MontoBase > 0 {
				unidades := math.Floor(monto / r.MontoBase)
				pts := unidades * r.PuntosPorMonto
				if pts > 0 {
					basePuntos = pts
					id := r.ID
					reglaUsadaID = &id
				}
			}
		case TipoReglaBonoFijo:
			if r.PuntosPorMonto > 0 {
				basePuntos = r.PuntosPorMonto
				id := r.ID
				reglaUsadaID = &id
			}
		}
		if reglaUsadaID != nil {
			break
		}
	}

	if basePuntos <= 0 {
		return 0, nil
	}

	// Si existe una regla MULTIPLICADOR aplicable, multiplicar el resultado base
	for _, r := range candidatas {
		if r.TipoRegla == TipoReglaMultiplicador && r.PuntosPorMonto > 0 {
			basePuntos = basePuntos * r.PuntosPorMonto
			break
		}
	}

	total := int64(math.Floor(basePuntos))
	if total <= 0 {
		return 0, nil
	}
	return total, reglaUsadaID
}

// =====================================================================
// Integración transaccional con Ventas
// =====================================================================

// AcumularPorVentaTx acumula puntos automáticamente dentro de la misma transacción
// de una venta si el cliente es miembro de lealtad y la empresa participa activamente
// en un plan con acumulación habilitada. Si el cliente no es miembro o no aplica ninguna
// regla, retorna nil sin interrumpir la venta.
func AcumularPorVentaTx(ctx context.Context, tx *gorm.DB, params AcumularVentaParams) error {
	if params.TotalVenta <= 0 || params.ClienteID <= 0 || params.EmpresaID <= 0 {
		return nil
	}

	// 1. Obtener datos del cliente (tercero_maestro)
	var tercero TerceroClienteRef
	if err := tx.WithContext(ctx).
		Where("id = ? AND empresa_id = ? AND deleted_at IS NULL", params.ClienteID, params.EmpresaID).
		First(&tercero).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	// Consumidor Final no acumula puntos
	if strings.TrimSpace(tercero.NumeroIdentificacion) == "222222222222" {
		return nil
	}

	paisID := int32(0)
	if tercero.PaisID != nil && *tercero.PaisID > 0 {
		paisID = *tercero.PaisID
	} else {
		var p PaisRef
		if err := tx.WithContext(ctx).
			Order("CASE WHEN iso_alpha_2 = 'CO' THEN 0 ELSE 1 END, id ASC").
			First(&p).Error; err != nil {
			return nil
		}
		paisID = p.ID
	}

	// 2. Buscar identidad global del miembro (deben coincidir país, tipo y número de documento)
	var miembro Miembro
	if err := tx.WithContext(ctx).
		Where("pais_id = ? AND upper(tipo_identificacion) = ? AND lower(trim(numero_identificacion)) = ?",
			paisID, strings.ToUpper(strings.TrimSpace(tercero.TipoIdentificacion)), strings.ToLower(strings.TrimSpace(tercero.NumeroIdentificacion))).
		First(&miembro).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}

	// 3. Buscar cuenta activa del miembro en un plan adherido con acumulación habilitada
	var cuenta Cuenta
	err := tx.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE", Table: clause.Table{Name: "lealtad_cuentas"}}).
		Joins("JOIN lealtad_programa_empresas lpe ON lpe.programa_id = lealtad_cuentas.programa_id AND lpe.empresa_id = ? AND lpe.activo = TRUE AND lpe.permite_acumular = TRUE", params.EmpresaID).
		Joins("JOIN lealtad_programas lp ON lp.id = lealtad_cuentas.programa_id AND lp.estado = ? AND lp.deleted_at IS NULL", EstadoActivo).
		Where("lealtad_cuentas.miembro_id = ? AND lealtad_cuentas.estado = ?", miembro.ID, EstadoCuentaActiva).
		Order("lpe.es_programa_principal DESC, lp.es_sistema DESC, lealtad_cuentas.id ASC").
		First(&cuenta).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}

	var programa Programa
	if err := tx.WithContext(ctx).Where("id = ?", cuenta.ProgramaID).First(&programa).Error; err != nil {
		return err
	}

	// 4. Evaluar reglas activas del programa (empresa o globales)
	var reglas []Regla
	if err := tx.WithContext(ctx).
		Where("programa_id = ? AND activo = TRUE AND (empresa_id IS NULL OR empresa_id = ?)", programa.ID, params.EmpresaID).
		Order("prioridad ASC, id ASC").
		Find(&reglas).Error; err != nil {
		return err
	}

	ahora := params.Fecha.UTC()
	if ahora.IsZero() {
		ahora = time.Now().UTC()
	}
	puntos, reglaID := CalcularPuntosPorReglas(params.TotalVenta, reglas, ahora)
	if puntos <= 0 {
		return nil
	}

	// 5. Actualizar saldo y registrar movimiento inmutable en el libro mayor
	nuevoSaldo := cuenta.SaldoPuntos + puntos
	if err := tx.WithContext(ctx).
		Model(&Cuenta{}).
		Where("id = ?", cuenta.ID).
		Updates(map[string]any{
			"saldo_puntos": nuevoSaldo,
			"updated_at":   time.Now().UTC(),
		}).Error; err != nil {
		return err
	}

	var fechaVenc *time.Time
	if programa.DiasVencimiento != nil && *programa.DiasVencimiento > 0 {
		v := ahora.AddDate(0, 0, int(*programa.DiasVencimiento))
		fechaVenc = &v
	}
	var sucPtr *int32
	if params.SucursalID > 0 {
		s := params.SucursalID
		sucPtr = &s
	}
	ventaID := params.VentaID
	obs := fmt.Sprintf("Acumulación automática por venta %s", params.Consecutivo)

	mov := Transaccion{
		CuentaID:         cuenta.ID,
		ProgramaID:       programa.ID,
		EmpresaID:        params.EmpresaID,
		SucursalID:       sucPtr,
		VentaID:          &ventaID,
		ReglaID:          reglaID,
		TipoTransaccion:  TipoTxAcumulacion,
		Puntos:           puntos,
		SaldoPosterior:   nuevoSaldo,
		FechaVencimiento: fechaVenc,
		Observaciones:    &obs,
		CreatedBy:        params.UsuarioID,
		CreatedAt:        time.Now().UTC(),
	}
	return tx.WithContext(ctx).Create(&mov).Error
}

// =====================================================================
// Repository & Service (Operación de transacciones y redenciones)
// =====================================================================

func (r *Repository) ListTransaccionesCuenta(ctx context.Context, cuentaID int64) ([]Transaccion, error) {
	var rows []Transaccion
	err := r.db.WithContext(ctx).
		Preload("Empresa").
		Preload("Regla").
		Preload("Usuario").
		Where("cuenta_id = ?", cuentaID).
		Order("created_at DESC, id DESC").
		Limit(100).
		Find(&rows).Error
	return rows, err
}

func (s *Service) ListarTransaccionesCuenta(ctx context.Context, claims *auth.Claims, cuentaID int64) ([]TransaccionResponse, error) {
	if _, err := s.repo.FindCuentaForEmpresa(ctx, claims.EmpresaID, cuentaID); err != nil {
		if errors.Is(err, errCuentaNotFound) {
			return nil, ErrCuentaNoEncontrada
		}
		return nil, err
	}
	rows, err := s.repo.ListTransaccionesCuenta(ctx, cuentaID)
	if err != nil {
		return nil, err
	}
	out := make([]TransaccionResponse, 0, len(rows))
	for _, row := range rows {
		out = append(out, toTransaccionResponse(row))
	}
	return out, nil
}

func (s *Service) AcumularOAjustarPuntos(ctx context.Context, claims *auth.Claims, cuentaID int64, req AcumularPuntosRequest) (*TransaccionResponse, error) {
	if req.MontoCompra == nil && (req.Puntos == nil || *req.Puntos == 0) {
		return nil, httpx.NewError(http.StatusBadRequest, "validacion", "debe indicar monto_compra o una cantidad de puntos distinta de cero")
	}

	var movCreado Transaccion
	err := s.repo.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var cuenta Cuenta
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", cuentaID).
			First(&cuenta).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrCuentaNoEncontrada
			}
			return err
		}

		// Validar adhesión activa de la empresa al programa de la cuenta
		var adh ProgramaEmpresa
		if err := tx.Where("programa_id = ? AND empresa_id = ? AND activo = TRUE", cuenta.ProgramaID, claims.EmpresaID).
			First(&adh).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrCuentaNoEncontrada
			}
			return err
		}
		if !adh.PermiteAcumular {
			return ErrAcumulacionNoPermitida
		}
		if cuenta.Estado != EstadoCuentaActiva {
			return ErrCuentaInactiva
		}

		var programa Programa
		if err := tx.Where("id = ? AND deleted_at IS NULL", cuenta.ProgramaID).First(&programa).Error; err != nil {
			return ErrProgramaNoEncontrado
		}

		now := time.Now().UTC()
		var puntos int64
		var tipoTx string
		var reglaID *int32
		var obs *string

		if req.Observaciones != nil && strings.TrimSpace(*req.Observaciones) != "" {
			v := strings.TrimSpace(*req.Observaciones)
			obs = &v
		}

		if req.MontoCompra != nil {
			var reglas []Regla
			if err := tx.Where("programa_id = ? AND activo = TRUE AND (empresa_id IS NULL OR empresa_id = ?)", programa.ID, claims.EmpresaID).
				Order("prioridad ASC, id ASC").
				Find(&reglas).Error; err != nil {
				return err
			}
			puntos, reglaID = CalcularPuntosPorReglas(*req.MontoCompra, reglas, now)
			if puntos <= 0 {
				return ErrSinReglaAplicable
			}
			tipoTx = TipoTxAcumulacion
			if obs == nil {
				v := fmt.Sprintf("Acumulación por compra de %.2f", *req.MontoCompra)
				obs = &v
			}
		} else {
			puntos = *req.Puntos
			if puntos > 0 {
				tipoTx = TipoTxAjustePositivo
			} else {
				tipoTx = TipoTxAjusteNegativo
			}
		}

		nuevoSaldo := cuenta.SaldoPuntos + puntos
		if nuevoSaldo < 0 {
			return ErrSaldoInsuficiente
		}

		if err := tx.Model(&Cuenta{}).Where("id = ?", cuenta.ID).Updates(map[string]any{
			"saldo_puntos": nuevoSaldo,
			"updated_at":   now,
		}).Error; err != nil {
			return err
		}

		var fechaVenc *time.Time
		if puntos > 0 && programa.DiasVencimiento != nil && *programa.DiasVencimiento > 0 {
			v := now.AddDate(0, 0, int(*programa.DiasVencimiento))
			fechaVenc = &v
		}
		var sucPtr *int32
		if claims.SucursalID > 0 {
			s := claims.SucursalID
			sucPtr = &s
		}

		mov := Transaccion{
			CuentaID:         cuenta.ID,
			ProgramaID:       programa.ID,
			EmpresaID:        claims.EmpresaID,
			SucursalID:       sucPtr,
			ReglaID:          reglaID,
			TipoTransaccion:  tipoTx,
			Puntos:           puntos,
			SaldoPosterior:   nuevoSaldo,
			FechaVencimiento: fechaVenc,
			Observaciones:    obs,
			CreatedBy:        claims.Sub,
			CreatedAt:        now,
		}
		if err := tx.Create(&mov).Error; err != nil {
			return err
		}
		return tx.Preload("Empresa").Preload("Regla").Preload("Usuario").Where("id = ?", mov.ID).First(&movCreado).Error
	})
	if err != nil {
		return nil, err
	}
	resp := toTransaccionResponse(movCreado)
	return &resp, nil
}

func (s *Service) RedimirPuntos(ctx context.Context, claims *auth.Claims, cuentaID int64, req RedimirPuntosRequest) (*RedencionResponse, error) {
	var redCreada Redencion
	var saldoPosterior int64

	err := s.repo.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var cuenta Cuenta
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", cuentaID).
			First(&cuenta).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrCuentaNoEncontrada
			}
			return err
		}

		var adh ProgramaEmpresa
		if err := tx.Where("programa_id = ? AND empresa_id = ? AND activo = TRUE", cuenta.ProgramaID, claims.EmpresaID).
			First(&adh).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrCuentaNoEncontrada
			}
			return err
		}
		if !adh.PermiteRedimir {
			return ErrRedencionNoPermitida
		}
		if cuenta.Estado != EstadoCuentaActiva {
			return ErrCuentaInactiva
		}
		if cuenta.SaldoPuntos < req.PuntosRedimidos {
			return ErrSaldoInsuficiente
		}

		var programa Programa
		if err := tx.Where("id = ? AND deleted_at IS NULL", cuenta.ProgramaID).First(&programa).Error; err != nil {
			return ErrProgramaNoEncontrado
		}

		now := time.Now().UTC()
		saldoPosterior = cuenta.SaldoPuntos - req.PuntosRedimidos
		montoDescontado := math.Round(float64(req.PuntosRedimidos)*programa.ValorPuntoRedencion*10000) / 10000

		if err := tx.Model(&Cuenta{}).Where("id = ?", cuenta.ID).Updates(map[string]any{
			"saldo_puntos": saldoPosterior,
			"updated_at":   now,
		}).Error; err != nil {
			return err
		}

		var sucPtr *int32
		if claims.SucursalID > 0 {
			s := claims.SucursalID
			sucPtr = &s
		}
		var obs *string
		if req.Observaciones != nil && strings.TrimSpace(*req.Observaciones) != "" {
			v := strings.TrimSpace(*req.Observaciones)
			obs = &v
		}

		mov := Transaccion{
			CuentaID:        cuenta.ID,
			ProgramaID:      programa.ID,
			EmpresaID:       claims.EmpresaID,
			SucursalID:      sucPtr,
			VentaID:         req.VentaID,
			TipoTransaccion: TipoTxRedencion,
			Puntos:          -req.PuntosRedimidos,
			SaldoPosterior:  saldoPosterior,
			Observaciones:   obs,
			CreatedBy:       claims.Sub,
			CreatedAt:       now,
		}
		if err := tx.Create(&mov).Error; err != nil {
			return err
		}

		red := Redencion{
			CuentaID:        cuenta.ID,
			TransaccionID:   mov.ID,
			EmpresaID:       claims.EmpresaID,
			SucursalID:      sucPtr,
			VentaID:         req.VentaID,
			PuntosRedimidos: req.PuntosRedimidos,
			MontoDescontado: montoDescontado,
			MonedaID:        programa.MonedaReferenciaID,
			Estado:          EstadoRedencionAplicada,
			Observaciones:   obs,
			CreatedBy:       claims.Sub,
			CreatedAt:       now,
		}
		if err := tx.Create(&red).Error; err != nil {
			return err
		}

		return tx.Preload("Empresa").Preload("Moneda").Where("id = ?", red.ID).First(&redCreada).Error
	})
	if err != nil {
		return nil, err
	}

	resp := toRedencionResponse(redCreada, saldoPosterior)
	return &resp, nil
}

// =====================================================================
// Handlers
// =====================================================================

func (h *Handler) listarTransaccionesCuenta(c *gin.Context) {
	claims, ok := auth.ClaimsFromContext(c)
	if !ok {
		httpx.RespondError(c, auth.ErrNoAutenticado)
		return
	}
	id, err := parseID64Param(c, "id")
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	resp, err := h.svc.ListarTransaccionesCuenta(c.Request.Context(), claims, id)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}

func (h *Handler) acumularPuntosCuenta(c *gin.Context) {
	claims, ok := auth.ClaimsFromContext(c)
	if !ok {
		httpx.RespondError(c, auth.ErrNoAutenticado)
		return
	}
	id, err := parseID64Param(c, "id")
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	var req AcumularPuntosRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondError(c, httpx.BindError(err))
		return
	}
	resp, err := h.svc.AcumularOAjustarPuntos(c.Request.Context(), claims, id, req)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *Handler) redimirPuntosCuenta(c *gin.Context) {
	claims, ok := auth.ClaimsFromContext(c)
	if !ok {
		httpx.RespondError(c, auth.ErrNoAutenticado)
		return
	}
	id, err := parseID64Param(c, "id")
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	var req RedimirPuntosRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondError(c, httpx.BindError(err))
		return
	}
	resp, err := h.svc.RedimirPuntos(c.Request.Context(), claims, id, req)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}
