import { Component, OnInit, computed, inject, signal } from '@angular/core';
import { CurrencyPipe, DatePipe, DecimalPipe } from '@angular/common';
import { RouterLink } from '@angular/router';

import { AuthService } from '../../core/auth/auth.service';
import { DashboardService, HomeDashboardData } from './dashboard.service';

interface ModuloCard {
  titulo: string;
  descripcion: string;
  route?: string;
  estado: 'Disponible' | 'Próximamente';
}

@Component({
  selector: 'app-dashboard',
  standalone: true,
  imports: [RouterLink, CurrencyPipe, DecimalPipe, DatePipe],
  templateUrl: './dashboard.html',
  styleUrl: './dashboard.scss',
})
export class DashboardComponent implements OnInit {
  private readonly auth = inject(AuthService);
  private readonly dashboardSvc = inject(DashboardService);

  readonly usuario = this.auth.usuario;
  readonly loading = signal(true);
  readonly data = signal<HomeDashboardData | null>(null);
  readonly hoy = new Date();

  readonly maxLeadsEtapa = computed(() => {
    const etapas = this.data()?.crm?.por_etapa ?? [];
    return Math.max(1, ...etapas.map((e) => e.cantidad));
  });

  readonly modulos: ModuloCard[] = [
    {
      titulo: 'CRM y Embudo',
      descripcion: 'Leads, etapas del pipeline, actividades de seguimiento y conversión.',
      route: '/crm',
      estado: 'Disponible',
    },
    {
      titulo: 'Empresas',
      descripcion: 'Administración multi-empresa, datos fiscales y configuración general.',
      route: '/empresas',
      estado: 'Disponible',
    },
    {
      titulo: 'Usuarios',
      descripcion: 'Gestión de accesos por sucursal, credenciales y estados.',
      route: '/usuarios',
      estado: 'Disponible',
    },
    {
      titulo: 'Terceros',
      descripcion: 'Clientes, proveedores y empleados bajo catálogo unificado.',
      estado: 'Próximamente',
    },
    {
      titulo: 'Productos e Inventario',
      descripcion: 'Catálogo maestro, bodegas, kardex y control de existencias.',
      estado: 'Próximamente',
    },
    {
      titulo: 'Ventas y Lealtad',
      descripcion: 'Facturación comercial, compras y programa de puntos.',
      estado: 'Próximamente',
    },
  ];

  ngOnInit(): void {
    this.recargar();
  }

  recargar(): void {
    const empresaId = this.usuario()?.empresa_id ?? 1;
    this.loading.set(true);
    this.dashboardSvc.cargarResumen(empresaId).subscribe((res) => {
      this.data.set(res);
      this.loading.set(false);
    });
  }

  porcentajeEtapa(cantidad: number): number {
    return Math.round((cantidad / this.maxLeadsEtapa()) * 100);
  }

  estaVencida(fecha: string | null): boolean {
    if (!fecha) return false;
    return new Date(fecha).getTime() < Date.now();
  }
}
