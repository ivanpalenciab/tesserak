import { Component, OnInit, computed, inject, signal } from '@angular/core';
import { FormArray, FormBuilder, FormGroup, ReactiveFormsModule, Validators } from '@angular/forms';
import { HttpErrorResponse } from '@angular/common/http';
import { catchError, forkJoin, of } from 'rxjs';

import { AuthService } from '../../../core/auth/auth.service';
import { ApiErrorBody } from '../../../core/auth/auth.models';
import { ProductosService } from '../../productos/productos.service';
import { GrupoImpuesto, Producto } from '../../productos/productos.models';
import { TercerosService } from '../../terceros/terceros.service';
import { Tercero } from '../../terceros/terceros.models';
import { InventarioService } from '../../inventario/inventario.service';
import { Ubicacion } from '../../inventario/inventario.models';
import { PreciosService } from '../../precios/precios.service';
import { DetallePrecio, ListaPrecio } from '../../precios/precios.models';
import { LealtadOperacionService } from '../../lealtad/lealtad.service';
import { CuentaClienteLealtad, ProgramaOperacionLealtad } from '../../lealtad/lealtad.models';
import { VentasService } from '../ventas.service';
import {
  CrearVentaRequest,
  EstadoVenta,
  Venta,
  VentaDetalleInput,
} from '../ventas.models';

export const CONSUMIDOR_FINAL_DOC = '222222222222';

export const ESTADOS_VENTA: { value: EstadoVenta; label: string }[] = [
  { value: 'REGISTRADA', label: 'Registrada' },
  { value: 'PAGADA', label: 'Pagada' },
  { value: 'BORRADOR', label: 'Borrador' },
  { value: 'ANULADA', label: 'Anulada' },
];

@Component({
  selector: 'app-ventas',
  standalone: true,
  imports: [ReactiveFormsModule],
  templateUrl: './ventas.html',
  styleUrl: './ventas.scss',
})
export class VentasComponent implements OnInit {
  private readonly svc = inject(VentasService);
  private readonly productosSvc = inject(ProductosService);
  private readonly tercerosSvc = inject(TercerosService);
  private readonly invSvc = inject(InventarioService);
  private readonly preciosSvc = inject(PreciosService);
  private readonly lealtadSvc = inject(LealtadOperacionService);
  private readonly auth = inject(AuthService);
  private readonly fb = inject(FormBuilder);

  readonly puedeEliminarVentas = this.auth.puedeEliminarVentas;
  readonly estados = ESTADOS_VENTA;

  // Listado y filtros
  readonly loading = signal(true);
  readonly ventas = signal<Venta[]>([]);
  readonly total = signal(0);
  readonly page = signal(1);
  readonly pageSize = 15;

  readonly filtroTexto = signal('');
  readonly filtroEstado = signal<EstadoVenta | ''>('');
  readonly filtroFechaDesde = signal('');
  readonly filtroFechaHasta = signal('');

  // Catálogos
  readonly articulos = signal<Producto[]>([]);
  readonly gruposImpuesto = signal<GrupoImpuesto[]>([]);
  readonly clientes = signal<Tercero[]>([]);
  readonly ubicaciones = signal<Ubicacion[]>([]);
  readonly listasPrecios = signal<ListaPrecio[]>([]);
  readonly listaPrecioSeleccionadaId = signal<number | null>(null);
  readonly detallesPrecio = signal<DetallePrecio[]>([]);

  // Lealtad del cliente seleccionado (para redención en nueva venta)
  readonly programasLealtad = signal<ProgramaOperacionLealtad[]>([]);
  readonly empresaAdheridaLealtad = computed(() =>
    this.programasLealtad().some((p) => p.estado === 'ACTIVO' && p.permite_redimir),
  );
  readonly planLealtadPrincipal = computed(
    () =>
      this.programasLealtad().find((p) => p.estado === 'ACTIVO' && p.es_programa_principal) ??
      this.programasLealtad().find((p) => p.estado === 'ACTIVO') ??
      null,
  );
  readonly cuentaLealtadCliente = signal<CuentaClienteLealtad | null>(null);
  readonly cargandoCuentaLealtad = signal(false);
  readonly redimirActivo = signal(false);
  readonly puntosRedimir = signal<number>(0);

