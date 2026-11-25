DROP INDEX IF EXISTS ux_usuarios_key;

ALTER TABLE usuarios
    DROP COLUMN IF EXISTS key;
