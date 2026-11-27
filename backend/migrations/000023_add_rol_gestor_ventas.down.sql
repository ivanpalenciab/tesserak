-- =====================================================================
-- 000023 (down) - Elimina el rol de sistema GESTOR_VENTAS
-- =====================================================================

DELETE FROM usuario_roles
WHERE rol_id IN (
    SELECT id FROM roles_sistema WHERE upper(codigo) = 'GESTOR_VENTAS'
);

DELETE FROM roles_sistema
WHERE upper(codigo) = 'GESTOR_VENTAS';
