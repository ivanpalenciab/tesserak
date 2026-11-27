package lealtad

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"tesserak/backend/internal/auth"
	"tesserak/backend/internal/httpx"
)

var (
	ErrProgramaNoEncontrado   = httpx.NewError(http.StatusNotFound, "programa_no_encontrado", "el plan de lealtad no existe")
	ErrAdhesionNoEncontrada   = httpx.NewError(http.StatusNotFound, "adhesion_no_encontrada", "la empresa no se encuentra adherida a este plan de lealtad")
	ErrCodigoDuplicado        = httpx.NewError(http.StatusConflict, "codigo_duplicado", "ya existe un plan de lealtad activo con ese código")
	ErrPlanSistemaProtegido   = httpx.NewError(http.StatusConflict, "plan_sistema_protegido", "el plan de lealtad por defecto del sistema no puede eliminarse ni cambiar de código")
	ErrSinPermisoPlanGlobal   = httpx.NewError(http.StatusForbidden, "solo_admin_plan_global", "solo el administrador designado del plan global (ipalenciab) puede administrar este plan")
	ErrSinPermisoPlanEmpresa  = httpx.NewError(http.StatusForbidden, "sin_permiso_plan", "solo un administrador de la empresa propietaria puede modificar o eliminar este plan")
	ErrMonedaNoEncontrada     = httpx.NewError(http.StatusBadRequest, "moneda_invalida", "la moneda de referencia indicada no existe")
	ErrVigenciaInvalida       = httpx.NewError(http.StatusBadRequest, "vigencia_invalida", "la fecha de fin debe ser igual o posterior a la fecha de inicio")
	ErrPlanNoDisponible       = httpx.NewError(http.StatusConflict, "plan_no_disponible", "solo es posible adherirse a planes de lealtad en estado ACTIVO")
	ErrAdhesionNoPermitida    = httpx.NewError(http.StatusForbidden, "adhesion_no_permitida", "este plan es de alcance LOCAL y solo permite la adhesión de su empresa propietaria")
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListarMonedas(ctx context.Context) ([]MonedaResponse, error) {
	items, err := s.repo.ListMonedas(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]MonedaResponse, 0, len(items))
	for _, m := range items {
		out = append(out, MonedaResponse{
			ID:           m.ID,
			CodigoISO:    m.CodigoISO,
			NombreMoneda: m.NombreMoneda,
			Simbolo:      m.Simbolo,
		})
	}
	return out, nil
}

func (s *Service) EsAdminGlobal(ctx context.Context, claims *auth.Claims) (bool, error) {
	return s.isPlatformLoyaltyAdmin(ctx, claims)
}

func (s *Service) isPlatformLoyaltyAdmin(ctx context.Context, claims *auth.Claims) (bool, error) {
	if claims == nil || !claims.IsAdmin() {
		return false, nil
	}
	username, err := s.repo.GetUsuarioUsername(ctx, claims.Sub)
	if err != nil {
		return false, err
	}
	return strings.EqualFold(strings.TrimSpace(username), UsuarioAdminGlobal), nil
}

func (s *Service) canManagePrograma(claims *auth.Claims, isGlobalAdmin bool, p Programa) bool {
	if claims == nil || !claims.IsAdmin() {
		return false
	}
	if p.Alcance == AlcanceGlobal {
		if p.UsuarioAdministradorID != nil && *p.UsuarioAdministradorID == claims.Sub {
			return true
		}
		return isGlobalAdmin
	}
	return p.EmpresaPropietariaID != nil && *p.EmpresaPropietariaID == claims.EmpresaID
}

func (s *Service) canAdherir(claims *auth.Claims, p Programa) bool {
	if claims == nil || !claims.IsAdmin() {
		return false
	}
	if p.Estado != EstadoActivo {
		return false
	}
	if p.Alcance == AlcanceLocal {
		return p.EmpresaPropietariaID != nil && *p.EmpresaPropietariaID == claims.EmpresaID
	}
	return true
}

func (s *Service) ListarProgramas(ctx context.Context, claims *auth.Claims, filter ProgramaFilter) ([]ProgramaResponse, error) {
	programas, err := s.repo.ListProgramasForEmpresa(ctx, claims.EmpresaID, filter)
	if err != nil {
		return nil, err
	}
	isGlobalAdmin, err := s.isPlatformLoyaltyAdmin(ctx, claims)
	if err != nil {
		return nil, err
	}

	ids := make([]int32, 0, len(programas))
	for _, p := range programas {
		ids = append(ids, p.ID)
	}
	adhesionesMap, err := s.repo.GetAdhesionesMapForEmpresa(ctx, claims.EmpresaID, ids)
	if err != nil {
		return nil, err
	}
	countsMap, err := s.repo.GetActiveAdhesionCounts(ctx, ids)
	if err != nil {
		return nil, err
	}

	out := make([]ProgramaResponse, 0, len(programas))
	for _, p := range programas {
		var adh *ProgramaEmpresa
		if row, ok := adhesionesMap[p.ID]; ok {
			copyRow := row
			adh = &copyRow
		}
		canManage := s.canManagePrograma(claims, isGlobalAdmin, p)
		out = append(out, toProgramaResponse(
			p,
			adh,
			countsMap[p.ID],
			canManage,
			canManage && !p.EsSistema,
			canManage,
			s.canAdherir(claims, p),
		))
	}
	return out, nil
}

