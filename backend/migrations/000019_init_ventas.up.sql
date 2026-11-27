-- =====================================================================
-- 000019 - Módulo de Ventas (ven_ventas, ven_ventas_detalle)
--          + Cliente por defecto "Consumidor Final" (222222222222) por empresa
-- Motor: PostgreSQL 14+
-- =====================================================================

-- ---------------------------------------------------------------------
-- 1. Backfill de "Consumidor Final" (222222222222, rol único CLIENTE)
--    para todas las empresas activas existentes que aún no lo tengan
-- ---------------------------------------------------------------------
INSERT INTO tercero_maestro (
    empresa_id,
    tipo_identificacion,
    numero_identificacion,
    razon_social,
    naturaleza_juridica,
    activo
)
SELECT
    e.id,
    'CC',
    '222222222222',
    'Consumidor Final',
    'NATURAL',
    TRUE
FROM empresas e
WHERE e.deleted_at IS NULL
  AND NOT EXISTS (
      SELECT 1
      FROM tercero_maestro t
      WHERE t.empresa_id = e.id
        AND lower(t.numero_identificacion) = '222222222222'
        AND t.deleted_at IS NULL
  )
ON CONFLICT DO NOTHING;

INSERT INTO tercero_empresa_roles (tercero_id, rol_id)
SELECT t.id, r.id
FROM tercero_maestro t
JOIN roles_tercero r ON r.codigo = 'CLIENTE'
WHERE lower(t.numero_identificacion) = '222222222222'
  AND t.deleted_at IS NULL
ON CONFLICT (tercero_id, rol_id) DO NOTHING;

-- ---------------------------------------------------------------------
-- 2. Cabecera: ven_ventas
-- ---------------------------------------------------------------------
CREATE TABLE ven_ventas (
    id                BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    empresa_id        INT           NOT NULL REFERENCES empresas (id),
    sucursal_id       INT           NOT NULL,
    ubicacion_id      INT           NOT NULL,
    cliente_id        BIGINT        NOT NULL,
    consecutivo       VARCHAR(50)   NOT NULL,
    moneda_id         INT           REFERENCES monedas (id),
    estado            VARCHAR(30)   NOT NULL DEFAULT 'REGISTRADA',
    fecha_venta       TIMESTAMPTZ   NOT NULL,
    fecha_vencimiento TIMESTAMPTZ,
    subtotal          NUMERIC(14,4) NOT NULL DEFAULT 0,
    total_iva         NUMERIC(14,4) NOT NULL DEFAULT 0,
    total             NUMERIC(14,4) NOT NULL DEFAULT 0,
    observaciones     TEXT,
    usuario_id        BIGINT        NOT NULL REFERENCES usuarios (id),
    created_at        TIMESTAMPTZ   NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ   NOT NULL DEFAULT now(),
    deleted_at        TIMESTAMPTZ,

    CONSTRAINT uq_ven_ventas_id_empresa UNIQUE (id, empresa_id),
    CONSTRAINT fk_ven_ventas_sucursal_empresa
        FOREIGN KEY (sucursal_id, empresa_id) REFERENCES sucursales (id, empresa_id),
    CONSTRAINT fk_ven_ventas_ubicacion_empresa
        FOREIGN KEY (ubicacion_id, empresa_id) REFERENCES inv_ubicaciones (id, empresa_id),
    CONSTRAINT fk_ven_ventas_cliente_empresa
        FOREIGN KEY (cliente_id, empresa_id) REFERENCES tercero_maestro (id, empresa_id),
    CONSTRAINT ck_ven_ventas_estado
        CHECK (estado IN ('BORRADOR', 'REGISTRADA', 'PAGADA', 'ANULADA')),
    CONSTRAINT ck_ven_ventas_fechas
        CHECK (fecha_vencimiento IS NULL OR fecha_vencimiento >= fecha_venta),
    CONSTRAINT ck_ven_ventas_totales
        CHECK (subtotal >= 0 AND total_iva >= 0 AND total >= 0)
);

CREATE UNIQUE INDEX ux_ven_ventas_empresa_consecutivo
    ON ven_ventas (empresa_id, lower(consecutivo))
    WHERE deleted_at IS NULL;

CREATE INDEX ix_ven_ventas_empresa_fecha
    ON ven_ventas (empresa_id, fecha_venta DESC)
    WHERE deleted_at IS NULL;

CREATE INDEX ix_ven_ventas_empresa_cliente
    ON ven_ventas (empresa_id, cliente_id)
    WHERE deleted_at IS NULL;

CREATE INDEX ix_ven_ventas_empresa_estado
    ON ven_ventas (empresa_id, estado)
    WHERE deleted_at IS NULL;

CREATE INDEX ix_ven_ventas_deleted_at
    ON ven_ventas (deleted_at);

-- ---------------------------------------------------------------------
-- 3. Detalle: ven_ventas_detalle
-- ---------------------------------------------------------------------
CREATE TABLE ven_ventas_detalle (
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    venta_id        BIGINT        NOT NULL REFERENCES ven_ventas (id) ON DELETE CASCADE,
    articulo_id     INT           NOT NULL REFERENCES articulos (id),
    descripcion     VARCHAR(255),
    cantidad        NUMERIC(14,4) NOT NULL,
    precio_unitario NUMERIC(14,4) NOT NULL DEFAULT 0,
    subtotal        NUMERIC(14,4) NOT NULL DEFAULT 0,
    impuesto        NUMERIC(14,4) NOT NULL DEFAULT 0,
    total           NUMERIC(14,4) NOT NULL DEFAULT 0,

    CONSTRAINT ck_ven_ventas_detalle_cantidad
        CHECK (cantidad > 0),
    CONSTRAINT ck_ven_ventas_detalle_valores
        CHECK (precio_unitario >= 0 AND subtotal >= 0 AND impuesto >= 0 AND total >= 0)
);

CREATE INDEX ix_ven_ventas_detalle_venta
    ON ven_ventas_detalle (venta_id);

CREATE INDEX ix_ven_ventas_detalle_articulo
    ON ven_ventas_detalle (articulo_id);
