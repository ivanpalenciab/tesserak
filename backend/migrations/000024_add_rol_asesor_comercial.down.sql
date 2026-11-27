-- =====================================================================
-- 000024 (down) - Elimina el rol de sistema ASESOR_COMERCIAL
-- =====================================================================

DELETE FROM usuario_roles
WHERE rol_id IN (
    SELECT id FROM roles_sistema WHERE upper(codigo) = 'ASESOR_COMERCIAL'
);

DELETE FROM roles_sistema
WHERE upper(codigo) = 'ASESOR_COMERCIAL';
