-- =====================================================================
-- 000017 - Módulo de Inventarios (prefijo inv_)
--          inv_ubicaciones, inv_tipos_movimiento, inv_documentos,
--          inv_movimientos_detalle, inv_stock
-- Motor: PostgreSQL 14+
-- =====================================================================

-- ---------------------------------------------------------------------
-- Semilla de tipo de documento para consecutivos de inventario
-- ---------------------------------------------------------------------
INSERT INTO tipos_documento (codigo, nombre, descripcion, prefijo_defecto, activo)
VALUES ('INVENTARIO', 'Movimiento de Inventario', 'Consecutivo automático para ajustes, traslados y movimientos de inventario', 'INV', TRUE)
ON CONFLICT (codigo) DO NOTHING;

-- ---------------------------------------------------------------------
-- 1. inv_ubicaciones (Bodegas, estantes o zonas por empresa y sucursal)
-- ---------------------------------------------------------------------
CREATE TABLE inv_ubicaciones (
    id          INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    empresa_id  INT          NOT NULL REFERENCES empresas (id),
    sucursal_id INT          NOT NULL,
    codigo      VARCHAR(40)  NOT NULL,
    nombre      VARCHAR(120) NOT NULL,
    descripcion VARCHAR(255),
    activo      BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    deleted_at  TIMESTAMPTZ,

    CONSTRAINT uq_inv_ubicaciones_id_empresa UNIQUE (id, empresa_id),
    CONSTRAINT fk_inv_ubicaciones_sucursal_empresa
        FOREIGN KEY (sucursal_id, empresa_id) REFERENCES sucursales (id, empresa_id)
);

CREATE UNIQUE INDEX ux_inv_ubicaciones_empresa_codigo
    ON inv_ubicaciones (empresa_id, lower(codigo))
    WHERE deleted_at IS NULL;

CREATE INDEX ix_inv_ubicaciones_empresa_sucursal
    ON inv_ubicaciones (empresa_id, sucursal_id)
    WHERE deleted_at IS NULL;

-- Crea una ubicación "Bodega Principal" para cada sucursal activa existente que aún no tenga ubicaciones
INSERT INTO inv_ubicaciones (empresa_id, sucursal_id, codigo, nombre, descripcion, activo)
SELECT s.empresa_id, s.id, 'BOD-' || s.id, 'Bodega Principal - ' || s.nombre, 'Ubicación principal por defecto', TRUE
FROM sucursales s
WHERE s.deleted_at IS NULL
  AND NOT EXISTS (
      SELECT 1 FROM inv_ubicaciones u
      WHERE u.sucursal_id = s.id AND u.deleted_at IS NULL
  )
ON CONFLICT DO NOTHING;

-- ---------------------------------------------------------------------
-- 2. inv_tipos_movimiento (Catálogo global con semilla)
--    naturaleza:
--      ENTRADA  (+): suma stock en ubicacion_destino_id
--      SALIDA   (-): resta stock en ubicacion_origen_id
--      TRASLADO (0): resta en ubicacion_origen_id y suma en ubicacion_destino_id
-- ---------------------------------------------------------------------
CREATE TABLE inv_tipos_movimiento (
    id          INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    codigo      VARCHAR(40)  NOT NULL UNIQUE,
    nombre      VARCHAR(100) NOT NULL,
    naturaleza  VARCHAR(20)  NOT NULL,
    descripcion VARCHAR(255),
    activo      BOOLEAN      NOT NULL DEFAULT TRUE,

    CONSTRAINT ck_inv_tipos_movimiento_naturaleza
        CHECK (naturaleza IN ('ENTRADA', 'SALIDA', 'TRASLADO'))
);

INSERT INTO inv_tipos_movimiento (codigo, nombre, naturaleza, descripcion, activo)
VALUES
    ('AJUSTE_POSITIVO',      'Ajuste positivo (sobrante)',  'ENTRADA',  'Incremento de existencias por conteo físico o ajuste', TRUE),
    ('AJUSTE_NEGATIVO',      'Ajuste negativo (faltante)',  'SALIDA',   'Disminución de existencias por conteo físico o ajuste', TRUE),
    ('PERDIDA',              'Pérdida / Merma',             'SALIDA',   'Baja de inventario por pérdida, hurto o merma', TRUE),
    ('AVERIA',               'Avería / Daño',               'SALIDA',   'Baja de inventario por mercancía averiada o vencida', TRUE),
    ('DEVOLUCION_PROVEEDOR', 'Devolución a proveedor',      'SALIDA',   'Salida de mercancía devuelta al proveedor', TRUE),
    ('DEVOLUCION_CLIENTE',   'Devolución de cliente',       'ENTRADA',  'Reingreso de mercancía devuelta por un cliente', TRUE),
    ('ENTRADA_COMPRA',       'Entrada por compra',          'ENTRADA',  'Ingreso de inventario proveniente de una compra', TRUE),
    ('SALIDA_VENTA',         'Salida por venta',            'SALIDA',   'Salida de inventario por venta o despacho', TRUE),
    ('TRASLADO',             'Traslado entre ubicaciones',  'TRASLADO', 'Transferencia de existencias de una ubicación a otra', TRUE)
