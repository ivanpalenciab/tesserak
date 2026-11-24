-- =====================================================================
-- 000005 - Módulo CRM: Campañas de marketing, atribución de leads y gastos
-- Motor: PostgreSQL 14+
-- =====================================================================

-- ---------------------------------------------------------------------
-- Restricción única compuesta en crm_leads para integridad multi-tenant
-- Permite referenciar (lead_id, empresa_id) desde crm_campana_leads.
-- ---------------------------------------------------------------------
ALTER TABLE crm_leads
    ADD CONSTRAINT uq_crm_leads_id_empresa UNIQUE (id, empresa_id);

-- ---------------------------------------------------------------------
-- crm_tipos_campana (catálogo de tipos de campaña por empresa)
-- ---------------------------------------------------------------------
CREATE TABLE crm_tipos_campana (
    id          INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    empresa_id  INT          NOT NULL REFERENCES empresas (id),
    nombre      VARCHAR(80)  NOT NULL,
    descripcion VARCHAR(255),
    activo      BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    deleted_at  TIMESTAMPTZ,

    CONSTRAINT uq_crm_tipos_campana_id_empresa UNIQUE (id, empresa_id)
);

CREATE UNIQUE INDEX ux_crm_tipos_campana_empresa_nombre
    ON crm_tipos_campana (empresa_id, lower(nombre))
    WHERE deleted_at IS NULL;

CREATE INDEX ix_crm_tipos_campana_empresa
    ON crm_tipos_campana (empresa_id)
    WHERE deleted_at IS NULL;

-- ---------------------------------------------------------------------
-- crm_campanas
--   * FK compuestas (..., empresa_id) garantizan que tipo_campana y
--     origen pertenezcan a la misma empresa de la campaña.
--   * Si se define presupuesto, debe indicarse moneda_id.
-- ---------------------------------------------------------------------
CREATE TABLE crm_campanas (
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    empresa_id      INT            NOT NULL REFERENCES empresas (id),
    tipo_campana_id INT            NOT NULL,
    origen_id       INT,
    moneda_id       INT            REFERENCES monedas (id),
    nombre          VARCHAR(150)   NOT NULL,
    descripcion     TEXT,
    objetivo        TEXT,
    fecha_inicio    DATE           NOT NULL,
    fecha_fin       DATE,
    presupuesto     NUMERIC(14,2),
    estado          VARCHAR(20)    NOT NULL DEFAULT 'BORRADOR',
    created_by      BIGINT         NOT NULL REFERENCES usuarios (id),
    created_at      TIMESTAMPTZ    NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ    NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT uq_crm_campanas_id_empresa UNIQUE (id, empresa_id),
    CONSTRAINT ck_crm_campanas_fechas
        CHECK (fecha_fin IS NULL OR fecha_fin >= fecha_inicio),
    CONSTRAINT ck_crm_campanas_presupuesto
        CHECK (presupuesto IS NULL OR presupuesto >= 0),
    CONSTRAINT ck_crm_campanas_presupuesto_moneda
        CHECK (presupuesto IS NULL OR moneda_id IS NOT NULL),
    CONSTRAINT ck_crm_campanas_estado
        CHECK (estado IN ('BORRADOR', 'PROGRAMADA', 'ACTIVA', 'PAUSADA', 'FINALIZADA')),
    CONSTRAINT fk_crm_campanas_tipo_empresa
        FOREIGN KEY (tipo_campana_id, empresa_id) REFERENCES crm_tipos_campana (id, empresa_id),
    CONSTRAINT fk_crm_campanas_origen_empresa
        FOREIGN KEY (origen_id, empresa_id) REFERENCES crm_origen_lead (id, empresa_id)
);

CREATE INDEX ix_crm_campanas_empresa_estado
    ON crm_campanas (empresa_id, estado)
    WHERE deleted_at IS NULL;

CREATE INDEX ix_crm_campanas_empresa_fechas
    ON crm_campanas (empresa_id, fecha_inicio, fecha_fin)
    WHERE deleted_at IS NULL;

CREATE INDEX ix_crm_campanas_deleted_at
    ON crm_campanas (deleted_at);

