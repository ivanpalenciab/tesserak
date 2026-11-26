import { Component, OnInit, computed, inject, signal } from '@angular/core';
import { FormArray, FormBuilder, FormGroup, ReactiveFormsModule, Validators } from '@angular/forms';
import { HttpErrorResponse } from '@angular/common/http';
import { catchError, forkJoin, of, switchMap } from 'rxjs';

import { ApiErrorBody } from '../../../core/auth/auth.models';
import { CrmService } from '../../crm/crm.service';
import { Moneda } from '../../crm/crm.models';
import { ProductosService } from '../../productos/productos.service';
import { Producto } from '../../productos/productos.models';
import { TercerosService } from '../../terceros/terceros.service';
import { RolTercero, Tercero } from '../../terceros/terceros.models';
import { ComprasService } from '../compras.service';
import {
  ActualizarCompraRequest,
  Compra,
  CompraDetalleInput,
  CrearCompraRequest,
  EstadoCompra,
} from '../compras.models';

export interface ProveedorOption {
  tercero_rol_id: number;
  tercero_id: number;
  identificacion: string;
  razon_social: string;
}

export const ESTADOS_COMPRA_OPCIONES: { value: EstadoCompra; label: string }[] = [
  { value: 'BORRADOR', label: 'Borrador' },
  { value: 'REGISTRADA', label: 'Registrada' },
  { value: 'PAGADA', label: 'Pagada' },
  { value: 'ANULADA', label: 'Anulada' },
];

@Component({
  selector: 'app-compras',
  standalone: true,
  imports: [ReactiveFormsModule],
  templateUrl: './compras.html',
  styleUrl: './compras.scss',
})
export class ComprasComponent implements OnInit {
  private readonly comprasSvc = inject(ComprasService);
  private readonly productosSvc = inject(ProductosService);
  private readonly tercerosSvc = inject(TercerosService);
  private readonly crmSvc = inject(CrmService);
  private readonly fb = inject(FormBuilder);

  readonly estadosCompra = ESTADOS_COMPRA_OPCIONES;

  // Estado general y alertas
  readonly loading = signal(true);
  readonly loadingDetalle = signal(false);
  readonly saving = signal(false);
  readonly errorMensaje = signal<string | null>(null);
  readonly exitoMensaje = signal<string | null>(null);
  readonly modalError = signal<string | null>(null);

  // Catálogos auxiliares
  readonly monedas = signal<Moneda[]>([]);
  readonly articulos = signal<Producto[]>([]);
  readonly proveedores = signal<ProveedorOption[]>([]);

  // Listado
  readonly compras = signal<Compra[]>([]);
  readonly total = signal(0);
  readonly page = signal(1);
  readonly pageSize = 15;

  // Filtros
  readonly filtroTexto = signal('');
  readonly filtroEstado = signal<EstadoCompra | ''>('');
  readonly filtroMonedaId = signal<number | null>(null);
  readonly filtroFechaDesde = signal('');
  readonly filtroFechaHasta = signal('');

  // Detalle seleccionado y modales
  readonly compraDetalle = signal<Compra | null>(null);
  readonly modalAbierto = signal(false);
  readonly editando = signal<Compra | null>(null);
  readonly aEliminar = signal<Compra | null>(null);

  // Computed
  readonly totalPaginas = computed(() => Math.max(1, Math.ceil(this.total() / this.pageSize)));
  readonly monedasMap = computed(() => new Map(this.monedas().map((m) => [m.id, m])));
  readonly articulosMap = computed(() => new Map(this.articulos().map((a) => [a.id, a])));
  readonly proveedoresMap = computed(
    () => new Map(this.proveedores().map((p) => [p.tercero_rol_id, p])),
  );

  readonly hayFiltrosActivos = computed(
    () =>
      this.filtroTexto().trim().length > 0 ||
      this.filtroEstado() !== '' ||
      this.filtroMonedaId() !== null ||
      this.filtroFechaDesde().trim().length > 0 ||
      this.filtroFechaHasta().trim().length > 0,
  );

  readonly form = this.fb.group({
    numero_factura_proveedor: ['', [Validators.maxLength(50)]],
    tercero_empresa_rol_id: [null as number | null, [Validators.required]],
    moneda_id: [null as number | null],
    estado: ['REGISTRADA' as EstadoCompra, [Validators.required]],
    fecha_compra: [this.hoyISO(), [Validators.required]],
    fecha_vencimiento: [''],
    observaciones: ['', [Validators.maxLength(5000)]],
    detalles: this.fb.array<FormGroup>([]),
  });

