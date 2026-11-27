# Tesserak

Plataforma SaaS multi-empresa para gestión comercial y operativa (empresas, usuarios, CRM, campañas de marketing, terceros, productos, inventario, ventas, compras y lealtad).

Este documento explica **cómo está diseñada y construida la aplicación** para que cualquier desarrollador pueda clonar el repositorio, levantar el entorno local y extender módulos siguiendo exactamente las mismas convenciones del proyecto.

---

## 1. Stack tecnológico

| Capa | Tecnología | Versión / Detalle |
|---|---|---|
| **Backend** | Go + Gin + GORM | Go 1.27, `gin-gonic/gin`, `gorm.io/gorm`, driver `pgx/v5` |
| **Base de datos** | PostgreSQL | 14+ (Docker usa `postgres:17-alpine`) |
| **Migraciones** | `golang-migrate/migrate/v4` | Archivos SQL embebidos en el binario (`backend/migrations`) |
| **Autenticación** | JWT (HS256) + `bcrypt` | Implementación propia sin dependencias pesadas en `internal/auth` |
| **Frontend** | Angular + TypeScript + SCSS | Angular 22 (componentes *standalone*, Signals, Reactive Forms) |
| **Pruebas** | `go test` + Vitest | Pruebas HTTP/unitarias e integración en Go; Vitest en Angular |
| **Contenedores** | Docker + Docker Compose | PostgreSQL, API Go y Frontend Angular servido con Nginx |

---

## 2. Arquitectura general

La aplicación sigue una arquitectura **monorepo cliente-servidor** con separación clara por dominios:

```text
┌──────────────────────────────────────────────────────────────────────┐
│                         Frontend (Angular 22)                        │
│  MainLayout (sidebar + submenús) → Rutas Lazy-Loaded por feature     │
│  AuthService (Signals) + authInterceptor (Bearer JWT) + authGuard    │
└──────────────────────────────────┬───────────────────────────────────┘
                                   │ HTTP/JSON (/api/v1/*)
┌──────────────────────────────────▼───────────────────────────────────┐
│                          Backend (Go + Gin)                          │
│  cmd/api/main.go → CORS + Logger + Recovery + Router /api/v1         │
│  ┌────────────────────────────────────────────────────────────────┐  │
│  │  internal/<modulo>/                                            │  │
│  │    handler.go     → Binding HTTP, validación, claims JWT       │  │
│  │    service.go     → Reglas de negocio, transacciones, errores  │  │
│  │    repository.go  → Consultas GORM acotadas por empresa_id     │  │
│  │    models.go      → Entidades GORM (TableName, gorm.DeletedAt) │  │
│  │    dto.go         → Request/Response DTOs y mappers            │  │
│  └────────────────────────────────────────────────────────────────┘  │
└──────────────────────────────────┬───────────────────────────────────┘
                                   │ pgx/v5 + GORM
┌──────────────────────────────────▼───────────────────────────────────┐
│                         PostgreSQL 14+ / 17                          │
│  Multi-tenant por empresa_id · FKs compuestas · Soft Delete          │
│  Migraciones SQL embebidas aplicadas automáticamente al iniciar      │
└──────────────────────────────────────────────────────────────────────┘
```

---

## 3. Estructura del repositorio