-- ---------------------------------------------------------------------
-- crm_campana_leads (relación N:M entre campañas y leads con atribución y UTM)
--   * FK compuestas garantizan que campaña y lead pertenezcan a la misma empresa.
--   * UNIQUE (empresa_id, campana_id, lead_id) evita asociaciones duplicadas.
--   * Índice único parcial garantiza como máximo un first_touch por lead.
-- ---------------------------------------------------------------------
CREATE TABLE crm_campana_leads (
    id                   BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    empresa_id           INT          NOT NULL REFERENCES empresas (id),
    campana_id           BIGINT       NOT NULL,
    lead_id              BIGINT       NOT NULL,
    fecha_asociacion     TIMESTAMPTZ  NOT NULL DEFAULT now(),
    tipo_atribucion      VARCHAR(20),
    utm_source           VARCHAR(255),
    utm_medium           VARCHAR(255),
    utm_campaign         VARCHAR(255),
    utm_content          VARCHAR(255),
    utm_term             VARCHAR(255),
    landing_page         TEXT,
    referrer_url         TEXT,
    external_campaign_id VARCHAR(150),
    external_ad_id       VARCHAR(150),
    created_at           TIMESTAMPTZ  NOT NULL DEFAULT now(),

    CONSTRAINT uq_crm_campana_leads_campana_lead
        UNIQUE (empresa_id, campana_id, lead_id),
    CONSTRAINT ck_crm_campana_leads_tipo_atribucion
        CHECK (tipo_atribucion IS NULL OR tipo_atribucion IN ('first_touch', 'last_touch', 'influencia')),
    CONSTRAINT fk_crm_campana_leads_campana_empresa
        FOREIGN KEY (campana_id, empresa_id) REFERENCES crm_campanas (id, empresa_id),
    CONSTRAINT fk_crm_campana_leads_lead_empresa
        FOREIGN KEY (lead_id, empresa_id) REFERENCES crm_leads (id, empresa_id)
);

CREATE UNIQUE INDEX ux_crm_campana_leads_first_touch
    ON crm_campana_leads (empresa_id, lead_id)
    WHERE tipo_atribucion = 'first_touch';

CREATE INDEX ix_crm_campana_leads_empresa_campana
    ON crm_campana_leads (empresa_id, campana_id, fecha_asociacion DESC);

CREATE INDEX ix_crm_campana_leads_empresa_lead
    ON crm_campana_leads (empresa_id, lead_id, fecha_asociacion DESC);

-- ---------------------------------------------------------------------
-- crm_campana_gastos
--   * proveedor_tercero_empresa_rol_id quedará listo para agregar su FK
--     cuando se cree el módulo de terceros (mismo patrón que crm_leads).
-- ---------------------------------------------------------------------
CREATE TABLE crm_campana_gastos (
    id                               BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    empresa_id                       INT            NOT NULL REFERENCES empresas (id),
    campana_id                       BIGINT         NOT NULL,
    fecha                            DATE           NOT NULL,
    concepto                         VARCHAR(150)   NOT NULL,
    categoria_gasto                  VARCHAR(50),
    valor                            NUMERIC(14,2)  NOT NULL,
    moneda_id                        INT            NOT NULL REFERENCES monedas (id),
    proveedor_tercero_empresa_rol_id BIGINT,
    created_by                       BIGINT         NOT NULL REFERENCES usuarios (id),
    created_at                       TIMESTAMPTZ    NOT NULL DEFAULT now(),
    updated_at                       TIMESTAMPTZ    NOT NULL DEFAULT now(),
    deleted_at                       TIMESTAMPTZ,

    CONSTRAINT ck_crm_campana_gastos_valor
        CHECK (valor >= 0),
    CONSTRAINT fk_crm_campana_gastos_campana_empresa
        FOREIGN KEY (campana_id, empresa_id) REFERENCES crm_campanas (id, empresa_id)
);

CREATE INDEX ix_crm_campana_gastos_empresa_campana_fecha
    ON crm_campana_gastos (empresa_id, campana_id, fecha DESC)
    WHERE deleted_at IS NULL;

CREATE INDEX ix_crm_campana_gastos_deleted_at
    ON crm_campana_gastos (deleted_at);

-- ---------------------------------------------------------------------
-- Valores iniciales de catálogo crm_tipos_campana para empresas existentes
-- ---------------------------------------------------------------------
INSERT INTO crm_tipos_campana (empresa_id, nombre, descripcion, activo)
SELECT e.id, t.nombre, t.descripcion, TRUE
FROM empresas e
CROSS JOIN (
    VALUES
        ('Publicidad digital', 'Anuncios pagos en redes sociales, buscadores y display'),
        ('Email marketing', 'Campañas de correo masivo, automatizaciones y boletines'),
        ('Marketing de contenidos', 'Artículos, webinars, guías descargables y SEO'),
        ('Eventos', 'Conferencias, lanzamientos, talleres y encuentros presenciales o virtuales'),
        ('Ferias comerciales', 'Participación con stand o patrocinio en ferias del sector'),
        ('Promociones', 'Descuentos, cupones y ofertas comerciales por tiempo limitado'),
        ('Referidos', 'Programas de recomendación de clientes y aliados')
) AS t(nombre, descripcion)
WHERE e.deleted_at IS NULL
ON CONFLICT DO NOTHING;
