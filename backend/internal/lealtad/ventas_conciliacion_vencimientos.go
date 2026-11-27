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

// =====================================================================
// 1. Consulta de cuenta de lealtad por cliente (para Ventas)
// =====================================================================

type CuentaClienteResponse struct {
	TieneCuenta         bool    `json:"tiene_cuenta"`
	CuentaID            int64   `json:"cuenta_id,omitempty"`
	ProgramaID          int32   `json:"programa_id,omitempty"`
	ProgramaCodigo      string  `json:"programa_codigo,omitempty"`
	ProgramaNombre      string  `json:"programa_nombre,omitempty"`
	SaldoPuntos         int64   `json:"saldo_puntos"`
	ValorPuntoRedencion float64 `json:"valor_punto_redencion"`
	ValorMaximoDescuento float64 `json:"valor_maximo_descuento"`
	MonedaCodigo        string  `json:"moneda_codigo,omitempty"`
	MonedaSimbolo       string  `json:"moneda_simbolo,omitempty"`
	PermiteAcumular     bool    `json:"permite_acumular"`
	PermiteRedimir      bool    `json:"permite_redimir"`
}

func (s *Service) ConsultarCuentaPorCliente(ctx context.Context, empresaID int32, terceroID int64) (*CuentaClienteResponse, error) {
	tercero, err := s.repo.FindTerceroByID(ctx, empresaID, terceroID)
	if err != nil {
		if errors.Is(err, errTerceroNotFound) {
			return &CuentaClienteResponse{TieneCuenta: false}, nil
		}
		return nil, err
	}
	if strings.TrimSpace(tercero.NumeroIdentificacion) == "222222222222" {
		return &CuentaClienteResponse{TieneCuenta: false}, nil
	}

	paisID := int32(0)
	if tercero.PaisID != nil && *tercero.PaisID > 0 {
		paisID = *tercero.PaisID
	} else {
		paisID, err = s.repo.GetDefaultPaisID(ctx)
		if err != nil {
			return &CuentaClienteResponse{TieneCuenta: false}, nil
		}
	}

	miembro, err := s.repo.FindMiembroByDocumento(ctx, paisID, tercero.TipoIdentificacion, tercero.NumeroIdentificacion)
	if err != nil {
		if errors.Is(err, errMiembroNotFound) {
			return &CuentaClienteResponse{TieneCuenta: false}, nil
		}
		return nil, err
	}

	type row struct {
		CuentaID            int64
		ProgramaID          int32
		ProgramaCodigo      string
		ProgramaNombre      string
		SaldoPuntos         int64
		ValorPuntoRedencion float64
		MonedaCodigo        string
		MonedaSimbolo       string
		PermiteAcumular     bool
		PermiteRedimir      bool
	}
	var foundRows []row
	err = s.repo.db.WithContext(ctx).
		Table("lealtad_cuentas lc").
		Select(`
			lc.id AS cuenta_id, lp.id AS programa_id, lp.codigo AS programa_codigo, lp.nombre AS programa_nombre,
			lc.saldo_puntos, lp.valor_punto_redencion,
			m.codigo_iso AS moneda_codigo, m.simbolo AS moneda_simbolo,
			lpe.permite_acumular, lpe.permite_redimir
		`).
		Joins("JOIN lealtad_programa_empresas lpe ON lpe.programa_id = lc.programa_id AND lpe.empresa_id = ? AND lpe.activo = TRUE", empresaID).
		Joins("JOIN lealtad_programas lp ON lp.id = lc.programa_id AND lp.estado = ? AND lp.deleted_at IS NULL", EstadoActivo).
		Joins("JOIN monedas m ON m.id = lp.moneda_referencia_id").
		Where("lc.miembro_id = ? AND lc.estado = ?", miembro.ID, EstadoCuentaActiva).
		Order("lpe.es_programa_principal DESC, lp.es_sistema DESC, lc.id ASC").
		Limit(1).
		Scan(&foundRows).Error
	if err != nil {
		return nil, err
	}
	if len(foundRows) == 0 {
		return &CuentaClienteResponse{TieneCuenta: false}, nil
	}
	found := foundRows[0]

	return &CuentaClienteResponse{
		TieneCuenta:          true,
		CuentaID:             found.CuentaID,
		ProgramaID:           found.ProgramaID,
		ProgramaCodigo:       found.ProgramaCodigo,
		ProgramaNombre:       found.ProgramaNombre,
		SaldoPuntos:          found.SaldoPuntos,
		ValorPuntoRedencion:  found.ValorPuntoRedencion,
		ValorMaximoDescuento: math.Round(float64(found.SaldoPuntos)*found.ValorPuntoRedencion*100) / 100,
		MonedaCodigo:         found.MonedaCodigo,
		MonedaSimbolo:        found.MonedaSimbolo,
		PermiteAcumular:      found.PermiteAcumular,
		PermiteRedimir:       found.PermiteRedimir,
	}, nil
}

