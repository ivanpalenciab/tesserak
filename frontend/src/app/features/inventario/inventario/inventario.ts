import { Component, OnInit, computed, inject, signal } from '@angular/core';
import { FormArray, FormBuilder, FormGroup, ReactiveFormsModule, Validators } from '@angular/forms';
import { HttpErrorResponse } from '@angular/common/http';
import { catchError, forkJoin, of } from 'rxjs';

import { ApiErrorBody } from '../../../core/auth/auth.models';
import { ProductosService } from '../../productos/productos.service';
import { Producto } from '../../productos/productos.models';
import { TercerosService } from '../../terceros/terceros.service';
import { Tercero } from '../../terceros/terceros.models';
import { InventarioService } from '../inventario.service';
import {
  CrearMovimientoRequest,
  DocumentoInventario,
  MovimientoLineaInput,
  NaturalezaMovimiento,
  StockItem,
  TipoMovimientoInventario,
  Ubicacion,
} from '../inventario.models';

@Component({
  selector: 'app-inventario',
  standalone: true,
  imports: [ReactiveFormsModule],
  templateUrl: './inventario.html',
  styleUrl: './inventario.scss',
})
export class InventarioComponent implements OnInit {
  private readonly invSvc = inject(InventarioService);
  private readonly productosSvc = inject(ProductosService);
  private readonly tercerosSvc = inject(TercerosService);
  private readonly fb = inject(FormBuilder);

  readonly tabActiva = signal<'ubicaciones' | 'movimientos'>('ubicaciones');

  // Alertas
  readonly errorMensaje = signal<string | null>(null);
  readonly exitoMensaje = signal<string | null>(null);

  // Catálogos
  readonly loading = signal(true);
  readonly tiposMovimiento = signal<TipoMovimientoInventario[]>([]);
  readonly ubicaciones = signal<Ubicacion[]>([]);
  readonly articulos = signal<Producto[]>([]);
  readonly terceros = signal<Tercero[]>([]);

  // Ubicación seleccionada: inicia en null (solo se muestran las ubicaciones hasta hacer clic)
  readonly ubicacionSeleccionada = signal<Ubicacion | null>(null);

  // Stock de la ubicación seleccionada
  readonly loadingStock = signal(false);
  readonly stockItems = signal<StockItem[]>([]);
  readonly totalStock = signal(0);
  readonly pageStock = signal(1);
  readonly pageSize = 15;
  readonly filtroStockTexto = signal('');
  readonly filtroSoloConStock = signal(true);

  // Movimientos / Kardex
  readonly loadingMovs = signal(false);
  readonly movimientos = signal<DocumentoInventario[]>([]);
  readonly totalMovs = signal(0);
  readonly pageMovs = signal(1);
  readonly filtroMovTipoId = signal<number | null>(null);
  readonly filtroMovUbicacionId = signal<number | null>(null);
  readonly filtroMovTexto = signal('');
  readonly movDetalle = signal<DocumentoInventario | null>(null);
  readonly loadingMovDetalle = signal(false);

  // Modal Nuevo Movimiento / Ajuste / Traslado
  readonly modalMovAbierto = signal(false);
  readonly savingMov = signal(false);
  readonly modalMovError = signal<string | null>(null);
  readonly tipoMovSeleccionadoId = signal<number | null>(null);

  // Computed
  readonly tiposMap = computed(() => new Map(this.tiposMovimiento().map((t) => [t.id, t])));
  readonly ubicacionesActivas = computed(() => this.ubicaciones().filter((u) => u.activo));
  readonly totalPaginasStock = computed(() => Math.max(1, Math.ceil(this.totalStock() / this.pageSize)));
  readonly totalPaginasMovs = computed(() => Math.max(1, Math.ceil(this.totalMovs() / this.pageSize)));

  readonly tipoMovSeleccionado = computed(() => {
    const id = this.tipoMovSeleccionadoId();
    return id ? this.tiposMap().get(id) ?? null : null;
  });