  get detallesArray(): FormArray<FormGroup> {
    return this.form.controls.detalles;
  }

  ngOnInit(): void {
    this.cargarCatalogosYCompras();
  }

  cargarCatalogosYCompras(): void {
    this.loading.set(true);
    this.errorMensaje.set(null);

    const proveedores$ = this.tercerosSvc.listarRoles(true).pipe(
      catchError(() => of([] as RolTercero[])),
      switchMap((roles) => {
        const rolProveedor = roles.find((r) => r.codigo === 'PROVEEDOR');
        if (!rolProveedor) {
          return of([] as ProveedorOption[]);
        }
        return this.tercerosSvc
          .listar({ rol_id: rolProveedor.id, activo: true, page: 1, page_size: 100 })
          .pipe(
            catchError(() => of({ data: [] as Tercero[], page: 1, page_size: 100, total: 0 })),
            switchMap((res) => {
              const opciones: ProveedorOption[] = [];
              for (const t of res.data ?? []) {
                const rolAsig = (t.roles ?? []).find(
                  (r) => r.rol_id === rolProveedor.id || r.codigo === 'PROVEEDOR',
                );
                if (rolAsig) {
                  opciones.push({
                    tercero_rol_id: rolAsig.tercero_rol_id,
                    tercero_id: t.id,
                    identificacion: `${t.tipo_identificacion} ${t.numero_identificacion}`,
                    razon_social: t.razon_social,
                  });
                }
              }
              return of(opciones);
            }),
          );
      }),
    );

    forkJoin({
      monedas: this.crmSvc.listarMonedas().pipe(catchError(() => of([] as Moneda[]))),
      articulos: this.productosSvc
        .listar({ estado: true, page: 1, page_size: 100 })
        .pipe(catchError(() => of({ data: [] as Producto[], page: 1, page_size: 100, total: 0 }))),
      proveedores: proveedores$,
    }).subscribe(({ monedas, articulos, proveedores }) => {
      this.monedas.set(monedas);
      this.articulos.set(articulos.data ?? []);
      this.proveedores.set(proveedores);
      this.cargarCompras();
    });
  }

  cargarCompras(): void {
    this.loading.set(true);
    this.errorMensaje.set(null);

    const desde = this.filtroFechaDesde().trim();
    const hasta = this.filtroFechaHasta().trim();
    if (desde && hasta && hasta < desde) {
      this.loading.set(false);
      this.errorMensaje.set('La fecha hasta debe ser mayor o igual a la fecha desde.');
      return;
    }

    this.comprasSvc
      .listar({
        estado: this.filtroEstado(),
        moneda_id: this.filtroMonedaId() ?? undefined,
        fecha_desde: desde || undefined,
        fecha_hasta: hasta || undefined,
        q: this.filtroTexto(),
        page: this.page(),
        page_size: this.pageSize,
      })
      .subscribe({
        next: (res) => {
          this.compras.set(res.data ?? []);
          this.total.set(res.total ?? 0);
          this.loading.set(false);
        },
        error: (err) => {
          this.loading.set(false);
          this.errorMensaje.set(
            this.extraerMensajeError(err, 'No se pudieron cargar las compras.'),
          );
        },
      });
  }

  // =====================================================================
  // Filtros y paginación
  // =====================================================================

  onBuscarTexto(valor: string): void {
    this.filtroTexto.set(valor);
    this.page.set(1);
    this.cargarCompras();
  }

  onFiltroEstado(valor: string): void {
    this.filtroEstado.set((valor || '') as EstadoCompra | '');
    this.page.set(1);
    this.cargarCompras();
  }

  onFiltroMoneda(valor: string): void {
    this.filtroMonedaId.set(valor ? Number(valor) : null);
    this.page.set(1);
    this.cargarCompras();
  }

  onFiltroFechaDesde(valor: string): void {
    this.filtroFechaDesde.set(valor);
    this.page.set(1);
    this.cargarCompras();
  }

  onFiltroFechaHasta(valor: string): void {
    this.filtroFechaHasta.set(valor);
    this.page.set(1);
    this.cargarCompras();
  }

