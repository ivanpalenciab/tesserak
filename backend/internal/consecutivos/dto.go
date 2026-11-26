package consecutivos

import (
	"fmt"
	"strings"
	"time"
)

type CrearTipoDocumentoRequest struct {
	Codigo         string  `json:"codigo"          binding:"required,min=2,max=20"`
	Nombre         string  `json:"nombre"          binding:"required,min=2,max=100"`
	Descripcion    *string `json:"descripcion"     binding:"omitnil,max=255"`
	PrefijoDefecto string  `json:"prefijo_defecto" binding:"required,min=1,max=10"`
	Activo         *bool   `json:"activo"`
}

type TipoDocumentoResponse struct {
	ID             int32     `json:"id"`
	Codigo         string    `json:"codigo"`
	Nombre         string    `json:"nombre"`
	Descripcion    *string   `json:"descripcion"`
	PrefijoDefecto string    `json:"prefijo_defecto"`
	Activo         bool      `json:"activo"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type CrearSecuenciaRequest struct {
	TipoDocumentoID    int32   `json:"tipo_documento_id"   binding:"required,gt=0"`
	Prefijo            string  `json:"prefijo"             binding:"required,min=1,max=10"`
	SiguienteNumero    *int64  `json:"siguiente_numero"    binding:"omitnil,gte=1"`
	LongitudRelleno    *int16  `json:"longitud_relleno"    binding:"omitnil,gte=1,lte=20"`
	FrecuenciaReinicio *string `json:"frecuencia_reinicio" binding:"omitnil,oneof=NUNCA ANUAL MENSUAL"`
	Activo             *bool   `json:"activo"`
}

type ActualizarSecuenciaRequest struct {
	Prefijo            *string `json:"prefijo"             binding:"omitnil,min=1,max=10"`
	SiguienteNumero    *int64  `json:"siguiente_numero"    binding:"omitnil,gte=1"`
	LongitudRelleno    *int16  `json:"longitud_relleno"    binding:"omitnil,gte=1,lte=20"`
	FrecuenciaReinicio *string `json:"frecuencia_reinicio" binding:"omitnil,oneof=NUNCA ANUAL MENSUAL"`
	Activo             *bool   `json:"activo"`
}

type SecuenciaResponse struct {
	ID                  int64     `json:"id"`
	EmpresaID           int32     `json:"empresa_id"`
	TipoDocumentoID     int32     `json:"tipo_documento_id"`
	TipoDocumentoCodigo string    `json:"tipo_documento_codigo,omitempty"`
	TipoDocumentoNombre string    `json:"tipo_documento_nombre,omitempty"`
	Prefijo             string    `json:"prefijo"`
	SiguienteNumero     int64     `json:"siguiente_numero"`
	LongitudRelleno     int16     `json:"longitud_relleno"`
	FrecuenciaReinicio  string    `json:"frecuencia_reinicio"`
	ClavePeriodo        *string   `json:"clave_periodo"`
	Activo              bool      `json:"activo"`
	EjemploProximo      string    `json:"ejemplo_proximo"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

func toTipoResponse(t *TipoDocumento) TipoDocumentoResponse {
	return TipoDocumentoResponse{
		ID:             t.ID,
		Codigo:         t.Codigo,
		Nombre:         t.Nombre,
		Descripcion:    t.Descripcion,
		PrefijoDefecto: t.PrefijoDefecto,
		Activo:         t.Activo,
		CreatedAt:      t.CreatedAt,
		UpdatedAt:      t.UpdatedAt,
	}
}

func toSecuenciaResponse(s *SecuenciaDocumento, ref time.Time) SecuenciaResponse {
	periodo := calcularClavePeriodo(s.FrecuenciaReinicio, ref)
	numero := s.SiguienteNumero
	if periodo != nil && (s.ClavePeriodo == nil || *s.ClavePeriodo != *periodo) {
		numero = 1
	}

	resp := SecuenciaResponse{
		ID:                 s.ID,
		EmpresaID:          s.EmpresaID,
		TipoDocumentoID:    s.TipoDocumentoID,
		Prefijo:            s.Prefijo,
		SiguienteNumero:    s.SiguienteNumero,
		LongitudRelleno:    s.LongitudRelleno,
		FrecuenciaReinicio: s.FrecuenciaReinicio,
		ClavePeriodo:       s.ClavePeriodo,
		Activo:             s.Activo,
		EjemploProximo:     FormatearConsecutivo(s.Prefijo, periodo, int(s.LongitudRelleno), numero),
		CreatedAt:          s.CreatedAt,
		UpdatedAt:          s.UpdatedAt,
	}
	if s.TipoDocumento != nil {
		resp.TipoDocumentoCodigo = s.TipoDocumento.Codigo
		resp.TipoDocumentoNombre = s.TipoDocumento.Nombre
	}
	return resp
}

// calcularClavePeriodo determina la clave de periodo para el reinicio.
// Devuelve nil cuando la frecuencia es NUNCA.
func calcularClavePeriodo(frecuencia string, ref time.Time) *string {
	t := ref.UTC()
	switch strings.ToUpper(strings.TrimSpace(frecuencia)) {
	case FrecuenciaAnual:
		v := t.Format("2006")
		return &v
	case FrecuenciaMensual:
		v := t.Format("2006-01")
		return &v
	default:
		return nil
	}
}

// FormatearConsecutivo arma el código final del documento.
// Ejemplos:
//   - NUNCA:   COM-000001
//   - ANUAL:   COM-2026-000001
//   - MENSUAL: COM-2026-11-000001
func FormatearConsecutivo(prefijo string, clavePeriodo *string, padding int, numero int64) string {
	p := strings.TrimSpace(prefijo)
	if padding < 1 {
		padding = 6
	}
	numStr := fmt.Sprintf("%0*d", padding, numero)
	if clavePeriodo != nil && strings.TrimSpace(*clavePeriodo) != "" {
		return fmt.Sprintf("%s-%s-%s", p, strings.TrimSpace(*clavePeriodo), numStr)
	}
	return fmt.Sprintf("%s-%s", p, numStr)
}
