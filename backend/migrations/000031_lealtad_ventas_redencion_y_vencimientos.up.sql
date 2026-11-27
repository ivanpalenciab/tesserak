-- =====================================================================
-- 000031 - Lealtad en Ventas (descuento por redención) + control
--          idempotente de vencimiento de puntos
-- Motor: PostgreSQL 14+
-- =====================================================================

-- 1. Campos de redención de lealtad en ven_ventas
ALTER TABLE ven_ventas
    ADD COLUMN IF NOT EXISTS puntos_redimidos  BIGINT        NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS descuento_lealtad NUMERIC(14,4) NOT NULL DEFAULT 0;

ALTER TABLE ven_ventas
    ADD CONSTRAINT ck_ven_ventas_puntos_redimidos CHECK (puntos_redimidos >= 0),
    ADD CONSTRAINT ck_ven_ventas_descuento_lealtad CHECK (descuento_lealtad >= 0);

-- 2. Bandera idempotente en lealtad_transacciones para procesar vencimientos una sola vez
ALTER TABLE lealtad_transacciones
    ADD COLUMN IF NOT EXISTS vencimiento_procesado BOOLEAN NOT NULL DEFAULT FALSE;

CREATE INDEX IF NOT EXISTS ix_lealtad_transacciones_vencimiento_pendiente
    ON lealtad_transacciones (programa_id, fecha_vencimiento)
    WHERE puntos > 0 AND fecha_vencimiento IS NOT NULL AND vencimiento_procesado = FALSE;
