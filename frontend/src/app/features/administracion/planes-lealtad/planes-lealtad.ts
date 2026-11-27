import { CommonModule } from '@angular/common';
import { HttpErrorResponse } from '@angular/common/http';
import { Component, OnInit, computed, inject, signal } from '@angular/core';
import { FormBuilder, FormsModule, ReactiveFormsModule, Validators } from '@angular/forms';
import { forkJoin } from 'rxjs';

import {
  ActualizarProgramaPayload,
  ActualizarReglaPayload,
  AdhesionPayload,
  AdhesionPrograma,
  AlcancePrograma,
  ConciliacionProgramaResponse,
  CrearProgramaPayload,
  CrearReglaPayload,
  EstadoPrograma,
  MonedaLealtad,
  ProgramaLealtad,
  ReglaLealtad,
  TipoReglaLealtad,
} from './planes-lealtad.models';
import { PlanesLealtadService } from './planes-lealtad.service';

@Component({
  selector: 'app-planes-lealtad',
  standalone: true,
  imports: [CommonModule, FormsModule, ReactiveFormsModule],
  templateUrl: './planes-lealtad.html',
  styleUrl: './planes-lealtad.scss',
})
export class PlanesLealtadComponent implements OnInit {
  private readonly fb = inject(FormBuilder);
  private readonly lealtadService = inject(PlanesLealtadService);

  readonly loading = signal(true);
  readonly saving = signal(false);
  readonly errorMensaje = signal<string | null>(null);
  readonly exitoMensaje = signal<string | null>(null);
  readonly modalError = signal<string | null>(null);

  readonly monedas = signal<MonedaLealtad[]>([]);
  readonly programas = signal<ProgramaLealtad[]>([]);
  readonly esAdminGlobal = signal(false);

  // Filtros
  readonly filtroSearch = signal('');
  readonly filtroAlcance = signal('');
  readonly filtroEstado = signal('');
  readonly filtroSoloAdheridos = signal(false);

  // Plan destacado por defecto (Lealtad Tesserak)
  readonly planDefecto = computed(() =>
    this.programas().find((p) => p.es_sistema || p.codigo.toUpperCase() === 'LEALTAD_TESSERAK') ?? null,
  );

  // Modal crear/editar plan
  readonly modalPlanAbierto = signal(false);
  readonly planEditando = signal<ProgramaLealtad | null>(null);

  readonly planForm = this.fb.nonNullable.group({
    codigo: ['', [Validators.required, Validators.minLength(2), Validators.maxLength(30)]],
    nombre: ['', [Validators.required, Validators.minLength(2), Validators.maxLength(100)]],
    descripcion: [''],
    alcance: ['LOCAL' as AlcancePrograma, [Validators.required]],
    moneda_referencia_id: [0, [Validators.required, Validators.min(1)]],
    valor_punto_redencion: [1, [Validators.required, Validators.min(0.0001)]],
    dias_vencimiento: [365 as number | null],
    estado: ['ACTIVO' as EstadoPrograma, [Validators.required]],
    fecha_inicio: [''],
    fecha_fin: [''],
  });

  // Modal adhesión / configuración de adhesión
  readonly modalAdhesionAbierto = signal(false);
  readonly planAdhesion = signal<ProgramaLealtad | null>(null);
  readonly editandoAdhesion = signal(false);

  readonly adhesionForm = this.fb.nonNullable.group({
    permite_acumular: [true],
    permite_redimir: [true],
    es_programa_principal: [false],
  });

  // Modal empresas adheridas
  readonly modalEmpresasAbierto = signal(false);
  readonly planEmpresas = signal<ProgramaLealtad | null>(null);
  readonly empresasAdheridas = signal<AdhesionPrograma[]>([]);
  readonly loadingEmpresas = signal(false);

  // Modal reglas de acumulación
  readonly modalReglasAbierto = signal(false);
  readonly planReglas = signal<ProgramaLealtad | null>(null);
  readonly reglas = signal<ReglaLealtad[]>([]);
  readonly loadingReglas = signal(false);
  readonly puedeCrearReglaGlobal = signal(false);
  readonly puedeCrearReglaEmpresa = signal(false);
  readonly mostrandoFormRegla = signal(false);
  readonly reglaEditando = signal<ReglaLealtad | null>(null);

