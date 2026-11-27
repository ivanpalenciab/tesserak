-- =====================================================================
-- 000026 - Agrega el rol de sistema GESTOR_COMPRAS
--   * Permite consultar, crear y modificar compras (sin eliminar).
--   * Combinable con otros roles en usuario_roles.
-- Motor: PostgreSQL 14+
-- =====================================================================

INSERT INTO roles_sistema (codigo, nombre, descripcion, es_sistema, activo)
VALUES (
    'GESTOR_COMPRAS',
    'Gestor de compras',
    'Puede consultar, crear y modificar compras (sin eliminar). Combinable con otros roles de sistema.',
    TRUE,
    TRUE
)
ON CONFLICT DO NOTHING;
