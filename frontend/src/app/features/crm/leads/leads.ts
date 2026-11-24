import { Component, OnInit, computed, inject, signal } from '@angular/core';
import { CurrencyPipe, DatePipe } from '@angular/common';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { HttpErrorResponse } from '@angular/common/http';
import { catchError, forkJoin, of, switchMap } from 'rxjs';

import { AuthService } from '../../../core/auth/auth.service';
import { ApiErrorBody } from '../../../core/auth/auth.models';
import { CrmService } from '../crm.service';
import {
  ActualizarLeadRequest,
  Campana,
  CampanaAsociadaLead,
  CrearLeadRequest,
  EstadoLead,
  Lead,
  MotivoDescarteLead,
  OrigenLead,
  TemperaturaLead,
  TipoAtribucionCampana,
} from '../crm.models';

@Component({
  selector: 'app-crm-leads',
  standalone: true,
  imports: [ReactiveFormsModule, CurrencyPipe, DatePipe],
  templateUrl: './leads.html',
  styleUrl: './leads.scss',
})
export class CrmLeadsComponent implements OnInit {
  private readonly crm = inject(CrmService);
  private readonly auth = inject(AuthService);
  private readonly fb = inject(FormBuilder);

  readonly usuario = this.auth.usuario;

  readonly loading = signal(true);
  readonly saving = signal(false);
  readonly errorMensaje = signal<string | null>(null);
  readonly exitoMensaje = signal<string | null>(null);

  readonly leads = signal<Lead[]>([]);
  readonly total = signal(0);
  readonly page = signal(1);
  readonly pageSize = 15;

  readonly origenes = signal<OrigenLead[]>([]);
  readonly estados = signal<EstadoLead[]>([]);
  readonly motivos = signal<MotivoDescarteLead[]>([]);
  readonly campanas = signal<Campana[]>([]);
  readonly campanasDelLead = signal<CampanaAsociadaLead[]>([]);

  readonly campanasHabilitadas = computed(() =>
    this.campanas().filter(
      (c) => c.estado === 'ACTIVA' || c.estado === 'PROGRAMADA' || c.estado === 'PAUSADA',
    ),
  );

  readonly filtroTexto = signal('');
  readonly filtroEstadoId = signal<number | null>(null);
  readonly filtroOrigenId = signal<number | null>(null);
  readonly filtroTemperatura = signal<TemperaturaLead | ''>('');
  readonly filtroConvertido = signal<'todos' | 'activos' | 'convertidos'>('todos');

  readonly modalAbierto = signal(false);
  readonly leadEditando = signal<Lead | null>(null);
  readonly leadAEliminar = signal<Lead | null>(null);

  readonly temperaturas: { value: TemperaturaLead; label: string }[] = [
    { value: 'FRIO', label: 'Frío' },
    { value: 'TIBIO', label: 'Tibio' },
    { value: 'CALIENTE', label: 'Caliente' },
  ];

  readonly tiposAtribucion: { value: TipoAtribucionCampana; label: string }[] = [
    { value: 'first_touch', label: 'Primer contacto (first_touch)' },
    { value: 'last_touch', label: 'Último contacto (last_touch)' },
    { value: 'influencia', label: 'Influencia (influencia)' },
  ];

  readonly totalPaginas = computed(() => Math.max(1, Math.ceil(this.total() / this.pageSize)));

  readonly estadosMap = computed(() => new Map(this.estados().map((e) => [e.id, e])));
  readonly origenesMap = computed(() => new Map(this.origenes().map((o) => [o.id, o])));

  readonly hayFiltrosActivos = computed(
    () =>
      this.filtroTexto().trim().length > 0 ||
      this.filtroEstadoId() !== null ||
      this.filtroOrigenId() !== null ||
      this.filtroTemperatura() !== '' ||
      this.filtroConvertido() !== 'todos',
  );

  readonly leadForm = this.fb.group({
    nombre: ['', [Validators.required, Validators.maxLength(100)]],
    apellido: ['', [Validators.required, Validators.maxLength(100)]],
    nombre_empresa: ['', [Validators.maxLength(150)]],
    cargo: ['', [Validators.maxLength(100)]],
    email: ['', [Validators.email, Validators.maxLength(255)]],
    telefono: ['', [Validators.maxLength(50)]],
    usuario_facebook: ['', [Validators.maxLength(100)]],
    origen_id: [null as number | null, [Validators.required]],
    estado_id: [null as number | null, [Validators.required]],
    temperatura: ['TIBIO' as TemperaturaLead, [Validators.required]],
    score: [0, [Validators.min(0), Validators.max(100)]],
    valor_estimado: [0, [Validators.min(0)]],
    campana_id: [null as number | null],
    tipo_atribucion: ['first_touch' as TipoAtribucionCampana],
    utm_source: ['', [Validators.maxLength(255)]],
    utm_medium: ['', [Validators.maxLength(255)]],
    utm_campaign: ['', [Validators.maxLength(255)]],
  });

