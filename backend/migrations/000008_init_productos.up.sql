-- =====================================================================
-- 000008 - Módulo Productos por empresa (fusión maestro + empresa_productos)
-- Motor: PostgreSQL 14+
-- =====================================================================

-- ---------------------------------------------------------------------
-- unidades_medida (catálogo global estándar)
-- ---------------------------------------------------------------------
CREATE TABLE unidades_medida (
    id          INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    nombre      VARCHAR(50)  NOT NULL UNIQUE,
    descripcion VARCHAR(150)
);

INSERT INTO unidades_medida (nombre, descripcion)
VALUES
    ('UND', 'Unidad'),
    ('KG',  'Kilogramo'),
    ('LT',  'Litro'),
    ('MT',  'Metro'),
    ('CJ',  'Caja')
ON CONFLICT (nombre) DO NOTHING;

-- ---------------------------------------------------------------------
-- categorias_producto (catálogo aislado por empresa, con soft delete)
-- ---------------------------------------------------------------------
CREATE TABLE categorias_producto (
    id          INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    empresa_id  INT          NOT NULL REFERENCES empresas (id),
    nombre      VARCHAR(100) NOT NULL,
    activo      BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    deleted_at  TIMESTAMPTZ,

    CONSTRAINT uq_categorias_producto_id_empresa UNIQUE (id, empresa_id)
);

CREATE UNIQUE INDEX ux_categorias_producto_empresa_nombre
    ON categorias_producto (empresa_id, lower(nombre))
    WHERE deleted_at IS NULL;

-- ---------------------------------------------------------------------
-- marcas (catálogo aislado por empresa, con soft delete)
-- ---------------------------------------------------------------------
CREATE TABLE marcas (
    id          INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    empresa_id  INT          NOT NULL REFERENCES empresas (id),
    nombre      VARCHAR(100) NOT NULL,
    activo      BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    deleted_at  TIMESTAMPTZ,

    CONSTRAINT uq_marcas_id_empresa UNIQUE (id, empresa_id)
);

CREATE UNIQUE INDEX ux_marcas_empresa_nombre
    ON marcas (empresa_id, lower(nombre))
    WHERE deleted_at IS NULL;

-- ---------------------------------------------------------------------
-- grupos_impuesto (catálogo aislado por empresa, con soft delete)
-- ---------------------------------------------------------------------
CREATE TABLE grupos_impuesto (
    id          INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    empresa_id  INT           NOT NULL REFERENCES empresas (id),
    nombre      VARCHAR(100)  NOT NULL,
    tarifa      NUMERIC(5,2)  NOT NULL DEFAULT 0,
    activo      BOOLEAN       NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ   NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ   NOT NULL DEFAULT now(),
    deleted_at  TIMESTAMPTZ,

    CONSTRAINT uq_grupos_impuesto_id_empresa UNIQUE (id, empresa_id),
    CONSTRAINT ck_grupos_impuesto_tarifa CHECK (tarifa >= 0 AND tarifa <= 100)
);

CREATE UNIQUE INDEX ux_grupos_impuesto_empresa_nombre
    ON grupos_impuesto (empresa_id, lower(nombre))
    WHERE deleted_at IS NULL;

-- ---------------------------------------------------------------------
-- productos (aislado por empresa, con FKs compuestas y soft delete)
-- ---------------------------------------------------------------------
CREATE TABLE productos (
    id                INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    empresa_id        INT          NOT NULL REFERENCES empresas (id),
    referencia        VARCHAR(80)  NOT NULL,
    nombre            VARCHAR(150) NOT NULL,
    descripcion       TEXT,
    unidad_medida_id  INT          REFERENCES unidades_medida (id),
    categoria_id      INT,
    marca_id          INT,
    grupo_impuesto_id INT,
    activo            BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at        TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ  NOT NULL DEFAULT now(),
    deleted_at        TIMESTAMPTZ,

    CONSTRAINT uq_productos_id_empresa UNIQUE (id, empresa_id),
    CONSTRAINT fk_productos_categoria_empresa
        FOREIGN KEY (categoria_id, empresa_id) REFERENCES categorias_producto (id, empresa_id),
    CONSTRAINT fk_productos_marca_empresa
        FOREIGN KEY (marca_id, empresa_id) REFERENCES marcas (id, empresa_id),
    CONSTRAINT fk_productos_impuesto_empresa
        FOREIGN KEY (grupo_impuesto_id, empresa_id) REFERENCES grupos_impuesto (id, empresa_id)
);

-- Referencia/SKU única por empresa entre productos no eliminados.
CREATE UNIQUE INDEX ux_productos_empresa_referencia
    ON productos (empresa_id, lower(referencia))
    WHERE deleted_at IS NULL;

CREATE INDEX ix_productos_empresa_activo
    ON productos (empresa_id, activo)
    WHERE deleted_at IS NULL;

CREATE INDEX ix_productos_empresa_categoria
    ON productos (empresa_id, categoria_id)
    WHERE deleted_at IS NULL;

CREATE INDEX ix_productos_deleted_at
    ON productos (deleted_at);