  // Alertas
  readonly errorMensaje = signal<string | null>(null);
  readonly exitoMensaje = signal<string | null>(null);

  // Detalle seleccionado
  readonly ventaDetalle = signal<Venta | null>(null);
  readonly loadingDetalle = signal(false);

  // Modal crear / editar
  readonly modalAbierto = signal(false);
  readonly editando = signal<Venta | null>(null);
  readonly saving = signal(false);
  readonly modalError = signal<string | null>(null);

  // Modal eliminar (soft delete)
  readonly ventaAEliminar = signal<Venta | null>(null);

  readonly totalPaginas = computed(() => Math.max(1, Math.ceil(this.total() / this.pageSize)));
  readonly articulosMap = computed(() => new Map(this.articulos().map((a) => [a.id, a])));
  readonly gruposImpuestoMap = computed(() => new Map(this.gruposImpuesto().map((g) => [g.id, g])));
  readonly consumidorFinal = computed(
    () =>
      this.clientes().find((c) => c.numero_identificacion === CONSUMIDOR_FINAL_DOC) ??
      this.clientes()[0] ??
      null,
  );

  readonly form = this.fb.group({
    cliente_id: [null as number | null, [Validators.required]],
    ubicacion_id: [null as number | null, [Validators.required]],
    estado: ['REGISTRADA' as EstadoVenta, [Validators.required]],
    fecha_venta: [this.hoyISO(), [Validators.required]],
    fecha_vencimiento: [''],
    observaciones: ['', [Validators.maxLength(5000)]],
    detalles: this.fb.array<FormGroup>([]),
  });

  get detallesArray(): FormArray<FormGroup> {
    return this.form.controls.detalles;
  }

  ngOnInit(): void {
    this.cargarCatalogos();
    this.cargarVentas();
  }

  cargarCatalogos(): void {
    forkJoin({
      articulos: this.productosSvc
        .listar({ estado: true, page: 1, page_size: 100 })
        .pipe(catchError(() => of({ data: [] as Producto[], page: 1, page_size: 100, total: 0 }))),
      gruposImpuesto: this.productosSvc
        .listarGruposImpuesto()
        .pipe(catchError(() => of([] as GrupoImpuesto[]))),
      terceros: this.tercerosSvc
        .listar({ activo: true, page: 1, page_size: 100 })
        .pipe(catchError(() => of({ data: [] as Tercero[], page: 1, page_size: 100, total: 0 }))),
      ubicaciones: this.invSvc
        .listarUbicaciones(true)
        .pipe(catchError(() => of([] as Ubicacion[]))),
      listasPrecios: this.preciosSvc
        .listarListas({ estado: true, page: 1, page_size: 100 })
        .pipe(catchError(() => of({ data: [] as ListaPrecio[], page: 1, page_size: 100, total: 0 }))),
      programasLealtad: this.lealtadSvc
        .listarProgramasAdheridos()
        .pipe(catchError(() => of([] as ProgramaOperacionLealtad[]))),
    }).subscribe(({ articulos, gruposImpuesto, terceros, ubicaciones, listasPrecios, programasLealtad }) => {
      this.articulos.set(articulos.data ?? []);
      this.gruposImpuesto.set(gruposImpuesto);
      const soloClientes = (terceros.data ?? []).filter((t) =>
        (t.roles ?? []).some((r) => r.codigo === 'CLIENTE'),
      );
      this.clientes.set(soloClientes);
      this.ubicaciones.set(ubicaciones);
      this.programasLealtad.set(programasLealtad ?? []);

      const listas = listasPrecios.data ?? [];
      this.listasPrecios.set(listas);
      const listaDefault = this.seleccionarListaPorDefecto(listas);
      if (listaDefault) {
        this.seleccionarListaPrecio(listaDefault.id, false);
      }
    });
  }

  onListaPrecioChange(valor: string | number | null): void {
    const id = valor ? Number(valor) : null;
    if (!id) {
      this.listaPrecioSeleccionadaId.set(null);
      this.detallesPrecio.set([]);
      return;
    }
    this.seleccionarListaPrecio(id, true);
  }