// =====================================================================
// 2. Redención y Reverso transaccional desde Ventas
// =====================================================================

type RedimirVentaParams struct {
	EmpresaID       int32
	SucursalID      int32
	ClienteID       int64
	UsuarioID       int64
	VentaID         int64
	Consecutivo     string
	PuntosRedimir   int64
	TotalAntesDesc  float64
}

// RedimirPorVentaTx descuenta puntos del cliente dentro de la transacción de la venta
// y retorna el monto descontado en dinero.
func RedimirPorVentaTx(ctx context.Context, tx *gorm.DB, params RedimirVentaParams) (float64, error) {
	if params.PuntosRedimir <= 0 {
		return 0, nil
	}

	var tercero TerceroClienteRef
	if err := tx.WithContext(ctx).
		Where("id = ? AND empresa_id = ? AND deleted_at IS NULL", params.ClienteID, params.EmpresaID).
		First(&tercero).Error; err != nil {
		return 0, ErrTerceroNoEncontrado
	}
	if strings.TrimSpace(tercero.NumeroIdentificacion) == "222222222222" {
		return 0, httpx.NewError(http.StatusBadRequest, "consumidor_final_sin_lealtad", "Consumidor Final no puede redimir puntos de lealtad")
	}

	paisID := int32(0)
	if tercero.PaisID != nil && *tercero.PaisID > 0 {
		paisID = *tercero.PaisID
	} else {
		var p PaisRef
		if err := tx.WithContext(ctx).
			Order("CASE WHEN iso_alpha_2 = 'CO' THEN 0 ELSE 1 END, id ASC").
			First(&p).Error; err != nil {
			return 0, ErrMiembroNoEncontrado
		}
		paisID = p.ID
	}

	var miembro Miembro
	if err := tx.WithContext(ctx).
		Where("pais_id = ? AND upper(tipo_identificacion) = ? AND lower(trim(numero_identificacion)) = ?",
			paisID, strings.ToUpper(strings.TrimSpace(tercero.TipoIdentificacion)), strings.ToLower(strings.TrimSpace(tercero.NumeroIdentificacion))).
		First(&miembro).Error; err != nil {
		return 0, ErrMiembroNoEncontrado
	}

	var cuenta Cuenta
	err := tx.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE", Table: clause.Table{Name: "lealtad_cuentas"}}).
		Joins("JOIN lealtad_programa_empresas lpe ON lpe.programa_id = lealtad_cuentas.programa_id AND lpe.empresa_id = ? AND lpe.activo = TRUE AND lpe.permite_redimir = TRUE", params.EmpresaID).
		Joins("JOIN lealtad_programas lp ON lp.id = lealtad_cuentas.programa_id AND lp.estado = ? AND lp.deleted_at IS NULL", EstadoActivo).
		Where("lealtad_cuentas.miembro_id = ? AND lealtad_cuentas.estado = ?", miembro.ID, EstadoCuentaActiva).
		Order("lpe.es_programa_principal DESC, lp.es_sistema DESC, lealtad_cuentas.id ASC").
		First(&cuenta).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, ErrRedencionNoPermitida
		}
		return 0, err
	}

	if cuenta.SaldoPuntos < params.PuntosRedimir {
		return 0, ErrSaldoInsuficiente
	}

	var programa Programa
	if err := tx.WithContext(ctx).Where("id = ?", cuenta.ProgramaID).First(&programa).Error; err != nil {
		return 0, err
	}

	montoDescontado := math.Round(float64(params.PuntosRedimir)*programa.ValorPuntoRedencion*10000) / 10000
	if montoDescontado > params.TotalAntesDesc {
		return 0, httpx.NewError(http.StatusBadRequest, "descuento_excede_total", "el descuento por puntos no puede superar el total de la venta")
	}

	now := time.Now().UTC()
	saldoPosterior := cuenta.SaldoPuntos - params.PuntosRedimir
	if err := tx.WithContext(ctx).
		Model(&Cuenta{}).
		Where("id = ?", cuenta.ID).
		Updates(map[string]any{
			"saldo_puntos": saldoPosterior,
			"updated_at":   now,
		}).Error; err != nil {
		return 0, err
	}

	var sucPtr *int32
	if params.SucursalID > 0 {
		s := params.SucursalID
		sucPtr = &s
	}
	ventaID := params.VentaID
	obs := fmt.Sprintf("Redención aplicada a venta %s", params.Consecutivo)

	mov := Transaccion{
		CuentaID:        cuenta.ID,
		ProgramaID:      programa.ID,
		EmpresaID:       params.EmpresaID,
		SucursalID:      sucPtr,
		VentaID:         &ventaID,
		TipoTransaccion: TipoTxRedencion,
		Puntos:          -params.PuntosRedimir,
		SaldoPosterior:  saldoPosterior,
		Observaciones:   &obs,
		CreatedBy:       params.UsuarioID,
		CreatedAt:       now,
	}
	if err := tx.WithContext(ctx).Create(&mov).Error; err != nil {
		return 0, err
	}

	red := Redencion{
		CuentaID:        cuenta.ID,
		TransaccionID:   mov.ID,
		EmpresaID:       params.EmpresaID,
		SucursalID:      sucPtr,
		VentaID:         &ventaID,
		PuntosRedimidos: params.PuntosRedimir,
		MontoDescontado: montoDescontado,
		MonedaID:        programa.MonedaReferenciaID,
		Estado:          EstadoRedencionAplicada,
		Observaciones:   &obs,
		CreatedBy:       params.UsuarioID,
		CreatedAt:       now,
	}
	if err := tx.WithContext(ctx).Create(&red).Error; err != nil {
		return 0, err
	}

	return montoDescontado, nil
}

