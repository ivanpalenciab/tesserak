package terceros

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"

	"tesserak/backend/internal/httpx"
)

var (
	ErrRolNoEncontrado     = httpx.NewError(http.StatusNotFound, "rol_tercero_no_encontrado", "el rol de tercero no existe")
	ErrCodigoDuplicado     = httpx.NewError(http.StatusConflict, "codigo_duplicado", "ya existe un rol de tercero activo con ese código")
	ErrNombreDuplicado     = httpx.NewError(http.StatusConflict, "nombre_duplicado", "ya existe un rol de tercero activo con ese nombre")
	ErrRolSistemaProtegido = httpx.NewError(http.StatusConflict, "rol_sistema_protegido", "los roles base del sistema (Cliente, Proveedor, Empleado) no pueden eliminarse ni cambiar su código")
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CrearRol(ctx context.Context, req CrearRolTerceroRequest) (*RolTerceroResponse, error) {
	activo := true
	if req.Activo != nil {
		activo = *req.Activo
	}

	rol := &RolTercero{
		Codigo:      normalizeCodigo(req.Codigo),
		NombreRol:   strings.TrimSpace(req.NombreRol),
		Descripcion: trimPtr(req.Descripcion),
		Activo:      activo,
		EsSistema:   false,
	}
	if err := s.repo.CreateRol(ctx, rol); err != nil {
		return nil, mapError(err)
	}
	resp := toRolResponse(rol)
	return &resp, nil
}

func (s *Service) ObtenerRol(ctx context.Context, id int32) (*RolTerceroResponse, error) {
	rol, err := s.repo.FindRolByID(ctx, id)
	if err != nil {
		return nil, mapError(err)
	}
	resp := toRolResponse(rol)
	return &resp, nil
}

func (s *Service) ListarRoles(ctx context.Context, q ListarRolesTerceroQuery) (*ListaRolesTerceroResponse, error) {
	items, err := s.repo.ListRoles(ctx, q)
	if err != nil {
		return nil, mapError(err)
	}
	data := make([]RolTerceroResponse, 0, len(items))
	for i := range items {
		data = append(data, toRolResponse(&items[i]))
	}
	return &ListaRolesTerceroResponse{Data: data}, nil
}

func (s *Service) ActualizarRol(ctx context.Context, id int32, req ActualizarRolTerceroRequest) (*RolTerceroResponse, error) {
	actual, err := s.repo.FindRolByID(ctx, id)
	if err != nil {
		return nil, mapError(err)
	}

	cambios := map[string]any{}
	if req.Codigo != nil {
		nuevoCodigo := normalizeCodigo(*req.Codigo)
		if actual.EsSistema && nuevoCodigo != actual.Codigo {
			return nil, ErrRolSistemaProtegido
		}
		cambios["codigo"] = nuevoCodigo
	}
	if req.NombreRol != nil {
		cambios["nombre_rol"] = strings.TrimSpace(*req.NombreRol)
	}
	if req.Descripcion != nil {
		cambios["descripcion"] = trimPtr(req.Descripcion)
	}
	if req.Activo != nil {
		cambios["activo"] = *req.Activo
	}

	if err := s.repo.UpdateRol(ctx, id, cambios); err != nil {
		return nil, mapError(err)
	}

	actualizado, err := s.repo.FindRolByID(ctx, id)
	if err != nil {
		return nil, mapError(err)
	}
	resp := toRolResponse(actualizado)
	return &resp, nil
}

func (s *Service) EliminarRol(ctx context.Context, id int32) error {
	actual, err := s.repo.FindRolByID(ctx, id)
	if err != nil {
		return mapError(err)
	}
	if actual.EsSistema {
		return ErrRolSistemaProtegido
	}
	return mapError(s.repo.SoftDeleteRol(ctx, id))
}

func normalizeCodigo(s string) string {
	return strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(s), " ", "_"))
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
	if errors.Is(err, errRolNotFound) {
		return ErrRolNoEncontrado
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		switch pgErr.ConstraintName {
		case "ux_roles_tercero_codigo":
			return ErrCodigoDuplicado
		case "ux_roles_tercero_nombre":
			return ErrNombreDuplicado
		default:
			return ErrNombreDuplicado
		}
	}
	return err
}