ON CONFLICT (codigo) DO NOTHING;

-- ---------------------------------------------------------------------
-- 3. inv_documentos (Cabecera de movimiento de inventario)
-- ---------------------------------------------------------------------
CREATE TABLE inv_documentos (
    id                    BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    empresa_id            INT          NOT NULL REFERENCES empresas (id),
    sucursal_id           INT          NOT NULL,
    tipo_movimiento_id    INT          NOT NULL REFERENCES inv_tipos_movimiento (id),
    consecutivo           VARCHAR(50)  NOT NULL,
    fecha                 TIMESTAMPTZ  NOT NULL DEFAULT now(),
    tercero_id            BIGINT       REFERENCES tercero_maestro (id),
    motivo                VARCHAR(255) NOT NULL,
    observaciones         TEXT,
    documento_origen_tipo VARCHAR(40),
    documento_origen_id   BIGINT,
    usuario_id            BIGINT       NOT NULL REFERENCES usuarios (id),
    estado                VARCHAR(20)  NOT NULL DEFAULT 'APLICADO',
    created_at            TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at            TIMESTAMPTZ  NOT NULL DEFAULT now(),

    CONSTRAINT uq_inv_documentos_empresa_consecutivo UNIQUE (empresa_id, consecutivo),
    CONSTRAINT fk_inv_documentos_sucursal_empresa
        FOREIGN KEY (sucursal_id, empresa_id) REFERENCES sucursales (id, empresa_id),
    CONSTRAINT ck_inv_documentos_estado
        CHECK (estado IN ('APLICADO', 'ANULADO'))
);

CREATE INDEX ix_inv_documentos_empresa_fecha
    ON inv_documentos (empresa_id, fecha DESC);

CREATE INDEX ix_inv_documentos_empresa_tipo
    ON inv_documentos (empresa_id, tipo_movimiento_id);

-- ---------------------------------------------------------------------
-- 4. inv_movimientos_detalle (Kardex / Líneas del documento)
-- ---------------------------------------------------------------------
CREATE TABLE inv_movimientos_detalle (
    id                   BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    documento_id         BIGINT        NOT NULL REFERENCES inv_documentos (id) ON DELETE CASCADE,
    articulo_id          INT           NOT NULL REFERENCES articulos (id),
    ubicacion_origen_id  INT           REFERENCES inv_ubicaciones (id),
    ubicacion_destino_id INT           REFERENCES inv_ubicaciones (id),
    cantidad             NUMERIC(14,4) NOT NULL,
    costo_unitario       NUMERIC(14,4) NOT NULL DEFAULT 0,
    created_at           TIMESTAMPTZ   NOT NULL DEFAULT now(),

    CONSTRAINT ck_inv_movimientos_detalle_cantidad CHECK (cantidad > 0),
    CONSTRAINT ck_inv_movimientos_detalle_costo CHECK (costo_unitario >= 0),
    CONSTRAINT ck_inv_movimientos_detalle_ubicacion
        CHECK (ubicacion_origen_id IS NOT NULL OR ubicacion_destino_id IS NOT NULL)
);

CREATE INDEX ix_inv_movimientos_detalle_documento
    ON inv_movimientos_detalle (documento_id);

CREATE INDEX ix_inv_movimientos_detalle_articulo
    ON inv_movimientos_detalle (articulo_id);

-- ---------------------------------------------------------------------
-- 5. inv_stock (Saldo actual materializado por artículo y ubicación)
-- ---------------------------------------------------------------------
CREATE TABLE inv_stock (
    id             BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    empresa_id     INT           NOT NULL REFERENCES empresas (id),
    articulo_id    INT           NOT NULL,
    ubicacion_id   INT           NOT NULL,
    cantidad       NUMERIC(14,4) NOT NULL DEFAULT 0,
    costo_promedio NUMERIC(14,4) NOT NULL DEFAULT 0,
    updated_at     TIMESTAMPTZ   NOT NULL DEFAULT now(),

    CONSTRAINT uq_inv_stock_articulo_ubicacion UNIQUE (articulo_id, ubicacion_id),
    CONSTRAINT fk_inv_stock_articulo_empresa
        FOREIGN KEY (articulo_id, empresa_id) REFERENCES articulos (id, empresa_id),
    CONSTRAINT fk_inv_stock_ubicacion_empresa
        FOREIGN KEY (ubicacion_id, empresa_id) REFERENCES inv_ubicaciones (id, empresa_id),
    CONSTRAINT ck_inv_stock_cantidad CHECK (cantidad >= 0),
    CONSTRAINT ck_inv_stock_costo CHECK (costo_promedio >= 0)
);

CREATE INDEX ix_inv_stock_empresa_ubicacion
    ON inv_stock (empresa_id, ubicacion_id);

CREATE INDEX ix_inv_stock_empresa_articulo
    ON inv_stock (empresa_id, articulo_id);
