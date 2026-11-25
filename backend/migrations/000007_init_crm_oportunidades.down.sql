DROP TABLE IF EXISTS crm_oportunidades;

ALTER TABLE IF EXISTS crm_estado_oportunidades
    RENAME CONSTRAINT uq_crm_estado_oportunidades_id_empresa TO uq_crm_etapa_oportunidad_id_empresa;

ALTER TABLE IF EXISTS crm_estado_oportunidades
    RENAME TO crm_etapa_oportunidad;
