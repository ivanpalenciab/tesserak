package ventas

import (
	"context"
	"errors"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"

	"tesserak/backend/internal/consecutivos"
	"tesserak/backend/internal/httpx"
	"tesserak/backend/internal/inventario"
	"tesserak/backend/internal/terceros"
)

const (
	defaultPageSize = 20
	maxPageSize     = 100
)

var (
	ErrNoEncontrada         = httpx.NewError(http.StatusNotFound, "venta_no_encontrada", "la venta no existe")
	ErrConsecutivoDuplicado = httpx.NewError(http.StatusConflict, "consecutivo_duplicado", "ya existe una venta activa con ese consecutivo en la empresa")
	ErrReferenciaInvalida   = httpx.NewError(http.StatusUnprocessableEntity, "referencia_invalida", "alguna de las referencias indicadas (sucursal, ubicación, cliente, moneda o artículo) no es válida")
	ErrClienteInvalido      = httpx.NewError(http.StatusUnprocessableEntity, "cliente_invalido", "el cliente indicado no existe, está inactivo o no tiene el rol CLIENTE")
	ErrUbicacionRequerida   = httpx.NewError(http.StatusUnprocessableEntity, "ubicacion_requerida", "no existe una ubicación activa configurada para descontar el inventario")
	ErrArticulosInvalidos   = httpx.NewError(http.StatusUnprocessableEntity, "articulo_invalido", "uno o más artículos indicados no existen o no pertenecen a la empresa")
	ErrFechaInvalida        = httpx.NewError(http.StatusBadRequest, "fecha_invalida", "el formato de fecha debe ser YYYY-MM-DD o RFC3339")
	ErrFechasInvalidas      = httpx.NewError(http.StatusBadRequest, "fechas_invalidas", "fecha_vencimiento debe ser mayor o igual a fecha_venta")
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// Crear registra la venta y genera automáticamente la salida de inventario (SALIDA_VENTA)
// dentro de la misma transacción. Si el movimiento de inventario falla (por ejemplo,
// por stock insuficiente), toda la transacción hace rollback y la venta no se guarda.
func (s *Service) Crear(ctx context.Context, empresaID, sucursalDefaultID int32, usuarioID int64, req CrearVentaRequest) (*VentaResponse, error) {
	fechaVenta, err := parseTimestamp(req.FechaVenta)
	if err != nil {
		return nil, ErrFechaInvalida
	}
	fechaVencimiento, err := parseTimestampPtr(req.FechaVencimiento)
	if err != nil {
		return nil, ErrFechaInvalida
	}
	if fechaVencimiento != nil && fechaVencimiento.Before(fechaVenta) {
		return nil, ErrFechasInvalidas
	}

	sucursalID := sucursalDefaultID
	if req.SucursalID != nil {
		sucursalID = *req.SucursalID
	}
	estado := EstadoRegistrada
	if req.Estado != nil {
		estado = *req.Estado
	}

	articuloIDs := extractArticuloIDs(req.Detalles)

	var creada *Venta
	err = s.repo.WithTx(ctx, func(tx *Repository) error {
		okArt, err := tx.ArticulosBelongToEmpresa(ctx, empresaID, articuloIDs)
		if err != nil {
			return err
		}
		if !okArt {
			return ErrArticulosInvalidos
		}

		tarifas, err := tx.GetTarifasImpuestoArticulos(ctx, empresaID, articuloIDs)
		if err != nil {
			return err
		}
		detalles, subtotal, totalIVA, total := buildDetalles(req.Detalles, tarifas)

		// Cliente: si no se envía cliente_id, usa automáticamente Consumidor Final (222222222222)
		var clienteID int64
		if req.ClienteID != nil {
			okCli, err := tx.ClienteBelongsToEmpresa(ctx, empresaID, *req.ClienteID)
			if err != nil {
				return err
			}
			if !okCli {
				return ErrClienteInvalido
			}
			clienteID = *req.ClienteID
		} else {
			clienteID, err = terceros.AsegurarConsumidorFinalTx(ctx, tx.DB(), empresaID)
			if err != nil {
				return err
			}
		}

		// Ubicación para descontar inventario
		var ubicacionID int32
		if req.UbicacionID != nil {
			ubicacionID = *req.UbicacionID
		} else {
			ubicacionID, err = tx.ResolveUbicacionDefault(ctx, empresaID, sucursalID)
			if err != nil {
				return err
			}
		}

		consecutivo := strings.TrimSpace(req.Consecutivo)
		if consecutivo == "" {
			consecutivo, err = consecutivos.GenerarConsecutivoTx(ctx, tx.DB(), empresaID, consecutivos.TipoVenta, fechaVenta)
			if err != nil {
				return err
			}
		}

		v := &Venta{
			EmpresaID:        empresaID,
			SucursalID:       sucursalID,
			UbicacionID:      ubicacionID,
			ClienteID:        clienteID,
			Consecutivo:      consecutivo,
			MonedaID:         req.MonedaID,
			Estado:           estado,
			FechaVenta:       fechaVenta,
			FechaVencimiento: fechaVencimiento,
			Subtotal:         subtotal,
			TotalIVA:         totalIVA,
			Total:            total,
			Observaciones:    trimPtr(req.Observaciones),
			UsuarioID:        usuarioID,
		}
		if err := tx.CreateVenta(ctx, v); err != nil {
			return err
		}

		for i := range detalles {
			detalles[i].VentaID = v.ID
		}
		if err := tx.CreateDetalles(ctx, detalles); err != nil {
			return err
		}

		if v.Estado == EstadoRegistrada || v.Estado == EstadoPagada {
			lineasInv := make([]inventario.SalidaVentaLinea, 0, len(detalles))
			for _, d := range detalles {
				lineasInv = append(lineasInv, inventario.SalidaVentaLinea{
					ArticuloID: d.ArticuloID,
					Cantidad:   d.Cantidad,
				})
			}
			if err := inventario.RegistrarSalidaVentaTx(ctx, tx.DB(), inventario.SalidaVentaParams{
				EmpresaID:     empresaID,
				SucursalID:    v.SucursalID,
				UbicacionID:   v.UbicacionID,
				ClienteID:     v.ClienteID,
				UsuarioID:     usuarioID,
				VentaID:       v.ID,
				Consecutivo:   v.Consecutivo,
				Fecha:         v.FechaVenta,
				Observaciones: v.Observaciones,
				Detalles:      lineasInv,
			}); err != nil {
				return err
			}
		}

		creada, err = tx.FindByID(ctx, empresaID, v.ID)
		return err
	})
	if err != nil {
		return nil, mapError(err)
	}

	resp := toResponse(creada, true)
	return &resp, nil
}

func (s *Service) Obtener(ctx context.Context, empresaID int32, id int64) (*VentaResponse, error) {
	v, err := s.repo.FindByID(ctx, empresaID, id)
	if err != nil {
		return nil, mapError(err)
	}
	resp := toResponse(v, true)
	return &resp, nil
}

func (s *Service) Listar(ctx context.Context, empresaID int32, q ListarVentasQuery) (*ListaVentasResponse, error) {
	var fechaDesde, fechaHasta *time.Time
	if strings.TrimSpace(q.FechaDesde) != "" {
		t, err := parseTimestamp(q.FechaDesde)
		if err != nil {
			return nil, ErrFechaInvalida
		}
		fechaDesde = &t
	}
	if strings.TrimSpace(q.FechaHasta) != "" {
		t, err := parseTimestamp(q.FechaHasta)
		if err != nil {
			return nil, ErrFechaInvalida
		}
		fechaHasta = &t
	}
	if fechaDesde != nil && fechaHasta != nil && fechaHasta.Before(*fechaDesde) {
		return nil, ErrFechasInvalidas
	}

	if q.Page < 1 {
		q.Page = 1
	}
	if q.PageSize < 1 {
		q.PageSize = defaultPageSize
	}
	if q.PageSize > maxPageSize {
		q.PageSize = maxPageSize
	}

	items, total, err := s.repo.List(ctx, empresaID, q, fechaDesde, fechaHasta)
	if err != nil {
		return nil, mapError(err)
	}

	data := make([]VentaResponse, 0, len(items))
	for i := range items {
		data = append(data, toResponse(&items[i], false))
	}
	return &ListaVentasResponse{
		Data:     data,
		Page:     q.Page,
		PageSize: q.PageSize,
		Total:    total,
	}, nil
}

func (s *Service) Actualizar(ctx context.Context, empresaID int32, id int64, req ActualizarVentaRequest) (*VentaResponse, error) {
	var parsedFechaVenta *time.Time
	if req.FechaVenta != nil {
		t, err := parseTimestamp(*req.FechaVenta)
		if err != nil {
			return nil, ErrFechaInvalida
		}
		parsedFechaVenta = &t
	}
	var parsedFechaVenc *time.Time
	if !req.LimpiarFechaVencimiento && req.FechaVencimiento != nil {
		t, err := parseTimestampPtr(req.FechaVencimiento)
		if err != nil {
			return nil, ErrFechaInvalida
		}
		parsedFechaVenc = t
	}

	var actualizada *Venta
	err := s.repo.WithTx(ctx, func(tx *Repository) error {
		actual, err := tx.FindByID(ctx, empresaID, id)
		if err != nil {
			return err
		}

		effVenta := actual.FechaVenta
		effVenc := actual.FechaVencimiento

		cambios := map[string]any{}
		if req.SucursalID != nil {
			cambios["sucursal_id"] = *req.SucursalID
		}
		if req.UbicacionID != nil {
			cambios["ubicacion_id"] = *req.UbicacionID
		}
		if req.ClienteID != nil {
			okCli, err := tx.ClienteBelongsToEmpresa(ctx, empresaID, *req.ClienteID)
			if err != nil {
				return err
			}
			if !okCli {
				return ErrClienteInvalido
			}
			cambios["cliente_id"] = *req.ClienteID
		}
		if req.Consecutivo != nil {
			cambios["consecutivo"] = strings.TrimSpace(*req.Consecutivo)
		}
		if req.MonedaID != nil {
			cambios["moneda_id"] = *req.MonedaID
		}
		if req.Estado != nil {
			cambios["estado"] = *req.Estado
		}
		if parsedFechaVenta != nil {
			effVenta = *parsedFechaVenta
			cambios["fecha_venta"] = *parsedFechaVenta
		}
		if req.LimpiarFechaVencimiento {
			effVenc = nil
			cambios["fecha_vencimiento"] = nil
		} else if req.FechaVencimiento != nil {
			effVenc = parsedFechaVenc
			cambios["fecha_vencimiento"] = parsedFechaVenc
		}
		if effVenc != nil && effVenc.Before(effVenta) {
			return ErrFechasInvalidas
		}
		if req.Observaciones != nil {
			cambios["observaciones"] = trimPtr(req.Observaciones)
		}

		if req.Detalles != nil {
			articuloIDs := extractArticuloIDs(*req.Detalles)
			okArt, err := tx.ArticulosBelongToEmpresa(ctx, empresaID, articuloIDs)
			if err != nil {
				return err
			}
			if !okArt {
				return ErrArticulosInvalidos
			}

			tarifas, err := tx.GetTarifasImpuestoArticulos(ctx, empresaID, articuloIDs)
			if err != nil {
				return err
			}
			nuevosDetalles, subtotal, totalIVA, total := buildDetalles(*req.Detalles, tarifas)
			for i := range nuevosDetalles {
				nuevosDetalles[i].VentaID = id
			}
			if err := tx.DeleteDetallesByVenta(ctx, id); err != nil {
				return err
			}
			if err := tx.CreateDetalles(ctx, nuevosDetalles); err != nil {
				return err
			}

			cambios["subtotal"] = subtotal
			cambios["total_iva"] = totalIVA
			cambios["total"] = total
		}

		if err := tx.UpdateVenta(ctx, empresaID, id, cambios); err != nil {
			return err
		}

		actualizada, err = tx.FindByID(ctx, empresaID, id)
		if err != nil {
			return err
		}

		if actualizada.Estado == EstadoRegistrada || actualizada.Estado == EstadoPagada {
			lineasInv := make([]inventario.SalidaVentaLinea, 0, len(actualizada.Detalles))
			for _, d := range actualizada.Detalles {
				lineasInv = append(lineasInv, inventario.SalidaVentaLinea{
					ArticuloID: d.ArticuloID,
					Cantidad:   d.Cantidad,
				})
			}
			if err := inventario.RegistrarSalidaVentaTx(ctx, tx.DB(), inventario.SalidaVentaParams{
				EmpresaID:     empresaID,
				SucursalID:    actualizada.SucursalID,
				UbicacionID:   actualizada.UbicacionID,
				ClienteID:     actualizada.ClienteID,
				UsuarioID:     actualizada.UsuarioID,
				VentaID:       actualizada.ID,
				Consecutivo:   actualizada.Consecutivo,
				Fecha:         actualizada.FechaVenta,
				Observaciones: actualizada.Observaciones,
				Detalles:      lineasInv,
			}); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, mapError(err)
	}

	resp := toResponse(actualizada, true)
	return &resp, nil
}

func (s *Service) Eliminar(ctx context.Context, empresaID int32, id int64) error {
	return mapError(s.repo.SoftDelete(ctx, empresaID, id))
}

func buildDetalles(inputs []VentaDetalleInput, tarifas map[int32]float64) (detalles []VentaDetalle, subtotal, totalIVA, total float64) {
	detalles = make([]VentaDetalle, 0, len(inputs))
	for _, in := range inputs {
		precio := *in.PrecioUnitario
		sub := round4(in.Cantidad * precio)
		tarifa := tarifas[in.ArticuloID]
		imp := round4(sub * tarifa / 100.0)
		tot := round4(sub + imp)

		detalles = append(detalles, VentaDetalle{
			ArticuloID:     in.ArticuloID,
			Descripcion:    trimPtr(in.Descripcion),
			Cantidad:       in.Cantidad,
			PrecioUnitario: precio,
			Subtotal:       sub,
			Impuesto:       imp,
			Total:          tot,
		})
		subtotal = round4(subtotal + sub)
		totalIVA = round4(totalIVA + imp)
		total = round4(total + tot)
	}
	return detalles, subtotal, totalIVA, total
}

func extractArticuloIDs(inputs []VentaDetalleInput) []int32 {
	ids := make([]int32, 0, len(inputs))
	for _, in := range inputs {
		ids = append(ids, in.ArticuloID)
	}
	return ids
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

func parseTimestampPtr(p *string) (*time.Time, error) {
	if p == nil {
		return nil, nil
	}
	v := strings.TrimSpace(*p)
	if v == "" {
		return nil, nil
	}
	t, err := parseTimestamp(v)
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
	if errors.Is(err, errNotFound) {
		return ErrNoEncontrada
	}
	if errors.Is(err, errUbicacionNotFound) {
		return ErrUbicacionRequerida
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return ErrConsecutivoDuplicado
		case "23503":
			return ErrReferenciaInvalida
		case "23514":
			return ErrFechasInvalidas
		}
	}
	return err
}
