-- =====================================================================
-- 000014 - Tabla maestra de Terceros (tercero_maestro) con FK a empresas
-- Motor: PostgreSQL 14+
-- =====================================================================

CREATE TABLE tercero_maestro (
    id                    BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    empresa_id            INT          NOT NULL REFERENCES empresas (id),
    tipo_identificacion   VARCHAR(30)  NOT NULL,
    numero_identificacion VARCHAR(50)  NOT NULL,
    razon_social          VARCHAR(200) NOT NULL,
    pais_id               INT          REFERENCES paises (id),
    naturaleza_juridica   VARCHAR(30)  NOT NULL,
    email                 VARCHAR(255),
    telefono              VARCHAR(50),
    direccion             VARCHAR(255),
    activo                BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at            TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at            TIMESTAMPTZ  NOT NULL DEFAULT now(),
    deleted_at            TIMESTAMPTZ,

    CONSTRAINT uq_tercero_maestro_id_empresa UNIQUE (id, empresa_id),
    CONSTRAINT ck_tercero_maestro_tipo_identificacion
        CHECK (tipo_identificacion IN ('NIT', 'CC', 'CE', 'PASAPORTE', 'RUT', 'DNI', 'OTRO')),
    CONSTRAINT ck_tercero_maestro_naturaleza_juridica
        CHECK (naturaleza_juridica IN ('NATURAL', 'JURIDICA'))
);

CREATE UNIQUE INDEX ux_tercero_maestro_empresa_identificacion
    ON tercero_maestro (empresa_id, upper(tipo_identificacion), lower(numero_identificacion))
    WHERE deleted_at IS NULL;

CREATE INDEX ix_tercero_maestro_empresa_activo
    ON tercero_maestro (empresa_id, activo)
    WHERE deleted_at IS NULL;

CREATE INDEX ix_tercero_maestro_empresa_naturaleza
    ON tercero_maestro (empresa_id, naturaleza_juridica)
    WHERE deleted_at IS NULL;

CREATE INDEX ix_tercero_maestro_deleted_at
    ON tercero_maestro (deleted_at);
