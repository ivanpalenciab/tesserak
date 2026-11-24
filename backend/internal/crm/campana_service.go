package crm

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"

	"tesserak/backend/internal/httpx"
)

var (
	ErrTipoCampanaNoEncontrado    = httpx.NewError(http.StatusNotFound, "tipo_campana_no_encontrado", "el tipo de campaña no existe")
	ErrCampanaNoEncontrada        = httpx.NewError(http.StatusNotFound, "campana_no_encontrada", "la campaña no existe")
	ErrCampanaLeadNoEncontrado    = httpx.NewError(http.StatusNotFound, "asociacion_no_encontrada", "la asociación entre campaña y lead no existe")
	ErrCampanaGastoNoEncontrado   = httpx.NewError(http.StatusNotFound, "gasto_no_encontrado", "el gasto de campaña no existe")
	ErrFechaInvalida              = httpx.NewError(http.StatusBadRequest, "fecha_invalida", "formato de fecha inválido, se espera YYYY-MM-DD")
	ErrRangoFechasInvalido        = httpx.NewError(http.StatusBadRequest, "rango_fechas_invalido", "la fecha de finalización no puede ser anterior a la fecha de inicio")
	ErrMonedaRequeridaPresupuesto = httpx.NewError(http.StatusBadRequest, "moneda_requerida", "cuando se define un presupuesto debe indicarse moneda_id")
	ErrTransicionEstadoInvalida   = httpx.NewError(http.StatusConflict, "transicion_estado_invalida", "la transición de estado solicitada no está permitida")
	ErrCampanaNoHabilitada        = httpx.NewError(http.StatusConflict, "campana_no_habilitada", "la campaña no está habilitada para asociar leads (debe estar PROGRAMADA, ACTIVA o PAUSADA)")
	ErrCampanaFinalizada          = httpx.NewError(http.StatusConflict, "campana_finalizada", "no se permiten modificaciones en una campaña finalizada")
	ErrAsociacionDuplicada        = httpx.NewError(http.StatusConflict, "asociacion_duplicada", "el lead ya se encuentra asociado a esta campaña")
	ErrFirstTouchDuplicado        = httpx.NewError(http.StatusConflict, "first_touch_duplicado", "el lead ya tiene una campaña registrada con atribución first_touch")
	ErrMonedaIncompatible         = httpx.NewError(http.StatusUnprocessableEntity, "moneda_incompatible", "la moneda del gasto debe coincidir con la moneda configurada en la campaña")
	ErrTipoCampanaInactivo        = httpx.NewError(http.StatusUnprocessableEntity, "tipo_campana_inactivo", "el tipo de campaña seleccionado no está activo")
)

// =====================================================================
// Tipos de campaña
// =====================================================================

func (s *Service) CrearTipoCampana(ctx context.Context, empresaID int32, req CrearTipoCampanaRequest) (*TipoCampanaResponse, error) {
	activo := true
	if req.Activo != nil {
		activo = *req.Activo
	}
	t := &TipoCampana{
		EmpresaID:   empresaID,
		Nombre:      strings.TrimSpace(req.Nombre),
		Descripcion: trimPtr(req.Descripcion),
		Activo:      activo,
	}
	if err := s.repo.CreateTipoCampana(ctx, t); err != nil {
		return nil, mapCampanaError(err)
	}
	resp := toTipoCampanaResponse(t)
	return &resp, nil
}

func (s *Service) ListarTiposCampana(ctx context.Context, empresaID int32, q ListarTiposCampanaQuery) ([]TipoCampanaResponse, error) {
	items, err := s.repo.ListTiposCampana(ctx, empresaID, q.Activo)
	if err != nil {
		return nil, err
	}
	out := make([]TipoCampanaResponse, 0, len(items))
	for i := range items {
		out = append(out, toTipoCampanaResponse(&items[i]))
	}
	return out, nil
}

func (s *Service) ObtenerTipoCampana(ctx context.Context, empresaID, id int32) (*TipoCampanaResponse, error) {
	t, err := s.repo.FindTipoCampanaByID(ctx, empresaID, id)
	if err != nil {
		return nil, mapCampanaError(err)
	}
	resp := toTipoCampanaResponse(t)
	return &resp, nil
}