func (s *Service) ObtenerPrograma(ctx context.Context, claims *auth.Claims, id int32) (*ProgramaResponse, error) {
	p, err := s.repo.FindProgramaByID(ctx, id)
	if err != nil {
		if errors.Is(err, errProgramaNotFound) {
			return nil, ErrProgramaNoEncontrado
		}
		return nil, err
	}
	isGlobalAdmin, err := s.isPlatformLoyaltyAdmin(ctx, claims)
	if err != nil {
		return nil, err
	}
	if p.Alcance == AlcanceLocal && (p.EmpresaPropietariaID == nil || *p.EmpresaPropietariaID != claims.EmpresaID) {
		return nil, ErrProgramaNoEncontrado
	}

	adh, err := s.repo.FindAdhesion(ctx, p.ID, claims.EmpresaID)
	if err != nil && !errors.Is(err, errAdhesionNotFound) {
		return nil, err
	}
	countsMap, err := s.repo.GetActiveAdhesionCounts(ctx, []int32{p.ID})
	if err != nil {
		return nil, err
	}
	canManage := s.canManagePrograma(claims, isGlobalAdmin, *p)
	resp := toProgramaResponse(
		*p,
		adh,
		countsMap[p.ID],
		canManage,
		canManage && !p.EsSistema,
		canManage,
		s.canAdherir(claims, *p),
	)
	return &resp, nil
}

func (s *Service) CrearPrograma(ctx context.Context, claims *auth.Claims, req CrearProgramaRequest) (*ProgramaResponse, error) {
	codigo := strings.ToUpper(strings.TrimSpace(req.Codigo))
	nombre := strings.TrimSpace(req.Nombre)
	if codigo == "" || nombre == "" {
		return nil, httpx.NewError(http.StatusBadRequest, "validacion", "código y nombre son obligatorios")
	}

	alcance := strings.ToUpper(strings.TrimSpace(req.Alcance))
	if alcance == "" {
		alcance = AlcanceLocal
	}

	isGlobalAdmin, err := s.isPlatformLoyaltyAdmin(ctx, claims)
	if err != nil {
		return nil, err
	}
	if alcance == AlcanceGlobal && !isGlobalAdmin {
		return nil, ErrSinPermisoPlanGlobal
	}

	okMoneda, err := s.repo.ExistsMoneda(ctx, req.MonedaReferenciaID)
	if err != nil {
		return nil, err
	}
	if !okMoneda {
		return nil, ErrMonedaNoEncontrada
	}

	estado := strings.ToUpper(strings.TrimSpace(req.Estado))
	if estado == "" {
		estado = EstadoActivo
	}

	fechaInicio := time.Now().UTC()
	if req.FechaInicio != nil {
		fechaInicio = req.FechaInicio.UTC()
	}
	if req.FechaFin != nil && req.FechaFin.UTC().Before(fechaInicio) {
		return nil, ErrVigenciaInvalida
	}

	var desc *string
	if req.Descripcion != nil {
		trimmed := strings.TrimSpace(*req.Descripcion)
		if trimmed != "" {
			desc = &trimmed
		}
	}

	var empresaPropietariaID *int32
	if alcance != AlcanceGlobal {
		empID := claims.EmpresaID
		empresaPropietariaID = &empID
	}
	adminID := claims.Sub

	p := &Programa{
		Codigo:                 codigo,
		Nombre:                 nombre,
		Descripcion:            desc,
		Alcance:                alcance,
		EmpresaPropietariaID:   empresaPropietariaID,
		UsuarioAdministradorID: &adminID,
		EsSistema:              false,
		MonedaReferenciaID:     req.MonedaReferenciaID,
		ValorPuntoRedencion:    req.ValorPuntoRedencion,
		DiasVencimiento:        req.DiasVencimiento,
		Estado:                 estado,
		FechaInicio:            fechaInicio,
		FechaFin:               req.FechaFin,
	}

	if err := s.repo.CreatePrograma(ctx, p, alcance != AlcanceGlobal); err != nil {
		if errors.Is(err, errCodigoDuplicado) {
			return nil, ErrCodigoDuplicado
		}
		return nil, err
	}
	return s.ObtenerPrograma(ctx, claims, p.ID)
}

