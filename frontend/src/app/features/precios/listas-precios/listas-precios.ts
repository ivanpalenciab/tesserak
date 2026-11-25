import { Component, OnInit, computed, inject, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { HttpErrorResponse } from '@angular/common/http';
import { catchError, forkJoin, of } from 'rxjs';

import { ApiErrorBody } from '../../../core/auth/auth.models';
import { CrmService } from '../../crm/crm.service';
import { Moneda } from '../../crm/crm.models';
import { ProductosService } from '../../productos/productos.service';
import { Producto, UnidadMedida } from '../../productos/productos.models';
import { PreciosService } from '../precios.service';
import {
  ActualizarDetallePrecioRequest,
  ActualizarListaPrecioRequest,
  CrearDetallePrecioRequest,
  CrearListaPrecioRequest,
  DetallePrecio,
  ListaPrecio,
  TipoListaPrecio,
} from '../precios.models';

export const TIPOS_LISTA_OPCIONES: { value: TipoListaPrecio; label: string }[] = [
  { value: 'GENERAL', label: 'General' },
  { value: 'MAYORISTA', label: 'Mayorista' },
  { value: 'MINORISTA', label: 'Minorista' },
  { value: 'ESPECIAL', label: 'Especial' },
];

export interface ArticuloPrecioFila {
  key: string;
  articulo: Producto;
  detalle: DetallePrecio | null;
}

@Component({
  selector: 'app-listas-precios',
  standalone: true,
  imports: [ReactiveFormsModule],
  templateUrl: './listas-precios.html',
  styleUrl: './listas-precios.scss',
})
export class ListasPreciosComponent implements OnInit {
  private readonly preciosSvc = inject(PreciosService);
  private readonly productosSvc = inject(ProductosService);
  private readonly crmSvc = inject(CrmService);
  private readonly fb = inject(FormBuilder);

  readonly tiposLista = TIPOS_LISTA_OPCIONES;

  // Estado general y alertas
  readonly loadingListas = signal(true);
  readonly loadingDetalles = signal(false);
  readonly savingLista = signal(false);
  readonly savingDetalle = signal(false);
  readonly errorMensaje = signal<string | null>(null);
  readonly exitoMensaje = signal<string | null>(null);
  readonly modalListaError = signal<string | null>(null);
  readonly modalDetalleError = signal<string | null>(null);

  // Catálogos auxiliares
  readonly monedas = signal<Moneda[]>([]);
  readonly articulos = signal<Producto[]>([]);
  readonly unidades = signal<UnidadMedida[]>([]);

  // Cabecera: listas_precios
  readonly listas = signal<ListaPrecio[]>([]);
  readonly totalListas = signal(0);
  readonly pageListas = signal(1);
  readonly pageSizeListas = 10;

  readonly filtroTextoLista = signal('');
  readonly filtroTipoLista = signal<TipoListaPrecio | ''>('');
  readonly filtroMonedaId = signal<number | null>(null);
  readonly filtroEstadoLista = signal<'todas' | 'activas' | 'inactivas'>('todas');

  // Inicia en null: solo se despliega al dar clic en una lista de precios
  readonly listaSeleccionada = signal<ListaPrecio | null>(null);

  // Detalle: lista_precios_detalle de la lista seleccionada
  readonly detalles = signal<DetallePrecio[]>([]);
  readonly pageDetalles = signal(1);
  readonly pageSizeDetalles = 15;

  readonly filtroTextoDetalle = signal('');
  readonly filtroUnidadDetalleId = signal<number | null>(null);
  readonly filtroAsignacion = signal<'todos' | 'con_precio' | 'sin_precio'>('todos');

  // Modales
  readonly modalListaAbierto = signal(false);
  readonly editandoLista = signal<ListaPrecio | null>(null);
  readonly listaAEliminar = signal<ListaPrecio | null>(null);

  readonly modalDetalleAbierto = signal(false);
  readonly editandoDetalle = signal<DetallePrecio | null>(null);
  readonly detalleAEliminar = signal<DetallePrecio | null>(null);

  // Computed
  readonly totalPaginasListas = computed(() =>
    Math.max(1, Math.ceil(this.totalListas() / this.pageSizeListas)),
  );
  readonly monedasMap = computed(() => new Map(this.monedas().map((m) => [m.id, m])));
  readonly articulosMap = computed(() => new Map(this.articulos().map((a) => [a.id, a])));
  readonly unidadesMap = computed(() => new Map(this.unidades().map((u) => [u.id, u])));

  readonly hayFiltrosLista = computed(
    () =>
      this.filtroTextoLista().trim().length > 0 ||
      this.filtroTipoLista() !== '' ||
      this.filtroMonedaId() !== null ||
      this.filtroEstadoLista() !== 'todas',
  );

  readonly hayFiltrosDetalle = computed(
    () =>
      this.filtroTextoDetalle().trim().length > 0 ||
      this.filtroUnidadDetalleId() !== null ||
      this.filtroAsignacion() !== 'todos',
  );

  // Cruza todos los artículos activos de la empresa con los precios asignados en la lista seleccionada
  readonly filasArticulosPrecio = computed<ArticuloPrecioFila[]>(() => {
    const detallesPorArticulo = new Map<number, DetallePrecio[]>();
    for (const det of this.detalles()) {
      const grupo = detallesPorArticulo.get(det.articulo_id) ?? [];
      grupo.push(det);
      detallesPorArticulo.set(det.articulo_id, grupo);
    }

    const filas: ArticuloPrecioFila[] = [];
    for (const art of this.articulos()) {
      const asignados = detallesPorArticulo.get(art.id);
      if (asignados && asignados.length > 0) {
        for (const det of asignados) {
          filas.push({ key: `det-${det.id}`, articulo: art, detalle: det });
        }
      } else {
        filas.push({ key: `art-${art.id}`, articulo: art, detalle: null });
      }
    }
    return filas;
  });

  readonly resumenAsignacion = computed(() => {
    const idsConPrecio = new Set(this.detalles().map((d) => d.articulo_id));
    const totalActivos = this.articulos().length;
    let conPrecio = 0;
    for (const art of this.articulos()) {
      if (idsConPrecio.has(art.id)) conPrecio++;
    }
    return {
      totalActivos,
      conPrecio,
      sinPrecio: Math.max(0, totalActivos - conPrecio),
    };
  });

  readonly filasFiltradas = computed<ArticuloPrecioFila[]>(() => {
    const texto = this.filtroTextoDetalle().trim().toLowerCase();
    const unidadId = this.filtroUnidadDetalleId();
    const asignacion = this.filtroAsignacion();

    return this.filasArticulosPrecio().filter((fila) => {
      if (asignacion === 'con_precio' && !fila.detalle) return false;
      if (asignacion === 'sin_precio' && fila.detalle) return false;

      if (unidadId !== null) {
        const effUnidad = fila.detalle ? fila.detalle.unidad_medida_id : fila.articulo.unidad_medida_id;
        if (effUnidad !== unidadId) return false;
      }

      if (texto) {
        const matchCodigo = fila.articulo.codigo.toLowerCase().includes(texto);
        const matchNombre = fila.articulo.nombre.toLowerCase().includes(texto);
        if (!matchCodigo && !matchNombre) return false;
      }

      return true;
    });
  });

  readonly totalPaginasDetalles = computed(() =>
    Math.max(1, Math.ceil(this.filasFiltradas().length / this.pageSizeDetalles)),
  );

  readonly filasPaginadas = computed<ArticuloPrecioFila[]>(() => {
    const inicio = (this.pageDetalles() - 1) * this.pageSizeDetalles;
    return this.filasFiltradas().slice(inicio, inicio + this.pageSizeDetalles);
  });

  readonly formLista = this.fb.group({
    nombre: ['', [Validators.required, Validators.maxLength(150)]],
    moneda_id: [null as number | null, [Validators.required]],
    tipo_lista: ['GENERAL' as TipoListaPrecio, [Validators.required]],
    vigencia_desde: [this.hoyISO(), [Validators.required]],
    vigencia_hasta: [''],
    estado: [true],
  });

  readonly formDetalle = this.fb.group({
    articulo_id: [null as number | null, [Validators.required]],
    precio: [null as number | null, [Validators.required, Validators.min(0)]],
    unidad_medida_id: [null as number | null, [Validators.required]],
    cantidad_minima: [1 as number | null, [Validators.required, Validators.min(0.0001)]],
    vigencia_desde: [''],
    vigencia_hasta: [''],
  });

  ngOnInit(): void {
    this.cargarCatalogosYListas();
  }

  cargarCatalogosYListas(): void {
    this.loadingListas.set(true);
    this.errorMensaje.set(null);

    forkJoin({
      monedas: this.crmSvc.listarMonedas().pipe(catchError(() => of([] as Moneda[]))),
      unidades: this.productosSvc
        .listarUnidadesMedida()
        .pipe(catchError(() => of([] as UnidadMedida[]))),
      articulos: this.productosSvc
        .listar({ estado: true, page: 1, page_size: 100 })
        .pipe(catchError(() => of({ data: [] as Producto[], page: 1, page_size: 100, total: 0 }))),
    }).subscribe(({ monedas, unidades, articulos }) => {
      this.monedas.set(monedas);
      this.unidades.set(unidades);
      this.articulos.set(articulos.data ?? []);
      this.cargarListas();
    });
  }

  // =====================================================================
  // Operaciones sobre listas_precios (Cabecera)
  // =====================================================================

  cargarListas(): void {
    this.loadingListas.set(true);
    this.errorMensaje.set(null);

    const estado =
      this.filtroEstadoLista() === 'todas' ? null : this.filtroEstadoLista() === 'activas';

    this.preciosSvc
      .listarListas({
        tipo_lista: this.filtroTipoLista(),
        moneda_id: this.filtroMonedaId() ?? undefined,
        estado,
        q: this.filtroTextoLista(),
        page: this.pageListas(),
        page_size: this.pageSizeListas,
      })
      .subscribe({
        next: (res) => {
          const data = res.data ?? [];
          this.listas.set(data);
          this.totalListas.set(res.total ?? 0);
          this.loadingListas.set(false);

          // Solo mantiene actualizada la lista si el usuario ya había hecho clic en una
          const actual = this.listaSeleccionada();
          if (actual) {
            const actualizada = data.find((l) => l.id === actual.id);
            if (actualizada) {
              this.listaSeleccionada.set(actualizada);
            } else {
              this.cerrarDetalleLista();
            }
          }
        },
        error: (err) => {
          this.loadingListas.set(false);
          this.errorMensaje.set(
            this.extraerMensajeError(err, 'No se pudieron cargar las listas de precios.'),
          );
        },
      });
  }

  seleccionarLista(lista: ListaPrecio): void {
    this.listaSeleccionada.set(lista);
    this.pageDetalles.set(1);
    this.filtroTextoDetalle.set('');
    this.filtroUnidadDetalleId.set(null);
    this.filtroAsignacion.set('todos');
    this.cargarDetalles();
  }

  cerrarDetalleLista(): void {
    this.listaSeleccionada.set(null);
    this.detalles.set([]);
    this.pageDetalles.set(1);
  }

  onBuscarLista(valor: string): void {
    this.filtroTextoLista.set(valor);
    this.pageListas.set(1);
    this.cargarListas();
  }

  onFiltroTipoLista(valor: string): void {
    this.filtroTipoLista.set((valor || '') as TipoListaPrecio | '');
    this.pageListas.set(1);
    this.cargarListas();
  }

  onFiltroMoneda(valor: string): void {
    this.filtroMonedaId.set(valor ? Number(valor) : null);
    this.pageListas.set(1);
    this.cargarListas();
  }

  onFiltroEstadoLista(valor: string): void {
    this.filtroEstadoLista.set(valor as 'todas' | 'activas' | 'inactivas');
    this.pageListas.set(1);
    this.cargarListas();
  }

  limpiarFiltrosLista(): void {
    this.filtroTextoLista.set('');
    this.filtroTipoLista.set('');
    this.filtroMonedaId.set(null);
    this.filtroEstadoLista.set('todas');
    this.pageListas.set(1);
    this.cargarListas();
  }

  irPaginaListas(p: number): void {
    if (p < 1 || p > this.totalPaginasListas()) return;
    this.pageListas.set(p);
    this.cargarListas();
  }

  abrirNuevaLista(): void {
    this.editandoLista.set(null);
    this.modalListaError.set(null);
    const monedaDefault = this.monedas()[0]?.id ?? null;
    this.formLista.reset({
      nombre: '',
      moneda_id: monedaDefault,
      tipo_lista: 'GENERAL',
      vigencia_desde: this.hoyISO(),
      vigencia_hasta: '',
      estado: true,
    });
    this.modalListaAbierto.set(true);
  }

  abrirEditarLista(lista: ListaPrecio, event?: Event): void {
    event?.stopPropagation();
    this.editandoLista.set(lista);
    this.modalListaError.set(null);
    this.formLista.reset({
      nombre: lista.nombre,
      moneda_id: lista.moneda_id,
      tipo_lista: lista.tipo_lista,
      vigencia_desde: lista.vigencia_desde,
      vigencia_hasta: lista.vigencia_hasta ?? '',
      estado: lista.estado,
    });
    this.modalListaAbierto.set(true);
  }

  cerrarModalLista(): void {
    if (this.savingLista()) return;
    this.modalListaAbierto.set(false);
    this.editandoLista.set(null);
    this.modalListaError.set(null);
  }

  guardarLista(): void {
    this.modalListaError.set(null);
    if (this.formLista.invalid) {
      this.formLista.markAllAsTouched();
      return;
    }

    const v = this.formLista.getRawValue();
    const desde = (v.vigencia_desde ?? '').trim();
    const hasta = (v.vigencia_hasta ?? '').trim();
    if (hasta && hasta < desde) {
      this.modalListaError.set('La fecha final debe ser mayor o igual a la fecha inicial.');
      return;
    }

    this.savingLista.set(true);
    const actual = this.editandoLista();

    if (actual) {
      const payload: ActualizarListaPrecioRequest = {
        nombre: (v.nombre ?? '').trim(),
        moneda_id: Number(v.moneda_id),
        tipo_lista: v.tipo_lista as TipoListaPrecio,
        vigencia_desde: desde,
        estado: Boolean(v.estado),
      };
      if (hasta) {
        payload.vigencia_hasta = hasta;
      } else if (actual.vigencia_hasta) {
        payload.limpiar_vigencia_hasta = true;
      }

      this.preciosSvc.actualizarLista(actual.id, payload).subscribe({
        next: (actualizada) => {
          this.savingLista.set(false);
          this.modalListaAbierto.set(false);
          this.editandoLista.set(null);
          if (this.listaSeleccionada()?.id === actualizada.id) {
            this.listaSeleccionada.set(actualizada);
          }
          this.mostrarExito(`Lista de precios "${actualizada.nombre}" actualizada.`);
          this.cargarListas();
        },
        error: (err) => {
          this.savingLista.set(false);
          this.modalListaError.set(
            this.extraerMensajeError(err, 'No se pudo actualizar la lista de precios.'),
          );
        },
      });
    } else {
      const payload: CrearListaPrecioRequest = {
        nombre: (v.nombre ?? '').trim(),
        moneda_id: Number(v.moneda_id),
        tipo_lista: v.tipo_lista as TipoListaPrecio,
        vigencia_desde: desde,
        vigencia_hasta: hasta || null,
        estado: Boolean(v.estado),
      };

      this.preciosSvc.crearLista(payload).subscribe({
        next: (creada) => {
          this.savingLista.set(false);
          this.modalListaAbierto.set(false);
          this.mostrarExito(`Lista de precios "${creada.nombre}" creada correctamente.`);
          this.pageListas.set(1);
          this.cargarListas();
        },
        error: (err) => {
          this.savingLista.set(false);
          this.modalListaError.set(
            this.extraerMensajeError(err, 'No se pudo crear la lista de precios.'),
          );
        },
      });
    }
  }

  toggleEstadoLista(lista: ListaPrecio, event?: Event): void {
    event?.stopPropagation();
    const nuevoEstado = !lista.estado;
    this.preciosSvc.actualizarLista(lista.id, { estado: nuevoEstado }).subscribe({
      next: () => {
        this.mostrarExito(`Lista "${lista.nombre}" ${nuevoEstado ? 'activada' : 'desactivada'}.`);
        this.cargarListas();
      },
      error: (err) => {
        this.errorMensaje.set(
          this.extraerMensajeError(err, 'No se pudo cambiar el estado de la lista.'),
        );
      },
    });
  }

  solicitarEliminarLista(lista: ListaPrecio, event?: Event): void {
    event?.stopPropagation();
    this.listaAEliminar.set(lista);
  }

  cancelarEliminarLista(): void {
    if (this.savingLista()) return;
    this.listaAEliminar.set(null);
  }

  confirmarEliminarLista(): void {
    const target = this.listaAEliminar();
    if (!target) return;
    this.savingLista.set(true);
    this.preciosSvc.eliminarLista(target.id).subscribe({
      next: () => {
        this.savingLista.set(false);
        this.listaAEliminar.set(null);
        if (this.listaSeleccionada()?.id === target.id) {
          this.cerrarDetalleLista();
        }
        this.mostrarExito(`Lista "${target.nombre}" eliminada.`);
        this.cargarListas();
      },
      error: (err) => {
        this.savingLista.set(false);
        this.listaAEliminar.set(null);
        this.errorMensaje.set(
          this.extraerMensajeError(err, 'No se pudo eliminar la lista de precios.'),
        );
      },
    });
  }

  // =====================================================================
  // Operaciones sobre lista_precios_detalle (Detalle)
  // =====================================================================

  cargarDetalles(): void {
    const lista = this.listaSeleccionada();
    if (!lista) {
      this.detalles.set([]);
      return;
    }

    this.loadingDetalles.set(true);
    this.preciosSvc.listarDetalles(lista.id, { page: 1, page_size: 100 }).subscribe({
      next: (res) => {
        this.detalles.set(res.data ?? []);
        this.loadingDetalles.set(false);
      },
      error: (err) => {
        this.loadingDetalles.set(false);
        this.errorMensaje.set(
          this.extraerMensajeError(err, 'No se pudieron cargar los precios de la lista.'),
        );
      },
    });
  }

  onBuscarDetalle(valor: string): void {
    this.filtroTextoDetalle.set(valor);
    this.pageDetalles.set(1);
  }

  onFiltroUnidadDetalle(valor: string): void {
    this.filtroUnidadDetalleId.set(valor ? Number(valor) : null);
    this.pageDetalles.set(1);
  }

  onFiltroAsignacion(valor: string): void {
    this.filtroAsignacion.set(valor as 'todos' | 'con_precio' | 'sin_precio');
    this.pageDetalles.set(1);
  }

  limpiarFiltrosDetalle(): void {
    this.filtroTextoDetalle.set('');
    this.filtroUnidadDetalleId.set(null);
    this.filtroAsignacion.set('todos');
    this.pageDetalles.set(1);
  }

  irPaginaDetalles(p: number): void {
    if (p < 1 || p > this.totalPaginasDetalles()) return;
    this.pageDetalles.set(p);
  }

  abrirNuevoDetalle(articuloPreseleccionado?: Producto): void {
    if (!this.listaSeleccionada()) return;
    this.editandoDetalle.set(null);
    this.modalDetalleError.set(null);
    const unidadDefault =
      articuloPreseleccionado?.unidad_medida_id ?? this.unidades()[0]?.id ?? null;
    this.formDetalle.reset({
      articulo_id: articuloPreseleccionado?.id ?? null,
      precio: null,
      unidad_medida_id: unidadDefault,
      cantidad_minima: 1,
      vigencia_desde: '',
      vigencia_hasta: '',
    });
    this.modalDetalleAbierto.set(true);
  }

  abrirEditarDetalle(detalle: DetallePrecio): void {
    this.editandoDetalle.set(detalle);
    this.modalDetalleError.set(null);
    this.formDetalle.reset({
      articulo_id: detalle.articulo_id,
      precio: detalle.precio,
      unidad_medida_id: detalle.unidad_medida_id,
      cantidad_minima: detalle.cantidad_minima,
      vigencia_desde: detalle.vigencia_desde ?? '',
      vigencia_hasta: detalle.vigencia_hasta ?? '',
    });
    this.modalDetalleAbierto.set(true);
  }

  onSeleccionarArticulo(valor: string): void {
    const id = valor ? Number(valor) : null;
    if (!id) return;
    const art = this.articulosMap().get(id);
    if (art?.unidad_medida_id) {
      this.formDetalle.controls.unidad_medida_id.setValue(art.unidad_medida_id);
    }
  }

  cerrarModalDetalle(): void {
    if (this.savingDetalle()) return;
    this.modalDetalleAbierto.set(false);
    this.editandoDetalle.set(null);
    this.modalDetalleError.set(null);
  }

  guardarDetalle(): void {
    const lista = this.listaSeleccionada();
    if (!lista) return;

    this.modalDetalleError.set(null);
    if (this.formDetalle.invalid) {
      this.formDetalle.markAllAsTouched();
      return;
    }

    const v = this.formDetalle.getRawValue();
    const desde = (v.vigencia_desde ?? '').trim();
    const hasta = (v.vigencia_hasta ?? '').trim();
    if (desde && hasta && hasta < desde) {
      this.modalDetalleError.set('La fecha final debe ser mayor o igual a la fecha inicial.');
      return;
    }

    this.savingDetalle.set(true);
    const actual = this.editandoDetalle();

    if (actual) {
      const payload: ActualizarDetallePrecioRequest = {
        articulo_id: Number(v.articulo_id),
        precio: Number(v.precio),
        unidad_medida_id: Number(v.unidad_medida_id),
        cantidad_minima: Number(v.cantidad_minima),
      };
      if (desde) {
        payload.vigencia_desde = desde;
      } else if (actual.vigencia_desde) {
        payload.limpiar_vigencia_desde = true;
      }
      if (hasta) {
        payload.vigencia_hasta = hasta;
      } else if (actual.vigencia_hasta) {
        payload.limpiar_vigencia_hasta = true;
      }

      this.preciosSvc.actualizarDetalle(lista.id, actual.id, payload).subscribe({
        next: () => {
          this.savingDetalle.set(false);
          this.modalDetalleAbierto.set(false);
          this.editandoDetalle.set(null);
          this.mostrarExito('Precio actualizado correctamente.');
          this.cargarDetalles();
        },
        error: (err) => {
          this.savingDetalle.set(false);
          this.modalDetalleError.set(
            this.extraerMensajeError(err, 'No se pudo actualizar el precio.'),
          );
        },
      });
    } else {
      const payload: CrearDetallePrecioRequest = {
        articulo_id: Number(v.articulo_id),
        precio: Number(v.precio),
        unidad_medida_id: Number(v.unidad_medida_id),
        cantidad_minima: Number(v.cantidad_minima ?? 1),
        vigencia_desde: desde || null,
        vigencia_hasta: hasta || null,
      };

      this.preciosSvc.crearDetalle(lista.id, payload).subscribe({
        next: () => {
          this.savingDetalle.set(false);
          this.modalDetalleAbierto.set(false);
          this.mostrarExito('Precio asignado correctamente.');
          this.cargarDetalles();
        },
        error: (err) => {
          this.savingDetalle.set(false);
          this.modalDetalleError.set(
            this.extraerMensajeError(err, 'No se pudo asignar el precio.'),
          );
        },
      });
    }
  }

  solicitarEliminarDetalle(detalle: DetallePrecio): void {
    this.detalleAEliminar.set(detalle);
  }

  cancelarEliminarDetalle(): void {
    if (this.savingDetalle()) return;
    this.detalleAEliminar.set(null);
  }

  confirmarEliminarDetalle(): void {
    const lista = this.listaSeleccionada();
    const target = this.detalleAEliminar();
    if (!lista || !target) return;

    this.savingDetalle.set(true);
    this.preciosSvc.eliminarDetalle(lista.id, target.id).subscribe({
      next: () => {
        this.savingDetalle.set(false);
        this.detalleAEliminar.set(null);
        this.mostrarExito('Precio eliminado de la lista.');
        this.cargarDetalles();
      },
      error: (err) => {
        this.savingDetalle.set(false);
        this.detalleAEliminar.set(null);
        this.errorMensaje.set(this.extraerMensajeError(err, 'No se pudo eliminar el precio.'));
      },
    });
  }

  // =====================================================================
  // Helpers de presentación
  // =====================================================================

  etiquetaTipo(tipo: TipoListaPrecio): string {
    return this.tiposLista.find((t) => t.value === tipo)?.label ?? tipo;
  }

  etiquetaMoneda(lista: ListaPrecio): string {
    if (lista.moneda_codigo) {
      return `${lista.moneda_codigo} (${lista.moneda_simbolo ?? '$'})`;
    }
    const m = this.monedasMap().get(lista.moneda_id);
    return m ? `${m.codigo_iso} (${m.simbolo})` : `#${lista.moneda_id}`;
  }

  formatearPrecio(precio: number, lista: ListaPrecio): string {
    const simbolo = lista.moneda_simbolo ?? this.monedasMap().get(lista.moneda_id)?.simbolo ?? '$';
    const codigo = lista.moneda_codigo ?? this.monedasMap().get(lista.moneda_id)?.codigo_iso ?? '';
    const formatted = new Intl.NumberFormat('es-CO', {
      minimumFractionDigits: 2,
      maximumFractionDigits: 2,
    }).format(precio);
    return `${simbolo} ${formatted}${codigo ? ' ' + codigo : ''}`;
  }

  nombreUnidad(id: number, nombreActual?: string): string {
    if (nombreActual) return nombreActual;
    return this.unidadesMap().get(id)?.nombre ?? `#${id}`;
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
