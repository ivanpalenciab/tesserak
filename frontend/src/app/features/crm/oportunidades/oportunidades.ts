import { Component, OnInit, computed, inject, signal } from '@angular/core';
import { CurrencyPipe } from '@angular/common';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { HttpErrorResponse } from '@angular/common/http';
import { catchError, forkJoin, of } from 'rxjs';

import { AuthService } from '../../../core/auth/auth.service';
import { ApiErrorBody } from '../../../core/auth/auth.models';
import { CrmService } from '../crm.service';
import {
  ActualizarOportunidadRequest,
  Campana,
  CrearOportunidadRequest,
  EtapaOportunidad,
  Lead,
  MotivoDescarteLead,
  Oportunidad,
  PrioridadOportunidad,
} from '../crm.models';

@Component({
  selector: 'app-crm-oportunidades',
  standalone: true,
  imports: [ReactiveFormsModule, CurrencyPipe],
  templateUrl: './oportunidades.html',
  styleUrl: './oportunidades.scss',
})
export class CrmOportunidadesComponent implements OnInit {
  private readonly crm = inject(CrmService);
  private readonly auth = inject(AuthService);
  private readonly fb = inject(FormBuilder);

  readonly usuario = this.auth.usuario;

  readonly loading = signal(true);
  readonly saving = signal(false);
  readonly errorMensaje = signal<string | null>(null);
  readonly exitoMensaje = signal<string | null>(null);

  readonly oportunidades = signal<Oportunidad[]>([]);
  readonly total = signal(0);
  readonly page = signal(1);
  readonly pageSize = 15;

  readonly etapas = signal<EtapaOportunidad[]>([]);
  readonly leads = signal<Lead[]>([]);
  readonly campanas = signal<Campana[]>([]);
  readonly motivos = signal<MotivoDescarteLead[]>([]);

  readonly filtroTexto = signal('');
  readonly filtroEstadoId = signal<number | null>(null);
  readonly filtroPrioridad = signal<PrioridadOportunidad | ''>('');
  readonly filtroCampanaId = signal<number | null>(null);

  readonly modalAbierto = signal(false);
  readonly editando = signal<Oportunidad | null>(null);
  readonly aEliminar = signal<Oportunidad | null>(null);

  readonly prioridades: { value: PrioridadOportunidad; label: string }[] = [
    { value: 'BAJA', label: 'Baja' },
    { value: 'MEDIA', label: 'Media' },
    { value: 'ALTA', label: 'Alta' },
  ];

  readonly totalPaginas = computed(() => Math.max(1, Math.ceil(this.total() / this.pageSize)));
  readonly etapasMap = computed(() => new Map(this.etapas().map((e) => [e.id, e])));
  readonly leadsMap = computed(() => new Map(this.leads().map((l) => [l.id, l])));
  readonly campanasMap = computed(() => new Map(this.campanas().map((c) => [c.id, c])));

  readonly hayFiltrosActivos = computed(
    () =>
      this.filtroTexto().trim().length > 0 ||
      this.filtroEstadoId() !== null ||
      this.filtroPrioridad() !== '' ||
      this.filtroCampanaId() !== null,
  );

  readonly form = this.fb.group({
    titulo: ['', [Validators.required, Validators.maxLength(150)]],
    nombre_cliente: ['', [Validators.maxLength(150)]],
    lead_id: [null as number | null],
    campana_id: [null as number | null],
    estado_oportunidad_id: [null as number | null, [Validators.required]],
    motivo_descarte_id: [null as number | null],
    prioridad: ['MEDIA' as PrioridadOportunidad, [Validators.required]],
    valor_estimado: [0, [Validators.min(0)]],
    moneda_id: [null as number | null, [Validators.min(1)]],
    probabilidad_pct: [0, [Validators.min(0), Validators.max(100)]],
    fecha_cierre_estimada: [''],
    descripcion: ['', [Validators.maxLength(5000)]],
  });

  ngOnInit(): void {
    this.cargarCatalogosYOportunidades();
  }

  private get empresaId(): number {
    return this.usuario()?.empresa_id ?? 1;
  }

  cargarCatalogosYOportunidades(): void {
    const empresaId = this.empresaId;
    this.loading.set(true);
    this.errorMensaje.set(null);

    forkJoin({
      etapas: this.crm
        .listarEtapasOportunidad(empresaId, true)
        .pipe(catchError(() => of([] as EtapaOportunidad[]))),
      leads: this.crm
        .listarLeads({ empresa_id: empresaId, page: 1, page_size: 100 })
        .pipe(catchError(() => of({ data: [] as Lead[], page: 1, page_size: 100, total: 0 }))),
      campanas: this.crm.listarCampanas().pipe(catchError(() => of([] as Campana[]))),
      motivos: this.crm
        .listarMotivosDescarte(empresaId, true)
        .pipe(catchError(() => of([] as MotivoDescarteLead[]))),
    }).subscribe(({ etapas, leads, campanas, motivos }) => {
      this.etapas.set([...etapas].sort((a, b) => a.orden - b.orden));
      this.leads.set(leads.data ?? []);
      this.campanas.set(campanas);
      this.motivos.set(motivos);
      this.cargarOportunidades();
    });
  }

