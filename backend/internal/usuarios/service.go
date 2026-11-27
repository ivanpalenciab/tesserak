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

	"tesserak/backend/internal/auth"
	"tesserak/backend/internal/httpx"
)

const (
	defaultPageSize = 20
	maxPageSize     = 100
	rolUsuarioStd   = "USUARIO"
)

var (
	ErrUsuarioNoEncontrado  = httpx.NewError(http.StatusNotFound, "usuario_no_encontrado", "el usuario no existe")
	ErrEmpresaRequerida     = httpx.NewError(http.StatusBadRequest, "empresa_requerida", "debe indicar empresa_id o al menos una sucursal")
	ErrSucursalRequerida    = httpx.NewError(http.StatusBadRequest, "sucursal_requerida", "los usuarios no administradores deben tener al menos una sucursal asignada")
	ErrSucursalNoEncontrada = httpx.NewError(http.StatusUnprocessableEntity, "sucursal_no_encontrada", "una o más sucursales indicadas no existen")
	ErrRolNoEncontrado      = httpx.NewError(http.StatusUnprocessableEntity, "rol_no_encontrado", "el rol de sistema indicado no existe o está inactivo")
	ErrCambioEmpresa        = httpx.NewError(http.StatusUnprocessableEntity, "sucursal_otra_empresa", "todas las sucursales deben pertenecer a la misma empresa")
	ErrEmailDuplicado          = httpx.NewError(http.StatusConflict, "email_duplicado", "ya existe un usuario con ese email en la empresa")
	ErrUsernameDuplicado       = httpx.NewError(http.StatusConflict, "nombre_usuario_duplicado", "ya existe un usuario con ese nombre de usuario en la empresa")
	ErrPasswordActualRequerida = httpx.NewError(http.StatusBadRequest, "password_actual_requerida", "debe ingresar su contraseña actual para establecer una nueva")
	ErrPasswordActualIncorrecta = httpx.NewError(http.StatusUnprocessableEntity, "password_actual_incorrecta", "la contraseña actual es incorrecta")
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// ListarRoles devuelve el catálogo de roles de sistema activos asignables por caller.
// Si caller no es ADMINISTRADOR global (ej. ADMINISTRADOR_SUCURSAL), excluye ADMINISTRADOR.
func (s *Service) ListarRoles(ctx context.Context, caller *auth.Claims) ([]RolSistemaResponse, error) {
	roles, err := s.repo.ListarRolesSistema(ctx)
	if err != nil {
		return nil, err
	}
	soloScopeSucursal := caller != nil && !caller.IsAdmin()
	out := make([]RolSistemaResponse, 0, len(roles))
	for _, r := range roles {
		if soloScopeSucursal && strings.EqualFold(r.Codigo, auth.RolAdministrador) {
			continue
		}
		out = append(out, toRolResponse(r))
	}
	return out, nil
}

// ListarSucursalesParaCaller devuelve las sucursales activas accesibles por caller:
// - ADMINISTRADOR: todas las sucursales activas de su empresa.
// - Otros roles (incl. ADMINISTRADOR_SUCURSAL): únicamente sus sucursales asignadas.
func (s *Service) ListarSucursalesParaCaller(ctx context.Context, caller *auth.Claims) ([]SucursalResumen, error) {
	if caller == nil {
		return nil, auth.ErrNoAutenticado
	}
	if caller.IsAdmin() {
		return s.repo.ListarSucursalesEmpresa(ctx, caller.EmpresaID)
	}
	ids, err := s.repo.SucursalesPorUsuario(ctx, caller.Sub)
	if err != nil {
		return nil, err
	}
	return s.repo.ListarSucursalesPorIDs(ctx, caller.EmpresaID, ids)
}

// Crear crea un usuario y le asigna rol de sistema + sucursales.
//   - Si la empresa aún no tiene usuarios, se permite sin token y se asigna ADMINISTRADOR automáticamente.
//   - Si la empresa ya tiene usuarios, exige que caller sea un ADMINISTRADOR autenticado de esa empresa.
//   - Si el rol final es ADMINISTRADOR, tiene acceso a todas las sucursales de la empresa.
//   - Si no es ADMINISTRADOR, requiere al menos una sucursal específica en sucursal_ids / sucursal_id.
func (s *Service) Crear(ctx context.Context, caller *auth.Claims, req CrearUsuarioRequest) (*UsuarioResponse, error) {
	sucursalIDs := normalizarSucursalIDs(req.SucursalID, req.SucursalIDs)

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

	var (
		u            *Usuario
		rolesFinales []RolSistema
		sidsFinales  []int32
	)

	err = s.repo.WithTx(ctx, func(tx *Repository) error {
		// 1. Determinar empresa_id y validar las sucursales enviadas (si hay)
		var empresaID int32
		if caller != nil {
			empresaID = caller.EmpresaID
		} else if req.EmpresaID != nil {
			empresaID = *req.EmpresaID
		}

		if len(sucursalIDs) > 0 {
			empresasMap, err := tx.EmpresasDeSucursales(ctx, sucursalIDs)
			if err != nil {
				return err
			}
			if len(empresasMap) != len(sucursalIDs) {
				return errSucursalNotFound
			}
			if empresaID == 0 {
				empresaID = empresasMap[sucursalIDs[0]]
			}
			for _, sid := range sucursalIDs {
				if empresasMap[sid] != empresaID {
					return ErrCambioEmpresa
				}
			}
		}
		if empresaID == 0 {
			return ErrEmpresaRequerida
		}

		// 2. Verificar si es el primer usuario de la empresa
		existentes, err := tx.ContarUsuariosEmpresa(ctx, empresaID)
		if err != nil {
			return err
		}
		esPrimerUsuario := existentes == 0

		if !esPrimerUsuario {
			if caller == nil {
				return auth.ErrNoAutenticado
			}
			if !caller.CanManageUsers() || caller.EmpresaID != empresaID {
				return auth.ErrSoloAdminOSucursal
			}
		}

		// 3. Resolver roles de sistema (soporta múltiples roles en rol_ids y/o rol_id)
		var rolesAsignar []RolSistema
		if esPrimerUsuario {
			rolAdmin, err := tx.BuscarRolPorCodigo(ctx, auth.RolAdministrador)
			if err != nil {
				return err
			}
			rolesAsignar = []RolSistema{*rolAdmin}
		} else {
			rolIDsReq := normalizarSucursalIDs(req.RolID, req.RolIDs)
			if len(rolIDsReq) > 0 {
				rolesAsignar, err = tx.BuscarRolesPorIDs(ctx, rolIDsReq)
				if err != nil {
					return err
				}
			} else {
				rolDefault, err := tx.BuscarRolPorCodigo(ctx, rolUsuarioStd)
				if err != nil {
					return err
				}
				rolesAsignar = []RolSistema{*rolDefault}
			}
		}

		esAdminNuevo := tieneRolAdmin(rolesAsignar)
		rolIDsFinales := make([]int32, 0, len(rolesAsignar))
		for _, rItem := range rolesAsignar {
			rolIDsFinales = append(rolIDsFinales, rItem.ID)
		}

		// Un ADMINISTRADOR_SUCURSAL no puede asignar el rol global ADMINISTRADOR
		// y solo puede asignar sucursales dentro de su propio alcance.
		if !esPrimerUsuario && caller != nil && !caller.IsAdmin() {
			if esAdminNuevo {
				return auth.ErrSoloAdministrador
			}
			callerSucs, err := tx.SucursalesPorUsuario(ctx, caller.Sub)
			if err != nil {
				return err
			}
			if len(sucursalIDs) == 0 && len(callerSucs) == 1 {
				sucursalIDs = callerSucs
			}
			if len(sucursalIDs) == 0 {
				return ErrSucursalRequerida
			}
			if !todasEnConjunto(sucursalIDs, callerSucs) {
				return auth.ErrSucursalNoAutorizada
			}
		}

		// 4. Validar/resolver sucursales según el rol
		if esAdminNuevo {
			if len(sucursalIDs) == 0 {
				sucursalesEmp, err := tx.ListarSucursalesEmpresa(ctx, empresaID)
				if err != nil {
					return err
				}
				for _, sItem := range sucursalesEmp {
					sucursalIDs = append(sucursalIDs, sItem.ID)
				}
			}
		} else if len(sucursalIDs) == 0 {
			return ErrSucursalRequerida
		}

		// 5. Insertar usuario, sucursales y roles
		u = &Usuario{
			Key:           key,
			EmpresaID:     empresaID,
			NombreUsuario: strings.TrimSpace(req.NombreUsuario),
			Nombres:       strings.TrimSpace(req.Nombres),
			Apellidos:     strings.TrimSpace(req.Apellidos),
			Email:         normalizarEmail(req.Email),
			PasswordHash:  hash,
			Activo:        activo,
		}
		if err := tx.Create(ctx, u); err != nil {
			return err
		}
		if err := tx.ReemplazarSucursales(ctx, u.ID, empresaID, sucursalIDs); err != nil {
			return err
		}
		if err := tx.ReemplazarRoles(ctx, u.ID, rolIDsFinales); err != nil {
			return err
		}

		rolesFinales = rolesAsignar
		sidsFinales = sucursalIDs
		return nil
	})
	if err != nil {
		return nil, mapError(err)
	}

	resp := toResponse(u, sidsFinales, rolesFinales)
	return &resp, nil
}

func (s *Service) Obtener(ctx context.Context, caller *auth.Claims, id int64) (*UsuarioResponse, error) {
	u, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, mapError(err)
	}
	if caller != nil && u.EmpresaID != caller.EmpresaID {
		return nil, ErrUsuarioNoEncontrado
	}
	roles, err := s.repo.RolesPorUsuario(ctx, u.ID)
	if err != nil {
		return nil, err
	}
	sucursalIDs, err := s.resolverSucursalesVisibles(ctx, u.EmpresaID, u.ID, roles)
	if err != nil {
		return nil, err
	}
	resp := toResponse(u, sucursalIDs, roles)
	return &resp, nil
}