  ngOnInit(): void {
    this.cargarCatalogosYLeads();
  }

  private get empresaId(): number {
    return this.usuario()?.empresa_id ?? 1;
  }

  cargarCatalogosYLeads(): void {
    const empresaId = this.empresaId;
    this.loading.set(true);
    this.errorMensaje.set(null);

    forkJoin({
      origenes: this.crm.listarOrigenes(empresaId, true).pipe(catchError(() => of([] as OrigenLead[]))),
      estados: this.crm.listarEstados(empresaId).pipe(catchError(() => of([] as EstadoLead[]))),
      motivos: this.crm
        .listarMotivosDescarte(empresaId, true)
        .pipe(catchError(() => of([] as MotivoDescarteLead[]))),
      campanas: this.crm.listarCampanas().pipe(catchError(() => of([] as Campana[]))),
    }).subscribe(({ origenes, estados, motivos, campanas }) => {
      this.origenes.set(origenes);
      this.estados.set([...estados].sort((a, b) => a.orden - b.orden));
      this.motivos.set(motivos);
      this.campanas.set(campanas);
      this.cargarLeads();
    });
  }

  cargarLeads(): void {
    this.loading.set(true);
    this.errorMensaje.set(null);

    const convertido =
      this.filtroConvertido() === 'todos'
        ? null
        : this.filtroConvertido() === 'convertidos';

    this.crm
      .listarLeads({
        empresa_id: this.empresaId,
        estado_id: this.filtroEstadoId() ?? undefined,
        origen_id: this.filtroOrigenId() ?? undefined,
        temperatura: this.filtroTemperatura(),
        convertido,
        q: this.filtroTexto(),
        page: this.page(),
        page_size: this.pageSize,
      })
      .subscribe({
        next: (res) => {
          this.leads.set(res.data ?? []);
          this.total.set(res.total ?? 0);
          this.loading.set(false);
        },
        error: (err: HttpErrorResponse) => {
          this.leads.set([]);
          this.total.set(0);
          this.errorMensaje.set(this.extraerError(err, 'No se pudieron cargar los leads del CRM.'));
          this.loading.set(false);
        },
      });
  }

  onBuscarTexto(valor: string): void {
    this.filtroTexto.set(valor);
    this.page.set(1);
    this.cargarLeads();
  }

  onFiltroEstado(valor: string): void {
    this.filtroEstadoId.set(valor ? Number(valor) : null);
    this.page.set(1);
    this.cargarLeads();
  }

  onFiltroOrigen(valor: string): void {
    this.filtroOrigenId.set(valor ? Number(valor) : null);
    this.page.set(1);
    this.cargarLeads();
  }

  onFiltroTemperatura(valor: string): void {
    this.filtroTemperatura.set(valor as TemperaturaLead | '');
    this.page.set(1);
    this.cargarLeads();
  }

  onFiltroConvertido(valor: string): void {
    this.filtroConvertido.set(valor as 'todos' | 'activos' | 'convertidos');
    this.page.set(1);
    this.cargarLeads();
  }

  limpiarFiltros(): void {
    this.filtroTexto.set('');
    this.filtroEstadoId.set(null);
    this.filtroOrigenId.set(null);
    this.filtroTemperatura.set('');
    this.filtroConvertido.set('todos');
    this.page.set(1);
    this.cargarLeads();
  }

  irAPagina(nuevaPagina: number): void {
    if (nuevaPagina < 1 || nuevaPagina > this.totalPaginas() || nuevaPagina === this.page()) {
      return;
    }
    this.page.set(nuevaPagina);
    this.cargarLeads();
  }

  abrirNuevoLead(): void {
    this.leadEditando.set(null);
    this.campanasDelLead.set([]);
    this.errorMensaje.set(null);
    this.leadForm.reset({
      nombre: '',
      apellido: '',
      nombre_empresa: '',
      cargo: '',
      email: '',
      telefono: '',
      usuario_facebook: '',
      origen_id: this.origenes()[0]?.id ?? null,
      estado_id: this.estados()[0]?.id ?? null,
      temperatura: 'TIBIO',
      score: 0,
      valor_estimado: 0,
      campana_id: null,
      tipo_atribucion: 'first_touch',
      utm_source: '',
      utm_medium: '',
      utm_campaign: '',
    });
    this.modalAbierto.set(true);
  }

