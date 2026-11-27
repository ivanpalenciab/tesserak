package empresas

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"

	"tesserak/backend/internal/auth"
	"tesserak/backend/internal/httpx"
)

// Sucursal mapea la tabla "sucursales" (con soft delete).
type Sucursal struct {
	ID        int32          `gorm:"column:id;primaryKey"`
	EmpresaID int32          `gorm:"column:empresa_id;not null"`
	CiudadID  *int32         `gorm:"column:ciudad_id"`
	Nombre    string         `gorm:"column:nombre;not null"`
	Direccion *string        `gorm:"column:direccion"`
	Activo    bool           `gorm:"column:activo;not null"`
	CreatedAt time.Time      `gorm:"column:created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (Sucursal) TableName() string { return "sucursales" }

type CrearSucursalRequest struct {
	Nombre    string  `json:"nombre"    binding:"required,min=1,max=150"`
	Direccion *string `json:"direccion" binding:"omitnil,max=255"`
	CiudadID  *int32  `json:"ciudad_id" binding:"omitnil,gt=0"`
	Activo    *bool   `json:"activo"`
}

type ActualizarSucursalRequest struct {
	Nombre    *string `json:"nombre"    binding:"omitnil,min=1,max=150"`
	Direccion *string `json:"direccion" binding:"omitnil,max=255"`
	CiudadID  *int32  `json:"ciudad_id" binding:"omitnil,gt=0"`
	Activo    *bool   `json:"activo"`
}

type SucursalResponse struct {
	ID        int32     `json:"id"`
	EmpresaID int32     `json:"empresa_id"`
	CiudadID  *int32    `json:"ciudad_id"`
	Nombre    string    `json:"nombre"`
	Direccion *string   `json:"direccion"`
	Activo    bool      `json:"activo"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

var (
	ErrSucursalNoEncontrada = httpx.NewError(http.StatusNotFound, "sucursal_no_encontrada", "la sucursal no existe")
	ErrUltimaSucursal       = httpx.NewError(http.StatusUnprocessableEntity, "ultima_sucursal", "no se puede eliminar la única sucursal de la empresa")
	ErrCiudadInvalida       = httpx.NewError(http.StatusUnprocessableEntity, "ciudad_invalida", "la ciudad indicada no existe")
)

func toSucursalResponse(s *Sucursal) SucursalResponse {
	return SucursalResponse{
		ID:        s.ID,
		EmpresaID: s.EmpresaID,
		CiudadID:  s.CiudadID,
		Nombre:    s.Nombre,
		Direccion: s.Direccion,
		Activo:    s.Activo,
		CreatedAt: s.CreatedAt,
		UpdatedAt: s.UpdatedAt,
	}
}

// =====================================================================
// Repository de sucursales
// =====================================================================

func (r *Repository) ListSucursales(ctx context.Context, empresaID int32) ([]Sucursal, error) {
	var items []Sucursal
	err := r.db.WithContext(ctx).
		Where("empresa_id = ?", empresaID).
		Order("id ASC").
		Find(&items).Error
	return items, err
}

func (r *Repository) FindSucursalByID(ctx context.Context, empresaID, id int32) (*Sucursal, error) {
	var s Sucursal
	err := r.db.WithContext(ctx).
		Where("empresa_id = ? AND id = ?", empresaID, id).
		First(&s).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrSucursalNoEncontrada
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *Repository) CountSucursales(ctx context.Context, empresaID int32) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).
		Model(&Sucursal{}).
		Where("empresa_id = ?", empresaID).
		Count(&total).Error
	return total, err
}

func (r *Repository) CreateSucursal(ctx context.Context, s *Sucursal) error {
	return r.db.WithContext(ctx).Create(s).Error
}

func (r *Repository) UpdateSucursal(ctx context.Context, s *Sucursal, cambios map[string]any) error {
	if len(cambios) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Model(s).Updates(cambios).Error
}

