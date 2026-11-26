-- =====================================================================
-- 000012 - Módulo de Compras (com_compras y com_compras_detalle)
-- Motor: PostgreSQL 14+
-- =====================================================================

-- ---------------------------------------------------------------------
-- Cabecera: com_compras
-- Registra las compras a proveedores por empresa y sucursal, con soft delete.
-- Nota: tercero_empresa_rol_id queda sin FK hasta implementar el módulo Terceros.
-- ---------------------------------------------------------------------
CREATE TABLE com_compras (
    id                       BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    empresa_id               INT           NOT NULL REFERENCES empresas (id),
    sucursal_id              INT           NOT NULL,
    tercero_empresa_rol_id   BIGINT,
    proveedor_nombre         VARCHAR(150),
    consecutivo              VARCHAR(50)   NOT NULL,
    numero_factura_proveedor VARCHAR(50),
    moneda_id                INT           REFERENCES monedas (id),
    estado                   VARCHAR(30)   NOT NULL DEFAULT 'REGISTRADA',
    fecha_compra             TIMESTAMPTZ   NOT NULL,
    fecha_vencimiento        TIMESTAMPTZ,
    subtotal                 NUMERIC(14,4) NOT NULL DEFAULT 0,
    total_iva                NUMERIC(14,4) NOT NULL DEFAULT 0,
    total                    NUMERIC(14,4) NOT NULL DEFAULT 0,
    observaciones            TEXT,
    usuario_id               BIGINT        NOT NULL REFERENCES usuarios (id),
    created_at               TIMESTAMPTZ   NOT NULL DEFAULT now(),
    updated_at               TIMESTAMPTZ   NOT NULL DEFAULT now(),
    deleted_at               TIMESTAMPTZ,

    CONSTRAINT uq_com_compras_id_empresa UNIQUE (id, empresa_id),
    CONSTRAINT fk_com_compras_sucursal_empresa
        FOREIGN KEY (sucursal_id, empresa_id) REFERENCES sucursales (id, empresa_id),
    CONSTRAINT ck_com_compras_estado
        CHECK (estado IN ('BORRADOR', 'REGISTRADA', 'PAGADA', 'ANULADA')),
    CONSTRAINT ck_com_compras_fechas
        CHECK (fecha_vencimiento IS NULL OR fecha_vencimiento >= fecha_compra),
    CONSTRAINT ck_com_compras_totales
        CHECK (subtotal >= 0 AND total_iva >= 0 AND total >= 0)
);

CREATE UNIQUE INDEX ux_com_compras_empresa_consecutivo
    ON com_compras (empresa_id, lower(consecutivo))
    WHERE deleted_at IS NULL;

CREATE INDEX ix_com_compras_empresa_fecha
    ON com_compras (empresa_id, fecha_compra DESC)
    WHERE deleted_at IS NULL;

CREATE INDEX ix_com_compras_empresa_sucursal
    ON com_compras (empresa_id, sucursal_id)
    WHERE deleted_at IS NULL;

CREATE INDEX ix_com_compras_empresa_estado
    ON com_compras (empresa_id, estado)
    WHERE deleted_at IS NULL;

CREATE INDEX ix_com_compras_deleted_at
    ON com_compras (deleted_at);

-- ---------------------------------------------------------------------
-- Detalle: com_compras_detalle
-- Líneas de artículos comprados asociadas a cada registro de com_compras.
-- ---------------------------------------------------------------------
CREATE TABLE com_compras_detalle (
    id             BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    compra_id      BIGINT        NOT NULL REFERENCES com_compras (id) ON DELETE CASCADE,
    articulo_id    INT           NOT NULL REFERENCES articulos (id),
    descripcion    VARCHAR(255),
    cantidad       NUMERIC(14,4) NOT NULL,
    costo_unitario NUMERIC(14,4) NOT NULL DEFAULT 0,
    subtotal       NUMERIC(14,4) NOT NULL DEFAULT 0,
    impuesto       NUMERIC(14,4) NOT NULL DEFAULT 0,
    total          NUMERIC(14,4) NOT NULL DEFAULT 0,

    CONSTRAINT ck_com_compras_detalle_cantidad
        CHECK (cantidad > 0),
    CONSTRAINT ck_com_compras_detalle_valores
        CHECK (costo_unitario >= 0 AND subtotal >= 0 AND impuesto >= 0 AND total >= 0)
);

CREATE INDEX ix_com_compras_detalle_compra
    ON com_compras_detalle (compra_id);

CREATE INDEX ix_com_compras_detalle_articulo
    ON com_compras_detalle (articulo_id);
