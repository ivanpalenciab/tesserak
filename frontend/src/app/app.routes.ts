import { Routes } from '@angular/router';

import {
  adminGuard,
  adminOrBranchAdminGuard,
  authGuard,
  comprasGuard,
  crmGuard,
  guestGuard,
  operacionGuard,
  tercerosGuard,
  ventasGuard,
} from './core/auth/auth.guard';

export const routes: Routes = [
  {
    path: 'login',
    canActivate: [guestGuard],
    loadComponent: () =>
      import('./features/auth/login/login').then((m) => m.LoginComponent),
    title: 'Iniciar sesión — Tesserak',
  },
  {
    path: '',
    canActivate: [authGuard],
    loadComponent: () =>
      import('./layout/main-layout/main-layout').then((m) => m.MainLayoutComponent),
    children: [
      {
        path: 'dashboard',
        loadComponent: () =>
          import('./features/dashboard/dashboard').then((m) => m.DashboardComponent),
        title: 'Inicio — Tesserak',
      },
      {
        path: 'crm',
        canActivate: [crmGuard],
        children: [
          {
            path: 'resumen',
            loadComponent: () =>
              import('./features/crm/resumen-comercial/resumen-comercial').then(
                (m) => m.ResumenComercialComponent,
              ),
            title: 'Resumen comercial — CRM · Tesserak',
          },
          {
            path: 'leads',
            loadComponent: () =>
              import('./features/crm/leads/leads').then((m) => m.CrmLeadsComponent),
            title: 'Administrar leads — CRM · Tesserak',
          },
          {
            path: 'empresas-contactos',
            loadComponent: () =>
              import('./features/crm/crm-seccion-vacia/crm-seccion-vacia').then(
                (m) => m.CrmSeccionVaciaComponent,
              ),
            data: {
              titulo: 'Empresas y contactos',
              subtitulo: 'Espacio reservado para la gestión de empresas y contactos comerciales.',
            },
            title: 'Empresas y contactos — CRM · Tesserak',
          },
          {
            path: 'oportunidades',
            loadComponent: () =>
              import('./features/crm/oportunidades/oportunidades').then(
                (m) => m.CrmOportunidadesComponent,
              ),
            title: 'Administrar oportunidades — CRM · Tesserak',
          },
          {
            path: 'pipeline',
            loadComponent: () =>
              import('./features/crm/crm-seccion-vacia/crm-seccion-vacia').then(
                (m) => m.CrmSeccionVaciaComponent,
              ),
            data: {
              titulo: 'Pipeline comercial',
              subtitulo: 'Espacio reservado para el embudo y seguimiento visual del pipeline.',
            },
            title: 'Pipeline comercial — CRM · Tesserak',
          },
          {
            path: 'actividades',
            loadComponent: () =>
              import('./features/crm/crm-seccion-vacia/crm-seccion-vacia').then(
                (m) => m.CrmSeccionVaciaComponent,
              ),
            data: {
              titulo: 'Actividades',
              subtitulo: 'Espacio reservado para la agenda y seguimiento de actividades del CRM.',
            },
            title: 'Actividades — CRM · Tesserak',
          },
          {
            path: 'campanas',
            loadComponent: () =>
              import('./features/crm/campanas/campanas').then((m) => m.CrmCampanasComponent),
            title: 'Campañas — CRM · Tesserak',
          },
          {
            path: 'configuracion',
            loadComponent: () =>
              import('./features/crm/configuracion/configuracion').then(
                (m) => m.CrmConfiguracionComponent,
              ),
            title: 'Configuración — CRM · Tesserak',
          },
          {
            path: '',
            pathMatch: 'full',
            redirectTo: 'leads',
          },
        ],
      },
      {
        path: 'productos',
        canActivate: [operacionGuard],
        children: [
          {
            path: 'catalogo',
            loadComponent: () =>
              import('./features/productos/catalogo/catalogo').then(
                (m) => m.ProductosCatalogoComponent,
              ),
            title: 'Catálogo de productos — Tesserak',
          },
          {
            path: 'categorias',
            loadComponent: () =>
              import('./features/productos/categorias/categorias').then(
                (m) => m.ProductosCategoriasComponent,
              ),
            title: 'Categorías de productos — Tesserak',
          },
          {
            path: 'listas-precios',
            loadComponent: () =>
              import('./features/precios/listas-precios/listas-precios').then(
                (m) => m.ListasPreciosComponent,
              ),
            title: 'Listas de precios — Tesserak',
          },
          {
            path: 'configuracion',
            loadComponent: () =>
              import('./features/productos/configuracion/configuracion').then(
                (m) => m.ProductosConfiguracionComponent,
              ),
            title: 'Configuración de productos — Tesserak',
          },
          {
            path: '',
            pathMatch: 'full',
            redirectTo: 'catalogo',
          },
        ],
      },
      {
        path: 'terceros',
        canActivate: [tercerosGuard],
        loadComponent: () =>
          import('./features/terceros/terceros/terceros').then((m) => m.TercerosComponent),
        title: 'Terceros — Tesserak',
      },
      {
        path: 'compras',
        canActivate: [comprasGuard],
        loadComponent: () =>
          import('./features/compras/compras/compras').then((m) => m.ComprasComponent),
        title: 'Compras — Tesserak',
      },
      {
        path: 'ventas',
        canActivate: [ventasGuard],
        loadComponent: () =>
          import('./features/ventas/ventas/ventas').then((m) => m.VentasComponent),
        title: 'Ventas — Tesserak',
      },
      {
        path: 'lealtad',
        canActivate: [operacionGuard],
        loadComponent: () =>
          import('./features/lealtad/miembros-cuentas/miembros-cuentas').then(
            (m) => m.LealtadMiembrosCuentasComponent,
          ),
        title: 'Lealtad — Miembros y cuentas · Tesserak',
      },
      {
        path: 'inventario',
        canActivate: [operacionGuard],
        loadComponent: () =>
          import('./features/inventario/inventario/inventario').then((m) => m.InventarioComponent),
        title: 'Inventarios — Tesserak',
      },
      {
        path: 'gestion-empresa',
        canActivate: [adminOrBranchAdminGuard],
        loadComponent: () =>
          import('./features/administracion/gestion-empresa/gestion-empresa').then(
            (m) => m.GestionEmpresaComponent,
          ),
        title: 'Gestión de empresa — Tesserak',
      },
      {
        path: 'administracion/planes-lealtad',
        canActivate: [adminGuard],
        loadComponent: () =>
          import('./features/administracion/planes-lealtad/planes-lealtad').then(
            (m) => m.PlanesLealtadComponent,
          ),
        title: 'Planes de lealtad — Administración · Tesserak',
      },
      {
        path: 'empresas',
        redirectTo: 'gestion-empresa',
        pathMatch: 'full',
      },
      {
        path: 'usuarios',
        canActivate: [operacionGuard],
        loadComponent: () =>
          import('./features/usuarios/usuarios/usuarios').then((m) => m.UsuariosComponent),
        title: 'Usuarios — Tesserak',
      },
      {
        path: 'mi-perfil',
        loadComponent: () =>
          import('./features/usuarios/mi-perfil/mi-perfil').then((m) => m.MiPerfilComponent),
        title: 'Mi perfil — Tesserak',
      },
      {
        path: '',
        pathMatch: 'full',
        redirectTo: 'dashboard',
      },
    ],
  },
  {
    path: '**',
    redirectTo: 'login',
  },
];
