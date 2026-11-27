package lealtad

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	errProgramaNotFound  = errors.New("plan de lealtad no encontrado")
	errAdhesionNotFound  = errors.New("adhesión de empresa no encontrada")
	errCodigoDuplicado   = errors.New("ya existe un plan de lealtad con ese código")
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) ListMonedas(ctx context.Context) ([]MonedaRef, error) {
	var items []MonedaRef
	err := r.db.WithContext(ctx).
		Order("CASE WHEN codigo_iso = 'COP' THEN 0 ELSE 1 END, codigo_iso ASC").
		Find(&items).Error
	return items, err
}

func (r *Repository) ExistsMoneda(ctx context.Context, monedaID int32) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&MonedaRef{}).
		Where("id = ?", monedaID).
		Count(&count).Error
	return count > 0, err
}

func (r *Repository) GetUsuarioUsername(ctx context.Context, userID int64) (string, error) {
	var u UsuarioRef
	err := r.db.WithContext(ctx).
		Select("id, nombre_usuario").
		Where("id = ? AND deleted_at IS NULL", userID).
		First(&u).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil
		}
		return "", err
	}
	return u.NombreUsuario, nil
}

// ListProgramasForEmpresa lista los planes visibles para el administrador de una empresa:
//   - Todos los planes GLOBAL (incluido Lealtad Tesserak)
//   - Todos los planes propios de su empresa (LOCAL o COALICION)
//   - Planes COALICION activos o donde su empresa esté adherida
func (r *Repository) ListProgramasForEmpresa(ctx context.Context, empresaID int32, f ProgramaFilter) ([]Programa, error) {
	q := r.db.WithContext(ctx).
		Model(&Programa{}).
		Preload("Moneda").
		Preload("EmpresaPropietaria").
		Preload("UsuarioAdministrador").
		Where(
			`(
				lealtad_programas.alcance = ?
				OR lealtad_programas.empresa_propietaria_id = ?
				OR (lealtad_programas.alcance = ? AND lealtad_programas.estado = ?)
				OR EXISTS (
					SELECT 1 FROM lealtad_programa_empresas lpe
					WHERE lpe.programa_id = lealtad_programas.id
					  AND lpe.empresa_id = ?
				)
			)`,
			AlcanceGlobal, empresaID, AlcanceCoalicion, EstadoActivo, empresaID,
		)

	if s := strings.TrimSpace(f.Search); s != "" {
		like := "%" + strings.ToLower(s) + "%"
		q = q.Where(
			"lower(lealtad_programas.nombre) LIKE ? OR lower(lealtad_programas.codigo) LIKE ? OR lower(COALESCE(lealtad_programas.descripcion, '')) LIKE ?",
			like, like, like,
		)
	}
	if a := strings.ToUpper(strings.TrimSpace(f.Alcance)); a != "" {
		q = q.Where("lealtad_programas.alcance = ?", a)
	}
	if e := strings.ToUpper(strings.TrimSpace(f.Estado)); e != "" {
		q = q.Where("lealtad_programas.estado = ?", e)
	}
	if f.SoloAdheridos {
		q = q.Where(
			`EXISTS (
				SELECT 1 FROM lealtad_programa_empresas lpe
				WHERE lpe.programa_id = lealtad_programas.id
				  AND lpe.empresa_id = ?
				  AND lpe.activo = TRUE
			)`,
			empresaID,
		)
	}

	var items []Programa
	err := q.
		Order("lealtad_programas.es_sistema DESC, CASE lealtad_programas.alcance WHEN 'GLOBAL' THEN 0 WHEN 'COALICION' THEN 1 ELSE 2 END, lealtad_programas.id ASC").
		Find(&items).Error
	return items, err
}

func (r *Repository) FindProgramaByID(ctx context.Context, id int32) (*Programa, error) {
	var p Programa
	err := r.db.WithContext(ctx).
		Preload("Moneda").
		Preload("EmpresaPropietaria").
		Preload("UsuarioAdministrador").
		Where("id = ?", id).
		First(&p).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errProgramaNotFound
		}
		return nil, err
	}
	return &p, nil
}

func (r *Repository) GetAdhesionesMapForEmpresa(ctx context.Context, empresaID int32, programaIDs []int32) (map[int32]ProgramaEmpresa, error) {
	result := make(map[int32]ProgramaEmpresa)
	if len(programaIDs) == 0 {
		return result, nil
	}
	var rows []ProgramaEmpresa
	err := r.db.WithContext(ctx).
		Where("empresa_id = ? AND programa_id IN ?", empresaID, programaIDs).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		result[row.ProgramaID] = row
	}
	return result, nil
}

func (r *Repository) GetActiveAdhesionCounts(ctx context.Context, programaIDs []int32) (map[int32]int64, error) {
	result := make(map[int32]int64)
	if len(programaIDs) == 0 {
		return result, nil
	}
	type countRow struct {
		ProgramaID int32 `gorm:"column:programa_id"`
		Total      int64 `gorm:"column:total"`
	}
	var rows []countRow
	err := r.db.WithContext(ctx).
		Model(&ProgramaEmpresa{}).
		Select("programa_id, COUNT(*) AS total").
		Where("programa_id IN ? AND activo = TRUE", programaIDs).
		Group("programa_id").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		result[row.ProgramaID] = row.Total
	}
	return result, nil
}

func (r *Repository) FindAdhesion(ctx context.Context, programaID, empresaID int32) (*ProgramaEmpresa, error) {
	var pe ProgramaEmpresa
	err := r.db.WithContext(ctx).
		Preload("Empresa").
		Where("programa_id = ? AND empresa_id = ?", programaID, empresaID).
		First(&pe).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errAdhesionNotFound
		}
		return nil, err
	}
	return &pe, nil
}

