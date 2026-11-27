-- =====================================================================
-- 000028 - Módulo de Lealtad: reglas de acumulación (lealtad_reglas)
--   * Reglas globales del programa (empresa_id IS NULL) y overrides por empresa
--   * Semilla de regla estándar para "Lealtad Tesserak" (1 pt por cada $1.000)
--   * Garantiza vinculación de "Lealtad Tesserak" al usuario "ipalenciab"
-- Motor: PostgreSQL 14+
-- =====================================================================

-- Asegurar que el plan global por defecto esté vinculado a ipalenciab si quedó sin asignar
UPDATE lealtad_programas
SET usuario_administrador_id = (
    SELECT id
    FROM usuarios
    WHERE lower(nombre_usuario) = 'ipalenciab'
      AND deleted_at IS NULL
    ORDER BY id ASC
    LIMIT 1
)
WHERE lower(codigo) = 'lealtad_tesserak'
  AND usuario_administrador_id IS NULL
  AND deleted_at IS NULL;

-- ---------------------------------------------------------------------
-- lealtad_reglas
-- ---------------------------------------------------------------------
CREATE TABLE lealtad_reglas (
    id                INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    programa_id       INT            NOT NULL REFERENCES lealtad_programas (id),
    empresa_id        INT            REFERENCES empresas (id),
    nombre            VARCHAR(100)   NOT NULL,
    tipo_regla        VARCHAR(30)    NOT NULL DEFAULT 'POR_MONTO',
    puntos_por_monto  NUMERIC(14,4)  NOT NULL DEFAULT 1,
    monto_base        NUMERIC(14,4)  NOT NULL DEFAULT 1000,
    compra_minima     NUMERIC(14,4)  NOT NULL DEFAULT 0,
    moneda_id         INT            NOT NULL REFERENCES monedas (id),
    prioridad         INT            NOT NULL DEFAULT 100,
    fecha_inicio      TIMESTAMPTZ    NOT NULL DEFAULT now(),
    fecha_fin         TIMESTAMPTZ,
    activo            BOOLEAN        NOT NULL DEFAULT TRUE,
    created_at        TIMESTAMPTZ    NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ    NOT NULL DEFAULT now(),
    deleted_at        TIMESTAMPTZ,

    CONSTRAINT ck_lealtad_reglas_tipo
        CHECK (tipo_regla IN ('POR_MONTO', 'BONO_FIJO', 'MULTIPLICADOR')),
    CONSTRAINT ck_lealtad_reglas_puntos
        CHECK (puntos_por_monto > 0),
    CONSTRAINT ck_lealtad_reglas_monto_base
        CHECK (monto_base > 0),
    CONSTRAINT ck_lealtad_reglas_compra_minima
        CHECK (compra_minima >= 0),
    CONSTRAINT ck_lealtad_reglas_prioridad
        CHECK (prioridad >= 1),
    CONSTRAINT ck_lealtad_reglas_vigencia
        CHECK (fecha_fin IS NULL OR fecha_fin >= fecha_inicio)
);

CREATE INDEX ix_lealtad_reglas_programa
    ON lealtad_reglas (programa_id, prioridad ASC)
    WHERE deleted_at IS NULL;

CREATE INDEX ix_lealtad_reglas_programa_empresa
    ON lealtad_reglas (programa_id, empresa_id, activo)
    WHERE deleted_at IS NULL;

-- ---------------------------------------------------------------------
-- Semilla: Regla base del plan "Lealtad Tesserak" (1 punto por cada $1.000)
-- ---------------------------------------------------------------------
INSERT INTO lealtad_reglas (
    programa_id,
    empresa_id,
    nombre,
    tipo_regla,
    puntos_por_monto,
    monto_base,
    compra_minima,
    moneda_id,
    prioridad,
    activo
)
SELECT
    lp.id,
    NULL,
    'Acumulación estándar Lealtad Tesserak',
    'POR_MONTO',
    1.0000,
    1000.0000,
    0.0000,
    lp.moneda_referencia_id,
    100,
    TRUE
FROM lealtad_programas lp
WHERE lower(lp.codigo) = 'lealtad_tesserak'
  AND lp.deleted_at IS NULL
  AND NOT EXISTS (
      SELECT 1
      FROM lealtad_reglas lr
      WHERE lr.programa_id = lp.id
        AND lr.empresa_id IS NULL
        AND lr.deleted_at IS NULL
  );