  onArticuloChange(index: number): void {
    const g = this.detallesArray.at(index);
    if (!g) return;
    this.aplicarPrecioALinea(g);
  }

  onCantidadChange(index: number): void {
    const g = this.detallesArray.at(index);
    if (!g) return;
    this.aplicarPrecioALinea(g);
  }

  onFechaVentaChange(): void {
    this.recalcularPreciosLineas();
  }

  private seleccionarListaPorDefecto(listas: ListaPrecio[]): ListaPrecio | null {
    if (listas.length === 0) return null;
    const hoy = this.hoyISO();
    const vigentes = listas.filter(
      (l) =>
        (!l.vigencia_desde || hoy >= l.vigencia_desde.slice(0, 10)) &&
        (!l.vigencia_hasta || hoy <= l.vigencia_hasta.slice(0, 10)),
    );
    const pool = vigentes.length > 0 ? vigentes : listas;
    return pool.find((l) => l.tipo_lista === 'GENERAL') ?? pool[0] ?? null;
  }

  private seleccionarListaPrecio(listaId: number, recalcularLineas: boolean): void {
    this.listaPrecioSeleccionadaId.set(listaId);
    this.preciosSvc
      .listarDetalles(listaId, { page: 1, page_size: 100 })
      .pipe(catchError(() => of({ data: [] as DetallePrecio[], page: 1, page_size: 100, total: 0 })))
      .subscribe((res) => {
        this.detallesPrecio.set(res.data ?? []);
        if (recalcularLineas) {
          this.recalcularPreciosLineas();
        }
      });
  }

  private recalcularPreciosLineas(): void {
    for (const g of this.detallesArray.controls) {
      this.aplicarPrecioALinea(g);
    }
  }

  private aplicarPrecioALinea(g: FormGroup): void {
    const artIdRaw = g.get('articulo_id')?.value;
    if (!artIdRaw) return;
    const precio = this.resolverPrecioArticulo(
      Number(artIdRaw),
      Number(g.get('cantidad')?.value ?? 1),
    );
    if (precio !== null) {
      g.get('precio_unitario')?.setValue(precio);
    }
  }

  private resolverPrecioArticulo(articuloId: number, cantidad: number): number | null {
    const candidatos = this.detallesPrecio().filter((d) => d.articulo_id === articuloId);
    if (candidatos.length === 0) return null;

    const fechaVenta = (this.form.controls.fecha_venta.value || this.hoyISO()).slice(0, 10);
    const vigentes = candidatos.filter(
      (d) =>
        (!d.vigencia_desde || fechaVenta >= d.vigencia_desde.slice(0, 10)) &&
        (!d.vigencia_hasta || fechaVenta <= d.vigencia_hasta.slice(0, 10)),
    );
    const pool = vigentes.length > 0 ? vigentes : candidatos;

    const cant = cantidad > 0 ? cantidad : 1;
    const aplicables = pool.filter((d) => cant >= Number(d.cantidad_minima ?? 0));
    if (aplicables.length > 0) {
      aplicables.sort((a, b) => Number(b.cantidad_minima ?? 0) - Number(a.cantidad_minima ?? 0));
      return Number(aplicables[0].precio);
    }

    const ordenados = [...pool].sort(
      (a, b) => Number(a.cantidad_minima ?? 0) - Number(b.cantidad_minima ?? 0),
    );
    return Number(ordenados[0].precio);
  }

  cargarVentas(): void {
    this.loading.set(true);
    this.errorMensaje.set(null);

    this.svc
      .listar({
        estado: this.filtroEstado(),
        fecha_desde: this.filtroFechaDesde(),
        fecha_hasta: this.filtroFechaHasta(),
        q: this.filtroTexto(),
        page: this.page(),
        page_size: this.pageSize,
      })
      .subscribe({
        next: (res) => {
          this.ventas.set(res.data ?? []);
          this.total.set(res.total ?? 0);
          this.loading.set(false);
        },
        error: (err) => {
          this.loading.set(false);
          this.errorMensaje.set(this.extraerMensajeError(err, 'No se pudieron cargar las ventas.'));
        },
      });
  }

