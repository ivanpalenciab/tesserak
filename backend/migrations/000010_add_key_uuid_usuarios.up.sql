-- Agrega el campo key (UUID v4) a la tabla usuarios.
-- gen_random_uuid() es nativo en PostgreSQL 13+ y genera UUID v4 (RFC 4122).
ALTER TABLE usuarios
    ADD COLUMN IF NOT EXISTS key UUID NOT NULL DEFAULT gen_random_uuid();

CREATE UNIQUE INDEX IF NOT EXISTS ux_usuarios_key
    ON usuarios (key);
