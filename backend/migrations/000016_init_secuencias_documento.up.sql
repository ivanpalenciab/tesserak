-- =====================================================================
-- 000016 - Catálogo de tipos de documento y secuencias de consecutivos
-- Motor: PostgreSQL 14+
-- =====================================================================

-- ---------------------------------------------------------------------
-- tipos_documento
-- Catálogo global de tipos de documento que generan consecutivos
-- (ej. COMPRA, VENTA, COTIZACION).
-- ---------------------------------------------------------------------
CREATE TABLE tipos_documento (
    id               INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    codigo           VARCHAR(20)  NOT NULL,
    nombre           VARCHAR(100) NOT NULL,
    descripcion      VARCHAR(255),
    prefijo_defecto  VARCHAR(10)  NOT NULL,
    activo           BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at       TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ  NOT NULL DEFAULT now(),

    CONSTRAINT uq_tipos_documento_codigo UNIQUE (codigo)
);

-- Semilla inicial: Compras (COM) y Ventas (VEN)
INSERT INTO tipos_documento (codigo, nombre, descripcion, prefijo_defecto, activo)
VALUES
    ('COMPRA', 'Compras', 'Documentos de compras a proveedores', 'COM', TRUE),
    ('VENTA',  'Ventas',  'Documentos de ventas a clientes',     'VEN', TRUE)
ON CONFLICT (codigo) DO NOTHING;

-- ---------------------------------------------------------------------
-- secuencias_documento
-- Control de consecutivos por empresa y tipo de documento.
-- Cada empresa configura su propio prefijo, relleno y frecuencia de reinicio.
-- ---------------------------------------------------------------------
CREATE TABLE secuencias_documento (
    id                   BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    empresa_id           INT         NOT NULL REFERENCES empresas (id),
    tipo_documento_id    INT         NOT NULL REFERENCES tipos_documento (id),
    prefijo              VARCHAR(10) NOT NULL,
    siguiente_numero     BIGINT      NOT NULL DEFAULT 1,
    longitud_relleno     SMALLINT    NOT NULL DEFAULT 6,
    frecuencia_reinicio  VARCHAR(20) NOT NULL DEFAULT 'NUNCA',
    clave_periodo        VARCHAR(20),
    activo               BOOLEAN     NOT NULL DEFAULT TRUE,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_secuencias_documento_empresa_tipo
        UNIQUE (empresa_id, tipo_documento_id),
    CONSTRAINT ck_secuencias_documento_siguiente
        CHECK (siguiente_numero >= 1),
    CONSTRAINT ck_secuencias_documento_relleno
        CHECK (longitud_relleno >= 1 AND longitud_relleno <= 20),
    CONSTRAINT ck_secuencias_documento_frecuencia
        CHECK (frecuencia_reinicio IN ('NUNCA', 'ANUAL', 'MENSUAL'))
);

CREATE INDEX ix_secuencias_documento_empresa
    ON secuencias_documento (empresa_id);

-- Semilla automática para las empresas existentes (incluyendo la empresa del usuario creado):
--   COMPRA -> prefijo 'COM'
--   VENTA  -> prefijo 'VEN'
INSERT INTO secuencias_documento (
    empresa_id,
    tipo_documento_id,
    prefijo,
    siguiente_numero,
    longitud_relleno,
    frecuencia_reinicio,
    clave_periodo,
    activo
)
SELECT
    e.id,
    td.id,
    td.prefijo_defecto,
    1,
    6,
    'NUNCA',
    NULL,
    TRUE
FROM empresas e
CROSS JOIN tipos_documento td
ON CONFLICT (empresa_id, tipo_documento_id) DO NOTHING;