func (s *Service) Listar(ctx context.Context, caller *auth.Claims, q ListarUsuariosQuery) (*ListaUsuariosResponse, error) {
	if caller != nil {
		q.EmpresaID = caller.EmpresaID
		if !caller.IsAdmin() {
			callerSucs, err := s.repo.SucursalesPorUsuario(ctx, caller.Sub)
			if err != nil {
				return nil, err
			}
			if len(callerSucs) == 0 {
				return &ListaUsuariosResponse{Data: []UsuarioResponse{}, Page: 1, PageSize: defaultPageSize, Total: 0}, nil
			}
			q.SucursalIDsScope = callerSucs
		}
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

	items, total, err := s.repo.List(ctx, q)
	if err != nil {
		return nil, err
	}

	userIDs := make([]int64, 0, len(items))
	for i := range items {
		userIDs = append(userIDs, items[i].ID)
	}
	sucursalesMap, err := s.repo.SucursalesPorUsuarios(ctx, userIDs)
	if err != nil {
		return nil, err
	}
	rolesMap, err := s.repo.RolesPorUsuarios(ctx, userIDs)
	if err != nil {
		return nil, err
	}

	// Cache de todas las sucursales activas por empresa para usuarios administradores
	todasSucursalesCache := make(map[int32][]int32)

	data := make([]UsuarioResponse, 0, len(items))
	for i := range items {
		u := &items[i]
		roles := rolesMap[u.ID]
		sids := sucursalesMap[u.ID]
		if tieneRolAdmin(roles) {
			cached, ok := todasSucursalesCache[u.EmpresaID]
			if !ok {
				sucs, err := s.repo.ListarSucursalesEmpresa(ctx, u.EmpresaID)
				if err != nil {
					return nil, err
				}
				cached = make([]int32, 0, len(sucs))
				for _, sc := range sucs {
					cached = append(cached, sc.ID)
				}
				todasSucursalesCache[u.EmpresaID] = cached
			}
			sids = combinarPrincipalYResto(sids, cached)
		}
		data = append(data, toResponse(u, sids, roles))
	}
	return &ListaUsuariosResponse{Data: data, Page: q.Page, PageSize: q.PageSize, Total: total}, nil
}

func (s *Service) Actualizar(ctx context.Context, caller *auth.Claims, id int64, req ActualizarUsuarioRequest) (*UsuarioResponse, error) {
	if caller != nil && !caller.CanManageUsers() {
		return nil, auth.ErrSoloAdminOSucursal
	}

	var (
		result      *Usuario
		sucursalIDs []int32
		roles       []RolSistema
	)

	err := s.repo.WithTx(ctx, func(tx *Repository) error {
		u, err := tx.FindByID(ctx, id)
		if err != nil {
			return err
		}
		if caller != nil && u.EmpresaID != caller.EmpresaID {
			return errNotFound
		}

		rolesActuales, err := tx.RolesPorUsuario(ctx, u.ID)
		if err != nil {
			return err
		}
		esAdminFinal := tieneRolAdmin(rolesActuales)

		var callerSucs []int32
		esAdminSucursalCaller := caller != nil && !caller.IsAdmin()
		if esAdminSucursalCaller {
			// Un ADMINISTRADOR_SUCURSAL no puede modificar a un ADMINISTRADOR global
			if esAdminFinal {
				return auth.ErrSoloAdministrador
			}
			callerSucs, err = tx.SucursalesPorUsuario(ctx, caller.Sub)
			if err != nil {
				return err
			}
			targetSucs, err := tx.SucursalesPorUsuario(ctx, u.ID)
			if err != nil {
				return err
			}
			if !compartenAlguna(targetSucs, callerSucs) {
				return auth.ErrSucursalNoAutorizada
			}
		}

		if req.RolIDs != nil || req.RolID != nil {
			var rolIDsReq []int32
			if req.RolIDs != nil {
				rolIDsReq = normalizarSucursalIDs(req.RolID, *req.RolIDs)
			} else {
				rolIDsReq = normalizarSucursalIDs(req.RolID, nil)
			}
			nuevosRoles, err := tx.BuscarRolesPorIDs(ctx, rolIDsReq)
			if err != nil {
				return err
			}
			if esAdminSucursalCaller && tieneRolAdmin(nuevosRoles) {
				return auth.ErrSoloAdministrador
			}
			if err := tx.ReemplazarRoles(ctx, u.ID, rolIDsReq); err != nil {
				return err
			}
			esAdminFinal = tieneRolAdmin(nuevosRoles)
		}

		if req.SucursalIDs != nil || req.SucursalID != nil {
			var nuevas []int32
			if req.SucursalIDs != nil {
				nuevas = normalizarSucursalIDs(req.SucursalID, *req.SucursalIDs)
			} else {
				nuevas = normalizarSucursalIDs(req.SucursalID, nil)
			}
			if !esAdminFinal && len(nuevas) == 0 {
				return ErrSucursalRequerida
			}
			if esAdminSucursalCaller && !todasEnConjunto(nuevas, callerSucs) {
				return auth.ErrSucursalNoAutorizada
			}
			if len(nuevas) > 0 {
				empresasMap, err := tx.EmpresasDeSucursales(ctx, nuevas)
				if err != nil {
					return err
				}
				if len(empresasMap) != len(nuevas) {
					return errSucursalNotFound
				}
				for _, sid := range nuevas {
					if empresasMap[sid] != u.EmpresaID {
						return ErrCambioEmpresa
					}
				}
			}
			if err := tx.ReemplazarSucursales(ctx, u.ID, u.EmpresaID, nuevas); err != nil {
				return err
			}
		} else if !esAdminFinal {
			// Si cambió de admin a no-admin sin enviar sucursales, validar que ya tenga al menos 1
			actuales, err := tx.SucursalesPorUsuario(ctx, u.ID)
			if err != nil {
				return err
			}
			if len(actuales) == 0 {
				return ErrSucursalRequerida
			}
		}

		cambios := map[string]any{}
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

		result, err = tx.FindByID(ctx, id)
		if err != nil {
			return err
		}
		roles, err = tx.RolesPorUsuario(ctx, id)
		if err != nil {
			return err
		}
		sucursalIDs, err = tx.SucursalesPorUsuario(ctx, id)
		if err != nil {
			return err
		}
		if tieneRolAdmin(roles) {
			sucs, err := tx.ListarSucursalesEmpresa(ctx, result.EmpresaID)
			if err != nil {
				return err
			}
			todas := make([]int32, 0, len(sucs))
			for _, sc := range sucs {
				todas = append(todas, sc.ID)
			}
			sucursalIDs = combinarPrincipalYResto(sucursalIDs, todas)
		}
		return nil
	})
	if err != nil {
		return nil, mapError(err)
	}

	resp := toResponse(result, sucursalIDs, roles)
	return &resp, nil
}

func (s *Service) Eliminar(ctx context.Context, caller *auth.Claims, id int64) error {
	if caller != nil && !caller.CanManageUsers() {
		return auth.ErrSoloAdminOSucursal
	}
	if caller != nil {
		u, err := s.repo.FindByID(ctx, id)
		if err != nil {
			return mapError(err)
		}
		if u.EmpresaID != caller.EmpresaID {
			return ErrUsuarioNoEncontrado
		}
		if !caller.IsAdmin() {
			roles, err := s.repo.RolesPorUsuario(ctx, u.ID)
			if err != nil {
				return err
			}
			if tieneRolAdmin(roles) {
				return auth.ErrSoloAdministrador
			}
			callerSucs, err := s.repo.SucursalesPorUsuario(ctx, caller.Sub)
			if err != nil {
				return err
			}
			targetSucs, err := s.repo.SucursalesPorUsuario(ctx, u.ID)
			if err != nil {
				return err
			}
			if !compartenAlguna(targetSucs, callerSucs) {
				return auth.ErrSucursalNoAutorizada
			}
		}
	}
	return mapError(s.repo.SoftDelete(ctx, id))
}

// ActualizarMiPerfil permite al usuario autenticado actualizar sus propios datos personales
// y cambiar su contraseña (validando obligatoriamente la contraseña actual).
func (s *Service) ActualizarMiPerfil(ctx context.Context, caller *auth.Claims, req ActualizarPerfilRequest) (*UsuarioResponse, error) {
	if caller == nil || caller.Sub <= 0 {
		return nil, auth.ErrNoAutenticado
	}

	var (
		result      *Usuario
		sucursalIDs []int32
		roles       []RolSistema
	)

	err := s.repo.WithTx(ctx, func(tx *Repository) error {
		u, err := tx.FindByID(ctx, caller.Sub)
		if err != nil {
			return err
		}
		if u.EmpresaID != caller.EmpresaID {
			return errNotFound
		}

		cambios := map[string]any{}
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

		if req.PasswordNueva != nil {
			if req.PasswordActual == nil || *req.PasswordActual == "" {
				return ErrPasswordActualRequerida
			}
			if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(*req.PasswordActual)); err != nil {
				return ErrPasswordActualIncorrecta
			}
			hash, err := hashPassword(*req.PasswordNueva)
			if err != nil {
				return err
			}
			cambios["password_hash"] = hash
		}

		if err := tx.Update(ctx, u, cambios); err != nil {
			return err
		}

		result, err = tx.FindByID(ctx, caller.Sub)
		if err != nil {
			return err
		}
		roles, err = tx.RolesPorUsuario(ctx, caller.Sub)
		if err != nil {
			return err
		}
		sucursalIDs, err = tx.SucursalesPorUsuario(ctx, caller.Sub)
		if err != nil {
			return err
		}
		if tieneRolAdmin(roles) {
			sucs, err := tx.ListarSucursalesEmpresa(ctx, result.EmpresaID)
			if err != nil {
				return err
			}
			todas := make([]int32, 0, len(sucs))
			for _, sc := range sucs {
				todas = append(todas, sc.ID)
			}
			sucursalIDs = combinarPrincipalYResto(sucursalIDs, todas)
		}
		return nil
	})
	if err != nil {
		return nil, mapError(err)
	}

	resp := toResponse(result, sucursalIDs, roles)
	return &resp, nil
}

