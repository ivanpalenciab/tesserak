package precios

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
	ErrListaNoEncontrada   = httpx.NewError(http.StatusNotFound, "lista_precio_no_encontrada", "la lista de precios no existe")
	ErrDetalleNoEncontrado = httpx.NewError(http.StatusNotFound, "detalle_precio_no_encontrado", "el detalle de precio no existe")
	ErrNombreDuplicado     = httpx.NewError(http.StatusConflict, "nombre_duplicado", "ya existe una lista de precios activa con ese nombre en la empresa")
	ErrReferenciaInvalida  = httpx.NewError(http.StatusUnprocessableEntity, "referencia_invalida", "alguna de las referencias indicadas (moneda, artículo o unidad de medida) no es válida")
	ErrArticuloInvalido    = httpx.NewError(http.StatusUnprocessableEntity, "articulo_invalido", "el artículo indicado no existe o no pertenece a la empresa")
	ErrFechaInvalida       = httpx.NewError(http.StatusBadRequest, "fecha_invalida", "el formato de fecha debe ser YYYY-MM-DD")
	ErrVigenciaInvalida    = httpx.NewError(http.StatusBadRequest, "vigencia_invalida", "vigencia_hasta debe ser mayor o igual a vigencia_desde")
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// =====================================================================
// listas_precios
// =====================================================================

func (s *Service) CrearLista(ctx context.Context, empresaID int32, req CrearListaPrecioRequest) (*ListaPrecioResponse, error) {
	vigenciaDesde, err := parseDate(req.VigenciaDesde)
	if err != nil {
		return nil, ErrFechaInvalida
	}
	vigenciaHasta, err := parseDatePtr(req.VigenciaHasta)
	if err != nil {
		return nil, ErrFechaInvalida
	}
	if vigenciaHasta != nil && vigenciaHasta.Before(vigenciaDesde) {
		return nil, ErrVigenciaInvalida
	}

	estado := true
	if req.Estado != nil {
		estado = *req.Estado
	}

	lp := &ListaPrecio{
		EmpresaID:     empresaID,
		Nombre:        strings.TrimSpace(req.Nombre),
		MonedaID:      req.MonedaID,
		TipoLista:     req.TipoLista,
		VigenciaDesde: vigenciaDesde,
		VigenciaHasta: vigenciaHasta,
		Estado:        estado,
	}
	if err := s.repo.CreateLista(ctx, lp); err != nil {
		return nil, mapError(err)
	}

	creada, err := s.repo.FindListaByID(ctx, empresaID, lp.ID)
	if err != nil {
		return nil, mapError(err)
	}
	resp := toListaResponse(creada)
	return &resp, nil
}

func (s *Service) ObtenerLista(ctx context.Context, empresaID, id int32) (*ListaPrecioResponse, error) {
	lp, err := s.repo.FindListaByID(ctx, empresaID, id)
	if err != nil {
		return nil, mapError(err)
	}
	resp := toListaResponse(lp)
	return &resp, nil
}

func (s *Service) ListarListas(ctx context.Context, empresaID int32, q ListarListasPreciosQuery) (*ListasPreciosPaginadasResponse, error) {
	q.Page, q.PageSize = normalizePagination(q.Page, q.PageSize)

	items, total, err := s.repo.ListListas(ctx, empresaID, q)
	if err != nil {
		return nil, mapError(err)
	}
	data := make([]ListaPrecioResponse, 0, len(items))
	for i := range items {
		data = append(data, toListaResponse(&items[i]))
	}
	return &ListasPreciosPaginadasResponse{
		Data:     data,
		Page:     q.Page,
		PageSize: q.PageSize,
		Total:    total,
	}, nil
}

