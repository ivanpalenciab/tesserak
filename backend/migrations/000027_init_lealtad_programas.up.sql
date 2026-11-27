-- =====================================================================
-- 000027 - Módulo de Lealtad: planes/programas y adhesión de empresas
--   * Crea lealtad_programas (LOCAL, COALICION, GLOBAL)
--   * Crea lealtad_programa_empresas (relación N:M de adhesión)
--   * Siembra el plan por defecto "Lealtad Tesserak" (GLOBAL, es_sistema)
--     administrado por el usuario existente "ipalenciab"
-- Motor: PostgreSQL 14+
-- =====================================================================

-- Garantizar al menos una moneda base para referenciar en el plan por defecto
INSERT INTO monedas (codigo_iso, nombre_moneda, simbolo)
VALUES ('COP', 'Peso Colombiano', '$')
ON CONFLICT (codigo_iso) DO NOTHING;

-- ---------------------------------------------------------------------
-- lealtad_programas
-- ---------------------------------------------------------------------
CREATE TABLE lealtad_programas (
    id                        INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    codigo                    VARCHAR(30)    NOT NULL,
    nombre                    VARCHAR(100)   NOT NULL,
    descripcion               TEXT,
    alcance                   VARCHAR(20)    NOT NULL DEFAULT 'LOCAL',
    empresa_propietaria_id    INT            REFERENCES empresas (id),
    usuario_administrador_id  BIGINT         REFERENCES usuarios (id),
    es_sistema                BOOLEAN        NOT NULL DEFAULT FALSE,
    moneda_referencia_id      INT            NOT NULL REFERENCES monedas (id),
    valor_punto_redencion     NUMERIC(14,4)  NOT NULL DEFAULT 1,
    dias_vencimiento          INT,
    estado                    VARCHAR(20)    NOT NULL DEFAULT 'ACTIVO',
    fecha_inicio              TIMESTAMPTZ    NOT NULL DEFAULT now(),
    fecha_fin                 TIMESTAMPTZ,
    created_at                TIMESTAMPTZ    NOT NULL DEFAULT now(),
    updated_at                TIMESTAMPTZ    NOT NULL DEFAULT now(),
    deleted_at                TIMESTAMPTZ,

    CONSTRAINT ck_lealtad_programas_alcance
        CHECK (alcance IN ('LOCAL', 'COALICION', 'GLOBAL')),
    CONSTRAINT ck_lealtad_programas_alcance_empresa
        CHECK (
            (alcance = 'GLOBAL' AND empresa_propietaria_id IS NULL)
            OR (alcance IN ('LOCAL', 'COALICION') AND empresa_propietaria_id IS NOT NULL)
        ),
    CONSTRAINT ck_lealtad_programas_valor_punto
        CHECK (valor_punto_redencion > 0),
    CONSTRAINT ck_lealtad_programas_dias_vencimiento
        CHECK (dias_vencimiento IS NULL OR dias_vencimiento > 0),
    CONSTRAINT ck_lealtad_programas_estado
        CHECK (estado IN ('ACTIVO', 'INACTIVO', 'PAUSADO')),
    CONSTRAINT ck_lealtad_programas_vigencia
        CHECK (fecha_fin IS NULL OR fecha_fin >= fecha_inicio)
);

CREATE UNIQUE INDEX ux_lealtad_programas_codigo
    ON lealtad_programas (lower(codigo))
    WHERE deleted_at IS NULL;

CREATE INDEX ix_lealtad_programas_empresa
    ON lealtad_programas (empresa_propietaria_id)
    WHERE deleted_at IS NULL;

CREATE INDEX ix_lealtad_programas_alcance_estado
    ON lealtad_programas (alcance, estado)
    WHERE deleted_at IS NULL;

-- ---------------------------------------------------------------------
-- lealtad_programa_empresas (adhesión N:M de empresas a planes)
-- ---------------------------------------------------------------------
CREATE TABLE lealtad_programa_empresas (
    id                     BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    programa_id            INT          NOT NULL REFERENCES lealtad_programas (id),
    empresa_id             INT          NOT NULL REFERENCES empresas (id),
    rol_empresa            VARCHAR(20)  NOT NULL DEFAULT 'ALIADA',
    permite_acumular       BOOLEAN      NOT NULL DEFAULT TRUE,
    permite_redimir        BOOLEAN      NOT NULL DEFAULT TRUE,
    es_programa_principal  BOOLEAN      NOT NULL DEFAULT FALSE,
    fecha_adhesion         TIMESTAMPTZ  NOT NULL DEFAULT now(),
    fecha_baja             TIMESTAMPTZ,
    activo                 BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at             TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at             TIMESTAMPTZ  NOT NULL DEFAULT now(),

    CONSTRAINT uq_lealtad_programa_empresas
        UNIQUE (programa_id, empresa_id),
    CONSTRAINT ck_lealtad_programa_empresas_rol
        CHECK (rol_empresa IN ('PROPIETARIA', 'ALIADA'))
);

CREATE UNIQUE INDEX ux_lealtad_empresa_principal
    ON lealtad_programa_empresas (empresa_id)
    WHERE es_programa_principal = TRUE AND activo = TRUE;

CREATE INDEX ix_lealtad_programa_empresas_empresa
    ON lealtad_programa_empresas (empresa_id)
    WHERE activo = TRUE;

-- ---------------------------------------------------------------------
-- Semilla: Plan de lealtad por defecto "Lealtad Tesserak"
--   * Administrado por el usuario existente "ipalenciab"
-- ---------------------------------------------------------------------
INSERT INTO lealtad_programas (
    codigo,
    nombre,
    descripcion,
    alcance,
    empresa_propietaria_id,
    usuario_administrador_id,
    es_sistema,
    moneda_referencia_id,
    valor_punto_redencion,
    dias_vencimiento,
    estado
)
SELECT
    'LEALTAD_TESSERAK',
    'Lealtad Tesserak',
    'Plan de lealtad global por defecto de Tesserak al que las empresas pueden adherirse.',
    'GLOBAL',
    NULL,
    (
        SELECT id
        FROM usuarios
        WHERE lower(nombre_usuario) = 'ipalenciab'
          AND deleted_at IS NULL
        ORDER BY id ASC
        LIMIT 1
    ),
    TRUE,
    (
        SELECT id
        FROM monedas
        ORDER BY CASE WHEN codigo_iso = 'COP' THEN 0 ELSE 1 END, id ASC
        LIMIT 1
    ),
    1.0000,
    365,
    'ACTIVO'
WHERE NOT EXISTS (
    SELECT 1
    FROM lealtad_programas
    WHERE lower(codigo) = 'lealtad_tesserak'
      AND deleted_at IS NULL
);