  limpiarFiltros(): void {
    this.filtroTexto.set('');
    this.filtroEstado.set('');
    this.filtroMonedaId.set(null);
    this.filtroFechaDesde.set('');
    this.filtroFechaHasta.set('');
    this.page.set(1);
    this.cargarCompras();
  }

  irPagina(p: number): void {
    if (p < 1 || p > this.totalPaginas()) return;
    this.page.set(p);
    this.cargarCompras();
  }

  // =====================================================================
  // Ver detalle de una compra
  // =====================================================================

  verDetalle(compra: Compra): void {
    this.loadingDetalle.set(true);
    this.compraDetalle.set(compra);
    this.comprasSvc.obtener(compra.id).subscribe({
      next: (completa) => {
        this.compraDetalle.set(completa);
        this.loadingDetalle.set(false);
      },
      error: (err) => {
        this.loadingDetalle.set(false);
        this.errorMensaje.set(
          this.extraerMensajeError(err, 'No se pudo cargar el detalle de la compra.'),
        );
      },
    });
  }

  cerrarDetalle(): void {
    this.compraDetalle.set(null);
  }

  // =====================================================================
  // Crear / Editar compra
  // =====================================================================

  abrirNueva(): void {
    this.editando.set(null);
    this.modalError.set(null);
    this.detallesArray.clear();
    this.agregarLinea();

    const monedaDefault = this.monedas()[0]?.id ?? null;
    this.form.patchValue({
      numero_factura_proveedor: '',
      tercero_empresa_rol_id: null,
      moneda_id: monedaDefault,
      estado: 'REGISTRADA',
      fecha_compra: this.hoyISO(),
      fecha_vencimiento: '',
      observaciones: '',
    });
    this.form.markAsUntouched();
    this.modalAbierto.set(true);
  }

  abrirEditar(compra: Compra, event?: Event): void {
    event?.stopPropagation();
    this.modalError.set(null);
    this.loadingDetalle.set(true);

    this.comprasSvc.obtener(compra.id).subscribe({
      next: (completa) => {
        this.loadingDetalle.set(false);
        this.editando.set(completa);
        this.detallesArray.clear();

        if (completa.detalles?.length) {
          for (const d of completa.detalles) {
            this.detallesArray.push(
              this.crearLineaGroup({
                articulo_id: d.articulo_id,
                descripcion: d.descripcion ?? '',
                cantidad: d.cantidad,
                costo_unitario: d.costo_unitario,
                impuesto: d.impuesto,
              }),
            );
          }
        } else {
          this.agregarLinea();
        }

        this.form.patchValue({
          numero_factura_proveedor: completa.numero_factura_proveedor ?? '',
          tercero_empresa_rol_id: completa.tercero_empresa_rol_id,
          moneda_id: completa.moneda_id,
          estado: completa.estado,
          fecha_compra: this.toDateInput(completa.fecha_compra),
          fecha_vencimiento: completa.fecha_vencimiento
            ? this.toDateInput(completa.fecha_vencimiento)
            : '',
          observaciones: completa.observaciones ?? '',
        });
        this.form.markAsUntouched();
        this.modalAbierto.set(true);
      },
      error: (err) => {
        this.loadingDetalle.set(false);
        this.errorMensaje.set(
          this.extraerMensajeError(err, 'No se pudo cargar la compra para editar.'),
        );
      },
    });
  }

  cerrarModal(): void {
    if (this.saving()) return;
    this.modalAbierto.set(false);
    this.editando.set(null);
    this.modalError.set(null);
  }

  agregarLinea(): void {
    this.detallesArray.push(this.crearLineaGroup());
  }

  eliminarLinea(index: number): void {
    if (this.detallesArray.length <= 1) return;
    this.detallesArray.removeAt(index);
  }

  onCambiarArticuloLinea(index: number, valor: string): void {
    const id = valor ? Number(valor) : null;
    if (!id) return;
    const art = this.articulosMap().get(id);
    const linea = this.detallesArray.at(index);
    const descActual = String(linea.get('descripcion')?.value ?? '').trim();
    if (art && !descActual) {
      linea.get('descripcion')?.setValue(art.nombre);
    }
  }