  readonly formMov = this.fb.group({
    tipo_movimiento_id: [null as number | null, [Validators.required]],
    fecha: [this.hoyISO()],
    tercero_id: [null as number | null],
    motivo: ['', [Validators.required, Validators.minLength(2), Validators.maxLength(255)]],
    observaciones: ['', [Validators.maxLength(5000)]],
    detalles: this.fb.array<FormGroup>([]),
  });

  get detallesArray(): FormArray<FormGroup> {
    return this.formMov.controls.detalles;
  }

  ngOnInit(): void {
    this.cargarCatalogos();
  }

  cambiarTab(tab: 'ubicaciones' | 'movimientos'): void {
    this.tabActiva.set(tab);
    this.errorMensaje.set(null);
    if (tab === 'movimientos' && this.movimientos().length === 0) {
      this.cargarMovimientos();
    }
  }

  cargarCatalogos(): void {
    this.loading.set(true);
    this.errorMensaje.set(null);

    forkJoin({
      tipos: this.invSvc
        .listarTiposMovimiento()
        .pipe(catchError(() => of([] as TipoMovimientoInventario[]))),
      ubicaciones: this.invSvc
        .listarUbicaciones(true)
        .pipe(catchError(() => of([] as Ubicacion[]))),
      articulos: this.productosSvc
        .listar({ estado: true, page: 1, page_size: 100 })
        .pipe(catchError(() => of({ data: [] as Producto[], page: 1, page_size: 100, total: 0 }))),
      terceros: this.tercerosSvc
        .listar({ activo: true, page: 1, page_size: 100 })
        .pipe(catchError(() => of({ data: [] as Tercero[], page: 1, page_size: 100, total: 0 }))),
    }).subscribe(({ tipos, ubicaciones, articulos, terceros }) => {
      this.tiposMovimiento.set(tipos);
      this.ubicaciones.set(ubicaciones);
      this.articulos.set(articulos.data ?? []);
      this.terceros.set(terceros.data ?? []);
      this.loading.set(false);

      // Si el usuario ya tenía una ubicación abierta, refresca su stock
      const sel = this.ubicacionSeleccionada();
      if (sel) {
        const actualizada = ubicaciones.find((u) => u.id === sel.id);
        if (actualizada) {
          this.ubicacionSeleccionada.set(actualizada);
          this.cargarStockUbicacion();
        } else {
          this.cerrarUbicacion();
        }
      }
    });
  }

  // =====================================================================
  // Selección de ubicación y detalle de productos (inv_stock)
  // =====================================================================

  seleccionarUbicacion(ubic: Ubicacion): void {
    this.ubicacionSeleccionada.set(ubic);
    this.pageStock.set(1);
    this.filtroStockTexto.set('');
    this.cargarStockUbicacion();
  }

  cerrarUbicacion(): void {
    this.ubicacionSeleccionada.set(null);
    this.stockItems.set([]);
    this.totalStock.set(0);
  }

  cargarStockUbicacion(): void {
    const ubic = this.ubicacionSeleccionada();
    if (!ubic) return;

    this.loadingStock.set(true);
    this.errorMensaje.set(null);

    this.invSvc
      .listarStock({
        ubicacion_id: ubic.id,
        solo_con_stock: this.filtroSoloConStock() ? true : undefined,
        q: this.filtroStockTexto(),
        page: this.pageStock(),
        page_size: this.pageSize,
      })
      .subscribe({
        next: (res) => {
          this.stockItems.set(res.data ?? []);
          this.totalStock.set(res.total ?? 0);
          this.loadingStock.set(false);
        },
        error: (err) => {
          this.loadingStock.set(false);
          this.errorMensaje.set(
            this.extraerMensajeError(err, 'No se pudieron cargar los productos de esta ubicación.'),
          );
        },
      });
  }

