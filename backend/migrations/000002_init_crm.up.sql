-- =====================================================================
-- 000002 - Módulo CRM y catálogo de monedas
-- Motor: PostgreSQL 14+
-- =====================================================================

-- ---------------------------------------------------------------------
-- monedas (catálogo global requerido por crm_leads y módulos futuros)
-- ---------------------------------------------------------------------
CREATE TABLE monedas (
    id             INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    codigo_iso     CHAR(3)     NOT NULL UNIQUE,
    nombre_moneda  VARCHAR(50) NOT NULL,
    simbolo        VARCHAR(10) NOT NULL
);

-- ---------------------------------------------------------------------
-- crm_origen_lead
-- ---------------------------------------------------------------------
CREATE TABLE crm_origen_lead (
    id          INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    empresa_id  INT          NOT NULL REFERENCES empresas (id),
    nombre      VARCHAR(100) NOT NULL,
    activo      BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    deleted_at  TIMESTAMPTZ,
    CONSTRAINT uq_crm_origen_lead_id_empresa UNIQUE (id, empresa_id)
);

CREATE UNIQUE INDEX ux_crm_origen_lead_empresa_nombre
    ON crm_origen_lead (empresa_id, lower(nombre))
    WHERE deleted_at IS NULL;

-- ---------------------------------------------------------------------
-- crm_estado_lead (etapas del embudo/pipeline por empresa)
-- ---------------------------------------------------------------------
CREATE TABLE crm_estado_lead (
    id               INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    empresa_id       INT         NOT NULL REFERENCES empresas (id),
    nombre           VARCHAR(50) NOT NULL,
    orden            INT         NOT NULL DEFAULT 1,
    es_estado_final  BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at       TIMESTAMPTZ,
    CONSTRAINT uq_crm_estado_lead_id_empresa UNIQUE (id, empresa_id)
);

CREATE UNIQUE INDEX ux_crm_estado_lead_empresa_nombre
    ON crm_estado_lead (empresa_id, lower(nombre))
    WHERE deleted_at IS NULL;

CREATE INDEX ix_crm_estado_lead_empresa_orden
    ON crm_estado_lead (empresa_id, orden)
    WHERE deleted_at IS NULL;

-- ---------------------------------------------------------------------
-- crm_motivo_descarte
-- ---------------------------------------------------------------------
CREATE TABLE crm_motivo_descarte (
    id           INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    empresa_id   INT          NOT NULL REFERENCES empresas (id),
    descripcion  VARCHAR(150) NOT NULL,
    activo       BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT now(),
    deleted_at   TIMESTAMPTZ,
    CONSTRAINT uq_crm_motivo_descarte_id_empresa UNIQUE (id, empresa_id)
);

CREATE UNIQUE INDEX ux_crm_motivo_descarte_empresa_descripcion
    ON crm_motivo_descarte (empresa_id, lower(descripcion))
    WHERE deleted_at IS NULL;

-- ---------------------------------------------------------------------
-- crm_leads
--   * FK compuestas (..., empresa_id) garantizan que sucursal, origen,
--     estado y motivo pertenezcan a la misma empresa del lead.
--   * convertido_tercero_empresa_rol_id quedará listo para agregar su FK
--     cuando se cree el módulo de terceros.
-- ---------------------------------------------------------------------
CREATE TABLE crm_leads (
    id                                 BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    empresa_id                         INT            NOT NULL REFERENCES empresas (id),
    sucursal_id                        INT,
    origen_id                          INT            NOT NULL,
    estado_id                          INT            NOT NULL,
    motivo_descarte_id                 INT,
    usuario_asignado_id                BIGINT         REFERENCES usuarios (id),
    nombre                             VARCHAR(100)   NOT NULL,
    apellido                           VARCHAR(100)   NOT NULL,
    nombre_empresa                     VARCHAR(150),
    cargo                              VARCHAR(100),
    email                              VARCHAR(255),
    telefono                           VARCHAR(50),
    temperatura                        VARCHAR(20)    NOT NULL DEFAULT 'TIBIO',
    score                              INT            NOT NULL DEFAULT 0,
    valor_estimado                     NUMERIC(14,4)  NOT NULL DEFAULT 0,
    moneda_id                          INT            REFERENCES monedas (id),
    convertido                         BOOLEAN        NOT NULL DEFAULT FALSE,
    convertido_tercero_empresa_rol_id  BIGINT,
    fecha_conversion                   TIMESTAMPTZ,
    created_by                         BIGINT         NOT NULL REFERENCES usuarios (id),
    created_at                         TIMESTAMPTZ    NOT NULL DEFAULT now(),
    updated_at                         TIMESTAMPTZ    NOT NULL DEFAULT now(),
    deleted_at                         TIMESTAMPTZ,

    CONSTRAINT ck_crm_leads_temperatura
        CHECK (temperatura IN ('FRIO', 'TIBIO', 'CALIENTE')),
    CONSTRAINT ck_crm_leads_score
        CHECK (score >= 0 AND score <= 100),
    CONSTRAINT ck_crm_leads_valor_estimado
        CHECK (valor_estimado >= 0),
    CONSTRAINT fk_crm_leads_sucursal_empresa
        FOREIGN KEY (sucursal_id, empresa_id) REFERENCES sucursales (id, empresa_id),
    CONSTRAINT fk_crm_leads_origen_empresa
        FOREIGN KEY (origen_id, empresa_id) REFERENCES crm_origen_lead (id, empresa_id),
    CONSTRAINT fk_crm_leads_estado_empresa
        FOREIGN KEY (estado_id, empresa_id) REFERENCES crm_estado_lead (id, empresa_id),
    CONSTRAINT fk_crm_leads_motivo_empresa
        FOREIGN KEY (motivo_descarte_id, empresa_id) REFERENCES crm_motivo_descarte (id, empresa_id)
);