type RevertirVentaParams struct {
	EmpresaID   int32
	SucursalID  int32
	UsuarioID   int64
	VentaID     int64
	Consecutivo string
	Motivo      string
}

// RevertirPorVentaTx revierte de manera idempotente cualquier acumulación o redención
// asociada a una venta cuando esta es anulada o eliminada.
func RevertirPorVentaTx(ctx context.Context, tx *gorm.DB, params RevertirVentaParams) error {
	if params.VentaID <= 0 {
		return nil
	}

	// Idempotencia: si ya existe al menos un REVERSO para esta venta, no repetir
	var reversosExistentes int64
	if err := tx.WithContext(ctx).
		Model(&Transaccion{}).
		Where("venta_id = ? AND tipo_transaccion = ?", params.VentaID, TipoTxReverso).
		Count(&reversosExistentes).Error; err != nil {
		return err
	}
	if reversosExistentes > 0 {
		return nil
	}

	var origs []Transaccion
	if err := tx.WithContext(ctx).
		Where("venta_id = ? AND tipo_transaccion IN ?", params.VentaID, []string{TipoTxAcumulacion, TipoTxRedencion}).
		Order("id ASC").
		Find(&origs).Error; err != nil {
		return err
	}
	if len(origs) == 0 {
		return nil
	}

	now := time.Now().UTC()
	var sucPtr *int32
	if params.SucursalID > 0 {
		s := params.SucursalID
		sucPtr = &s
	}

	// Agrupar delta neto a revertir por cuenta (inverso de los puntos originales)
	deltaPorCuenta := make(map[int64]int64)
	progPorCuenta := make(map[int64]int32)
	for _, t := range origs {
		deltaPorCuenta[t.CuentaID] += -t.Puntos
		progPorCuenta[t.CuentaID] = t.ProgramaID
	}

	// Marcar redenciones asociadas como ANULADA
	if err := tx.WithContext(ctx).
		Model(&Redencion{}).
		Where("venta_id = ? AND estado = ?", params.VentaID, EstadoRedencionAplicada).
		Update("estado", EstadoRedencionAnulada).Error; err != nil {
		return err
	}

	for cuentaID, delta := range deltaPorCuenta {
		if delta == 0 {
			continue
		}
		var cuenta Cuenta
		if err := tx.WithContext(ctx).
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", cuentaID).
			First(&cuenta).Error; err != nil {
			return err
		}

		nuevoSaldo := cuenta.SaldoPuntos + delta
		puntosReverso := delta
		if nuevoSaldo < 0 {
			// Evitar violar CHECK (saldo_puntos >= 0) si el cliente ya gastó parte de los puntos
			puntosReverso = -cuenta.SaldoPuntos
			nuevoSaldo = 0
		}
		if puntosReverso == 0 {
			continue
		}

		if err := tx.WithContext(ctx).
			Model(&Cuenta{}).
			Where("id = ?", cuenta.ID).
			Updates(map[string]any{
				"saldo_puntos": nuevoSaldo,
				"updated_at":   now,
			}).Error; err != nil {
			return err
		}

		obs := fmt.Sprintf("Reverso automático por %s de venta %s", params.Motivo, params.Consecutivo)
		ventaID := params.VentaID
		rev := Transaccion{
			CuentaID:        cuenta.ID,
			ProgramaID:      progPorCuenta[cuentaID],
			EmpresaID:       params.EmpresaID,
			SucursalID:      sucPtr,
			VentaID:         &ventaID,
			TipoTransaccion: TipoTxReverso,
			Puntos:          puntosReverso,
			SaldoPosterior:  nuevoSaldo,
			Observaciones:   &obs,
			CreatedBy:       params.UsuarioID,
			CreatedAt:       now,
		}
		if err := tx.WithContext(ctx).Create(&rev).Error; err != nil {
			return err
		}
	}

	return nil
}

