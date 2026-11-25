package productos

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"

	"tesserak/backend/internal/httpx"
)

const (
	defaultPageSize = 20
	maxPageSize     = 100
)

var (
	errTipoArticuloNotFound = errors.New("tipo de artículo no encontrado")

	ErrNoEncontrado              = httpx.NewError(http.StatusNotFound, "articulo_no_encontrado", "el artículo no existe")
	ErrTipoArticuloNoEncontrado  = httpx.NewError(http.StatusUnprocessableEntity, "tipo_articulo_invalido", "el tipo de artículo indicado no existe")
	ErrCodigoDuplicado           = httpx.NewError(http.StatusConflict, "codigo_duplicado", "ya existe un artículo activo con ese código en la empresa")
	ErrReferenciaInvalida        = httpx.NewError(http.StatusUnprocessableEntity, "referencia_invalida", "alguna de las referencias indicadas (unidad, categoría, marca, impuesto, método de valoración o modalidad) no es válida")
	ErrMetodoValoracionRequerido = httpx.NewError(http.StatusBadRequest, "metodo_valoracion_requerido", "cuando control_inventario está activo debes seleccionar metodo_valoracion_id")
	ErrModalidadRequerida        = httpx.NewError(http.StatusBadRequest, "modalidad_requerida", "para un servicio debes seleccionar modalidad_id")
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Crear(ctx context.Context, empresaID int32, req CrearProductoRequest) (*ProductoResponse, error) {
	estado := true
	if req.Estado != nil {
		estado = *req.Estado
	}

	var creado *Articulo
	err := s.repo.WithTx(ctx, func(tx *Repository) error {
		tipo, err := tx.FindTipoArticuloByID(ctx, req.TipoArticuloID)
		if err != nil {
			return err
		}

		// Validaciones previas por subtipo
		var prodExt *ProductoExtension
		var servExt *ServicioExtension

		switch tipo.Codigo {
		case TipoCodigoProducto:
			controlInv := true
			permiteCompra := true
			permiteVenta := true
			var metodoID *int32
			if req.Producto != nil {
				if req.Producto.ControlInventario != nil {
					controlInv = *req.Producto.ControlInventario
				}
				if req.Producto.PermiteCompra != nil {
					permiteCompra = *req.Producto.PermiteCompra
				}
				if req.Producto.PermiteVenta != nil {
					permiteVenta = *req.Producto.PermiteVenta
				}
				metodoID = req.Producto.MetodoValoracionID
			}
			if controlInv && metodoID == nil {
				return ErrMetodoValoracionRequerido
			}
			prodExt = &ProductoExtension{
				ControlInventario:  controlInv,
				MetodoValoracionID: metodoID,
				PermiteCompra:      permiteCompra,
				PermiteVenta:       permiteVenta,
			}

		case TipoCodigoServicio:
			if req.Servicio == nil || req.Servicio.ModalidadID == nil {
				return ErrModalidadRequerida
			}
			reqProg := req.Servicio.RequiereProgramacion != nil && *req.Servicio.RequiereProgramacion
			ctrlCons := req.Servicio.ControlConsumos != nil && *req.Servicio.ControlConsumos
			servExt = &ServicioExtension{
				ModalidadID:          *req.Servicio.ModalidadID,
				DuracionEstimada:     req.Servicio.DuracionEstimada,
				RequiereProgramacion: reqProg,
				ControlConsumos:      ctrlCons,
			}

		default:
			return ErrTipoArticuloNoEncontrado
		}

		a := &Articulo{
			EmpresaID:       empresaID,
			Codigo:          strings.TrimSpace(req.Codigo),
			Nombre:          strings.TrimSpace(req.Nombre),
			Descripcion:     trimPtr(req.Descripcion),
			TipoArticuloID:  req.TipoArticuloID,
			UnidadMedidaID:  req.UnidadMedidaID,
			CategoriaID:     req.CategoriaID,
			MarcaID:         req.MarcaID,
			GrupoImpuestoID: req.GrupoImpuestoID,
			Estado:          estado,
		}
		if err := tx.CreateArticulo(ctx, a); err != nil {
			return err
		}

		if prodExt != nil {
			prodExt.ArticuloID = a.ID
			if err := tx.CreateProductoExt(ctx, prodExt); err != nil {
				return err
			}
		}
		if servExt != nil {
			servExt.ArticuloID = a.ID
			if err := tx.CreateServicioExt(ctx, servExt); err != nil {
				return err
			}
		}

		creado, err = tx.FindByID(ctx, empresaID, a.ID)
		return err
	})
	if err != nil {
		return nil, mapError(err)
	}
	resp := toResponse(creado)
	return &resp, nil
}

func (s *Service) Obtener(ctx context.Context, empresaID, id int32) (*ProductoResponse, error) {
	a, err := s.repo.FindByID(ctx, empresaID, id)
	if err != nil {
		return nil, mapError(err)
	}
	resp := toResponse(a)
	return &resp, nil
}

func (s *Service) Listar(ctx context.Context, empresaID int32, q ListarProductosQuery) (*ListaProductosResponse, error) {
	if q.Page < 1 {
		q.Page = 1
	}
	if q.PageSize < 1 {
		q.PageSize = defaultPageSize
	}
	if q.PageSize > maxPageSize {
		q.PageSize = maxPageSize
	}

	items, total, err := s.repo.List(ctx, empresaID, q)
	if err != nil {
		return nil, err
	}
	data := make([]ProductoResponse, 0, len(items))
	for i := range items {
		data = append(data, toResponse(&items[i]))
	}
	return &ListaProductosResponse{
		Data:     data,
		Page:     q.Page,
		PageSize: q.PageSize,
		Total:    total,
	}, nil
}

func (s *Service) Actualizar(ctx context.Context, empresaID, id int32, req ActualizarProductoRequest) (*ProductoResponse, error) {
	var res *Articulo
	err := s.repo.WithTx(ctx, func(tx *Repository) error {
		a, err := tx.FindByID(ctx, empresaID, id)
		if err != nil {
			return err
		}

		cambios := map[string]any{}
		if req.Codigo != nil {
			cambios["codigo"] = strings.TrimSpace(*req.Codigo)
		}
		if req.Nombre != nil {
			cambios["nombre"] = strings.TrimSpace(*req.Nombre)
		}
		if req.Descripcion != nil {
			cambios["descripcion"] = trimPtr(req.Descripcion)
		}
		if req.UnidadMedidaID != nil {
			cambios["unidad_medida_id"] = *req.UnidadMedidaID
		}
		if req.CategoriaID != nil {
			cambios["categoria_id"] = *req.CategoriaID
		}
		if req.LimpiarMarca {
			cambios["marca_id"] = nil
		} else if req.MarcaID != nil {
			cambios["marca_id"] = *req.MarcaID
		}
		if req.GrupoImpuestoID != nil {
			cambios["grupo_impuesto_id"] = *req.GrupoImpuestoID
		}
		if req.Estado != nil {
			cambios["estado"] = *req.Estado
		}

		if err := tx.UpdateArticulo(ctx, a, cambios); err != nil {
			return err
		}

		// Actualizar extensión según subtipo del artículo
		if a.Producto != nil && req.Producto != nil {
			effControl := a.Producto.ControlInventario
			if req.Producto.ControlInventario != nil {
				effControl = *req.Producto.ControlInventario
			}
			effMetodo := a.Producto.MetodoValoracionID
			if req.Producto.MetodoValoracionID != nil {
				effMetodo = req.Producto.MetodoValoracionID
			}
			if effControl && effMetodo == nil {
				return ErrMetodoValoracionRequerido
			}

			cambiosProd := map[string]any{}
			if req.Producto.ControlInventario != nil {
				cambiosProd["control_inventario"] = *req.Producto.ControlInventario
			}
			if req.Producto.MetodoValoracionID != nil {
				cambiosProd["metodo_valoracion_id"] = *req.Producto.MetodoValoracionID
			}
			if req.Producto.PermiteCompra != nil {
				cambiosProd["permite_compra"] = *req.Producto.PermiteCompra
			}
			if req.Producto.PermiteVenta != nil {
				cambiosProd["permite_venta"] = *req.Producto.PermiteVenta
			}
			if err := tx.UpdateProductoExt(ctx, a.ID, cambiosProd); err != nil {
				return err
			}
		}

		if a.Servicio != nil && req.Servicio != nil {
			cambiosServ := map[string]any{}
			if req.Servicio.ModalidadID != nil {
				cambiosServ["modalidad_id"] = *req.Servicio.ModalidadID
			}
			if req.Servicio.DuracionEstimada != nil {
				cambiosServ["duracion_estimada"] = *req.Servicio.DuracionEstimada
			}
			if req.Servicio.RequiereProgramacion != nil {
				cambiosServ["requiere_programacion"] = *req.Servicio.RequiereProgramacion
			}
			if req.Servicio.ControlConsumos != nil {
				cambiosServ["control_consumos"] = *req.Servicio.ControlConsumos
			}
			if err := tx.UpdateServicioExt(ctx, a.ID, cambiosServ); err != nil {
				return err
			}
		}

		res, err = tx.FindByID(ctx, empresaID, id)
		return err
	})
	if err != nil {
		return nil, mapError(err)
	}
	resp := toResponse(res)
	return &resp, nil
}

func (s *Service) Eliminar(ctx context.Context, empresaID, id int32) error {
	return mapError(s.repo.SoftDelete(ctx, empresaID, id))
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
	if errors.Is(err, errNotFound) {
		return ErrNoEncontrado
	}
	if errors.Is(err, errTipoArticuloNotFound) {
		return ErrTipoArticuloNoEncontrado
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