func (s *Service) ActualizarTipoCampana(ctx context.Context, empresaID, id int32, req ActualizarTipoCampanaRequest) (*TipoCampanaResponse, error) {
	var res *TipoCampana
	err := s.repo.WithTx(ctx, func(tx *Repository) error {
		t, err := tx.FindTipoCampanaByID(ctx, empresaID, id)
		if err != nil {
			return err
		}
		cambios := map[string]any{}
		if req.Nombre != nil {
			cambios["nombre"] = strings.TrimSpace(*req.Nombre)
		}
		if req.Descripcion != nil {
			cambios["descripcion"] = trimPtr(req.Descripcion)
		}
		if req.Activo != nil {
			cambios["activo"] = *req.Activo
		}
		if err := tx.UpdateTipoCampana(ctx, t, cambios); err != nil {
			return err
		}
		res, err = tx.FindTipoCampanaByID(ctx, empresaID, id)
		return err
	})
	if err != nil {
		return nil, mapCampanaError(err)
	}
	resp := toTipoCampanaResponse(res)
	return &resp, nil
}

func (s *Service) EliminarTipoCampana(ctx context.Context, empresaID, id int32) error {
	return mapCampanaError(s.repo.SoftDeleteTipoCampana(ctx, empresaID, id))
}

// =====================================================================
// Campañas
// =====================================================================

func (s *Service) CrearCampana(ctx context.Context, empresaID int32, usuarioID int64, req CrearCampanaRequest) (*CampanaResponse, error) {
	fechaInicio, err := parseDate(req.FechaInicio)
	if err != nil {
		return nil, ErrFechaInvalida
	}
	var fechaFin *time.Time
	if req.FechaFin != nil && strings.TrimSpace(*req.FechaFin) != "" {
		ff, err := parseDate(*req.FechaFin)
		if err != nil {
			return nil, ErrFechaInvalida
		}
		if ff.Before(fechaInicio) {
			return nil, ErrRangoFechasInvalido
		}
		fechaFin = &ff
	}

	if req.Presupuesto != nil && req.MonedaID == nil {
		return nil, ErrMonedaRequeridaPresupuesto
	}

	okUser, err := s.repo.UsuarioPerteneceAEmpresa(ctx, usuarioID, empresaID)
	if err != nil {
		return nil, err
	}
	if !okUser {
		return nil, ErrUsuarioOtraEmpresa
	}

	tipo, err := s.repo.FindTipoCampanaByID(ctx, empresaID, req.TipoCampanaID)
	if err != nil {
		return nil, mapCampanaError(err)
	}
	if !tipo.Activo {
		return nil, ErrTipoCampanaInactivo
	}

	if req.OrigenID != nil {
		okOrigen, err := s.repo.OrigenPerteneceAEmpresa(ctx, empresaID, *req.OrigenID)
		if err != nil {
			return nil, err
		}
		if !okOrigen {
			return nil, ErrReferenciaInvalida
		}
	}

	if req.MonedaID != nil {
		okMoneda, err := s.repo.MonedaExiste(ctx, *req.MonedaID)
		if err != nil {
			return nil, err
		}
		if !okMoneda {
			return nil, ErrReferenciaInvalida
		}
	}

	estado := CampanaEstadoBorrador
	if req.Estado != nil {
		estado = *req.Estado
	}

	c := &Campana{
		EmpresaID:     empresaID,
		TipoCampanaID: req.TipoCampanaID,
		OrigenID:      req.OrigenID,
		MonedaID:      req.MonedaID,
		Nombre:        strings.TrimSpace(req.Nombre),
		Descripcion:   trimPtr(req.Descripcion),
		Objetivo:      trimPtr(req.Objetivo),
		FechaInicio:   fechaInicio,
		FechaFin:      fechaFin,
		Presupuesto:   req.Presupuesto,
		Estado:        estado,
		CreatedBy:     usuarioID,
	}
	if err := s.repo.CreateCampana(ctx, c); err != nil {
		return nil, mapCampanaError(err)
	}
	resp := toCampanaResponse(c)
	return &resp, nil
}

func (s *Service) ObtenerCampana(ctx context.Context, empresaID int32, id int64) (*CampanaResponse, error) {
	c, err := s.repo.FindCampanaByID(ctx, empresaID, id)
	if err != nil {
		return nil, mapCampanaError(err)
	}
	resp := toCampanaResponse(c)
	return &resp, nil
}

