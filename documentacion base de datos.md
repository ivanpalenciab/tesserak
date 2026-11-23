# Documentación de la Base de Datos — Tesserak

Este documento describe cómo está diseñada y construida la base de datos de **Tesserak**, los **patrones de diseño** que sigue, el propósito de **cada tabla** y el estado actual de implementación en PostgreSQL.

- **Motor:** PostgreSQL 14+
- **Gestión de esquema:** Migraciones SQL versionadas con `golang-migrate` (embebidas en el binario Go en `backend/migrations/`). No se usa `AutoMigrate` de GORM para no perder índices únicos parciales, restricciones `CHECK` ni claves foráneas compuestas.
- **Acceso a datos:** GORM sobre `pgx/v5`.
- **Fuente del diseño completo:** `Diseño aplicación.drawio` (47 tablas en 8 dominios).

---

## 1. Patrones de diseño de la base de datos

| Patrón | Dónde se aplica | Qué resuelve |
|---|---|---|
| **Multi-tenant con discriminador (`empresa_id`)** | `sucursales`, `usuarios`, `crm_*`, `ventas`, `compras`, `lealtad_*`, `empresa_productos`, `tercero_empresa` | Múltiples empresas comparten la misma base de datos con aislamiento lógico por `empresa_id`. |
| **Integridad multi-tenant por FK compuesta** | `usuarios (sucursal_id, empresa_id)`, `crm_leads (sucursal_id/origen_id/estado_id/motivo_descarte_id, empresa_id)` | Impide a nivel de motor que un registro de la empresa A referencie una sucursal, origen, estado o motivo de la empresa B. |
| **Soft Delete + Índices únicos parciales** | `empresas`, `sucursales`, `usuarios`, `crm_origen_lead`, `crm_estado_lead`, `crm_motivo_descarte`, `crm_leads`, `crm_lead_actividades` | `deleted_at TIMESTAMPTZ` conserva el historial sin borrar físicamente. Los índices `UNIQUE (...) WHERE deleted_at IS NULL` permiten reutilizar emails, nombres de usuario o NITs de registros eliminados. |
| **Catálogo maestro + Vinculación por empresa** | `maestro_productos` → `empresa_productos`; `tercero_maestro` → `tercero_empresa` | Evita duplicar información base (productos o personas/empresas) entre tenants, permitiendo configuración propia por empresa. |
| **Party / Role (Supertipo–Subtipo / Class Table Inheritance)** | `tercero_maestro` → `tercero_empresa` → `tercero_empresa_roles` → `clientes` / `proveedores` / `empleados` | Una misma persona o empresa puede ser simultáneamente cliente, proveedor y empleado sin duplicar sus datos de identidad. |
| **Header–Detail (Cabecera–Detalle)** | `ventas` / `ventas_detalle`, `compras` / `compras_detalle`, `documento_inventario` / `movimiento_inventario_detalle` | Separa metadatos y totales del documento de sus líneas de ítems. |
| **Ledger (Libro mayor inmutable) + Saldo materializado** | `lealtad_transacciones` + `lealtad_cuentas.saldo_puntos`; `movimiento_inventario_detalle` + `inventario_stock` | Mantiene trazabilidad completa de cada movimiento mientras conserva un saldo/stock actualizado para lecturas rápidas. |
| **RBAC (Control de acceso basado en roles)** | `usuarios` ↔ `usuario_roles` ↔ `roles_sistema` ↔ `rol_permisos` ↔ `permisos` | Desacopla usuarios de permisos mediante roles asignables. |
| **Feature Flags por tenant (Suscripción modular)** | `modulos` ↔ `modulo_empresa` | Permite activar o desactivar módulos funcionales por empresa. |
| **Pipeline / Máquina de estados configurable** | `crm_estado_lead` + `crm_leads` + `crm_motivo_descarte` | Cada empresa define sus propias etapas de venta, orden y estados finales, con seguimiento vía `crm_lead_actividades`. |
| **Auditoría temporal** | Tablas transaccionales y maestras | Uso estándar de `created_at`, `updated_at` (y `deleted_at` donde aplica) en `TIMESTAMPTZ` (UTC). |

---

## 2. Diagrama ER de las tablas implementadas (Migraciones `000001` y `000002`)

