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

const (
	defaultPageSize = 20
	maxPageSize     = 100
)

var (
	ErrOrigenNoEncontrado       = httpx.NewError(http.StatusNotFound, "origen_no_encontrado", "el origen de lead no existe")
	ErrEstadoNoEncontrado       = httpx.NewError(http.StatusNotFound, "estado_no_encontrado", "el estado de lead no existe")
	ErrMotivoNoEncontrado       = httpx.NewError(http.StatusNotFound, "motivo_descarte_no_encontrado", "el motivo de descarte no existe")
	ErrLeadNoEncontrado         = httpx.NewError(http.StatusNotFound, "lead_no_encontrado", "el lead no existe")
	ErrActividadNoEncontrada    = httpx.NewError(http.StatusNotFound, "actividad_no_encontrada", "la actividad no existe")
	ErrInteresNoEncontrado      = httpx.NewError(http.StatusNotFound, "interes_no_encontrado", "el interés de producto no existe")
	ErrLeadYaConvertido         = httpx.NewError(http.StatusConflict, "lead_ya_convertido", "el lead ya fue convertido previamente")
	ErrUsuarioOtraEmpresa       = httpx.NewError(http.StatusUnprocessableEntity, "usuario_invalido", "el usuario no existe o no pertenece a la empresa del lead")
	ErrEstadoOtraEmpresa        = httpx.NewError(http.StatusUnprocessableEntity, "estado_invalido", "el estado no pertenece a la empresa del lead")
	ErrInteresSinReferencia     = httpx.NewError(http.StatusBadRequest, "interes_sin_referencia", "debe indicar empresa_producto_id, categoria_id o notas")
	ErrNombreDuplicado          = httpx.NewError(http.StatusConflict, "duplicado", "ya existe un registro con ese nombre en la empresa")
	ErrReferenciaInvalida       = httpx.NewError(http.StatusUnprocessableEntity, "referencia_invalida", "alguna de las referencias indicadas (empresa, sucursal, origen, estado, motivo o moneda) no es válida")
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// =====================================================================
// Orígenes de lead
// =====================================================================

func (s *Service) CrearOrigen(ctx context.Context, req CrearOrigenRequest) (*OrigenResponse, error) {
	activo := true
	if req.Activo != nil {
		activo = *req.Activo
	}
	o := &OrigenLead{
		EmpresaID: req.EmpresaID,
		Nombre:    strings.TrimSpace(req.Nombre),
		Activo:    activo,
	}
	if err := s.repo.CreateOrigen(ctx, o); err != nil {
		return nil, mapError(err)
	}
	resp := toOrigenResponse(o)
	return &resp, nil
}

func (s *Service) ListarOrigenes(ctx context.Context, q ListarCatalogoQuery) ([]OrigenResponse, error) {
	items, err := s.repo.ListOrigenes(ctx, q)
	if err != nil {
		return nil, err
	}
	out := make([]OrigenResponse, 0, len(items))
	for i := range items {
		out = append(out, toOrigenResponse(&items[i]))
	}
	return out, nil
}

func (s *Service) ObtenerOrigen(ctx context.Context, id int32) (*OrigenResponse, error) {
	o, err := s.repo.FindOrigenByID(ctx, id)
	if err != nil {
		return nil, mapError(err)
	}
	resp := toOrigenResponse(o)
	return &resp, nil
}

func (s *Service) ActualizarOrigen(ctx context.Context, id int32, req ActualizarOrigenRequest) (*OrigenResponse, error) {
	var res *OrigenLead
	err := s.repo.WithTx(ctx, func(tx *Repository) error {
		o, err := tx.FindOrigenByID(ctx, id)
		if err != nil {
			return err
		}
		cambios := map[string]any{}
		if req.Nombre != nil {
			cambios["nombre"] = strings.TrimSpace(*req.Nombre)
		}
		if req.Activo != nil {
			cambios["activo"] = *req.Activo
		}
		if err := tx.UpdateOrigen(ctx, o, cambios); err != nil {
			return err
		}
		res, err = tx.FindOrigenByID(ctx, id)
		return err
	})
	if err != nil {
		return nil, mapError(err)
	}
	resp := toOrigenResponse(res)
	return &resp, nil
}

func (s *Service) EliminarOrigen(ctx context.Context, id int32) error {
	return mapError(s.repo.SoftDeleteOrigen(ctx, id))
}

// =====================================================================
// Estados de lead
// =====================================================================

func (s *Service) CrearEstado(ctx context.Context, req CrearEstadoRequest) (*EstadoResponse, error) {
	orden := int32(1)
	if req.Orden != nil {
		orden = *req.Orden
	}
	esFinal := false
	if req.EsEstadoFinal != nil {
		esFinal = *req.EsEstadoFinal
	}
	e := &EstadoLead{
		EmpresaID:     req.EmpresaID,
		Nombre:        strings.TrimSpace(req.Nombre),
		Orden:         orden,
		EsEstadoFinal: esFinal,
	}
	if err := s.repo.CreateEstado(ctx, e); err != nil {
		return nil, mapError(err)
	}
	resp := toEstadoResponse(e)
	return &resp, nil
}

func (s *Service) ListarEstados(ctx context.Context, empresaID int32) ([]EstadoResponse, error) {
	items, err := s.repo.ListEstados(ctx, empresaID)
	if err != nil {
		return nil, err
	}
	out := make([]EstadoResponse, 0, len(items))
	for i := range items {
		out = append(out, toEstadoResponse(&items[i]))
	}
	return out, nil
}

func (s *Service) ObtenerEstado(ctx context.Context, id int32) (*EstadoResponse, error) {
	e, err := s.repo.FindEstadoByID(ctx, id)
	if err != nil {
		return nil, mapError(err)
	}
	resp := toEstadoResponse(e)
	return &resp, nil
}

func (s *Service) ActualizarEstado(ctx context.Context, id int32, req ActualizarEstadoRequest) (*EstadoResponse, error) {
	var res *EstadoLead
	err := s.repo.WithTx(ctx, func(tx *Repository) error {
		e, err := tx.FindEstadoByID(ctx, id)
		if err != nil {
			return err
		}
		cambios := map[string]any{}
		if req.Nombre != nil {
			cambios["nombre"] = strings.TrimSpace(*req.Nombre)
		}
		if req.Orden != nil {
			cambios["orden"] = *req.Orden
		}
		if req.EsEstadoFinal != nil {
			cambios["es_estado_final"] = *req.EsEstadoFinal
		}
		if err := tx.UpdateEstado(ctx, e, cambios); err != nil {
			return err
		}
		res, err = tx.FindEstadoByID(ctx, id)
		return err
	})
	if err != nil {
		return nil, mapError(err)
	}
	resp := toEstadoResponse(res)
	return &resp, nil
}

func (s *Service) EliminarEstado(ctx context.Context, id int32) error {
	return mapError(s.repo.SoftDeleteEstado(ctx, id))
}

// =====================================================================
// Motivos de descarte
// =====================================================================

func (s *Service) CrearMotivo(ctx context.Context, req CrearMotivoDescarteRequest) (*MotivoDescarteResponse, error) {
	activo := true
	if req.Activo != nil {
		activo = *req.Activo
	}
	m := &MotivoDescarte{
		EmpresaID:   req.EmpresaID,
		Descripcion: strings.TrimSpace(req.Descripcion),
		Activo:      activo,
	}
	if err := s.repo.CreateMotivo(ctx, m); err != nil {
		return nil, mapError(err)
	}
	resp := toMotivoResponse(m)
	return &resp, nil
}

func (s *Service) ListarMotivos(ctx context.Context, q ListarCatalogoQuery) ([]MotivoDescarteResponse, error) {
	items, err := s.repo.ListMotivos(ctx, q)
	if err != nil {
		return nil, err
	}
	out := make([]MotivoDescarteResponse, 0, len(items))
	for i := range items {
		out = append(out, toMotivoResponse(&items[i]))
	}
	return out, nil
}

func (s *Service) ObtenerMotivo(ctx context.Context, id int32) (*MotivoDescarteResponse, error) {
	m, err := s.repo.FindMotivoByID(ctx, id)
	if err != nil {
		return nil, mapError(err)
	}
	resp := toMotivoResponse(m)
	return &resp, nil
}

func (s *Service) ActualizarMotivo(ctx context.Context, id int32, req ActualizarMotivoDescarteRequest) (*MotivoDescarteResponse, error) {
	var res *MotivoDescarte
	err := s.repo.WithTx(ctx, func(tx *Repository) error {
		m, err := tx.FindMotivoByID(ctx, id)
		if err != nil {
			return err
		}
		cambios := map[string]any{}
		if req.Descripcion != nil {
			cambios["descripcion"] = strings.TrimSpace(*req.Descripcion)
		}
		if req.Activo != nil {
			cambios["activo"] = *req.Activo
		}
		if err := tx.UpdateMotivo(ctx, m, cambios); err != nil {
			return err
		}
		res, err = tx.FindMotivoByID(ctx, id)
		return err
	})
	if err != nil {
		return nil, mapError(err)
	}
	resp := toMotivoResponse(res)
	return &resp, nil
}

func (s *Service) EliminarMotivo(ctx context.Context, id int32) error {
	return mapError(s.repo.SoftDeleteMotivo(ctx, id))
}

// =====================================================================
// Leads
// =====================================================================

func (s *Service) CrearLead(ctx context.Context, req CrearLeadRequest) (*LeadResponse, error) {
	ok, err := s.repo.UsuarioPerteneceAEmpresa(ctx, req.CreatedBy, req.EmpresaID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrUsuarioOtraEmpresa
	}
	if req.UsuarioAsignadoID != nil {
		ok, err := s.repo.UsuarioPerteneceAEmpresa(ctx, *req.UsuarioAsignadoID, req.EmpresaID)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, ErrUsuarioOtraEmpresa
		}
	}

	temp := TempTibio
	if req.Temperatura != nil {
		temp = *req.Temperatura
	}
	var score int32
	if req.Score != nil {
		score = *req.Score
	}
	var valor float64
	if req.ValorEstimado != nil {
		valor = *req.ValorEstimado
	}

	l := &Lead{
		EmpresaID:         req.EmpresaID,
		SucursalID:        req.SucursalID,
		OrigenID:          req.OrigenID,
		EstadoID:          req.EstadoID,
		MotivoDescarteID:  req.MotivoDescarteID,
		UsuarioAsignadoID: req.UsuarioAsignadoID,
		Nombre:            strings.TrimSpace(req.Nombre),
		Apellido:          strings.TrimSpace(req.Apellido),
		NombreEmpresa:     trimPtr(req.NombreEmpresa),
		Cargo:             trimPtr(req.Cargo),
		Email:             normalizarEmailPtr(req.Email),
		Telefono:          trimPtr(req.Telefono),
		UsuarioFacebook:   trimPtr(req.UsuarioFacebook),
		Temperatura:       temp,
		Score:             score,
		ValorEstimado:     valor,
		MonedaID:          req.MonedaID,
		Convertido:        false,
		CreatedBy:         req.CreatedBy,
	}
	if err := s.repo.CreateLead(ctx, l); err != nil {
		return nil, mapError(err)
	}
	resp := toLeadResponse(l)
	return &resp, nil
}

