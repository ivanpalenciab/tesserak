import { Component, OnInit, computed, inject, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { HttpErrorResponse } from '@angular/common/http';
import { catchError, forkJoin, of } from 'rxjs';

import { AuthService } from '../../../core/auth/auth.service';
import { ApiErrorBody } from '../../../core/auth/auth.models';
import { CrmService } from '../crm.service';
import {
  EstadoLead,
  EtapaOportunidad,
  MotivoDescarteLead,
  OrigenLead,
} from '../crm.models';

type ConfigTab = 'origenes' | 'etapas' | 'etapas_oportunidad' | 'motivos';

interface ConfirmDeleteState {
  tipo: ConfigTab;
  id: number;
  etiqueta: string;
}

@Component({
  selector: 'app-crm-configuracion',
  standalone: true,
  imports: [ReactiveFormsModule],
  templateUrl: './configuracion.html',
  styleUrl: './configuracion.scss',
})
export class CrmConfiguracionComponent implements OnInit {
  private readonly crm = inject(CrmService);
  private readonly auth = inject(AuthService);
  private readonly fb = inject(FormBuilder);

  readonly usuario = this.auth.usuario;

  readonly tabActiva = signal<ConfigTab>('origenes');
  readonly loading = signal(true);
  readonly saving = signal(false);
  readonly errorMensaje = signal<string | null>(null);
  readonly exitoMensaje = signal<string | null>(null);

  // Datos de los 4 catálogos
  readonly origenes = signal<OrigenLead[]>([]);
  readonly etapas = signal<EstadoLead[]>([]);
  readonly etapasOportunidad = signal<EtapaOportunidad[]>([]);
  readonly motivos = signal<MotivoDescarteLead[]>([]);

  readonly filtroEstadoOrigen = signal<'todos' | 'activos' | 'inactivos'>('todos');
  readonly filtroEstadoMotivo = signal<'todos' | 'activos' | 'inactivos'>('todos');

  readonly origenesFiltrados = computed(() => {
    const f = this.filtroEstadoOrigen();
    const lista = this.origenes();
    if (f === 'activos') return lista.filter((o) => o.activo);
    if (f === 'inactivos') return lista.filter((o) => !o.activo);
    return lista;
  });

  readonly motivosFiltrados = computed(() => {
    const f = this.filtroEstadoMotivo();
    const lista = this.motivos();
    if (f === 'activos') return lista.filter((m) => m.activo);
    if (f === 'inactivos') return lista.filter((m) => !m.activo);
    return lista;
  });

  // Modales
  readonly modalOrigenAbierto = signal(false);
  readonly origenEditando = signal<OrigenLead | null>(null);

  readonly modalEtapaAbierto = signal(false);
  readonly etapaEditando = signal<EstadoLead | null>(null);

  readonly modalEtapaOpAbierto = signal(false);
  readonly etapaOpEditando = signal<EtapaOportunidad | null>(null);

  readonly modalMotivoAbierto = signal(false);
  readonly motivoEditando = signal<MotivoDescarteLead | null>(null);

  readonly eliminando = signal<ConfirmDeleteState | null>(null);

  // Formularios
  readonly origenForm = this.fb.group({
    nombre: ['', [Validators.required, Validators.maxLength(100)]],
    activo: [true],
  });

  readonly etapaForm = this.fb.group({
    nombre: ['', [Validators.required, Validators.maxLength(50)]],
    orden: [1, [Validators.required, Validators.min(0)]],
    es_estado_final: [false],
  });

  readonly etapaOpForm = this.fb.group({
    nombre: ['', [Validators.required, Validators.maxLength(80)]],
    orden: [1, [Validators.required, Validators.min(0)]],
    probabilidad_pct: [0, [Validators.required, Validators.min(0), Validators.max(100)]],
    es_ganada: [false],
    es_perdida: [false],
    activo: [true],
  });

  readonly motivoForm = this.fb.group({
    descripcion: ['', [Validators.required, Validators.maxLength(150)]],
    activo: [true],
  });

  ngOnInit(): void {
    this.cargarTodo();
  }

  private get empresaId(): number {
    return this.usuario()?.empresa_id ?? 1;
  }

  cargarTodo(): void {
    this.loading.set(true);
    this.errorMensaje.set(null);

    forkJoin({
      origenes: this.crm.listarOrigenes(this.empresaId).pipe(catchError(() => of([] as OrigenLead[]))),
      etapas: this.crm.listarEstados(this.empresaId).pipe(catchError(() => of([] as EstadoLead[]))),
      etapasOp: this.crm
        .listarEtapasOportunidad(this.empresaId)
        .pipe(catchError(() => of([] as EtapaOportunidad[]))),
      motivos: this.crm
        .listarMotivosDescarte(this.empresaId)
        .pipe(catchError(() => of([] as MotivoDescarteLead[]))),
    }).subscribe(({ origenes, etapas, etapasOp, motivos }) => {
      this.origenes.set(origenes);
      this.etapas.set([...etapas].sort((a, b) => a.orden - b.orden));
      this.etapasOportunidad.set([...etapasOp].sort((a, b) => a.orden - b.orden));
      this.motivos.set(motivos);
      this.loading.set(false);
    });
  }

  // =====================================================================
  // 1. Orígenes (CRUD con soft delete)
  // =====================================================================

  abrirNuevoOrigen(): void {
    this.origenEditando.set(null);
    this.errorMensaje.set(null);
    this.origenForm.reset({ nombre: '', activo: true });
    this.modalOrigenAbierto.set(true);
  }

  abrirEditarOrigen(origen: OrigenLead): void {
    this.origenEditando.set(origen);
    this.errorMensaje.set(null);
    this.origenForm.reset({ nombre: origen.nombre, activo: origen.activo });
    this.modalOrigenAbierto.set(true);
  }

  cerrarModalOrigen(): void {
    if (this.saving()) return;
    this.modalOrigenAbierto.set(false);
    this.origenEditando.set(null);
  }

  guardarOrigen(): void {
    if (this.origenForm.invalid) {
      this.origenForm.markAllAsTouched();
      return;
    }
    const raw = this.origenForm.getRawValue();
    const editando = this.origenEditando();
    this.saving.set(true);
    this.errorMensaje.set(null);

    const obs = editando
      ? this.crm.actualizarOrigen(editando.id, { nombre: raw.nombre!.trim(), activo: !!raw.activo })
      : this.crm.crearOrigen({ empresa_id: this.empresaId, nombre: raw.nombre!.trim(), activo: !!raw.activo });

    obs.subscribe({
      next: () => {
        this.saving.set(false);
        this.modalOrigenAbierto.set(false);
        this.origenEditando.set(null);
        this.mostrarExito(editando ? 'Origen actualizado.' : 'Origen creado.');
        this.recargarOrigenes();
      },
      error: (err: HttpErrorResponse) => {
        this.saving.set(false);
        this.errorMensaje.set(this.extraerError(err, 'No se pudo guardar el origen.'));
      },
    });
  }

  toggleActivoOrigen(origen: OrigenLead): void {
    this.crm.actualizarOrigen(origen.id, { activo: !origen.activo }).subscribe({
      next: () => {
        this.mostrarExito(`Origen "${origen.nombre}" ${!origen.activo ? 'activado' : 'desactivado'}.`);
        this.recargarOrigenes();
      },
      error: (err: HttpErrorResponse) =>
        this.errorMensaje.set(this.extraerError(err, 'No se pudo actualizar el origen.')),
    });
  }

  private recargarOrigenes(): void {
    this.crm.listarOrigenes(this.empresaId).subscribe({
      next: (lista) => this.origenes.set(lista),
      error: (err: HttpErrorResponse) =>
        this.errorMensaje.set(this.extraerError(err, 'No se pudieron recargar los orígenes.')),
    });
  }

  // =====================================================================
  // 2. Etapas de leads (CRUD con soft delete)
  // =====================================================================

  abrirNuevaEtapa(): void {
    this.etapaEditando.set(null);
    this.errorMensaje.set(null);
    const sigOrden = this.etapas().length + 1;
    this.etapaForm.reset({ nombre: '', orden: sigOrden, es_estado_final: false });
    this.modalEtapaAbierto.set(true);
  }

  abrirEditarEtapa(etapa: EstadoLead): void {
    this.etapaEditando.set(etapa);
    this.errorMensaje.set(null);
    this.etapaForm.reset({
      nombre: etapa.nombre,
      orden: etapa.orden,
      es_estado_final: etapa.es_estado_final,
    });
    this.modalEtapaAbierto.set(true);
  }

  cerrarModalEtapa(): void {
    if (this.saving()) return;
    this.modalEtapaAbierto.set(false);
    this.etapaEditando.set(null);
  }

  guardarEtapa(): void {
    if (this.etapaForm.invalid) {
      this.etapaForm.markAllAsTouched();
      return;
    }
    const raw = this.etapaForm.getRawValue();
    const editando = this.etapaEditando();
    this.saving.set(true);
    this.errorMensaje.set(null);

    const obs = editando
      ? this.crm.actualizarEstado(editando.id, {
          nombre: raw.nombre!.trim(),
          orden: Number(raw.orden ?? 1),
          es_estado_final: !!raw.es_estado_final,
        })
      : this.crm.crearEstado({
          empresa_id: this.empresaId,
          nombre: raw.nombre!.trim(),
          orden: Number(raw.orden ?? 1),
          es_estado_final: !!raw.es_estado_final,
        });

    obs.subscribe({
      next: () => {
        this.saving.set(false);
        this.modalEtapaAbierto.set(false);
        this.etapaEditando.set(null);
        this.mostrarExito(editando ? 'Etapa actualizada.' : 'Etapa creada.');
        this.recargarEtapas();
      },
      error: (err: HttpErrorResponse) => {
        this.saving.set(false);
        this.errorMensaje.set(this.extraerError(err, 'No se pudo guardar la etapa.'));
      },
    });
  }

  private recargarEtapas(): void {
    this.crm.listarEstados(this.empresaId).subscribe({
      next: (lista) => this.etapas.set([...lista].sort((a, b) => a.orden - b.orden)),
      error: (err: HttpErrorResponse) =>
        this.errorMensaje.set(this.extraerError(err, 'No se pudieron recargar las etapas.')),
    });
  }

  // =====================================================================
  // 3. Etapas de oportunidades (CRUD con soft delete)
  // =====================================================================

  abrirNuevaEtapaOp(): void {
    this.etapaOpEditando.set(null);
    this.errorMensaje.set(null);
    const sigOrden = this.etapasOportunidad().length + 1;
    this.etapaOpForm.reset({
      nombre: '',
      orden: sigOrden,
      probabilidad_pct: 10,
      es_ganada: false,
      es_perdida: false,
      activo: true,
    });
    this.modalEtapaOpAbierto.set(true);
  }

  abrirEditarEtapaOp(etapa: EtapaOportunidad): void {
    this.etapaOpEditando.set(etapa);
    this.errorMensaje.set(null);
    this.etapaOpForm.reset({
      nombre: etapa.nombre,
      orden: etapa.orden,
      probabilidad_pct: etapa.probabilidad_pct,
      es_ganada: etapa.es_ganada,
      es_perdida: etapa.es_perdida,
      activo: etapa.activo,
    });
    this.modalEtapaOpAbierto.set(true);
  }

  cerrarModalEtapaOp(): void {
    if (this.saving()) return;
    this.modalEtapaOpAbierto.set(false);
    this.etapaOpEditando.set(null);
  }

  onToggleGanada(checked: boolean): void {
    if (checked) {
      this.etapaOpForm.patchValue({ es_perdida: false, probabilidad_pct: 100 });
    }
  }

  onTogglePerdida(checked: boolean): void {
    if (checked) {
      this.etapaOpForm.patchValue({ es_ganada: false, probabilidad_pct: 0 });
    }
  }

  guardarEtapaOp(): void {
    if (this.etapaOpForm.invalid) {
      this.etapaOpForm.markAllAsTouched();
      return;
    }
    const raw = this.etapaOpForm.getRawValue();
    if (raw.es_ganada && raw.es_perdida) {
      this.errorMensaje.set('Una etapa no puede ser ganada y perdida al mismo tiempo.');
      return;
    }

    const editando = this.etapaOpEditando();
    this.saving.set(true);
    this.errorMensaje.set(null);

    const obs = editando
      ? this.crm.actualizarEtapaOportunidad(editando.id, {
          nombre: raw.nombre!.trim(),
          orden: Number(raw.orden ?? 1),
          probabilidad_pct: Number(raw.probabilidad_pct ?? 0),
          es_ganada: !!raw.es_ganada,
          es_perdida: !!raw.es_perdida,
          activo: !!raw.activo,
        })
      : this.crm.crearEtapaOportunidad({
          empresa_id: this.empresaId,
          nombre: raw.nombre!.trim(),
          orden: Number(raw.orden ?? 1),
          probabilidad_pct: Number(raw.probabilidad_pct ?? 0),
          es_ganada: !!raw.es_ganada,
          es_perdida: !!raw.es_perdida,
          activo: !!raw.activo,
        });

    obs.subscribe({
      next: () => {
        this.saving.set(false);
        this.modalEtapaOpAbierto.set(false);
        this.etapaOpEditando.set(null);
        this.mostrarExito(editando ? 'Etapa de oportunidad actualizada.' : 'Etapa de oportunidad creada.');
        this.recargarEtapasOp();
      },
      error: (err: HttpErrorResponse) => {
        this.saving.set(false);
        this.errorMensaje.set(this.extraerError(err, 'No se pudo guardar la etapa de oportunidad.'));
      },
    });
  }

  private recargarEtapasOp(): void {
    this.crm.listarEtapasOportunidad(this.empresaId).subscribe({
      next: (lista) => this.etapasOportunidad.set([...lista].sort((a, b) => a.orden - b.orden)),
      error: (err: HttpErrorResponse) =>
        this.errorMensaje.set(this.extraerError(err, 'No se pudieron recargar las etapas de oportunidad.')),
    });
  }

  // =====================================================================
  // 4. Motivos de descarte (CRUD con soft delete)
  // =====================================================================

  abrirNuevoMotivo(): void {
    this.motivoEditando.set(null);
    this.errorMensaje.set(null);
    this.motivoForm.reset({ descripcion: '', activo: true });
    this.modalMotivoAbierto.set(true);
  }

  abrirEditarMotivo(motivo: MotivoDescarteLead): void {
    this.motivoEditando.set(motivo);
    this.errorMensaje.set(null);
    this.motivoForm.reset({ descripcion: motivo.descripcion, activo: motivo.activo });
    this.modalMotivoAbierto.set(true);
  }

  cerrarModalMotivo(): void {
    if (this.saving()) return;
    this.modalMotivoAbierto.set(false);
    this.motivoEditando.set(null);
  }

  guardarMotivo(): void {
    if (this.motivoForm.invalid) {
      this.motivoForm.markAllAsTouched();
      return;
    }
    const raw = this.motivoForm.getRawValue();
    const editando = this.motivoEditando();
    this.saving.set(true);
    this.errorMensaje.set(null);

    const obs = editando
      ? this.crm.actualizarMotivoDescarte(editando.id, {
          descripcion: raw.descripcion!.trim(),
          activo: !!raw.activo,
        })
      : this.crm.crearMotivoDescarte({
          empresa_id: this.empresaId,
          descripcion: raw.descripcion!.trim(),
          activo: !!raw.activo,
        });

    obs.subscribe({
      next: () => {
        this.saving.set(false);
        this.modalMotivoAbierto.set(false);
        this.motivoEditando.set(null);
        this.mostrarExito(editando ? 'Motivo de descarte actualizado.' : 'Motivo de descarte creado.');
        this.recargarMotivos();
      },
      error: (err: HttpErrorResponse) => {
        this.saving.set(false);
        this.errorMensaje.set(this.extraerError(err, 'No se pudo guardar el motivo de descarte.'));
      },
    });
  }

  toggleActivoMotivo(motivo: MotivoDescarteLead): void {
    this.crm.actualizarMotivoDescarte(motivo.id, { activo: !motivo.activo }).subscribe({
      next: () => {
        this.mostrarExito(`Motivo ${!motivo.activo ? 'activado' : 'desactivado'}.`);
        this.recargarMotivos();
      },
      error: (err: HttpErrorResponse) =>
        this.errorMensaje.set(this.extraerError(err, 'No se pudo actualizar el motivo.')),
    });
  }

  private recargarMotivos(): void {
    this.crm.listarMotivosDescarte(this.empresaId).subscribe({
      next: (lista) => this.motivos.set(lista),
      error: (err: HttpErrorResponse) =>
        this.errorMensaje.set(this.extraerError(err, 'No se pudieron recargar los motivos.')),
    });
  }

  // =====================================================================
  // Eliminación lógica (soft delete) unificada
  // =====================================================================

  solicitarEliminar(tipo: ConfigTab, id: number, etiqueta: string): void {
    this.eliminando.set({ tipo, id, etiqueta });
  }

  cancelarEliminar(): void {
    if (this.saving()) return;
    this.eliminando.set(null);
  }

  confirmarEliminar(): void {
    const item = this.eliminando();
    if (!item) return;

    this.saving.set(true);
    this.errorMensaje.set(null);

    const onSuccess = () => {
      this.saving.set(false);
      this.eliminando.set(null);
      this.mostrarExito(`"${item.etiqueta}" eliminado correctamente.`);
      if (item.tipo === 'origenes') this.recargarOrigenes();
      if (item.tipo === 'etapas') this.recargarEtapas();
      if (item.tipo === 'etapas_oportunidad') this.recargarEtapasOp();
      if (item.tipo === 'motivos') this.recargarMotivos();
    };

    const onError = (err: HttpErrorResponse) => {
      this.saving.set(false);
      this.eliminando.set(null);
      this.errorMensaje.set(this.extraerError(err, 'No se pudo eliminar el registro.'));
    };

    switch (item.tipo) {
      case 'origenes':
        this.crm.eliminarOrigen(item.id).subscribe({ next: onSuccess, error: onError });
        break;
      case 'etapas':
        this.crm.eliminarEstado(item.id).subscribe({ next: onSuccess, error: onError });
        break;
      case 'etapas_oportunidad':
        this.crm.eliminarEtapaOportunidad(item.id).subscribe({ next: onSuccess, error: onError });
        break;
      case 'motivos':
        this.crm.eliminarMotivoDescarte(item.id).subscribe({ next: onSuccess, error: onError });
        break;
    }
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
