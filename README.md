# Tesserak

Plataforma SaaS multi-empresa para gestión comercial y operativa (empresas, usuarios, CRM, terceros, productos, inventario, ventas, compras y lealtad).

- **Backend:** Go + Gin + GORM + PostgreSQL 14+
- **Frontend:** Angular 22 + TypeScript + SCSS + Vitest
- **Migraciones:** SQL versionado con `golang-migrate` (embebidas en el binario)

---

## Estructura del repositorio

```text
.
├── Diseño aplicación.drawio          # Diagrama ER completo del sistema (47 tablas, 8 dominios)
├── documentacion base de datos.md    # Arquitectura, patrones de diseño y catálogo de tablas
├── README.md                         # Documentación general y estado del proyecto (este archivo)
└── backend/
    ├── cmd/api/main.go               # Punto de entrada HTTP + ejecución de migraciones
    ├── migrations/
    │   ├── 000001_init_usuarios.*    # Geografía, empresas, sucursales y usuarios
    │   ├── 000002_init_crm.*         # Monedas y módulo CRM completo
    │   └── embed.go                  # Embebe los .sql en el binario
    ├── internal/
    │   ├── config/                   # Carga de variables de entorno
    │   ├── database/                 # Conexión GORM (pgx/v5) y runner de migraciones
    │   ├── httpx/                    # Manejo uniforme de errores y validación HTTP
    │   ├── empresas/                 # CRUD de empresas (soft delete)
    │   ├── usuarios/                 # CRUD de usuarios (1 sucursal, email único por empresa, bcrypt)
    │   └── crm/                      # Catálogos CRM, leads, pipeline, actividades, intereses y resumen
    ├── docker-compose.yml            # PostgreSQL 17 local
    ├── .env.example                  # Variables de entorno de ejemplo
    └── README.md                     # Guía rápida técnica del backend
```

---

## Estado actual de lo construido

| Módulo | Estado | Detalle |
|---|---|---|
| **Análisis y diseño de BD** | Completado | Revisión de las 47 tablas del diagrama, corrección de inconsistencias y documentación en [documentacion base de datos.md](documentacion%20base%20de%20datos.md). |
| **Infraestructura backend** | Completado | Servidor Gin, configuración por env, conexión GORM a PostgreSQL, migraciones automáticas, healthcheck (`/health`) y cierre limpio (*graceful shutdown*). |
| **Migraciones SQL** | Completado (`000001` y `000002`) | 12 tablas creadas con FKs compuestas multi-tenant, soft delete, índices únicos parciales y constraints `CHECK`. |
| **CRUD Empresas** | Completado | `/api/v1/empresas` con filtros, paginación, identificador fiscal único activo y soft delete. |
| **CRUD Usuarios** | Completado | `/api/v1/usuarios`, usuario vinculado a una sucursal, email y username únicos por empresa, `bcrypt`, soft delete. |
| **Autenticación (Backend)** | Completado | `POST /api/v1/auth/login` con `bcrypt`, JWT HS256, soporte multi-empresa (`identificador_fiscal` opcional) y CORS. |
| **Módulo CRM** | Completado | `/api/v1/crm/*`: orígenes, estados del pipeline, motivos de descarte, leads, cambio de etapa, conversión, actividades/agenda, intereses en productos y resumen de KPIs. |
| **Frontend Angular (Login)** | Completado | Proyecto en `frontend/` con pantalla de login, `AuthService` (signals), interceptor Bearer, guards (`/login`, `/dashboard`) y pruebas con Vitest. |
| **Pruebas unitarias de validación** | Completado | 35 casos de prueba HTTP pasando (`usuarios`: 9, `empresas`: 9, `crm`: 17). |
| **Pruebas de integración con PostgreSQL** | Pendiente | Pendiente ejecutar con Docker Desktop encendido para validar migraciones y consultas en vivo. |

---

## Resumen de endpoints implementados

### Salud del servicio
- `GET /health` — Verifica conectividad con PostgreSQL.

### Empresas (`/api/v1/empresas`)
- `POST /api/v1/empresas` — Crear empresa.
- `GET /api/v1/empresas` — Listar (`activo`, `q`, `page`, `page_size`).
- `GET /api/v1/empresas/:id` — Consultar por ID.
- `PATCH /api/v1/empresas/:id` — Actualización parcial.
- `DELETE /api/v1/empresas/:id` — Soft delete (`204`).