```mermaid
erDiagram
    paises ||--o{ ciudades : "tiene"
    empresas ||--o{ sucursales : "opera"
    ciudades ||--o{ sucursales : "ubica"
    empresas ||--o{ usuarios : "pertenece"
    sucursales ||--o{ usuarios : "asigna"

    empresas ||--o{ crm_origen_lead : "define"
    empresas ||--o{ crm_estado_lead : "define"
    empresas ||--o{ crm_motivo_descarte : "define"
    empresas ||--o{ crm_leads : "gestiona"
    sucursales ||--o{ crm_leads : "origina"
    crm_origen_lead ||--o{ crm_leads : "clasifica"
    crm_estado_lead ||--o{ crm_leads : "etapa"
    crm_motivo_descarte ||--o{ crm_leads : "descarta"
    monedas ||--o{ crm_leads : "valora"
    usuarios ||--o{ crm_leads : "asigna / crea"

    crm_leads ||--o{ crm_lead_actividades : "registra"
    usuarios ||--o{ crm_lead_actividades : "ejecuta"
    crm_leads ||--o{ crm_lead_interes_productos : "interesa"
```

---

## 3. Catálogo completo de tablas por dominio

> **Leyenda de estado:**
> - **Implementada**: existe en `backend/migrations/` y está lista para PostgreSQL.
> - **Diseñada**: definida en `Diseño aplicación.drawio`, pendiente de migración SQL.

### 3.1 Geografía y Monedas (Catálogos globales)

| Tabla | Estado | Para qué sirve | Patrón de diseño |
|---|---|---|---|
| `paises` | Implementada (`000001`, `000003`) | Catálogo ISO de países (`nombre`, `iso_alpha_2`, `iso_alpha_3`, `codigo_numerico`, `nombre_oficial`). | **Lookup / Catálogo de referencia global**. |
| `ciudades` | Implementada (`000001`) | Ciudades asociadas a un país (`pais_id`, `nombre` único por país). | **Catálogo jerárquico** (1:N con `paises`). |
| `monedas` | Implementada (`000002`) | Catálogo ISO de divisas (`codigo_iso`, `nombre_moneda`, `simbolo`) usado en CRM, ventas, inventario y lealtad. | **Lookup / Catálogo de referencia global**. |

### 3.2 Empresas, Sucursales y Módulos (Multi-tenancy)

| Tabla | Estado | Para qué sirve | Patrón de diseño |
|---|---|---|---|
| `empresas` | Implementada (`000001`) | Entidad raíz de cada tenant (`nombre_legal`, `razon_social`, `identificador_fiscal` único activo, datos de contacto y logo). | **Tenant Root + Soft Delete + Índice único parcial**. |
| `sucursales` | Implementada (`000001`) | Sedes físicas o puntos de operación de una empresa. Expone `UNIQUE (id, empresa_id)` para FKs compuestas. | **Jerarquía de Tenant + Composite Key para aislamiento**. |
| `modulos` | Diseñada | Catálogo de módulos funcionales de la plataforma (CRM, Inventario, Ventas, Lealtad, etc.). | **Catálogo de referencia / Feature registry**. |
| `modulo_empresa` | Diseñada | Activa módulos contratados por cada empresa con su fecha de activación. | **Tabla puente N:M + Feature Flags por tenant**. |

### 3.3 Seguridad, Usuarios y RBAC

| Tabla | Estado | Para qué sirve | Patrón de diseño |
|---|---|---|---|
| `usuarios` | Implementada (`000001`) | Credenciales y perfil de acceso. Cada usuario pertenece a una sucursal y empresa (`FK (sucursal_id, empresa_id)`). Email y `nombre_usuario` son únicos por empresa (insensibles a mayúsculas) entre usuarios no borrados. Contraseña con `bcrypt`. | **Multi-tenant con FK compuesta + Soft Delete + Índice único parcial**. |
| `roles_sistema` | Diseñada | Define roles de acceso al sistema (ej. Administrador, Vendedor, Cajero). | **RBAC (Role)**. |
| `permisos` | Diseñada | Define permisos granulares de acciones en el sistema. | **RBAC (Permission)**. |
| `usuario_roles` | Diseñada | Asigna uno o varios roles del sistema a cada usuario. | **Tabla puente N:M (RBAC)**. |
| `rol_permisos` *(en drawio sin título)* | Diseñada | Asocia permisos a cada rol del sistema. | **Tabla puente N:M (RBAC)**. |

### 3.4 Módulo CRM (Gestión de prospectos y pipeline)

