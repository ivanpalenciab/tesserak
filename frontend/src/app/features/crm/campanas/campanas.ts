import { Component, OnInit, computed, inject, signal } from '@angular/core';
import { DatePipe, DecimalPipe } from '@angular/common';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { HttpErrorResponse } from '@angular/common/http';
import { catchError, forkJoin, of } from 'rxjs';

import { AuthService } from '../../../core/auth/auth.service';
import { ApiErrorBody } from '../../../core/auth/auth.models';
import { CrmService } from '../crm.service';
import {
  ActualizarCampanaRequest,
  ActualizarGastoCampanaRequest,
  Campana,
  CrearCampanaRequest,
  CrearGastoCampanaRequest,
  EstadoCampana,
  GastoCampana,
  IndicadoresCampana,
  Lead,
  LeadAsociadoCampana,
  Moneda,
  OrigenLead,
  TipoAtribucionCampana,
  TipoCampana,
} from '../crm.models';

@Component({
  selector: 'app-crm-campanas',
  standalone: true,
  imports: [ReactiveFormsModule, DecimalPipe, DatePipe],
  templateUrl: './campanas.html',
  styleUrl: './campanas.scss',
})
export class CrmCampanasComponent implements OnInit {
  private readonly crm = inject(CrmService);
  private readonly auth = inject(AuthService);
  private readonly fb = inject(FormBuilder);

  readonly usuario = this.auth.usuario;

  // Vista principal: campañas vs catálogo de tipos
  readonly vistaPrincipal = signal<'campanas' | 'tipos'>('campanas');

  // Estado general
  readonly loading = signal(true);
  readonly saving = signal(false);
  readonly errorMensaje = signal<string | null>(null);
  readonly exitoMensaje = signal<string | null>(null);

  // Listado de campañas y catálogos
  readonly campanas = signal<Campana[]>([]);
  readonly total = signal(0);
  readonly page = signal(1);
  readonly pageSize = 15;

  readonly tipos = signal<TipoCampana[]>([]);
  readonly origenes = signal<OrigenLead[]>([]);
  readonly monedas = signal<Moneda[]>([]);
  readonly leadsEmpresa = signal<Lead[]>([]);

  // Filtros de campañas
  readonly filtroTexto = signal('');
  readonly filtroEstado = signal<EstadoCampana | ''>('');
  readonly filtroTipoId = signal<number | null>(null);
  readonly filtroOrigenId = signal<number | null>(null);
  readonly filtroFechaDesde = signal('');
  readonly filtroFechaHasta = signal('');

  // Modal crear/editar campaña
  readonly modalCampanaAbierto = signal(false);
  readonly campanaEditando = signal<Campana | null>(null);

  // Modal detalle de campaña
  readonly campanaDetalle = signal<Campana | null>(null);
  readonly tabDetalle = signal<'indicadores' | 'leads' | 'gastos'>('indicadores');
  readonly loadingDetalle = signal(false);
  readonly indicadores = signal<IndicadoresCampana | null>(null);
  readonly leadsCampana = signal<LeadAsociadoCampana[]>([]);
  readonly gastosCampana = signal<GastoCampana[]>([]);
  readonly gastoEditando = signal<GastoCampana | null>(null);

  // Modal crear/editar tipo de campaña
  readonly modalTipoAbierto = signal(false);
  readonly tipoEditando = signal<TipoCampana | null>(null);

  readonly estadosCampana: { value: EstadoCampana; label: string }[] = [
    { value: 'BORRADOR', label: 'Borrador' },
    { value: 'PROGRAMADA', label: 'Programada' },
    { value: 'ACTIVA', label: 'Activa' },
    { value: 'PAUSADA', label: 'Pausada' },
    { value: 'FINALIZADA', label: 'Finalizada' },
  ];

  readonly tiposAtribucion: { value: TipoAtribucionCampana; label: string }[] = [
    { value: 'first_touch', label: 'Primer contacto (first_touch)' },
    { value: 'last_touch', label: 'Último contacto (last_touch)' },
    { value: 'influencia', label: 'Influencia (influencia)' },
  ];

