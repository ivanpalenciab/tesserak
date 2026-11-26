-- =====================================================================
-- 000018 - Migración de compras existentes al inventario (inv_documentos,
--          inv_movimientos_detalle e inv_stock) de forma idempotente.
-- Motor: PostgreSQL 14+
-- =====================================================================

-- 1. Asegurar que exista al menos una ubicación por cada sucursal con compras
INSERT INTO inv_ubicaciones (empresa_id, sucursal_id, codigo, nombre, descripcion, activo)
SELECT DISTINCT c.empresa_id, c.sucursal_id, 'BOD-' || c.sucursal_id, 'Bodega Principal', 'Ubicación generada automáticamente', TRUE
FROM com_compras c
WHERE c.deleted_at IS NULL
  AND NOT EXISTS (
      SELECT 1 FROM inv_ubicaciones u
      WHERE u.sucursal_id = c.sucursal_id AND u.empresa_id = c.empresa_id AND u.deleted_at IS NULL
  )
ON CONFLICT DO NOTHING;

-- 2. Crear inv_documentos (ENTRADA_COMPRA) para compras activas sin documento de inventario previo
INSERT INTO inv_documentos (
    empresa_id,
    sucursal_id,
    tipo_movimiento_id,
    consecutivo,
    fecha,
    tercero_id,
    motivo,
    observaciones,
    documento_origen_tipo,
    documento_origen_id,
    usuario_id,
    estado
)
SELECT
    c.empresa_id,
    c.sucursal_id,
    (SELECT id FROM inv_tipos_movimiento WHERE codigo = 'ENTRADA_COMPRA'),
    'INV-COMP-' || c.id,
    c.fecha_compra,
    NULL,
    'Ingreso automático por compra ' || c.consecutivo,
    c.observaciones,
    'COMPRA',
    c.id,
    c.usuario_id,
    'APLICADO'
FROM com_compras c
WHERE c.deleted_at IS NULL
  AND c.estado IN ('REGISTRADA', 'PAGADA')
  AND EXISTS (SELECT 1 FROM com_compras_detalle cd WHERE cd.compra_id = c.id)
  AND NOT EXISTS (
      SELECT 1 FROM inv_documentos d
      WHERE d.empresa_id = c.empresa_id
        AND d.documento_origen_tipo = 'COMPRA'
        AND d.documento_origen_id = c.id
  );

-- 3. Insertar las líneas en inv_movimientos_detalle para los documentos recién creados
INSERT INTO inv_movimientos_detalle (
    documento_id,
    articulo_id,
    ubicacion_origen_id,
    ubicacion_destino_id,
    cantidad,
    costo_unitario
)
SELECT
    d.id,
    cd.articulo_id,
    NULL,
    (
        SELECT u.id
        FROM inv_ubicaciones u
        WHERE u.empresa_id = d.empresa_id
          AND u.sucursal_id = d.sucursal_id
          AND u.deleted_at IS NULL
        ORDER BY u.activo DESC, u.id ASC
        LIMIT 1
    ),
    cd.cantidad,
    cd.costo_unitario
FROM inv_documentos d
JOIN com_compras_detalle cd ON cd.compra_id = d.documento_origen_id
WHERE d.documento_origen_tipo = 'COMPRA'
  AND d.consecutivo = 'INV-COMP-' || d.documento_origen_id
  AND NOT EXISTS (
      SELECT 1 FROM inv_movimientos_detalle m WHERE m.documento_id = d.id
  );

-- 4. Materializar/actualizar inv_stock con las líneas insertadas por el backfill
INSERT INTO inv_stock (empresa_id, articulo_id, ubicacion_id, cantidad, costo_promedio, updated_at)
SELECT
    d.empresa_id,
    m.articulo_id,
    m.ubicacion_destino_id,
    SUM(m.cantidad) AS cantidad_total,
    CASE
        WHEN SUM(m.cantidad) > 0 THEN ROUND(SUM(m.cantidad * m.costo_unitario) / SUM(m.cantidad), 4)
        ELSE 0
    END AS costo_prom,
    now()
FROM inv_documentos d
JOIN inv_movimientos_detalle m ON m.documento_id = d.id
WHERE d.documento_origen_tipo = 'COMPRA'
  AND d.consecutivo = 'INV-COMP-' || d.documento_origen_id
  AND m.ubicacion_destino_id IS NOT NULL
GROUP BY d.empresa_id, m.articulo_id, m.ubicacion_destino_id
ON CONFLICT (articulo_id, ubicacion_id) DO UPDATE
SET
    costo_promedio = CASE
        WHEN (inv_stock.cantidad + EXCLUDED.cantidad) > 0
        THEN ROUND(((inv_stock.cantidad * inv_stock.costo_promedio) + (EXCLUDED.cantidad * EXCLUDED.costo_promedio)) / (inv_stock.cantidad + EXCLUDED.cantidad), 4)
        ELSE inv_stock.costo_promedio
    END,
    cantidad = inv_stock.cantidad + EXCLUDED.cantidad,
    updated_at = now();