// =====================================================================
// 3. Reporte de Conciliación de Coalición
// =====================================================================

type ConciliacionEmpresaItem struct {
	EmpresaID          int32   `json:"empresa_id"`
	EmpresaNombre      string  `json:"empresa_nombre"`
	EmpresaNIT         string  `json:"empresa_nit"`
	PuntosEmitidos     int64   `json:"puntos_emitidos"`
	PuntosRedimidos    int64   `json:"puntos_redimidos"`
	PuntosReversados   int64   `json:"puntos_reversados"`
	PuntosVencidos     int64   `json:"puntos_vencidos"`
	BalanceNetoPuntos  int64   `json:"balance_neto_puntos"`
	ValorNeto          float64 `json:"valor_neto"`
	TotalTransacciones int64   `json:"total_transacciones"`
}

type ConciliacionProgramaResponse struct {
	ProgramaID          int32                     `json:"programa_id"`
	ProgramaCodigo      string                    `json:"programa_codigo"`
	ProgramaNombre      string                    `json:"programa_nombre"`
	ValorPuntoRedencion float64                   `json:"valor_punto_redencion"`
	MonedaCodigo        string                    `json:"moneda_codigo,omitempty"`
	MonedaSimbolo       string                    `json:"moneda_simbolo,omitempty"`
	TotalEmitidos       int64                     `json:"total_emitidos"`
	TotalRedimidos      int64                     `json:"total_redimidos"`
	TotalVencidos       int64                     `json:"total_vencidos"`
	Empresas            []ConciliacionEmpresaItem `json:"empresas"`
}