func (s *Service) ActualizarLista(ctx context.Context, empresaID, id int32, req ActualizarListaPrecioRequest) (*ListaPrecioResponse, error) {
	actual, err := s.repo.FindListaByID(ctx, empresaID, id)
	if err != nil {
		return nil, mapError(err)
	}

	effDesde := actual.VigenciaDesde
	effHasta := actual.VigenciaHasta

	cambios := map[string]any{}
	if req.Nombre != nil {
		cambios["nombre"] = strings.TrimSpace(*req.Nombre)
	}
	if req.MonedaID != nil {
		cambios["moneda_id"] = *req.MonedaID
	}
	if req.TipoLista != nil {
		cambios["tipo_lista"] = *req.TipoLista
	}
	if req.VigenciaDesde != nil {
		d, err := parseDate(*req.VigenciaDesde)
		if err != nil {
			return nil, ErrFechaInvalida
		}
		effDesde = d
		cambios["vigencia_desde"] = d
	}
	if req.LimpiarVigenciaHasta {
		effHasta = nil
		cambios["vigencia_hasta"] = nil
	} else if req.VigenciaHasta != nil {
		h, err := parseDatePtr(req.VigenciaHasta)
		if err != nil {
			return nil, ErrFechaInvalida
		}
		effHasta = h
		cambios["vigencia_hasta"] = h
	}
	if effHasta != nil && effHasta.Before(effDesde) {
		return nil, ErrVigenciaInvalida
	}
	if req.Estado != nil {
		cambios["estado"] = *req.Estado
	}

	if err := s.repo.UpdateLista(ctx, empresaID, id, cambios); err != nil {
		return nil, mapError(err)
	}

	actualizada, err := s.repo.FindListaByID(ctx, empresaID, id)
	if err != nil {
		return nil, mapError(err)
	}
	resp := toListaResponse(actualizada)
	return &resp, nil
}

func (s *Service) EliminarLista(ctx context.Context, empresaID, id int32) error {
	return mapError(s.repo.SoftDeleteLista(ctx, empresaID, id))
}

// =====================================================================
// lista_precios_detalle
// =====================================================================

func (s *Service) CrearDetalle(ctx context.Context, empresaID, listaPrecioID int32, req CrearDetallePrecioRequest) (*DetallePrecioResponse, error) {
	vigenciaDesde, err := parseDatePtr(req.VigenciaDesde)
	if err != nil {
		return nil, ErrFechaInvalida
	}
	vigenciaHasta, err := parseDatePtr(req.VigenciaHasta)
	if err != nil {
		return nil, ErrFechaInvalida
	}
	if vigenciaDesde != nil && vigenciaHasta != nil && vigenciaHasta.Before(*vigenciaDesde) {
		return nil, ErrVigenciaInvalida
	}

	if _, err := s.repo.FindListaByID(ctx, empresaID, listaPrecioID); err != nil {
		return nil, mapError(err)
	}

	ok, err := s.repo.ArticuloBelongsToEmpresa(ctx, empresaID, req.ArticuloID)
	if err != nil {
		return nil, mapError(err)
	}
	if !ok {
		return nil, ErrArticuloInvalido
	}

	cantidadMinima := 1.0
	if req.CantidadMinima != nil {
		cantidadMinima = *req.CantidadMinima
	}

	d := &ListaPrecioDetalle{
		ListaPrecioID:  listaPrecioID,
		ArticuloID:     req.ArticuloID,
		Precio:         *req.Precio,
		UnidadMedidaID: req.UnidadMedidaID,
		CantidadMinima: cantidadMinima,
		VigenciaDesde:  vigenciaDesde,
		VigenciaHasta:  vigenciaHasta,
	}
	if err := s.repo.CreateDetalle(ctx, d); err != nil {
		return nil, mapError(err)
	}

	creado, err := s.repo.FindDetalleByID(ctx, listaPrecioID, d.ID)
	if err != nil {
		return nil, mapError(err)
	}
	resp := toDetalleResponse(creado)
	return &resp, nil
}

func (s *Service) ObtenerDetalle(ctx context.Context, empresaID, listaPrecioID int32, detalleID int64) (*DetallePrecioResponse, error) {
	if _, err := s.repo.FindListaByID(ctx, empresaID, listaPrecioID); err != nil {
		return nil, mapError(err)
	}
	d, err := s.repo.FindDetalleByID(ctx, listaPrecioID, detalleID)
	if err != nil {
		return nil, mapError(err)
	}
	resp := toDetalleResponse(d)
	return &resp, nil
}

func (s *Service) ListarDetalles(ctx context.Context, empresaID, listaPrecioID int32, q ListarDetallesPrecioQuery) (*DetallesPrecioPaginadosResponse, error) {
	if _, err := s.repo.FindListaByID(ctx, empresaID, listaPrecioID); err != nil {
		return nil, mapError(err)
	}
	q.Page, q.PageSize = normalizePagination(q.Page, q.PageSize)

	items, total, err := s.repo.ListDetalles(ctx, listaPrecioID, q)
	if err != nil {
		return nil, mapError(err)
	}
	data := make([]DetallePrecioResponse, 0, len(items))
	for i := range items {
		data = append(data, toDetalleResponse(&items[i]))
	}
	return &DetallesPrecioPaginadosResponse{
		Data:     data,
		Page:     q.Page,
		PageSize: q.PageSize,
		Total:    total,
	}, nil
}

