-- =====================================================================
-- 000020 - Relación N:M entre usuarios y sucursales (usuario_sucursales)
--   * Se mantiene la FK directa empresas <- sucursales (sucursales.empresa_id)
--     y empresas <- usuarios (usuarios.empresa_id).
--   * Se crea la tabla intermedia usuario_sucursales con FKs compuestas
--     (usuario_id, empresa_id) y (sucursal_id, empresa_id) para garantizar
--     aislamiento multi-tenant a nivel de motor.
--   * Se migran las asignaciones actuales de usuarios.sucursal_id antes de
--     eliminar la FK y la columna usuarios.sucursal_id.
-- Motor: PostgreSQL 14+
-- =====================================================================

-- 1. Restricción única compuesta en usuarios para habilitar FK compuesta
--    desde usuario_sucursales (usuario_id, empresa_id).
ALTER TABLE usuarios
    ADD CONSTRAINT uq_usuarios_id_empresa UNIQUE (id, empresa_id);

-- 2. Tabla intermedia usuario_sucursales (N:M entre usuarios y sucursales)
CREATE TABLE usuario_sucursales (
    usuario_id   BIGINT      NOT NULL,
    sucursal_id  INT         NOT NULL,
    empresa_id   INT         NOT NULL,
    es_principal BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),

    PRIMARY KEY (usuario_id, sucursal_id),
    CONSTRAINT fk_usuario_sucursales_empresa
        FOREIGN KEY (empresa_id) REFERENCES empresas (id),
    CONSTRAINT fk_usuario_sucursales_usuario_empresa
        FOREIGN KEY (usuario_id, empresa_id) REFERENCES usuarios (id, empresa_id) ON DELETE CASCADE,
    CONSTRAINT fk_usuario_sucursales_sucursal_empresa
        FOREIGN KEY (sucursal_id, empresa_id) REFERENCES sucursales (id, empresa_id) ON DELETE CASCADE
);

-- Garantiza como máximo una sucursal principal por usuario.
CREATE UNIQUE INDEX ux_usuario_sucursales_principal
    ON usuario_sucursales (usuario_id)
    WHERE es_principal = TRUE;

CREATE INDEX ix_usuario_sucursales_sucursal ON usuario_sucursales (sucursal_id);
CREATE INDEX ix_usuario_sucursales_empresa  ON usuario_sucursales (empresa_id);

-- 3. Backfill: migrar las asignaciones actuales de usuarios.sucursal_id
--    marcándolas como sucursal principal para no perder información.
INSERT INTO usuario_sucursales (usuario_id, sucursal_id, empresa_id, es_principal)
SELECT id, sucursal_id, empresa_id, TRUE
FROM usuarios
WHERE sucursal_id IS NOT NULL
ON CONFLICT (usuario_id, sucursal_id) DO NOTHING;

-- 4. Eliminar índice, llave foránea y columna sucursal_id en usuarios
DROP INDEX IF EXISTS ix_usuarios_sucursal;
ALTER TABLE usuarios DROP CONSTRAINT IF EXISTS fk_usuarios_sucursal_empresa;
ALTER TABLE usuarios DROP COLUMN sucursal_id;
