package crm

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"tesserak/backend/internal/httpx"
)

type Moneda struct {
	ID           int32  `gorm:"column:id;primaryKey"`
	CodigoISO    string `gorm:"column:codigo_iso;not null"`
	NombreMoneda string `gorm:"column:nombre_moneda;not null"`
	Simbolo      string `gorm:"column:simbolo;not null"`
}

func (Moneda) TableName() string { return "monedas" }

type MonedaResponse struct {
	ID           int32  `json:"id"`
	CodigoISO    string `json:"codigo_iso"`
	NombreMoneda string `json:"nombre_moneda"`
	Simbolo      string `json:"simbolo"`
}

func (r *Repository) ListMonedas(ctx context.Context) ([]Moneda, error) {
	var items []Moneda
	err := r.db.WithContext(ctx).Order("codigo_iso ASC").Find(&items).Error
	return items, err
}

func (s *Service) ListarMonedas(ctx context.Context) ([]MonedaResponse, error) {
	items, err := s.repo.ListMonedas(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]MonedaResponse, 0, len(items))
	for i := range items {
		out = append(out, MonedaResponse{
			ID:           items[i].ID,
			CodigoISO:    strings.TrimSpace(items[i].CodigoISO),
			NombreMoneda: items[i].NombreMoneda,
			Simbolo:      items[i].Simbolo,
		})
	}
	return out, nil
}

func (h *Handler) listarMonedas(c *gin.Context) {
	resp, err := h.svc.ListarMonedas(c.Request.Context())
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}