func (s *Service) ActualizarDetalle(ctx context.Context, empresaID, listaPrecioID int32, detalleID int64, req ActualizarDetallePrecioRequest) (*DetallePrecioResponse, error) {
	if _, err := s.repo.FindListaByID(ctx, empresaID, listaPrecioID); err != nil {
		return nil, mapError(err)
	}
	actual, err := s.repo.FindDetalleByID(ctx, listaPrecioID, detalleID)
	if err != nil {
		return nil, mapError(err)
	}

	effDesde := actual.VigenciaDesde
	effHasta := actual.VigenciaHasta

	cambios := map[string]any{}
	if req.ArticuloID != nil {
		ok, err := s.repo.ArticuloBelongsToEmpresa(ctx, empresaID, *req.ArticuloID)
		if err != nil {
			return nil, mapError(err)
		}
		if !ok {
			return nil, ErrArticuloInvalido
		}
		cambios["articulo_id"] = *req.ArticuloID
	}
	if req.Precio != nil {
		cambios["precio"] = *req.Precio
	}
	if req.UnidadMedidaID != nil {
		cambios["unidad_medida_id"] = *req.UnidadMedidaID
	}
	if req.CantidadMinima != nil {
		cambios["cantidad_minima"] = *req.CantidadMinima
	}
	if req.LimpiarVigenciaDesde {
		effDesde = nil
		cambios["vigencia_desde"] = nil
	} else if req.VigenciaDesde != nil {
		d, err := parseDatePtr(req.VigenciaDesde)
		if err != nil {
			return nil, ErrFechaInvalida
		}
		effDesde = d
		cambios["vigencia_desde"] = d
	}
	if req.LimpiarVigenciaHasta {
		effHasta = nil
		cambios["vigencia_hasta"] = nil
	} else if req.VigenciaHasta != nil {
		h, err := parseDatePtr(req.VigenciaHasta)
		if err != nil {
			return nil, ErrFechaInvalida
		}
		effHasta = h
		cambios["vigencia_hasta"] = h
	}
	if effDesde != nil && effHasta != nil && effHasta.Before(*effDesde) {
		return nil, ErrVigenciaInvalida
	}

	if err := s.repo.UpdateDetalle(ctx, listaPrecioID, detalleID, cambios); err != nil {
		return nil, mapError(err)
	}

	actualizado, err := s.repo.FindDetalleByID(ctx, listaPrecioID, detalleID)
	if err != nil {
		return nil, mapError(err)
	}
	resp := toDetalleResponse(actualizado)
	return &resp, nil
}

func (s *Service) EliminarDetalle(ctx context.Context, empresaID, listaPrecioID int32, detalleID int64) error {
	if _, err := s.repo.FindListaByID(ctx, empresaID, listaPrecioID); err != nil {
		return mapError(err)
	}
	return mapError(s.repo.DeleteDetalle(ctx, listaPrecioID, detalleID))
}

// =====================================================================
// Helpers
// =====================================================================

func normalizePagination(page, pageSize int) (int, int) {
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

func parseDate(s string) (time.Time, error) {
	return time.Parse(dateLayout, strings.TrimSpace(s))
}

func parseDatePtr(p *string) (*time.Time, error) {
	if p == nil {
		return nil, nil
	}
	v := strings.TrimSpace(*p)
	if v == "" {
		return nil, nil
	}
	t, err := time.Parse(dateLayout, v)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func mapError(err error) error {
	if err == nil {
		return nil
	}
	var appErr *httpx.AppError
	if errors.As(err, &appErr) {
		return err
	}
	if errors.Is(err, errListaNotFound) {
		return ErrListaNoEncontrada
	}
	if errors.Is(err, errDetalleNotFound) {
		return ErrDetalleNoEncontrado
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return ErrNombreDuplicado
		case "23503":
			return ErrReferenciaInvalida
		case "23514":
			return ErrVigenciaInvalida
		}
	}
	return err
}