func (r *Repository) CreatePrograma(ctx context.Context, p *Programa, createPropietariaAdhesion bool) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(p).Error; err != nil {
			if isUniqueViolation(err) {
				return errCodigoDuplicado
			}
			return err
		}
		if createPropietariaAdhesion && p.EmpresaPropietariaID != nil {
			now := time.Now().UTC()
			adhesion := ProgramaEmpresa{
				ProgramaID:          p.ID,
				EmpresaID:           *p.EmpresaPropietariaID,
				RolEmpresa:          RolEmpresaPropietaria,
				PermiteAcumular:     true,
				PermiteRedimir:      true,
				EsProgramaPrincipal: false,
				FechaAdhesion:       now,
				Activo:              true,
			}
			if err := tx.Create(&adhesion).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *Repository) UpdatePrograma(ctx context.Context, id int32, updates map[string]any) error {
	res := r.db.WithContext(ctx).
		Model(&Programa{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Updates(updates)
	if res.Error != nil {
		if isUniqueViolation(res.Error) {
			return errCodigoDuplicado
		}
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errProgramaNotFound
	}
	return nil
}

func (r *Repository) DeletePrograma(ctx context.Context, id int32) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now().UTC()
		res := tx.Model(&Programa{}).
			Where("id = ? AND deleted_at IS NULL", id).
			Updates(map[string]any{
				"estado":     EstadoInactivo,
				"deleted_at": now,
				"updated_at": now,
			})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return errProgramaNotFound
		}
		if err := tx.Model(&ProgramaEmpresa{}).
			Where("programa_id = ? AND activo = TRUE", id).
			Updates(map[string]any{
				"activo":                false,
				"es_programa_principal": false,
				"fecha_baja":            now,
				"updated_at":            now,
			}).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *Repository) UpsertAdhesion(
	ctx context.Context,
	programaID, empresaID int32,
	rolEmpresa string,
	permiteAcumular, permiteRedimir, esPrincipal bool,
) (*ProgramaEmpresa, error) {
	now := time.Now().UTC()
	var result ProgramaEmpresa

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if esPrincipal {
			if err := tx.Model(&ProgramaEmpresa{}).
				Where("empresa_id = ? AND es_programa_principal = TRUE AND activo = TRUE AND programa_id <> ?", empresaID, programaID).
				Updates(map[string]any{
					"es_programa_principal": false,
					"updated_at":            now,
				}).Error; err != nil {
				return err
			}
		}

		pe := ProgramaEmpresa{
			ProgramaID:          programaID,
			EmpresaID:           empresaID,
			RolEmpresa:          rolEmpresa,
			PermiteAcumular:     permiteAcumular,
			PermiteRedimir:      permiteRedimir,
			EsProgramaPrincipal: esPrincipal,
			FechaAdhesion:       now,
			FechaBaja:           nil,
			Activo:              true,
			UpdatedAt:           now,
		}

		if err := tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "programa_id"}, {Name: "empresa_id"}},
			DoUpdates: clause.Assignments(map[string]any{
				"rol_empresa":           rolEmpresa,
				"permite_acumular":      permiteAcumular,
				"permite_redimir":       permiteRedimir,
				"es_programa_principal": esPrincipal,
				"fecha_adhesion":        now,
				"fecha_baja":            nil,
				"activo":                true,
				"updated_at":            now,
			}),
		}).Create(&pe).Error; err != nil {
			return err
		}

		return tx.Preload("Empresa").
			Where("programa_id = ? AND empresa_id = ?", programaID, empresaID).
			First(&result).Error
	})
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (r *Repository) UpdateAdhesion(
	ctx context.Context,
	programaID, empresaID int32,
	updates map[string]any,
	setPrincipal bool,
) (*ProgramaEmpresa, error) {
	now := time.Now().UTC()
	var result ProgramaEmpresa

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if setPrincipal {
			if err := tx.Model(&ProgramaEmpresa{}).
				Where("empresa_id = ? AND es_programa_principal = TRUE AND activo = TRUE AND programa_id <> ?", empresaID, programaID).
				Updates(map[string]any{
					"es_programa_principal": false,
					"updated_at":            now,
				}).Error; err != nil {
				return err
			}
		}

		updates["updated_at"] = now
		res := tx.Model(&ProgramaEmpresa{}).
			Where("programa_id = ? AND empresa_id = ? AND activo = TRUE", programaID, empresaID).
			Updates(updates)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return errAdhesionNotFound
		}

		return tx.Preload("Empresa").
			Where("programa_id = ? AND empresa_id = ?", programaID, empresaID).
			First(&result).Error
	})
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (r *Repository) DeactivateAdhesion(ctx context.Context, programaID, empresaID int32) error {
	now := time.Now().UTC()
	res := r.db.WithContext(ctx).
		Model(&ProgramaEmpresa{}).
		Where("programa_id = ? AND empresa_id = ? AND activo = TRUE", programaID, empresaID).
		Updates(map[string]any{
			"activo":                false,
			"es_programa_principal": false,
			"fecha_baja":            now,
			"updated_at":            now,
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errAdhesionNotFound
	}
	return nil
}

func (r *Repository) ListEmpresasAdheridas(ctx context.Context, programaID int32) ([]ProgramaEmpresa, error) {
	var rows []ProgramaEmpresa
	err := r.db.WithContext(ctx).
		Preload("Empresa").
		Where("programa_id = ?", programaID).
		Order("activo DESC, rol_empresa DESC, fecha_adhesion ASC").
		Find(&rows).Error
	return rows, err
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return true
	}
	return strings.Contains(strings.ToLower(err.Error()), "duplicate key")
}