func (s *Service) ActualizarPrograma(ctx context.Context, claims *auth.Claims, id int32, req ActualizarProgramaRequest) (*ProgramaResponse, error) {
	p, err := s.repo.FindProgramaByID(ctx, id)
	if err != nil {
		if errors.Is(err, errProgramaNotFound) {
			return nil, ErrProgramaNoEncontrado
		}
		return nil, err
	}

	isGlobalAdmin, err := s.isPlatformLoyaltyAdmin(ctx, claims)
	if err != nil {
		return nil, err
	}
	if !s.canManagePrograma(claims, isGlobalAdmin, *p) {
		if p.Alcance == AlcanceGlobal {
			return nil, ErrSinPermisoPlanGlobal
		}
		return nil, ErrSinPermisoPlanEmpresa
	}

	updates := map[string]any{
		"updated_at": time.Now().UTC(),
	}

	if req.Codigo != nil {
		codigo := strings.ToUpper(strings.TrimSpace(*req.Codigo))
		if p.EsSistema && !strings.EqualFold(codigo, p.Codigo) {
			return nil, ErrPlanSistemaProtegido
		}
		if codigo != "" {
			updates["codigo"] = codigo
		}
	}
	if req.Nombre != nil {
		nombre := strings.TrimSpace(*req.Nombre)
		if nombre != "" {
			updates["nombre"] = nombre
		}
	}
	if req.Descripcion != nil {
		trimmed := strings.TrimSpace(*req.Descripcion)
		if trimmed == "" {
			updates["descripcion"] = nil
		} else {
			updates["descripcion"] = trimmed
		}
	}
	if req.MonedaReferenciaID != nil {
		okMoneda, err := s.repo.ExistsMoneda(ctx, *req.MonedaReferenciaID)
		if err != nil {
			return nil, err
		}
		if !okMoneda {
			return nil, ErrMonedaNoEncontrada
		}
		updates["moneda_referencia_id"] = *req.MonedaReferenciaID
	}
	if req.ValorPuntoRedencion != nil {
		updates["valor_punto_redencion"] = *req.ValorPuntoRedencion
	}
	if req.LimpiarDiasVencimiento {
		updates["dias_vencimiento"] = nil
	} else if req.DiasVencimiento != nil {
		updates["dias_vencimiento"] = *req.DiasVencimiento
	}
	if req.Estado != nil {
		updates["estado"] = strings.ToUpper(strings.TrimSpace(*req.Estado))
	}

	fechaInicio := p.FechaInicio
	if req.FechaInicio != nil {
		fechaInicio = req.FechaInicio.UTC()
		updates["fecha_inicio"] = fechaInicio
	}
	if req.LimpiarFechaFin {
		updates["fecha_fin"] = nil
	} else if req.FechaFin != nil {
		ff := req.FechaFin.UTC()
		if ff.Before(fechaInicio) {
			return nil, ErrVigenciaInvalida
		}
		updates["fecha_fin"] = ff
	} else if p.FechaFin != nil && p.FechaFin.Before(fechaInicio) {
		return nil, ErrVigenciaInvalida
	}

	// Si el plan global por defecto aún no tenía usuario_administrador_id (p. ej. semilla previa), vincularlo
	if p.Alcance == AlcanceGlobal && p.UsuarioAdministradorID == nil && isGlobalAdmin {
		updates["usuario_administrador_id"] = claims.Sub
	}

	if err := s.repo.UpdatePrograma(ctx, id, updates); err != nil {
		if errors.Is(err, errCodigoDuplicado) {
			return nil, ErrCodigoDuplicado
		}
		if errors.Is(err, errProgramaNotFound) {
			return nil, ErrProgramaNoEncontrado
		}
		return nil, err
	}
	return s.ObtenerPrograma(ctx, claims, id)
}

func (s *Service) EliminarPrograma(ctx context.Context, claims *auth.Claims, id int32) error {
	p, err := s.repo.FindProgramaByID(ctx, id)
	if err != nil {
		if errors.Is(err, errProgramaNotFound) {
			return ErrProgramaNoEncontrado
		}
		return err
	}
	if p.EsSistema {
		return ErrPlanSistemaProtegido
	}
	isGlobalAdmin, err := s.isPlatformLoyaltyAdmin(ctx, claims)
	if err != nil {
		return err
	}
	if !s.canManagePrograma(claims, isGlobalAdmin, *p) {
		if p.Alcance == AlcanceGlobal {
			return ErrSinPermisoPlanGlobal
		}
		return ErrSinPermisoPlanEmpresa
	}
	if err := s.repo.DeletePrograma(ctx, id); err != nil {
		if errors.Is(err, errProgramaNotFound) {
			return ErrProgramaNoEncontrado
		}
		return err
	}
	return nil
}