  onBuscar(valor: string): void {
    this.filtroTexto.set(valor);
    this.page.set(1);
    this.cargarVentas();
  }

  onFiltrarEstado(valor: string): void {
    this.filtroEstado.set((valor || '') as EstadoVenta | '');
    this.page.set(1);
    this.cargarVentas();
  }

  onFiltrarFechaDesde(valor: string): void {
    this.filtroFechaDesde.set(valor);
    this.page.set(1);
    this.cargarVentas();
  }

  onFiltrarFechaHasta(valor: string): void {
    this.filtroFechaHasta.set(valor);
    this.page.set(1);
    this.cargarVentas();
  }

  irPagina(p: number): void {
    if (p < 1 || p > this.totalPaginas()) return;
    this.page.set(p);
    this.cargarVentas();
  }

  verDetalle(v: Venta): void {
    this.ventaDetalle.set(v);
    this.loadingDetalle.set(true);
    this.svc.obtener(v.id).subscribe({
      next: (completa) => {
        this.ventaDetalle.set(completa);
        this.loadingDetalle.set(false);
      },
      error: (err) => {
        this.loadingDetalle.set(false);
        this.errorMensaje.set(
          this.extraerMensajeError(err, 'No se pudo cargar el detalle de la venta.'),
        );
      },
    });
  }

  cerrarDetalle(): void {
    this.ventaDetalle.set(null);
  }

  abrirNuevaVenta(): void {
    this.editando.set(null);
    this.modalError.set(null);
    this.cuentaLealtadCliente.set(null);
    this.redimirActivo.set(false);
    this.puntosRedimir.set(0);
    this.detallesArray.clear();

    const cf = this.consumidorFinal();
    const ubicDefault = this.ubicaciones()[0]?.id ?? null;

    this.form.reset({
      cliente_id: cf?.id ?? null,
      ubicacion_id: ubicDefault,
      estado: 'REGISTRADA',
      fecha_venta: this.hoyISO(),
      fecha_vencimiento: '',
      observaciones: '',
    });

    this.agregarLinea();
    this.modalAbierto.set(true);
    if (cf?.id && this.empresaAdheridaLealtad()) {
      this.cargarCuentaLealtadCliente(cf.id);
    }
  }

  esClienteConsumidorFinal(): boolean {
    const cliId = this.form.controls.cliente_id.value;
    if (!cliId) return false;
    const cli = this.clientes().find((c) => c.id === Number(cliId));
    return cli?.numero_identificacion === CONSUMIDOR_FINAL_DOC;
  }

  onClienteChange(): void {
    this.redimirActivo.set(false);
    this.puntosRedimir.set(0);
    const cliId = this.form.controls.cliente_id.value;
    if (!cliId || this.editando() || !this.empresaAdheridaLealtad()) {
      this.cuentaLealtadCliente.set(null);
      return;
    }
    this.cargarCuentaLealtadCliente(Number(cliId));
  }

  onToggleRedimir(checked: boolean): void {
    this.redimirActivo.set(checked);
    if (!checked) {
      this.puntosRedimir.set(0);
    } else if (this.puntosRedimir() <= 0) {
      this.puntosRedimir.set(this.maxPuntosRedimibles());
    }
  }

  onPuntosRedimirChange(valor: string): void {
    const num = Math.max(0, Math.floor(Number(valor) || 0));
    const maxPts = this.maxPuntosRedimibles();
    this.puntosRedimir.set(Math.min(num, maxPts));
  }

  usarMaximosPuntos(): void {
    this.redimirActivo.set(true);
    this.puntosRedimir.set(this.maxPuntosRedimibles());
  }

  maxPuntosRedimibles(): number {
    const cuenta = this.cuentaLealtadCliente();
    if (!cuenta || !cuenta.tiene_cuenta || !cuenta.permite_redimir || cuenta.saldo_puntos <= 0) {
      return 0;
    }
    const valorPunto = cuenta.valor_punto_redencion || 1;
    const totalBruto = this.calcularSubtotal() + this.calcularIVA();
    const maxPorTotal = Math.floor(totalBruto / valorPunto);
    return Math.max(0, Math.min(cuenta.saldo_puntos, maxPorTotal));
  }

