-- =====================================================================
-- 000031 (DOWN) - Revertir campos de redención en ventas y vencimientos
-- =====================================================================

DROP INDEX IF EXISTS ix_lealtad_transacciones_vencimiento_pendiente;

ALTER TABLE lealtad_transacciones
    DROP COLUMN IF EXISTS vencimiento_procesado;

ALTER TABLE ven_ventas
    DROP CONSTRAINT IF EXISTS ck_ven_ventas_descuento_lealtad,
    DROP CONSTRAINT IF EXISTS ck_ven_ventas_puntos_redimidos,
    DROP COLUMN IF EXISTS descuento_lealtad,
    DROP COLUMN IF EXISTS puntos_redimidos;
