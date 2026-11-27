-- =====================================================================
-- 000020 (down) - Revierte usuario_sucursales y restaura usuarios.sucursal_id
-- =====================================================================

-- 1. Restaurar columna sucursal_id en usuarios (temporalmente nullable)
ALTER TABLE usuarios ADD COLUMN sucursal_id INT;

-- 2. Restaurar sucursal principal (o primera disponible) desde usuario_sucursales
UPDATE usuarios u
SET sucursal_id = sub.sucursal_id
FROM (
    SELECT DISTINCT ON (usuario_id)
        usuario_id,
        sucursal_id
    FROM usuario_sucursales
    ORDER BY usuario_id, es_principal DESC, sucursal_id ASC
) sub
WHERE u.id = sub.usuario_id;

-- 3. Fallback: si algún usuario quedó sin sucursal asignada, tomar la primera de su empresa
UPDATE usuarios u
SET sucursal_id = sub.id
FROM (
    SELECT DISTINCT ON (empresa_id)
        empresa_id,
        id
    FROM sucursales
    ORDER BY empresa_id, id ASC
) sub
WHERE u.sucursal_id IS NULL
  AND u.empresa_id = sub.empresa_id;

-- 4. Restaurar NOT NULL, FK compuesta e índice
ALTER TABLE usuarios ALTER COLUMN sucursal_id SET NOT NULL;

ALTER TABLE usuarios
    ADD CONSTRAINT fk_usuarios_sucursal_empresa
        FOREIGN KEY (sucursal_id, empresa_id) REFERENCES sucursales (id, empresa_id);

CREATE INDEX ix_usuarios_sucursal ON usuarios (sucursal_id) WHERE deleted_at IS NULL;

-- 5. Eliminar tabla intermedia y restricción única compuesta en usuarios
DROP TABLE IF EXISTS usuario_sucursales;
ALTER TABLE usuarios DROP CONSTRAINT IF EXISTS uq_usuarios_id_empresa;
