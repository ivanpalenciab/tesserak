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
├── tesserak_modelo_corregido.drawio       # Diagrama ER actualizado (incluye campañas y etapas de oportunidad)
├── docker-compose.yml                     # Orquestación completa: PostgreSQL + Backend + Frontend
│
├── backend/
│   ├── cmd/api/main.go                    # Arranque HTTP, CORS, /health, migraciones y graceful shutdown
│   ├── Dockerfile                         # Build multi-stage Go
│   ├── docker-compose.yml                 # PostgreSQL 17 aislado para desarrollo local
│   ├── .env.example                       # Plantilla de variables de entorno
│   ├── go.mod / go.sum
│   ├── migrations/                        # Migraciones SQL versionadas (up/down)
│   │   ├── embed.go                       # //go:embed *.sql
│   │   ├── 000001_init_usuarios.*         # paises, ciudades, empresas, sucursales, usuarios
│   │   ├── 000002_init_crm.*              # monedas, orígenes, estados, motivos, leads, actividades, intereses
│   │   ├── 000003_add_nombre_paises.*     # Columna nombre en paises
│   │   ├── 000004_add_usuario_facebook_crm_leads.* # Campo opcional usuario_facebook en crm_leads
│   │   ├── 000005_init_crm_campanas.*     # crm_tipos_campana, crm_campanas, crm_campana_leads, crm_campana_gastos
│   │   └── 000006_init_crm_etapas_oportunidad.*    # crm_etapa_oportunidad
│   └── internal/
│       ├── auth/                          # Login, bcrypt, JWT HS256 y middleware RequireAuth
│       ├── config/                        # Carga de .env y variables de entorno
│       ├── database/                      # Conexión GORM y runner de golang-migrate
│       ├── httpx/                         # AppError uniforme y formateo de errores de validación
│       ├── empresas/                      # CRUD de empresas (soft delete)
│       ├── usuarios/                      # CRUD de usuarios (1 sucursal, bcrypt, soft delete)
│       └── crm/                           # Módulo CRM completo:
│           ├── models.go / dto.go / repository.go / service.go / handler.go
│           ├── campana_models.go / campana_dto.go / campana_repository.go / campana_service.go / campana_handler.go
│           ├── etapa_oportunidad.go       # Catálogo de etapas de oportunidades
│           └── handler_test.go / campana_handler_test.go / campana_integration_test.go
│
└── frontend/
    ├── Dockerfile / nginx.conf            # Build Angular + servidor estático Nginx
    ├── package.json / angular.json
    └── src/
        ├── environments/environment.ts    # URL base de la API (http://localhost:8080/api/v1)
        ├── styles.scss                    # Variables CSS globales (--primary, --bg-page, etc.)
        └── app/
            ├── app.routes.ts              # Rutas lazy-loaded protegidas por authGuard / guestGuard
            ├── app.config.ts              # provideRouter, provideHttpClient(withInterceptors([authInterceptor]))
            ├── core/auth/                 # AuthService (signals), auth.interceptor, auth.guard, auth.models
            ├── layout/main-layout/        # Shell principal: sidebar con submenús desplegables, topbar, responsive
            └── features/
                ├── auth/login/            # Pantalla de inicio de sesión
                ├── dashboard/             # Panel principal con KPIs en vivo del CRM y accesos rápidos
                ├── coming-soon/           # Pantalla placeholder para módulos pendientes
                └── crm/
                    ├── crm.models.ts      # Interfaces TypeScript del dominio CRM
                    ├── crm.service.ts     # Cliente HTTP hacia /api/v1/crm/*
                    ├── resumen-comercial/ # /crm/resumen
                    ├── leads/             # /crm/leads (CRUD, filtros, etapas, conversión, atribución de campaña)
                    ├── campanas/          # /crm/campanas (campañas, tipos, detalle con indicadores, leads y gastos)
                    ├── configuracion/     # /crm/configuracion (Orígenes, Etapas, Etapas de oportunidades, Motivos)
                    └── crm-seccion-vacia/ # Placeholder reutilizable para secciones CRM en construcción
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
| `000006` | Etapas de oportunidades | `crm_etapa_oportunidad` (orden, probabilidad %, ganada/perdida, soft delete) |

El diagrama relacional completo y actualizado está en [tesserak_modelo_corregido.drawio](tesserak_modelo_corregido.drawio) (editable en [diagrams.net](https://app.diagrams.net/)) y la documentación detallada de convenciones de BD en [documentacion base de datos.md](documentacion%20base%20de%20datos.md).

---

## 6. Referencia de API REST

Base URL local: `http://localhost:8080`

### Salud y Autenticación
- `GET /health` — Verifica conectividad con PostgreSQL.
- `POST /api/v1/auth/login` — Autenticación por email o usuario + password (+ `identificador_fiscal` opcional). Retorna JWT HS256 y datos del usuario.

### Empresas (`/api/v1/empresas`) y Usuarios (`/api/v1/usuarios`)
- `POST`, `GET` (filtros + paginación), `GET /:id`, `PATCH /:id`, `DELETE /:id` (soft delete).

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
| `/crm/oportunidades` | Base lista | Sección pendiente |
| `/crm/pipeline` | Base lista | Sección pendiente |
| `/crm/actividades` | Base lista | Sección pendiente |

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

- **Implementado:** Infraestructura base, Auth JWT, Empresas, Usuarios, CRM Leads (con `usuario_facebook` y atribución a campaña), Campañas de marketing (tipos, campañas, leads atribuidos, UTM, gastos e indicadores) y Configuración CRM (Orígenes, Etapas, Etapas de oportunidades, Motivos de descarte).
- **Próximos módulos:** Oportunidades y Pipeline comercial (Kanban), Agenda de Actividades en UI, Resumen comercial, Terceros (Clientes/Proveedores/Empleados y FK de conversión), Productos/Inventario, Ventas, Compras y Lealtad.