func (s *Service) ObtenerLead(ctx context.Context, id int64) (*LeadResponse, error) {
	l, err := s.repo.FindLeadByID(ctx, id)
	if err != nil {
		return nil, mapError(err)
	}
	resp := toLeadResponse(l)
	return &resp, nil
}

func (s *Service) ListarLeads(ctx context.Context, q ListarLeadsQuery) (*ListaLeadsResponse, error) {
	q.Page, q.PageSize = normalizarPaginacion(q.Page, q.PageSize)
	items, total, err := s.repo.ListLeads(ctx, q)
	if err != nil {
		return nil, err
	}
	data := make([]LeadResponse, 0, len(items))
	for i := range items {
		data = append(data, toLeadResponse(&items[i]))
	}
	return &ListaLeadsResponse{Data: data, Page: q.Page, PageSize: q.PageSize, Total: total}, nil
}

func (s *Service) ActualizarLead(ctx context.Context, id int64, req ActualizarLeadRequest) (*LeadResponse, error) {
	var res *Lead
	err := s.repo.WithTx(ctx, func(tx *Repository) error {
		l, err := tx.FindLeadByID(ctx, id)
		if err != nil {
			return err
		}

		cambios := map[string]any{}
		if req.UsuarioAsignadoID != nil {
			ok, err := tx.UsuarioPerteneceAEmpresa(ctx, *req.UsuarioAsignadoID, l.EmpresaID)
			if err != nil {
				return err
			}
			if !ok {
				return ErrUsuarioOtraEmpresa
			}
			cambios["usuario_asignado_id"] = *req.UsuarioAsignadoID
		}
		if req.SucursalID != nil {
			cambios["sucursal_id"] = *req.SucursalID
		}
		if req.OrigenID != nil {
			cambios["origen_id"] = *req.OrigenID
		}
		if req.EstadoID != nil {
			cambios["estado_id"] = *req.EstadoID
		}
		if req.MotivoDescarteID != nil {
			cambios["motivo_descarte_id"] = *req.MotivoDescarteID
		}
		if req.Nombre != nil {
			cambios["nombre"] = strings.TrimSpace(*req.Nombre)
		}
		if req.Apellido != nil {
			cambios["apellido"] = strings.TrimSpace(*req.Apellido)
		}
		if req.NombreEmpresa != nil {
			cambios["nombre_empresa"] = trimPtr(req.NombreEmpresa)
		}
		if req.Cargo != nil {
			cambios["cargo"] = trimPtr(req.Cargo)
		}
		if req.Email != nil {
			cambios["email"] = normalizarEmailPtr(req.Email)
		}
		if req.Telefono != nil {
			cambios["telefono"] = trimPtr(req.Telefono)
		}
		if req.UsuarioFacebook != nil {
			cambios["usuario_facebook"] = trimPtr(req.UsuarioFacebook)
		}
		if req.Temperatura != nil {
			cambios["temperatura"] = *req.Temperatura
		}
		if req.Score != nil {
			cambios["score"] = *req.Score
		}
		if req.ValorEstimado != nil {
			cambios["valor_estimado"] = *req.ValorEstimado
		}
		if req.MonedaID != nil {
			cambios["moneda_id"] = *req.MonedaID
		}

		if err := tx.UpdateLead(ctx, l, cambios); err != nil {
			return err
		}
		res, err = tx.FindLeadByID(ctx, id)
		return err
	})
	if err != nil {
		return nil, mapError(err)
	}
	resp := toLeadResponse(res)
	return &resp, nil
}

