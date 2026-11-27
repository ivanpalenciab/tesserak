package empresas

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"

	"tesserak/backend/internal/httpx"
	"tesserak/backend/internal/terceros"
)

const (
	defaultPageSize = 20
	maxPageSize     = 100
)

var (
	ErrEmpresaNoEncontrada         = httpx.NewError(http.StatusNotFound, "empresa_no_encontrada", "la empresa no existe")
	ErrIdentificadorFiscalDuplicado = httpx.NewError(http.StatusConflict, "identificador_fiscal_duplicado", "ya existe una empresa con ese identificador fiscal")
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Crear(ctx context.Context, req CrearEmpresaRequest) (*EmpresaResponse, error) {
	activo := true
	if req.Activo != nil {
		activo = *req.Activo
	}

	e := &Empresa{
		NombreLegal:         strings.TrimSpace(req.NombreLegal),
		RazonSocial:         strings.TrimSpace(req.RazonSocial),
		IdentificadorFiscal: strings.TrimSpace(req.IdentificadorFiscal),
		Email:               normalizarEmailPtr(req.Email),
		Direccion:           trimPtr(req.Direccion),
		Logo:                trimPtr(req.Logo),
		Activo:              activo,
	}
	err := s.repo.WithTx(ctx, func(tx *Repository) error {
		if err := tx.Create(ctx, e); err != nil {
			return err
		}
		if _, err := terceros.AsegurarConsumidorFinalTx(ctx, tx.DB(), e.ID); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, mapError(err)
	}
	resp := toResponse(e)
	return &resp, nil
}

func (s *Service) Obtener(ctx context.Context, id int32) (*EmpresaResponse, error) {
	e, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, mapError(err)
	}
	resp := toResponse(e)
	return &resp, nil
}

func (s *Service) Listar(ctx context.Context, q ListarEmpresasQuery) (*ListaEmpresasResponse, error) {
	if q.Page < 1 {
		q.Page = 1
	}
	if q.PageSize < 1 {
		q.PageSize = defaultPageSize
	}
	if q.PageSize > maxPageSize {
		q.PageSize = maxPageSize
	}

	items, total, err := s.repo.List(ctx, q)
	if err != nil {
		return nil, err
	}

	data := make([]EmpresaResponse, 0, len(items))
	for i := range items {
		data = append(data, toResponse(&items[i]))
	}
	return &ListaEmpresasResponse{Data: data, Page: q.Page, PageSize: q.PageSize, Total: total}, nil
}

func (s *Service) Actualizar(ctx context.Context, id int32, req ActualizarEmpresaRequest) (*EmpresaResponse, error) {
	var result *Empresa

	err := s.repo.WithTx(ctx, func(tx *Repository) error {
		e, err := tx.FindByID(ctx, id)
		if err != nil {
			return err
		}

		cambios := map[string]any{}

		if req.NombreLegal != nil {
			cambios["nombre_legal"] = strings.TrimSpace(*req.NombreLegal)
		}
		if req.RazonSocial != nil {
			cambios["razon_social"] = strings.TrimSpace(*req.RazonSocial)
		}
		if req.IdentificadorFiscal != nil {
			cambios["identificador_fiscal"] = strings.TrimSpace(*req.IdentificadorFiscal)
		}
		if req.Email != nil {
			cambios["email"] = normalizarEmailPtr(req.Email)
		}
		if req.Direccion != nil {
			cambios["direccion"] = trimPtr(req.Direccion)
		}
		if req.Logo != nil {
			cambios["logo"] = trimPtr(req.Logo)
		}
		if req.Activo != nil {
			cambios["activo"] = *req.Activo
		}

		if err := tx.Update(ctx, e, cambios); err != nil {
			return err
		}

		result, err = tx.FindByID(ctx, id)
		return err
	})
	if err != nil {
		return nil, mapError(err)
	}

	resp := toResponse(result)
	return &resp, nil
}

func (s *Service) Eliminar(ctx context.Context, id int32) error {
	return mapError(s.repo.SoftDelete(ctx, id))
}

// ---------------------------------------------------------------------

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

	if errors.Is(err, errNotFound) {
		return ErrEmpresaNoEncontrada
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code == "23505" { // unique_violation
			if pgErr.ConstraintName == "ux_empresas_identificador_fiscal" {
				return ErrIdentificadorFiscalDuplicado
			}
			return httpx.NewError(http.StatusConflict, "duplicado", "el registro ya existe")
		}
	}
	return err
}
