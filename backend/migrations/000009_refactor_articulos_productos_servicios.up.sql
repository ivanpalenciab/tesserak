-- =====================================================================
-- 000009 - Refactorización Supertipo-Subtipo: articulos -> productos / servicios
-- Motor: PostgreSQL 14+
-- =====================================================================

-- ---------------------------------------------------------------------
-- Catálogos globales con valores iniciales
-- ---------------------------------------------------------------------
CREATE TABLE tipos_articulo (
    id      INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    codigo  VARCHAR(30) NOT NULL UNIQUE,
    nombre  VARCHAR(60) NOT NULL
);

INSERT INTO tipos_articulo (codigo, nombre)
VALUES
    ('PRODUCTO', 'Producto'),
    ('SERVICIO', 'Servicio')
ON CONFLICT (codigo) DO NOTHING;

CREATE TABLE metodos_valoracion (
    id      INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    codigo  VARCHAR(40)  NOT NULL UNIQUE,
    nombre  VARCHAR(100) NOT NULL
);

INSERT INTO metodos_valoracion (codigo, nombre)
VALUES
    ('FIFO',                      'FIFO (Primero en entrar, primero en salir)'),
    ('PROMEDIO_PONDERADO',        'Costo promedio ponderado'),
    ('IDENTIFICACION_ESPECIFICA', 'Identificación específica')
ON CONFLICT (codigo) DO NOTHING;

CREATE TABLE modalidades_servicio (
    id      INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    codigo  VARCHAR(40)  NOT NULL UNIQUE,
    nombre  VARCHAR(100) NOT NULL
);

INSERT INTO modalidades_servicio (codigo, nombre)
VALUES
    ('POR_HORA',    'Por hora'),
    ('PROYECTO',    'Por proyecto'),
    ('UNIDAD',      'Por unidad'),
    ('SUSCRIPCION', 'Suscripción')
ON CONFLICT (codigo) DO NOTHING;

-- ---------------------------------------------------------------------
-- Asegura un grupo de impuesto base ('Exento 0%') por empresa existente
-- ---------------------------------------------------------------------
INSERT INTO grupos_impuesto (empresa_id, nombre, tarifa, activo)
SELECT e.id, 'Exento (0%)', 0.00, TRUE
FROM empresas e
WHERE e.deleted_at IS NULL
  AND NOT EXISTS (
      SELECT 1 FROM grupos_impuesto gi
      WHERE gi.empresa_id = e.id AND gi.deleted_at IS NULL
  )
ON CONFLICT DO NOTHING;

-- Si hay productos legados sin categoría, crea categoría 'General' para su empresa
INSERT INTO categorias_producto (empresa_id, nombre, activo)
SELECT DISTINCT p.empresa_id, 'General', TRUE
FROM productos p
WHERE p.categoria_id IS NULL
  AND NOT EXISTS (
      SELECT 1 FROM categorias_producto c
      WHERE c.empresa_id = p.empresa_id AND c.deleted_at IS NULL
  )
ON CONFLICT DO NOTHING;

-- ---------------------------------------------------------------------
-- Tabla supertipo: articulos (común a productos y servicios, por empresa)
-- ---------------------------------------------------------------------
CREATE TABLE articulos (
    id                INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    empresa_id        INT          NOT NULL REFERENCES empresas (id),
    codigo            VARCHAR(80)  NOT NULL,
    nombre            VARCHAR(150) NOT NULL,
    descripcion       TEXT,
    tipo_articulo_id  INT          NOT NULL REFERENCES tipos_articulo (id),
    unidad_medida_id  INT          NOT NULL REFERENCES unidades_medida (id),
    categoria_id      INT          NOT NULL,
    marca_id          INT,
    grupo_impuesto_id INT          NOT NULL,
    estado            BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at        TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ  NOT NULL DEFAULT now(),
    deleted_at        TIMESTAMPTZ,

    CONSTRAINT uq_articulos_id_empresa UNIQUE (id, empresa_id),
    CONSTRAINT fk_articulos_categoria_empresa
        FOREIGN KEY (categoria_id, empresa_id) REFERENCES categorias_producto (id, empresa_id),
    CONSTRAINT fk_articulos_marca_empresa
        FOREIGN KEY (marca_id, empresa_id) REFERENCES marcas (id, empresa_id),
    CONSTRAINT fk_articulos_impuesto_empresa
        FOREIGN KEY (grupo_impuesto_id, empresa_id) REFERENCES grupos_impuesto (id, empresa_id)
);