func (s *Service) CambiarEstadoLead(ctx context.Context, id int64, req CambiarEstadoLeadRequest) (*LeadResponse, error) {
	var res *Lead
	err := s.repo.WithTx(ctx, func(tx *Repository) error {
		l, err := tx.FindLeadByID(ctx, id)
		if err != nil {
			return err
		}
		estado, err := tx.FindEstadoByID(ctx, req.EstadoID)
		if err != nil {
			return err
		}
		if estado.EmpresaID != l.EmpresaID {
			return ErrEstadoOtraEmpresa
		}

		cambios := map[string]any{"estado_id": req.EstadoID}
		if req.MotivoDescarteID != nil {
			cambios["motivo_descarte_id"] = *req.MotivoDescarteID
		} else if !estado.EsEstadoFinal {
			// Si vuelve a una etapa abierta, limpiamos un posible motivo de descarte previo.
			cambios["motivo_descarte_id"] = nil
		}

		if err := tx.UpdateLead(ctx, l, cambios); err != nil {
			return err
		}
		res, err = tx.FindLeadByID(ctx, id)
		return err
	})
	if err != nil {
		return nil, mapError(err)
	}
	resp := toLeadResponse(res)
	return &resp, nil
}

func (s *Service) ConvertirLead(ctx context.Context, id int64, req ConvertirLeadRequest) (*LeadResponse, error) {
	var res *Lead
	err := s.repo.WithTx(ctx, func(tx *Repository) error {
		l, err := tx.FindLeadByID(ctx, id)
		if err != nil {
			return err
		}
		if l.Convertido {
			return ErrLeadYaConvertido
		}

		now := time.Now().UTC()
		cambios := map[string]any{
			"convertido":         true,
			"fecha_conversion":   now,
			"motivo_descarte_id": nil,
		}
		if req.TerceroEmpresaRolID != nil {
			cambios["convertido_tercero_empresa_rol_id"] = *req.TerceroEmpresaRolID
		}
		if req.EstadoID != nil {
			estado, err := tx.FindEstadoByID(ctx, *req.EstadoID)
			if err != nil {
				return err
			}
			if estado.EmpresaID != l.EmpresaID {
				return ErrEstadoOtraEmpresa
			}
			cambios["estado_id"] = *req.EstadoID
		}

		if err := tx.UpdateLead(ctx, l, cambios); err != nil {
			return err
		}
		res, err = tx.FindLeadByID(ctx, id)
		return err
	})
	if err != nil {
		return nil, mapError(err)
	}
	resp := toLeadResponse(res)
	return &resp, nil
}

