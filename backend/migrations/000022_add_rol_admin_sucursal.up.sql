-- =====================================================================
-- 000022 - Agrega el rol de sistema ADMINISTRADOR_SUCURSAL
--   * Acceso administrativo a todos los módulos dentro del alcance
--     de las sucursales asignadas al usuario.
--   * Puede crear, editar y eliminar usuarios únicamente dentro de
--     sus sucursales asignadas (sin poder asignar el rol global
--     ADMINISTRADOR).
-- Motor: PostgreSQL 14+
-- =====================================================================

INSERT INTO roles_sistema (codigo, nombre, descripcion, es_sistema, activo)
VALUES (
    'ADMINISTRADOR_SUCURSAL',
    'Administrador de sucursal',
    'Acceso administrativo a todos los módulos dentro de sus sucursales asignadas, incluyendo creación y gestión de usuarios solo para dichas sucursales.',
    TRUE,
    TRUE
)
ON CONFLICT DO NOTHING;
