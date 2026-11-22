-- =====================================================================
-- 000001 - Esquema base: geografía, empresas, sucursales y usuarios
-- Motor: PostgreSQL 14+
-- =====================================================================

-- ---------------------------------------------------------------------
-- paises
-- ---------------------------------------------------------------------
CREATE TABLE paises (
    id               INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    iso_alpha_2      CHAR(2)      NOT NULL UNIQUE,
    iso_alpha_3      CHAR(3)      NOT NULL UNIQUE,
    codigo_numerico  INT          NOT NULL UNIQUE,
    nombre_oficial   VARCHAR(150) NOT NULL
);

-- ---------------------------------------------------------------------
-- ciudades
-- ---------------------------------------------------------------------
CREATE TABLE ciudades (
    id       INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    pais_id  INT          NOT NULL REFERENCES paises (id),
    nombre   VARCHAR(100) NOT NULL,
    CONSTRAINT uq_ciudades_pais_nombre UNIQUE (pais_id, nombre)
);

-- ---------------------------------------------------------------------
-- empresas
-- ---------------------------------------------------------------------
CREATE TABLE empresas (
    id                    INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    nombre_legal          VARCHAR(255) NOT NULL,
    razon_social          VARCHAR(255) NOT NULL,
    identificador_fiscal  VARCHAR(50)  NOT NULL,
    email                 VARCHAR(255),
    direccion             VARCHAR(255),
    logo                  VARCHAR(500),
    activo                BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at            TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at            TIMESTAMPTZ  NOT NULL DEFAULT now(),
    deleted_at            TIMESTAMPTZ
);

CREATE UNIQUE INDEX ux_empresas_identificador_fiscal
    ON empresas (identificador_fiscal)
    WHERE deleted_at IS NULL;

-- ---------------------------------------------------------------------
-- sucursales
-- ---------------------------------------------------------------------
CREATE TABLE sucursales (
    id          INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    empresa_id  INT          NOT NULL REFERENCES empresas (id),
    ciudad_id   INT          REFERENCES ciudades (id),
    nombre      VARCHAR(150) NOT NULL,
    direccion   VARCHAR(255),
    activo      BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    deleted_at  TIMESTAMPTZ,
    -- Permite que otras tablas referencien (sucursal_id, empresa_id) y así
    -- garantizar que la sucursal pertenece a la empresa indicada.
    CONSTRAINT uq_sucursales_id_empresa UNIQUE (id, empresa_id)
);

CREATE INDEX ix_sucursales_empresa ON sucursales (empresa_id);

-- ---------------------------------------------------------------------
-- usuarios
--   * Cada usuario pertenece a UNA sucursal (y por ende a una empresa).
--   * empresa_id se guarda desnormalizado para poder aplicar la regla
--     "email único por empresa"; la FK compuesta evita inconsistencias.
--   * Soft delete mediante deleted_at.
-- ---------------------------------------------------------------------
CREATE TABLE usuarios (
    id               BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    empresa_id       INT          NOT NULL,
    sucursal_id      INT          NOT NULL,
    nombre_usuario   VARCHAR(50)  NOT NULL,
    nombres          VARCHAR(100) NOT NULL,
    apellidos        VARCHAR(100) NOT NULL,
    email            VARCHAR(255) NOT NULL,
    password_hash    VARCHAR(255) NOT NULL,
    activo           BOOLEAN      NOT NULL DEFAULT TRUE,
    ultimo_login_at  TIMESTAMPTZ,
    created_at       TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ  NOT NULL DEFAULT now(),
    deleted_at       TIMESTAMPTZ,

    CONSTRAINT fk_usuarios_empresa
        FOREIGN KEY (empresa_id) REFERENCES empresas (id),
    CONSTRAINT fk_usuarios_sucursal_empresa
        FOREIGN KEY (sucursal_id, empresa_id) REFERENCES sucursales (id, empresa_id)
);

-- Email único por empresa (sin distinguir mayúsculas) solo entre usuarios no borrados.
CREATE UNIQUE INDEX ux_usuarios_empresa_email
    ON usuarios (empresa_id, lower(email))
    WHERE deleted_at IS NULL;

-- Nombre de usuario único por empresa solo entre usuarios no borrados.
CREATE UNIQUE INDEX ux_usuarios_empresa_nombre_usuario
    ON usuarios (empresa_id, lower(nombre_usuario))
    WHERE deleted_at IS NULL;

CREATE INDEX ix_usuarios_sucursal ON usuarios (sucursal_id) WHERE deleted_at IS NULL;
CREATE INDEX ix_usuarios_deleted_at ON usuarios (deleted_at);