  abrirEditarLead(lead: Lead): void {
    this.leadEditando.set(lead);
    this.campanasDelLead.set([]);
    this.errorMensaje.set(null);
    this.leadForm.reset({
      nombre: lead.nombre,
      apellido: lead.apellido,
      nombre_empresa: lead.nombre_empresa ?? '',
      cargo: lead.cargo ?? '',
      email: lead.email ?? '',
      telefono: lead.telefono ?? '',
      usuario_facebook: lead.usuario_facebook ?? '',
      origen_id: lead.origen_id,
      estado_id: lead.estado_id,
      temperatura: lead.temperatura,
      score: lead.score,
      valor_estimado: lead.valor_estimado,
      campana_id: null,
      tipo_atribucion: 'influencia',
      utm_source: '',
      utm_medium: '',
      utm_campaign: '',
    });
    this.modalAbierto.set(true);
    this.crm
      .listarCampanasDeLead(lead.id)
      .pipe(catchError(() => of([] as CampanaAsociadaLead[])))
      .subscribe((lista) => this.campanasDelLead.set(lista));
  }

  cerrarModal(): void {
    if (this.saving()) return;
    this.modalAbierto.set(false);
    this.leadEditando.set(null);
  }

  guardarLead(): void {
    if (this.leadForm.invalid) {
      this.leadForm.markAllAsTouched();
      return;
    }

    const raw = this.leadForm.getRawValue();
    const editando = this.leadEditando();
    this.saving.set(true);
    this.errorMensaje.set(null);

    if (editando) {
      const payload: ActualizarLeadRequest = {
        nombre: raw.nombre?.trim(),
        apellido: raw.apellido?.trim(),
        nombre_empresa: this.normalizarOpcional(raw.nombre_empresa),
        cargo: this.normalizarOpcional(raw.cargo),
        email: this.normalizarOpcional(raw.email),
        telefono: this.normalizarOpcional(raw.telefono),
        usuario_facebook: this.normalizarOpcional(raw.usuario_facebook),
        origen_id: raw.origen_id ?? undefined,
        estado_id: raw.estado_id ?? undefined,
        temperatura: raw.temperatura ?? 'TIBIO',
        score: Number(raw.score ?? 0),
        valor_estimado: Number(raw.valor_estimado ?? 0),
      };

      this.crm
        .actualizarLead(editando.id, payload)
        .pipe(
          switchMap((actualizado) => {
            if (!raw.campana_id) {
              return of({ actualizado, campanaError: null as string | null });
            }
            return this.crm
              .asociarLeadACampana(Number(raw.campana_id), {
                lead_id: actualizado.id,
                tipo_atribucion: raw.tipo_atribucion ?? 'influencia',
                utm_source: this.normalizarOpcional(raw.utm_source),
                utm_medium: this.normalizarOpcional(raw.utm_medium),
                utm_campaign: this.normalizarOpcional(raw.utm_campaign),
              })
              .pipe(
                switchMap(() => of({ actualizado, campanaError: null as string | null })),
                catchError((err: HttpErrorResponse) =>
                  of({
                    actualizado,
                    campanaError: this.extraerError(
                      err,
                      'El lead se actualizó, pero no se pudo asociar a la campaña.',
                    ),
                  }),
                ),
              );
          }),
        )
        .subscribe({
          next: ({ campanaError }) => {
            this.saving.set(false);
            this.modalAbierto.set(false);
            this.leadEditando.set(null);
            if (campanaError) {
              this.errorMensaje.set(campanaError);
            } else {
              this.mostrarExito('Lead actualizado correctamente.');
            }
            this.cargarLeads();
          },
          error: (err: HttpErrorResponse) => {
            this.saving.set(false);
            this.errorMensaje.set(this.extraerError(err, 'No se pudo actualizar el lead.'));
          },
        });
      return;
    }

    const payload: CrearLeadRequest = {
      empresa_id: this.empresaId,
      sucursal_id: this.usuario()?.sucursal_id ?? null,
      origen_id: Number(raw.origen_id),
      estado_id: Number(raw.estado_id),
      nombre: raw.nombre!.trim(),
      apellido: raw.apellido!.trim(),
      nombre_empresa: this.normalizarOpcional(raw.nombre_empresa),
      cargo: this.normalizarOpcional(raw.cargo),
      email: this.normalizarOpcional(raw.email),
      telefono: this.normalizarOpcional(raw.telefono),
      usuario_facebook: this.normalizarOpcional(raw.usuario_facebook),
      temperatura: raw.temperatura ?? 'TIBIO',
      score: Number(raw.score ?? 0),
      valor_estimado: Number(raw.valor_estimado ?? 0),
      created_by: this.usuario()?.id ?? 1,
    };

    this.crm
      .crearLead(payload)
      .pipe(
        switchMap((creado) => {
          if (!raw.campana_id) {
            return of({ creado, campanaError: null as string | null });
          }
          return this.crm
            .asociarLeadACampana(Number(raw.campana_id), {
              lead_id: creado.id,
              tipo_atribucion: raw.tipo_atribucion ?? 'first_touch',
              utm_source: this.normalizarOpcional(raw.utm_source),
              utm_medium: this.normalizarOpcional(raw.utm_medium),
              utm_campaign: this.normalizarOpcional(raw.utm_campaign),
            })
            .pipe(
              switchMap(() => of({ creado, campanaError: null as string | null })),
              catchError((err: HttpErrorResponse) =>
                of({
                  creado,
                  campanaError: this.extraerError(
                    err,
                    'El lead se creó, pero no se pudo asociar a la campaña seleccionada.',
                  ),
                }),
              ),
            );
        }),
      )
      .subscribe({
        next: ({ campanaError }) => {
          this.saving.set(false);
          this.modalAbierto.set(false);
          if (campanaError) {
            this.errorMensaje.set(campanaError);
          } else if (raw.campana_id) {
            this.mostrarExito('Lead creado y asociado a la campaña correctamente.');
          } else {
            this.mostrarExito('Lead creado correctamente.');
          }
          this.page.set(1);
          this.cargarLeads();
        },
        error: (err: HttpErrorResponse) => {
          this.saving.set(false);
          this.errorMensaje.set(this.extraerError(err, 'No se pudo crear el lead.'));
        },
      });
  }