  onBuscarStock(valor: string): void {
    this.filtroStockTexto.set(valor);
    this.pageStock.set(1);
    this.cargarStockUbicacion();
  }

  onToggleSoloConStock(checked: boolean): void {
    this.filtroSoloConStock.set(checked);
    this.pageStock.set(1);
    this.cargarStockUbicacion();
  }

  irPaginaStock(p: number): void {
    if (p < 1 || p > this.totalPaginasStock()) return;
    this.pageStock.set(p);
    this.cargarStockUbicacion();
  }

  // =====================================================================
  // Movimientos / Ajustes / Traslados
  // =====================================================================

  cargarMovimientos(): void {
    this.loadingMovs.set(true);
    this.invSvc
      .listarMovimientos({
        tipo_movimiento_id: this.filtroMovTipoId() ?? undefined,
        ubicacion_id: this.filtroMovUbicacionId() ?? undefined,
        q: this.filtroMovTexto(),
        page: this.pageMovs(),
        page_size: this.pageSize,
      })
      .subscribe({
        next: (res) => {
          this.movimientos.set(res.data ?? []);
          this.totalMovs.set(res.total ?? 0);
          this.loadingMovs.set(false);
        },
        error: (err) => {
          this.loadingMovs.set(false);
          this.errorMensaje.set(
            this.extraerMensajeError(err, 'No se pudieron cargar los movimientos de inventario.'),
          );
        },
      });
  }

  onFiltroMovTipo(valor: string): void {
    this.filtroMovTipoId.set(valor ? Number(valor) : null);
    this.pageMovs.set(1);
    this.cargarMovimientos();
  }

  onFiltroMovUbicacion(valor: string): void {
    this.filtroMovUbicacionId.set(valor ? Number(valor) : null);
    this.pageMovs.set(1);
    this.cargarMovimientos();
  }

  onBuscarMov(valor: string): void {
    this.filtroMovTexto.set(valor);
    this.pageMovs.set(1);
    this.cargarMovimientos();
  }

  verDetalleMovimiento(doc: DocumentoInventario): void {
    this.movDetalle.set(doc);
    this.loadingMovDetalle.set(true);
    this.invSvc.obtenerMovimiento(doc.id).subscribe({
      next: (completo) => {
        this.movDetalle.set(completo);
        this.loadingMovDetalle.set(false);
      },
      error: (err) => {
        this.loadingMovDetalle.set(false);
        this.errorMensaje.set(
          this.extraerMensajeError(err, 'No se pudo cargar el detalle del movimiento.'),
        );
      },
    });
  }

  cerrarDetalleMovimiento(): void {
    this.movDetalle.set(null);
  }

  abrirTraslado(pre?: { articulo_id?: number; ubicacion_origen_id?: number }): void {
    const tipoTraslado = this.tiposMovimiento().find((t) => t.codigo === 'TRASLADO');
    this.abrirNuevoMovimiento({
      tipo_movimiento_id: tipoTraslado?.id,
      articulo_id: pre?.articulo_id,
      ubicacion_origen_id: pre?.ubicacion_origen_id ?? this.ubicacionSeleccionada()?.id,
      motivo: 'Traslado interno entre ubicaciones',
    });
  }

