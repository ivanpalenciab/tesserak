-- =====================================================================
-- 000003 - Agrega columna nombre a paises (idempotente para BDs existentes)
-- Nota: en instalaciones limpias la columna ya nace despues de id en 000001.
-- =====================================================================

ALTER TABLE paises
    ADD COLUMN IF NOT EXISTS nombre VARCHAR(100);

UPDATE paises
SET nombre = nombre_oficial
WHERE nombre IS NULL;

ALTER TABLE paises
    ALTER COLUMN nombre SET NOT NULL;
