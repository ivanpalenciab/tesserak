# Tesserak Backend (Go + Gin + GORM + PostgreSQL)

## Estructura

```
backend/
├── cmd/api/main.go              # punto de entrada (servidor HTTP + migraciones)
├── internal/
│   ├── auth/                    # autenticación JWT HS256, bcrypt y middleware RequireAuth
│   ├── config/                  # variables de entorno
│   ├── consecutivos/            # secuencias automáticas configurables (COMPRA, VENTA, INVENTARIO)
│   ├── database/                # conexión GORM y runner de migraciones
│   ├── httpx/                   # errores y validación HTTP comunes
│   ├── crm/                     # módulo CRM: catálogos, leads, oportunidades, campañas, actividades y resumen
│   ├── compras/                 # módulo compras: com_compras y com_compras_detalle (entrada atómica a inventario)
│   ├── empresas/                # módulo empresas: CRUD + creación automática de Consumidor Final (222222222222)
│   ├── inventario/              # módulo inventarios: inv_ubicaciones, inv_stock, ajustes, traslados y hooks compras/ventas
│   ├── precios/                 # módulo listas de precios: listas_precios y lista_precios_detalle
│   ├── productos/               # módulo artículos: productos, servicios y catálogos (marcas, impuestos, categorías)
│   ├── terceros/                # módulo terceros: tercero_maestro, roles_tercero globales y tercero_empresa_roles
│   ├── usuarios/                # módulo usuarios: CRUD, múltiples sucursales (usuario_sucursales), key UUID v4 y regeneración
│   └── ventas/                  # módulo ventas: ven_ventas, ven_ventas_detalle, IVA automático y salida atómica de stock
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

## API de empresas (`/api/v1/empresas`) y sucursales (`/api/v1/empresas/sucursales`)

| Método | Ruta | Descripción |
|---|---|---|
| POST | `/api/v1/empresas` | Crear empresa (crea automáticamente `Consumidor Final`) |
| GET | `/api/v1/empresas` | Listar. Filtros: `activo`, `q`, `page`, `page_size` |
| GET | `/api/v1/empresas/:id` | Obtener por id |
| PATCH | `/api/v1/empresas/:id` | Actualización parcial (solo campos enviados) |
| DELETE | `/api/v1/empresas/:id` | **Soft delete** (`deleted_at = now()`), responde 204 |
| GET | `/api/v1/empresas/sucursales` | Listar sucursales de la empresa autenticada (JWT; filtrado a sucursales autorizadas para no admin) |
| POST | `/api/v1/empresas/sucursales` | Crear sucursal + ubicación por defecto `BOD-<id>` en la misma transacción (solo `ADMINISTRADOR`) |
| PATCH | `/api/v1/empresas/sucursales/:id` | Actualizar nombre, dirección o estado de la sucursal (solo `ADMINISTRADOR`) |
| DELETE | `/api/v1/empresas/sucursales/:id` | **Soft delete** de sucursal; impide eliminar la única sucursal de la empresa (solo `ADMINISTRADOR`) |

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
| GET | `/api/v1/usuarios/roles` | Listar roles de sistema asignables (`ADMINISTRADOR`, `ADMINISTRADOR_SUCURSAL`, `GESTOR_VENTAS`, `ASESOR_COMERCIAL`, `GESTOR_COMPRAS`, `USUARIO`) |
| GET | `/api/v1/usuarios/sucursales` | Listar sucursales activas accesibles por el usuario autenticado |
| GET | `/api/v1/usuarios/mi-perfil` | Obtener perfil completo del usuario autenticado (JWT) |
| PATCH | `/api/v1/usuarios/mi-perfil` | Actualizar datos propios (`nombres`, `apellidos`, `nombre_usuario`, `email`) y/o contraseña (exige `password_actual`) |
| POST | `/api/v1/usuarios` | Crear usuario con múltiples roles (`rol_ids`) y sucursales (`sucursal_ids`). 1er usuario de empresa recibe `ADMINISTRADOR`; luego solo `ADMINISTRADOR` o `ADMINISTRADOR_SUCURSAL` |
| GET | `/api/v1/usuarios` | Listar (acotado por empresa y, para no admin, a sus sucursales autorizadas). Filtros: `sucursal_id`, `activo`, `q`, `page`, `page_size` |
| GET | `/api/v1/usuarios/:id` | Obtener por id |
| PATCH | `/api/v1/usuarios/:id` | Actualización parcial (permite reemplazar `rol_ids` y `sucursal_ids`; solo `ADMINISTRADOR` o `ADMINISTRADOR_SUCURSAL` en su alcance) |
| POST | `/api/v1/usuarios/:id/regenerar-key` | Regenerar `key` UUID v4 |
| DELETE | `/api/v1/usuarios/:id` | **Soft delete** (`deleted_at = now()`), responde 204 |

Ejemplo:

```powershell
curl -X POST http://localhost:8080/api/v1/usuarios -H "Authorization: Bearer <TOKEN>" -H "Content-Type: application/json" -d '{
  "rol_ids": [3, 4],
  "sucursal_ids": [1, 2],
  "nombre_usuario": "jperez",
  "nombres": "Juan",
  "apellidos": "Pérez",
  "email": "juan@tesserak.com",
  "password": "Secreta123"
}'
```

### Reglas de negocio (usuarios, roles y aislamiento por sucursal)
- Un usuario pertenece a una `empresa` (`usuarios.empresa_id`), puede tener **uno o varios roles de sistema** en `usuario_roles` (permisos aditivos) y acceso a **una o varias sucursales** en `usuario_sucursales`.
- El primer usuario creado en cada empresa recibe automáticamente `ADMINISTRADOR` (acceso total a todas las sucursales, consecutivos, sucursales y usuarios).
- `ADMINISTRADOR_SUCURSAL` gestiona usuarios y ubicaciones únicamente dentro de sus sucursales asignadas, sin poder asignar `ADMINISTRADOR` global.
- Todos los módulos operativos (`ventas`, `compras`, `inventario`, `crm`, `usuarios`, `sucursales`) filtran automáticamente los datos a las sucursales autorizadas del usuario (`AllowedSucursalIDs()`), devolviendo `403 sucursal_no_autorizada` ante accesos fuera de alcance.
- Las FKs compuestas `(usuario_id, empresa_id)` y `(sucursal_id, empresa_id)` en `usuario_sucursales` garantizan a nivel de PostgreSQL que no se puedan mezclar sucursales de distintas empresas.
- **Email único por empresa** (sin distinguir mayúsculas) entre usuarios no eliminados → `409 email_duplicado`.
- **Nombre de usuario único por empresa** → `409 nombre_usuario_duplicado`.
- Un usuario eliminado libera su email/nombre de usuario para volver a usarse.
- No se pueden asignar sucursales de otra empresa → `422 sucursal_otra_empresa`.
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

## API de Terceros — Roles (`/api/v1/terceros/roles`)

Todos los endpoints requieren autenticación JWT (`Authorization: Bearer <token>`).

La migración `000013_init_roles_tercero` inserta automáticamente los roles base del sistema (`es_sistema = true`):
- **`CLIENTE`** — *Cliente*
- **`PROVEEDOR`** — *Proveedor*
- **`EMPLEADO`** — *Empleado*

| Método | Ruta | Descripción |
|---|---|---|
| POST | `/api/v1/terceros/roles` | Crear un rol de tercero personalizado (`codigo`, `nombre_rol`, `descripcion`, `activo`) |
| GET | `/api/v1/terceros/roles` | Listar roles de tercero. Filtros: `activo`, `q` |
| GET | `/api/v1/terceros/roles/:id` | Obtener un rol por id |
| PATCH | `/api/v1/terceros/roles/:id` | Actualización parcial de rol |
| DELETE | `/api/v1/terceros/roles/:id` | **Soft delete** (`deleted_at = now()`), responde 204 |

### Reglas de negocio (roles de tercero)
- `codigo` (normalizado a mayúsculas) y `nombre_rol` son únicos entre roles activos no eliminados → `409 codigo_duplicado` / `409 nombre_duplicado`.
- Los roles base del sistema (`CLIENTE`, `PROVEEDOR`, `EMPLEADO`, con `es_sistema = true`) están protegidos: no pueden eliminarse ni cambiar su `codigo` → `409 rol_sistema_protegido`.

## API de Terceros — Maestro (`/api/v1/terceros`)

Todos los endpoints requieren autenticación JWT (`Authorization: Bearer <token>`) y aíslan automáticamente los registros por `empresa_id` del token (`tercero_maestro.empresa_id`, sin tabla intermedia `tercero_empresa`).

| Método | Ruta | Descripción |
|---|---|---|
| POST | `/api/v1/terceros` | Crear tercero (`tipo_identificacion`, `numero_identificacion`, `razon_social`, `naturaleza_juridica`, `pais_id`, `email`, `telefono`, `direccion`, `activo`) |
| GET | `/api/v1/terceros` | Listar terceros paginados. Filtros: `tipo_identificacion`, `naturaleza_juridica`, `pais_id`, `activo`, `q`, `page`, `page_size` |
| GET | `/api/v1/terceros/:id` | Obtener tercero por id |
| PATCH | `/api/v1/terceros/:id` | Actualización parcial (soporta `limpiar_pais`) |
| DELETE | `/api/v1/terceros/:id` | **Soft delete** (`deleted_at = now()`), responde 204 |

### Reglas de negocio (terceros)
- `tipo_identificacion` acepta `NIT`, `CC`, `CE`, `PASAPORTE`, `RUT`, `DNI`, `OTRO`.
- `naturaleza_juridica` acepta `NATURAL` o `JURIDICA`.
- La combinación `(empresa_id, tipo_identificacion, numero_identificacion)` es única entre terceros no eliminados → `409 identificacion_duplicada`.
- Al crear una empresa se genera automáticamente el tercero **Consumidor Final** (`CC 222222222222`, `NATURAL`, rol único `CLIENTE`).

## API de Consecutivos (`/api/v1/consecutivos`)

Requiere JWT. Permite configurar prefijos, siguiente número, longitud de relleno y frecuencia de reinicio (`NUNCA`, `ANUAL`, `MENSUAL`) para `COMPRA`, `VENTA` e `INVENTARIO`.

| Método | Ruta | Descripción |
|---|---|---|
| GET | `/api/v1/consecutivos/tipos-documento` | Listar tipos de documento |
| GET / POST | `/api/v1/consecutivos/secuencias` | Listar o crear secuencia para la empresa |
| PATCH | `/api/v1/consecutivos/secuencias/:id` | Actualizar prefijo / parámetros de la secuencia |

## API de Compras (`/api/v1/compras`)

Requiere JWT. CRUD completo con **soft delete** (`DELETE /api/v1/compras/:id`) y consecutivo automático (`COMPRA`).
- Al crear/activar (`REGISTRADA` o `PAGADA`) se registra automáticamente `ENTRADA_COMPRA` en `inv_documentos` / `inv_stock` dentro de la misma transacción (solo artículos de tipo `PRODUCTO` con `control_inventario = TRUE`). Si falla el movimiento de inventario se revierte toda la compra.

## API de Inventarios (`/api/v1/inventario`)

Requiere JWT. Tablas con prefijo `inv_` y stock materializado en `inv_stock`.

| Método | Ruta | Descripción |
|---|---|---|
| GET | `/api/v1/inventario/sucursales` | Sucursales activas de la empresa |
| GET | `/api/v1/inventario/tipos-movimiento` | Catálogo global de tipos (`ENTRADA`, `SALIDA`, `TRASLADO`) |
| POST / GET / PATCH `:id` / DELETE `:id` | `/api/v1/inventario/ubicaciones` | CRUD de ubicaciones/bodegas por sucursal (soft delete) |
| GET | `/api/v1/inventario/stock` | Existencias consolidadas (`inv_stock`) con costo promedio y valor total |
| POST / GET / GET `:id` | `/api/v1/inventario/movimientos` | Ajustes, pérdidas, averías, devoluciones y traslados internos entre ubicaciones |

## API de Ventas (`/api/v1/ventas`)

Requiere JWT. CRUD completo (`POST`, `GET`, `GET /:id`, `PATCH /:id`, `DELETE /:id` con **soft delete**).
- **Cliente por defecto:** Si no se envía `cliente_id`, se asigna automáticamente **Consumidor Final** (`222222222222`).
- **Consecutivo automático:** Generado con la secuencia `VENTA` de la empresa.
- **Impuesto automático:** Cada línea calcula `impuesto = subtotal * (grupos_impuesto.tarifa / 100)` a partir del grupo de impuestos del artículo seleccionado.
- **Salida atómica de inventario:** Al guardar en estado `REGISTRADA` o `PAGADA`, descuenta `inv_stock` (`SALIDA_VENTA`) dentro de la misma transacción. Los **servicios** (y productos sin `control_inventario`) no validan ni descuentan existencias; si algún producto inventariable no tiene stock suficiente, la transacción completa hace **rollback**.