  // Modal Conciliación de coalición y vencimientos
  readonly modalConciliacionAbierto = signal(false);
  readonly planConciliacion = signal<ProgramaLealtad | null>(null);
  readonly conciliacionData = signal<ConciliacionProgramaResponse | null>(null);
  readonly loadingConciliacion = signal(false);
  readonly procesandoVencimientos = signal(false);
  readonly conciliacionDesde = signal('');
  readonly conciliacionHasta = signal('');

  readonly reglaForm = this.fb.nonNullable.group({
    nombre: ['', [Validators.required, Validators.minLength(2), Validators.maxLength(100)]],
    es_global: [true],
    tipo_regla: ['POR_MONTO' as TipoReglaLealtad, [Validators.required]],
    puntos_por_monto: [1, [Validators.required, Validators.min(0.0001)]],
    monto_base: [1000 as number | null, [Validators.min(0.0001)]],
    compra_minima: [0 as number | null, [Validators.min(0)]],
    moneda_id: [0, [Validators.required, Validators.min(1)]],
    prioridad: [100, [Validators.required, Validators.min(1)]],
    fecha_inicio: [''],
    fecha_fin: [''],
    activo: [true],
  });

  ngOnInit(): void {
    this.cargarInicial();
  }

  cargarInicial(): void {
    this.loading.set(true);
    this.errorMensaje.set(null);

    forkJoin({
      monedas: this.lealtadService.listarMonedas(),
      programas: this.lealtadService.listarProgramas(this.construirFiltros()),
    }).subscribe({
      next: ({ monedas, programas }) => {
        this.monedas.set(monedas);
        this.programas.set(programas.data);
        this.esAdminGlobal.set(programas.es_admin_global);
        this.loading.set(false);
      },
      error: (err: HttpErrorResponse) => {
        this.loading.set(false);
        this.errorMensaje.set(this.extraerError(err, 'No se pudieron cargar los planes de lealtad.'));
      },
    });
  }

  cargarProgramas(): void {
    this.loading.set(true);
    this.errorMensaje.set(null);

    this.lealtadService.listarProgramas(this.construirFiltros()).subscribe({
      next: (res) => {
        this.programas.set(res.data);
        this.esAdminGlobal.set(res.es_admin_global);
        this.loading.set(false);
      },
      error: (err: HttpErrorResponse) => {
        this.loading.set(false);
        this.errorMensaje.set(this.extraerError(err, 'No se pudieron consultar los planes de lealtad.'));
      },
    });
  }

  limpiarFiltros(): void {
    this.filtroSearch.set('');
    this.filtroAlcance.set('');
    this.filtroEstado.set('');
    this.filtroSoloAdheridos.set(false);
    this.cargarProgramas();
  }

  // =====================================================================
  // Planes (Crear / Editar / Eliminar)
  // =====================================================================

  abrirCrearPlan(): void {
    this.planEditando.set(null);
    this.modalError.set(null);
    const defaultMonedaId = this.monedas()[0]?.id ?? 0;
    this.planForm.reset({
      codigo: '',
      nombre: '',
      descripcion: '',
      alcance: 'LOCAL',
      moneda_referencia_id: defaultMonedaId,
      valor_punto_redencion: 1,
      dias_vencimiento: 365,
      estado: 'ACTIVO',
      fecha_inicio: this.toDateInputValue(new Date().toISOString()),
      fecha_fin: '',
    });
    this.planForm.controls.codigo.enable();
    this.planForm.controls.alcance.enable();
    this.modalPlanAbierto.set(true);
  }