CREATE INDEX ix_crm_leads_empresa_estado
    ON crm_leads (empresa_id, estado_id)
    WHERE deleted_at IS NULL;

CREATE INDEX ix_crm_leads_empresa_asignado
    ON crm_leads (empresa_id, usuario_asignado_id)
    WHERE deleted_at IS NULL;

CREATE INDEX ix_crm_leads_deleted_at
    ON crm_leads (deleted_at);

-- ---------------------------------------------------------------------
-- crm_lead_actividades
-- ---------------------------------------------------------------------
CREATE TABLE crm_lead_actividades (
    id                 BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    lead_id            BIGINT       NOT NULL REFERENCES crm_leads (id),
    usuario_id         BIGINT       NOT NULL REFERENCES usuarios (id),
    tipo_actividad     VARCHAR(30)  NOT NULL,
    fecha_actividad    TIMESTAMPTZ  NOT NULL DEFAULT now(),
    fecha_vencimiento  TIMESTAMPTZ,
    estado             VARCHAR(20)  NOT NULL DEFAULT 'PENDIENTE',
    asunto             VARCHAR(200) NOT NULL,
    observaciones      TEXT,
    created_at         TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at         TIMESTAMPTZ  NOT NULL DEFAULT now(),
    deleted_at         TIMESTAMPTZ,

    CONSTRAINT ck_crm_actividades_tipo
        CHECK (tipo_actividad IN ('LLAMADA', 'CORREO', 'REUNION', 'WHATSAPP', 'TAREA', 'NOTA')),
    CONSTRAINT ck_crm_actividades_estado
        CHECK (estado IN ('PENDIENTE', 'COMPLETADA', 'CANCELADA'))
);

CREATE INDEX ix_crm_actividades_lead
    ON crm_lead_actividades (lead_id, fecha_actividad DESC)
    WHERE deleted_at IS NULL;

CREATE INDEX ix_crm_actividades_usuario_estado
    ON crm_lead_actividades (usuario_id, estado, fecha_vencimiento)
    WHERE deleted_at IS NULL;

-- ---------------------------------------------------------------------
-- crm_lead_interes_productos
--   * empresa_producto_id y categoria_id quedarán listos para agregar
--     sus FK cuando se cree el módulo de productos.
-- ---------------------------------------------------------------------
CREATE TABLE crm_lead_interes_productos (
    id                   BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    lead_id              BIGINT        NOT NULL REFERENCES crm_leads (id) ON DELETE CASCADE,
    empresa_producto_id  INT,
    categoria_id         INT,
    cantidad_estimada    NUMERIC(14,2) NOT NULL DEFAULT 1,
    notas                VARCHAR(255),
    created_at           TIMESTAMPTZ   NOT NULL DEFAULT now(),

    CONSTRAINT ck_crm_interes_cantidad
        CHECK (cantidad_estimada > 0),
    CONSTRAINT ck_crm_interes_referencia
        CHECK (empresa_producto_id IS NOT NULL OR categoria_id IS NOT NULL OR (notas IS NOT NULL AND length(trim(notas)) > 0))
);

CREATE INDEX ix_crm_interes_lead ON crm_lead_interes_productos (lead_id);