  abrirNuevoMovimiento(pre?: {
    tipo_movimiento_id?: number;
    articulo_id?: number;
    ubicacion_origen_id?: number;
    ubicacion_destino_id?: number;
    motivo?: string;
  }): void {
    this.modalMovError.set(null);
    this.detallesArray.clear();

    const defaultTipo = pre?.tipo_movimiento_id ?? this.tiposMovimiento()[0]?.id ?? null;
    this.tipoMovSeleccionadoId.set(defaultTipo);

    const activas = this.ubicacionesActivas();
    const origenDefault =
      pre?.ubicacion_origen_id ?? this.ubicacionSeleccionada()?.id ?? activas[0]?.id ?? null;
    const destinoDefault =
      pre?.ubicacion_destino_id ??
      (defaultTipo && this.tiposMap().get(defaultTipo)?.naturaleza === 'TRASLADO'
        ? (activas.find((u) => u.id !== origenDefault)?.id ?? null)
        : origenDefault);

    this.formMov.reset({
      tipo_movimiento_id: defaultTipo,
      fecha: this.hoyISO(),
      tercero_id: null,
      motivo: pre?.motivo ?? '',
      observaciones: '',
    });

    this.detallesArray.push(
      this.crearLineaMovGroup({
        articulo_id: pre?.articulo_id ?? null,
        ubicacion_origen_id: origenDefault,
        ubicacion_destino_id: destinoDefault,
        cantidad: 1,
        costo_unitario: 0,
      }),
    );
    this.modalMovAbierto.set(true);
  }

  onCambiarTipoMovimiento(): void {
    const raw = this.formMov.controls.tipo_movimiento_id.value;
    const id = raw !== null && raw !== undefined && !Number.isNaN(Number(raw)) ? Number(raw) : null;
    this.tipoMovSeleccionadoId.set(id);

    const tipo = id ? (this.tiposMap().get(id) ?? null) : null;
    const activas = this.ubicacionesActivas();
    if (!tipo || activas.length === 0) return;

    for (const g of this.detallesArray.controls) {
      const origen = g.get('ubicacion_origen_id')?.value;
      const destino = g.get('ubicacion_destino_id')?.value;
      const fallback = this.ubicacionSeleccionada()?.id ?? activas[0].id;

      if (tipo.naturaleza === 'TRASLADO') {
        const origenFinal = origen ?? fallback;
        g.get('ubicacion_origen_id')?.setValue(origenFinal);
        if (!destino || destino === origenFinal) {
          const otra = activas.find((u) => u.id !== origenFinal);
          g.get('ubicacion_destino_id')?.setValue(otra?.id ?? null);
        }
      } else if (tipo.naturaleza === 'ENTRADA' && !destino) {
        g.get('ubicacion_destino_id')?.setValue(fallback);
      } else if (tipo.naturaleza === 'SALIDA' && !origen) {
        g.get('ubicacion_origen_id')?.setValue(fallback);
      }
    }
  }

  agregarLineaMov(): void {
    const activas = this.ubicacionesActivas();
    const origenDefault = this.ubicacionSeleccionada()?.id ?? activas[0]?.id ?? null;
    const destinoDefault =
      this.tipoMovSeleccionado()?.naturaleza === 'TRASLADO'
        ? (activas.find((u) => u.id !== origenDefault)?.id ?? null)
        : origenDefault;

    this.detallesArray.push(
      this.crearLineaMovGroup({
        ubicacion_origen_id: origenDefault,
        ubicacion_destino_id: destinoDefault,
        cantidad: 1,
        costo_unitario: 0,
      }),
    );
  }

  eliminarLineaMov(index: number): void {
    if (this.detallesArray.length <= 1) return;
    this.detallesArray.removeAt(index);
  }

  cerrarModalMov(): void {
    if (this.savingMov()) return;
    this.modalMovAbierto.set(false);
    this.modalMovError.set(null);
  }

