-- =====================================================================
-- 000006 - Módulo CRM: Etapas de oportunidades (con soft delete)
-- Motor: PostgreSQL 14+
-- =====================================================================

CREATE TABLE crm_etapa_oportunidad (
    id                INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    empresa_id        INT          NOT NULL REFERENCES empresas (id),
    nombre            VARCHAR(80)  NOT NULL,
    orden             INT          NOT NULL DEFAULT 1,
    probabilidad_pct  INT          NOT NULL DEFAULT 0,
    es_ganada         BOOLEAN      NOT NULL DEFAULT FALSE,
    es_perdida        BOOLEAN      NOT NULL DEFAULT FALSE,
    activo            BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at        TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ  NOT NULL DEFAULT now(),
    deleted_at        TIMESTAMPTZ,

    CONSTRAINT uq_crm_etapa_oportunidad_id_empresa UNIQUE (id, empresa_id),
    CONSTRAINT ck_crm_etapa_oportunidad_orden CHECK (orden >= 0),
    CONSTRAINT ck_crm_etapa_oportunidad_probabilidad CHECK (probabilidad_pct >= 0 AND probabilidad_pct <= 100),
    CONSTRAINT ck_crm_etapa_oportunidad_cierre CHECK (NOT (es_ganada AND es_perdida))
);

CREATE UNIQUE INDEX ux_crm_etapa_oportunidad_empresa_nombre
    ON crm_etapa_oportunidad (empresa_id, lower(nombre))
    WHERE deleted_at IS NULL;

CREATE INDEX ix_crm_etapa_oportunidad_empresa_orden
    ON crm_etapa_oportunidad (empresa_id, orden)
    WHERE deleted_at IS NULL;