  subtotalLinea(index: number): number {
    const g = this.detallesArray.at(index);
    const cant = Number(g.get('cantidad')?.value ?? 0);
    const costo = Number(g.get('costo_unitario')?.value ?? 0);
    return Math.max(0, cant * costo);
  }

  totalLinea(index: number): number {
    const g = this.detallesArray.at(index);
    const imp = Number(g.get('impuesto')?.value ?? 0);
    return this.subtotalLinea(index) + Math.max(0, imp);
  }

  resumenFormulario(): { subtotal: number; totalIva: number; total: number } {
    let subtotal = 0;
    let totalIva = 0;
    for (let i = 0; i < this.detallesArray.length; i++) {
      const sub = this.subtotalLinea(i);
      const imp = Math.max(0, Number(this.detallesArray.at(i).get('impuesto')?.value ?? 0));
      subtotal += sub;
      totalIva += imp;
    }
    return { subtotal, totalIva, total: subtotal + totalIva };
  }

  guardar(): void {
    this.modalError.set(null);
    if (this.form.invalid || this.detallesArray.length === 0) {
      this.form.markAllAsTouched();
      return;
    }

    const v = this.form.getRawValue();
    const fechaCompra = (v.fecha_compra ?? '').trim();
    const fechaVenc = (v.fecha_vencimiento ?? '').trim();
    if (fechaVenc && fechaVenc < fechaCompra) {
      this.modalError.set('La fecha de vencimiento debe ser mayor o igual a la fecha de compra.');
      return;
    }

    const detalles: CompraDetalleInput[] = this.detallesArray.controls.map((g) => ({
      articulo_id: Number(g.get('articulo_id')?.value),
      descripcion: String(g.get('descripcion')?.value ?? '').trim() || null,
      cantidad: Number(g.get('cantidad')?.value),
      costo_unitario: Number(g.get('costo_unitario')?.value),
      impuesto: Number(g.get('impuesto')?.value ?? 0),
    }));

    const proveedorId = v.tercero_empresa_rol_id ? Number(v.tercero_empresa_rol_id) : null;
    const proveedorSel = proveedorId ? this.proveedoresMap().get(proveedorId) : undefined;
    const proveedorNombre = proveedorSel?.razon_social ?? null;

    this.saving.set(true);
    const actual = this.editando();

    if (actual) {
      const payload: ActualizarCompraRequest = {
        numero_factura_proveedor: (v.numero_factura_proveedor ?? '').trim() || null,
        tercero_empresa_rol_id: proveedorId,
        proveedor_nombre: proveedorNombre ?? actual.proveedor_nombre,
        moneda_id: v.moneda_id ? Number(v.moneda_id) : null,
        estado: v.estado as EstadoCompra,
        fecha_compra: fechaCompra,
        observaciones: (v.observaciones ?? '').trim() || null,
        detalles,
      };
      if (fechaVenc) {
        payload.fecha_vencimiento = fechaVenc;
      } else if (actual.fecha_vencimiento) {
        payload.limpiar_fecha_vencimiento = true;
      }

      this.comprasSvc.actualizar(actual.id, payload).subscribe({
        next: (actualizada) => {
          this.saving.set(false);
          this.modalAbierto.set(false);
          this.editando.set(null);
          if (this.compraDetalle()?.id === actualizada.id) {
            this.compraDetalle.set(actualizada);
          }
          this.mostrarExito(`Compra "${actualizada.consecutivo}" actualizada correctamente.`);
          this.cargarCompras();
        },
        error: (err) => {
          this.saving.set(false);
          this.modalError.set(
            this.extraerMensajeError(err, 'No se pudo actualizar la compra.'),
          );
        },
      });
    } else {
      const payload: CrearCompraRequest = {
        numero_factura_proveedor: (v.numero_factura_proveedor ?? '').trim() || null,
        tercero_empresa_rol_id: proveedorId,
        proveedor_nombre: proveedorNombre,
        moneda_id: v.moneda_id ? Number(v.moneda_id) : null,
        estado: v.estado as EstadoCompra,
        fecha_compra: fechaCompra,
        fecha_vencimiento: fechaVenc || null,
        observaciones: (v.observaciones ?? '').trim() || null,
        detalles,
      };

      this.comprasSvc.crear(payload).subscribe({
        next: (creada) => {
          this.saving.set(false);
          this.modalAbierto.set(false);
          this.mostrarExito(`Compra "${creada.consecutivo}" registrada correctamente.`);
          this.page.set(1);
          this.cargarCompras();
        },
        error: (err) => {
          this.saving.set(false);
          this.modalError.set(this.extraerMensajeError(err, 'No se pudo registrar la compra.'));
        },
      });
    }
  }

