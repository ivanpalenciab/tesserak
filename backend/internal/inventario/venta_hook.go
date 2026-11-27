package inventario

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"tesserak/backend/internal/consecutivos"
)

type SalidaVentaLinea struct {
	ArticuloID int32
	Cantidad   float64
}

type SalidaVentaParams struct {
	EmpresaID     int32
	SucursalID    int32
	UbicacionID   int32
	ClienteID     int64
	UsuarioID     int64
	VentaID       int64
	Consecutivo   string
	Fecha         time.Time
	Observaciones *string
	Detalles      []SalidaVentaLinea
}

// RegistrarSalidaVentaTx crea el movimiento de inventario de tipo SALIDA_VENTA
// y descuenta inv_stock dentro de la transacción de la venta, solo si aún no existe
// un documento de inventario asociado a esa venta (operación idempotente).
// Si el stock es insuficiente o ocurre cualquier error, devuelve error para que
// la transacción de la venta realice rollback completo.
func RegistrarSalidaVentaTx(ctx context.Context, txDB *gorm.DB, p SalidaVentaParams) error {
	if len(p.Detalles) == 0 {
		return nil
	}

	// Los servicios y productos sin control de inventario no validan ni descuentan stock.
	artIDs := make([]int32, 0, len(p.Detalles))
	for _, d := range p.Detalles {
		artIDs = append(artIDs, d.ArticuloID)
	}
	invSet, err := filtrarArticulosInventariablesTx(ctx, txDB, p.EmpresaID, artIDs)
	if err != nil {
		return err
	}
	lineasInv := make([]SalidaVentaLinea, 0, len(p.Detalles))
	for _, d := range p.Detalles {
		if invSet[d.ArticuloID] {
			lineasInv = append(lineasInv, d)
		}
	}
	if len(lineasInv) == 0 {
		return nil
	}

	var existente int64
	if err := txDB.WithContext(ctx).
		Model(&DocumentoInventario{}).
		Where("empresa_id = ? AND documento_origen_tipo = ? AND documento_origen_id = ?", p.EmpresaID, "VENTA", p.VentaID).
		Count(&existente).Error; err != nil {
		return err
	}
	if existente > 0 {
		return nil
	}

	var tipo TipoMovimientoInventario
	if err := txDB.WithContext(ctx).
		Where("codigo = ? AND activo = TRUE", "SALIDA_VENTA").
		First(&tipo).Error; err != nil {
		return err
	}

	txRepo := NewRepository(txDB)

	okUbic, err := txRepo.UbicacionesBelongToEmpresa(ctx, p.EmpresaID, []int32{p.UbicacionID})
	if err != nil {
		return err
	}
	if !okUbic {
		return ErrReferenciaInvalida
	}

	consecutivoInv, err := consecutivos.GenerarConsecutivoTx(ctx, txDB, p.EmpresaID, consecutivos.TipoInventario, p.Fecha)
	if err != nil {
		return err
	}

	origenTipo := "VENTA"
	origenID := p.VentaID
	clienteID := p.ClienteID

	doc := &DocumentoInventario{
		EmpresaID:           p.EmpresaID,
		SucursalID:          p.SucursalID,
		TipoMovimientoID:    tipo.ID,
		Consecutivo:         consecutivoInv,
		Fecha:               p.Fecha,
		TerceroID:           &clienteID,
		Motivo:              fmt.Sprintf("Salida automática por venta %s", p.Consecutivo),
		Observaciones:       p.Observaciones,
		DocumentoOrigenTipo: &origenTipo,
		DocumentoOrigenID:   &origenID,
		UsuarioID:           p.UsuarioID,
		Estado:              EstadoAplicado,
	}
	if err := txRepo.CreateDocumento(ctx, doc); err != nil {
		return err
	}

	detallesInv := make([]MovimientoInventarioDetalle, 0, len(lineasInv))
	for _, line := range lineasInv {
		costoAplicado, err := aplicarSalida(ctx, txRepo, p.EmpresaID, line.ArticuloID, p.UbicacionID, line.Cantidad, nil)
		if err != nil {
			return err
		}
		ubicOrigen := p.UbicacionID
		detallesInv = append(detallesInv, MovimientoInventarioDetalle{
			DocumentoID:       doc.ID,
			ArticuloID:        line.ArticuloID,
			UbicacionOrigenID: &ubicOrigen,
			Cantidad:          line.Cantidad,
			CostoUnitario:     costoAplicado,
		})
	}

	return txRepo.CreateDetalles(ctx, detallesInv)
}