  readonly totalPaginas = computed(() => Math.max(1, Math.ceil(this.total() / this.pageSize)));
  readonly tiposActivos = computed(() => this.tipos().filter((t) => t.activo));
  readonly tiposMap = computed(() => new Map(this.tipos().map((t) => [t.id, t])));
  readonly origenesMap = computed(() => new Map(this.origenes().map((o) => [o.id, o])));
  readonly monedasMap = computed(() => new Map(this.monedas().map((m) => [m.id, m])));

  readonly hayFiltrosActivos = computed(
    () =>
      this.filtroTexto().trim().length > 0 ||
      this.filtroEstado() !== '' ||
      this.filtroTipoId() !== null ||
      this.filtroOrigenId() !== null ||
      this.filtroFechaDesde() !== '' ||
      this.filtroFechaHasta() !== '',
  );

  // Formulario de campaña
  readonly campanaForm = this.fb.group({
    nombre: ['', [Validators.required, Validators.maxLength(150)]],
    tipo_campana_id: [null as number | null, [Validators.required]],
    origen_id: [null as number | null],
    fecha_inicio: ['', [Validators.required]],
    fecha_fin: [''],
    presupuesto: [null as number | null, [Validators.min(0)]],
    moneda_id: [null as number | null, [Validators.min(1)]],
    estado: ['BORRADOR' as EstadoCampana],
    descripcion: ['', [Validators.maxLength(5000)]],
    objetivo: ['', [Validators.maxLength(5000)]],
  });

  // Formulario de asociación de lead a campaña
  readonly asociarLeadForm = this.fb.group({
    lead_id: [null as number | null, [Validators.required]],
    tipo_atribucion: ['first_touch' as TipoAtribucionCampana],
    utm_source: ['', [Validators.maxLength(255)]],
    utm_medium: ['', [Validators.maxLength(255)]],
    utm_campaign: ['', [Validators.maxLength(255)]],
    utm_content: ['', [Validators.maxLength(255)]],
    utm_term: ['', [Validators.maxLength(255)]],
    landing_page: ['', [Validators.maxLength(2000)]],
    external_campaign_id: ['', [Validators.maxLength(150)]],
  });

  // Formulario de gasto de campaña
  readonly gastoForm = this.fb.group({
    fecha: ['', [Validators.required]],
    concepto: ['', [Validators.required, Validators.maxLength(150)]],
    categoria_gasto: ['', [Validators.maxLength(50)]],
    valor: [0, [Validators.required, Validators.min(0)]],
    moneda_id: [1 as number | null, [Validators.required, Validators.min(1)]],
    proveedor_tercero_empresa_rol_id: [null as number | null, [Validators.min(1)]],
  });

  // Formulario de tipo de campaña
  readonly tipoForm = this.fb.group({
    nombre: ['', [Validators.required, Validators.maxLength(80)]],
    descripcion: ['', [Validators.maxLength(255)]],
    activo: [true],
  });

  ngOnInit(): void {
    this.cargarCatalogosYCampanas();
  }

  private get empresaId(): number {
    return this.usuario()?.empresa_id ?? 1;
  }

  cargarCatalogosYCampanas(): void {
    this.loading.set(true);
    this.errorMensaje.set(null);

    forkJoin({
      tipos: this.crm.listarTiposCampana().pipe(catchError(() => of([] as TipoCampana[]))),
      origenes: this.crm
        .listarOrigenes(this.empresaId, true)
        .pipe(catchError(() => of([] as OrigenLead[]))),
      monedas: this.crm.listarMonedas().pipe(catchError(() => of([] as Moneda[]))),
      leads: this.crm
        .listarLeads({ empresa_id: this.empresaId, page: 1, page_size: 100 })
        .pipe(catchError(() => of({ data: [] as Lead[], page: 1, page_size: 100, total: 0 }))),
    }).subscribe(({ tipos, origenes, monedas, leads }) => {
      this.tipos.set(tipos);
      this.origenes.set(origenes);
      this.monedas.set(monedas);
      this.leadsEmpresa.set(leads.data ?? []);
      this.cargarCampanas();
    });
  }

