package terceros

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
	ErrTerceroNoEncontrado     = httpx.NewError(http.StatusNotFound, "tercero_no_encontrado", "el tercero no existe")
	ErrIdentificacionDuplicada = httpx.NewError(http.StatusConflict, "identificacion_duplicada", "ya existe un tercero activo con ese tipo y número de identificación en la empresa")
	ErrReferenciaInvalida      = httpx.NewError(http.StatusUnprocessableEntity, "referencia_invalida", "el país o rol indicado no es válido")
	ErrRolesInvalidos          = httpx.NewError(http.StatusUnprocessableEntity, "roles_invalidos", "uno o más roles seleccionados no existen o están inactivos")
)

func (s *Service) CrearTercero(ctx context.Context, empresaID int32, req CrearTerceroRequest) (*TerceroResponse, error) {
	rolIDs := deduplicateInt32(req.RolIDs)
	activo := true
	if req.Activo != nil {
		activo = *req.Activo
	}

	var creado *TerceroMaestro
	err := s.repo.WithTx(ctx, func(tx *Repository) error {
		ok, err := tx.RolesActivosExist(ctx, rolIDs)
		if err != nil {
			return err
		}
		if !ok {
			return ErrRolesInvalidos
		}

		t := &TerceroMaestro{
			EmpresaID:            empresaID,
			TipoIdentificacion:   strings.ToUpper(strings.TrimSpace(req.TipoIdentificacion)),
			NumeroIdentificacion: strings.TrimSpace(req.NumeroIdentificacion),
			RazonSocial:          strings.TrimSpace(req.RazonSocial),
			PaisID:               req.PaisID,
			NaturalezaJuridica:   strings.ToUpper(strings.TrimSpace(req.NaturalezaJuridica)),
			Email:                trimPtr(req.Email),
			Telefono:             trimPtr(req.Telefono),
			Direccion:            trimPtr(req.Direccion),
			Activo:               activo,
		}
		if err := tx.CreateTercero(ctx, t); err != nil {
			return err
		}
		if err := tx.ReplaceTerceroRoles(ctx, t.ID, rolIDs); err != nil {
			return err
		}

		creado, err = tx.FindTerceroByID(ctx, empresaID, t.ID)
		return err
	})
	if err != nil {
		return nil, mapTerceroError(err)
	}

	resp := toTerceroResponse(creado)
	return &resp, nil
}

func (s *Service) ObtenerTercero(ctx context.Context, empresaID int32, id int64) (*TerceroResponse, error) {
	t, err := s.repo.FindTerceroByID(ctx, empresaID, id)
	if err != nil {
		return nil, mapTerceroError(err)
	}
	resp := toTerceroResponse(t)
	return &resp, nil
}

func (s *Service) ListarTerceros(ctx context.Context, empresaID int32, q ListarTercerosQuery) (*ListaTercerosResponse, error) {
	if q.Page < 1 {
		q.Page = 1
	}
	if q.PageSize < 1 {
		q.PageSize = defaultPageSize
	}
	if q.PageSize > maxPageSize {
		q.PageSize = maxPageSize
	}

	items, total, err := s.repo.ListTerceros(ctx, empresaID, q)
	if err != nil {
		return nil, mapTerceroError(err)
	}
	data := make([]TerceroResponse, 0, len(items))
	for i := range items {
		data = append(data, toTerceroResponse(&items[i]))
	}
	return &ListaTercerosResponse{
		Data:     data,
		Page:     q.Page,
		PageSize: q.PageSize,
		Total:    total,
	}, nil
}

func (s *Service) ActualizarTercero(ctx context.Context, empresaID int32, id int64, req ActualizarTerceroRequest) (*TerceroResponse, error) {
	var actualizado *TerceroMaestro
	err := s.repo.WithTx(ctx, func(tx *Repository) error {
		if _, err := tx.FindTerceroByID(ctx, empresaID, id); err != nil {
			return err
		}

		cambios := map[string]any{}
		if req.TipoIdentificacion != nil {
			cambios["tipo_identificacion"] = strings.ToUpper(strings.TrimSpace(*req.TipoIdentificacion))
		}
		if req.NumeroIdentificacion != nil {
			cambios["numero_identificacion"] = strings.TrimSpace(*req.NumeroIdentificacion)
		}
		if req.RazonSocial != nil {
			cambios["razon_social"] = strings.TrimSpace(*req.RazonSocial)
		}
		if req.LimpiarPais {
			cambios["pais_id"] = nil
		} else if req.PaisID != nil {
			cambios["pais_id"] = *req.PaisID
		}
		if req.NaturalezaJuridica != nil {
			cambios["naturaleza_juridica"] = strings.ToUpper(strings.TrimSpace(*req.NaturalezaJuridica))
		}
		if req.Email != nil {
			cambios["email"] = trimPtr(req.Email)
		}
		if req.Telefono != nil {
			cambios["telefono"] = trimPtr(req.Telefono)
		}
		if req.Direccion != nil {
			cambios["direccion"] = trimPtr(req.Direccion)
		}
		if req.Activo != nil {
			cambios["activo"] = *req.Activo
		}

		if err := tx.UpdateTercero(ctx, empresaID, id, cambios); err != nil {
			return err
		}

		if req.RolIDs != nil {
			rolIDs := deduplicateInt32(*req.RolIDs)
			ok, err := tx.RolesActivosExist(ctx, rolIDs)
			if err != nil {
				return err
			}
			if !ok {
				return ErrRolesInvalidos
			}
			if err := tx.ReplaceTerceroRoles(ctx, id, rolIDs); err != nil {
				return err
			}
		}

		var err error
		actualizado, err = tx.FindTerceroByID(ctx, empresaID, id)
		return err
	})
	if err != nil {
		return nil, mapTerceroError(err)
	}

	resp := toTerceroResponse(actualizado)
	return &resp, nil
}

func (s *Service) EliminarTercero(ctx context.Context, empresaID int32, id int64) error {
	return mapTerceroError(s.repo.SoftDeleteTercero(ctx, empresaID, id))
}

func deduplicateInt32(ids []int32) []int32 {
	seen := make(map[int32]struct{}, len(ids))
	out := make([]int32, 0, len(ids))
	for _, id := range ids {
		if _, ok := seen[id]; !ok {
			seen[id] = struct{}{}
			out = append(out, id)
		}
	}
	return out
}

func mapTerceroError(err error) error {
	if err == nil {
		return nil
	}
	var appErr *httpx.AppError
	if errors.As(err, &appErr) {
		return err
	}
	if errors.Is(err, errTerceroNotFound) {
		return ErrTerceroNoEncontrado
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return ErrIdentificacionDuplicada
		case "23503":
			return ErrReferenciaInvalida
		}
	}
	return err
}
