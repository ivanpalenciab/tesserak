-- =====================================================================
-- 000023 - Agrega el rol de sistema GESTOR_VENTAS
--   * Solo puede crear y modificar ventas y terceros.
--   * En terceros, únicamente puede crear clientes (rol CLIENTE) o
--     asignar el rol CLIENTE a terceros existentes, sin agregar ni
--     quitar otros roles ni eliminar registros.
-- Motor: PostgreSQL 14+
-- =====================================================================

INSERT INTO roles_sistema (codigo, nombre, descripcion, es_sistema, activo)
VALUES (
    'GESTOR_VENTAS',
    'Gestor de ventas',
    'Puede crear y modificar ventas y gestionar terceros únicamente con el rol CLIENTE (crear clientes o asignar el rol CLIENTE a terceros existentes).',
    TRUE,
    TRUE
)
ON CONFLICT DO NOTHING;
