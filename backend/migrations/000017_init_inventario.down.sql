-- =====================================================================
-- 000017 - Rollback de Módulo de Inventarios (prefijo inv_)
-- =====================================================================

DROP TABLE IF EXISTS inv_stock;
DROP TABLE IF EXISTS inv_movimientos_detalle;
DROP TABLE IF EXISTS inv_documentos;
DROP TABLE IF EXISTS inv_tipos_movimiento;
DROP TABLE IF EXISTS inv_ubicaciones;
DELETE FROM tipos_documento WHERE codigo = 'INVENTARIO';
