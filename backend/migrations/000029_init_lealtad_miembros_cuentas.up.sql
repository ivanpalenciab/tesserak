-- =====================================================================
-- 000029 - Módulo de Lealtad: miembros globales y cuentas de puntos
--   * lealtad_miembros: identidad única por documento (independiente de empresa_id)
--   * lealtad_cuentas: billetera de puntos por (programa_id, miembro_id)
-- Motor: PostgreSQL 14+
-- =====================================================================

-- ---------------------------------------------------------------------
-- lealtad_miembros (identidad global desacoplada de empresa_id)
-- ---------------------------------------------------------------------
CREATE TABLE lealtad_miembros (
    id                    BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    pais_id               INT          NOT NULL REFERENCES paises (id),
    tipo_identificacion   VARCHAR(30)  NOT NULL,
    numero_identificacion VARCHAR(50)  NOT NULL,
    nombre_completo       VARCHAR(200) NOT NULL,
    email                 VARCHAR(255),
    telefono              VARCHAR(50),
    created_at            TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at            TIMESTAMPTZ  NOT NULL DEFAULT now(),

    CONSTRAINT ck_lealtad_miembros_tipo_identificacion
        CHECK (tipo_identificacion IN ('NIT', 'CC', 'CE', 'PASAPORTE', 'RUT', 'DNI', 'OTRO'))
);

CREATE UNIQUE INDEX ux_lealtad_miembros_documento
    ON lealtad_miembros (pais_id, upper(tipo_identificacion), lower(numero_identificacion));

CREATE INDEX ix_lealtad_miembros_nombre
    ON lealtad_miembros (lower(nombre_completo));

-- ---------------------------------------------------------------------
-- lealtad_cuentas (billetera de puntos del miembro dentro de un programa)
-- ---------------------------------------------------------------------
CREATE TABLE lealtad_cuentas (
    id            BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    programa_id   INT          NOT NULL REFERENCES lealtad_programas (id),
    miembro_id    BIGINT       NOT NULL REFERENCES lealtad_miembros (id),
    saldo_puntos  BIGINT       NOT NULL DEFAULT 0,
    estado        VARCHAR(20)  NOT NULL DEFAULT 'ACTIVA',
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT now(),

    CONSTRAINT uq_lealtad_cuentas_programa_miembro
        UNIQUE (programa_id, miembro_id),
    CONSTRAINT ck_lealtad_cuentas_saldo
        CHECK (saldo_puntos >= 0),
    CONSTRAINT ck_lealtad_cuentas_estado
        CHECK (estado IN ('ACTIVA', 'BLOQUEADA', 'CERRADA'))
);

CREATE INDEX ix_lealtad_cuentas_miembro
    ON lealtad_cuentas (miembro_id);

CREATE INDEX ix_lealtad_cuentas_programa_estado
    ON lealtad_cuentas (programa_id, estado);