func (s *Service) EliminarLead(ctx context.Context, id int64) error {
	return mapError(s.repo.SoftDeleteLead(ctx, id))
}

// =====================================================================
// Actividades
// =====================================================================

func (s *Service) CrearActividad(ctx context.Context, leadID int64, req CrearActividadRequest) (*ActividadResponse, error) {
	l, err := s.repo.FindLeadByID(ctx, leadID)
	if err != nil {
		return nil, mapError(err)
	}
	ok, err := s.repo.UsuarioPerteneceAEmpresa(ctx, req.UsuarioID, l.EmpresaID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrUsuarioOtraEmpresa
	}

	fechaAct := time.Now().UTC()
	if req.FechaActividad != nil {
		fechaAct = req.FechaActividad.UTC()
	}
	estado := ActEstadoPendiente
	if req.Estado != nil {
		estado = *req.Estado
	}

	a := &LeadActividad{
		LeadID:           leadID,
		UsuarioID:        req.UsuarioID,
		TipoActividad:    req.TipoActividad,
		FechaActividad:   fechaAct,
		FechaVencimiento: req.FechaVencimiento,
		Estado:           estado,
		Asunto:           strings.TrimSpace(req.Asunto),
		Observaciones:    trimPtr(req.Observaciones),
	}
	if err := s.repo.CreateActividad(ctx, a); err != nil {
		return nil, mapError(err)
	}
	resp := toActividadResponse(a)
	return &resp, nil
}