  abrirEditarPlan(plan: ProgramaLealtad): void {
    this.planEditando.set(plan);
    this.modalError.set(null);
    this.planForm.reset({
      codigo: plan.codigo,
      nombre: plan.nombre,
      descripcion: plan.descripcion ?? '',
      alcance: plan.alcance,
      moneda_referencia_id: plan.moneda_referencia_id,
      valor_punto_redencion: plan.valor_punto_redencion,
      dias_vencimiento: plan.dias_vencimiento ?? null,
      estado: plan.estado,
      fecha_inicio: this.toDateInputValue(plan.fecha_inicio),
      fecha_fin: plan.fecha_fin ? this.toDateInputValue(plan.fecha_fin) : '',
    });
    if (plan.es_sistema) {
      this.planForm.controls.codigo.disable();
    } else {
      this.planForm.controls.codigo.enable();
    }
    this.planForm.controls.alcance.disable();
    this.modalPlanAbierto.set(true);
  }

  cerrarModalPlan(): void {
    this.modalPlanAbierto.set(false);
    this.planEditando.set(null);
    this.modalError.set(null);
  }

  guardarPlan(): void {
    if (this.planForm.invalid) {
      this.planForm.markAllAsTouched();
      return;
    }

    const raw = this.planForm.getRawValue();
    const editando = this.planEditando();
    const diasVenc =
      raw.dias_vencimiento !== null && String(raw.dias_vencimiento).trim() !== ''
        ? Number(raw.dias_vencimiento)
        : null;

    this.saving.set(true);
    this.modalError.set(null);
    this.errorMensaje.set(null);
    this.exitoMensaje.set(null);

    if (!editando) {
      const payload: CrearProgramaPayload = {
        codigo: raw.codigo.trim().toUpperCase(),
        nombre: raw.nombre.trim(),
        descripcion: raw.descripcion.trim() || null,
        alcance: raw.alcance,
        moneda_referencia_id: Number(raw.moneda_referencia_id),
        valor_punto_redencion: Number(raw.valor_punto_redencion),
        dias_vencimiento: diasVenc && diasVenc > 0 ? diasVenc : null,
        estado: raw.estado,
        fecha_inicio: raw.fecha_inicio ? new Date(`${raw.fecha_inicio}T00:00:00Z`).toISOString() : null,
        fecha_fin: raw.fecha_fin ? new Date(`${raw.fecha_fin}T23:59:59Z`).toISOString() : null,
      };

      this.lealtadService.crearPrograma(payload).subscribe({
        next: (creado) => {
          this.saving.set(false);
          this.cerrarModalPlan();
          this.exitoMensaje.set(`Plan de lealtad "${creado.nombre}" creado correctamente.`);
          this.cargarProgramas();
        },
        error: (err: HttpErrorResponse) => {
          this.saving.set(false);
          this.modalError.set(this.extraerError(err, 'No se pudo crear el plan de lealtad.'));
        },
      });
      return;
    }

    const payload: ActualizarProgramaPayload = {
      nombre: raw.nombre.trim(),
      descripcion: raw.descripcion.trim(),
      moneda_referencia_id: Number(raw.moneda_referencia_id),
      valor_punto_redencion: Number(raw.valor_punto_redencion),
      estado: raw.estado,
      fecha_inicio: raw.fecha_inicio ? new Date(`${raw.fecha_inicio}T00:00:00Z`).toISOString() : null,
    };
    if (!editando.es_sistema) {
      payload.codigo = raw.codigo.trim().toUpperCase();
    }
    if (diasVenc && diasVenc > 0) {
      payload.dias_vencimiento = diasVenc;
    } else {
      payload.limpiar_dias_vencimiento = true;
    }
    if (raw.fecha_fin) {
      payload.fecha_fin = new Date(`${raw.fecha_fin}T23:59:59Z`).toISOString();
    } else {
      payload.limpiar_fecha_fin = true;
    }

    this.lealtadService.actualizarPrograma(editando.id, payload).subscribe({
      next: (actualizado) => {
        this.saving.set(false);
        this.cerrarModalPlan();
        this.exitoMensaje.set(`Plan de lealtad "${actualizado.nombre}" actualizado correctamente.`);
        this.cargarProgramas();
      },
      error: (err: HttpErrorResponse) => {
        this.saving.set(false);
        this.modalError.set(this.extraerError(err, 'No se pudo actualizar el plan de lealtad.'));
      },
    });
  }