// RegenerarKey asigna un nuevo UUID v4 en usuarios.key y devuelve el usuario actualizado.
func (s *Service) RegenerarKey(ctx context.Context, caller *auth.Claims, id int64) (*UsuarioResponse, error) {
	nuevaKey, err := newUUIDv4()
	if err != nil {
		return nil, err
	}

	var (
		result      *Usuario
		sucursalIDs []int32
		roles       []RolSistema
	)
	err = s.repo.WithTx(ctx, func(tx *Repository) error {
		u, err := tx.FindByID(ctx, id)
		if err != nil {
			return err
		}
		if caller != nil && u.EmpresaID != caller.EmpresaID {
			return errNotFound
		}
		if caller != nil && caller.Sub != u.ID && !caller.IsAdmin() {
			return auth.ErrSoloAdministrador
		}
		if err := tx.Update(ctx, u, map[string]any{"key": nuevaKey}); err != nil {
			return err
		}
		result, err = tx.FindByID(ctx, id)
		if err != nil {
			return err
		}
		roles, err = tx.RolesPorUsuario(ctx, id)
		if err != nil {
			return err
		}
		sucursalIDs, err = tx.SucursalesPorUsuario(ctx, id)
		return err
	})
	if err != nil {
		return nil, mapError(err)
	}
	resp := toResponse(result, sucursalIDs, roles)
	return &resp, nil
}