func (s *Service) ListarActividades(ctx context.Context, leadID *int64, q ListarActividadesQuery) (*ListaActividadesResponse, error) {
	if leadID != nil {
		if _, err := s.repo.FindLeadByID(ctx, *leadID); err != nil {
			return nil, mapError(err)
		}
	}
	q.Page, q.PageSize = normalizarPaginacion(q.Page, q.PageSize)
	items, total, err := s.repo.ListActividades(ctx, leadID, q)
	if err != nil {
		return nil, err
	}
	data := make([]ActividadResponse, 0, len(items))
	for i := range items {
		data = append(data, toActividadResponse(&items[i]))
	}
	return &ListaActividadesResponse{Data: data, Page: q.Page, PageSize: q.PageSize, Total: total}, nil
}

func (s *Service) ActualizarActividad(ctx context.Context, id int64, req ActualizarActividadRequest) (*ActividadResponse, error) {
	var res *LeadActividad
	err := s.repo.WithTx(ctx, func(tx *Repository) error {
		a, err := tx.FindActividadByID(ctx, id)
		if err != nil {
			return err
		}
		cambios := map[string]any{}
		if req.TipoActividad != nil {
			cambios["tipo_actividad"] = *req.TipoActividad
		}
		if req.FechaActividad != nil {
			cambios["fecha_actividad"] = req.FechaActividad.UTC()
		}
		if req.FechaVencimiento != nil {
			cambios["fecha_vencimiento"] = req.FechaVencimiento.UTC()
		}
		if req.Estado != nil {
			cambios["estado"] = *req.Estado
		}
		if req.Asunto != nil {
			cambios["asunto"] = strings.TrimSpace(*req.Asunto)
		}
		if req.Observaciones != nil {
			cambios["observaciones"] = trimPtr(req.Observaciones)
		}
		if err := tx.UpdateActividad(ctx, a, cambios); err != nil {
			return err
		}
		res, err = tx.FindActividadByID(ctx, id)
		return err
	})
	if err != nil {
		return nil, mapError(err)
	}
	resp := toActividadResponse(res)
	return &resp, nil
}