  eliminarPlan(plan: ProgramaLealtad): void {
    if (!plan.puede_eliminar) return;
    if (!window.confirm(`¿Deseas eliminar el plan de lealtad "${plan.nombre}" (${plan.codigo})?`)) {
      return;
    }
    this.errorMensaje.set(null);
    this.exitoMensaje.set(null);

    this.lealtadService.eliminarPrograma(plan.id).subscribe({
      next: () => {
        this.exitoMensaje.set(`Plan de lealtad "${plan.nombre}" eliminado correctamente.`);
        this.cargarProgramas();
      },
      error: (err: HttpErrorResponse) => {
        this.errorMensaje.set(this.extraerError(err, 'No se pudo eliminar el plan de lealtad.'));
      },
    });
  }

  // =====================================================================
  // Adhesiones
  // =====================================================================

  abrirModalAdhesion(plan: ProgramaLealtad, editar = false): void {
    this.planAdhesion.set(plan);
    this.editandoAdhesion.set(editar);
    this.modalError.set(null);
    const adh = plan.adhesion;
    this.adhesionForm.reset({
      permite_acumular: editar && adh ? adh.permite_acumular : true,
      permite_redimir: editar && adh ? adh.permite_redimir : true,
      es_programa_principal: editar && adh ? adh.es_programa_principal : false,
    });
    this.modalAdhesionAbierto.set(true);
  }

  cerrarModalAdhesion(): void {
    this.modalAdhesionAbierto.set(false);
    this.planAdhesion.set(null);
    this.modalError.set(null);
  }

  guardarAdhesion(): void {
    const plan = this.planAdhesion();
    if (!plan) return;

    const payload: AdhesionPayload = this.adhesionForm.getRawValue();
    this.saving.set(true);
    this.modalError.set(null);
    this.errorMensaje.set(null);
    this.exitoMensaje.set(null);

    const obs = this.editandoAdhesion()
      ? this.lealtadService.actualizarAdhesion(plan.id, payload)
      : this.lealtadService.adherirEmpresa(plan.id, payload);

    obs.subscribe({
      next: () => {
        this.saving.set(false);
        this.cerrarModalAdhesion();
        this.exitoMensaje.set(
          this.editandoAdhesion()
            ? `Configuración de adhesión a "${plan.nombre}" actualizada.`
            : `Tu empresa se adhirió correctamente al plan "${plan.nombre}".`,
        );
        this.cargarProgramas();
      },
      error: (err: HttpErrorResponse) => {
        this.saving.set(false);
        this.modalError.set(this.extraerError(err, 'No se pudo guardar la adhesión al plan.'));
      },
    });
  }

  desadherirEmpresa(plan: ProgramaLealtad): void {
    if (!window.confirm(`¿Deseas retirar a tu empresa del plan "${plan.nombre}"?`)) {
      return;
    }
    this.errorMensaje.set(null);
    this.exitoMensaje.set(null);

    this.lealtadService.desadherirEmpresa(plan.id).subscribe({
      next: () => {
        this.exitoMensaje.set(`Tu empresa fue retirada del plan "${plan.nombre}".`);
        this.cargarProgramas();
      },
      error: (err: HttpErrorResponse) => {
        this.errorMensaje.set(this.extraerError(err, 'No se pudo retirar la adhesión al plan.'));
      },
    });
  }

  // =====================================================================
  // Empresas adheridas
  // =====================================================================

  abrirEmpresasAdheridas(plan: ProgramaLealtad): void {
    this.planEmpresas.set(plan);
    this.empresasAdheridas.set([]);
    this.modalError.set(null);
    this.loadingEmpresas.set(true);
    this.modalEmpresasAbierto.set(true);

    this.lealtadService.listarEmpresasAdheridas(plan.id).subscribe({
      next: (rows) => {
        this.empresasAdheridas.set(rows);
        this.loadingEmpresas.set(false);
      },
      error: (err: HttpErrorResponse) => {
        this.loadingEmpresas.set(false);
        this.modalError.set(this.extraerError(err, 'No se pudieron consultar las empresas adheridas.'));
      },
    });
  }

