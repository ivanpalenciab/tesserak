# Tesserak Backend (Go + Gin + GORM + PostgreSQL)

## Estructura

```
backend/
├── cmd/api/main.go              # punto de entrada (servidor HTTP + migraciones)
├── internal/
│   ├── config/                  # variables de entorno
│   ├── database/                # conexión GORM y runner de migraciones
│   ├── httpx/                   # errores y validación HTTP comunes
│   ├── crm/                     # módulo CRM: catálogos, leads, actividades, intereses y resumen
│   ├── empresas/                # módulo empresas: model, dto, repository, service, handler
│   └── usuarios/                # módulo usuarios: model, dto, repository, service, handler
├── migrations/                  # SQL versionado (golang-migrate), embebido en el binario
└── docker-compose.yml           # PostgreSQL local
```

## Ejecutar

```powershell
# 1. Levantar PostgreSQL (requiere Docker Desktop abierto)
docker compose up -d --wait

# 2. Variables de entorno
$env:DATABASE_URL = "postgres://tesserak:tesserak@localhost:5432/tesserak?sslmode=disable"

# 3. Arrancar (aplica migraciones automáticamente)
go run ./cmd/api
#   go run ./cmd/api -migrate        -> solo aplica migraciones
#   go run ./cmd/api -skip-migrate   -> arranca sin migrar
```

### Datos mínimos para probar
Un usuario necesita una sucursal, y la sucursal una empresa:

```sql
INSERT INTO empresas (nombre_legal, razon_social, identificador_fiscal)
VALUES ('Tesserak SAS', 'Tesserak SAS', '900123456-7');

INSERT INTO sucursales (empresa_id, nombre) VALUES (1, 'Principal');
```

## API de empresas (`/api/v1/empresas`)

| Método | Ruta | Descripción |
|---|---|---|
| POST | `/api/v1/empresas` | Crear empresa |
| GET | `/api/v1/empresas` | Listar. Filtros: `activo`, `q`, `page`, `page_size` |
| GET | `/api/v1/empresas/:id` | Obtener por id |
| PATCH | `/api/v1/empresas/:id` | Actualización parcial (solo campos enviados) |
| DELETE | `/api/v1/empresas/:id` | **Soft delete** (`deleted_at = now()`), responde 204 |

Ejemplo:

```powershell
curl -X POST http://localhost:8080/api/v1/empresas -H "Content-Type: application/json" -d '{
  "nombre_legal": "Tesserak SAS",
  "razon_social": "Tesserak SAS",
  "identificador_fiscal": "900123456-7",
  "email": "contacto@tesserak.com",
  "direccion": "Calle 100 # 10-20"
}'
```

### Reglas de negocio (empresas)
- `identificador_fiscal` es único entre empresas no eliminadas → `409 identificador_fiscal_duplicado`.
- Una empresa eliminada (soft delete) libera su identificador fiscal para volver a registrarse.
- Los listados y consultas excluyen automáticamente las empresas eliminadas.

## API de usuarios (`/api/v1/usuarios`)

| Método | Ruta | Descripción |
|---|---|---|
| POST | `/api/v1/usuarios` | Crear usuario (la empresa se deduce de la sucursal) |
| GET | `/api/v1/usuarios` | Listar. Filtros: `empresa_id`, `sucursal_id`, `activo`, `q`, `page`, `page_size` |
| GET | `/api/v1/usuarios/:id` | Obtener por id |
| PATCH | `/api/v1/usuarios/:id` | Actualización parcial (solo campos enviados) |
| DELETE | `/api/v1/usuarios/:id` | **Soft delete** (`deleted_at = now()`), responde 204 |

Ejemplo:

```powershell
curl -X POST http://localhost:8080/api/v1/usuarios -H "Content-Type: application/json" -d '{
  "sucursal_id": 1,
  "nombre_usuario": "jperez",
  "nombres": "Juan",
  "apellidos": "Pérez",
  "email": "juan@tesserak.com",
  "password": "Secreta123"
}'
```

### Reglas de negocio
- Cada usuario pertenece a **una sucursal**; `empresa_id` se guarda desnormalizado y una FK compuesta `(sucursal_id, empresa_id)` garantiza la consistencia.
- **Email único por empresa** (sin distinguir mayúsculas) entre usuarios no eliminados → `409 email_duplicado`.
- **Nombre de usuario único por empresa** → `409 nombre_usuario_duplicado`.
- Un usuario eliminado libera su email/nombre de usuario para volver a usarse.
- No se puede mover un usuario a una sucursal de otra empresa → `422 sucursal_otra_empresa`.
- Contraseñas con bcrypt; `password_hash` nunca se expone en la API.
- Los usuarios eliminados no aparecen en listados ni consultas (GORM filtra `deleted_at IS NULL`).

## API de CRM (`/api/v1/crm`)

### Control / Resumen
| Método | Ruta | Descripción |
|---|---|---|
| GET | `/api/v1/crm/resumen?empresa_id=` | KPIs del embudo: totales, tasa de conversión, valor estimado, desglose por etapa y temperatura, actividades pendientes y vencidas |

### Catálogos por empresa (CRUD con soft delete)
| Recurso | Rutas |
|---|---|
| Orígenes de lead | `POST/GET /api/v1/crm/origenes`, `GET/PATCH/DELETE /api/v1/crm/origenes/:id` |
| Estados / Etapas | `POST/GET /api/v1/crm/estados`, `GET/PATCH/DELETE /api/v1/crm/estados/:id` |
| Motivos de descarte | `POST/GET /api/v1/crm/motivos-descarte`, `GET/PATCH/DELETE /api/v1/crm/motivos-descarte/:id` |

### Leads, actividades e intereses
| Método | Ruta | Descripción |
|---|---|---|
| POST | `/api/v1/crm/leads` | Crear lead |
| GET | `/api/v1/crm/leads?empresa_id=` | Listar leads (filtros: `sucursal_id`, `origen_id`, `estado_id`, `usuario_asignado_id`, `temperatura`, `convertido`, `q`, `page`, `page_size`) |
| GET | `/api/v1/crm/leads/:id` | Ver lead |
| PATCH | `/api/v1/crm/leads/:id` | Actualización parcial |
| PATCH | `/api/v1/crm/leads/:id/estado` | Mover lead de etapa (y asignar o limpiar motivo de descarte) |
| POST | `/api/v1/crm/leads/:id/convertir` | Marcar lead como convertido (`convertido = true`, `fecha_conversion = now()`) |
| DELETE | `/api/v1/crm/leads/:id` | Soft delete |
| POST / GET | `/api/v1/crm/leads/:id/actividades` | Registrar y listar actividades de seguimiento del lead |
| GET | `/api/v1/crm/actividades` | Agenda global (`empresa_id`, `usuario_id`, `tipo_actividad`, `estado`, `solo_vencidas=true`) |
| PATCH / DELETE | `/api/v1/crm/actividades/:id` | Actualizar o eliminar actividad |
| POST / GET / DELETE | `/api/v1/crm/leads/:id/intereses` (`/:interesId`) | Gestionar productos/categorías de interés del lead |

