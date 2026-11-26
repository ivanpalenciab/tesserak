package consecutivos

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"

	"tesserak/backend/internal/httpx"
)

var (
	ErrTipoNoEncontrado      = httpx.NewError(http.StatusNotFound, "tipo_documento_no_encontrado", "el tipo de documento no existe o está inactivo")
	ErrTipoDuplicado         = httpx.NewError(http.StatusConflict, "tipo_documento_duplicado", "ya existe un tipo de documento con ese código")
	ErrSecuenciaNoEncontrada = httpx.NewError(http.StatusNotFound, "secuencia_no_encontrada", "la secuencia de documento no existe")
	ErrSecuenciaDuplicada    = httpx.NewError(http.StatusConflict, "secuencia_duplicada", "la empresa ya tiene una secuencia configurada para ese tipo de documento")
	ErrSecuenciaInactiva     = httpx.NewError(http.StatusUnprocessableEntity, "secuencia_inactiva", "la secuencia de consecutivos para este tipo de documento está inactiva")
	ErrReferenciaInvalida    = httpx.NewError(http.StatusUnprocessableEntity, "referencia_invalida", "el tipo de documento o empresa indicado no es válido")
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListarTipos(ctx context.Context) ([]TipoDocumentoResponse, error) {
	items, err := s.repo.ListTipos(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]TipoDocumentoResponse, 0, len(items))
	for i := range items {
		out = append(out, toTipoResponse(&items[i]))
	}
	return out, nil
}

func (s *Service) CrearTipo(ctx context.Context, req CrearTipoDocumentoRequest) (*TipoDocumentoResponse, error) {
	activo := true
	if req.Activo != nil {
		activo = *req.Activo
	}
	t := &TipoDocumento{
		Codigo:         strings.ToUpper(strings.TrimSpace(req.Codigo)),
		Nombre:         strings.TrimSpace(req.Nombre),
		Descripcion:    trimPtr(req.Descripcion),
		PrefijoDefecto: strings.ToUpper(strings.TrimSpace(req.PrefijoDefecto)),
		Activo:         activo,
	}
	if err := s.repo.CreateTipo(ctx, t); err != nil {
		return nil, mapError(err)
	}
	resp := toTipoResponse(t)
	return &resp, nil
}

func (s *Service) ListarSecuencias(ctx context.Context, empresaID int32) ([]SecuenciaResponse, error) {
	items, err := s.repo.ListSecuencias(ctx, empresaID)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	out := make([]SecuenciaResponse, 0, len(items))
	for i := range items {
		out = append(out, toSecuenciaResponse(&items[i], now))
	}
	return out, nil
}

func (s *Service) CrearSecuencia(ctx context.Context, empresaID int32, req CrearSecuenciaRequest) (*SecuenciaResponse, error) {
	siguiente := int64(1)
	if req.SiguienteNumero != nil {
		siguiente = *req.SiguienteNumero
	}
	relleno := int16(6)
	if req.LongitudRelleno != nil {
		relleno = *req.LongitudRelleno
	}
	frecuencia := FrecuenciaNunca
	if req.FrecuenciaReinicio != nil {
		frecuencia = *req.FrecuenciaReinicio
	}
	activo := true
	if req.Activo != nil {
		activo = *req.Activo
	}

	now := time.Now().UTC()
	seq := &SecuenciaDocumento{
		EmpresaID:          empresaID,
		TipoDocumentoID:    req.TipoDocumentoID,
		Prefijo:            strings.ToUpper(strings.TrimSpace(req.Prefijo)),
		SiguienteNumero:    siguiente,
		LongitudRelleno:    relleno,
		FrecuenciaReinicio: frecuencia,
		ClavePeriodo:       calcularClavePeriodo(frecuencia, now),
		Activo:             activo,
	}
	if err := s.repo.CreateSecuencia(ctx, seq); err != nil {
		return nil, mapError(err)
	}

	creada, err := s.repo.FindSecuenciaByID(ctx, empresaID, seq.ID)
	if err != nil {
		return nil, mapError(err)
	}
	resp := toSecuenciaResponse(creada, now)
	return &resp, nil
}

func (s *Service) ActualizarSecuencia(ctx context.Context, empresaID int32, id int64, req ActualizarSecuenciaRequest) (*SecuenciaResponse, error) {
	actual, err := s.repo.FindSecuenciaByID(ctx, empresaID, id)
	if err != nil {
		return nil, mapError(err)
	}

	now := time.Now().UTC()
	cambios := map[string]any{}
	if req.Prefijo != nil {
		cambios["prefijo"] = strings.ToUpper(strings.TrimSpace(*req.Prefijo))
	}
	if req.SiguienteNumero != nil {
		cambios["siguiente_numero"] = *req.SiguienteNumero
	}
	if req.LongitudRelleno != nil {
		cambios["longitud_relleno"] = *req.LongitudRelleno
	}
	if req.FrecuenciaReinicio != nil {
		cambios["frecuencia_reinicio"] = *req.FrecuenciaReinicio
		if *req.FrecuenciaReinicio != actual.FrecuenciaReinicio {
			cambios["clave_periodo"] = calcularClavePeriodo(*req.FrecuenciaReinicio, now)
		}
	}
	if req.Activo != nil {
		cambios["activo"] = *req.Activo
	}

	if err := s.repo.UpdateSecuencia(ctx, empresaID, id, cambios); err != nil {
		return nil, mapError(err)
	}

	actualizada, err := s.repo.FindSecuenciaByID(ctx, empresaID, id)
	if err != nil {
		return nil, mapError(err)
	}
	resp := toSecuenciaResponse(actualizada, now)
	return &resp, nil
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
	switch {
	case errors.Is(err, errTipoNotFound):
		return ErrTipoNoEncontrado
	case errors.Is(err, errSecuenciaNotFound):
		return ErrSecuenciaNoEncontrada
	case errors.Is(err, errSecuenciaInactiva):
		return ErrSecuenciaInactiva
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			if pgErr.ConstraintName == "uq_tipos_documento_codigo" {
				return ErrTipoDuplicado
			}
			return ErrSecuenciaDuplicada
		case "23503":
			return ErrReferenciaInvalida
		}
	}
	return err
}