```text
.
├── README.md                              # Documentación general y arquitectura (este archivo)
├── documentacion base de datos.md         # Patrones de diseño de BD y catálogo completo de tablas
├── diseno_base_datos.drawio               # Diagrama ER base
├── tesserak_modelo_corregido.drawio       # Diagrama ER actualizado (incluye CRM, Productos, Precios, Terceros, Compras, Inventarios y Ventas)
├── docker-compose.yml                     # Orquestación completa: PostgreSQL + Backend + Frontend
│
├── backend/
│   ├── cmd/api/main.go                    # Arranque HTTP, CORS, /health, migraciones y graceful shutdown
│   ├── Dockerfile                         # Build multi-stage Go
│   ├── docker-compose.yml                 # PostgreSQL 17 aislado para desarrollo local
│   ├── .env.example                       # Plantilla de variables de entorno
│   ├── go.mod / go.sum
│   ├── migrations/                        # Migraciones SQL versionadas (000001 a 000031, up/down)
│   │   ├── embed.go                       # //go:embed *.sql
│   │   ├── 000001_init_usuarios.*         # paises, ciudades, empresas, sucursales, usuarios
│   │   ├── 000002_init_crm.*              # monedas, orígenes, estados, motivos, leads, actividades, intereses
│   │   ├── 000003_add_nombre_paises.*     # Columna nombre en paises
│   │   ├── 000004_add_usuario_facebook_crm_leads.* # Campo opcional usuario_facebook en crm_leads
│   │   ├── 000005_init_crm_campanas.*     # crm_tipos_campana, crm_campanas, crm_campana_leads, crm_campana_gastos
│   │   ├── 000006_init_crm_etapas_oportunidad.*    # crm_etapa_oportunidad (renombrada en 000007)
│   │   ├── 000007_init_crm_oportunidades.*         # crm_estado_oportunidades y crm_oportunidades
│   │   ├── 000008_init_productos.*                 # unidades_medida, categorias_producto, marcas, grupos_impuesto, productos
│   │   ├── 000009_refactor_articulos_productos_servicios.* # articulos -> productos / servicios + catálogos
│   │   ├── 000010_add_key_uuid_usuarios.*          # Columna key UUID v4 en usuarios
│   │   ├── 000011_init_listas_precios.*            # listas_precios y lista_precios_detalle
│   │   ├── 000012_init_compras.*                   # com_compras y com_compras_detalle
│   │   ├── 000013_init_roles_tercero.*             # roles_tercero (semilla global: CLIENTE, PROVEEDOR, EMPLEADO)
│   │   ├── 000014_init_tercero_maestro.*           # tercero_maestro (aislado por empresa_id)
│   │   ├── 000015_init_tercero_empresa_roles.*     # tercero_empresa_roles (N:M terceros ↔ roles)
│   │   ├── 000016_init_secuencias_documento.*      # tipos_documento y secuencias_documento (consecutivos automáticos)
│   │   ├── 000017_init_inventario.*                # inv_ubicaciones, inv_tipos_movimiento, inv_documentos, inv_movimientos_detalle, inv_stock
│   │   ├── 000018_backfill_compras_a_inventario.*  # Migración de compras históricas hacia inv_documentos / inv_stock
│   │   ├── 000019_init_ventas.*                    # ven_ventas, ven_ventas_detalle y backfill Consumidor Final (222222222222)
│   │   ├── 000020_usuario_sucursales.*             # Tabla intermedia usuario_sucursales (N:M) y eliminación de usuarios.sucursal_id
│   │   ├── 000021_init_roles_sistema.*             # roles_sistema (ADMINISTRADOR, USUARIO) y usuario_roles (N:M)
│   │   ├── 000022_add_rol_admin_sucursal.*         # Rol ADMINISTRADOR_SUCURSAL
│   │   ├── 000023_add_rol_gestor_ventas.*          # Rol GESTOR_VENTAS
│   │   ├── 000024_add_rol_asesor_comercial.*       # Rol ASESOR_COMERCIAL
│   │   ├── 000025_update_rol_asesor_comercial_solo_crm.* # ASESOR_COMERCIAL exclusivo CRM + permisos multi-rol aditivos
│   │   ├── 000026_add_rol_gestor_compras.*         # Rol GESTOR_COMPRAS
│   │   ├── 000027_init_lealtad_programas.*         # lealtad_programas, lealtad_programa_empresas y semilla "Lealtad Tesserak" (ipalenciab)
│   │   ├── 000028_init_lealtad_reglas.*            # lealtad_reglas y regla estándar de Lealtad Tesserak (1 pt / $1.000)
│   │   ├── 000029_init_lealtad_miembros_cuentas.*  # lealtad_miembros (identidad global) y lealtad_cuentas (saldos)
│   │   ├── 000030_init_lealtad_transacciones_redenciones.* # lealtad_transacciones (ledger) y lealtad_redenciones
│   │   └── 000031_lealtad_ventas_redencion_y_vencimientos.* # puntos_redimidos / descuento_lealtad en ven_ventas y vencimiento_procesado
│   └── internal/
│       ├── auth/                          # Login, bcrypt, JWT HS256, roles aditivos, aislamiento por sucursal y middlewares
│       ├── config/                        # Carga de .env y variables de entorno
│       ├── consecutivos/                  # Secuencias configurables (COMPRA, VENTA, INVENTARIO) con bloqueo transaccional
│       ├── database/                      # Conexión GORM y runner de golang-migrate
│       ├── httpx/                         # AppError uniforme y formateo de errores de validación
│       ├── empresas/                      # CRUD de empresas y sucursales (con ubicación por defecto BOD-<id> y Consumidor Final)
│       ├── usuarios/                      # CRUD de usuarios (N:M sucursales y roles_sistema), /mi-perfil, bcrypt, key UUID, soft delete
│       ├── productos/                     # Artículos (productos físicos y servicios), categorías, marcas, impuestos y catálogos
│       ├── precios/                       # Listas de precios y detalle por artículo
│       ├── terceros/                      # Terceros por empresa, asignación de roles globales y Consumidor Final
│       ├── compras/                       # CRUD de compras acotado por sucursales autorizadas y entrada atómica a inventario
│       ├── inventario/                    # Ubicaciones, stock materializado (inv_stock), ajustes y traslados acotados por sucursal
│       ├── ventas/                        # CRUD de ventas acotado por sucursales, IVA automático, salida de stock y lealtad (acumulación/redención/reverso)
│       ├── lealtad/                       # Módulo de lealtad híbrido: planes, adhesiones, reglas, miembros, cuentas, ledger, redenciones, conciliación y vencimientos
│       └── crm/                           # Módulo CRM completo (leads, oportunidades, campañas, configuración) acotado por sucursal
│
└── frontend/
    ├── Dockerfile / nginx.conf            # Build Angular + servidor estático Nginx
    ├── package.json / angular.json
    └── src/
        ├── environments/environment.ts    # URL base de la API (http://localhost:8080/api/v1)
        ├── styles.scss                    # Variables CSS globales (--primary, --bg-page, etc.)
        └── app/
            ├── app.routes.ts              # Rutas lazy-loaded protegidas por guards de autenticación y roles aditivos
            ├── app.config.ts              # provideRouter, provideHttpClient(withInterceptors([authInterceptor]))
            ├── core/auth/                 # AuthService (signals y capacidades aditivas), auth.interceptor, guards, auth.models
            ├── layout/main-layout/        # Shell principal: sidebar dinámico por roles, tarjeta clickeable a /mi-perfil, topbar
            └── features/
                ├── auth/login/            # Pantalla de inicio de sesión
                ├── dashboard/             # Panel principal con KPIs en vivo del CRM y accesos rápidos
                ├── coming-soon/           # Pantalla placeholder para módulos pendientes
                ├── administracion/
                │   ├── gestion-empresa/   # /gestion-empresa (consecutivos, CRUD sucursales, ubicaciones por sucursal, key UUID)
                │   └── planes-lealtad/    # /administracion/planes-lealtad (solo ADMINISTRADOR: planes, adhesiones, reglas, conciliación y vencimientos)
                ├── lealtad/
                │   └── miembros-cuentas/  # /lealtad (miembros globales, inscripción desde Terceros, cuentas, extracto, acumulación y redención)
                ├── usuarios/              # /usuarios (CRUD multi-rol y multi-sucursal) y /mi-perfil (datos propios + contraseña)
                ├── crm/                   # /crm/{leads, oportunidades, campanas, configuracion, resumen}
                ├── productos/             # /productos/{catalogo, categorias, configuracion}
                ├── precios/
                │   └── listas-precios/    # /productos/listas-precios
                ├── terceros/              # /terceros (terceros con chips de roles globales de solo lectura)
                ├── compras/               # /compras (proveedores filtrados, consecutivo automático, soft delete)
                ├── inventario/            # /inventario (vista por ubicación -> stock, ajustes, pérdidas, averías, devoluciones y traslados)
                └── ventas/                # /ventas (precios automáticos desde listas de precios, redención de puntos, Consumidor Final, IVA, salida de stock)
```

