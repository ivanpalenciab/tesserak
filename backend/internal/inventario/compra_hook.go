package inventario

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"tesserak/backend/internal/consecutivos"
)

type EntradaCompraLinea struct {
	ArticuloID    int32
	Cantidad      float64
	CostoUnitario float64
}

type EntradaCompraParams struct {
	EmpresaID     int32
	SucursalID    int32
	UsuarioID     int64
	CompraID      int64
	Consecutivo   string
	Fecha         time.Time
	Observaciones *string
	Detalles      []EntradaCompraLinea
}

// RegistrarEntradaCompraTx crea el movimiento de inventario de tipo ENTRADA_COMPRA
// y actualiza inv_stock dentro de la transacción de la compra, solo si aún no existe
// un documento de inventario asociado a esa compra (operación idempotente).
func RegistrarEntradaCompraTx(ctx context.Context, txDB *gorm.DB, p EntradaCompraParams) error {
	if len(p.Detalles) == 0 {
		return nil
	}

	var existente int64
	if err := txDB.WithContext(ctx).
		Model(&DocumentoInventario{}).
		Where("empresa_id = ? AND documento_origen_tipo = ? AND documento_origen_id = ?", p.EmpresaID, "COMPRA", p.CompraID).
		Count(&existente).Error; err != nil {
		return err
	}
	if existente > 0 {
		return nil
	}

	var tipo TipoMovimientoInventario
	if err := txDB.WithContext(ctx).
		Where("codigo = ? AND activo = TRUE", "ENTRADA_COMPRA").
		First(&tipo).Error; err != nil {
		return err
	}

	ubicacionID, err := resolverUbicacionDestinoTx(ctx, txDB, p.EmpresaID, p.SucursalID)
	if err != nil {
		return err
	}

	consecutivoInv, err := consecutivos.GenerarConsecutivoTx(ctx, txDB, p.EmpresaID, consecutivos.TipoInventario, p.Fecha)
	if err != nil {
		return err
	}

	origenTipo := "COMPRA"
	origenID := p.CompraID
	txRepo := NewRepository(txDB)

	doc := &DocumentoInventario{
		EmpresaID:           p.EmpresaID,
		SucursalID:          p.SucursalID,
		TipoMovimientoID:    tipo.ID,
		Consecutivo:         consecutivoInv,
		Fecha:               p.Fecha,
		Motivo:              fmt.Sprintf("Ingreso automático por compra %s", p.Consecutivo),
		Observaciones:       p.Observaciones,
		DocumentoOrigenTipo: &origenTipo,
		DocumentoOrigenID:   &origenID,
		UsuarioID:           p.UsuarioID,
		Estado:              EstadoAplicado,
	}
	if err := txRepo.CreateDocumento(ctx, doc); err != nil {
		return err
	}

	detallesInv := make([]MovimientoInventarioDetalle, 0, len(p.Detalles))
	for _, line := range p.Detalles {
		if err := aplicarEntrada(ctx, txRepo, p.EmpresaID, line.ArticuloID, ubicacionID, line.Cantidad, line.CostoUnitario); err != nil {
			return err
		}
		destID := ubicacionID
		detallesInv = append(detallesInv, MovimientoInventarioDetalle{
			DocumentoID:        doc.ID,
			ArticuloID:         line.ArticuloID,
			UbicacionDestinoID: &destID,
			Cantidad:           line.Cantidad,
			CostoUnitario:      line.CostoUnitario,
		})
	}

	return txRepo.CreateDetalles(ctx, detallesInv)
}

func resolverUbicacionDestinoTx(ctx context.Context, txDB *gorm.DB, empresaID, sucursalID int32) (int32, error) {
	var u Ubicacion
	err := txDB.WithContext(ctx).
		Where("empresa_id = ? AND sucursal_id = ? AND activo = TRUE", empresaID, sucursalID).
		Order("id ASC").
		First(&u).Error
	if err == nil {
		return u.ID, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, err
	}

	// Fallback: primera ubicación activa de la empresa
	err = txDB.WithContext(ctx).
		Where("empresa_id = ? AND activo = TRUE", empresaID).
		Order("id ASC").
		First(&u).Error
	if err == nil {
		return u.ID, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, err
	}

	// Si la empresa aún no tiene ninguna ubicación, crea Bodega Principal automáticamente
	desc := "Ubicación principal creada automáticamente"
	nueva := &Ubicacion{
		EmpresaID:   empresaID,
		SucursalID:  sucursalID,
		Codigo:      fmt.Sprintf("BOD-%d", sucursalID),
		Nombre:      "Bodega Principal",
		Descripcion: &desc,
		Activo:      true,
	}
	if err := txDB.WithContext(ctx).Create(nueva).Error; err != nil {
		return 0, err
	}
	return nueva.ID, nil
}
