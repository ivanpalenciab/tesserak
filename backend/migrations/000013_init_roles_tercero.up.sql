-- =====================================================================
-- 000013 - Catálogo de Roles de Tercero (roles_tercero) con semilla inicial
-- Motor: PostgreSQL 14+
-- =====================================================================

CREATE TABLE roles_tercero (
    id          INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    codigo      VARCHAR(30)  NOT NULL,
    nombre_rol  VARCHAR(80)  NOT NULL,
    descripcion VARCHAR(255),
    activo      BOOLEAN      NOT NULL DEFAULT TRUE,
    es_sistema  BOOLEAN      NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    deleted_at  TIMESTAMPTZ
);

CREATE UNIQUE INDEX ux_roles_tercero_codigo
    ON roles_tercero (upper(codigo))
    WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX ux_roles_tercero_nombre
    ON roles_tercero (lower(nombre_rol))
    WHERE deleted_at IS NULL;

CREATE INDEX ix_roles_tercero_activo
    ON roles_tercero (activo)
    WHERE deleted_at IS NULL;

CREATE INDEX ix_roles_tercero_deleted_at
    ON roles_tercero (deleted_at);

-- ---------------------------------------------------------------------
-- Semilla de roles base del sistema: Cliente, Proveedor y Empleado
-- ---------------------------------------------------------------------
INSERT INTO roles_tercero (codigo, nombre_rol, descripcion, activo, es_sistema)
VALUES
    ('CLIENTE',   'Cliente',   'Persona natural o jurídica que adquiere productos o servicios', TRUE, TRUE),
    ('PROVEEDOR', 'Proveedor', 'Persona natural o jurídica que suministra bienes o servicios',  TRUE, TRUE),
    ('EMPLEADO',  'Empleado',  'Colaborador vinculado laboralmente a la empresa',               TRUE, TRUE)
ON CONFLICT DO NOTHING;
