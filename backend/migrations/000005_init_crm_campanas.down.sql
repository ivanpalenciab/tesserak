DROP TABLE IF EXISTS crm_campana_gastos;
DROP TABLE IF EXISTS crm_campana_leads;
DROP TABLE IF EXISTS crm_campanas;
DROP TABLE IF EXISTS crm_tipos_campana;

ALTER TABLE crm_leads
    DROP CONSTRAINT IF EXISTS uq_crm_leads_id_empresa;
