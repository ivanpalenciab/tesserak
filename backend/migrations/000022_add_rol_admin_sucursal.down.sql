-- =====================================================================
-- 000022 (down) - Elimina el rol de sistema ADMINISTRADOR_SUCURSAL
-- =====================================================================

DELETE FROM usuario_roles
WHERE rol_id IN (
    SELECT id FROM roles_sistema WHERE upper(codigo) = 'ADMINISTRADOR_SUCURSAL'
);

DELETE FROM roles_sistema
WHERE upper(codigo) = 'ADMINISTRADOR_SUCURSAL';