func (s *Service) ListarCampanas(ctx context.Context, empresaID int32, q ListarCampanasQuery) (*ListaCampanasResponse, error) {
	var fechaDesde, fechaHasta *time.Time
	if strings.TrimSpace(q.FechaDesde) != "" {
		d, err := parseDate(q.FechaDesde)
		if err != nil {
			return nil, ErrFechaInvalida
		}
		fechaDesde = &d
	}
	if strings.TrimSpace(q.FechaHasta) != "" {
		h, err := parseDate(q.FechaHasta)
		if err != nil {
			return nil, ErrFechaInvalida
		}
		fechaHasta = &h
	}
	if fechaDesde != nil && fechaHasta != nil && fechaHasta.Before(*fechaDesde) {
		return nil, ErrRangoFechasInvalido
	}

	q.Page, q.PageSize = normalizarPaginacion(q.Page, q.PageSize)
	items, total, err := s.repo.ListCampanas(ctx, empresaID, q, fechaDesde, fechaHasta)
	if err != nil {
		return nil, err
	}
	data := make([]CampanaResponse, 0, len(items))
	for i := range items {
		data = append(data, toCampanaResponse(&items[i]))
	}
	return &ListaCampanasResponse{
		Data:     data,
		Page:     q.Page,
		PageSize: q.PageSize,
		Total:    total,
	}, nil
}

func (s *Service) ActualizarCampana(ctx context.Context, empresaID int32, id int64, req ActualizarCampanaRequest) (*CampanaResponse, error) {
	var parsedInicio *time.Time
	if req.FechaInicio != nil {
		fi, err := parseDate(*req.FechaInicio)
		if err != nil {
			return nil, ErrFechaInvalida
		}
		parsedInicio = &fi
	}
	var parsedFin *time.Time
	clearFechaFin := false
	if req.FechaFin != nil {
		if strings.TrimSpace(*req.FechaFin) == "" {
			clearFechaFin = true
		} else {
			ff, err := parseDate(*req.FechaFin)
			if err != nil {
				return nil, ErrFechaInvalida
			}
			parsedFin = &ff
		}
	}

	var res *Campana
	err := s.repo.WithTx(ctx, func(tx *Repository) error {
		c, err := tx.FindCampanaByID(ctx, empresaID, id)
		if err != nil {
			return err
		}
		if c.Estado == CampanaEstadoFinalizada {
			return ErrCampanaFinalizada
		}

		effectiveInicio := c.FechaInicio
		if parsedInicio != nil {
			effectiveInicio = *parsedInicio
		}
		effectiveFin := c.FechaFin
		if clearFechaFin {
			effectiveFin = nil
		} else if parsedFin != nil {
			effectiveFin = parsedFin
		}
		if effectiveFin != nil && effectiveFin.Before(effectiveInicio) {
			return ErrRangoFechasInvalido
		}

		effectiveMoneda := c.MonedaID
		if req.MonedaID != nil {
			okMoneda, err := tx.MonedaExiste(ctx, *req.MonedaID)
			if err != nil {
				return err
			}
			if !okMoneda {
				return ErrReferenciaInvalida
			}
			// No permitir cambiar la moneda si ya existen gastos en otra moneda.
			gastosPorMoneda, err := tx.SumGastosPorMonedaCampana(ctx, empresaID, id)
			if err != nil {
				return err
			}
			for _, gm := range gastosPorMoneda {
				if gm.MonedaID != *req.MonedaID {
					return ErrMonedaIncompatible
				}
			}
			effectiveMoneda = req.MonedaID
		}
		effectivePresupuesto := c.Presupuesto
		if req.Presupuesto != nil {
			effectivePresupuesto = req.Presupuesto
		}
		if effectivePresupuesto != nil && effectiveMoneda == nil {
			return ErrMonedaRequeridaPresupuesto
		}

		cambios := map[string]any{}
		if req.TipoCampanaID != nil {
			tipo, err := tx.FindTipoCampanaByID(ctx, empresaID, *req.TipoCampanaID)
			if err != nil {
				return err
			}
			if !tipo.Activo {
				return ErrTipoCampanaInactivo
			}
			cambios["tipo_campana_id"] = *req.TipoCampanaID
		}
		if req.OrigenID != nil {
			okOrigen, err := tx.OrigenPerteneceAEmpresa(ctx, empresaID, *req.OrigenID)
			if err != nil {
				return err
			}
			if !okOrigen {
				return ErrReferenciaInvalida
			}
			cambios["origen_id"] = *req.OrigenID
		}
		if req.MonedaID != nil {
			cambios["moneda_id"] = *req.MonedaID
		}
		if req.Nombre != nil {
			cambios["nombre"] = strings.TrimSpace(*req.Nombre)
		}
		if req.Descripcion != nil {
			cambios["descripcion"] = trimPtr(req.Descripcion)
		}
		if req.Objetivo != nil {
			cambios["objetivo"] = trimPtr(req.Objetivo)
		}
		if parsedInicio != nil {
			cambios["fecha_inicio"] = *parsedInicio
		}
		if clearFechaFin {
			cambios["fecha_fin"] = nil
		} else if parsedFin != nil {
			cambios["fecha_fin"] = *parsedFin
		}
		if req.Presupuesto != nil {
			cambios["presupuesto"] = *req.Presupuesto
		}

		if err := tx.UpdateCampana(ctx, c, cambios); err != nil {
			return err
		}
		res, err = tx.FindCampanaByID(ctx, empresaID, id)
		return err
	})
	if err != nil {
		return nil, mapCampanaError(err)
	}
	resp := toCampanaResponse(res)
	return &resp, nil
}