  cambiarEstado(lead: Lead, nuevoEstadoIdStr: string): void {
    const nuevoEstadoId = Number(nuevoEstadoIdStr);
    if (!nuevoEstadoId || nuevoEstadoId === lead.estado_id) return;

    this.crm.cambiarEstadoLead(lead.id, { estado_id: nuevoEstadoId }).subscribe({
      next: (actualizado) => {
        this.leads.update((lista) =>
          lista.map((item) => (item.id === actualizado.id ? actualizado : item)),
        );
        this.mostrarExito('Etapa del lead actualizada.');
      },
      error: (err: HttpErrorResponse) => {
        this.errorMensaje.set(this.extraerError(err, 'No se pudo cambiar la etapa del lead.'));
        this.cargarLeads();
      },
    });
  }

  convertirLead(lead: Lead): void {
    if (lead.convertido) return;

    this.crm.convertirLead(lead.id).subscribe({
      next: (actualizado) => {
        this.leads.update((lista) =>
          lista.map((item) => (item.id === actualizado.id ? actualizado : item)),
        );
        this.mostrarExito(`Lead "${lead.nombre} ${lead.apellido}" marcado como convertido.`);
      },
      error: (err: HttpErrorResponse) => {
        this.errorMensaje.set(this.extraerError(err, 'No se pudo convertir el lead.'));
      },
    });
  }

  solicitarEliminar(lead: Lead): void {
    this.leadAEliminar.set(lead);
  }

  cancelarEliminar(): void {
    if (this.saving()) return;
    this.leadAEliminar.set(null);
  }

  confirmarEliminar(): void {
    const lead = this.leadAEliminar();
    if (!lead) return;

    this.saving.set(true);
    this.crm.eliminarLead(lead.id).subscribe({
      next: () => {
        this.saving.set(false);
        this.leadAEliminar.set(null);
        this.mostrarExito('Lead eliminado.');
        this.cargarLeads();
      },
      error: (err: HttpErrorResponse) => {
        this.saving.set(false);
        this.leadAEliminar.set(null);
        this.errorMensaje.set(this.extraerError(err, 'No se pudo eliminar el lead.'));
      },
    });
  }

  nombreEstado(estadoId: number): string {
    return this.estadosMap().get(estadoId)?.nombre ?? `Etapa #${estadoId}`;
  }

  nombreOrigen(origenId: number): string {
    return this.origenesMap().get(origenId)?.nombre ?? `Origen #${origenId}`;
  }

  etiquetaTemperatura(temp: TemperaturaLead): string {
    switch (temp) {
      case 'CALIENTE':
        return 'Caliente';
      case 'TIBIO':
        return 'Tibio';
      case 'FRIO':
        return 'Frío';
      default:
        return temp;
    }
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