func (s *Service) ObtenerConciliacionPrograma(ctx context.Context, claims *auth.Claims, programaID int32, fechaDesde, fechaHasta string) (*ConciliacionProgramaResponse, error) {
	p, err := s.repo.FindProgramaByID(ctx, programaID)
	if err != nil {
		if errors.Is(err, errProgramaNotFound) {
			return nil, ErrProgramaNoEncontrado
		}
		return nil, err
	}
	isGlobalAdmin, err := s.isPlatformLoyaltyAdmin(ctx, claims)
	if err != nil {
		return nil, err
	}
	if !s.canManagePrograma(claims, isGlobalAdmin, *p) {
		if p.Alcance == AlcanceGlobal {
			return nil, ErrSinPermisoPlanGlobal
		}
		return nil, ErrSinPermisoPlanEmpresa
	}

	q := s.repo.db.WithContext(ctx).
		Table("lealtad_transacciones lt").
		Select(`
			lt.empresa_id,
			COALESCE(NULLIF(trim(e.razon_social), ''), e.nombre_legal) AS empresa_nombre,
			e.identificador_fiscal AS empresa_nit,
			COALESCE(SUM(CASE WHEN lt.tipo_transaccion IN ('ACUMULACION', 'AJUSTE_POSITIVO') THEN lt.puntos ELSE 0 END), 0) AS puntos_emitidos,
			COALESCE(SUM(CASE WHEN lt.tipo_transaccion = 'REDENCION' THEN -lt.puntos ELSE 0 END), 0) AS puntos_redimidos,
			COALESCE(SUM(CASE WHEN lt.tipo_transaccion = 'REVERSO' THEN lt.puntos ELSE 0 END), 0) AS puntos_reversados,
			COALESCE(SUM(CASE WHEN lt.tipo_transaccion = 'VENCIMIENTO' THEN -lt.puntos ELSE 0 END), 0) AS puntos_vencidos,
			COUNT(*) AS total_transacciones
		`).
		Joins("JOIN empresas e ON e.id = lt.empresa_id").
		Where("lt.programa_id = ?", p.ID)

	if strings.TrimSpace(fechaDesde) != "" {
		if t, err := time.Parse("2006-01-02", strings.TrimSpace(fechaDesde)); err == nil {
			q = q.Where("lt.created_at >= ?", t.UTC())
		}
	}
	if strings.TrimSpace(fechaHasta) != "" {
		if t, err := time.Parse("2006-01-02", strings.TrimSpace(fechaHasta)); err == nil {
			q = q.Where("lt.created_at < ?", t.UTC().AddDate(0, 0, 1))
		}
	}

	var rows []ConciliacionEmpresaItem
	if err := q.Group("lt.empresa_id, e.razon_social, e.nombre_legal, e.identificador_fiscal").
		Order("puntos_emitidos DESC, empresa_nombre ASC").
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	resp := &ConciliacionProgramaResponse{
		ProgramaID:          p.ID,
		ProgramaCodigo:      p.Codigo,
		ProgramaNombre:      p.Nombre,
		ValorPuntoRedencion: p.ValorPuntoRedencion,
		Empresas:            make([]ConciliacionEmpresaItem, 0, len(rows)),
	}
	if p.Moneda != nil {
		resp.MonedaCodigo = p.Moneda.CodigoISO
		resp.MonedaSimbolo = p.Moneda.Simbolo
	}

	for _, r := range rows {
		r.BalanceNetoPuntos = r.PuntosEmitidos - r.PuntosRedimidos
		r.ValorNeto = math.Round(float64(r.BalanceNetoPuntos)*p.ValorPuntoRedencion*100) / 100
		resp.TotalEmitidos += r.PuntosEmitidos
		resp.TotalRedimidos += r.PuntosRedimidos
		resp.TotalVencidos += r.PuntosVencidos
		resp.Empresas = append(resp.Empresas, r)
	}

	return resp, nil
}

// =====================================================================
// 4. Procesamiento de vencimiento de puntos (VENCIMIENTO)
// =====================================================================

type ProcesarVencimientosResponse struct {
	ProgramaID          int32     `json:"programa_id"`
	CuentasAfectadas    int       `json:"cuentas_afectadas"`
	LotesProcesados     int       `json:"lotes_procesados"`
	PuntosVencidosTotal int64     `json:"puntos_vencidos_total"`
	ProcesadoAt         time.Time `json:"procesado_at"`
}