---

## 4. Patrones y convenciones de diseño

Para mantener consistencia al agregar nuevas funcionalidades, el proyecto sigue estas reglas:

### 4.1. Multi-tenancy e integridad referencial
- Todas las tablas operativas incluyen `empresa_id` como discriminador de tenant.
- Para impedir a nivel de motor que un registro de la empresa A referencie entidades de la empresa B, se utilizan **claves foráneas compuestas** `(entidad_id, empresa_id)` apoyadas en restricciones `UNIQUE (id, empresa_id)`.
- En los endpoints protegidos por JWT (`auth.RequireAuth`), el `empresa_id` y el `usuario_id` (`claims.Sub`) se obtienen **exclusivamente del token autenticado**, nunca del cuerpo o query enviado por el cliente.

### 4.2. Soft Delete + índices únicos parciales
- Los catálogos y entidades principales (`empresas`, `sucursales`, `usuarios`, `crm_origen_lead`, `crm_estado_lead`, `crm_etapa_oportunidad`, `crm_motivo_descarte`, `crm_leads`, `crm_lead_actividades`, `crm_tipos_campana`, `crm_campanas`, `crm_campana_gastos`) usan `deleted_at TIMESTAMPTZ` mapeado con `gorm.DeletedAt`.
- Al eliminar con `DELETE`, GORM realiza un **soft delete** (`deleted_at = now()`), conservando integridad histórica.
- Los índices de unicidad se declaran como **índices únicos parciales** (`WHERE deleted_at IS NULL`), permitiendo reutilizar nombres, emails o NITs de registros eliminados.

### 4.3. Arquitectura por capas en Go
Cada paquete de dominio en `backend/internal/<modulo>` separa responsabilidades:
1. **`handler.go`**: parsea parámetros de ruta/query/JSON con Gin, valida tags `binding:"..."` vía `httpx.BindError`, extrae claims JWT si aplica y delega al servicio.
2. **`service.go`**: contiene validaciones de dominio, transiciones de estado, manejo de transacciones (`repo.WithTx`) y traducción de errores PostgreSQL (`*pgconn.PgError`) a `httpx.AppError`.
3. **`repository.go`**: encapsula consultas GORM y SQL explícito. GORM se configura sin `TranslateError` para poder inspeccionar `pgErr.ConstraintName` y devolver errores HTTP precisos (`409`, `422`).
4. **`models.go` / `dto.go`**: los modelos GORM nunca se serializan directo al cliente; siempre se mapean a DTOs de respuesta (`to*Response`).

### 4.4. Manejo uniforme de errores HTTP
Todos los errores responden con el mismo sobre JSON definido en [errors.go](backend/internal/httpx/errors.go):
```json
{
  "error": {
    "code": "codigo_de_error",
    "message": "Descripción legible para el cliente",
    "details": []
  }
}
```

### 4.5. Convenciones del Frontend Angular
- **Standalone Components** en toda la aplicación (sin `NgModules`).
- **Estado reactivo con Signals** (`signal`, `computed`) y flujo de control nativo (`@if`, `@for`).
- **Rutas con Lazy Loading** (`loadComponent`) definidas en [app.routes.ts](frontend/src/app/app.routes.ts).
- **AuthInterceptor** adjunta automáticamente `Authorization: Bearer <token>` a todas las peticiones HTTP y limpia sesión si recibe `401`.
- **Variables CSS centralizadas** en [styles.scss](frontend/src/styles.scss) (`--primary`, `--bg-page`, `--bg-surface`, `--border-color`, `--text-primary`, `--text-secondary`).

---

## 5. Modelo de datos y migraciones

Las migraciones están en `backend/migrations/`, se embeben en el binario mediante `embed.FS` y se ejecutan automáticamente al iniciar la API (salvo que se pase `-skip-migrate`).

