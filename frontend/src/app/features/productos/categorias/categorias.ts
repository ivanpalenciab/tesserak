import { Component, OnInit, computed, inject, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { HttpErrorResponse } from '@angular/common/http';

import { ApiErrorBody } from '../../../core/auth/auth.models';
import { ProductosService } from '../productos.service';
import { CategoriaProducto } from '../productos.models';

@Component({
  selector: 'app-productos-categorias',
  standalone: true,
  imports: [ReactiveFormsModule],
  templateUrl: './categorias.html',
  styleUrl: './categorias.scss',
})
export class ProductosCategoriasComponent implements OnInit {
  private readonly productosSvc = inject(ProductosService);
  private readonly fb = inject(FormBuilder);

  readonly loading = signal(true);
  readonly saving = signal(false);
  readonly errorMensaje = signal<string | null>(null);
  readonly exitoMensaje = signal<string | null>(null);

  readonly categorias = signal<CategoriaProducto[]>([]);
  readonly filtroTexto = signal('');
  readonly filtroEstado = signal<'todos' | 'activos' | 'inactivos'>('todos');

  readonly modalAbierto = signal(false);
  readonly editando = signal<CategoriaProducto | null>(null);
  readonly aEliminar = signal<CategoriaProducto | null>(null);

  readonly hayFiltrosActivos = computed(
    () => this.filtroTexto().trim().length > 0 || this.filtroEstado() !== 'todos',
  );

  readonly form = this.fb.group({
    nombre: ['', [Validators.required, Validators.maxLength(100)]],
    activo: [true],
  });

  ngOnInit(): void {
    this.cargarCategorias();
  }

  cargarCategorias(): void {
    this.loading.set(true);
    this.errorMensaje.set(null);

    const activo =
      this.filtroEstado() === 'todos' ? undefined : this.filtroEstado() === 'activos';

    this.productosSvc.listarCategorias(activo, this.filtroTexto()).subscribe({
      next: (lista) => {
        this.categorias.set(lista);
        this.loading.set(false);
      },
      error: (err: HttpErrorResponse) => {
        this.categorias.set([]);
        this.errorMensaje.set(this.extraerError(err, 'No se pudieron cargar las categorías.'));
        this.loading.set(false);
      },
    });
  }

  onBuscarTexto(valor: string): void {
    this.filtroTexto.set(valor);
    this.cargarCategorias();
  }

  onFiltroEstado(valor: string): void {
    this.filtroEstado.set(valor as 'todos' | 'activos' | 'inactivos');
    this.cargarCategorias();
  }

  limpiarFiltros(): void {
    this.filtroTexto.set('');
    this.filtroEstado.set('todos');
    this.cargarCategorias();
  }

  abrirNueva(): void {
    this.editando.set(null);
    this.errorMensaje.set(null);
    this.form.reset({ nombre: '', activo: true });
    this.modalAbierto.set(true);
  }

  abrirEditar(cat: CategoriaProducto): void {
    this.editando.set(cat);
    this.errorMensaje.set(null);
    this.form.reset({ nombre: cat.nombre, activo: cat.activo });
    this.modalAbierto.set(true);
  }

  cerrarModal(): void {
    if (this.saving()) return;
    this.modalAbierto.set(false);
    this.editando.set(null);
  }

  guardar(): void {
    if (this.form.invalid) {
      this.form.markAllAsTouched();
      return;
    }

    const raw = this.form.getRawValue();
    const editando = this.editando();
    this.saving.set(true);
    this.errorMensaje.set(null);

    const obs = editando
      ? this.productosSvc.actualizarCategoria(editando.id, {
          nombre: raw.nombre!.trim(),
          activo: !!raw.activo,
        })
      : this.productosSvc.crearCategoria({
          nombre: raw.nombre!.trim(),
          activo: !!raw.activo,
        });

    obs.subscribe({
      next: () => {
        this.saving.set(false);
        this.modalAbierto.set(false);
        this.editando.set(null);
        this.mostrarExito(editando ? 'Categoría actualizada.' : 'Categoría creada.');
        this.cargarCategorias();
      },
      error: (err: HttpErrorResponse) => {
        this.saving.set(false);
        this.errorMensaje.set(this.extraerError(err, 'No se pudo guardar la categoría.'));
      },
    });
  }

  toggleActivo(cat: CategoriaProducto): void {
    this.errorMensaje.set(null);
    const nuevoEstado = !cat.activo;
    this.productosSvc.actualizarCategoria(cat.id, { activo: nuevoEstado }).subscribe({
      next: (actualizada) => {
        this.categorias.update((lista) =>
          lista.map((c) => (c.id === actualizada.id ? actualizada : c)),
        );
        this.mostrarExito(
          `Categoría "${actualizada.nombre}" ${nuevoEstado ? 'activada' : 'desactivada'}.`,
        );
      },
      error: (err: HttpErrorResponse) => {
        this.errorMensaje.set(
          this.extraerError(err, 'No se pudo cambiar el estado de la categoría.'),
        );
      },
    });
  }

  solicitarEliminar(cat: CategoriaProducto): void {
    this.aEliminar.set(cat);
  }

  cancelarEliminar(): void {
    if (this.saving()) return;
    this.aEliminar.set(null);
  }

  confirmarEliminar(): void {
    const cat = this.aEliminar();
    if (!cat) return;

    this.saving.set(true);
    this.productosSvc.eliminarCategoria(cat.id).subscribe({
      next: () => {
        this.saving.set(false);
        this.aEliminar.set(null);
        this.mostrarExito(`Categoría "${cat.nombre}" eliminada.`);
        this.cargarCategorias();
      },
      error: (err: HttpErrorResponse) => {
        this.saving.set(false);
        this.aEliminar.set(null);
        this.errorMensaje.set(this.extraerError(err, 'No se pudo eliminar la categoría.'));
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