func (s *Service) EliminarActividad(ctx context.Context, id int64) error {
	return mapError(s.repo.SoftDeleteActividad(ctx, id))
}

// =====================================================================
// Intereses en productos
// =====================================================================

func (s *Service) AgregarInteres(ctx context.Context, leadID int64, req CrearInteresProductoRequest) (*InteresProductoResponse, error) {
	notas := trimPtr(req.Notas)
	if req.EmpresaProductoID == nil && req.CategoriaID == nil && notas == nil {
		return nil, ErrInteresSinReferencia
	}
	if _, err := s.repo.FindLeadByID(ctx, leadID); err != nil {
		return nil, mapError(err)
	}

	cant := 1.0
	if req.CantidadEstimada != nil {
		cant = *req.CantidadEstimada
	}
	i := &LeadInteresProducto{
		LeadID:            leadID,
		EmpresaProductoID: req.EmpresaProductoID,
		CategoriaID:       req.CategoriaID,
		CantidadEstimada:  cant,
		Notas:             notas,
	}
	if err := s.repo.CreateInteres(ctx, i); err != nil {
		return nil, mapError(err)
	}
	resp := toInteresResponse(i)
	return &resp, nil
}

func (s *Service) ListarIntereses(ctx context.Context, leadID int64) ([]InteresProductoResponse, error) {
	if _, err := s.repo.FindLeadByID(ctx, leadID); err != nil {
		return nil, mapError(err)
	}
	items, err := s.repo.ListInteresesByLead(ctx, leadID)
	if err != nil {
		return nil, err
	}
	out := make([]InteresProductoResponse, 0, len(items))
	for i := range items {
		out = append(out, toInteresResponse(&items[i]))
	}
	return out, nil
}

func (s *Service) EliminarInteres(ctx context.Context, leadID, interesID int64) error {
	return mapError(s.repo.DeleteInteres(ctx, leadID, interesID))
}

// =====================================================================
// Resumen / Control del CRM
// =====================================================================

func (s *Service) Resumen(ctx context.Context, empresaID int32) (*ResumenCRMResponse, error) {
	return s.repo.ResumenCRM(ctx, empresaID)
}

// =====================================================================
// Helpers
// =====================================================================

func normalizarPaginacion(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = defaultPageSize
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}
	return page, pageSize
}

func normalizarEmailPtr(p *string) *string {
	if p == nil {
		return nil
	}
	v := strings.ToLower(strings.TrimSpace(*p))
	if v == "" {
		return nil
	}
	return &v
}

func trimPtr(p *string) *string {
	if p == nil {
		return nil
	}
	v := strings.TrimSpace(*p)
	if v == "" {
		return nil
	}
	return &v
}

func mapError(err error) error {
	if err == nil {
		return nil
	}

	var appErr *httpx.AppError
	if errors.As(err, &appErr) {
		return err
	}

	switch {
	case errors.Is(err, errOrigenNotFound):
		return ErrOrigenNoEncontrado
	case errors.Is(err, errEstadoNotFound):
		return ErrEstadoNoEncontrado
	case errors.Is(err, errMotivoNotFound):
		return ErrMotivoNoEncontrado
	case errors.Is(err, errLeadNotFound):
		return ErrLeadNoEncontrado
	case errors.Is(err, errActividadNotFound):
		return ErrActividadNoEncontrada
	case errors.Is(err, errInteresNotFound):
		return ErrInteresNoEncontrado
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505": // unique_violation
			return ErrNombreDuplicado
		case "23503": // foreign_key_violation
			return ErrReferenciaInvalida
		}
	}
	return err
}