| Tabla | Estado | Para qué sirve | Patrón de diseño |
|---|---|---|---|
| `crm_origen_lead` | Implementada (`000002`) | Canales de captación configurables por empresa (Web, Referido, WhatsApp, Evento, etc.). | **Catálogo por tenant + Soft Delete + FK compuesta**. |
| `crm_estado_lead` | Implementada (`000002`) | Etapas ordenadas del embudo comercial por empresa (`orden`, `es_estado_final`). | **Máquina de estados configurable + Catálogo por tenant**. |
| `crm_motivo_descarte` | Implementada (`000002`) | Razones por las que se descarta o pierde un lead (Precio, Competencia, Sin respuesta, etc.). | **Catálogo por tenant + Soft Delete**. |
| `crm_leads` | Implementada (`000002`) | Prospectos comerciales con datos de contacto, calificación (`temperatura`: `FRIO`/`TIBIO`/`CALIENTE`, `score`: 0–100), `valor_estimado`, asignación a usuario, estado actual y trazabilidad de conversión (`convertido`, `fecha_conversion`, `convertido_tercero_empresa_rol_id`). | **Entidad transaccional multi-tenant + FK compuestas + Soft Delete + Conversión a Party/Role**. |
| `crm_lead_actividades` | Implementada (`000002`) | Historial y agenda de seguimiento sobre cada lead (`LLAMADA`, `CORREO`, `REUNION`, `WHATSAPP`, `TAREA`, `NOTA`) con fecha de vencimiento y estado (`PENDIENTE`, `COMPLETADA`, `CANCELADA`). | **Activity Log / Task Tracking + Soft Delete**. |
| `crm_lead_interes_productos` | Implementada (`000002`) | Productos, categorías o notas de interés asociadas a un lead con su cantidad estimada. | **Relación 1:N de detalle con referencia flexible (producto, categoría o nota)**. |

### 3.5 Terceros (Clientes, Proveedores y Empleados)

| Tabla | Estado | Para qué sirve | Patrón de diseño |
|---|---|---|---|
| `tercero_maestro` | Diseñada | Registro único de identidad de una persona natural o jurídica (tipo/número de identificación, razón social, país, naturaleza jurídica). | **Party Pattern (Supertipo maestro compartido)**. |
| `tercero_empresa` | Diseñada | Vincula un tercero del maestro con una empresa específica. | **Tabla puente N:M (Catálogo compartido → Tenant)**. |
| `roles_tercero` | Diseñada | Catálogo de roles de negocio que puede tener un tercero (Cliente, Proveedor, Empleado). | **Catálogo de roles de negocio**. |
| `tercero_empresa_roles` | Diseñada | Asigna roles concretos a un tercero dentro de una empresa. Es la referencia central que usan ventas, compras, lealtad y CRM. | **Party–Role Pattern**. |
| `clientes` | Diseñada | Datos específicos de un tercero con rol cliente (`limite_credito`, `condiciones_pago`, `fecha_registro`). | **Class Table Inheritance (Subtipo de rol)**. |
| `proveedores` | Diseñada | Datos específicos de un tercero con rol proveedor (`evaluacion`). | **Class Table Inheritance (Subtipo de rol)**. |
| `empleados` | Diseñada | Datos laborales de un tercero con rol empleado (`fecha_contratacion`, `salario`, `tipo_contrato_id`, `cargo_id`, `estado`). | **Class Table Inheritance (Subtipo de rol)**. |
| `cargos` | Diseñada | Catálogo de cargos laborales para empleados. | **Lookup / Catálogo**. |
| `tipos_contrato` | Diseñada | Catálogo de tipos de contrato laboral (Indefinido, Fijo, Prestación de servicios, etc.). | **Lookup / Catálogo**. |

### 3.6 Catálogo de Productos

| Tabla | Estado | Para qué sirve | Patrón de diseño |
|---|---|---|---|
| `maestro_productos` | Diseñada | Definición base compartida de productos (`referencia`, `nombre`, `descripcion`, unidad de medida, categoría, marca, grupo de impuesto). | **Catálogo maestro compartido**. |
| `empresa_productos` | Diseñada | Habilita y vincula productos del maestro para una empresa específica. | **Tabla puente N:M (Maestro → Tenant)**. |
| `categorias_producto` | Diseñada | Clasificación de productos por categoría. | **Lookup / Catálogo**. |
| `marcas` | Diseñada | Catálogo de marcas comerciales. | **Lookup / Catálogo**. |
| `unidades_medida` | Diseñada | Unidades de medida (Unidad, Kg, Litro, Caja, etc.). | **Lookup / Catálogo**. |
| `grupos_impuesto` | Diseñada | Esquemas tributarios y tarifas porcentuales aplicables a productos (ej. IVA 19%, Exento 0%). | **Lookup / Catálogo fiscal**. |

