-- =====================================================================
-- 000018 - Rollback de backfill de compras a inventario
-- =====================================================================

-- Nota: al revertir el backfill se eliminan únicamente los documentos generados con prefijo INV-COMP-
DELETE FROM inv_documentos
WHERE documento_origen_tipo = 'COMPRA'
  AND consecutivo LIKE 'INV-COMP-%';
