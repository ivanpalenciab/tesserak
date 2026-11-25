-- Rollback de 000009: restaura tabla única productos desde articulos
DROP TABLE IF EXISTS servicios;
DROP TABLE IF EXISTS productos;

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

INSERT INTO productos (
    id, empresa_id, referencia, nombre, descripcion,
    unidad_medida_id, categoria_id, marca_id, grupo_impuesto_id,
    activo, created_at, updated_at, deleted_at
)
OVERRIDING SYSTEM VALUE
SELECT
    a.id, a.empresa_id, a.codigo, a.nombre, a.descripcion,
    a.unidad_medida_id, a.categoria_id, a.marca_id, a.grupo_impuesto_id,
    a.estado, a.created_at, a.updated_at, a.deleted_at
FROM articulos a;

SELECT setval(
    pg_get_serial_sequence('productos', 'id'),
    COALESCE((SELECT MAX(id) FROM productos), 1),
    EXISTS (SELECT 1 FROM productos)
);

DROP TABLE IF EXISTS articulos;
DROP TABLE IF EXISTS modalidades_servicio;
DROP TABLE IF EXISTS metodos_valoracion;
DROP TABLE IF EXISTS tipos_articulo;