CREATE UNIQUE INDEX ux_articulos_empresa_codigo
    ON articulos (empresa_id, lower(codigo))
    WHERE deleted_at IS NULL;

CREATE INDEX ix_articulos_empresa_tipo
    ON articulos (empresa_id, tipo_articulo_id)
    WHERE deleted_at IS NULL;

CREATE INDEX ix_articulos_empresa_estado
    ON articulos (empresa_id, estado)
    WHERE deleted_at IS NULL;

CREATE INDEX ix_articulos_deleted_at
    ON articulos (deleted_at);

-- ---------------------------------------------------------------------
-- Migración de registros existentes de la antigua tabla productos -> articulos
-- ---------------------------------------------------------------------
INSERT INTO articulos (
    id, empresa_id, codigo, nombre, descripcion,
    tipo_articulo_id, unidad_medida_id, categoria_id, marca_id, grupo_impuesto_id,
    estado, created_at, updated_at, deleted_at
)
OVERRIDING SYSTEM VALUE
SELECT
    p.id,
    p.empresa_id,
    p.referencia,
    p.nombre,
    p.descripcion,
    (SELECT id FROM tipos_articulo WHERE codigo = 'PRODUCTO'),
    COALESCE(p.unidad_medida_id, (SELECT id FROM unidades_medida ORDER BY id ASC LIMIT 1)),
    COALESCE(p.categoria_id, (SELECT id FROM categorias_producto c WHERE c.empresa_id = p.empresa_id AND c.deleted_at IS NULL ORDER BY id ASC LIMIT 1)),
    p.marca_id,
    COALESCE(p.grupo_impuesto_id, (SELECT id FROM grupos_impuesto gi WHERE gi.empresa_id = p.empresa_id AND gi.deleted_at IS NULL ORDER BY id ASC LIMIT 1)),
    p.activo,
    p.created_at,
    p.updated_at,
    p.deleted_at
FROM productos p;

SELECT setval(
    pg_get_serial_sequence('articulos', 'id'),
    COALESCE((SELECT MAX(id) FROM articulos), 1),
    EXISTS (SELECT 1 FROM articulos)
);

-- ---------------------------------------------------------------------
-- Reemplazar tabla antigua productos por extensión 1:1 productos
-- ---------------------------------------------------------------------
DROP TABLE productos;

CREATE TABLE productos (
    articulo_id          INT     PRIMARY KEY REFERENCES articulos (id) ON DELETE CASCADE,
    control_inventario   BOOLEAN NOT NULL DEFAULT TRUE,
    metodo_valoracion_id INT     REFERENCES metodos_valoracion (id),
    permite_compra       BOOLEAN NOT NULL DEFAULT TRUE,
    permite_venta        BOOLEAN NOT NULL DEFAULT TRUE,

    CONSTRAINT ck_productos_valoracion_inventario
        CHECK (NOT control_inventario OR metodo_valoracion_id IS NOT NULL)
);

INSERT INTO productos (articulo_id, control_inventario, metodo_valoracion_id, permite_compra, permite_venta)
SELECT
    a.id,
    TRUE,
    (SELECT id FROM metodos_valoracion WHERE codigo = 'PROMEDIO_PONDERADO'),
    TRUE,
    TRUE
FROM articulos a
WHERE a.tipo_articulo_id = (SELECT id FROM tipos_articulo WHERE codigo = 'PRODUCTO');

-- ---------------------------------------------------------------------
-- Tabla extensión 1:1 servicios (duracion_estimada en horas)
-- ---------------------------------------------------------------------
CREATE TABLE servicios (
    articulo_id           INT          PRIMARY KEY REFERENCES articulos (id) ON DELETE CASCADE,
    modalidad_id          INT          NOT NULL REFERENCES modalidades_servicio (id),
    duracion_estimada     NUMERIC(6,2),
    requiere_programacion BOOLEAN      NOT NULL DEFAULT FALSE,
    control_consumos      BOOLEAN      NOT NULL DEFAULT FALSE,

    CONSTRAINT ck_servicios_duracion
        CHECK (duracion_estimada IS NULL OR duracion_estimada > 0)
);
