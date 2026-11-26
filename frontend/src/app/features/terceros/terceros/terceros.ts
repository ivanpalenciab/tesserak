import { Component, OnInit, computed, inject, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { HttpErrorResponse } from '@angular/common/http';
import { catchError, of } from 'rxjs';

import { ApiErrorBody } from '../../../core/auth/auth.models';
import { TercerosService } from '../terceros.service';
import {
  ActualizarTerceroRequest,
  CrearTerceroRequest,
  NaturalezaJuridica,
  RolTercero,
  Tercero,
  TipoIdentificacion,
} from '../terceros.models';

export const TIPOS_IDENTIFICACION: { value: TipoIdentificacion; label: string }[] = [
  { value: 'NIT', label: 'NIT' },
  { value: 'CC', label: 'Cédula de ciudadanía (CC)' },
  { value: 'CE', label: 'Cédula de extranjería (CE)' },
  { value: 'PASAPORTE', label: 'Pasaporte' },
  { value: 'RUT', label: 'RUT' },
  { value: 'DNI', label: 'DNI' },
  { value: 'OTRO', label: 'Otro' },
];

export const NATURALEZAS_JURIDICAS: { value: NaturalezaJuridica; label: string }[] = [
  { value: 'JURIDICA', label: 'Persona jurídica' },
  { value: 'NATURAL', label: 'Persona natural' },
];

@Component({
  selector: 'app-terceros',
  standalone: true,
  imports: [ReactiveFormsModule],
  templateUrl: './terceros.html',
  styleUrl: './terceros.scss',
})
export class TercerosComponent implements OnInit {
  private readonly tercerosSvc = inject(TercerosService);
  private readonly fb = inject(FormBuilder);

  readonly tiposIdentificacion = TIPOS_IDENTIFICACION;
  readonly naturalezasJuridicas = NATURALEZAS_JURIDICAS;

  // Alertas globales
  readonly errorMensaje = signal<string | null>(null);
  readonly exitoMensaje = signal<string | null>(null);

  // Catálogo global de roles (solo lectura)
  readonly rolesDisponibles = signal<RolTercero[]>([]);

  // Estado Terceros
  readonly loadingTerceros = signal(true);
  readonly savingTercero = signal(false);
  readonly modalTerceroError = signal<string | null>(null);

  readonly terceros = signal<Tercero[]>([]);
  readonly totalTerceros = signal(0);
  readonly page = signal(1);
  readonly pageSize = 15;

  readonly filtroTexto = signal('');
  readonly filtroTipoId = signal<TipoIdentificacion | ''>('');
  readonly filtroNaturaleza = signal<NaturalezaJuridica | ''>('');
  readonly filtroRolId = signal<number | null>(null);
  readonly filtroEstado = signal<'todos' | 'activos' | 'inactivos'>('todos');

  readonly modalTerceroAbierto = signal(false);
  readonly editandoTercero = signal<Tercero | null>(null);
  readonly terceroAEliminar = signal<Tercero | null>(null);

  // Roles seleccionados en el modal de crear/editar tercero
  readonly rolesSeleccionados = signal<number[]>([]);
  readonly rolesTouched = signal(false);

  // Computed
  readonly totalPaginas = computed(() => Math.max(1, Math.ceil(this.totalTerceros() / this.pageSize)));

  readonly hayFiltrosTerceros = computed(
    () =>
      this.filtroTexto().trim().length > 0 ||
      this.filtroTipoId() !== '' ||
      this.filtroNaturaleza() !== '' ||
      this.filtroRolId() !== null ||
      this.filtroEstado() !== 'todos',
  );

  readonly formTercero = this.fb.group({
    tipo_identificacion: ['NIT' as TipoIdentificacion, [Validators.required]],
    numero_identificacion: ['', [Validators.required, Validators.maxLength(50)]],
    razon_social: ['', [Validators.required, Validators.maxLength(200)]],
    naturaleza_juridica: ['JURIDICA' as NaturalezaJuridica, [Validators.required]],
    email: ['', [Validators.email, Validators.maxLength(255)]],
    telefono: ['', [Validators.maxLength(50)]],
    direccion: ['', [Validators.maxLength(255)]],
    activo: [true],
  });

  ngOnInit(): void {
    this.tercerosSvc
      .listarRoles(true)
      .pipe(catchError(() => of([] as RolTercero[])))
      .subscribe((roles) => {
        this.rolesDisponibles.set(roles);
      });
    this.cargarTerceros();
  }

  cargarTerceros(): void {
    this.loadingTerceros.set(true);
    this.errorMensaje.set(null);

    const activo =
      this.filtroEstado() === 'todos' ? null : this.filtroEstado() === 'activos';

    this.tercerosSvc
      .listar({
        tipo_identificacion: this.filtroTipoId(),
        naturaleza_juridica: this.filtroNaturaleza(),
        rol_id: this.filtroRolId() ?? undefined,
        activo,
        q: this.filtroTexto(),
        page: this.page(),
        page_size: this.pageSize,
      })
      .subscribe({
        next: (res) => {
          this.terceros.set(res.data ?? []);
          this.totalTerceros.set(res.total ?? 0);
          this.loadingTerceros.set(false);
        },
        error: (err) => {
          this.loadingTerceros.set(false);
          this.errorMensaje.set(
            this.extraerMensajeError(err, 'No se pudieron cargar los terceros.'),
          );
        },
      });
  }

  onBuscarTercero(valor: string): void {
    this.filtroTexto.set(valor);
    this.page.set(1);
    this.cargarTerceros();
  }

  onFiltroTipoId(valor: string): void {
    this.filtroTipoId.set((valor || '') as TipoIdentificacion | '');
    this.page.set(1);
    this.cargarTerceros();
  }

  onFiltroNaturaleza(valor: string): void {
    this.filtroNaturaleza.set((valor || '') as NaturalezaJuridica | '');
    this.page.set(1);
    this.cargarTerceros();
  }

  onFiltroRol(valor: string): void {
    this.filtroRolId.set(valor ? Number(valor) : null);
    this.page.set(1);
    this.cargarTerceros();
  }

  onFiltroEstado(valor: string): void {
    this.filtroEstado.set(valor as 'todos' | 'activos' | 'inactivos');
    this.page.set(1);
    this.cargarTerceros();
  }

  limpiarFiltrosTerceros(): void {
    this.filtroTexto.set('');
    this.filtroTipoId.set('');
    this.filtroNaturaleza.set('');
    this.filtroRolId.set(null);
    this.filtroEstado.set('todos');
    this.page.set(1);
    this.cargarTerceros();
  }

  irPagina(p: number): void {
    if (p < 1 || p > this.totalPaginas()) return;
    this.page.set(p);
    this.cargarTerceros();
  }

  // =====================================================================
  // Selección múltiple de roles en el formulario
  // =====================================================================

  tieneRolSeleccionado(rolId: number): boolean {
    return this.rolesSeleccionados().includes(rolId);
  }

  toggleRolSeleccionado(rolId: number): void {
    this.rolesTouched.set(true);
    this.rolesSeleccionados.update((actuales) =>
      actuales.includes(rolId) ? actuales.filter((id) => id !== rolId) : [...actuales, rolId],
    );
  }

  // =====================================================================
  // Crear / Editar / Eliminar Tercero
  // =====================================================================

  abrirNuevoTercero(): void {
    this.editandoTercero.set(null);
    this.modalTerceroError.set(null);
    this.rolesSeleccionados.set([]);
    this.rolesTouched.set(false);
    this.formTercero.reset({
      tipo_identificacion: 'NIT',
      numero_identificacion: '',
      razon_social: '',
      naturaleza_juridica: 'JURIDICA',
      email: '',
      telefono: '',
      direccion: '',
      activo: true,
    });
    this.modalTerceroAbierto.set(true);
  }

  abrirEditarTercero(t: Tercero): void {
    this.editandoTercero.set(t);
    this.modalTerceroError.set(null);
    this.rolesSeleccionados.set((t.roles ?? []).map((r) => r.rol_id));
    this.rolesTouched.set(false);
    this.formTercero.reset({
      tipo_identificacion: t.tipo_identificacion,
      numero_identificacion: t.numero_identificacion,
      razon_social: t.razon_social,
      naturaleza_juridica: t.naturaleza_juridica,
      email: t.email ?? '',
      telefono: t.telefono ?? '',
      direccion: t.direccion ?? '',
      activo: t.activo,
    });
    this.modalTerceroAbierto.set(true);
  }

  cerrarModalTercero(): void {
    if (this.savingTercero()) return;
    this.modalTerceroAbierto.set(false);
    this.editandoTercero.set(null);
    this.modalTerceroError.set(null);
  }

  guardarTercero(): void {
    this.modalTerceroError.set(null);
    this.rolesTouched.set(true);

    if (this.formTercero.invalid || this.rolesSeleccionados().length === 0) {
      this.formTercero.markAllAsTouched();
      return;
    }

    const v = this.formTercero.getRawValue();
    const rolIds = [...this.rolesSeleccionados()];
    this.savingTercero.set(true);
    const actual = this.editandoTercero();

    if (actual) {
      const payload: ActualizarTerceroRequest = {
        tipo_identificacion: v.tipo_identificacion as TipoIdentificacion,
        numero_identificacion: (v.numero_identificacion ?? '').trim(),
        razon_social: (v.razon_social ?? '').trim(),
        naturaleza_juridica: v.naturaleza_juridica as NaturalezaJuridica,
        email: (v.email ?? '').trim() || null,
        telefono: (v.telefono ?? '').trim() || null,
        direccion: (v.direccion ?? '').trim() || null,
        activo: Boolean(v.activo),
        rol_ids: rolIds,
      };

      this.tercerosSvc.actualizar(actual.id, payload).subscribe({
        next: (actualizado) => {
          this.savingTercero.set(false);
          this.modalTerceroAbierto.set(false);
          this.editandoTercero.set(null);
          this.mostrarExito(`Tercero "${actualizado.razon_social}" actualizado.`);
          this.cargarTerceros();
        },
        error: (err) => {
          this.savingTercero.set(false);
          this.modalTerceroError.set(
            this.extraerMensajeError(err, 'No se pudo actualizar el tercero.'),
          );
        },
      });
    } else {
      const payload: CrearTerceroRequest = {
        tipo_identificacion: v.tipo_identificacion as TipoIdentificacion,
        numero_identificacion: (v.numero_identificacion ?? '').trim(),
        razon_social: (v.razon_social ?? '').trim(),
        naturaleza_juridica: v.naturaleza_juridica as NaturalezaJuridica,
        email: (v.email ?? '').trim() || null,
        telefono: (v.telefono ?? '').trim() || null,
        direccion: (v.direccion ?? '').trim() || null,
        activo: Boolean(v.activo),
        rol_ids: rolIds,
      };

      this.tercerosSvc.crear(payload).subscribe({
        next: (creado) => {
          this.savingTercero.set(false);
          this.modalTerceroAbierto.set(false);
          this.mostrarExito(`Tercero "${creado.razon_social}" registrado correctamente.`);
          this.page.set(1);
          this.cargarTerceros();
        },
        error: (err) => {
          this.savingTercero.set(false);
          this.modalTerceroError.set(
            this.extraerMensajeError(err, 'No se pudo registrar el tercero.'),
          );
        },
      });
    }
  }

  toggleActivoTercero(t: Tercero): void {
    const nuevoEstado = !t.activo;
    this.tercerosSvc.actualizar(t.id, { activo: nuevoEstado }).subscribe({
      next: () => {
        this.mostrarExito(
          `Tercero "${t.razon_social}" ${nuevoEstado ? 'activado' : 'desactivado'}.`,
        );
        this.cargarTerceros();
      },
      error: (err) => {
        this.errorMensaje.set(
          this.extraerMensajeError(err, 'No se pudo cambiar el estado del tercero.'),
        );
      },
    });
  }

  solicitarEliminarTercero(t: Tercero): void {
    this.terceroAEliminar.set(t);
  }

  cancelarEliminarTercero(): void {
    if (this.savingTercero()) return;
    this.terceroAEliminar.set(null);
  }

  confirmarEliminarTercero(): void {
    const target = this.terceroAEliminar();
    if (!target) return;

    this.savingTercero.set(true);
    this.tercerosSvc.eliminar(target.id).subscribe({
      next: () => {
        this.savingTercero.set(false);
        this.terceroAEliminar.set(null);
        this.mostrarExito(`Tercero "${target.razon_social}" eliminado.`);
        this.cargarTerceros();
      },
      error: (err) => {
        this.savingTercero.set(false);
        this.terceroAEliminar.set(null);
        this.errorMensaje.set(this.extraerMensajeError(err, 'No se pudo eliminar el tercero.'));
      },
    });
  }

  etiquetaNaturaleza(nat: NaturalezaJuridica): string {
    return nat === 'JURIDICA' ? 'Jurídica' : 'Natural';
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