  calcularDescuentoLealtad(): number {
    if (this.editando()) return Number(this.editando()?.descuento_lealtad ?? 0);
    if (!this.redimirActivo()) return 0;
    const cuenta = this.cuentaLealtadCliente();
    const pts = this.puntosRedimir();
    if (!cuenta || !cuenta.tiene_cuenta || pts <= 0) return 0;
    const desc = Math.round(pts * (cuenta.valor_punto_redencion || 1) * 100) / 100;
    const totalBruto = this.calcularSubtotal() + this.calcularIVA();
    return Math.min(desc, totalBruto);
  }

  private cargarCuentaLealtadCliente(terceroId: number): void {
    this.cargandoCuentaLealtad.set(true);
    this.lealtadSvc
      .consultarCuentaCliente(terceroId)
      .pipe(catchError(() => of(null)))
      .subscribe((res) => {
        this.cargandoCuentaLealtad.set(false);
        this.cuentaLealtadCliente.set(res?.tiene_cuenta ? res : null);
      });
  }

  abrirEditarVenta(v: Venta): void {
    this.modalError.set(null);
    this.svc.obtener(v.id).subscribe({
      next: (completa) => {
        this.editando.set(completa);
        this.detallesArray.clear();
        this.form.reset({
          cliente_id: completa.cliente_id,
          ubicacion_id: completa.ubicacion_id,
          estado: completa.estado,
          fecha_venta: this.formatearFecha(completa.fecha_venta),
          fecha_vencimiento: completa.fecha_vencimiento
            ? this.formatearFecha(completa.fecha_vencimiento)
            : '',
          observaciones: completa.observaciones ?? '',
        });
        const lineas = completa.detalles ?? [];
        if (lineas.length === 0) {
          this.agregarLinea();
        } else {
          for (const d of lineas) {
            this.detallesArray.push(
              this.crearLineaGroup({
                articulo_id: d.articulo_id,
                descripcion: d.descripcion ?? '',
                cantidad: d.cantidad,
                precio_unitario: d.precio_unitario,
                impuesto: d.impuesto,
              }),
            );
          }
        }
        this.modalAbierto.set(true);
      },
      error: (err) => {
        this.errorMensaje.set(
          this.extraerMensajeError(err, 'No se pudo cargar la venta para editar.'),
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

  tarifaLinea(g: FormGroup): number {
    const artId = g.get('articulo_id')?.value;
    if (!artId) return 0;
    const art = this.articulosMap().get(Number(artId));
    if (!art) return 0;
    const grupo = this.gruposImpuestoMap().get(art.grupo_impuesto_id);
    return grupo ? Number(grupo.tarifa ?? 0) : 0;
  }

  impuestoLinea(g: FormGroup): number {
    const cant = Number(g.get('cantidad')?.value ?? 0);
    const precio = Number(g.get('precio_unitario')?.value ?? 0);
    const tarifa = this.tarifaLinea(g);
    return Math.round(cant * precio * (tarifa / 100) * 100) / 100;
  }

  calcularSubtotal(): number {
    return this.detallesArray.controls.reduce((acc, g) => {
      const cant = Number(g.get('cantidad')?.value ?? 0);
      const precio = Number(g.get('precio_unitario')?.value ?? 0);
      return acc + cant * precio;
    }, 0);
  }

  calcularIVA(): number {
    return this.detallesArray.controls.reduce((acc, g) => acc + this.impuestoLinea(g), 0);
  }

  calcularTotal(): number {
    return Math.max(0, this.calcularSubtotal() + this.calcularIVA() - this.calcularDescuentoLealtad());
  }

  guardarVenta(): void {
    this.modalError.set(null);
    if (this.form.invalid || this.detallesArray.length === 0) {
      this.form.markAllAsTouched();
      return;
    }

    const v = this.form.getRawValue();
    const detalles: VentaDetalleInput[] = this.detallesArray.controls.map((g) => ({
      articulo_id: Number(g.get('articulo_id')?.value),
      descripcion: String(g.get('descripcion')?.value ?? '').trim() || null,
      cantidad: Number(g.get('cantidad')?.value),
      precio_unitario: Number(g.get('precio_unitario')?.value ?? 0),
      impuesto: this.impuestoLinea(g),
    }));

    this.saving.set(true);
    const actual = this.editando();

    const req$ = actual
      ? this.svc.actualizar(actual.id, {
          cliente_id: v.cliente_id ? Number(v.cliente_id) : undefined,
          ubicacion_id: v.ubicacion_id ? Number(v.ubicacion_id) : undefined,
          estado: v.estado as EstadoVenta,
          fecha_venta: v.fecha_venta ?? undefined,
          fecha_vencimiento: v.fecha_vencimiento?.trim() || null,
          limpiar_fecha_vencimiento: !v.fecha_vencimiento?.trim(),
          observaciones: (v.observaciones ?? '').trim() || null,
          detalles,
        })
      : this.svc.crear({
          cliente_id: v.cliente_id ? Number(v.cliente_id) : null,
          ubicacion_id: v.ubicacion_id ? Number(v.ubicacion_id) : null,
          estado: v.estado as EstadoVenta,
          fecha_venta: v.fecha_venta ?? this.hoyISO(),
          fecha_vencimiento: v.fecha_vencimiento?.trim() || null,
          observaciones: (v.observaciones ?? '').trim() || null,
          puntos_redimir: this.redimirActivo() && this.puntosRedimir() > 0 ? this.puntosRedimir() : null,
          detalles,
        } as CrearVentaRequest);

    req$.subscribe({
      next: (guardada) => {
        this.saving.set(false);
        this.modalAbierto.set(false);
        this.editando.set(null);
        this.mostrarExito(
          actual
            ? `Venta ${guardada.consecutivo} actualizada correctamente.`
            : `Venta ${guardada.consecutivo} registrada y descontada del inventario.`,
        );
        this.cargarVentas();
        if (this.ventaDetalle()?.id === guardada.id) {
          this.ventaDetalle.set(guardada);
        }
      },
      error: (err) => {
        this.saving.set(false);
        this.modalError.set(
          this.extraerMensajeError(err, 'No se pudo guardar la venta. Se revirtió la transacción.'),
        );
      },
    });
  }

  solicitarEliminar(v: Venta): void {
    this.ventaAEliminar.set(v);
  }

  cancelarEliminar(): void {
    if (this.saving()) return;
    this.ventaAEliminar.set(null);
  }

  confirmarEliminar(): void {
    const target = this.ventaAEliminar();
    if (!target) return;
    this.saving.set(true);
    this.svc.eliminar(target.id).subscribe({
      next: () => {
        this.saving.set(false);
        this.ventaAEliminar.set(null);
        if (this.ventaDetalle()?.id === target.id) {
          this.ventaDetalle.set(null);
        }
        this.mostrarExito(`Venta ${target.consecutivo} eliminada.`);
        this.cargarVentas();
      },
      error: (err) => {
        this.saving.set(false);
        this.ventaAEliminar.set(null);
        this.errorMensaje.set(this.extraerMensajeError(err, 'No se pudo eliminar la venta.'));
      },
    });
  }

  formatearMonto(valor: number): string {
    return (
      '$ ' +
      new Intl.NumberFormat('es-CO', {
        minimumFractionDigits: 2,
        maximumFractionDigits: 2,
      }).format(valor)
    );
  }

  formatearFecha(iso: string): string {
    return iso ? iso.slice(0, 10) : '—';
  }

  private crearLineaGroup(init?: {
    articulo_id?: number | null;
    descripcion?: string;
    cantidad?: number;
    precio_unitario?: number;
    impuesto?: number;
  }): FormGroup {
    return this.fb.group({
      articulo_id: [init?.articulo_id ?? null, [Validators.required]],
      descripcion: [init?.descripcion ?? '', [Validators.maxLength(255)]],
      cantidad: [init?.cantidad ?? 1, [Validators.required, Validators.min(0.0001)]],
      precio_unitario: [init?.precio_unitario ?? 0, [Validators.required, Validators.min(0)]],
      impuesto: [init?.impuesto ?? 0, [Validators.min(0)]],
    });
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