  cargarOportunidades(): void {
    this.loading.set(true);
    this.errorMensaje.set(null);

    this.crm
      .listarOportunidades({
        empresa_id: this.empresaId,
        estado_oportunidad_id: this.filtroEstadoId() ?? undefined,
        prioridad: this.filtroPrioridad(),
        campana_id: this.filtroCampanaId() ?? undefined,
        q: this.filtroTexto(),
        page: this.page(),
        page_size: this.pageSize,
      })
      .subscribe({
        next: (res) => {
          this.oportunidades.set(res.data ?? []);
          this.total.set(res.total ?? 0);
          this.loading.set(false);
        },
        error: (err: HttpErrorResponse) => {
          this.oportunidades.set([]);
          this.total.set(0);
          this.errorMensaje.set(
            this.extraerError(err, 'No se pudieron cargar las oportunidades del CRM.'),
          );
          this.loading.set(false);
        },
      });
  }

  onBuscarTexto(valor: string): void {
    this.filtroTexto.set(valor);
    this.page.set(1);
    this.cargarOportunidades();
  }

  onFiltroEstado(valor: string): void {
    this.filtroEstadoId.set(valor ? Number(valor) : null);
    this.page.set(1);
    this.cargarOportunidades();
  }

  onFiltroPrioridad(valor: string): void {
    this.filtroPrioridad.set(valor as PrioridadOportunidad | '');
    this.page.set(1);
    this.cargarOportunidades();
  }

  onFiltroCampana(valor: string): void {
    this.filtroCampanaId.set(valor ? Number(valor) : null);
    this.page.set(1);
    this.cargarOportunidades();
  }

  limpiarFiltros(): void {
    this.filtroTexto.set('');
    this.filtroEstadoId.set(null);
    this.filtroPrioridad.set('');
    this.filtroCampanaId.set(null);
    this.page.set(1);
    this.cargarOportunidades();
  }

  irAPagina(nueva: number): void {
    if (nueva < 1 || nueva > this.totalPaginas() || nueva === this.page()) return;
    this.page.set(nueva);
    this.cargarOportunidades();
  }

  abrirNueva(): void {
    this.editando.set(null);
    this.errorMensaje.set(null);
    const primera = this.etapas()[0];
    this.form.reset({
      titulo: '',
      nombre_cliente: '',
      lead_id: null,
      campana_id: null,
      estado_oportunidad_id: primera?.id ?? null,
      motivo_descarte_id: null,
      prioridad: 'MEDIA',
      valor_estimado: 0,
      moneda_id: null,
      probabilidad_pct: primera?.probabilidad_pct ?? 0,
      fecha_cierre_estimada: '',
      descripcion: '',
    });
    this.modalAbierto.set(true);
  }

  abrirEditar(op: Oportunidad): void {
    this.editando.set(op);
    this.errorMensaje.set(null);
    this.form.reset({
      titulo: op.titulo,
      nombre_cliente: op.nombre_cliente ?? '',
      lead_id: op.lead_id,
      campana_id: op.campana_id,
      estado_oportunidad_id: op.estado_oportunidad_id,
      motivo_descarte_id: op.motivo_descarte_id,
      prioridad: op.prioridad,
      valor_estimado: op.valor_estimado,
      moneda_id: op.moneda_id,
      probabilidad_pct: op.probabilidad_pct,
      fecha_cierre_estimada: op.fecha_cierre_estimada ?? '',
      descripcion: op.descripcion ?? '',
    });
    this.modalAbierto.set(true);
  }

  cerrarModal(): void {
    if (this.saving()) return;
    this.modalAbierto.set(false);
    this.editando.set(null);
  }

  onCambioEtapaForm(etapaId: number | null): void {
    if (!etapaId) return;
    const etapa = this.etapasMap().get(Number(etapaId));
    if (etapa) {
      this.form.patchValue({ probabilidad_pct: etapa.probabilidad_pct });
    }
  }

