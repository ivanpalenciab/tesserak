-- =====================================================================
-- 000007 - Módulo CRM: crm_estado_oportunidades y crm_oportunidades
-- Motor: PostgreSQL 14+
-- =====================================================================

-- ---------------------------------------------------------------------
-- Renombra crm_etapa_oportunidad -> crm_estado_oportunidades
-- ---------------------------------------------------------------------
ALTER TABLE IF EXISTS crm_etapa_oportunidad
    RENAME TO crm_estado_oportunidades;

ALTER TABLE crm_estado_oportunidades
    RENAME CONSTRAINT uq_crm_etapa_oportunidad_id_empresa TO uq_crm_estado_oportunidades_id_empresa;

-- ---------------------------------------------------------------------
-- crm_oportunidades (con FK compuestas multi-tenant y soft delete)
-- ---------------------------------------------------------------------
CREATE TABLE crm_oportunidades (
    id                     BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    empresa_id             INT            NOT NULL REFERENCES empresas (id),
    sucursal_id            INT,
    lead_id                BIGINT,
    campana_id             BIGINT,
    estado_oportunidad_id  INT            NOT NULL,
    motivo_descarte_id     INT,
    usuario_asignado_id    BIGINT         REFERENCES usuarios (id),
    titulo                 VARCHAR(150)   NOT NULL,
    descripcion            TEXT,
    nombre_cliente         VARCHAR(150),
    valor_estimado         NUMERIC(14,2)  NOT NULL DEFAULT 0,
    moneda_id              INT            REFERENCES monedas (id),
    probabilidad_pct       INT            NOT NULL DEFAULT 0,
    fecha_cierre_estimada  DATE,
    fecha_cierre_real      TIMESTAMPTZ,
    prioridad              VARCHAR(20)    NOT NULL DEFAULT 'MEDIA',
    created_by             BIGINT         NOT NULL REFERENCES usuarios (id),
    created_at             TIMESTAMPTZ    NOT NULL DEFAULT now(),
    updated_at             TIMESTAMPTZ    NOT NULL DEFAULT now(),
    deleted_at             TIMESTAMPTZ,

    CONSTRAINT uq_crm_oportunidades_id_empresa UNIQUE (id, empresa_id),
    CONSTRAINT ck_crm_oportunidades_valor CHECK (valor_estimado >= 0),
    CONSTRAINT ck_crm_oportunidades_probabilidad CHECK (probabilidad_pct >= 0 AND probabilidad_pct <= 100),
    CONSTRAINT ck_crm_oportunidades_prioridad CHECK (prioridad IN ('BAJA', 'MEDIA', 'ALTA')),
    CONSTRAINT fk_crm_oportunidades_sucursal_empresa
        FOREIGN KEY (sucursal_id, empresa_id) REFERENCES sucursales (id, empresa_id),
    CONSTRAINT fk_crm_oportunidades_lead_empresa
        FOREIGN KEY (lead_id, empresa_id) REFERENCES crm_leads (id, empresa_id),
    CONSTRAINT fk_crm_oportunidades_campana_empresa
        FOREIGN KEY (campana_id, empresa_id) REFERENCES crm_campanas (id, empresa_id),
    CONSTRAINT fk_crm_oportunidades_estado_empresa
        FOREIGN KEY (estado_oportunidad_id, empresa_id) REFERENCES crm_estado_oportunidades (id, empresa_id),
    CONSTRAINT fk_crm_oportunidades_motivo_empresa
        FOREIGN KEY (motivo_descarte_id, empresa_id) REFERENCES crm_motivo_descarte (id, empresa_id)
);

CREATE INDEX ix_crm_oportunidades_empresa_estado
    ON crm_oportunidades (empresa_id, estado_oportunidad_id)
    WHERE deleted_at IS NULL;

CREATE INDEX ix_crm_oportunidades_empresa_lead
    ON crm_oportunidades (empresa_id, lead_id)
    WHERE deleted_at IS NULL;

CREATE INDEX ix_crm_oportunidades_empresa_cierre
    ON crm_oportunidades (empresa_id, fecha_cierre_estimada)
    WHERE deleted_at IS NULL;

CREATE INDEX ix_crm_oportunidades_deleted_at
    ON crm_oportunidades (deleted_at);