func (s *Service) AdherirEmpresa(ctx context.Context, claims *auth.Claims, programaID int32, req AdherirEmpresaRequest) (*AdhesionResponse, error) {
	p, err := s.repo.FindProgramaByID(ctx, programaID)
	if err != nil {
		if errors.Is(err, errProgramaNotFound) {
			return nil, ErrProgramaNoEncontrado
		}
		return nil, err
	}
	if p.Estado != EstadoActivo {
		return nil, ErrPlanNoDisponible
	}
	if p.Alcance == AlcanceLocal && (p.EmpresaPropietariaID == nil || *p.EmpresaPropietariaID != claims.EmpresaID) {
		return nil, ErrAdhesionNoPermitida
	}

	rol := RolEmpresaAliada
	if p.EmpresaPropietariaID != nil && *p.EmpresaPropietariaID == claims.EmpresaID {
		rol = RolEmpresaPropietaria
	}

	permiteAcumular := true
	if req.PermiteAcumular != nil {
		permiteAcumular = *req.PermiteAcumular
	}
	permiteRedimir := true
	if req.PermiteRedimir != nil {
		permiteRedimir = *req.PermiteRedimir
	}
	esPrincipal := false
	if req.EsProgramaPrincipal != nil {
		esPrincipal = *req.EsProgramaPrincipal
	}

	row, err := s.repo.UpsertAdhesion(ctx, programaID, claims.EmpresaID, rol, permiteAcumular, permiteRedimir, esPrincipal)
	if err != nil {
		return nil, err
	}
	resp := toAdhesionResponse(*row)
	return &resp, nil
}

func (s *Service) ActualizarAdhesion(ctx context.Context, claims *auth.Claims, programaID int32, req ActualizarAdhesionRequest) (*AdhesionResponse, error) {
	if _, err := s.repo.FindProgramaByID(ctx, programaID); err != nil {
		if errors.Is(err, errProgramaNotFound) {
			return nil, ErrProgramaNoEncontrado
		}
		return nil, err
	}

	updates := map[string]any{}
	setPrincipal := false
	if req.PermiteAcumular != nil {
		updates["permite_acumular"] = *req.PermiteAcumular
	}
	if req.PermiteRedimir != nil {
		updates["permite_redimir"] = *req.PermiteRedimir
	}
	if req.EsProgramaPrincipal != nil {
		updates["es_programa_principal"] = *req.EsProgramaPrincipal
		setPrincipal = *req.EsProgramaPrincipal
	}

	row, err := s.repo.UpdateAdhesion(ctx, programaID, claims.EmpresaID, updates, setPrincipal)
	if err != nil {
		if errors.Is(err, errAdhesionNotFound) {
			return nil, ErrAdhesionNoEncontrada
		}
		return nil, err
	}
	resp := toAdhesionResponse(*row)
	return &resp, nil
}

func (s *Service) DesadherirEmpresa(ctx context.Context, claims *auth.Claims, programaID int32) error {
	if _, err := s.repo.FindProgramaByID(ctx, programaID); err != nil {
		if errors.Is(err, errProgramaNotFound) {
			return ErrProgramaNoEncontrado
		}
		return err
	}
	if err := s.repo.DeactivateAdhesion(ctx, programaID, claims.EmpresaID); err != nil {
		if errors.Is(err, errAdhesionNotFound) {
			return ErrAdhesionNoEncontrada
		}
		return err
	}
	return nil
}

func (s *Service) ListarEmpresasAdheridas(ctx context.Context, claims *auth.Claims, programaID int32) ([]AdhesionResponse, error) {
	p, err := s.repo.FindProgramaByID(ctx, programaID)
	if err != nil {
		if errors.Is(err, errProgramaNotFound) {
			return nil, ErrProgramaNoEncontrado
		}
		return nil, err
	}
	isGlobalAdmin, err := s.isPlatformLoyaltyAdmin(ctx, claims)
	if err != nil {
		return nil, err
	}
	if !s.canManagePrograma(claims, isGlobalAdmin, *p) {
		if p.Alcance == AlcanceGlobal {
			return nil, ErrSinPermisoPlanGlobal
		}
		return nil, ErrSinPermisoPlanEmpresa
	}

	rows, err := s.repo.ListEmpresasAdheridas(ctx, programaID)
	if err != nil {
		return nil, err
	}
	out := make([]AdhesionResponse, 0, len(rows))
	for _, row := range rows {
		out = append(out, toAdhesionResponse(row))
	}
	return out, nil
}
