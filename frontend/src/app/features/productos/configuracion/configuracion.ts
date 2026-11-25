import { Component, OnInit, computed, inject, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { HttpErrorResponse } from '@angular/common/http';
import { catchError, forkJoin, of } from 'rxjs';

import { ApiErrorBody } from '../../../core/auth/auth.models';
import { ProductosService } from '../productos.service';
import { GrupoImpuesto, MarcaProducto } from '../productos.models';

type ConfigTab = 'marcas' | 'impuestos';
type FiltroEstado = 'todos' | 'activos' | 'inactivos';

interface ItemAEliminar {
  tipo: ConfigTab;
  id: number;
  nombre: string;
}

@Component({
  selector: 'app-productos-configuracion',
  standalone: true,
  imports: [ReactiveFormsModule],
  templateUrl: './configuracion.html',
  styleUrl: './configuracion.scss',
})
export class ProductosConfiguracionComponent implements OnInit {
  private readonly productosSvc = inject(ProductosService);
  private readonly fb = inject(FormBuilder);

  readonly tabActiva = signal<ConfigTab>('marcas');
  readonly loading = signal(true);
  readonly saving = signal(false);
  readonly errorMensaje = signal<string | null>(null);
  readonly modalError = signal<string | null>(null);
  readonly exitoMensaje = signal<string | null>(null);

  // Marcas
  readonly marcas = signal<MarcaProducto[]>([]);
  readonly filtroTextoMarca = signal('');
  readonly filtroEstadoMarca = signal<FiltroEstado>('todos');
  readonly modalMarcaAbierto = signal(false);
  readonly editandoMarca = signal<MarcaProducto | null>(null);

  // Grupos de impuestos
  readonly impuestos = signal<GrupoImpuesto[]>([]);
  readonly filtroTextoImpuesto = signal('');
  readonly filtroEstadoImpuesto = signal<FiltroEstado>('todos');
  readonly modalImpuestoAbierto = signal(false);
  readonly editandoImpuesto = signal<GrupoImpuesto | null>(null);

  // Confirmación soft delete
  readonly aEliminar = signal<ItemAEliminar | null>(null);

  readonly marcasFiltradas = computed(() => {
    const q = this.filtroTextoMarca().trim().toLowerCase();
    const estado = this.filtroEstadoMarca();
    return this.marcas().filter((m) => {
      if (estado === 'activos' && !m.activo) return false;
      if (estado === 'inactivos' && m.activo) return false;
      if (q && !m.nombre.toLowerCase().includes(q)) return false;
      return true;
    });
  });

  readonly impuestosFiltrados = computed(() => {
    const q = this.filtroTextoImpuesto().trim().toLowerCase();
    const estado = this.filtroEstadoImpuesto();
    return this.impuestos().filter((g) => {
      if (estado === 'activos' && !g.activo) return false;
      if (estado === 'inactivos' && g.activo) return false;
      if (q && !g.nombre.toLowerCase().includes(q) && !String(g.tarifa).includes(q)) return false;
      return true;
    });
  });

  readonly formMarca = this.fb.group({
    nombre: ['', [Validators.required, Validators.maxLength(100)]],
    activo: [true],
  });

  readonly formImpuesto = this.fb.group({
    nombre: ['', [Validators.required, Validators.maxLength(100)]],
    tarifa: [0 as number | null, [Validators.required, Validators.min(0), Validators.max(100)]],
    activo: [true],
  });

  ngOnInit(): void {
    this.cargarTodo();
  }

  cargarTodo(): void {
    this.loading.set(true);
    this.errorMensaje.set(null);

    forkJoin({
      marcas: this.productosSvc
        .listarMarcas()
        .pipe(catchError(() => of([] as MarcaProducto[]))),
      impuestos: this.productosSvc
        .listarGruposImpuesto()
        .pipe(catchError(() => of([] as GrupoImpuesto[]))),
    }).subscribe({
      next: ({ marcas, impuestos }) => {
        this.marcas.set(marcas);
        this.impuestos.set(impuestos);
        this.loading.set(false);
      },
      error: (err: HttpErrorResponse) => {
        this.errorMensaje.set(
          this.extraerError(err, 'No se pudo cargar la configuración de productos.'),
        );
        this.loading.set(false);
      },
    });
  }

  // =====================================================================
  // Marcas
  // =====================================================================

  abrirNuevaMarca(): void {
    this.editandoMarca.set(null);
    this.modalError.set(null);
    this.formMarca.reset({ nombre: '', activo: true });
    this.modalMarcaAbierto.set(true);
  }

  abrirEditarMarca(marca: MarcaProducto): void {
    this.editandoMarca.set(marca);
    this.modalError.set(null);
    this.formMarca.reset({ nombre: marca.nombre, activo: marca.activo });
    this.modalMarcaAbierto.set(true);
  }

  cerrarModalMarca(): void {
    if (this.saving()) return;
    this.modalMarcaAbierto.set(false);
    this.editandoMarca.set(null);
    this.modalError.set(null);
  }

  guardarMarca(): void {
    if (this.formMarca.invalid) {
      this.formMarca.markAllAsTouched();
      this.modalError.set('Ingresa un nombre válido para la marca.');
      return;
    }

    const raw = this.formMarca.getRawValue();
    const editando = this.editandoMarca();
    this.saving.set(true);
    this.modalError.set(null);

    const obs = editando
      ? this.productosSvc.actualizarMarca(editando.id, {
          nombre: raw.nombre!.trim(),
          activo: !!raw.activo,
        })
      : this.productosSvc.crearMarca({
          nombre: raw.nombre!.trim(),
          activo: !!raw.activo,
        });

    obs.subscribe({
      next: () => {
        this.saving.set(false);
        this.modalMarcaAbierto.set(false);
        this.editandoMarca.set(null);
        this.mostrarExito(editando ? 'Marca actualizada correctamente.' : 'Marca creada correctamente.');
        this.recargarMarcas();
      },
      error: (err: HttpErrorResponse) => {
        this.saving.set(false);
        this.modalError.set(this.extraerError(err, 'No se pudo guardar la marca.'));
      },
    });
  }

  toggleActivoMarca(marca: MarcaProducto): void {
    this.errorMensaje.set(null);
    const nuevoEstado = !marca.activo;
    this.productosSvc.actualizarMarca(marca.id, { activo: nuevoEstado }).subscribe({
      next: (actualizada) => {
        this.marcas.update((lista) =>
          lista.map((m) => (m.id === actualizada.id ? actualizada : m)),
        );
        this.mostrarExito(`Marca "${actualizada.nombre}" ${nuevoEstado ? 'activada' : 'desactivada'}.`);
      },
      error: (err: HttpErrorResponse) => {
        this.errorMensaje.set(this.extraerError(err, 'No se pudo cambiar el estado de la marca.'));
      },
    });
  }

  private recargarMarcas(): void {
    this.productosSvc.listarMarcas().subscribe({
      next: (lista) => this.marcas.set(lista),
      error: (err: HttpErrorResponse) =>
        this.errorMensaje.set(this.extraerError(err, 'No se pudieron recargar las marcas.')),
    });
  }

  // =====================================================================
  // Grupos de impuestos
  // =====================================================================

  abrirNuevoImpuesto(): void {
    this.editandoImpuesto.set(null);
    this.modalError.set(null);
    this.formImpuesto.reset({ nombre: '', tarifa: 0, activo: true });
    this.modalImpuestoAbierto.set(true);
  }

  abrirEditarImpuesto(grupo: GrupoImpuesto): void {
    this.editandoImpuesto.set(grupo);
    this.modalError.set(null);
    this.formImpuesto.reset({
      nombre: grupo.nombre,
      tarifa: grupo.tarifa,
      activo: grupo.activo,
    });
    this.modalImpuestoAbierto.set(true);
  }

  cerrarModalImpuesto(): void {
    if (this.saving()) return;
    this.modalImpuestoAbierto.set(false);
    this.editandoImpuesto.set(null);
    this.modalError.set(null);
  }

  guardarImpuesto(): void {
    if (this.formImpuesto.invalid) {
      this.formImpuesto.markAllAsTouched();
      this.modalError.set('Completa el nombre y una tarifa válida entre 0% y 100%.');
      return;
    }

    const raw = this.formImpuesto.getRawValue();
    const editando = this.editandoImpuesto();
    this.saving.set(true);
    this.modalError.set(null);

    const obs = editando
      ? this.productosSvc.actualizarGrupoImpuesto(editando.id, {
          nombre: raw.nombre!.trim(),
          tarifa: Number(raw.tarifa),
          activo: !!raw.activo,
        })
      : this.productosSvc.crearGrupoImpuesto({
          nombre: raw.nombre!.trim(),
          tarifa: Number(raw.tarifa),
          activo: !!raw.activo,
        });

    obs.subscribe({
      next: () => {
        this.saving.set(false);
        this.modalImpuestoAbierto.set(false);
        this.editandoImpuesto.set(null);
        this.mostrarExito(
          editando
            ? 'Grupo de impuesto actualizado correctamente.'
            : 'Grupo de impuesto creado correctamente.',
        );
        this.recargarImpuestos();
      },
      error: (err: HttpErrorResponse) => {
        this.saving.set(false);
        this.modalError.set(this.extraerError(err, 'No se pudo guardar el grupo de impuesto.'));
      },
    });
  }

  toggleActivoImpuesto(grupo: GrupoImpuesto): void {
    this.errorMensaje.set(null);
    const nuevoEstado = !grupo.activo;
    this.productosSvc.actualizarGrupoImpuesto(grupo.id, { activo: nuevoEstado }).subscribe({
      next: (actualizado) => {
        this.impuestos.update((lista) =>
          lista.map((g) => (g.id === actualizado.id ? actualizado : g)),
        );
        this.mostrarExito(
          `Grupo "${actualizado.nombre}" ${nuevoEstado ? 'activado' : 'desactivado'}.`,
        );
      },
      error: (err: HttpErrorResponse) => {
        this.errorMensaje.set(
          this.extraerError(err, 'No se pudo cambiar el estado del grupo de impuesto.'),
        );
      },
    });
  }

  private recargarImpuestos(): void {
    this.productosSvc.listarGruposImpuesto().subscribe({
      next: (lista) => this.impuestos.set(lista),
      error: (err: HttpErrorResponse) =>
        this.errorMensaje.set(
          this.extraerError(err, 'No se pudieron recargar los grupos de impuestos.'),
        ),
    });
  }

  // =====================================================================
  // Eliminación lógica (Soft Delete)
  // =====================================================================

  solicitarEliminar(tipo: ConfigTab, id: number, nombre: string): void {
    this.aEliminar.set({ tipo, id, nombre });
  }

  cancelarEliminar(): void {
    if (this.saving()) return;
    this.aEliminar.set(null);
  }

  confirmarEliminar(): void {
    const item = this.aEliminar();
    if (!item) return;

    this.saving.set(true);
    const obs =
      item.tipo === 'marcas'
        ? this.productosSvc.eliminarMarca(item.id)
        : this.productosSvc.eliminarGrupoImpuesto(item.id);

    obs.subscribe({
      next: () => {
        this.saving.set(false);
        this.aEliminar.set(null);
        this.mostrarExito(`"${item.nombre}" eliminado correctamente.`);
        if (item.tipo === 'marcas') {
          this.recargarMarcas();
        } else {
          this.recargarImpuestos();
        }
      },
      error: (err: HttpErrorResponse) => {
        this.saving.set(false);
        this.aEliminar.set(null);
        this.errorMensaje.set(this.extraerError(err, 'No se pudo eliminar el registro.'));
      },
    });
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