func (s *Service) CambiarEstadoCampana(ctx context.Context, empresaID int32, id int64, nuevoEstado string) (*CampanaResponse, error) {
	var res *Campana
	err := s.repo.WithTx(ctx, func(tx *Repository) error {
		c, err := tx.FindCampanaByID(ctx, empresaID, id)
		if err != nil {
			return err
		}
		if c.Estado == nuevoEstado {
			res = c
			return nil
		}
		if !esTransicionEstadoValida(c.Estado, nuevoEstado) {
			return ErrTransicionEstadoInvalida
		}
		if err := tx.UpdateCampana(ctx, c, map[string]any{"estado": nuevoEstado}); err != nil {
			return err
		}
		res, err = tx.FindCampanaByID(ctx, empresaID, id)
		return err
	})
	if err != nil {
		return nil, mapCampanaError(err)
	}
	resp := toCampanaResponse(res)
	return &resp, nil
}

func (s *Service) ObtenerIndicadoresCampana(ctx context.Context, empresaID int32, id int64) (*IndicadoresCampanaResponse, error) {
	c, err := s.repo.FindCampanaByID(ctx, empresaID, id)
	if err != nil {
		return nil, mapCampanaError(err)
	}
	totalLeads, err := s.repo.CountLeadsCampana(ctx, empresaID, id)
	if err != nil {
		return nil, err
	}
	gastosPorMoneda, err := s.repo.SumGastosPorMonedaCampana(ctx, empresaID, id)
	if err != nil {
		return nil, err
	}

	var gastoAcumulado *float64
	var presupuestoDisponible *float64
	if c.MonedaID != nil {
		var suma float64
		for _, gm := range gastosPorMoneda {
			if gm.MonedaID == *c.MonedaID {
				suma = gm.TotalGastado
				break
			}
		}
		gastoAcumulado = &suma
		if c.Presupuesto != nil {
			disp := *c.Presupuesto - suma
			presupuestoDisponible = &disp
		}
	}

	return &IndicadoresCampanaResponse{
		CampanaID:             c.ID,
		EmpresaID:             c.EmpresaID,
		Nombre:                c.Nombre,
		Estado:                c.Estado,
		MonedaID:              c.MonedaID,
		Presupuesto:           c.Presupuesto,
		TotalLeads:            totalLeads,
		GastoAcumulado:        gastoAcumulado,
		PresupuestoDisponible: presupuestoDisponible,
		GastosPorMoneda:       gastosPorMoneda,
	}, nil
}

// =====================================================================
// Asociación Campaña - Lead
// =====================================================================

