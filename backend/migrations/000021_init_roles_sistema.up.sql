-- =====================================================================
-- 000021 - Roles de sistema (roles_sistema) y asignación (usuario_roles)
--   * Crea el catálogo global roles_sistema con semilla de sistema:
--       - ADMINISTRADOR: acceso total a la empresa, todas las sucursales,
--         gestión de empresa y administración/creación de usuarios.
--       - USUARIO: acceso operativo restringido a sus sucursales asignadas.
--   * Crea la tabla puente N:M usuario_roles.
--   * Backfill: asigna ADMINISTRADOR al primer usuario activo de cada
--     empresa existente y USUARIO a los demás usuarios existentes.
-- Motor: PostgreSQL 14+
-- =====================================================================

-- 1. Catálogo global de roles de sistema
CREATE TABLE roles_sistema (
    id          INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    codigo      VARCHAR(50)  NOT NULL,
    nombre      VARCHAR(100) NOT NULL,
    descripcion TEXT,
    es_sistema  BOOLEAN      NOT NULL DEFAULT FALSE,
    activo      BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX ux_roles_sistema_codigo
    ON roles_sistema (upper(codigo));

INSERT INTO roles_sistema (codigo, nombre, descripcion, es_sistema, activo)
VALUES
    (
        'ADMINISTRADOR',
        'Administrador',
        'Acceso total a la empresa, todas las sucursales, gestión de empresa y creación/administración de usuarios.',
        TRUE,
        TRUE
    ),
    (
        'USUARIO',
        'Usuario',
        'Acceso operativo a los módulos de la empresa restringido a las sucursales asignadas.',
        TRUE,
        TRUE
    )
ON CONFLICT DO NOTHING;

-- 2. Tabla puente N:M usuario_roles
CREATE TABLE usuario_roles (
    usuario_id BIGINT      NOT NULL REFERENCES usuarios (id) ON DELETE CASCADE,
    rol_id     INT         NOT NULL REFERENCES roles_sistema (id) ON DELETE RESTRICT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    PRIMARY KEY (usuario_id, rol_id)
);

CREATE INDEX ix_usuario_roles_rol ON usuario_roles (rol_id);

-- 3. Backfill: asignar ADMINISTRADOR al primer usuario (id más bajo) de cada empresa
INSERT INTO usuario_roles (usuario_id, rol_id)
SELECT primer.id, r.id
FROM (
    SELECT DISTINCT ON (empresa_id) id
    FROM usuarios
    WHERE deleted_at IS NULL
    ORDER BY empresa_id, id ASC
) primer
JOIN roles_sistema r ON r.codigo = 'ADMINISTRADOR'
ON CONFLICT (usuario_id, rol_id) DO NOTHING;

-- 4. Backfill: asignar USUARIO al resto de usuarios existentes que no tengan rol
INSERT INTO usuario_roles (usuario_id, rol_id)
SELECT u.id, r.id
FROM usuarios u
JOIN roles_sistema r ON r.codigo = 'USUARIO'
WHERE u.deleted_at IS NULL
  AND NOT EXISTS (
      SELECT 1 FROM usuario_roles ur WHERE ur.usuario_id = u.id
  )
ON CONFLICT (usuario_id, rol_id) DO NOTHING;
