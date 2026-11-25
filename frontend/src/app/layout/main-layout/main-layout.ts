import { Component, computed, inject, signal } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { NavigationEnd, Router, RouterLink, RouterLinkActive, RouterOutlet } from '@angular/router';
import { filter } from 'rxjs';

import { AuthService } from '../../core/auth/auth.service';

interface NavSubItem {
  label: string;
  route?: string;
  badge?: string;
  disabled?: boolean;
}

interface NavItem {
  label: string;
  route?: string;
  icon: string;
  badge?: string;
  disabled?: boolean;
  children?: NavSubItem[];
}

interface NavSection {
  title: string;
  items: NavItem[];
}

@Component({
  selector: 'app-main-layout',
  standalone: true,
  imports: [RouterOutlet, RouterLink, RouterLinkActive],
  templateUrl: './main-layout.html',
  styleUrl: './main-layout.scss',
})
export class MainLayoutComponent {
  private readonly auth = inject(AuthService);
  private readonly router = inject(Router);

  readonly usuario = this.auth.usuario;
  readonly mobileMenuOpen = signal(false);
  readonly currentUrl = signal(this.router.url);
  readonly expandedMenus = signal<Record<string, boolean>>({});

  readonly iniciales = computed(() => {
    const u = this.usuario();
    if (!u) return 'TS';
    const n = u.nombres?.trim().charAt(0) ?? '';
    const a = u.apellidos?.trim().charAt(0) ?? '';
    return (n + a).toUpperCase() || 'TS';
  });

  readonly navSections: NavSection[] = [
    {
      title: 'Principal',
      items: [{ label: 'Inicio', route: '/dashboard', icon: 'home' }],
    },
    {
      title: 'Comercial',
      items: [
        {
          label: 'CRM',
          icon: 'crm',
          children: [
            { label: 'Resumen comercial', disabled: true, badge: 'Pronto' },
            { label: 'Leads', route: '/crm/leads' },
            { label: 'Empresas y contactos', disabled: true, badge: 'Pronto' },
            { label: 'Oportunidades', route: '/crm/oportunidades' },
            { label: 'Pipeline comercial', disabled: true, badge: 'Pronto' },
            { label: 'Actividades', disabled: true, badge: 'Pronto' },
            { label: 'Campañas', route: '/crm/campanas' },
            { label: 'Configuración', route: '/crm/configuracion' },
          ],
        },
        { label: 'Ventas', icon: 'sales', disabled: true, badge: 'Pronto' },
        { label: 'Lealtad', icon: 'loyalty', disabled: true, badge: 'Pronto' },
      ],
    },
    {
      title: 'Operación',
      items: [
        { label: 'Terceros', icon: 'contacts', disabled: true, badge: 'Pronto' },
        {
          label: 'Productos',
          icon: 'products',
          children: [
            { label: 'Catálogo', route: '/productos/catalogo' },
            { label: 'Categorías', route: '/productos/categorias' },
            { label: 'Configuración', route: '/productos/configuracion' },
          ],
        },
        { label: 'Inventario', icon: 'inventory', disabled: true, badge: 'Pronto' },
      ],
    },
    {
      title: 'Administración',
      items: [
        { label: 'Empresas', route: '/empresas', icon: 'company', badge: 'API' },
        { label: 'Usuarios', route: '/usuarios', icon: 'users', badge: 'API' },
      ],
    },
  ];

  constructor() {
    this.syncExpandedWithUrl(this.router.url);

    this.router.events
      .pipe(
        filter((event): event is NavigationEnd => event instanceof NavigationEnd),
        takeUntilDestroyed(),
      )
      .subscribe((event) => {
        this.currentUrl.set(event.urlAfterRedirects);
        this.syncExpandedWithUrl(event.urlAfterRedirects);
      });
  }

  toggleSubmenu(label: string): void {
    this.expandedMenus.update((state) => ({
      ...state,
      [label]: !state[label],
    }));
  }

  isSubmenuOpen(label: string): boolean {
    return !!this.expandedMenus()[label];
  }

  isGroupActive(item: NavItem): boolean {
    if (!item.children?.length) return false;
    const url = this.currentUrl();
    return item.children.some((child) => !!child.route && url.startsWith(child.route));
  }

  toggleMobileMenu(): void {
    this.mobileMenuOpen.update((v) => !v);
  }

  closeMobileMenu(): void {
    this.mobileMenuOpen.set(false);
  }

  logout(): void {
    this.auth.logout();
  }

  private syncExpandedWithUrl(url: string): void {
    for (const section of this.navSections) {
      for (const item of section.items) {
        if (item.children?.some((child) => !!child.route && url.startsWith(child.route))) {
          this.expandedMenus.update((state) =>
            state[item.label] ? state : { ...state, [item.label]: true },
          );
        }
      }
    }
  }
}