  guardarMovimiento(): void {
    this.modalMovError.set(null);
    if (this.formMov.invalid || this.detallesArray.length === 0) {
      this.formMov.markAllAsTouched();
      return;
    }

    const v = this.formMov.getRawValue();
    const tipoId =
      v.tipo_movimiento_id !== null && v.tipo_movimiento_id !== undefined
        ? Number(v.tipo_movimiento_id)
        : this.tipoMovSeleccionadoId();
    const tipo = tipoId && !Number.isNaN(tipoId) ? (this.tiposMap().get(tipoId) ?? null) : null;
    if (!tipo) {
      this.modalMovError.set('Selecciona un tipo de movimiento.');
      return;
    }

    const detalles: MovimientoLineaInput[] = [];

    for (const g of this.detallesArray.controls) {
      const articuloId = Number(g.get('articulo_id')?.value);
      const origenId = g.get('ubicacion_origen_id')?.value
        ? Number(g.get('ubicacion_origen_id')?.value)
        : null;
      const destinoId = g.get('ubicacion_destino_id')?.value
        ? Number(g.get('ubicacion_destino_id')?.value)
        : null;
      const cantidad = Number(g.get('cantidad')?.value);
      const costo = Number(g.get('costo_unitario')?.value ?? 0);

      if (tipo.naturaleza === 'ENTRADA' && !destinoId) {
        this.modalMovError.set('Selecciona la ubicación de destino para todas las líneas.');
        return;
      }
      if (tipo.naturaleza === 'SALIDA' && !origenId) {
        this.modalMovError.set('Selecciona la ubicación de origen para todas las líneas.');
        return;
      }
      if (tipo.naturaleza === 'TRASLADO') {
        if (!origenId || !destinoId) {
          this.modalMovError.set('Selecciona ubicación de origen y destino en cada traslado.');
          return;
        }
        if (origenId === destinoId) {
          this.modalMovError.set('En un traslado la ubicación de origen y destino deben ser distintas.');
          return;
        }
      }

      detalles.push({
        articulo_id: articuloId,
        ubicacion_origen_id: tipo.naturaleza === 'ENTRADA' ? null : origenId,
        ubicacion_destino_id: tipo.naturaleza === 'SALIDA' ? null : destinoId,
        cantidad,
        costo_unitario: costo,
      });
    }

    const payload: CrearMovimientoRequest = {
      tipo_movimiento_id: tipo.id,
      fecha: (v.fecha ?? '').trim() || undefined,
      tercero_id: v.tercero_id ? Number(v.tercero_id) : null,
      motivo: (v.motivo ?? '').trim(),
      observaciones: (v.observaciones ?? '').trim() || null,
      detalles,
    };

    this.savingMov.set(true);
    this.invSvc.crearMovimiento(payload).subscribe({
      next: (doc) => {
        this.savingMov.set(false);
        this.modalMovAbierto.set(false);
        this.mostrarExito(
          `Movimiento ${doc.consecutivo} (${tipo.nombre}) aplicado correctamente al inventario.`,
        );
        if (this.ubicacionSeleccionada()) {
          this.cargarStockUbicacion();
        }
        if (this.tabActiva() === 'movimientos' || this.movimientos().length > 0) {
          this.cargarMovimientos();
        }
      },
      error: (err) => {
        this.savingMov.set(false);
        this.modalMovError.set(
          this.extraerMensajeError(err, 'No se pudo registrar el movimiento de inventario.'),
        );
      },
    });
  }

  // =====================================================================
  // Helpers
  // =====================================================================

  etiquetaNaturaleza(nat?: NaturalezaMovimiento): string {
    if (nat === 'ENTRADA') return '+ Entrada';
    if (nat === 'SALIDA') return '− Salida';
    if (nat === 'TRASLADO') return '⇄ Traslado';
    return '';
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

  private crearLineaMovGroup(init?: {
    articulo_id?: number | null;
    ubicacion_origen_id?: number | null;
    ubicacion_destino_id?: number | null;
    cantidad?: number;
    costo_unitario?: number;
  }): FormGroup {
    return this.fb.group({
      articulo_id: [init?.articulo_id ?? null, [Validators.required]],
      ubicacion_origen_id: [init?.ubicacion_origen_id ?? null],
      ubicacion_destino_id: [init?.ubicacion_destino_id ?? null],
      cantidad: [init?.cantidad ?? 1, [Validators.required, Validators.min(0.0001)]],
      costo_unitario: [init?.costo_unitario ?? 0, [Validators.min(0)]],
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
