package compras

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
)

const (
	defaultPageSize = 20
	maxPageSize     = 100
)

var (
	ErrNoEncontrada        = httpx.NewError(http.StatusNotFound, "compra_no_encontrada", "la compra no existe")
	ErrConsecutivoDuplicado = httpx.NewError(http.StatusConflict, "consecutivo_duplicado", "ya existe una compra activa con ese consecutivo en la empresa")
	ErrReferenciaInvalida  = httpx.NewError(http.StatusUnprocessableEntity, "referencia_invalida", "alguna de las referencias indicadas (sucursal, moneda, usuario o artículo) no es válida")
	ErrArticulosInvalidos  = httpx.NewError(http.StatusUnprocessableEntity, "articulo_invalido", "uno o más artículos indicados no existen o no pertenecen a la empresa")
	ErrFechaInvalida       = httpx.NewError(http.StatusBadRequest, "fecha_invalida", "el formato de fecha debe ser YYYY-MM-DD o RFC3339")
	ErrFechasInvalidas     = httpx.NewError(http.StatusBadRequest, "fechas_invalidas", "fecha_vencimiento debe ser mayor o igual a fecha_compra")
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Crear(ctx context.Context, empresaID, sucursalDefaultID int32, usuarioID int64, req CrearCompraRequest) (*CompraResponse, error) {
	fechaCompra, err := parseTimestamp(req.FechaCompra)
	if err != nil {
		return nil, ErrFechaInvalida
	}
	fechaVencimiento, err := parseTimestampPtr(req.FechaVencimiento)
	if err != nil {
		return nil, ErrFechaInvalida
	}
	if fechaVencimiento != nil && fechaVencimiento.Before(fechaCompra) {
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

	detalles, subtotal, totalIVA, total := buildDetalles(req.Detalles)
	articuloIDs := extractArticuloIDs(req.Detalles)

	var creada *Compra
	err = s.repo.WithTx(ctx, func(tx *Repository) error {
		ok, err := tx.ArticulosBelongToEmpresa(ctx, empresaID, articuloIDs)
		if err != nil {
			return err
		}
		if !ok {
			return ErrArticulosInvalidos
		}

		consecutivo := strings.TrimSpace(req.Consecutivo)
		if consecutivo == "" {
			consecutivo, err = consecutivos.GenerarConsecutivoTx(ctx, tx.DB(), empresaID, consecutivos.TipoCompra, fechaCompra)
			if err != nil {
				return err
			}
		}

		c := &Compra{
			EmpresaID:              empresaID,
			SucursalID:             sucursalID,
			TerceroEmpresaRolID:    req.TerceroEmpresaRolID,
			ProveedorNombre:        trimPtr(req.ProveedorNombre),
			Consecutivo:            consecutivo,
			NumeroFacturaProveedor: trimPtr(req.NumeroFacturaProveedor),
			MonedaID:               req.MonedaID,
			Estado:                 estado,
			FechaCompra:            fechaCompra,
			FechaVencimiento:       fechaVencimiento,
			Subtotal:               subtotal,
			TotalIVA:               totalIVA,
			Total:                  total,
			Observaciones:          trimPtr(req.Observaciones),
			UsuarioID:              usuarioID,
		}
		if err := tx.CreateCompra(ctx, c); err != nil {
			return err
		}

		for i := range detalles {
			detalles[i].CompraID = c.ID
		}
		if err := tx.CreateDetalles(ctx, detalles); err != nil {
			return err
		}

		if c.Estado == EstadoRegistrada || c.Estado == EstadoPagada {
			lineasInv := make([]inventario.EntradaCompraLinea, 0, len(detalles))
			for _, d := range detalles {
				lineasInv = append(lineasInv, inventario.EntradaCompraLinea{
					ArticuloID:    d.ArticuloID,
					Cantidad:      d.Cantidad,
					CostoUnitario: d.CostoUnitario,
				})
			}
			if err := inventario.RegistrarEntradaCompraTx(ctx, tx.DB(), inventario.EntradaCompraParams{
				EmpresaID:     empresaID,
				SucursalID:    c.SucursalID,
				UsuarioID:     usuarioID,
				CompraID:      c.ID,
				Consecutivo:   c.Consecutivo,
				Fecha:         c.FechaCompra,
				Observaciones: c.Observaciones,
				Detalles:      lineasInv,
			}); err != nil {
				return err
			}
		}

		creada, err = tx.FindByID(ctx, empresaID, c.ID)
		return err
	})
	if err != nil {
		return nil, mapError(err)
	}

	resp := toResponse(creada, true)
	return &resp, nil
}

func (s *Service) Obtener(ctx context.Context, empresaID int32, id int64) (*CompraResponse, error) {
	c, err := s.repo.FindByID(ctx, empresaID, id)
	if err != nil {
		return nil, mapError(err)
	}
	resp := toResponse(c, true)
	return &resp, nil
}

func (s *Service) Listar(ctx context.Context, empresaID int32, q ListarComprasQuery) (*ListaComprasResponse, error) {
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

	data := make([]CompraResponse, 0, len(items))
	for i := range items {
		data = append(data, toResponse(&items[i], false))
	}
	return &ListaComprasResponse{
		Data:     data,
		Page:     q.Page,
		PageSize: q.PageSize,
		Total:    total,
	}, nil
}

func (s *Service) Actualizar(ctx context.Context, empresaID int32, id int64, req ActualizarCompraRequest) (*CompraResponse, error) {
	var parsedFechaCompra *time.Time
	if req.FechaCompra != nil {
		t, err := parseTimestamp(*req.FechaCompra)
		if err != nil {
			return nil, ErrFechaInvalida
		}
		parsedFechaCompra = &t
	}
	var parsedFechaVenc *time.Time
	if !req.LimpiarFechaVencimiento && req.FechaVencimiento != nil {
		t, err := parseTimestampPtr(req.FechaVencimiento)
		if err != nil {
			return nil, ErrFechaInvalida
		}
		parsedFechaVenc = t
	}

	var actualizada *Compra
	err := s.repo.WithTx(ctx, func(tx *Repository) error {
		actual, err := tx.FindByID(ctx, empresaID, id)
		if err != nil {
			return err
		}

		effCompra := actual.FechaCompra
		effVenc := actual.FechaVencimiento

		cambios := map[string]any{}
		if req.SucursalID != nil {
			cambios["sucursal_id"] = *req.SucursalID
		}
		if req.TerceroEmpresaRolID != nil {
			cambios["tercero_empresa_rol_id"] = *req.TerceroEmpresaRolID
		}
		if req.ProveedorNombre != nil {
			cambios["proveedor_nombre"] = trimPtr(req.ProveedorNombre)
		}
		if req.Consecutivo != nil {
			cambios["consecutivo"] = strings.TrimSpace(*req.Consecutivo)
		}
		if req.NumeroFacturaProveedor != nil {
			cambios["numero_factura_proveedor"] = trimPtr(req.NumeroFacturaProveedor)
		}
		if req.MonedaID != nil {
			cambios["moneda_id"] = *req.MonedaID
		}
		if req.Estado != nil {
			cambios["estado"] = *req.Estado
		}
		if parsedFechaCompra != nil {
			effCompra = *parsedFechaCompra
			cambios["fecha_compra"] = *parsedFechaCompra
		}
		if req.LimpiarFechaVencimiento {
			effVenc = nil
			cambios["fecha_vencimiento"] = nil
		} else if req.FechaVencimiento != nil {
			effVenc = parsedFechaVenc
			cambios["fecha_vencimiento"] = parsedFechaVenc
		}
		if effVenc != nil && effVenc.Before(effCompra) {
			return ErrFechasInvalidas
		}
		if req.Observaciones != nil {
			cambios["observaciones"] = trimPtr(req.Observaciones)
		}

		if req.Detalles != nil {
			articuloIDs := extractArticuloIDs(*req.Detalles)
			ok, err := tx.ArticulosBelongToEmpresa(ctx, empresaID, articuloIDs)
			if err != nil {
				return err
			}
			if !ok {
				return ErrArticulosInvalidos
			}

			nuevosDetalles, subtotal, totalIVA, total := buildDetalles(*req.Detalles)
			for i := range nuevosDetalles {
				nuevosDetalles[i].CompraID = id
			}

			if err := tx.DeleteDetallesByCompra(ctx, id); err != nil {
				return err
			}
			if err := tx.CreateDetalles(ctx, nuevosDetalles); err != nil {
				return err
			}

			cambios["subtotal"] = subtotal
			cambios["total_iva"] = totalIVA
			cambios["total"] = total
		}

		if err := tx.UpdateCompra(ctx, empresaID, id, cambios); err != nil {
			return err
		}

		actualizada, err = tx.FindByID(ctx, empresaID, id)
		if err != nil {
			return err
		}
		if actualizada.Estado == EstadoRegistrada || actualizada.Estado == EstadoPagada {
			lineasInv := make([]inventario.EntradaCompraLinea, 0, len(actualizada.Detalles))
			for _, d := range actualizada.Detalles {
				lineasInv = append(lineasInv, inventario.EntradaCompraLinea{
					ArticuloID:    d.ArticuloID,
					Cantidad:      d.Cantidad,
					CostoUnitario: d.CostoUnitario,
				})
			}
			if err := inventario.RegistrarEntradaCompraTx(ctx, tx.DB(), inventario.EntradaCompraParams{
				EmpresaID:     empresaID,
				SucursalID:    actualizada.SucursalID,
				UsuarioID:     actualizada.UsuarioID,
				CompraID:      actualizada.ID,
				Consecutivo:   actualizada.Consecutivo,
				Fecha:         actualizada.FechaCompra,
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

func buildDetalles(inputs []CompraDetalleInput) (detalles []CompraDetalle, subtotal, totalIVA, total float64) {
	detalles = make([]CompraDetalle, 0, len(inputs))
	for _, in := range inputs {
		costo := *in.CostoUnitario
		impuesto := 0.0
		if in.Impuesto != nil {
			impuesto = *in.Impuesto
		}
		sub := round4(in.Cantidad * costo)
		imp := round4(impuesto)
		tot := round4(sub + imp)

		detalles = append(detalles, CompraDetalle{
			ArticuloID:    in.ArticuloID,
			Descripcion:   trimPtr(in.Descripcion),
			Cantidad:      in.Cantidad,
			CostoUnitario: costo,
			Subtotal:      sub,
			Impuesto:      imp,
			Total:         tot,
		})
		subtotal = round4(subtotal + sub)
		totalIVA = round4(totalIVA + imp)
		total = round4(total + tot)
	}
	return detalles, subtotal, totalIVA, total
}

func extractArticuloIDs(inputs []CompraDetalleInput) []int32 {
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
