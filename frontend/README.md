# Tesserak Frontend (Angular 22 + Signals + Vitest)

Aplicación web SPA multi-empresa construida con **Angular 22** (componentes *standalone*, Signals, Reactive Forms y carga diferida por ruta).

## Comandos principales

```powershell
# Instalar dependencias
npm install

# Servidor de desarrollo (http://localhost:4200)
npm start

# Build de producción (salida en dist/)
npm run build

# Pruebas unitarias con Vitest
npm test -- --watch=false
```

> En Windows con políticas restringidas de PowerShell para `npx`, se puede invocar el CLI directamente con:
> `node ./node_modules/@angular/cli/bin/ng.js build` y `node ./node_modules/@angular/cli/bin/ng.js test --watch=false`.

## Módulos implementados

| Ruta | Módulo | Funcionalidad principal |
|---|---|---|
| `/login` | Autenticación | Inicio de sesión con JWT HS256 y persistencia de sesión |
| `/dashboard` | Inicio | KPIs en vivo y accesos rápidos |
| `/crm/leads` | CRM | CRUD de leads, etapas, conversión y atribución a campañas |
| `/crm/oportunidades` | CRM | Gestión de oportunidades comerciales |
| `/crm/campanas` | CRM | Campañas, tipos, leads atribuidos, gastos e indicadores |
| `/crm/configuracion` | CRM | Catálogos CRM (orígenes, etapas, estados de oportunidad, motivos) |
| `/productos/catalogo` | Productos | Productos físicos y servicios (`articulos` → `productos` / `servicios`) |
| `/productos/categorias` | Productos | CRUD de categorías (soft delete) |
| `/productos/listas-precios` | Precios | Listas de precios y precios asignados por artículo |
| `/productos/configuracion` | Productos | Marcas y grupos de impuesto (`grupos_impuesto`) |
| `/terceros` | Terceros | Maestro de terceros por empresa con roles globales de solo lectura |
| `/compras` | Compras | Compras con selector de proveedores, consecutivo automático y soft delete |
| `/inventario` | Inventario | Ubicaciones → stock (`inv_stock`), ajustes, pérdidas, averías, devoluciones y traslados |
| `/ventas` | Ventas | Ventas con **Consumidor Final** (`222222222222`) por defecto, IVA automático, salida de stock y soft delete |
| `/gestion-empresa` | Administración | Prefijos de consecutivos, CRUD de ubicaciones por sucursal y gestión de `key` UUID de usuario |