  cargarCampanas(): void {
    this.loading.set(true);
    this.errorMensaje.set(null);

    this.crm
      .listarCampanasPaginado({
        tipo_campana_id: this.filtroTipoId() ?? undefined,
        origen_id: this.filtroOrigenId() ?? undefined,
        estado: this.filtroEstado(),
        fecha_desde: this.filtroFechaDesde() || undefined,
        fecha_hasta: this.filtroFechaHasta() || undefined,
        q: this.filtroTexto(),
        page: this.page(),
        page_size: this.pageSize,
      })
      .subscribe({
        next: (res) => {
          this.campanas.set(res.data ?? []);
          this.total.set(res.total ?? 0);
          this.loading.set(false);
        },
        error: (err: HttpErrorResponse) => {
          this.campanas.set([]);
          this.total.set(0);
          this.errorMensaje.set(this.extraerError(err, 'No se pudieron cargar las campañas.'));
          this.loading.set(false);
        },
      });
  }

  cargarTipos(): void {
    this.crm.listarTiposCampana().subscribe({
      next: (lista) => this.tipos.set(lista),
      error: (err: HttpErrorResponse) =>
        this.errorMensaje.set(this.extraerError(err, 'No se pudieron cargar los tipos de campaña.')),
    });
  }

  // =====================================================================
  // Filtros y paginación
  // =====================================================================

  onBuscarTexto(valor: string): void {
    this.filtroTexto.set(valor);
    this.page.set(1);
    this.cargarCampanas();
  }

  onFiltroEstado(valor: string): void {
    this.filtroEstado.set(valor as EstadoCampana | '');
    this.page.set(1);
    this.cargarCampanas();
  }

  onFiltroTipo(valor: string): void {
    this.filtroTipoId.set(valor ? Number(valor) : null);
    this.page.set(1);
    this.cargarCampanas();
  }

  onFiltroOrigen(valor: string): void {
    this.filtroOrigenId.set(valor ? Number(valor) : null);
    this.page.set(1);
    this.cargarCampanas();
  }

  onFiltroFechaDesde(valor: string): void {
    this.filtroFechaDesde.set(valor);
    this.page.set(1);
    this.cargarCampanas();
  }

  onFiltroFechaHasta(valor: string): void {
    this.filtroFechaHasta.set(valor);
    this.page.set(1);
    this.cargarCampanas();
  }

  limpiarFiltros(): void {
    this.filtroTexto.set('');
    this.filtroEstado.set('');
    this.filtroTipoId.set(null);
    this.filtroOrigenId.set(null);
    this.filtroFechaDesde.set('');
    this.filtroFechaHasta.set('');
    this.page.set(1);
    this.cargarCampanas();
  }

  irAPagina(nueva: number): void {
    if (nueva < 1 || nueva > this.totalPaginas() || nueva === this.page()) return;
    this.page.set(nueva);
    this.cargarCampanas();
  }

  // =====================================================================
  // Crear / Editar Campaña y transiciones de estado
  // =====================================================================

  abrirNuevaCampana(): void {
    this.campanaEditando.set(null);
    this.errorMensaje.set(null);
    const hoy = new Date().toISOString().slice(0, 10);
    this.campanaForm.reset({
      nombre: '',
      tipo_campana_id: this.tiposActivos()[0]?.id ?? null,
      origen_id: null,
      fecha_inicio: hoy,
      fecha_fin: '',
      presupuesto: null,
      moneda_id: null,
      estado: 'BORRADOR',
      descripcion: '',
      objetivo: '',
    });
    this.modalCampanaAbierto.set(true);
  }

  abrirEditarCampana(c: Campana): void {
    if (c.estado === 'FINALIZADA') return;
    this.campanaEditando.set(c);
    this.errorMensaje.set(null);
    this.campanaForm.reset({
      nombre: c.nombre,
      tipo_campana_id: c.tipo_campana_id,
      origen_id: c.origen_id,
      fecha_inicio: c.fecha_inicio,
      fecha_fin: c.fecha_fin ?? '',
      presupuesto: c.presupuesto,
      moneda_id: c.moneda_id,
      estado: c.estado,
      descripcion: c.descripcion ?? '',
      objetivo: c.objetivo ?? '',
    });
    this.modalCampanaAbierto.set(true);
  }