// ---------------------------------------------------------------------

func (s *Service) resolverSucursalesVisibles(ctx context.Context, empresaID int32, usuarioID int64, roles []RolSistema) ([]int32, error) {
	asignadas, err := s.repo.SucursalesPorUsuario(ctx, usuarioID)
	if err != nil {
		return nil, err
	}
	if !tieneRolAdmin(roles) {
		return asignadas, nil
	}
	sucs, err := s.repo.ListarSucursalesEmpresa(ctx, empresaID)
	if err != nil {
		return nil, err
	}
	todas := make([]int32, 0, len(sucs))
	for _, sc := range sucs {
		todas = append(todas, sc.ID)
	}
	return combinarPrincipalYResto(asignadas, todas), nil
}

func tieneRolAdmin(roles []RolSistema) bool {
	for _, r := range roles {
		if strings.EqualFold(r.Codigo, auth.RolAdministrador) {
			return true
		}
	}
	return false
}

func todasEnConjunto(sub, conjunto []int32) bool {
	set := make(map[int32]struct{}, len(conjunto))
	for _, id := range conjunto {
		set[id] = struct{}{}
	}
	for _, id := range sub {
		if _, ok := set[id]; !ok {
			return false
		}
	}
	return true
}

func compartenAlguna(a, b []int32) bool {
	set := make(map[int32]struct{}, len(b))
	for _, id := range b {
		set[id] = struct{}{}
	}
	for _, id := range a {
		if _, ok := set[id]; ok {
			return true
		}
	}
	return false
}

func combinarPrincipalYResto(asignadas, todas []int32) []int32 {
	seen := make(map[int32]struct{}, len(todas))
	out := make([]int32, 0, len(todas))
	if len(asignadas) > 0 {
		for _, id := range todas {
			if id == asignadas[0] {
				seen[id] = struct{}{}
				out = append(out, id)
				break
			}
		}
	}
	for _, id := range todas {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}

// normalizarSucursalIDs combina sucursal_id y sucursal_ids eliminando duplicados
// y conservando el orden (la primera queda como sucursal principal).
func normalizarSucursalIDs(principal *int32, lista []int32) []int32 {
	seen := make(map[int32]struct{}, len(lista)+1)
	out := make([]int32, 0, len(lista)+1)
	if principal != nil && *principal > 0 {
		seen[*principal] = struct{}{}
		out = append(out, *principal)
	}
	for _, id := range lista {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}

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
	case errors.Is(err, errRolNotFound):
		return ErrRolNoEncontrado
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
			return httpx.NewError(http.StatusUnprocessableEntity, "referencia_invalida", "la sucursal, rol o empresa indicada no es válida")
		}
	}
	return err
}
