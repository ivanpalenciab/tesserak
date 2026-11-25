-- =====================================================================
-- 000011 - Estructura de Listas de Precios (listas_precios y lista_precios_detalle)
-- Motor: PostgreSQL 14+
-- =====================================================================

-- ---------------------------------------------------------------------
-- Cabecera: listas_precios
-- Define los diferentes esquemas comerciales de precios por empresa.
-- ---------------------------------------------------------------------
CREATE TABLE listas_precios (
    id             INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    empresa_id     INT          NOT NULL REFERENCES empresas (id),
    nombre         VARCHAR(150) NOT NULL,
    moneda_id      INT          NOT NULL REFERENCES monedas (id),
    tipo_lista     VARCHAR(30)  NOT NULL,
    vigencia_desde DATE         NOT NULL,
    vigencia_hasta DATE,
    estado         BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ  NOT NULL DEFAULT now(),
    deleted_at     TIMESTAMPTZ,

    CONSTRAINT uq_listas_precios_id_empresa UNIQUE (id, empresa_id),
    CONSTRAINT ck_listas_precios_tipo
        CHECK (tipo_lista IN ('GENERAL', 'MAYORISTA', 'MINORISTA', 'ESPECIAL')),
    CONSTRAINT ck_listas_precios_vigencia
        CHECK (vigencia_hasta IS NULL OR vigencia_hasta >= vigencia_desde)
);

CREATE UNIQUE INDEX ux_listas_precios_empresa_nombre
    ON listas_precios (empresa_id, lower(nombre))
    WHERE deleted_at IS NULL;

CREATE INDEX ix_listas_precios_empresa_estado
    ON listas_precios (empresa_id, estado)
    WHERE deleted_at IS NULL;

CREATE INDEX ix_listas_precios_empresa_tipo
    ON listas_precios (empresa_id, tipo_lista)
    WHERE deleted_at IS NULL;

CREATE INDEX ix_listas_precios_deleted_at
    ON listas_precios (deleted_at);

-- ---------------------------------------------------------------------
-- Detalle: lista_precios_detalle
-- Contiene los precios específicos de cada artículo dentro de una lista.
-- ---------------------------------------------------------------------
CREATE TABLE lista_precios_detalle (
    id               BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    lista_precio_id  INT           NOT NULL REFERENCES listas_precios (id) ON DELETE CASCADE,
    articulo_id      INT           NOT NULL REFERENCES articulos (id),
    precio           NUMERIC(14,2) NOT NULL,
    unidad_medida_id INT           NOT NULL REFERENCES unidades_medida (id),
    cantidad_minima  NUMERIC(14,4) NOT NULL DEFAULT 1,
    vigencia_desde   DATE,
    vigencia_hasta   DATE,
    created_at       TIMESTAMPTZ   NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ   NOT NULL DEFAULT now(),

    CONSTRAINT ck_lista_precios_detalle_precio
        CHECK (precio >= 0),
    CONSTRAINT ck_lista_precios_detalle_cantidad_minima
        CHECK (cantidad_minima > 0),
    CONSTRAINT ck_lista_precios_detalle_vigencia
        CHECK (vigencia_desde IS NULL OR vigencia_hasta IS NULL OR vigencia_hasta >= vigencia_desde)
);

CREATE INDEX ix_lista_precios_detalle_lista
    ON lista_precios_detalle (lista_precio_id);

CREATE INDEX ix_lista_precios_detalle_articulo
    ON lista_precios_detalle (articulo_id);

CREATE INDEX ix_lista_precios_detalle_lookup
    ON lista_precios_detalle (lista_precio_id, articulo_id, unidad_medida_id, cantidad_minima);
