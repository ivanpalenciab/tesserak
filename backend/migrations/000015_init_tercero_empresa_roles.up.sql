-- =====================================================================
-- 000015 - Asignación de Roles a Terceros (tercero_empresa_roles)
-- Motor: PostgreSQL 14+
-- =====================================================================

CREATE TABLE tercero_empresa_roles (
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tercero_id BIGINT      NOT NULL REFERENCES tercero_maestro (id) ON DELETE CASCADE,
    rol_id     INT         NOT NULL REFERENCES roles_tercero (id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_tercero_empresa_roles_tercero_rol UNIQUE (tercero_id, rol_id)
);

CREATE INDEX ix_tercero_empresa_roles_tercero
    ON tercero_empresa_roles (tercero_id);

CREATE INDEX ix_tercero_empresa_roles_rol
    ON tercero_empresa_roles (rol_id);