  cerrarModalEmpresas(): void {
    this.modalEmpresasAbierto.set(false);
    this.planEmpresas.set(null);
    this.empresasAdheridas.set([]);
    this.modalError.set(null);
  }

  // =====================================================================
  // Reglas de acumulación (lealtad_reglas)
  // =====================================================================

  abrirModalReglas(plan: ProgramaLealtad): void {
    this.planReglas.set(plan);
    this.reglas.set([]);
    this.mostrandoFormRegla.set(false);
    this.reglaEditando.set(null);
    this.modalError.set(null);
    this.modalReglasAbierto.set(true);
    this.cargarReglas(plan.id);
  }

  cerrarModalReglas(): void {
    this.modalReglasAbierto.set(false);
    this.planReglas.set(null);
    this.reglas.set([]);
    this.mostrandoFormRegla.set(false);
    this.reglaEditando.set(null);
    this.modalError.set(null);
  }

  cargarReglas(programaId: number): void {
    this.loadingReglas.set(true);
    this.modalError.set(null);

    this.lealtadService.listarReglas(programaId).subscribe({
      next: (res) => {
        this.reglas.set(res.data);
        this.puedeCrearReglaGlobal.set(res.puede_crear_global);
        this.puedeCrearReglaEmpresa.set(res.puede_crear_empresa);
        this.loadingReglas.set(false);
      },
      error: (err: HttpErrorResponse) => {
        this.loadingReglas.set(false);
        this.modalError.set(this.extraerError(err, 'No se pudieron cargar las reglas del plan.'));
      },
    });
  }

  abrirCrearRegla(): void {
    const plan = this.planReglas();
    if (!plan) return;
    this.reglaEditando.set(null);
    this.modalError.set(null);
    const defaultGlobal = this.puedeCrearReglaGlobal();
    this.reglaForm.reset({
      nombre: '',
      es_global: defaultGlobal,
      tipo_regla: 'POR_MONTO',
      puntos_por_monto: 1,
      monto_base: 1000,
      compra_minima: 0,
      moneda_id: plan.moneda_referencia_id || (this.monedas()[0]?.id ?? 0),
      prioridad: 100,
      fecha_inicio: this.toDateInputValue(new Date().toISOString()),
      fecha_fin: '',
      activo: true,
    });
    if (this.puedeCrearReglaGlobal() && this.puedeCrearReglaEmpresa()) {
      this.reglaForm.controls.es_global.enable();
    } else {
      this.reglaForm.controls.es_global.disable();
    }
    this.mostrandoFormRegla.set(true);
  }

  abrirEditarRegla(regla: ReglaLealtad): void {
    this.reglaEditando.set(regla);
    this.modalError.set(null);
    this.reglaForm.reset({
      nombre: regla.nombre,
      es_global: regla.es_global,
      tipo_regla: regla.tipo_regla,
      puntos_por_monto: regla.puntos_por_monto,
      monto_base: regla.monto_base,
      compra_minima: regla.compra_minima,
      moneda_id: regla.moneda_id,
      prioridad: regla.prioridad,
      fecha_inicio: this.toDateInputValue(regla.fecha_inicio),
      fecha_fin: regla.fecha_fin ? this.toDateInputValue(regla.fecha_fin) : '',
      activo: regla.activo,
    });
    this.reglaForm.controls.es_global.disable();
    this.mostrandoFormRegla.set(true);
  }

  cancelarFormRegla(): void {
    this.mostrandoFormRegla.set(false);
    this.reglaEditando.set(null);
    this.modalError.set(null);
  }

