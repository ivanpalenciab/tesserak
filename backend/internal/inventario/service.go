package inventario

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"

	"tesserak/backend/internal/consecutivos"
	"tesserak/backend/internal/httpx"
)

const (
	defaultPageSize = 20
	maxPageSize     = 100
)

var (
	ErrUbicacionNoEncontrada = httpx.NewError(http.StatusNotFound, "ubicacion_no_encontrada", "la ubicación no existe")
	ErrDocumentoNoEncontrado = httpx.NewError(http.StatusNotFound, "documento_no_encontrado", "el movimiento de inventario no existe")
	ErrCodigoDuplicado       = httpx.NewError(http.StatusConflict, "codigo_duplicado", "ya existe una ubicación activa con ese código en la empresa")
	ErrTipoMovInvalido       = httpx.NewError(http.StatusUnprocessableEntity, "tipo_movimiento_invalido", "el tipo de movimiento indicado no existe o está inactivo")
	ErrReferenciaInvalida    = httpx.NewError(http.StatusUnprocessableEntity, "referencia_invalida", "alguna referencia (sucursal, artículo, ubicación o tercero) no es válida")
	ErrUbicacionRequerida    = httpx.NewError(http.StatusBadRequest, "ubicacion_requerida", "debes indicar la ubicación de origen o destino según el tipo de movimiento")
	ErrTrasladoMismaUbic     = httpx.NewError(http.StatusBadRequest, "ubicaciones_iguales", "en un traslado la ubicación de origen y destino deben ser diferentes")
	ErrFechaInvalida         = httpx.NewError(http.StatusBadRequest, "fecha_invalida", "el formato de fecha debe ser YYYY-MM-DD o RFC3339")
	ErrFechasInvalidas       = httpx.NewError(http.StatusBadRequest, "fechas_invalidas", "fecha_hasta debe ser mayor o igual a fecha_desde")
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// =====================================================================
// Sucursales
// =====================================================================

func (s *Service) ListarSucursales(ctx context.Context, empresaID int32) ([]SucursalResponse, error) {
	items, err := s.repo.ListSucursales(ctx, empresaID)
	if err != nil {
		return nil, mapError(err)
	}
	out := make([]SucursalResponse, 0, len(items))
	for i := range items {
		out = append(out, SucursalResponse{
			ID:     items[i].ID,
			Nombre: items[i].Nombre,
			Activo: items[i].Activo,
		})
	}
	return out, nil
}

// =====================================================================
// Ubicaciones
// =====================================================================

func (s *Service) CrearUbicacion(ctx context.Context, empresaID, sucursalDefaultID int32, req CrearUbicacionRequest) (*UbicacionResponse, error) {
	sucursalID := sucursalDefaultID
	if req.SucursalID != nil {
		sucursalID = *req.SucursalID
	}
	activo := true
	if req.Activo != nil {
		activo = *req.Activo
	}

	u := &Ubicacion{
		EmpresaID:   empresaID,
		SucursalID:  sucursalID,
		Codigo:      strings.ToUpper(strings.TrimSpace(req.Codigo)),
		Nombre:      strings.TrimSpace(req.Nombre),
		Descripcion: trimPtr(req.Descripcion),
		Activo:      activo,
	}
	if err := s.repo.CreateUbicacion(ctx, u); err != nil {
		return nil, mapError(err)
	}
	creada, err := s.repo.FindUbicacionByID(ctx, empresaID, u.ID)
	if err != nil {
		return nil, mapError(err)
	}
	resp := toUbicacionResponse(creada)
	return &resp, nil
}

func (s *Service) ListarUbicaciones(ctx context.Context, empresaID int32, q ListarUbicacionesQuery) ([]UbicacionResponse, error) {
	items, err := s.repo.ListUbicaciones(ctx, empresaID, q)
	if err != nil {
		return nil, mapError(err)
	}
	out := make([]UbicacionResponse, 0, len(items))
	for i := range items {
		out = append(out, toUbicacionResponse(&items[i]))
	}
	return out, nil
}

func (s *Service) ActualizarUbicacion(ctx context.Context, empresaID, id int32, req ActualizarUbicacionRequest) (*UbicacionResponse, error) {
	if _, err := s.repo.FindUbicacionByID(ctx, empresaID, id); err != nil {
		return nil, mapError(err)
	}
	cambios := map[string]any{}
	if req.SucursalID != nil {
		cambios["sucursal_id"] = *req.SucursalID
	}
	if req.Codigo != nil {
		cambios["codigo"] = strings.ToUpper(strings.TrimSpace(*req.Codigo))
	}
	if req.Nombre != nil {
		cambios["nombre"] = strings.TrimSpace(*req.Nombre)
	}
	if req.Descripcion != nil {
		cambios["descripcion"] = trimPtr(req.Descripcion)
	}
	if req.Activo != nil {
		cambios["activo"] = *req.Activo
	}
	if err := s.repo.UpdateUbicacion(ctx, empresaID, id, cambios); err != nil {
		return nil, mapError(err)
	}
	actualizada, err := s.repo.FindUbicacionByID(ctx, empresaID, id)
	if err != nil {
		return nil, mapError(err)
	}
	resp := toUbicacionResponse(actualizada)
	return &resp, nil
}

func (s *Service) EliminarUbicacion(ctx context.Context, empresaID, id int32) error {
	return mapError(s.repo.SoftDeleteUbicacion(ctx, empresaID, id))
}

// =====================================================================
// Tipos de movimiento
// =====================================================================

func (s *Service) ListarTiposMovimiento(ctx context.Context) ([]TipoMovimientoResponse, error) {
	items, err := s.repo.ListTiposMovimiento(ctx)
	if err != nil {
		return nil, mapError(err)
	}
	out := make([]TipoMovimientoResponse, 0, len(items))
	for i := range items {
		out = append(out, toTipoMovimientoResponse(&items[i]))
	}
	return out, nil
}

// =====================================================================
// Stock por ubicación
// =====================================================================

func (s *Service) ListarStock(ctx context.Context, empresaID int32, q ListarStockQuery) (*ListaStockResponse, error) {
	q.Page, q.PageSize = normalizePagination(q.Page, q.PageSize)

	items, total, err := s.repo.ListStock(ctx, empresaID, q)
	if err != nil {
		return nil, mapError(err)
	}
	data := make([]StockItemResponse, 0, len(items))
	for i := range items {
		data = append(data, toStockResponse(&items[i]))
	}
	return &ListaStockResponse{
		Data:     data,
		Page:     q.Page,
		PageSize: q.PageSize,
		Total:    total,
	}, nil
}

// =====================================================================
// Movimientos y ajustes de inventario (atómicos sobre inventario_stock)
// =====================================================================

func (s *Service) CrearMovimiento(ctx context.Context, empresaID, sucursalDefaultID int32, usuarioID int64, req CrearMovimientoRequest) (*DocumentoInventarioResponse, error) {
	fecha := time.Now().UTC()
	if req.Fecha != nil && strings.TrimSpace(*req.Fecha) != "" {
		t, err := parseTimestamp(*req.Fecha)
		if err != nil {
			return nil, ErrFechaInvalida
		}
		fecha = t
	}

	sucursalID := sucursalDefaultID
	if req.SucursalID != nil {
		sucursalID = *req.SucursalID
	}

	var creado *DocumentoInventario
	err := s.repo.WithTx(ctx, func(tx *Repository) error {
		tipo, err := tx.FindTipoMovimientoByID(ctx, req.TipoMovimientoID)
		if err != nil {
			return err
		}

		// Validar ubicaciones según la naturaleza del movimiento y recolectar IDs
		artIDs := make([]int32, 0, len(req.Detalles))
		ubicIDs := make([]int32, 0, len(req.Detalles)*2)

		for _, line := range req.Detalles {
			artIDs = append(artIDs, line.ArticuloID)
			switch tipo.Naturaleza {
			case NaturalezaEntrada:
				if line.UbicacionDestinoID == nil {
					return ErrUbicacionRequerida
				}
				ubicIDs = append(ubicIDs, *line.UbicacionDestinoID)
			case NaturalezaSalida:
				if line.UbicacionOrigenID == nil {
					return ErrUbicacionRequerida
				}
				ubicIDs = append(ubicIDs, *line.UbicacionOrigenID)
			case NaturalezaTraslado:
				if line.UbicacionOrigenID == nil || line.UbicacionDestinoID == nil {
					return ErrUbicacionRequerida
				}
				if *line.UbicacionOrigenID == *line.UbicacionDestinoID {
					return ErrTrasladoMismaUbic
				}
				ubicIDs = append(ubicIDs, *line.UbicacionOrigenID, *line.UbicacionDestinoID)
			}
		}

		if ok, err := tx.ArticulosBelongToEmpresa(ctx, empresaID, artIDs); err != nil {
			return err
		} else if !ok {
			return ErrReferenciaInvalida
		}

		if ok, err := tx.UbicacionesBelongToEmpresa(ctx, empresaID, ubicIDs); err != nil {
			return err
		} else if !ok {
			return ErrReferenciaInvalida
		}

		if req.TerceroID != nil {
			if ok, err := tx.TerceroBelongsToEmpresa(ctx, empresaID, *req.TerceroID); err != nil {
				return err
			} else if !ok {
				return ErrReferenciaInvalida
			}
		}

		consecutivo, err := consecutivos.GenerarConsecutivoTx(ctx, tx.DB(), empresaID, consecutivos.TipoInventario, fecha)
		if err != nil {
			return err
		}

		doc := &DocumentoInventario{
			EmpresaID:           empresaID,
			SucursalID:          sucursalID,
			TipoMovimientoID:    tipo.ID,
			Consecutivo:         consecutivo,
			Fecha:               fecha,
			TerceroID:           req.TerceroID,
			Motivo:              strings.TrimSpace(req.Motivo),
			Observaciones:       trimPtr(req.Observaciones),
			DocumentoOrigenTipo: trimPtr(req.DocumentoOrigenTipo),
			DocumentoOrigenID:   req.DocumentoOrigenID,
			UsuarioID:           usuarioID,
			Estado:              EstadoAplicado,
		}
		if err := tx.CreateDocumento(ctx, doc); err != nil {
			return err
		}

		detalles := make([]MovimientoInventarioDetalle, 0, len(req.Detalles))
		for _, line := range req.Detalles {
			costo := 0.0
			if line.CostoUnitario != nil {
				costo = *line.CostoUnitario
			}

			switch tipo.Naturaleza {
			case NaturalezaEntrada:
				if err := aplicarEntrada(ctx, tx, empresaID, line.ArticuloID, *line.UbicacionDestinoID, line.Cantidad, costo); err != nil {
					return err
				}
				detalles = append(detalles, MovimientoInventarioDetalle{
					DocumentoID:        doc.ID,
					ArticuloID:         line.ArticuloID,
					UbicacionDestinoID: line.UbicacionDestinoID,
					Cantidad:           line.Cantidad,
					CostoUnitario:      costo,
				})

			case NaturalezaSalida:
				costoAplicado, err := aplicarSalida(ctx, tx, empresaID, line.ArticuloID, *line.UbicacionOrigenID, line.Cantidad, line.CostoUnitario)
				if err != nil {
					return err
				}
				detalles = append(detalles, MovimientoInventarioDetalle{
					DocumentoID:       doc.ID,
					ArticuloID:        line.ArticuloID,
					UbicacionOrigenID: line.UbicacionOrigenID,
					Cantidad:          line.Cantidad,
					CostoUnitario:     costoAplicado,
				})

			case NaturalezaTraslado:
				costoOrigen, err := aplicarSalida(ctx, tx, empresaID, line.ArticuloID, *line.UbicacionOrigenID, line.Cantidad, line.CostoUnitario)
				if err != nil {
					return err
				}
				if err := aplicarEntrada(ctx, tx, empresaID, line.ArticuloID, *line.UbicacionDestinoID, line.Cantidad, costoOrigen); err != nil {
					return err
				}
				detalles = append(detalles, MovimientoInventarioDetalle{
					DocumentoID:        doc.ID,
					ArticuloID:         line.ArticuloID,
					UbicacionOrigenID:  line.UbicacionOrigenID,
					UbicacionDestinoID: line.UbicacionDestinoID,
					Cantidad:           line.Cantidad,
					CostoUnitario:      costoOrigen,
				})
			}
		}

		if err := tx.CreateDetalles(ctx, detalles); err != nil {
			return err
		}

		creado, err = tx.FindDocumentoByID(ctx, empresaID, doc.ID)
		return err
	})
	if err != nil {
		return nil, mapError(err)
	}

	resp := toDocumentoResponse(creado, true)
	return &resp, nil
}

func (s *Service) ListarMovimientos(ctx context.Context, empresaID int32, q ListarMovimientosQuery) (*ListaMovimientosResponse, error) {
	var desde, hasta *time.Time
	if strings.TrimSpace(q.FechaDesde) != "" {
		t, err := parseTimestamp(q.FechaDesde)
		if err != nil {
			return nil, ErrFechaInvalida
		}
		desde = &t
	}
	if strings.TrimSpace(q.FechaHasta) != "" {
		t, err := parseTimestamp(q.FechaHasta)
		if err != nil {
			return nil, ErrFechaInvalida
		}
		hasta = &t
	}
	if desde != nil && hasta != nil && hasta.Before(*desde) {
		return nil, ErrFechasInvalidas
	}

	q.Page, q.PageSize = normalizePagination(q.Page, q.PageSize)

	items, total, err := s.repo.ListDocumentos(ctx, empresaID, q, desde, hasta)
	if err != nil {
		return nil, mapError(err)
	}
	data := make([]DocumentoInventarioResponse, 0, len(items))
	for i := range items {
		data = append(data, toDocumentoResponse(&items[i], false))
	}
	return &ListaMovimientosResponse{
		Data:     data,
		Page:     q.Page,
		PageSize: q.PageSize,
		Total:    total,
	}, nil
}

func (s *Service) ObtenerMovimiento(ctx context.Context, empresaID int32, id int64) (*DocumentoInventarioResponse, error) {
	doc, err := s.repo.FindDocumentoByID(ctx, empresaID, id)
	if err != nil {
		return nil, mapError(err)
	}
	resp := toDocumentoResponse(doc, true)
	return &resp, nil
}

// =====================================================================
// Lógica de actualización de inventario_stock
// =====================================================================

func aplicarEntrada(ctx context.Context, tx *Repository, empresaID, articuloID, ubicacionID int32, cantidad, costoUnitario float64) error {
	stock, err := tx.FindStockForUpdate(ctx, empresaID, articuloID, ubicacionID)
	if err != nil {
		return err
	}
	if stock == nil {
		stock = &InventarioStock{
			EmpresaID:     empresaID,
			ArticuloID:    articuloID,
			UbicacionID:   ubicacionID,
			Cantidad:      round4(cantidad),
			CostoPromedio: round4(costoUnitario),
		}
		return tx.UpsertStock(ctx, stock)
	}

	nuevaCantidad := round4(stock.Cantidad + cantidad)
	nuevoCosto := stock.CostoPromedio
	if costoUnitario > 0 && nuevaCantidad > 0 {
		valorAnterior := stock.Cantidad * stock.CostoPromedio
		valorEntrada := cantidad * costoUnitario
		nuevoCosto = round4((valorAnterior + valorEntrada) / nuevaCantidad)
	}
	stock.Cantidad = nuevaCantidad
	stock.CostoPromedio = nuevoCosto
	return tx.UpsertStock(ctx, stock)
}

func aplicarSalida(ctx context.Context, tx *Repository, empresaID, articuloID, ubicacionID int32, cantidad float64, costoInput *float64) (float64, error) {
	stock, err := tx.FindStockForUpdate(ctx, empresaID, articuloID, ubicacionID)
	if err != nil {
		return 0, err
	}
	disponible := 0.0
	if stock != nil {
		disponible = stock.Cantidad
	}
	if disponible < cantidad {
		return 0, httpx.NewError(
			http.StatusConflict,
			"stock_insuficiente",
			fmt.Sprintf("stock insuficiente para el artículo #%d en la ubicación #%d (disponible: %.2f, requerido: %.2f)", articuloID, ubicacionID, disponible, cantidad),
		)
	}

	costoAplicado := stock.CostoPromedio
	if costoInput != nil && *costoInput > 0 {
		costoAplicado = *costoInput
	}
	stock.Cantidad = round4(stock.Cantidad - cantidad)
	if err := tx.UpsertStock(ctx, stock); err != nil {
		return 0, err
	}
	return costoAplicado, nil
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

func round4(v float64) float64 {
	return math.Round(v*10000) / 10000
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

func parseTimestamp(s string) (time.Time, error) {
	v := strings.TrimSpace(s)
	if t, err := time.Parse(time.RFC3339, v); err == nil {
		return t.UTC(), nil
	}
	if t, err := time.Parse(dateLayout, v); err == nil {
		return t.UTC(), nil
	}
	return time.Time{}, ErrFechaInvalida
}

func mapError(err error) error {
	if err == nil {
		return nil
	}
	var appErr *httpx.AppError
	if errors.As(err, &appErr) {
		return err
	}
	if errors.Is(err, errUbicacionNotFound) {
		return ErrUbicacionNoEncontrada
	}
	if errors.Is(err, errTipoMovNotFound) {
		return ErrTipoMovInvalido
	}
	if errors.Is(err, errDocumentoNotFound) {
		return ErrDocumentoNoEncontrado
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return ErrCodigoDuplicado
		case "23503":
			return ErrReferenciaInvalida
		}
	}
	return err
}