### 3.7 Inventario y Bodegas

| Tabla | Estado | Para qué sirve | Patrón de diseño |
|---|---|---|---|
| `ubicaciones` | Diseñada | Bodegas, estantes o zonas físicas dentro de una sucursal. | **Jerarquía física (Sucursal → Ubicación)**. |
| `tipos_movimiento_inventario` | Diseñada | Catálogo de tipos de movimiento (Entrada, Salida, Traslado, Ajuste). | **Lookup / Catálogo**. |
| `documento_inventario` | Diseñada | Cabecera de un movimiento de inventario (consecutivo, fecha, tercero, motivo, usuario, sucursal y referencia polimórfica `documento_origen_tipo` + `documento_origen_id`). | **Header–Detail + Referencia polimórfica**. |
| `movimiento_inventario_detalle` | Diseñada | Líneas de cada documento de inventario con producto, ubicación origen/destino, cantidad, costo unitario y moneda. | **Header–Detail + Ledger de inventario (Kardex)**. |
| `inventario_stock` | Diseñada | Saldo actual y costo promedio ponderado por `empresa_producto_id` y `ubicacion_id`. | **Materialized Balance / Snapshot de existencias**. |

### 3.8 Ventas y Compras

| Tabla | Estado | Para qué sirve | Patrón de diseño |
|---|---|---|---|
| `ventas` | Diseñada | Cabecera de ventas (consecutivo, empresa, sucursal, cliente vía `tercero_empresa_rol_id`, vendedor, estado, subtotal, descuentos, impuestos y total). | **Header–Detail + Documento transaccional** (se anula por `estado`, no se borra). |
| `ventas_detalle` | Diseñada | Ítems vendidos con cantidad, precio unitario, descuento, impuesto y total por línea. | **Header–Detail**. |
| `compras` | Diseñada | Cabecera de compras a proveedores (consecutivo, factura del proveedor, fechas de compra y vencimiento, totales, usuario y sucursal). | **Header–Detail + Documento transaccional**. |
| `compras_detalle` | Diseñada | Ítems comprados con cantidad, subtotal, impuesto y total. | **Header–Detail**. |

### 3.9 Programa de Lealtad

| Tabla | Estado | Para qué sirve | Patrón de diseño |
|---|---|---|---|
| `lealtad_programas` | Diseñada | Programas de fidelización definidos por empresa con vigencia y estado. | **Entidad de configuración por tenant**. |
| `lealtad_reglas` | Diseñada | Reglas de acumulación de puntos (`puntos_por_monto`, `compra_minima`, moneda, vigencia). | **Rule Engine ligero / Estrategia parametrizada**. |
| `lealtad_cuentas` | Diseñada | Cuenta de puntos de un cliente dentro de un programa, con `saldo_puntos` materializado. | **Materialized Balance**. |
| `lealtad_transacciones` | Diseñada | Historial inmutable de acumulación, ajuste o vencimiento de puntos con `saldo_posterior` y vínculo a la venta origen. | **Append-only Ledger (Libro mayor de puntos)**. |
| `lealtad_redenciones` | Diseñada | Registro de canje de puntos por descuentos aplicados en una venta. | **Documento transaccional vinculado a cuenta y venta**. |

---

## 4. Convenciones SQL aplicadas en las migraciones

1. **Nomenclatura:** `snake_case` en minúsculas para tablas, columnas, índices y constraints. Claves foráneas siempre como `<entidad>_id`.
2. **Claves primarias:** `GENERATED ALWAYS AS IDENTITY` (`INT` para catálogos/entidades de volumen medio, `BIGINT` para entidades de alto volumen como `usuarios`, `crm_leads`, `crm_lead_actividades`).
3. **Índices y constraints nombrados explícitamente:** `uq_*` (unique), `ux_*` (unique index parcial), `ix_*` (index), `fk_*` (foreign key), `ck_*` (check). Esto permite al backend inspeccionar `pgErr.ConstraintName` y devolver errores HTTP precisos (`409`, `422`).
4. **Fechas:** siempre `TIMESTAMPTZ` almacenadas en UTC.
5. **Seguridad:** las contraseñas usan `VARCHAR(255)` (preparado para `bcrypt` o `Argon2id`) y nunca se exponen en DTOs de salida.