func (r *Repository) SoftDeleteSucursal(ctx context.Context, empresaID, id int32) error {
	res := r.db.WithContext(ctx).
		Where("empresa_id = ? AND id = ?", empresaID, id).
		Delete(&Sucursal{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrSucursalNoEncontrada
	}
	return nil
}

// =====================================================================
// Service de sucursales
// =====================================================================

func (s *Service) ListarSucursales(ctx context.Context, empresaID int32) ([]SucursalResponse, error) {
	items, err := s.repo.ListSucursales(ctx, empresaID)
	if err != nil {
		return nil, mapSucursalError(err)
	}
	out := make([]SucursalResponse, 0, len(items))
	for i := range items {
		out = append(out, toSucursalResponse(&items[i]))
	}
	return out, nil
}

func (r *Repository) CreateUbicacionDefault(ctx context.Context, empresaID, sucursalID int32, sucursalNombre string) error {
	nombre := fmt.Sprintf("Bodega Principal - %s", strings.TrimSpace(sucursalNombre))
	if len(nombre) > 120 {
		nombre = nombre[:120]
	}
	codigo := fmt.Sprintf("BOD-%d", sucursalID)
	desc := "Ubicación principal creada automáticamente al registrar la sucursal"
	return r.db.WithContext(ctx).
		Table("inv_ubicaciones").
		Create(map[string]any{
			"empresa_id":  empresaID,
			"sucursal_id": sucursalID,
			"codigo":      codigo,
			"nombre":      nombre,
			"descripcion": desc,
			"activo":      true,
		}).Error
}

func (s *Service) CrearSucursal(ctx context.Context, empresaID int32, req CrearSucursalRequest) (*SucursalResponse, error) {
	activo := true
	if req.Activo != nil {
		activo = *req.Activo
	}

	suc := &Sucursal{
		EmpresaID: empresaID,
		CiudadID:  req.CiudadID,
		Nombre:    strings.TrimSpace(req.Nombre),
		Direccion: trimPtr(req.Direccion),
		Activo:    activo,
	}
	err := s.repo.WithTx(ctx, func(tx *Repository) error {
		if err := tx.CreateSucursal(ctx, suc); err != nil {
			return err
		}
		return tx.CreateUbicacionDefault(ctx, empresaID, suc.ID, suc.Nombre)
	})
	if err != nil {
		return nil, mapSucursalError(err)
	}
	resp := toSucursalResponse(suc)
	return &resp, nil
}

func (s *Service) ActualizarSucursal(ctx context.Context, empresaID, id int32, req ActualizarSucursalRequest) (*SucursalResponse, error) {
	var result *Sucursal
	err := s.repo.WithTx(ctx, func(tx *Repository) error {
		suc, err := tx.FindSucursalByID(ctx, empresaID, id)
		if err != nil {
			return err
		}

		cambios := map[string]any{}
		if req.Nombre != nil {
			cambios["nombre"] = strings.TrimSpace(*req.Nombre)
		}
		if req.Direccion != nil {
			cambios["direccion"] = trimPtr(req.Direccion)
		}
		if req.CiudadID != nil {
			cambios["ciudad_id"] = *req.CiudadID
		}
		if req.Activo != nil {
			cambios["activo"] = *req.Activo
		}

		if err := tx.UpdateSucursal(ctx, suc, cambios); err != nil {
			return err
		}
		result, err = tx.FindSucursalByID(ctx, empresaID, id)
		return err
	})
	if err != nil {
		return nil, mapSucursalError(err)
	}
	resp := toSucursalResponse(result)
	return &resp, nil
}

func (s *Service) EliminarSucursal(ctx context.Context, empresaID, id int32) error {
	return mapSucursalError(s.repo.WithTx(ctx, func(tx *Repository) error {
		if _, err := tx.FindSucursalByID(ctx, empresaID, id); err != nil {
			return err
		}
		total, err := tx.CountSucursales(ctx, empresaID)
		if err != nil {
			return err
		}
		if total <= 1 {
			return ErrUltimaSucursal
		}
		return tx.SoftDeleteSucursal(ctx, empresaID, id)
	}))
}

func mapSucursalError(err error) error {
	if err == nil {
		return nil
	}
	var appErr *httpx.AppError
	if errors.As(err, &appErr) {
		return err
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23503" {
		return ErrCiudadInvalida
	}
	return err
}

// =====================================================================
// Handlers HTTP de sucursales (/api/v1/empresas/sucursales)
// =====================================================================

func (h *Handler) listarSucursales(c *gin.Context) {
	claims, ok := auth.ClaimsFromContext(c)
	if !ok {
		httpx.RespondError(c, auth.ErrNoAutenticado)
		return
	}
	resp, err := h.svc.ListarSucursales(c.Request.Context(), claims.EmpresaID)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	if !claims.IsAdmin() {
		filtradas := make([]SucursalResponse, 0, len(resp))
		for _, s := range resp {
			if claims.HasSucursal(s.ID) {
				filtradas = append(filtradas, s)
			}
		}
		resp = filtradas
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}

func (h *Handler) crearSucursal(c *gin.Context) {
	claims, ok := auth.ClaimsFromContext(c)
	if !ok {
		httpx.RespondError(c, auth.ErrNoAutenticado)
		return
	}
	var req CrearSucursalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondError(c, httpx.BindError(err))
		return
	}
	resp, err := h.svc.CrearSucursal(c.Request.Context(), claims.EmpresaID, req)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *Handler) actualizarSucursal(c *gin.Context) {
	claims, ok := auth.ClaimsFromContext(c)
	if !ok {
		httpx.RespondError(c, auth.ErrNoAutenticado)
		return
	}
	id, err := parseID(c)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	var req ActualizarSucursalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondError(c, httpx.BindError(err))
		return
	}
	resp, err := h.svc.ActualizarSucursal(c.Request.Context(), claims.EmpresaID, id, req)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) eliminarSucursal(c *gin.Context) {
	claims, ok := auth.ClaimsFromContext(c)
	if !ok {
		httpx.RespondError(c, auth.ErrNoAutenticado)
		return
	}
	id, err := parseID(c)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	if err := h.svc.EliminarSucursal(c.Request.Context(), claims.EmpresaID, id); err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
