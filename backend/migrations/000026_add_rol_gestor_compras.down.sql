-- =====================================================================
-- 000026 (down) - Elimina el rol de sistema GESTOR_COMPRAS
-- =====================================================================

DELETE FROM usuario_roles
WHERE rol_id IN (
    SELECT id FROM roles_sistema WHERE upper(codigo) = 'GESTOR_COMPRAS'
);

DELETE FROM roles_sistema
WHERE upper(codigo) = 'GESTOR_COMPRAS';
