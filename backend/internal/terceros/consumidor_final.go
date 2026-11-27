package terceros

import (
	"context"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	ConsumidorFinalTipoID     = "CC"
	ConsumidorFinalNumeroID   = "222222222222"
	ConsumidorFinalNombre     = "Consumidor Final"
	ConsumidorFinalNaturaleza = "NATURAL"
)

// AsegurarConsumidorFinalTx garantiza que la empresa tenga creado el cliente
// por defecto "Consumidor Final" (identificación 222222222222) con rol único CLIENTE
// dentro de la transacción activa y devuelve su tercero_maestro.id.
func AsegurarConsumidorFinalTx(ctx context.Context, txDB *gorm.DB, empresaID int32) (int64, error) {
	var existente TerceroMaestro
	err := txDB.WithContext(ctx).
		Where("empresa_id = ? AND lower(numero_identificacion) = ?", empresaID, ConsumidorFinalNumeroID).
		First(&existente).Error

	var terceroID int64
	if err == nil {
		terceroID = existente.ID
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		nuevo := &TerceroMaestro{
			EmpresaID:            empresaID,
			TipoIdentificacion:   ConsumidorFinalTipoID,
			NumeroIdentificacion: ConsumidorFinalNumeroID,
			RazonSocial:          ConsumidorFinalNombre,
			NaturalezaJuridica:   ConsumidorFinalNaturaleza,
			Activo:               true,
		}
		if err := txDB.WithContext(ctx).Omit(clause.Associations).Create(nuevo).Error; err != nil {
			return 0, err
		}
		terceroID = nuevo.ID
	} else {
		return 0, err
	}

	var rolCliente RolTercero
	if err := txDB.WithContext(ctx).
		Where("codigo = ?", RolCodigoCliente).
		First(&rolCliente).Error; err != nil {
		return 0, err
	}

	asignacion := TerceroRol{
		TerceroID: terceroID,
		RolID:     rolCliente.ID,
	}
	if err := txDB.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(&asignacion).Error; err != nil {
		return 0, err
	}

	return terceroID, nil
}