  guardarRegla(): void {
    const plan = this.planReglas();
    if (!plan) return;
    if (this.reglaForm.invalid) {
      this.reglaForm.markAllAsTouched();
      this.modalError.set('Revisa los campos obligatorios de la regla antes de guardar.');
      return;
    }

    const raw = this.reglaForm.getRawValue();
    if (raw.tipo_regla === 'POR_MONTO' && (!raw.monto_base || Number(raw.monto_base) <= 0)) {
      this.modalError.set('El monto base debe ser mayor a 0 para reglas por monto.');
      return;
    }

    const editando = this.reglaEditando();
    const compraMinima =
      raw.compra_minima !== null && String(raw.compra_minima).trim() !== ''
        ? Number(raw.compra_minima)
        : 0;

    this.saving.set(true);
    this.modalError.set(null);

    if (!editando) {
      const payload: CrearReglaPayload = {
        nombre: raw.nombre.trim(),
        es_global: Boolean(raw.es_global),
        tipo_regla: raw.tipo_regla,
        puntos_por_monto: Number(raw.puntos_por_monto),
        monto_base: raw.tipo_regla === 'POR_MONTO' ? Number(raw.monto_base) : 1,
        compra_minima: compraMinima,
        moneda_id: Number(raw.moneda_id),
        prioridad: Number(raw.prioridad),
        fecha_inicio: raw.fecha_inicio ? new Date(`${raw.fecha_inicio}T00:00:00Z`).toISOString() : null,
        fecha_fin: raw.fecha_fin ? new Date(`${raw.fecha_fin}T23:59:59Z`).toISOString() : null,
        activo: Boolean(raw.activo),
      };

      this.lealtadService.crearRegla(plan.id, payload).subscribe({
        next: () => {
          this.saving.set(false);
          this.mostrandoFormRegla.set(false);
          this.exitoMensaje.set(`Regla "${payload.nombre}" creada correctamente.`);
          this.cargarReglas(plan.id);
        },
        error: (err: HttpErrorResponse) => {
          this.saving.set(false);
          this.modalError.set(this.extraerError(err, 'No se pudo crear la regla de acumulación.'));
        },
      });
      return;
    }

    const payload: ActualizarReglaPayload = {
      nombre: raw.nombre.trim(),
      tipo_regla: raw.tipo_regla,
      puntos_por_monto: Number(raw.puntos_por_monto),
      monto_base: raw.tipo_regla === 'POR_MONTO' ? Number(raw.monto_base) : 1,
      compra_minima: compraMinima,
      moneda_id: Number(raw.moneda_id),
      prioridad: Number(raw.prioridad),
      fecha_inicio: raw.fecha_inicio ? new Date(`${raw.fecha_inicio}T00:00:00Z`).toISOString() : null,
      activo: Boolean(raw.activo),
    };
    if (raw.fecha_fin) {
      payload.fecha_fin = new Date(`${raw.fecha_fin}T23:59:59Z`).toISOString();
    } else {
      payload.limpiar_fecha_fin = true;
    }

    this.lealtadService.actualizarRegla(plan.id, editando.id, payload).subscribe({
      next: () => {
        this.saving.set(false);
        this.mostrandoFormRegla.set(false);
        this.reglaEditando.set(null);
        this.exitoMensaje.set(`Regla "${payload.nombre}" actualizada correctamente.`);
        this.cargarReglas(plan.id);
      },
      error: (err: HttpErrorResponse) => {
        this.saving.set(false);
        this.modalError.set(this.extraerError(err, 'No se pudo actualizar la regla.'));
      },
    });
  }

  eliminarRegla(regla: ReglaLealtad): void {
    const plan = this.planReglas();
    if (!plan || !regla.puede_eliminar) return;
    if (!window.confirm(`¿Deseas eliminar la regla "${regla.nombre}"?`)) {
      return;
    }
    this.modalError.set(null);

    this.lealtadService.eliminarRegla(plan.id, regla.id).subscribe({
      next: () => {
        this.exitoMensaje.set(`Regla "${regla.nombre}" eliminada correctamente.`);
        this.cargarReglas(plan.id);
      },
      error: (err: HttpErrorResponse) => {
        this.modalError.set(this.extraerError(err, 'No se pudo eliminar la regla.'));
      },
    });
  }

  etiquetaTipoRegla(tipo: TipoReglaLealtad): string {
    switch (tipo) {
      case 'BONO_FIJO':
        return 'Bono fijo';
      case 'MULTIPLICADOR':
        return 'Multiplicador';
      default:
        return 'Por monto';
    }
  }

