-- =====================================================================
-- 000004 - Agrega columna opcional usuario_facebook a crm_leads
-- Nota: en instalaciones limpias la columna ya nace en 000002.
-- =====================================================================

ALTER TABLE crm_leads
    ADD COLUMN IF NOT EXISTS usuario_facebook VARCHAR(100);