| Migración | Descripción | Tablas / Cambios |
|---|---|---|
| `000001` | Esquema base multi-empresa | `paises`, `ciudades`, `empresas`, `sucursales`, `usuarios` |
| `000002` | Monedas y núcleo del CRM | `monedas`, `crm_origen_lead`, `crm_estado_lead`, `crm_motivo_descarte`, `crm_leads`, `crm_lead_actividades`, `crm_lead_interes_productos` |
| `000003` | Nombre corto en países | Agrega columna `nombre` en `paises` |
| `000004` | Contacto social en leads | Agrega columna opcional `usuario_facebook VARCHAR(100)` en `crm_leads` |
| `000005` | Módulo de campañas de marketing | `crm_tipos_campana` (con 7 tipos semilla), `crm_campanas`, `crm_campana_leads` (atribución + UTM), `crm_campana_gastos` y `uq_crm_leads_id_empresa` |
| `000006` | Etapas de oportunidades | `crm_etapa_oportunidad` (renombrada en `000007`) |
| `000007` | Oportunidades comerciales | Renombra a `crm_estado_oportunidades` y crea `crm_oportunidades` |
| `000008` | Catálogos de productos por empresa | `unidades_medida`, `categorias_producto`, `marcas`, `grupos_impuesto`, `productos` |
| `000009` | Supertipo–Subtipo de artículos | `tipos_articulo`, `metodos_valoracion`, `modalidades_servicio`, `articulos` → `productos` (1:1) / `servicios` (1:1, duración en horas), `grupo_impuesto_id` obligatorio y `marca_id` opcional |
| `000010` | Clave pública UUID v4 en usuarios | Agrega columna `key UUID NOT NULL DEFAULT gen_random_uuid()` con índice único `ux_usuarios_key` en `usuarios` |
| `000011` | Listas de precios | `listas_precios` y `lista_precios_detalle` |
| `000012` | Módulo de compras | `com_compras` y `com_compras_detalle` (con soft delete) |
| `000013` | Roles globales de tercero | `roles_tercero` con semilla de sistema (`CLIENTE`, `PROVEEDOR`, `EMPLEADO`) |
| `000014` | Maestro de terceros por empresa | `tercero_maestro` con FK directa a `empresas` y unicidad por identificación |
| `000015` | Relación terceros ↔ roles | `tercero_empresa_roles` (N:M) |
| `000016` | Consecutivos configurables | `tipos_documento` (`COMPRA`, `VENTA`, `INVENTARIO`) y `secuencias_documento` |
| `000017` | Módulo de inventarios (`inv_`) | `inv_ubicaciones`, `inv_tipos_movimiento` (9 tipos semilla), `inv_documentos`, `inv_movimientos_detalle`, `inv_stock` (stock materializado + costo promedio) |
| `000018` | Backfill de compras a inventario | Genera movimientos `ENTRADA_COMPRA` e inicializa `inv_stock` para compras históricas existentes |
| `000019` | Módulo de ventas y Consumidor Final | `ven_ventas`, `ven_ventas_detalle` (soft delete) y backfill de `Consumidor Final` (`CC 222222222222`, rol `CLIENTE`) por empresa |
| `000020` | Relación N:M usuarios ↔ sucursales | Crea tabla intermedia `usuario_sucursales` (con FKs compuestas y `es_principal`), migra `usuarios.sucursal_id` y elimina la FK y columna `sucursal_id` de `usuarios` |
| `000021` | Roles de sistema y rol Administrador | `roles_sistema` (semilla `ADMINISTRADOR` y `USUARIO`), `usuario_roles` (N:M) y backfill de `ADMINISTRADOR` al primer usuario de cada empresa |
| `000022` | Rol Administrador de sucursal | Inserta rol de sistema `ADMINISTRADOR_SUCURSAL` con acceso administrativo acotado a sus sucursales asignadas |
| `000023` | Rol Gestor de ventas | Inserta rol de sistema `GESTOR_VENTAS` (solo crear/modificar ventas y crear/asignar rol `CLIENTE` en terceros) |
| `000024` | Rol Asesor comercial | Inserta rol de sistema `ASESOR_COMERCIAL` en `roles_sistema` |
| `000025` | Asesor comercial solo CRM + multi-rol | Actualiza `ASESOR_COMERCIAL` a acceso exclusivo de CRM, combinable con múltiples roles (ej. `GESTOR_VENTAS`) |
| `000026` | Rol Gestor de compras | Inserta rol de sistema `GESTOR_COMPRAS` (consultar, crear y modificar compras, combinable con otros roles) |
| `000027` | Planes de lealtad y adhesiones | `lealtad_programas`, `lealtad_programa_empresas` y semilla global **Lealtad Tesserak** (`LEALTAD_TESSERAK`) administrada por `ipalenciab` |
| `000028` | Reglas de acumulación de lealtad | `lealtad_reglas` (globales y overrides por empresa) y regla base de **Lealtad Tesserak** (`1 pt / $1.000`) |
| `000029` | Miembros y cuentas de lealtad | `lealtad_miembros` (identidad global por documento) y `lealtad_cuentas` (billeteras de puntos) |
| `000030` | Libro mayor y redenciones de lealtad | `lealtad_transacciones` (append-only ledger) y `lealtad_redenciones` (1:1 con movimiento) |
| `000031` | Redención en ventas y vencimientos | Agrega `puntos_redimidos` y `descuento_lealtad` a `ven_ventas`, y `vencimiento_procesado` a `lealtad_transacciones` |