  resumenFormulaRegla(r: ReglaLealtad): string {
    const sym = r.moneda_simbolo || '$';
    const cod = r.moneda_codigo || '';
    if (r.tipo_regla === 'BONO_FIJO') {
      return `+${r.puntos_por_monto} pts fijos por venta`;
    }
    if (r.tipo_regla === 'MULTIPLICADOR') {
      return `x${r.puntos_por_monto} sobre puntos base`;
    }
    return `${r.puntos_por_monto} pts por cada ${sym} ${r.monto_base.toLocaleString('es-CO')} ${cod}`.trim();
  }

  etiquetaAlcance(alcance: AlcancePrograma): string {
    switch (alcance) {
      case 'GLOBAL':
        return 'Global';
      case 'COALICION':
        return 'Coalición';
      default:
        return 'Local';
    }
  }

  etiquetaEstado(estado: EstadoPrograma): string {
    switch (estado) {
      case 'ACTIVO':
        return 'Activo';
      case 'PAUSADO':
        return 'Pausado';
      default:
        return 'Inactivo';
    }
  }

  // =====================================================================
  // Conciliación de coalición y procesamiento de vencimientos
  // =====================================================================

  abrirConciliacion(plan: ProgramaLealtad): void {
    this.planConciliacion.set(plan);
    this.conciliacionData.set(null);
    this.conciliacionDesde.set('');
    this.conciliacionHasta.set('');
    this.modalError.set(null);
    this.modalConciliacionAbierto.set(true);
    this.cargarConciliacion();
  }

  cerrarConciliacion(): void {
    this.modalConciliacionAbierto.set(false);
    this.planConciliacion.set(null);
    this.conciliacionData.set(null);
    this.modalError.set(null);
  }

  cargarConciliacion(): void {
    const plan = this.planConciliacion();
    if (!plan) return;
    this.loadingConciliacion.set(true);
    this.modalError.set(null);

    this.lealtadService
      .obtenerConciliacion(plan.id, this.conciliacionDesde(), this.conciliacionHasta())
      .subscribe({
        next: (data) => {
          this.conciliacionData.set(data);
          this.loadingConciliacion.set(false);
        },
        error: (err: HttpErrorResponse) => {
          this.loadingConciliacion.set(false);
          this.modalError.set(this.extraerError(err, 'No se pudo cargar el reporte de conciliación.'));
        },
      });
  }

  ejecutarVencimientos(plan: ProgramaLealtad): void {
    if (
      !window.confirm(
        `¿Deseas procesar los puntos vencidos del plan "${plan.nombre}"? Esta acción descontará los lotes expirados en el libro mayor.`,
      )
    ) {
      return;
    }
    this.procesandoVencimientos.set(true);
    this.modalError.set(null);
    this.errorMensaje.set(null);
    this.exitoMensaje.set(null);

    this.lealtadService.procesarVencimientos(plan.id).subscribe({
      next: (res) => {
        this.procesandoVencimientos.set(false);
        this.exitoMensaje.set(
          res.puntos_vencidos_total > 0
            ? `Vencimientos procesados en "${plan.nombre}": ${res.puntos_vencidos_total} pts vencidos en ${res.cuentas_afectadas} cuenta(s).`
            : `No se encontraron puntos vencidos pendientes en "${plan.nombre}".`,
        );
        if (this.modalConciliacionAbierto() && this.planConciliacion()?.id === plan.id) {
          this.cargarConciliacion();
        }
      },
      error: (err: HttpErrorResponse) => {
        this.procesandoVencimientos.set(false);
        const msg = this.extraerError(err, 'No se pudieron procesar los vencimientos.');
        if (this.modalConciliacionAbierto()) {
          this.modalError.set(msg);
        } else {
          this.errorMensaje.set(msg);
        }
      },
    });
  }

  private construirFiltros() {
    return {
      search: this.filtroSearch(),
      alcance: this.filtroAlcance(),
      estado: this.filtroEstado(),
      solo_adheridos: this.filtroSoloAdheridos(),
    };
  }

  private toDateInputValue(iso: string): string {
    if (!iso) return '';
    return iso.slice(0, 10);
  }

  private extraerError(err: HttpErrorResponse, fallback: string): string {
    const body = err.error as { error?: { message?: string } } | null;
    return body?.error?.message ?? fallback;
  }
}