func (s *Service) AsociarLeadACampana(ctx context.Context, empresaID int32, campanaID int64, req AsociarLeadCampanaRequest) (*CampanaLeadResponse, error) {
	var created *CampanaLead
	err := s.repo.WithTx(ctx, func(tx *Repository) error {
		c, err := tx.FindCampanaByID(ctx, empresaID, campanaID)
		if err != nil {
			return err
		}
		if !campanaHabilitadaParaLeads(c.Estado) {
			return ErrCampanaNoHabilitada
		}
		if _, err := tx.FindLeadByIDAndEmpresa(ctx, empresaID, req.LeadID); err != nil {
			return err
		}

		fecha := time.Now().UTC()
		if req.FechaAsociacion != nil {
			fecha = req.FechaAsociacion.UTC()
		}

		cl := &CampanaLead{
			EmpresaID:          empresaID,
			CampanaID:          campanaID,
			LeadID:             req.LeadID,
			FechaAsociacion:    fecha,
			TipoAtribucion:     req.TipoAtribucion,
			UTMSource:          trimPtr(req.UTMSource),
			UTMMedium:          trimPtr(req.UTMMedium),
			UTMCampaign:        trimPtr(req.UTMCampaign),
			UTMContent:         trimPtr(req.UTMContent),
			UTMTerm:            trimPtr(req.UTMTerm),
			LandingPage:        trimPtr(req.LandingPage),
			ReferrerURL:        trimPtr(req.ReferrerURL),
			ExternalCampaignID: trimPtr(req.ExternalCampaignID),
			ExternalAdID:       trimPtr(req.ExternalAdID),
		}
		if err := tx.CreateCampanaLead(ctx, cl); err != nil {
			return err
		}
		created = cl
		return nil
	})
	if err != nil {
		return nil, mapCampanaError(err)
	}
	resp := toCampanaLeadResponse(created)
	return &resp, nil
}

func (s *Service) ListarLeadsDeCampana(ctx context.Context, empresaID int32, campanaID int64, q ListarLeadsCampanaQuery) (*ListaLeadsCampanaResponse, error) {
	if _, err := s.repo.FindCampanaByID(ctx, empresaID, campanaID); err != nil {
		return nil, mapCampanaError(err)
	}
	q.Page, q.PageSize = normalizarPaginacion(q.Page, q.PageSize)
	items, total, err := s.repo.ListLeadsByCampana(ctx, empresaID, campanaID, q)
	if err != nil {
		return nil, err
	}
	return &ListaLeadsCampanaResponse{
		Data:     items,
		Page:     q.Page,
		PageSize: q.PageSize,
		Total:    total,
	}, nil
}

func (s *Service) ListarCampanasDeLead(ctx context.Context, empresaID int32, leadID int64) ([]CampanaAsociadaLeadResponse, error) {
	if _, err := s.repo.FindLeadByIDAndEmpresa(ctx, empresaID, leadID); err != nil {
		return nil, mapCampanaError(err)
	}
	return s.repo.ListCampanasByLead(ctx, empresaID, leadID)
}

func (s *Service) DesasociarLeadDeCampana(ctx context.Context, empresaID int32, campanaID, leadID int64) error {
	return mapCampanaError(s.repo.WithTx(ctx, func(tx *Repository) error {
		c, err := tx.FindCampanaByID(ctx, empresaID, campanaID)
		if err != nil {
			return err
		}
		if c.Estado == CampanaEstadoFinalizada {
			return ErrCampanaFinalizada
		}
		return tx.DeleteCampanaLead(ctx, empresaID, campanaID, leadID)
	}))
}

// =====================================================================
// Gastos de campaña
// =====================================================================