  cerrarModalCampana(): void {
    if (this.saving()) return;
    this.modalCampanaAbierto.set(false);
    this.campanaEditando.set(null);
  }

  guardarCampana(): void {
    if (this.campanaForm.invalid) {
      this.campanaForm.markAllAsTouched();
      return;
    }

    const raw = this.campanaForm.getRawValue();
    if (raw.fecha_fin && raw.fecha_inicio && raw.fecha_fin < raw.fecha_inicio) {
      this.errorMensaje.set('La fecha de finalización no puede ser anterior a la fecha de inicio.');
      return;
    }
    if (raw.presupuesto !== null && raw.presupuesto !== undefined && !raw.moneda_id) {
      this.errorMensaje.set('Cuando defines un presupuesto debes seleccionar una moneda.');
      return;
    }

    const editando = this.campanaEditando();
    this.saving.set(true);
    this.errorMensaje.set(null);

    if (editando) {
      const payload: ActualizarCampanaRequest = {
        nombre: raw.nombre!.trim(),
        tipo_campana_id: Number(raw.tipo_campana_id),
        origen_id: raw.origen_id ? Number(raw.origen_id) : undefined,
        moneda_id: raw.moneda_id ? Number(raw.moneda_id) : undefined,
        fecha_inicio: raw.fecha_inicio!,
        fecha_fin: raw.fecha_fin ?? '',
        presupuesto: raw.presupuesto !== null ? Number(raw.presupuesto) : undefined,
        descripcion: this.normalizarOpcional(raw.descripcion),
        objetivo: this.normalizarOpcional(raw.objetivo),
      };

      this.crm.actualizarCampana(editando.id, payload).subscribe({
        next: (actualizada) => {
          this.saving.set(false);
          this.modalCampanaAbierto.set(false);
          this.campanaEditando.set(null);
          if (this.campanaDetalle()?.id === actualizada.id) {
            this.campanaDetalle.set(actualizada);
            this.cargarDetalleCampana(actualizada.id);
          }
          this.mostrarExito('Campaña actualizada correctamente.');
          this.cargarCampanas();
        },
        error: (err: HttpErrorResponse) => {
          this.saving.set(false);
          this.errorMensaje.set(this.extraerError(err, 'No se pudo actualizar la campaña.'));
        },
      });
      return;
    }

    const payload: CrearCampanaRequest = {
      nombre: raw.nombre!.trim(),
      tipo_campana_id: Number(raw.tipo_campana_id),
      origen_id: raw.origen_id ? Number(raw.origen_id) : null,
      moneda_id: raw.moneda_id ? Number(raw.moneda_id) : null,
      fecha_inicio: raw.fecha_inicio!,
      fecha_fin: raw.fecha_fin?.trim() ? raw.fecha_fin.trim() : undefined,
      presupuesto: raw.presupuesto !== null ? Number(raw.presupuesto) : null,
      estado: raw.estado ?? 'BORRADOR',
      descripcion: this.normalizarOpcional(raw.descripcion),
      objetivo: this.normalizarOpcional(raw.objetivo),
    };

    this.crm.crearCampana(payload).subscribe({
      next: () => {
        this.saving.set(false);
        this.modalCampanaAbierto.set(false);
        this.mostrarExito('Campaña creada correctamente.');
        this.page.set(1);
        this.cargarCampanas();
      },
      error: (err: HttpErrorResponse) => {
        this.saving.set(false);
        this.errorMensaje.set(this.extraerError(err, 'No se pudo crear la campaña.'));
      },
    });
  }

  cambiarEstado(c: Campana, nuevoEstado: EstadoCampana): void {
    this.errorMensaje.set(null);
    this.crm.cambiarEstadoCampana(c.id, nuevoEstado).subscribe({
      next: (actualizada) => {
        this.campanas.update((lista) =>
          lista.map((item) => (item.id === actualizada.id ? actualizada : item)),
        );
        if (this.campanaDetalle()?.id === actualizada.id) {
          this.campanaDetalle.set(actualizada);
          this.cargarDetalleCampana(actualizada.id);
        }
        this.mostrarExito(`Campaña "${actualizada.nombre}" pasó a estado ${this.etiquetaEstado(nuevoEstado)}.`);
      },
      error: (err: HttpErrorResponse) => {
        this.errorMensaje.set(this.extraerError(err, 'No se pudo cambiar el estado de la campaña.'));
      },
    });
  }

