-- =====================================================================
-- 000025 (down) - Restaura la descripción anterior de ASESOR_COMERCIAL
-- =====================================================================

UPDATE roles_sistema
SET
    descripcion = 'Acceso completo a todo el módulo CRM y permiso para consultar y crear ventas.',
    updated_at  = now()
WHERE upper(codigo) = 'ASESOR_COMERCIAL';