El diagrama relacional completo y actualizado está en [tesserak_modelo_corregido.drawio](tesserak_modelo_corregido.drawio) (editable en [diagrams.net](https://app.diagrams.net/)) y la documentación detallada de convenciones de BD en [documentacion base de datos.md](documentacion%20base%20de%20datos.md).

---

## 6. Referencia de API REST

Base URL local: `http://localhost:8080`

### Salud y Autenticación
- `GET /health` — Verifica conectividad con PostgreSQL.
- `POST /api/v1/auth/login` — Autenticación por email o usuario + password (+ `identificador_fiscal` opcional). Retorna JWT HS256 y datos del usuario (`roles`, flags de rol, `sucursal_ids` autorizadas y `sucursal_id` principal). Para `ADMINISTRADOR`, resuelve automáticamente todas las sucursales activas de su empresa.

### Empresas (`/api/v1/empresas`), Sucursales (`/api/v1/empresas/sucursales`) y Usuarios (`/api/v1/usuarios`)
- **Empresas:** `POST`, `GET` (filtros + paginación), `GET /:id`, `PATCH /:id`, `DELETE /:id` (soft delete). Al crear una empresa se crea automáticamente **Consumidor Final** (`222222222222`, rol único `CLIENTE`).
- **Sucursales (requieren JWT; `POST/PATCH/DELETE` exclusivos `ADMINISTRADOR`):**
  - `GET /api/v1/empresas/sucursales` — Lista las sucursales de la empresa (acotado a sucursales autorizadas para no administradores).
  - `POST /api/v1/empresas/sucursales` — Crea una sucursal y **genera automáticamente** su ubicación por defecto `BOD-<sucursal_id>` (`Bodega Principal - <nombre>`) en `inv_ubicaciones` dentro de la misma transacción.
  - `PATCH /api/v1/empresas/sucursales/:id` y `DELETE /api/v1/empresas/sucursales/:id` (soft delete; impide eliminar la única sucursal de la empresa).
- **Usuarios y Roles de Sistema:**
  - `GET /api/v1/usuarios/roles` — Catálogo de roles de sistema asignables (`ADMINISTRADOR`, `ADMINISTRADOR_SUCURSAL`, `GESTOR_VENTAS`, `ASESOR_COMERCIAL`, `GESTOR_COMPRAS`, `USUARIO`).
  - `GET /api/v1/usuarios/sucursales` — Sucursales activas accesibles por el usuario autenticado.
  - `GET / PATCH /api/v1/usuarios/mi-perfil` — Autogestión del perfil propio (`nombres`, `apellidos`, `nombre_usuario`, `email`) y cambio de contraseña verificando obligatoriamente `password_actual`.
  - `POST / GET / GET :id / PATCH :id / DELETE :id` en `/api/v1/usuarios` — Soportan múltiples roles combinables (`rol_ids`) y múltiples sucursales (`sucursal_ids`). El primer usuario de cada empresa recibe automáticamente `ADMINISTRADOR`; a partir de ahí solo `ADMINISTRADOR` (toda la empresa) o `ADMINISTRADOR_SUCURSAL` (solo dentro de sus sucursales asignadas) pueden gestionar usuarios.
  - `POST /api/v1/usuarios/:id/regenerar-key` — Regenera la `key` (UUID v4) del usuario.

### Roles de Sistema (RBAC Multi-rol Aditivo) y Aislamiento por Sucursal
Los usuarios pueden tener **uno o varios roles** en `usuario_roles`; los permisos se suman entre sí y todos los módulos operativos (`ventas`, `compras`, `inventario`, `crm`, `usuarios`, `sucursales`) filtran los datos únicamente a las sucursales autorizadas del usuario:

| Rol (`codigo`) | Alcance y permisos |
|---|---|
| `ADMINISTRADOR` | Acceso total a la empresa, todas las sucursales, consecutivos, CRUD de sucursales, **administración de planes de lealtad**, gestión de empresa y usuarios. Asignado automáticamente al primer usuario de cada empresa. |
| `ADMINISTRADOR_SUCURSAL` | Acceso administrativo a todos los módulos operativos, ubicaciones y gestión de usuarios únicamente dentro de sus sucursales asignadas (sin acceso a administración de planes de lealtad, consecutivos ni sucursales de empresa; no puede asignar rol `ADMINISTRADOR`). |
| `GESTOR_VENTAS` | Consultar, crear y modificar ventas (sin eliminar) y gestionar terceros únicamente con el rol `CLIENTE` (crear clientes o asignar rol `CLIENTE` a terceros existentes, sin eliminar). |
| `ASESOR_COMERCIAL` | Acceso completo y exclusivo al módulo CRM (leads, oportunidades, campañas, actividades y configuración). Combinable con otros roles (ej. `GESTOR_VENTAS`). |
| `GESTOR_COMPRAS` | Consultar, crear y modificar compras (sin eliminar). Combinable con otros roles. |
| `USUARIO` | Acceso operativo estándar acotado a sus sucursales asignadas. |

### CRM — Núcleo y Catálogos (`/api/v1/crm`)
- `GET /api/v1/crm/resumen?empresa_id=` — KPIs del embudo, conversión, valor estimado, etapas, temperatura y actividades pendientes/vencidas.
- **Catálogos de configuración** (`POST`, `GET`, `GET /:id`, `PATCH /:id`, `DELETE /:id` con soft delete):
  - `/api/v1/crm/origenes` — Orígenes de leads (`crm_origen_lead`)
  - `/api/v1/crm/estados` — Etapas de leads (`crm_estado_lead`)
  - `/api/v1/crm/etapas-oportunidad` — Etapas de oportunidades (`crm_etapa_oportunidad`)
  - `/api/v1/crm/motivos-descarte` — Motivos de descarte (`crm_motivo_descarte`)
- **Leads**:
  - `POST /api/v1/crm/leads`, `GET /api/v1/crm/leads?empresa_id=`, `GET / PATCH / DELETE /api/v1/crm/leads/:id`
  - `PATCH /api/v1/crm/leads/:id/estado` — Cambio de etapa y motivo de descarte.
  - `POST /api/v1/crm/leads/:id/convertir` — Marca el lead como convertido.
- **Actividades e intereses**:
  - `POST / GET /api/v1/crm/leads/:id/actividades`, `GET /api/v1/crm/actividades`, `PATCH / DELETE /api/v1/crm/actividades/:id`
  - `POST / GET /api/v1/crm/leads/:id/intereses`, `DELETE /api/v1/crm/leads/:id/intereses/:interesId`

### CRM — Campañas de Marketing (requieren `Authorization: Bearer <jwt>`)
- **Tipos de campaña**: `POST / GET / GET :id / PATCH :id / DELETE :id` en `/api/v1/crm/tipos-campana`
- **Campañas**:
  - `POST /api/v1/crm/campanas`, `GET /api/v1/crm/campanas`, `GET /:id`, `PATCH /:id`
  - `PATCH /api/v1/crm/campanas/:id/estado`, `POST /:id/activar`, `POST /:id/pausar`, `POST /:id/finalizar`
  - `GET /api/v1/crm/campanas/:id/indicadores` — Presupuesto, gasto acumulado, disponible, total de leads y desglose por moneda.
- **Atribución Campaña ↔ Leads**:
  - `POST / GET /api/v1/crm/campanas/:id/leads`
  - `DELETE /api/v1/crm/campanas/:id/leads/:leadId`
  - `GET /api/v1/crm/leads/:id/campanas`
- **Gastos de campaña**:
  - `POST / GET /api/v1/crm/campanas/:id/gastos`
  - `GET / PATCH / DELETE /api/v1/crm/gastos/:id`

### Productos y Servicios (`/api/v1/productos`, requieren `Authorization: Bearer <jwt>`)
- **Catálogos**:
  - `GET /api/v1/productos/tipos-articulo` (`PRODUCTO`, `SERVICIO`)
  - `GET /api/v1/productos/metodos-valoracion` (`FIFO`, `PROMEDIO_PONDERADO`, `IDENTIFICACION_ESPECIFICA`)
  - `GET /api/v1/productos/modalidades-servicio` (`POR_HORA`, `PROYECTO`, `UNIDAD`, `SUSCRIPCION`)
  - `GET /api/v1/productos/unidades-medida`
  - `GET /api/v1/productos/grupos-impuesto`
  - `POST / GET / GET :id / PATCH :id / DELETE :id` en `/api/v1/productos/categorias` (soft delete)
  - `POST / GET / PATCH :id / DELETE :id` en `/api/v1/productos/marcas` (soft delete)
- **Artículos (`articulos` + extensión `productos` o `servicios`)**:
  - `POST /api/v1/productos` — Crea artículo + subtipo en transacción (`codigo`, `nombre`, `tipo_articulo_id`, `unidad_medida_id`, `categoria_id`, `grupo_impuesto_id` obligatorios; `marca_id` opcional; bloque `producto` o `servicio` con `duracion_estimada` en horas).
  - `GET /api/v1/productos` — Listado paginado (`tipo_articulo_id`, `categoria_id`, `estado`, `q`, `page`, `page_size`).
  - `GET / PATCH / DELETE /api/v1/productos/:id` — Consulta, actualización transaccional y **soft delete**.

### Listas de Precios (`/api/v1/precios`, requieren JWT)
- `POST / GET / GET :id / PATCH :id / DELETE :id` en `/api/v1/precios/listas` (soft delete).
- `GET / PUT / DELETE` en `/api/v1/precios/listas/:id/detalles` (`/:articuloId`) — Gestión de precios por artículo dentro de cada lista.

### Terceros (`/api/v1/terceros`, requieren JWT)
- `GET /api/v1/terceros/roles` — Catálogo global de roles (`CLIENTE`, `PROVEEDOR`, `EMPLEADO`; de solo lectura en frontend).
- `POST / GET / GET :id / PATCH :id / DELETE :id` en `/api/v1/terceros` — CRUD aislado por `empresa_id` con asignación múltiple de roles (`rol_ids`) y soft delete.

### Consecutivos (`/api/v1/consecutivos`, requieren JWT)
- `GET /api/v1/consecutivos/tipos-documento` — Tipos de documento configurables (`COMPRA`, `VENTA`, `INVENTARIO`).
- `POST / GET / PATCH :id` en `/api/v1/consecutivos/secuencias` — Configuración de prefijo, siguiente número, relleno y frecuencia de reinicio (`NUNCA`, `ANUAL`, `MENSUAL`).

### Compras (`/api/v1/compras`, requieren JWT)
- `POST / GET / GET :id / PATCH :id / DELETE :id` en `/api/v1/compras` (soft delete).
- Genera consecutivo automático (`COMPRA`) y ejecuta `ENTRADA_COMPRA` en `inv_stock` dentro de la misma transacción (solo para productos con `control_inventario = TRUE`); si falla el movimiento de inventario se realiza rollback completo.

### Inventarios (`/api/v1/inventario`, requieren JWT)
- `GET /api/v1/inventario/sucursales` — Sucursales activas de la empresa (para asignar al crear/editar ubicaciones).
- `GET /api/v1/inventario/tipos-movimiento` — Catálogo global (`AJUSTE_POSITIVO`, `AJUSTE_NEGATIVO`, `PERDIDA`, `AVERIA`, `DEVOLUCION_PROVEEDOR`, `DEVOLUCION_CLIENTE`, `ENTRADA_COMPRA`, `SALIDA_VENTA`, `TRASLADO`).
- `POST / GET / PATCH :id / DELETE :id` en `/api/v1/inventario/ubicaciones` (soft delete).
- `GET /api/v1/inventario/stock` — Consulta de stock materializado en `inv_stock` por ubicación/artículo.
- `POST / GET / GET :id` en `/api/v1/inventario/movimientos` — Ajustes, pérdidas, averías, devoluciones y traslados internos entre ubicaciones con actualización atómica de `inv_stock`.

### Ventas (`/api/v1/ventas`, requieren JWT)
- `POST / GET / GET :id / PATCH :id / DELETE :id` en `/api/v1/ventas` (soft delete).
- Consecutivo automático (`VENTA`), cliente por defecto **Consumidor Final** (`222222222222`), cálculo automático de impuesto desde `grupos_impuesto.tarifa`, autocompletado de precios desde listas de precios en frontend y salida atómica `SALIDA_VENTA` de `inv_stock`.
- **Integración con Lealtad:** acepta `puntos_redimir` opcional en `POST /api/v1/ventas` (aplica `descuento_lealtad = puntos × valor_punto_redencion` y descuenta saldo con `FOR UPDATE`), acumula automáticamente puntos sobre el total neto pagado si el cliente coincide por `(pais_id, tipo_identificacion, numero_identificacion)` en un plan adherido, y genera reverso automático (`REVERSO`) al anular (`ANULADA`) o eliminar la venta.

### Lealtad (`/api/v1/lealtad`, requieren JWT)
- **Administración de planes (exclusivo `ADMINISTRADOR`; `ADMINISTRADOR_SUCURSAL` recibe `403`):**
  - `GET /api/v1/lealtad/monedas`
  - `GET / POST / GET :id / PATCH :id / DELETE :id` en `/api/v1/lealtad/programas` (plan global por defecto **Lealtad Tesserak** administrado por `ipalenciab`, protegido contra eliminación).
  - `POST / PATCH / DELETE` en `/api/v1/lealtad/programas/:id/adherir` (`/adhesion`) y `GET /api/v1/lealtad/programas/:id/empresas`.
  - `GET / POST / PATCH :reglaId / DELETE :reglaId` en `/api/v1/lealtad/programas/:id/reglas` (reglas globales del plan y overrides por empresa).
  - `GET /api/v1/lealtad/programas/:id/conciliacion` (balance emitidos vs. redimidos por empresa aliada) y `POST /api/v1/lealtad/programas/:id/procesar-vencimientos`.
- **Operación de lealtad (`CanManageOperacion` y consultas desde `CanAccessVentas`):**
  - `GET /api/v1/lealtad/operacion/programas`, `/cuenta-cliente/:terceroId`, `/paises`, `/clientes`.
  - `GET /api/v1/lealtad/miembros/buscar`, `POST /api/v1/lealtad/miembros`, `POST /api/v1/lealtad/miembros/desde-tercero`, `PATCH /api/v1/lealtad/miembros/:id`.
  - `GET / POST / PATCH :id/estado` en `/api/v1/lealtad/cuentas`.
  - `GET /api/v1/lealtad/cuentas/:id/transacciones`, `POST /api/v1/lealtad/cuentas/:id/acumular` y `POST /api/v1/lealtad/cuentas/:id/redimir`.

---

## 7. Módulos y navegación del Frontend

| Ruta | Estado | Descripción |
|---|---|---|
| `/login` | Implementado | Autenticación contra `/api/v1/auth/login` |
| `/dashboard` | Implementado | Resumen ejecutivo con KPIs de CRM, embudo y accesos rápidos |
| `/crm/resumen` | Base lista | Resumen comercial (lienzo reservado para métricas comerciales) |
| `/crm/leads` | Implementado | Administración completa de leads, campo `usuario_facebook` y asociación opcional a campaña + UTMs al crear/editar |
| `/crm/campanas` | Implementado | Administración de campañas, tipos de campaña, ciclo de vida, indicadores, leads asociados y gastos |
| `/crm/configuracion` | Implementado | Configuración CRUD (con soft delete) de Orígenes, Etapas, Etapas de oportunidades y Motivos de descarte |
| `/crm/empresas-contactos` | Base lista | Sección pendiente |
| `/crm/oportunidades` | Implementado | Administración de oportunidades (`crm_oportunidades` y `crm_estado_oportunidades`) |
| `/crm/pipeline` | Base lista (`Pronto`) | Sección pendiente |
| `/crm/actividades` | Base lista (`Pronto`) | Sección pendiente |
| `/productos/catalogo` | Implementado | Catálogo de productos físicos y servicios (`articulos` → `productos` / `servicios`) |
| `/productos/categorias` | Implementado | Administración CRUD (soft delete) de categorías de productos |
| `/productos/listas-precios` | Implementado | Listas de precios y asignación de precios por artículo |
| `/productos/configuracion` | Implementado | Configuración de marcas y grupos de impuesto (`grupos_impuesto`) |
| `/terceros` | Implementado | Maestro de terceros con asignación de roles globales de solo lectura (`CLIENTE`, `PROVEEDOR`, `EMPLEADO`) |
| `/compras` | Implementado | Registro de compras con proveedores filtrados, consecutivo automático, entrada a inventario y soft delete |
| `/inventario` | Implementado | Vista por ubicaciones → detalle de stock (`inv_stock`), ajustes, pérdidas, averías, devoluciones y traslados internos |
| `/ventas` | Implementado | Registro de ventas con precios automáticos desde listas de precios, redención de puntos de lealtad, Consumidor Final por defecto, cálculo automático de IVA y salida atómica de stock |
| `/lealtad` | Implementado | Módulo operativo de lealtad: registro de miembros globales, inscripción desde Terceros, billeteras de puntos, extracto de movimientos, acumulación/ajustes y redención |
| `/gestion-empresa` | Implementado | Prefijos de consecutivos y CRUD de sucursales (`ADMINISTRADOR`), ubicaciones/bodegas por sucursal (`ADMINISTRADOR` / `ADMINISTRADOR_SUCURSAL`) y key UUID de usuario |
| `/administracion/planes-lealtad` | Implementado | Administración exclusiva `ADMINISTRADOR`: plan global **Lealtad Tesserak** (`ipalenciab`), planes propios, adhesiones, reglas de acumulación, conciliación de coalición y procesamiento de vencimientos |
| `/usuarios` | Implementado | Administración de usuarios por empresa (`ADMINISTRADOR` y `ADMINISTRADOR_SUCURSAL` en su alcance): creación/edición, asignación de múltiples roles combinables y permisos por sucursal |
| `/mi-perfil` | Implementado | Autogestión del perfil propio desde la tarjeta del usuario en el sidebar: edición de datos personales y cambio de contraseña validando la contraseña actual |

---

## 8. Cómo clonar y ejecutar localmente

### Opción A: Todo con Docker Compose (recomendado)

Requisitos: **Docker Desktop**.

```powershell
git clone <url-del-repositorio>
cd "codigo fuennte tesserak"
docker compose up -d --build
```

Servicios levantados:
- **Frontend:** `http://localhost:4200`
- **Backend API:** `http://localhost:8080/health`
- **PostgreSQL:** `localhost:5432` (BD `tesserak`, usuario/clave `tesserak`)

Para detener:
```powershell
docker compose down
```

### Opción B: Desarrollo local (hot reload)

Requisitos: **Go 1.24+**, **Node.js 22+**, **PostgreSQL 14+** (o Docker solo para la BD).

1. **Levantar PostgreSQL**:
   ```powershell
   cd backend
   docker compose up -d --wait
   ```
2. **Configurar variables de entorno** (copiar `backend/.env.example` a `backend/.env` o exportar en consola):
   ```powershell
   $env:DATABASE_URL = "postgres://tesserak:tesserak@localhost:5432/tesserak?sslmode=disable"
   $env:JWT_SECRET   = "tesserak-dev-secret-cambiar-en-produccion"
   $env:CORS_ORIGIN  = "http://localhost:4200"
   ```
3. **Ejecutar el backend** (aplica migraciones automáticamente):
   ```powershell
   go run ./cmd/api
   ```
   Flags útiles:
   - `go run ./cmd/api -migrate` — Solo aplica migraciones y termina.
   - `go run ./cmd/api -skip-migrate` — Inicia el servidor sin correr migraciones.
4. **Ejecutar el frontend** (en otra terminal):
   ```powershell
   cd frontend
   npm install
   npm start
   ```
   Disponible en `http://localhost:4200`.

---

## 9. Cómo ejecutar las pruebas

### Backend (Go)
```powershell
cd backend

# Pruebas unitarias, de autenticación y validaciones HTTP
go test -v ./...

# Pruebas de integración contra PostgreSQL (campañas, aislamiento multiempresa y migraciones)
$env:TEST_DATABASE_URL = "postgres://tesserak:tesserak@localhost:5432/tesserak?sslmode=disable"
go test -v ./internal/crm -run TestIntegracionCampanas
```

### Frontend (Angular + Vitest)
```powershell
cd frontend

# Pruebas unitarias
npm test -- --watch=false

# Build de producción
npm run build
```

---

## 10. Guía rápida para extender el proyecto

- **Agregar una tabla o columna:** crear un par `00000N_descripcion.up.sql` y `.down.sql` en `backend/migrations/`, actualizar `tesserak_modelo_corregido.drawio` y respetar `empresa_id` + FK compuesta + `deleted_at` cuando aplique.
- **Agregar endpoints:** seguir el patrón `models -> dto -> repository -> service -> handler` dentro de `backend/internal/<modulo>/`, proteger nuevas rutas con `auth.RequireAuth(cfg.JWTSecret)` y extraer `empresa_id` desde `auth.ClaimsFromContext(c)`.
- **Agregar una pantalla en Angular:** crear un componente *standalone* en `frontend/src/app/features/<modulo>/`, registrar la ruta lazy en `app.routes.ts` y agregar la entrada en `navSections` de `main-layout.ts`.

---

## 11. Estado actual y hoja de ruta

- **Implementado:** Infraestructura base, Auth JWT con RBAC multi-rol aditivo (`ADMINISTRADOR`, `ADMINISTRADOR_SUCURSAL`, `GESTOR_VENTAS`, `ASESOR_COMERCIAL`, `GESTOR_COMPRAS`, `USUARIO`) y aislamiento por sucursales autorizadas (`usuario_sucursales`), Empresas (con creación automática de `Consumidor Final`), Sucursales (CRUD con ubicación por defecto `BOD-<id>`), Usuarios UI y Mi perfil (con `key` UUID y cambio de contraseña), Gestión de empresa (prefijos de consecutivos, sucursales, ubicaciones por sucursal, key de usuario), CRM (Leads, Oportunidades, Campañas y Configuración), Productos y Servicios (Catálogo, Categorías, Listas de precios, Marcas e Impuestos), Terceros, Compras, Inventarios (`inv_stock`, ajustes y traslados), Ventas (precios automáticos desde listas de precios, cálculo automático de IVA, descuento atómico de stock e integración con lealtad) y **Lealtad** (plan global **Lealtad Tesserak** administrado por `ipalenciab`, planes locales/coalición, adhesiones, reglas de acumulación, miembros globales, cuentas, libro mayor, redenciones en Ventas con reverso automático, conciliación de coalición y vencimientos).
- **Próximos módulos:** Pipeline comercial (Kanban), Agenda de Actividades en UI, Resumen comercial y módulos contratados por empresa (`modulos`/`modulo_empresa`).