### Usuarios (`/api/v1/usuarios`)
- `POST /api/v1/usuarios` — Crear usuario (deduce `empresa_id` desde `sucursal_id`).
- `GET /api/v1/usuarios` — Listar (`empresa_id`, `sucursal_id`, `activo`, `q`, `page`, `page_size`).
- `GET /api/v1/usuarios/:id` — Consultar por ID.
- `PATCH /api/v1/usuarios/:id` — Actualización parcial (no permite mover a una sucursal de otra empresa).
- `DELETE /api/v1/usuarios/:id` — Soft delete (`204`).

### CRM (`/api/v1/crm`)
- `GET /api/v1/crm/resumen?empresa_id=` — KPIs del embudo (totales, conversión, valor estimado, desglose por etapa y temperatura, actividades pendientes y vencidas).
- `POST / GET / GET :id / PATCH :id / DELETE :id` en:
  - `/api/v1/crm/origenes`
  - `/api/v1/crm/estados`
  - `/api/v1/crm/motivos-descarte`
- **Leads:**
  - `POST /api/v1/crm/leads`, `GET /api/v1/crm/leads?empresa_id=`, `GET / PATCH / DELETE /api/v1/crm/leads/:id`
  - `PATCH /api/v1/crm/leads/:id/estado` — Transición de etapa y motivo de descarte.
  - `POST /api/v1/crm/leads/:id/convertir` — Conversión de lead a cliente/tercero.
- **Actividades e intereses:**
  - `POST / GET /api/v1/crm/leads/:id/actividades`
  - `GET /api/v1/crm/actividades` (agenda global con filtro `solo_vencidas=true`)
  - `PATCH / DELETE /api/v1/crm/actividades/:id`
  - `POST / GET /api/v1/crm/leads/:id/intereses` y `DELETE /api/v1/crm/leads/:id/intereses/:interesId`

---

## Cómo ejecutar localmente

### Opción A: Todo dockerizado (recomendado)

Requisitos: **Docker Desktop**.

```powershell
# Desde la raíz del repositorio: levanta PostgreSQL, Backend Go y Frontend Angular
docker compose up -d --build
```

- **Frontend:** `http://localhost:4200`
- **Backend API:** `http://localhost:8080/health`
- **PostgreSQL:** `localhost:5432` (BD `tesserak`, usuario/contraseña `tesserak`)

Para detener todo:
```powershell
docker compose down
```

### Opción B: Desarrollo local sin contenedores de app

Requisitos: **Go 1.22+**, **Node 22+** y **Docker Desktop** (o PostgreSQL 14+).

```powershell
# 1. Ir al directorio del backend
cd backend

# 2. Levantar PostgreSQL
docker compose up -d --wait

# 3. Configurar la conexión
$env:DATABASE_URL = "postgres://tesserak:tesserak@localhost:5432/tesserak?sslmode=disable"

# 4. Iniciar la API (aplica las migraciones automáticamente)
go run ./cmd/api
```

Opciones útiles del binario:
- `go run ./cmd/api -migrate` — Solo aplica migraciones y termina.
- `go run ./cmd/api -skip-migrate` — Inicia el servidor sin ejecutar migraciones.

---

## Próximos pasos sugeridos

1. **Sucursales y Geografía:** CRUD de `sucursales` (y carga inicial de `paises`, `ciudades`, `monedas`) para operar todo sin inserts manuales.
2. **Autenticación y RBAC:** Login con JWT, middleware de sesión/tenant, roles (`roles_sistema`, `permisos`) y auditoría (`created_by`, `updated_by`, `deleted_by`).
3. **Terceros:** `tercero_maestro`, `tercero_empresa`, `tercero_empresa_roles`, `clientes`, `proveedores`, `empleados` (y conectar la FK de conversión desde `crm_leads`).
4. **Productos e Inventario:** Catálogo maestro, `empresa_productos`, `ubicaciones`, kardex y stock (y conectar FKs desde `crm_lead_interes_productos`).
5. **Ventas, Compras y Lealtad.**