func (s *Service) CrearGastoCampana(ctx context.Context, empresaID int32, usuarioID int64, campanaID int64, req CrearGastoCampanaRequest) (*GastoCampanaResponse, error) {
	fecha, err := parseDate(req.Fecha)
	if err != nil {
		return nil, ErrFechaInvalida
	}

	var created *CampanaGasto
	err = s.repo.WithTx(ctx, func(tx *Repository) error {
		c, err := tx.FindCampanaByID(ctx, empresaID, campanaID)
		if err != nil {
			return err
		}
		if c.Estado == CampanaEstadoFinalizada {
			return ErrCampanaFinalizada
		}
		okUser, err := tx.UsuarioPerteneceAEmpresa(ctx, usuarioID, empresaID)
		if err != nil {
			return err
		}
		if !okUser {
			return ErrUsuarioOtraEmpresa
		}
		okMoneda, err := tx.MonedaExiste(ctx, req.MonedaID)
		if err != nil {
			return err
		}
		if !okMoneda {
			return ErrReferenciaInvalida
		}
		if c.MonedaID != nil && *c.MonedaID != req.MonedaID {
			return ErrMonedaIncompatible
		}

		g := &CampanaGasto{
			EmpresaID:                    empresaID,
			CampanaID:                    campanaID,
			Fecha:                        fecha,
			Concepto:                     strings.TrimSpace(req.Concepto),
			CategoriaGasto:               trimPtr(req.CategoriaGasto),
			Valor:                        req.Valor,
			MonedaID:                     req.MonedaID,
			ProveedorTerceroEmpresaRolID: req.ProveedorTerceroEmpresaRolID,
			CreatedBy:                    usuarioID,
		}
		if err := tx.CreateCampanaGasto(ctx, g); err != nil {
			return err
		}
		created = g
		return nil
	})
	if err != nil {
		return nil, mapCampanaError(err)
	}
	resp := toGastoCampanaResponse(created)
	return &resp, nil
}

func (s *Service) ListarGastosCampana(ctx context.Context, empresaID int32, campanaID int64, q ListarGastosCampanaQuery) (*ListaGastosCampanaResponse, error) {
	if _, err := s.repo.FindCampanaByID(ctx, empresaID, campanaID); err != nil {
		return nil, mapCampanaError(err)
	}
	var fechaDesde, fechaHasta *time.Time
	if strings.TrimSpace(q.FechaDesde) != "" {
		d, err := parseDate(q.FechaDesde)
		if err != nil {
			return nil, ErrFechaInvalida
		}
		fechaDesde = &d
	}
	if strings.TrimSpace(q.FechaHasta) != "" {
		h, err := parseDate(q.FechaHasta)
		if err != nil {
			return nil, ErrFechaInvalida
		}
		fechaHasta = &h
	}
	if fechaDesde != nil && fechaHasta != nil && fechaHasta.Before(*fechaDesde) {
		return nil, ErrRangoFechasInvalido
	}

	q.Page, q.PageSize = normalizarPaginacion(q.Page, q.PageSize)
	items, total, err := s.repo.ListGastosByCampana(ctx, empresaID, campanaID, q, fechaDesde, fechaHasta)
	if err != nil {
		return nil, err
	}
	data := make([]GastoCampanaResponse, 0, len(items))
	for i := range items {
		data = append(data, toGastoCampanaResponse(&items[i]))
	}
	return &ListaGastosCampanaResponse{
		Data:     data,
		Page:     q.Page,
		PageSize: q.PageSize,
		Total:    total,
	}, nil
}

func (s *Service) ObtenerGastoCampana(ctx context.Context, empresaID int32, id int64) (*GastoCampanaResponse, error) {
	g, err := s.repo.FindCampanaGastoByID(ctx, empresaID, id)
	if err != nil {
		return nil, mapCampanaError(err)
	}
	resp := toGastoCampanaResponse(g)
	return &resp, nil
}

