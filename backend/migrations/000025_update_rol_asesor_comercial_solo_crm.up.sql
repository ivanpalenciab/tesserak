-- =====================================================================
-- 000025 - Actualiza el rol ASESOR_COMERCIAL a acceso exclusivo de CRM
--   * Los roles son combinables (N:M en usuario_roles): un usuario puede
--     tener ASESOR_COMERCIAL (CRM) y extenderse con GESTOR_VENTAS
--     (Ventas + Terceros clientes) u otros roles de sistema.
-- Motor: PostgreSQL 14+
-- =====================================================================

UPDATE roles_sistema
SET
    descripcion = 'Acceso completo a todo el módulo CRM (leads, oportunidades, campañas, actividades y configuración). Combinable con otros roles como Gestor de ventas.',
    updated_at  = now()
WHERE upper(codigo) = 'ASESOR_COMERCIAL';
