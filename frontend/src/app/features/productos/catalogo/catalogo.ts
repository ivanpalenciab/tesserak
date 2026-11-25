import { Component, OnInit, computed, inject, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { HttpErrorResponse } from '@angular/common/http';
import { catchError, forkJoin, of } from 'rxjs';

import { ApiErrorBody } from '../../../core/auth/auth.models';
import { ProductosService } from '../productos.service';
import {
  ActualizarProductoRequest,
  CategoriaProducto,
  CrearProductoRequest,
  GrupoImpuesto,
  MarcaProducto,
  MetodoValoracion,
  ModalidadServicio,
  Producto,
  TipoArticulo,
  TipoArticuloCodigo,
  UnidadMedida,
} from '../productos.models';

@Component({
  selector: 'app-productos-catalogo',
  standalone: true,
  imports: [ReactiveFormsModule],
  templateUrl: './catalogo.html',
  styleUrl: './catalogo.scss',
})
export class ProductosCatalogoComponent implements OnInit {
  private readonly productosSvc = inject(ProductosService);
  private readonly fb = inject(FormBuilder);

  readonly loading = signal(true);
  readonly saving = signal(false);
  readonly errorMensaje = signal<string | null>(null);
  readonly modalError = signal<string | null>(null);
  readonly exitoMensaje = signal<string | null>(null);
  readonly tipoSeleccionadoId = signal<number | null>(null);

  readonly productos = signal<Producto[]>([]);
  readonly total = signal(0);
  readonly page = signal(1);
  readonly pageSize = 15;

  readonly tiposArticulo = signal<TipoArticulo[]>([]);
  readonly metodosValoracion = signal<MetodoValoracion[]>([]);
  readonly modalidadesServicio = signal<ModalidadServicio[]>([]);
  readonly gruposImpuesto = signal<GrupoImpuesto[]>([]);
  readonly unidades = signal<UnidadMedida[]>([]);
  readonly categorias = signal<CategoriaProducto[]>([]);
  readonly marcas = signal<MarcaProducto[]>([]);

  readonly filtroTexto = signal('');
  readonly filtroTipoId = signal<number | null>(null);
  readonly filtroEstado = signal<'todos' | 'activos' | 'inactivos'>('todos');
  readonly filtroCategoriaId = signal<number | null>(null);

  readonly modalAbierto = signal(false);
  readonly editando = signal<Producto | null>(null);
  readonly aEliminar = signal<Producto | null>(null);

  readonly totalPaginas = computed(() => Math.max(1, Math.ceil(this.total() / this.pageSize)));
  readonly tiposMap = computed(() => new Map(this.tiposArticulo().map((t) => [t.id, t])));
  readonly unidadesMap = computed(() => new Map(this.unidades().map((u) => [u.id, u])));
  readonly categoriasMap = computed(() => new Map(this.categorias().map((c) => [c.id, c])));
  readonly marcasMap = computed(() => new Map(this.marcas().map((m) => [m.id, m])));
  readonly impuestosMap = computed(() => new Map(this.gruposImpuesto().map((g) => [g.id, g])));

  readonly hayFiltrosActivos = computed(
    () =>
      this.filtroTexto().trim().length > 0 ||
      this.filtroTipoId() !== null ||
      this.filtroEstado() !== 'todos' ||
      this.filtroCategoriaId() !== null,
  );

  readonly form = this.fb.group({
    // Comunes (articulos)
    codigo: ['', [Validators.required, Validators.maxLength(80)]],
    nombre: ['', [Validators.required, Validators.maxLength(150)]],
    tipo_articulo_id: [null as number | null, [Validators.required]],
    unidad_medida_id: [null as number | null, [Validators.required]],
    categoria_id: [null as number | null, [Validators.required]],
    grupo_impuesto_id: [null as number | null, [Validators.required]],
    marca_id: [null as number | null],
    descripcion: ['', [Validators.maxLength(5000)]],
    estado: [true],

    // Extensión producto
    control_inventario: [true],
    metodo_valoracion_id: [null as number | null],
    permite_compra: [true],
    permite_venta: [true],

    // Extensión servicio
    modalidad_id: [null as number | null],
    duracion_estimada: [null as number | null, [Validators.min(0.01)]],
    requiere_programacion: [false],
    control_consumos: [false],
  });

  ngOnInit(): void {
    this.cargarCatalogosYProductos();
  }

  cargarCatalogosYProductos(): void {
    this.loading.set(true);
    this.errorMensaje.set(null);

    forkJoin({
      tipos: this.productosSvc
        .listarTiposArticulo()
        .pipe(catchError(() => of([] as TipoArticulo[]))),
      metodos: this.productosSvc
        .listarMetodosValoracion()
        .pipe(catchError(() => of([] as MetodoValoracion[]))),
      modalidades: this.productosSvc
        .listarModalidadesServicio()
        .pipe(catchError(() => of([] as ModalidadServicio[]))),
      impuestos: this.productosSvc
        .listarGruposImpuesto(true)
        .pipe(catchError(() => of([] as GrupoImpuesto[]))),
      unidades: this.productosSvc
        .listarUnidadesMedida()
        .pipe(catchError(() => of([] as UnidadMedida[]))),
      categorias: this.productosSvc
        .listarCategorias(true)
        .pipe(catchError(() => of([] as CategoriaProducto[]))),
      marcas: this.productosSvc
        .listarMarcas(true)
        .pipe(catchError(() => of([] as MarcaProducto[]))),
    }).subscribe(({ tipos, metodos, modalidades, impuestos, unidades, categorias, marcas }) => {
      this.tiposArticulo.set(tipos);
      this.metodosValoracion.set(metodos);
      this.modalidadesServicio.set(modalidades);
      this.gruposImpuesto.set(impuestos);
      this.unidades.set(unidades);
      this.categorias.set(categorias);
      this.marcas.set(marcas);
      this.cargarProductos();
    });
  }

  cargarProductos(): void {
    this.loading.set(true);
    this.errorMensaje.set(null);

    const estado =
      this.filtroEstado() === 'todos' ? null : this.filtroEstado() === 'activos';

    this.productosSvc
      .listar({
        tipo_articulo_id: this.filtroTipoId() ?? undefined,
        estado,
        categoria_id: this.filtroCategoriaId() ?? undefined,
        q: this.filtroTexto(),
        page: this.page(),
        page_size: this.pageSize,
      })
      .subscribe({
        next: (res) => {
          this.productos.set(res.data ?? []);
          this.total.set(res.total ?? 0);
          this.loading.set(false);
        },
        error: (err: HttpErrorResponse) => {
          this.productos.set([]);
          this.total.set(0);
          this.errorMensaje.set(this.extraerError(err, 'No se pudieron cargar los artículos.'));
          this.loading.set(false);
        },
      });
  }

  onBuscarTexto(valor: string): void {
    this.filtroTexto.set(valor);
    this.page.set(1);
    this.cargarProductos();
  }

  onFiltroTipo(valor: string): void {
    this.filtroTipoId.set(valor ? Number(valor) : null);
    this.page.set(1);
    this.cargarProductos();
  }

  onFiltroEstado(valor: string): void {
    this.filtroEstado.set(valor as 'todos' | 'activos' | 'inactivos');
    this.page.set(1);
    this.cargarProductos();
  }

  onFiltroCategoria(valor: string): void {
    this.filtroCategoriaId.set(valor ? Number(valor) : null);
    this.page.set(1);
    this.cargarProductos();
  }

  limpiarFiltros(): void {
    this.filtroTexto.set('');
    this.filtroTipoId.set(null);
    this.filtroEstado.set('todos');
    this.filtroCategoriaId.set(null);
    this.page.set(1);
    this.cargarProductos();
  }

  irAPagina(nueva: number): void {
    if (nueva < 1 || nueva > this.totalPaginas() || nueva === this.page()) return;
    this.page.set(nueva);
    this.cargarProductos();
  }

  tipoSeleccionadoCodigo(): TipoArticuloCodigo | null {
    const id = this.form.controls.tipo_articulo_id.value ?? this.tipoSeleccionadoId();
    if (!id) return null;
    return this.tiposMap().get(Number(id))?.codigo ?? null;
  }

  seleccionarTipo(id: number): void {
    if (this.editando()) return;
    this.modalError.set(null);
    this.tipoSeleccionadoId.set(id);
    this.form.controls.tipo_articulo_id.setValue(id);
    this.form.controls.tipo_articulo_id.markAsTouched();
  }

  abrirNuevo(): void {
    this.editando.set(null);
    this.errorMensaje.set(null);
    this.modalError.set(null);
    this.tipoSeleccionadoId.set(null);
    this.form.controls.tipo_articulo_id.enable();
    this.form.reset({
      codigo: '',
      nombre: '',
      tipo_articulo_id: null,
      unidad_medida_id: this.unidades()[0]?.id ?? null,
      categoria_id: this.categorias()[0]?.id ?? null,
      grupo_impuesto_id: this.gruposImpuesto()[0]?.id ?? null,
      marca_id: null,
      descripcion: '',
      estado: true,
      control_inventario: true,
      metodo_valoracion_id: this.metodosValoracion()[0]?.id ?? null,
      permite_compra: true,
      permite_venta: true,
      modalidad_id: this.modalidadesServicio()[0]?.id ?? null,
      duracion_estimada: null,
      requiere_programacion: false,
      control_consumos: false,
    });
    this.modalAbierto.set(true);
  }

  abrirEditar(articulo: Producto): void {
    this.editando.set(articulo);
    this.errorMensaje.set(null);
    this.modalError.set(null);
    this.tipoSeleccionadoId.set(articulo.tipo_articulo_id);
    this.form.reset({
      codigo: articulo.codigo,
      nombre: articulo.nombre,
      tipo_articulo_id: articulo.tipo_articulo_id,
      unidad_medida_id: articulo.unidad_medida_id ?? this.unidades()[0]?.id ?? null,
      categoria_id: articulo.categoria_id ?? this.categorias()[0]?.id ?? null,
      grupo_impuesto_id: articulo.grupo_impuesto_id ?? this.gruposImpuesto()[0]?.id ?? null,
      marca_id: articulo.marca_id,
      descripcion: articulo.descripcion ?? '',
      estado: articulo.estado,
      control_inventario: articulo.producto?.control_inventario ?? true,
      metodo_valoracion_id:
        articulo.producto?.metodo_valoracion_id ?? this.metodosValoracion()[0]?.id ?? null,
      permite_compra: articulo.producto?.permite_compra ?? true,
      permite_venta: articulo.producto?.permite_venta ?? true,
      modalidad_id: articulo.servicio?.modalidad_id ?? this.modalidadesServicio()[0]?.id ?? null,
      duracion_estimada: articulo.servicio?.duracion_estimada ?? null,
      requiere_programacion: articulo.servicio?.requiere_programacion ?? false,
      control_consumos: articulo.servicio?.control_consumos ?? false,
    });
    this.form.controls.tipo_articulo_id.disable();
    this.modalAbierto.set(true);
  }

  cerrarModal(): void {
    if (this.saving()) return;
    this.modalAbierto.set(false);
    this.editando.set(null);
    this.modalError.set(null);
    this.tipoSeleccionadoId.set(null);
  }

  guardar(): void {
    const tipoCodigo = this.tipoSeleccionadoCodigo();
    if (!tipoCodigo) {
      this.modalError.set('Primero selecciona si deseas crear un Producto o un Servicio.');
      return;
    }

    if (this.form.invalid) {
      this.form.markAllAsTouched();
      this.modalError.set(
        'Completa los campos obligatorios: Código, Nombre, Unidad, Categoría y Grupo de impuestos.',
      );
      return;
    }

    const raw = this.form.getRawValue();

    if (tipoCodigo === 'PRODUCTO' && raw.control_inventario && !raw.metodo_valoracion_id) {
      this.modalError.set(
        'Cuando el producto controla inventario debes seleccionar un método de valoración.',
      );
      return;
    }
    if (tipoCodigo === 'SERVICIO' && !raw.modalidad_id) {
      this.modalError.set('Para un servicio debes seleccionar una modalidad.');
      return;
    }

    const editando = this.editando();
    this.saving.set(true);
    this.errorMensaje.set(null);
    this.modalError.set(null);

    if (editando) {
      const payload: ActualizarProductoRequest = {
        codigo: raw.codigo!.trim(),
        nombre: raw.nombre!.trim(),
        unidad_medida_id: Number(raw.unidad_medida_id),
        categoria_id: Number(raw.categoria_id),
        grupo_impuesto_id: Number(raw.grupo_impuesto_id),
        marca_id: raw.marca_id ? Number(raw.marca_id) : undefined,
        limpiar_marca: !raw.marca_id,
        descripcion: this.normalizarOpcional(raw.descripcion),
        estado: !!raw.estado,
      };
      if (tipoCodigo === 'PRODUCTO') {
        payload.producto = {
          control_inventario: !!raw.control_inventario,
          metodo_valoracion_id: raw.metodo_valoracion_id ? Number(raw.metodo_valoracion_id) : null,
          permite_compra: !!raw.permite_compra,
          permite_venta: !!raw.permite_venta,
        };
      } else if (tipoCodigo === 'SERVICIO') {
        payload.servicio = {
          modalidad_id: Number(raw.modalidad_id),
          duracion_estimada:
            raw.duracion_estimada !== null && raw.duracion_estimada !== undefined
              ? Number(raw.duracion_estimada)
              : null,
          requiere_programacion: !!raw.requiere_programacion,
          control_consumos: !!raw.control_consumos,
        };
      }

      this.productosSvc.actualizar(editando.id, payload).subscribe({
        next: () => {
          this.saving.set(false);
          this.modalAbierto.set(false);
          this.editando.set(null);
          this.mostrarExito('Artículo actualizado correctamente.');
          this.cargarProductos();
        },
        error: (err: HttpErrorResponse) => {
          this.saving.set(false);
          this.modalError.set(this.extraerError(err, 'No se pudo actualizar el artículo.'));
        },
      });
      return;
    }

    const payload: CrearProductoRequest = {
      codigo: raw.codigo!.trim(),
      nombre: raw.nombre!.trim(),
      tipo_articulo_id: Number(raw.tipo_articulo_id),
      unidad_medida_id: Number(raw.unidad_medida_id),
      categoria_id: Number(raw.categoria_id),
      grupo_impuesto_id: Number(raw.grupo_impuesto_id),
      marca_id: raw.marca_id ? Number(raw.marca_id) : null,
      descripcion: this.normalizarOpcional(raw.descripcion),
      estado: !!raw.estado,
    };
    if (tipoCodigo === 'PRODUCTO') {
      payload.producto = {
        control_inventario: !!raw.control_inventario,
        metodo_valoracion_id: raw.metodo_valoracion_id ? Number(raw.metodo_valoracion_id) : null,
        permite_compra: !!raw.permite_compra,
        permite_venta: !!raw.permite_venta,
      };
    } else if (tipoCodigo === 'SERVICIO') {
      payload.servicio = {
        modalidad_id: Number(raw.modalidad_id),
        duracion_estimada:
          raw.duracion_estimada !== null && raw.duracion_estimada !== undefined
            ? Number(raw.duracion_estimada)
            : null,
        requiere_programacion: !!raw.requiere_programacion,
        control_consumos: !!raw.control_consumos,
      };
    }

    this.productosSvc.crear(payload).subscribe({
      next: () => {
        this.saving.set(false);
        this.modalAbierto.set(false);
        this.mostrarExito('Artículo creado correctamente.');
        this.page.set(1);
        this.cargarProductos();
      },
      error: (err: HttpErrorResponse) => {
        this.saving.set(false);
        this.modalError.set(this.extraerError(err, 'No se pudo crear el artículo.'));
      },
    });
  }

  toggleActivo(articulo: Producto): void {
    this.errorMensaje.set(null);
    const nuevoEstado = !articulo.estado;
    this.productosSvc.actualizar(articulo.id, { estado: nuevoEstado }).subscribe({
      next: (actualizado) => {
        this.productos.update((lista) =>
          lista.map((p) => (p.id === actualizado.id ? actualizado : p)),
        );
        this.mostrarExito(
          `"${actualizado.nombre}" ${nuevoEstado ? 'activado' : 'desactivado'}.`,
        );
      },
      error: (err: HttpErrorResponse) => {
        this.errorMensaje.set(
          this.extraerError(err, 'No se pudo cambiar el estado del artículo.'),
        );
      },
    });
  }

  solicitarEliminar(articulo: Producto): void {
    this.aEliminar.set(articulo);
  }

  cancelarEliminar(): void {
    if (this.saving()) return;
    this.aEliminar.set(null);
  }

  confirmarEliminar(): void {
    const articulo = this.aEliminar();
    if (!articulo) return;

    this.saving.set(true);
    this.productosSvc.eliminar(articulo.id).subscribe({
      next: () => {
        this.saving.set(false);
        this.aEliminar.set(null);
        this.mostrarExito(`"${articulo.nombre}" eliminado.`);
        this.cargarProductos();
      },
      error: (err: HttpErrorResponse) => {
        this.saving.set(false);
        this.aEliminar.set(null);
        this.errorMensaje.set(this.extraerError(err, 'No se pudo eliminar el artículo.'));
      },
    });
  }

  nombreTipo(articulo: Producto): string {
    return this.tiposMap().get(articulo.tipo_articulo_id)?.nombre ?? articulo.tipo_articulo_codigo ?? '—';
  }

  codigoTipo(articulo: Producto): string {
    return articulo.tipo_articulo_codigo ?? this.tiposMap().get(articulo.tipo_articulo_id)?.codigo ?? '';
  }

  nombreUnidad(id: number | null): string {
    if (!id) return '—';
    return this.unidadesMap().get(id)?.nombre ?? `Unidad #${id}`;
  }

  nombreCategoria(id: number | null): string {
    if (!id) return '—';
    return this.categoriasMap().get(id)?.nombre ?? `Cat. #${id}`;
  }

  nombreMarca(id: number | null): string {
    if (!id) return '—';
    return this.marcasMap().get(id)?.nombre ?? `Marca #${id}`;
  }

  nombreImpuesto(id: number | null): string {
    if (!id) return '—';
    const g = this.impuestosMap().get(id);
    return g ? `${g.nombre} (${g.tarifa}%)` : `Imp. #${id}`;
  }

  private normalizarOpcional(valor: string | null | undefined): string | undefined {
    const limpio = valor?.trim();
    return limpio ? limpio : undefined;
  }

  private mostrarExito(mensaje: string): void {
    this.exitoMensaje.set(mensaje);
    setTimeout(() => {
      if (this.exitoMensaje() === mensaje) {
        this.exitoMensaje.set(null);
      }
    }, 3500);
  }

  private extraerError(err: HttpErrorResponse, fallback: string): string {
    const body = err.error as ApiErrorBody | null;
    if (body?.error?.message) return body.error.message;
    if (err.status === 0) {
      return 'No se pudo conectar con el servidor backend. Verifica que la API esté en ejecución.';
    }
    return fallback;
  }
}
