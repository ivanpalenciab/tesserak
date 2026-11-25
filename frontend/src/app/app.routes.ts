import { Routes } from '@angular/router';

import { authGuard, guestGuard } from './core/auth/auth.guard';

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
        path: 'empresas',
        loadComponent: () =>
          import('./features/coming-soon/coming-soon').then((m) => m.ComingSoonComponent),
        data: {
          titulo: 'Gestión de Empresas',
          descripcion:
            'El backend CRUD de empresas está activo en /api/v1/empresas. Aquí agregaremos el listado, creación y edición de empresas.',
        },
        title: 'Empresas — Tesserak',
      },
      {
        path: 'usuarios',
        loadComponent: () =>
          import('./features/coming-soon/coming-soon').then((m) => m.ComingSoonComponent),
        data: {
          titulo: 'Gestión de Usuarios',
          descripcion:
            'El backend CRUD de usuarios está activo en /api/v1/usuarios. Aquí agregaremos el listado por sucursal, creación y control de estado.',
        },
        title: 'Usuarios — Tesserak',
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
