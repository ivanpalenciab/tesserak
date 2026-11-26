package consecutivos

import (
	"context"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	errTipoNotFound      = errors.New("tipo de documento no encontrado")
	errSecuenciaNotFound = errors.New("secuencia de documento no encontrada")
	errSecuenciaInactiva = errors.New("la secuencia de documento está inactiva")
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// =====================================================================
// Tipos de documento
// =====================================================================

func (r *Repository) ListTipos(ctx context.Context) ([]TipoDocumento, error) {
	var items []TipoDocumento
	err := r.db.WithContext(ctx).Order("id ASC").Find(&items).Error
	return items, err
}

func (r *Repository) CreateTipo(ctx context.Context, t *TipoDocumento) error {
	return r.db.WithContext(ctx).Create(t).Error
}

// =====================================================================
// Secuencias por empresa
// =====================================================================

func (r *Repository) ListSecuencias(ctx context.Context, empresaID int32) ([]SecuenciaDocumento, error) {
	var items []SecuenciaDocumento
	err := r.db.WithContext(ctx).
		Preload("TipoDocumento").
		Where("empresa_id = ?", empresaID).
		Order("tipo_documento_id ASC, id ASC").
		Find(&items).Error
	return items, err
}

func (r *Repository) FindSecuenciaByID(ctx context.Context, empresaID int32, id int64) (*SecuenciaDocumento, error) {
	var s SecuenciaDocumento
	err := r.db.WithContext(ctx).
		Preload("TipoDocumento").
		Where("empresa_id = ? AND id = ?", empresaID, id).
		First(&s).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errSecuenciaNotFound
	}
	return &s, err
}

func (r *Repository) CreateSecuencia(ctx context.Context, s *SecuenciaDocumento) error {
	return r.db.WithContext(ctx).Omit(clause.Associations).Create(s).Error
}

func (r *Repository) UpdateSecuencia(ctx context.Context, empresaID int32, id int64, cambios map[string]any) error {
	if len(cambios) == 0 {
		return nil
	}
	res := r.db.WithContext(ctx).
		Model(&SecuenciaDocumento{}).
		Where("empresa_id = ? AND id = ?", empresaID, id).
		Updates(cambios)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errSecuenciaNotFound
	}
	return nil
}

// =====================================================================
// Generación atómica de consecutivos (dentro de la transacción del documento)
// =====================================================================

// GenerarConsecutivoTx obtiene y reserva el siguiente consecutivo disponible para
// (empresaID, codigoTipo) usando bloqueo de fila FOR UPDATE dentro de la transacción tx.
// Si la empresa aún no tiene fila en secuencias_documento, la crea con el prefijo por
// defecto de tipos_documento.
func GenerarConsecutivoTx(ctx context.Context, tx *gorm.DB, empresaID int32, codigoTipo string, fechaRef time.Time) (string, error) {
	codigo := strings.ToUpper(strings.TrimSpace(codigoTipo))

	var tipo TipoDocumento
	if err := tx.WithContext(ctx).Where("codigo = ? AND activo = TRUE", codigo).First(&tipo).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", ErrTipoNoEncontrado
		}
		return "", err
	}

	// Asegura que exista la secuencia para esta empresa (útil para empresas recién creadas).
	seed := SecuenciaDocumento{
		EmpresaID:          empresaID,
		TipoDocumentoID:    tipo.ID,
		Prefijo:            tipo.PrefijoDefecto,
		SiguienteNumero:    1,
		LongitudRelleno:    6,
		FrecuenciaReinicio: FrecuenciaNunca,
		Activo:             true,
	}
	if err := tx.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "empresa_id"}, {Name: "tipo_documento_id"}},
			DoNothing: true,
		}).
		Omit(clause.Associations).
		Create(&seed).Error; err != nil {
		return "", err
	}

	// Bloqueo pesimista FOR UPDATE: serializa la emisión del consecutivo por empresa y tipo.
	var seq SecuenciaDocumento
	if err := tx.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("empresa_id = ? AND tipo_documento_id = ?", empresaID, tipo.ID).
		First(&seq).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", ErrSecuenciaNoEncontrada
		}
		return "", err
	}

	if !seq.Activo {
		return "", ErrSecuenciaInactiva
	}

	if fechaRef.IsZero() {
		fechaRef = time.Now().UTC()
	}
	periodoActual := calcularClavePeriodo(seq.FrecuenciaReinicio, fechaRef)

	numeroAUsar := seq.SiguienteNumero
	if periodoActual != nil && (seq.ClavePeriodo == nil || *seq.ClavePeriodo != *periodoActual) {
		numeroAUsar = 1
	}

	cambios := map[string]any{
		"siguiente_numero": numeroAUsar + 1,
		"clave_periodo":    periodoActual,
		"updated_at":       time.Now().UTC(),
	}
	if err := tx.WithContext(ctx).
		Model(&SecuenciaDocumento{}).
		Where("id = ?", seq.ID).
		Updates(cambios).Error; err != nil {
		return "", err
	}

	return FormatearConsecutivo(seq.Prefijo, periodoActual, int(seq.LongitudRelleno), numeroAUsar), nil
}