  // =====================================================================
  // Detalle de campaña: Indicadores, Leads y Gastos
  // =====================================================================

  abrirDetalle(c: Campana, tab: 'indicadores' | 'leads' | 'gastos' = 'indicadores'): void {
    this.campanaDetalle.set(c);
    this.tabDetalle.set(tab);
    this.gastoEditando.set(null);
    this.errorMensaje.set(null);
    this.resetearFormAsociarLead();
    this.resetearFormGasto(c);
    this.cargarDetalleCampana(c.id);
  }

  cerrarDetalle(): void {
    this.campanaDetalle.set(null);
    this.indicadores.set(null);
    this.leadsCampana.set([]);
    this.gastosCampana.set([]);
    this.gastoEditando.set(null);
  }

  cargarDetalleCampana(campanaId: number): void {
    this.loadingDetalle.set(true);
    forkJoin({
      indicadores: this.crm
        .obtenerIndicadoresCampana(campanaId)
        .pipe(catchError(() => of(null))),
      leads: this.crm
        .listarLeadsDeCampana(campanaId, 1, 50)
        .pipe(catchError(() => of({ data: [] as LeadAsociadoCampana[], page: 1, page_size: 50, total: 0 }))),
      gastos: this.crm
        .listarGastosCampana(campanaId, 1, 50)
        .pipe(catchError(() => of({ data: [] as GastoCampana[], page: 1, page_size: 50, total: 0 }))),
    }).subscribe(({ indicadores, leads, gastos }) => {
      this.indicadores.set(indicadores);
      this.leadsCampana.set(leads.data ?? []);
      this.gastosCampana.set(gastos.data ?? []);
      this.loadingDetalle.set(false);
    });
  }

  puedeAsociarLeads(c: Campana): boolean {
    return c.estado === 'ACTIVA' || c.estado === 'PROGRAMADA' || c.estado === 'PAUSADA';
  }

  asociarLead(): void {
    const campana = this.campanaDetalle();
    if (!campana || this.asociarLeadForm.invalid) {
      this.asociarLeadForm.markAllAsTouched();
      return;
    }

    const raw = this.asociarLeadForm.getRawValue();
    this.saving.set(true);
    this.errorMensaje.set(null);

    this.crm
      .asociarLeadACampana(campana.id, {
        lead_id: Number(raw.lead_id),
        tipo_atribucion: raw.tipo_atribucion ?? 'first_touch',
        utm_source: this.normalizarOpcional(raw.utm_source),
        utm_medium: this.normalizarOpcional(raw.utm_medium),
        utm_campaign: this.normalizarOpcional(raw.utm_campaign),
        utm_content: this.normalizarOpcional(raw.utm_content),
        utm_term: this.normalizarOpcional(raw.utm_term),
        landing_page: this.normalizarOpcional(raw.landing_page),
        external_campaign_id: this.normalizarOpcional(raw.external_campaign_id),
      })
      .subscribe({
        next: () => {
          this.saving.set(false);
          this.resetearFormAsociarLead();
          this.mostrarExito('Lead asociado a la campaña.');
          this.cargarDetalleCampana(campana.id);
        },
        error: (err: HttpErrorResponse) => {
          this.saving.set(false);
          this.errorMensaje.set(this.extraerError(err, 'No se pudo asociar el lead a la campaña.'));
        },
      });
  }

  desasociarLead(leadId: number): void {
    const campana = this.campanaDetalle();
    if (!campana) return;
    this.crm.desasociarLeadDeCampana(campana.id, leadId).subscribe({
      next: () => {
        this.mostrarExito('Asociación eliminada.');
        this.cargarDetalleCampana(campana.id);
      },
      error: (err: HttpErrorResponse) => {
        this.errorMensaje.set(this.extraerError(err, 'No se pudo eliminar la asociación.'));
      },
    });
  }