func (s *Service) ActualizarGastoCampana(ctx context.Context, empresaID int32, id int64, req ActualizarGastoCampanaRequest) (*GastoCampanaResponse, error) {
	var parsedFecha *time.Time
	if req.Fecha != nil {
		f, err := parseDate(*req.Fecha)
		if err != nil {
			return nil, ErrFechaInvalida
		}
		parsedFecha = &f
	}

	var res *CampanaGasto
	err := s.repo.WithTx(ctx, func(tx *Repository) error {
		g, err := tx.FindCampanaGastoByID(ctx, empresaID, id)
		if err != nil {
			return err
		}
		c, err := tx.FindCampanaByID(ctx, empresaID, g.CampanaID)
		if err != nil {
			return err
		}
		if c.Estado == CampanaEstadoFinalizada {
			return ErrCampanaFinalizada
		}

		cambios := map[string]any{}
		if parsedFecha != nil {
			cambios["fecha"] = *parsedFecha
		}
		if req.Concepto != nil {
			cambios["concepto"] = strings.TrimSpace(*req.Concepto)
		}
		if req.CategoriaGasto != nil {
			cambios["categoria_gasto"] = trimPtr(req.CategoriaGasto)
		}
		if req.Valor != nil {
			cambios["valor"] = *req.Valor
		}
		if req.MonedaID != nil {
			okMoneda, err := tx.MonedaExiste(ctx, *req.MonedaID)
			if err != nil {
				return err
			}
			if !okMoneda {
				return ErrReferenciaInvalida
			}
			if c.MonedaID != nil && *c.MonedaID != *req.MonedaID {
				return ErrMonedaIncompatible
			}
			cambios["moneda_id"] = *req.MonedaID
		}
		if req.ProveedorTerceroEmpresaRolID != nil {
			cambios["proveedor_tercero_empresa_rol_id"] = *req.ProveedorTerceroEmpresaRolID
		}

		if err := tx.UpdateCampanaGasto(ctx, g, cambios); err != nil {
			return err
		}
		res, err = tx.FindCampanaGastoByID(ctx, empresaID, id)
		return err
	})
	if err != nil {
		return nil, mapCampanaError(err)
	}
	resp := toGastoCampanaResponse(res)
	return &resp, nil
}

func (s *Service) EliminarGastoCampana(ctx context.Context, empresaID int32, id int64) error {
	return mapCampanaError(s.repo.WithTx(ctx, func(tx *Repository) error {
		g, err := tx.FindCampanaGastoByID(ctx, empresaID, id)
		if err != nil {
			return err
		}
		c, err := tx.FindCampanaByID(ctx, empresaID, g.CampanaID)
		if err != nil {
			return err
		}
		if c.Estado == CampanaEstadoFinalizada {
			return ErrCampanaFinalizada
		}
		return tx.SoftDeleteCampanaGasto(ctx, empresaID, id)
	}))
}

// =====================================================================
// Helpers de campañas
// =====================================================================

func parseDate(v string) (time.Time, error) {
	return time.Parse(dateLayout, strings.TrimSpace(v))
}

func campanaHabilitadaParaLeads(estado string) bool {
	switch estado {
	case CampanaEstadoProgramada, CampanaEstadoActiva, CampanaEstadoPausada:
		return true
	default:
		return false
	}
}

func esTransicionEstadoValida(actual, nuevo string) bool {
	switch actual {
	case CampanaEstadoBorrador:
		return nuevo == CampanaEstadoProgramada || nuevo == CampanaEstadoActiva || nuevo == CampanaEstadoPausada || nuevo == CampanaEstadoFinalizada
	case CampanaEstadoProgramada:
		return nuevo == CampanaEstadoBorrador || nuevo == CampanaEstadoActiva || nuevo == CampanaEstadoPausada || nuevo == CampanaEstadoFinalizada
	case CampanaEstadoActiva:
		return nuevo == CampanaEstadoPausada || nuevo == CampanaEstadoFinalizada
	case CampanaEstadoPausada:
		return nuevo == CampanaEstadoActiva || nuevo == CampanaEstadoFinalizada
	case CampanaEstadoFinalizada:
		return false
	default:
		return false
	}
}

func mapCampanaError(err error) error {
	if err == nil {
		return nil
	}

	var appErr *httpx.AppError
	if errors.As(err, &appErr) {
		return err
	}

	switch {
	case errors.Is(err, errTipoCampanaNotFound):
		return ErrTipoCampanaNoEncontrado
	case errors.Is(err, errCampanaNotFound):
		return ErrCampanaNoEncontrada
	case errors.Is(err, errCampanaLeadNotFound):
		return ErrCampanaLeadNoEncontrado
	case errors.Is(err, errCampanaGastoNotFound):
		return ErrCampanaGastoNoEncontrado
	case errors.Is(err, errLeadNotFound):
		return ErrLeadNoEncontrado
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			switch pgErr.ConstraintName {
			case "uq_crm_campana_leads_campana_lead":
				return ErrAsociacionDuplicada
			case "ux_crm_campana_leads_first_touch":
				return ErrFirstTouchDuplicado
			default:
				return ErrNombreDuplicado
			}
		case "23503":
			return ErrReferenciaInvalida
		}
	}
	return mapError(err)
}
