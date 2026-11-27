-- =====================================================================
-- 000024 - Agrega el rol de sistema ASESOR_COMERCIAL
--   * Acceso completo al módulo CRM (leads, oportunidades, campañas,
--     actividades y configuración).
--   * Permiso para consultar y crear ventas (sin modificar ni eliminar).
-- Motor: PostgreSQL 14+
-- =====================================================================

INSERT INTO roles_sistema (codigo, nombre, descripcion, es_sistema, activo)
VALUES (
    'ASESOR_COMERCIAL',
    'Asesor comercial',
    'Acceso completo a todo el módulo CRM y permiso para consultar y crear ventas.',
    TRUE,
    TRUE
)
ON CONFLICT DO NOTHING;