func (s *Service) ProcesarVencimientosPrograma(ctx context.Context, claims *auth.Claims, programaID int32) (*ProcesarVencimientosResponse, error) {
	p, err := s.repo.FindProgramaByID(ctx, programaID)
	if err != nil {
		if errors.Is(err, errProgramaNotFound) {
			return nil, ErrProgramaNoEncontrado
		}
		return nil, err
	}
	isGlobalAdmin, err := s.isPlatformLoyaltyAdmin(ctx, claims)
	if err != nil {
		return nil, err
	}
	if !s.canManagePrograma(claims, isGlobalAdmin, *p) {
		if p.Alcance == AlcanceGlobal {
			return nil, ErrSinPermisoPlanGlobal
		}
		return nil, ErrSinPermisoPlanEmpresa
	}

	now := time.Now().UTC()
	out := &ProcesarVencimientosResponse{
		ProgramaID:  p.ID,
		ProcesadoAt: now,
	}

	err = s.repo.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		type loteVencido struct {
			ID       int64
			CuentaID int64
			Puntos   int64
		}
		var lotes []loteVencido
		if err := tx.Table("lealtad_transacciones").
			Select("id, cuenta_id, puntos").
			Where("programa_id = ? AND puntos > 0 AND fecha_vencimiento IS NOT NULL AND fecha_vencimiento <= ? AND vencimiento_procesado = FALSE", p.ID, now).
			Order("cuenta_id ASC, id ASC").
			Scan(&lotes).Error; err != nil {
			return err
		}
		if len(lotes) == 0 {
			return nil
		}

		out.LotesProcesados = len(lotes)
		porCuenta := make(map[int64]int64)
		loteIDs := make([]int64, 0, len(lotes))
		for _, l := range lotes {
			porCuenta[l.CuentaID] += l.Puntos
			loteIDs = append(loteIDs, l.ID)
		}

		if err := tx.Table("lealtad_transacciones").
			Where("id IN ?", loteIDs).
			Update("vencimiento_procesado", true).Error; err != nil {
			return err
		}

		for cuentaID, puntosExpirados := range porCuenta {
			var cuenta Cuenta
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
				Where("id = ?", cuentaID).
				First(&cuenta).Error; err != nil {
				return err
			}

			descontar := puntosExpirados
			if descontar > cuenta.SaldoPuntos {
				descontar = cuenta.SaldoPuntos
			}
			if descontar <= 0 {
				continue
			}

			nuevoSaldo := cuenta.SaldoPuntos - descontar
			if err := tx.Model(&Cuenta{}).Where("id = ?", cuenta.ID).Updates(map[string]any{
				"saldo_puntos": nuevoSaldo,
				"updated_at":   now,
			}).Error; err != nil {
				return err
			}

			obs := fmt.Sprintf("Vencimiento automático de %d puntos expirados", descontar)
			mov := Transaccion{
				CuentaID:        cuenta.ID,
				ProgramaID:      p.ID,
				EmpresaID:       claims.EmpresaID,
				TipoTransaccion: TipoTxVencimiento,
				Puntos:          -descontar,
				SaldoPosterior:  nuevoSaldo,
				Observaciones:   &obs,
				CreatedBy:       claims.Sub,
				CreatedAt:       now,
			}
			if err := tx.Create(&mov).Error; err != nil {
				return err
			}
			out.CuentasAfectadas++
			out.PuntosVencidosTotal += descontar
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return out, nil
}

// =====================================================================
// Handlers
// =====================================================================

func (h *Handler) consultarCuentaCliente(c *gin.Context) {
	claims, ok := auth.ClaimsFromContext(c)
	if !ok {
		httpx.RespondError(c, auth.ErrNoAutenticado)
		return
	}
	terceroID, err := parseID64Param(c, "terceroId")
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	resp, err := h.svc.ConsultarCuentaPorCliente(c.Request.Context(), claims.EmpresaID, terceroID)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) obtenerConciliacion(c *gin.Context) {
	claims, ok := auth.ClaimsFromContext(c)
	if !ok {
		httpx.RespondError(c, auth.ErrNoAutenticado)
		return
	}
	id, err := parseID32(c)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	resp, err := h.svc.ObtenerConciliacionPrograma(
		c.Request.Context(),
		claims,
		id,
		c.Query("fecha_desde"),
		c.Query("fecha_hasta"),
	)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) procesarVencimientos(c *gin.Context) {
	claims, ok := auth.ClaimsFromContext(c)
	if !ok {
		httpx.RespondError(c, auth.ErrNoAutenticado)
		return
	}
	id, err := parseID32(c)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	resp, err := h.svc.ProcesarVencimientosPrograma(c.Request.Context(), claims, id)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}