  // =====================================================================
  // Eliminar (Soft delete)
  // =====================================================================

  solicitarEliminar(compra: Compra, event?: Event): void {
    event?.stopPropagation();
    this.aEliminar.set(compra);
  }

  cancelarEliminar(): void {
    if (this.saving()) return;
    this.aEliminar.set(null);
  }

  confirmarEliminar(): void {
    const target = this.aEliminar();
    if (!target) return;

    this.saving.set(true);
    this.comprasSvc.eliminar(target.id).subscribe({
      next: () => {
        this.saving.set(false);
        this.aEliminar.set(null);
        if (this.compraDetalle()?.id === target.id) {
          this.compraDetalle.set(null);
        }
        this.mostrarExito(`Compra "${target.consecutivo}" eliminada.`);
        this.cargarCompras();
      },
      error: (err) => {
        this.saving.set(false);
        this.aEliminar.set(null);
        this.errorMensaje.set(this.extraerMensajeError(err, 'No se pudo eliminar la compra.'));
      },
    });
  }

  // =====================================================================
  // Helpers
  // =====================================================================

  etiquetaEstado(estado: EstadoCompra): string {
    return this.estadosCompra.find((e) => e.value === estado)?.label ?? estado;
  }

  formatearFecha(iso: string | null): string {
    if (!iso) return '—';
    return iso.slice(0, 10);
  }

  formatearMonto(monto: number, compra?: Compra | null): string {
    const simbolo =
      compra?.moneda_simbolo ??
      (compra?.moneda_id ? this.monedasMap().get(compra.moneda_id)?.simbolo : undefined) ??
      '$';
    const codigo =
      compra?.moneda_codigo ??
      (compra?.moneda_id ? this.monedasMap().get(compra.moneda_id)?.codigo_iso : undefined) ??
      '';
    const formatted = new Intl.NumberFormat('es-CO', {
      minimumFractionDigits: 2,
      maximumFractionDigits: 2,
    }).format(monto);
    return `${simbolo} ${formatted}${codigo ? ' ' + codigo : ''}`;
  }

  formatearMontoForm(monto: number): string {
    const monedaId = this.form.controls.moneda_id.value;
    const m = monedaId ? this.monedasMap().get(monedaId) : undefined;
    const simbolo = m?.simbolo ?? '$';
    const codigo = m?.codigo_iso ?? '';
    const formatted = new Intl.NumberFormat('es-CO', {
      minimumFractionDigits: 2,
      maximumFractionDigits: 2,
    }).format(monto);
    return `${simbolo} ${formatted}${codigo ? ' ' + codigo : ''}`;
  }

  private crearLineaGroup(init?: {
    articulo_id?: number | null;
    descripcion?: string;
    cantidad?: number;
    costo_unitario?: number;
    impuesto?: number;
  }): FormGroup {
    return this.fb.group({
      articulo_id: [init?.articulo_id ?? null, [Validators.required]],
      descripcion: [init?.descripcion ?? '', [Validators.maxLength(255)]],
      cantidad: [init?.cantidad ?? 1, [Validators.required, Validators.min(0.0001)]],
      costo_unitario: [init?.costo_unitario ?? 0, [Validators.required, Validators.min(0)]],
      impuesto: [init?.impuesto ?? 0, [Validators.min(0)]],
    });
  }

  private toDateInput(iso: string): string {
    return iso ? iso.slice(0, 10) : '';
  }

  private hoyISO(): string {
    return new Date().toISOString().slice(0, 10);
  }

  private mostrarExito(msg: string): void {
    this.exitoMensaje.set(msg);
    setTimeout(() => {
      if (this.exitoMensaje() === msg) {
        this.exitoMensaje.set(null);
      }
    }, 4500);
  }

  private extraerMensajeError(err: unknown, fallback: string): string {
    if (err instanceof HttpErrorResponse) {
      const body = err.error as ApiErrorBody | null;
      if (body?.error?.message) return body.error.message;
    }
    return fallback;
  }
}