  editarGasto(g: GastoCampana): void {
    this.gastoEditando.set(g);
    this.gastoForm.reset({
      fecha: g.fecha,
      concepto: g.concepto,
      categoria_gasto: g.categoria_gasto ?? '',
      valor: g.valor,
      moneda_id: g.moneda_id,
      proveedor_tercero_empresa_rol_id: g.proveedor_tercero_empresa_rol_id,
    });
  }

  cancelarEdicionGasto(): void {
    const campana = this.campanaDetalle();
    this.gastoEditando.set(null);
    if (campana) this.resetearFormGasto(campana);
  }

  guardarGasto(): void {
    const campana = this.campanaDetalle();
    if (!campana || this.gastoForm.invalid) {
      this.gastoForm.markAllAsTouched();
      return;
    }

    const raw = this.gastoForm.getRawValue();
    const editando = this.gastoEditando();
    this.saving.set(true);
    this.errorMensaje.set(null);

    if (editando) {
      const payload: ActualizarGastoCampanaRequest = {
        fecha: raw.fecha!,
        concepto: raw.concepto!.trim(),
        categoria_gasto: this.normalizarOpcional(raw.categoria_gasto),
        valor: Number(raw.valor),
        moneda_id: Number(raw.moneda_id),
        proveedor_tercero_empresa_rol_id: raw.proveedor_tercero_empresa_rol_id
          ? Number(raw.proveedor_tercero_empresa_rol_id)
          : undefined,
      };
      this.crm.actualizarGastoCampana(editando.id, payload).subscribe({
        next: () => {
          this.saving.set(false);
          this.gastoEditando.set(null);
          this.resetearFormGasto(campana);
          this.mostrarExito('Gasto actualizado.');
          this.cargarDetalleCampana(campana.id);
        },
        error: (err: HttpErrorResponse) => {
          this.saving.set(false);
          this.errorMensaje.set(this.extraerError(err, 'No se pudo actualizar el gasto.'));
        },
      });
      return;
    }

    const payload: CrearGastoCampanaRequest = {
      fecha: raw.fecha!,
      concepto: raw.concepto!.trim(),
      categoria_gasto: this.normalizarOpcional(raw.categoria_gasto),
      valor: Number(raw.valor),
      moneda_id: Number(raw.moneda_id),
      proveedor_tercero_empresa_rol_id: raw.proveedor_tercero_empresa_rol_id
        ? Number(raw.proveedor_tercero_empresa_rol_id)
        : null,
    };

    this.crm.crearGastoCampana(campana.id, payload).subscribe({
      next: () => {
        this.saving.set(false);
        this.resetearFormGasto(campana);
        this.mostrarExito('Gasto registrado correctamente.');
        this.cargarDetalleCampana(campana.id);
      },
      error: (err: HttpErrorResponse) => {
        this.saving.set(false);
        this.errorMensaje.set(this.extraerError(err, 'No se pudo registrar el gasto.'));
      },
    });
  }

  eliminarGasto(gastoId: number): void {
    const campana = this.campanaDetalle();
    if (!campana) return;
    this.crm.eliminarGastoCampana(gastoId).subscribe({
      next: () => {
        this.mostrarExito('Gasto eliminado.');
        this.cargarDetalleCampana(campana.id);
      },
      error: (err: HttpErrorResponse) => {
        this.errorMensaje.set(this.extraerError(err, 'No se pudo eliminar el gasto.'));
      },
    });
  }

  // =====================================================================
  // Catálogo de Tipos de Campaña
  // =====================================================================

  abrirNuevoTipo(): void {
    this.tipoEditando.set(null);
    this.errorMensaje.set(null);
    this.tipoForm.reset({ nombre: '', descripcion: '', activo: true });
    this.modalTipoAbierto.set(true);
  }

  abrirEditarTipo(t: TipoCampana): void {
    this.tipoEditando.set(t);
    this.errorMensaje.set(null);
    this.tipoForm.reset({
      nombre: t.nombre,
      descripcion: t.descripcion ?? '',
      activo: t.activo,
    });
    this.modalTipoAbierto.set(true);
  }

  cerrarModalTipo(): void {
    if (this.saving()) return;
    this.modalTipoAbierto.set(false);
    this.tipoEditando.set(null);
  }

