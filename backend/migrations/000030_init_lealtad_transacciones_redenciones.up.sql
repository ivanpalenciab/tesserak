-- =====================================================================
-- 000030 - Módulo de Lealtad: libro mayor (lealtad_transacciones) y
--          redenciones (lealtad_redenciones)
--   * Ledger inmutable con saldo_posterior materializado
--   * Trazabilidad multi-empresa (empresa_id, sucursal_id, venta_id)
-- Motor: PostgreSQL 14+
-- =====================================================================

-- ---------------------------------------------------------------------
-- lealtad_transacciones (Append-only Ledger)
-- ---------------------------------------------------------------------
CREATE TABLE lealtad_transacciones (
    id                 BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    cuenta_id          BIGINT       NOT NULL REFERENCES lealtad_cuentas (id),
    programa_id        INT          NOT NULL REFERENCES lealtad_programas (id),
    empresa_id         INT          NOT NULL REFERENCES empresas (id),
    sucursal_id        INT          REFERENCES sucursales (id),
    venta_id           BIGINT       REFERENCES ven_ventas (id),
    regla_id           INT          REFERENCES lealtad_reglas (id),
    tipo_transaccion   VARCHAR(30)  NOT NULL,
    puntos             BIGINT       NOT NULL,
    saldo_posterior    BIGINT       NOT NULL,
    fecha_vencimiento  TIMESTAMPTZ,
    observaciones      TEXT,
    created_by         BIGINT       NOT NULL REFERENCES usuarios (id),
    created_at         TIMESTAMPTZ  NOT NULL DEFAULT now(),

    CONSTRAINT ck_lealtad_transacciones_tipo
        CHECK (tipo_transaccion IN ('ACUMULACION', 'REDENCION', 'AJUSTE_POSITIVO', 'AJUSTE_NEGATIVO', 'VENCIMIENTO', 'REVERSO')),
    CONSTRAINT ck_lealtad_transacciones_puntos_no_cero
        CHECK (puntos <> 0),
    CONSTRAINT ck_lealtad_transacciones_saldo_no_negativo
        CHECK (saldo_posterior >= 0)
);

CREATE INDEX ix_lealtad_transacciones_cuenta_fecha
    ON lealtad_transacciones (cuenta_id, created_at DESC, id DESC);

CREATE INDEX ix_lealtad_transacciones_empresa_fecha
    ON lealtad_transacciones (empresa_id, created_at DESC);

CREATE INDEX ix_lealtad_transacciones_venta
    ON lealtad_transacciones (venta_id)
    WHERE venta_id IS NOT NULL;

-- ---------------------------------------------------------------------
-- lealtad_redenciones (documento 1:1 con movimiento de redención)
-- ---------------------------------------------------------------------
CREATE TABLE lealtad_redenciones (
    id                BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    cuenta_id         BIGINT         NOT NULL REFERENCES lealtad_cuentas (id),
    transaccion_id    BIGINT         NOT NULL UNIQUE REFERENCES lealtad_transacciones (id),
    empresa_id        INT            NOT NULL REFERENCES empresas (id),
    sucursal_id       INT            REFERENCES sucursales (id),
    venta_id          BIGINT         REFERENCES ven_ventas (id),
    puntos_redimidos  BIGINT         NOT NULL,
    monto_descontado  NUMERIC(14,4)  NOT NULL,
    moneda_id         INT            NOT NULL REFERENCES monedas (id),
    estado            VARCHAR(20)    NOT NULL DEFAULT 'APLICADA',
    observaciones     TEXT,
    created_by        BIGINT         NOT NULL REFERENCES usuarios (id),
    created_at        TIMESTAMPTZ    NOT NULL DEFAULT now(),

    CONSTRAINT ck_lealtad_redenciones_puntos
        CHECK (puntos_redimidos > 0),
    CONSTRAINT ck_lealtad_redenciones_monto
        CHECK (monto_descontado >= 0),
    CONSTRAINT ck_lealtad_redenciones_estado
        CHECK (estado IN ('APLICADA', 'ANULADA'))
);

CREATE INDEX ix_lealtad_redenciones_cuenta
    ON lealtad_redenciones (cuenta_id, created_at DESC);

CREATE INDEX ix_lealtad_redenciones_empresa
    ON lealtad_redenciones (empresa_id, created_at DESC);
