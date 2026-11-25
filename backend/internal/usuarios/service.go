package usuarios

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"

	"tesserak/backend/internal/httpx"
)

const (
	defaultPageSize = 20
	maxPageSize     = 100
)

var (
	ErrUsuarioNoEncontrado  = httpx.NewError(http.StatusNotFound, "usuario_no_encontrado", "el usuario no existe")
	ErrSucursalNoEncontrada = httpx.NewError(http.StatusUnprocessableEntity, "sucursal_no_encontrada", "la sucursal no existe")
	ErrCambioEmpresa        = httpx.NewError(http.StatusUnprocessableEntity, "sucursal_otra_empresa", "la nueva sucursal pertenece a otra empresa")
	ErrEmailDuplicado       = httpx.NewError(http.StatusConflict, "email_duplicado", "ya existe un usuario con ese email en la empresa")
	ErrUsernameDuplicado    = httpx.NewError(http.StatusConflict, "nombre_usuario_duplicado", "ya existe un usuario con ese nombre de usuario en la empresa")
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Crear(ctx context.Context, req CrearUsuarioRequest) (*UsuarioResponse, error) {
	empresaID, err := s.repo.EmpresaDeSucursal(ctx, req.SucursalID)
	if err != nil {
		return nil, mapError(err)
	}

	hash, err := hashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	activo := true
	if req.Activo != nil {
		activo = *req.Activo
	}

	key, err := newUUIDv4()
	if err != nil {
		return nil, err
	}

	u := &Usuario{
		Key:           key,
		EmpresaID:     empresaID,
		SucursalID:    req.SucursalID,
		NombreUsuario: strings.TrimSpace(req.NombreUsuario),
		Nombres:       strings.TrimSpace(req.Nombres),
		Apellidos:     strings.TrimSpace(req.Apellidos),
		Email:         normalizarEmail(req.Email),
		PasswordHash:  hash,
		Activo:        activo,
	}
	if err := s.repo.Create(ctx, u); err != nil {
		return nil, mapError(err)
	}
	resp := toResponse(u)
	return &resp, nil
}

func (s *Service) Obtener(ctx context.Context, id int64) (*UsuarioResponse, error) {
	u, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, mapError(err)
	}
	resp := toResponse(u)
	return &resp, nil
}

func (s *Service) Listar(ctx context.Context, q ListarUsuariosQuery) (*ListaUsuariosResponse, error) {
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

	data := make([]UsuarioResponse, 0, len(items))
	for i := range items {
		data = append(data, toResponse(&items[i]))
	}
	return &ListaUsuariosResponse{Data: data, Page: q.Page, PageSize: q.PageSize, Total: total}, nil
}

func (s *Service) Actualizar(ctx context.Context, id int64, req ActualizarUsuarioRequest) (*UsuarioResponse, error) {
	var result *Usuario

	err := s.repo.WithTx(ctx, func(tx *Repository) error {
		u, err := tx.FindByID(ctx, id)
		if err != nil {
			return err
		}

		cambios := map[string]any{}

		if req.SucursalID != nil && *req.SucursalID != u.SucursalID {
			empresaID, err := tx.EmpresaDeSucursal(ctx, *req.SucursalID)
			if err != nil {
				return err
			}
			// El email es único por empresa: no se permite mover un usuario a otra empresa.
			if empresaID != u.EmpresaID {
				return ErrCambioEmpresa
			}
			cambios["sucursal_id"] = *req.SucursalID
		}
		if req.NombreUsuario != nil {
			cambios["nombre_usuario"] = strings.TrimSpace(*req.NombreUsuario)
		}
		if req.Nombres != nil {
			cambios["nombres"] = strings.TrimSpace(*req.Nombres)
		}
		if req.Apellidos != nil {
			cambios["apellidos"] = strings.TrimSpace(*req.Apellidos)
		}
		if req.Email != nil {
			cambios["email"] = normalizarEmail(*req.Email)
		}
		if req.Password != nil {
			hash, err := hashPassword(*req.Password)
			if err != nil {
				return err
			}
			cambios["password_hash"] = hash
		}
		if req.Activo != nil {
			cambios["activo"] = *req.Activo
		}

		if err := tx.Update(ctx, u, cambios); err != nil {
			return err
		}

		// Releer para devolver el estado final (incluye updated_at).
		result, err = tx.FindByID(ctx, id)
		return err
	})
	if err != nil {
		return nil, mapError(err)
	}

	resp := toResponse(result)
	return &resp, nil
}

func (s *Service) Eliminar(ctx context.Context, id int64) error {
	return mapError(s.repo.SoftDelete(ctx, id))
}

// ---------------------------------------------------------------------

func normalizarEmail(e string) string {
	return strings.ToLower(strings.TrimSpace(e))
}

func hashPassword(p string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(p), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// newUUIDv4 genera un UUID v4 estándar (RFC 4122) usando crypto/rand.
func newUUIDv4() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	b[6] = (b[6] & 0x0f) | 0x40 // versión 4
	b[8] = (b[8] & 0x3f) | 0x80 // variante RFC 4122
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16]), nil
}

// mapError convierte errores de repositorio/PostgreSQL en errores HTTP de dominio.
func mapError(err error) error {
	if err == nil {
		return nil
	}

	var appErr *httpx.AppError
	if errors.As(err, &appErr) {
		return err
	}

	switch {
	case errors.Is(err, errNotFound):
		return ErrUsuarioNoEncontrado
	case errors.Is(err, errSucursalNotFound):
		return ErrSucursalNoEncontrada
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505": // unique_violation
			switch pgErr.ConstraintName {
			case "ux_usuarios_empresa_email":
				return ErrEmailDuplicado
			case "ux_usuarios_empresa_nombre_usuario":
				return ErrUsernameDuplicado
			}
			return httpx.NewError(http.StatusConflict, "duplicado", "el registro ya existe")
		case "23503": // foreign_key_violation
			return httpx.NewError(http.StatusUnprocessableEntity, "referencia_invalida", "la sucursal o empresa indicada no es válida")
		}
	}
	return err
}