  esEtapaPerdidaSeleccionada(): boolean {
    const id = this.form.controls.estado_oportunidad_id.value;
    if (!id) return false;
    return !!this.etapasMap().get(Number(id))?.es_perdida;
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

    if (editando) {
      const payload: ActualizarOportunidadRequest = {
        titulo: raw.titulo!.trim(),
        nombre_cliente: this.normalizarOpcional(raw.nombre_cliente),
        lead_id: raw.lead_id ? Number(raw.lead_id) : undefined,
        campana_id: raw.campana_id ? Number(raw.campana_id) : undefined,
        estado_oportunidad_id: Number(raw.estado_oportunidad_id),
        motivo_descarte_id: raw.motivo_descarte_id ? Number(raw.motivo_descarte_id) : undefined,
        prioridad: raw.prioridad ?? 'MEDIA',
        valor_estimado: Number(raw.valor_estimado ?? 0),
        moneda_id: raw.moneda_id ? Number(raw.moneda_id) : undefined,
        probabilidad_pct: Number(raw.probabilidad_pct ?? 0),
        fecha_cierre_estimada: raw.fecha_cierre_estimada ?? '',
        descripcion: this.normalizarOpcional(raw.descripcion),
      };

      this.crm.actualizarOportunidad(editando.id, payload).subscribe({
        next: () => {
          this.saving.set(false);
          this.modalAbierto.set(false);
          this.editando.set(null);
          this.mostrarExito('Oportunidad actualizada correctamente.');
          this.cargarOportunidades();
        },
        error: (err: HttpErrorResponse) => {
          this.saving.set(false);
          this.errorMensaje.set(this.extraerError(err, 'No se pudo actualizar la oportunidad.'));
        },
      });
      return;
    }

    const payload: CrearOportunidadRequest = {
      empresa_id: this.empresaId,
      sucursal_id: this.usuario()?.sucursal_id ?? null,
      lead_id: raw.lead_id ? Number(raw.lead_id) : null,
      campana_id: raw.campana_id ? Number(raw.campana_id) : null,
      estado_oportunidad_id: Number(raw.estado_oportunidad_id),
      motivo_descarte_id: raw.motivo_descarte_id ? Number(raw.motivo_descarte_id) : null,
      titulo: raw.titulo!.trim(),
      nombre_cliente: this.normalizarOpcional(raw.nombre_cliente),
      valor_estimado: Number(raw.valor_estimado ?? 0),
      moneda_id: raw.moneda_id ? Number(raw.moneda_id) : null,
      probabilidad_pct: Number(raw.probabilidad_pct ?? 0),
      fecha_cierre_estimada: raw.fecha_cierre_estimada?.trim()
        ? raw.fecha_cierre_estimada.trim()
        : undefined,
      prioridad: raw.prioridad ?? 'MEDIA',
      descripcion: this.normalizarOpcional(raw.descripcion),
      created_by: this.usuario()?.id ?? 1,
    };

    this.crm.crearOportunidad(payload).subscribe({
      next: () => {
        this.saving.set(false);
        this.modalAbierto.set(false);
        this.mostrarExito('Oportunidad creada correctamente.');
        this.page.set(1);
        this.cargarOportunidades();
      },
      error: (err: HttpErrorResponse) => {
        this.saving.set(false);
        this.errorMensaje.set(this.extraerError(err, 'No se pudo crear la oportunidad.'));
      },
    });
  }

  cambiarEstado(op: Oportunidad, nuevoEstadoIdStr: string): void {
    const nuevoEstadoId = Number(nuevoEstadoIdStr);
    if (!nuevoEstadoId || nuevoEstadoId === op.estado_oportunidad_id) return;

    this.crm
      .cambiarEstadoOportunidad(op.id, { estado_oportunidad_id: nuevoEstadoId })
      .subscribe({
        next: (actualizada) => {
          this.oportunidades.update((lista) =>
            lista.map((item) => (item.id === actualizada.id ? actualizada : item)),
          );
          this.mostrarExito('Etapa de oportunidad actualizada.');
        },
        error: (err: HttpErrorResponse) => {
          this.errorMensaje.set(
            this.extraerError(err, 'No se pudo cambiar la etapa de la oportunidad.'),
          );
          this.cargarOportunidades();
        },
      });
  }

  solicitarEliminar(op: Oportunidad): void {
    this.aEliminar.set(op);
  }

  cancelarEliminar(): void {
    if (this.saving()) return;
    this.aEliminar.set(null);
  }

  confirmarEliminar(): void {
    const op = this.aEliminar();
    if (!op) return;

    this.saving.set(true);
    this.crm.eliminarOportunidad(op.id).subscribe({
      next: () => {
        this.saving.set(false);
        this.aEliminar.set(null);
        this.mostrarExito('Oportunidad eliminada.');
        this.cargarOportunidades();
      },
      error: (err: HttpErrorResponse) => {
        this.saving.set(false);
        this.aEliminar.set(null);
        this.errorMensaje.set(this.extraerError(err, 'No se pudo eliminar la oportunidad.'));
      },
    });
  }

  nombreEtapa(id: number): string {
    return this.etapasMap().get(id)?.nombre ?? `Etapa #${id}`;
  }

  nombreLead(id: number | null): string | null {
    if (!id) return null;
    const l = this.leadsMap().get(id);
    return l ? `${l.nombre} ${l.apellido}` : `Lead #${id}`;
  }

  nombreCampana(id: number | null): string | null {
    if (!id) return null;
    return this.campanasMap().get(id)?.nombre ?? `Campaña #${id}`;
  }

  etiquetaPrioridad(p: PrioridadOportunidad): string {
    if (p === 'ALTA') return 'Alta';
    if (p === 'BAJA') return 'Baja';
    return 'Media';
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