  guardarTipo(): void {
    if (this.tipoForm.invalid) {
      this.tipoForm.markAllAsTouched();
      return;
    }

    const raw = this.tipoForm.getRawValue();
    const editando = this.tipoEditando();
    this.saving.set(true);
    this.errorMensaje.set(null);

    if (editando) {
      this.crm
        .actualizarTipoCampana(editando.id, {
          nombre: raw.nombre!.trim(),
          descripcion: this.normalizarOpcional(raw.descripcion),
          activo: !!raw.activo,
        })
        .subscribe({
          next: () => {
            this.saving.set(false);
            this.modalTipoAbierto.set(false);
            this.tipoEditando.set(null);
            this.mostrarExito('Tipo de campaña actualizado.');
            this.cargarTipos();
          },
          error: (err: HttpErrorResponse) => {
            this.saving.set(false);
            this.errorMensaje.set(this.extraerError(err, 'No se pudo actualizar el tipo de campaña.'));
          },
        });
      return;
    }

    this.crm
      .crearTipoCampana({
        nombre: raw.nombre!.trim(),
        descripcion: this.normalizarOpcional(raw.descripcion),
        activo: !!raw.activo,
      })
      .subscribe({
        next: () => {
          this.saving.set(false);
          this.modalTipoAbierto.set(false);
          this.mostrarExito('Tipo de campaña creado.');
          this.cargarTipos();
        },
        error: (err: HttpErrorResponse) => {
          this.saving.set(false);
          this.errorMensaje.set(this.extraerError(err, 'No se pudo crear el tipo de campaña.'));
        },
      });
  }

  toggleActivoTipo(t: TipoCampana): void {
    this.crm.actualizarTipoCampana(t.id, { activo: !t.activo }).subscribe({
      next: () => {
        this.mostrarExito(`Tipo "${t.nombre}" ${!t.activo ? 'activado' : 'desactivado'}.`);
        this.cargarTipos();
      },
      error: (err: HttpErrorResponse) => {
        this.errorMensaje.set(this.extraerError(err, 'No se pudo cambiar el estado del tipo.'));
      },
    });
  }

  eliminarTipo(t: TipoCampana): void {
    this.crm.eliminarTipoCampana(t.id).subscribe({
      next: () => {
        this.mostrarExito(`Tipo "${t.nombre}" eliminado.`);
        this.cargarTipos();
      },
      error: (err: HttpErrorResponse) => {
        this.errorMensaje.set(this.extraerError(err, 'No se pudo eliminar el tipo de campaña.'));
      },
    });
  }

  // =====================================================================
  // Helpers de presentación
  // =====================================================================

  nombreTipo(id: number): string {
    return this.tiposMap().get(id)?.nombre ?? `Tipo #${id}`;
  }

  nombreOrigen(id: number | null): string {
    if (!id) return 'Sin origen';
    return this.origenesMap().get(id)?.nombre ?? `Origen #${id}`;
  }

  etiquetaMoneda(id: number | null): string {
    if (!id) return 'Sin moneda';
    const m = this.monedasMap().get(id);
    return m ? `${m.codigo_iso} (${m.simbolo})` : `Moneda #${id}`;
  }

  etiquetaEstado(estado: EstadoCampana): string {
    switch (estado) {
      case 'BORRADOR':
        return 'Borrador';
      case 'PROGRAMADA':
        return 'Programada';
      case 'ACTIVA':
        return 'Activa';
      case 'PAUSADA':
        return 'Pausada';
      case 'FINALIZADA':
        return 'Finalizada';
      default:
        return estado;
    }
  }

  private resetearFormAsociarLead(): void {
    this.asociarLeadForm.reset({
      lead_id: null,
      tipo_atribucion: 'first_touch',
      utm_source: '',
      utm_medium: '',
      utm_campaign: '',
      utm_content: '',
      utm_term: '',
      landing_page: '',
      external_campaign_id: '',
    });
  }

  private resetearFormGasto(c: Campana): void {
    const hoy = new Date().toISOString().slice(0, 10);
    this.gastoForm.reset({
      fecha: hoy,
      concepto: '',
      categoria_gasto: '',
      valor: 0,
      moneda_id: c.moneda_id ?? this.monedas()[0]?.id ?? null,
      proveedor_tercero_empresa_rol_id: null,
    });
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
