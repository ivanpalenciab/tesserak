import { CommonModule } from '@angular/common';
import { HttpErrorResponse } from '@angular/common/http';
import { Component, OnInit, computed, inject, signal } from '@angular/core';
import { FormBuilder, FormsModule, ReactiveFormsModule, Validators } from '@angular/forms';
import { RouterLink } from '@angular/router';
import { forkJoin } from 'rxjs';

import { AuthService } from '../../../core/auth/auth.service';
import {
  ClienteDisponibleLealtad,
  CuentaLealtad,
  EstadoCuentaLealtad,
  PaisLealtad,
  ProgramaOperacionLealtad,
  TipoIdentificacionLealtad,
  TipoTransaccionLealtad,
  TransaccionLealtad,
} from '../lealtad.models';
import { LealtadOperacionService } from '../lealtad.service';

@Component({
  selector: 'app-lealtad-miembros-cuentas',
  standalone: true,
  imports: [CommonModule, FormsModule, ReactiveFormsModule, RouterLink],
  templateUrl: './miembros-cuentas.html',
  styleUrl: './miembros-cuentas.scss',
})
export class LealtadMiembrosCuentasComponent implements OnInit {
  private readonly fb = inject(FormBuilder);
  private readonly lealtadService = inject(LealtadOperacionService);
  private readonly auth = inject(AuthService);

  readonly esAdministrador = this.auth.esAdministrador;

  readonly loading = signal(true);
  readonly saving = signal(false);
  readonly errorMensaje = signal<string | null>(null);
  readonly exitoMensaje = signal<string | null>(null);
  readonly modalError = signal<string | null>(null);

  readonly paises = signal<PaisLealtad[]>([]);
  readonly programas = signal<ProgramaOperacionLealtad[]>([]);
  readonly cuentas = signal<CuentaLealtad[]>([]);

  // Filtros
  readonly filtroSearch = signal('');
  readonly filtroProgramaId = signal<number | null>(null);
  readonly filtroEstado = signal('');

  // Resumen KPI
  readonly totalCuentas = computed(() => this.cuentas().length);
  readonly cuentasActivas = computed(() => this.cuentas().filter((c) => c.estado === 'ACTIVA').length);
  readonly totalPuntos = computed(() => this.cuentas().reduce((acc, c) => acc + c.saldo_puntos, 0));

  readonly tiposDocumento: TipoIdentificacionLealtad[] = [
    'CC',
    'NIT',
    'CE',
    'PASAPORTE',
    'RUT',
    'DNI',
    'OTRO',
  ];

  // Modal registrar miembro + abrir cuenta
  readonly modalRegistroAbierto = signal(false);
  readonly buscandoDoc = signal(false);
  readonly miembroEncontradoAviso = signal<string | null>(null);

  readonly registroForm = this.fb.nonNullable.group({
    programa_id: [0, [Validators.required, Validators.min(1)]],
    pais_id: [0, [Validators.required, Validators.min(1)]],
    tipo_identificacion: ['CC' as TipoIdentificacionLealtad, [Validators.required]],
    numero_identificacion: ['', [Validators.required, Validators.minLength(2), Validators.maxLength(50)]],
    nombre_completo: ['', [Validators.required, Validators.minLength(2), Validators.maxLength(200)]],
    email: ['', [Validators.email]],
    telefono: [''],
  });

  // Modal inscribir desde Terceros (Clientes)
  readonly modalTerceroAbierto = signal(false);
  readonly clientes = signal<ClienteDisponibleLealtad[]>([]);
  readonly loadingClientes = signal(false);
  readonly filtroClienteSearch = signal('');

  readonly terceroForm = this.fb.nonNullable.group({
    programa_id: [0, [Validators.required, Validators.min(1)]],
    tercero_id: [0, [Validators.required, Validators.min(1)]],
    pais_id: [0],
  });

  // Modal editar datos de contacto del miembro
  readonly modalEditarMiembroAbierto = signal(false);
  readonly cuentaEditandoMiembro = signal<CuentaLealtad | null>(null);

  readonly editarMiembroForm = this.fb.nonNullable.group({
    nombre_completo: ['', [Validators.required, Validators.minLength(2), Validators.maxLength(200)]],
    email: ['', [Validators.email]],
    telefono: [''],
  });

  // Modal Movimientos (Libro mayor)
  readonly modalMovimientosAbierto = signal(false);
  readonly cuentaMovimientos = signal<CuentaLealtad | null>(null);
  readonly transacciones = signal<TransaccionLealtad[]>([]);
  readonly loadingMovimientos = signal(false);

  // Modal Acumular / Ajustar puntos
  readonly modalAcumularAbierto = signal(false);
  readonly cuentaAcumular = signal<CuentaLealtad | null>(null);
  readonly modoAcumular = signal<'MONTO' | 'PUNTOS'>('MONTO');

  readonly acumularForm = this.fb.nonNullable.group({
    monto_compra: [null as number | null],
    puntos: [null as number | null],
    observaciones: [''],
  });

  // Modal Redimir puntos
  readonly modalRedimirAbierto = signal(false);
  readonly cuentaRedimir = signal<CuentaLealtad | null>(null);

  readonly redimirForm = this.fb.nonNullable.group({
    puntos_redimidos: [1, [Validators.required, Validators.min(1)]],
    observaciones: [''],
  });

  ngOnInit(): void {
    this.cargarInicial();
  }

  cargarInicial(): void {
    this.loading.set(true);
    this.errorMensaje.set(null);

    forkJoin({
      paises: this.lealtadService.listarPaises(),
      programas: this.lealtadService.listarProgramasAdheridos(),
      cuentas: this.lealtadService.listarCuentas(this.construirFiltros()),
    }).subscribe({
      next: ({ paises, programas, cuentas }) => {
        this.paises.set(paises);
        this.programas.set(programas);
        this.cuentas.set(cuentas);
        this.loading.set(false);
      },
      error: (err: HttpErrorResponse) => {
        this.loading.set(false);
        this.errorMensaje.set(this.extraerError(err, 'No se pudo cargar el módulo de lealtad.'));
      },
    });
  }

  cargarCuentas(): void {
    this.loading.set(true);
    this.errorMensaje.set(null);

    this.lealtadService.listarCuentas(this.construirFiltros()).subscribe({
      next: (rows) => {
        this.cuentas.set(rows);
        this.loading.set(false);
      },
      error: (err: HttpErrorResponse) => {
        this.loading.set(false);
        this.errorMensaje.set(this.extraerError(err, 'No se pudieron consultar las cuentas de lealtad.'));
      },
    });
  }

  limpiarFiltros(): void {
    this.filtroSearch.set('');
    this.filtroProgramaId.set(null);
    this.filtroEstado.set('');
    this.cargarCuentas();
  }

  // =====================================================================
  // Registrar miembro manualmente (con búsqueda global por documento)
  // =====================================================================

  abrirRegistroMiembro(): void {
    this.modalError.set(null);
    this.miembroEncontradoAviso.set(null);
    const defaultProg = this.programas().find((p) => p.es_programa_principal)?.id ?? this.programas()[0]?.id ?? 0;
    const defaultPais = this.paises()[0]?.id ?? 0;

    this.registroForm.reset({
      programa_id: defaultProg,
      pais_id: defaultPais,
      tipo_identificacion: 'CC',
      numero_identificacion: '',
      nombre_completo: '',
      email: '',
      telefono: '',
    });
    this.modalRegistroAbierto.set(true);
  }

  cerrarRegistroMiembro(): void {
    this.modalRegistroAbierto.set(false);
    this.modalError.set(null);
    this.miembroEncontradoAviso.set(null);
  }

  buscarDocumentoGlobal(): void {
    const raw = this.registroForm.getRawValue();
    if (!raw.pais_id || !raw.numero_identificacion.trim()) {
      this.modalError.set('Selecciona país e ingresa el número de documento para buscar.');
      return;
    }
    this.buscandoDoc.set(true);
    this.modalError.set(null);
    this.miembroEncontradoAviso.set(null);

    this.lealtadService
      .buscarMiembroPorDocumento(raw.pais_id, raw.tipo_identificacion, raw.numero_identificacion)
      .subscribe({
        next: (m) => {
          this.buscandoDoc.set(false);
          this.registroForm.patchValue({
            nombre_completo: m.nombre_completo,
            email: m.email ?? '',
            telefono: m.telefono ?? '',
          });
          this.miembroEncontradoAviso.set(
            `Miembro encontrado en la red de lealtad: ${m.nombre_completo}. Al confirmar se vinculará su cuenta al plan seleccionado.`,
          );
        },
        error: () => {
          this.buscandoDoc.set(false);
          this.miembroEncontradoAviso.set('No existe aún en la red de lealtad. Completa sus datos para registrarlo.');
        },
      });
  }

  guardarRegistroMiembro(): void {
    if (this.registroForm.invalid) {
      this.registroForm.markAllAsTouched();
      this.modalError.set('Revisa los campos obligatorios antes de continuar.');
      return;
    }
    const raw = this.registroForm.getRawValue();
    this.saving.set(true);
    this.modalError.set(null);
    this.exitoMensaje.set(null);

    this.lealtadService
      .registrarMiembro({
        programa_id: Number(raw.programa_id),
        pais_id: Number(raw.pais_id),
        tipo_identificacion: raw.tipo_identificacion,
        numero_identificacion: raw.numero_identificacion.trim(),
        nombre_completo: raw.nombre_completo.trim(),
        email: raw.email.trim() || null,
        telefono: raw.telefono.trim() || null,
      })
      .subscribe({
        next: (cuenta) => {
          this.saving.set(false);
          this.cerrarRegistroMiembro();
          this.exitoMensaje.set(
            `Miembro "${cuenta.miembro_nombre}" inscrito en ${cuenta.programa_nombre} (saldo actual: ${cuenta.saldo_puntos} pts).`,
          );
          this.cargarCuentas();
        },
        error: (err: HttpErrorResponse) => {
          this.saving.set(false);
          this.modalError.set(this.extraerError(err, 'No se pudo registrar el miembro en el plan.'));
        },
      });
  }

  // =====================================================================
  // Inscribir desde Terceros (Clientes)
  // =====================================================================

  abrirInscribirDesdeTercero(): void {
    this.modalError.set(null);
    this.filtroClienteSearch.set('');
    const defaultProg = this.programas().find((p) => p.es_programa_principal)?.id ?? this.programas()[0]?.id ?? 0;
    const defaultPais = this.paises()[0]?.id ?? 0;

    this.terceroForm.reset({
      programa_id: defaultProg,
      tercero_id: 0,
      pais_id: defaultPais,
    });
    this.modalTerceroAbierto.set(true);
    this.buscarClientesTerceros();
  }

  cerrarModalTercero(): void {
    this.modalTerceroAbierto.set(false);
    this.modalError.set(null);
  }

  buscarClientesTerceros(): void {
    this.loadingClientes.set(true);
    this.lealtadService.listarClientesDisponibles(this.filtroClienteSearch()).subscribe({
      next: (rows) => {
        this.clientes.set(rows);
        this.loadingClientes.set(false);
      },
      error: (err: HttpErrorResponse) => {
        this.loadingClientes.set(false);
        this.modalError.set(this.extraerError(err, 'No se pudieron consultar los clientes de Terceros.'));
      },
    });
  }

  guardarInscripcionTercero(): void {
    if (this.terceroForm.invalid) {
      this.terceroForm.markAllAsTouched();
      this.modalError.set('Selecciona un plan y un cliente de Terceros.');
      return;
    }
    const raw = this.terceroForm.getRawValue();
    this.saving.set(true);
    this.modalError.set(null);
    this.exitoMensaje.set(null);

    this.lealtadService
      .inscribirDesdeTercero({
        programa_id: Number(raw.programa_id),
        tercero_id: Number(raw.tercero_id),
        pais_id: raw.pais_id ? Number(raw.pais_id) : null,
      })
      .subscribe({
        next: (cuenta) => {
          this.saving.set(false);
          this.cerrarModalTercero();
          this.exitoMensaje.set(
            `Cliente "${cuenta.miembro_nombre}" inscrito correctamente en ${cuenta.programa_nombre}.`,
          );
          this.cargarCuentas();
        },
        error: (err: HttpErrorResponse) => {
          this.saving.set(false);
          this.modalError.set(this.extraerError(err, 'No se pudo inscribir el cliente en el plan.'));
        },
      });
  }

  // =====================================================================
  // Editar miembro y cambiar estado de cuenta
  // =====================================================================

  abrirEditarMiembro(cuenta: CuentaLealtad): void {
    this.cuentaEditandoMiembro.set(cuenta);
    this.modalError.set(null);
    this.editarMiembroForm.reset({
      nombre_completo: cuenta.miembro_nombre ?? '',
      email: cuenta.miembro_email ?? '',
      telefono: cuenta.miembro_telefono ?? '',
    });
    this.modalEditarMiembroAbierto.set(true);
  }

  cerrarEditarMiembro(): void {
    this.modalEditarMiembroAbierto.set(false);
    this.cuentaEditandoMiembro.set(null);
    this.modalError.set(null);
  }

  guardarEditarMiembro(): void {
    const cuenta = this.cuentaEditandoMiembro();
    if (!cuenta) return;
    if (this.editarMiembroForm.invalid) {
      this.editarMiembroForm.markAllAsTouched();
      this.modalError.set('Revisa los datos del miembro antes de guardar.');
      return;
    }

    const raw = this.editarMiembroForm.getRawValue();
    this.saving.set(true);
    this.modalError.set(null);

    this.lealtadService
      .actualizarMiembro(cuenta.miembro_id, {
        nombre_completo: raw.nombre_completo.trim(),
        email: raw.email.trim(),
        telefono: raw.telefono.trim(),
      })
      .subscribe({
        next: () => {
          this.saving.set(false);
          this.cerrarEditarMiembro();
          this.exitoMensaje.set('Datos del miembro actualizados correctamente.');
          this.cargarCuentas();
        },
        error: (err: HttpErrorResponse) => {
          this.saving.set(false);
          this.modalError.set(this.extraerError(err, 'No se pudieron actualizar los datos del miembro.'));
        },
      });
  }

  cambiarEstadoCuenta(cuenta: CuentaLealtad, nuevoEstado: EstadoCuentaLealtad): void {
    const accion =
      nuevoEstado === 'ACTIVA' ? 'activar' : nuevoEstado === 'BLOQUEADA' ? 'bloquear' : 'cerrar';
    if (
      !window.confirm(
        `¿Deseas ${accion} la cuenta de "${cuenta.miembro_nombre}" en ${cuenta.programa_nombre}?`,
      )
    ) {
      return;
    }
    this.errorMensaje.set(null);
    this.exitoMensaje.set(null);

    this.lealtadService.actualizarEstadoCuenta(cuenta.id, nuevoEstado).subscribe({
      next: () => {
        this.exitoMensaje.set(`Estado de la cuenta actualizado a ${this.etiquetaEstado(nuevoEstado)}.`);
        this.cargarCuentas();
      },
      error: (err: HttpErrorResponse) => {
        this.errorMensaje.set(this.extraerError(err, 'No se pudo cambiar el estado de la cuenta.'));
      },
    });
  }

  etiquetaEstado(estado: EstadoCuentaLealtad): string {
    switch (estado) {
      case 'ACTIVA':
        return 'Activa';
      case 'BLOQUEADA':
        return 'Bloqueada';
      default:
        return 'Cerrada';
    }
  }

  // =====================================================================
  // Movimientos (Extracto / Libro mayor)
  // =====================================================================

  abrirMovimientos(cuenta: CuentaLealtad): void {
    this.cuentaMovimientos.set(cuenta);
    this.transacciones.set([]);
    this.modalError.set(null);
    this.loadingMovimientos.set(true);
    this.modalMovimientosAbierto.set(true);

    this.lealtadService.listarTransacciones(cuenta.id).subscribe({
      next: (rows) => {
        this.transacciones.set(rows);
        this.loadingMovimientos.set(false);
      },
      error: (err: HttpErrorResponse) => {
        this.loadingMovimientos.set(false);
        this.modalError.set(this.extraerError(err, 'No se pudo cargar el extracto de movimientos.'));
      },
    });
  }

  cerrarMovimientos(): void {
    this.modalMovimientosAbierto.set(false);
    this.cuentaMovimientos.set(null);
    this.transacciones.set([]);
    this.modalError.set(null);
  }

  etiquetaTipoTransaccion(tipo: TipoTransaccionLealtad): string {
    switch (tipo) {
      case 'ACUMULACION':
        return 'Acumulación';
      case 'REDENCION':
        return 'Redención';
      case 'AJUSTE_POSITIVO':
        return 'Ajuste +';
      case 'AJUSTE_NEGATIVO':
        return 'Ajuste -';
      case 'VENCIMIENTO':
        return 'Vencimiento';
      default:
        return 'Reverso';
    }
  }

  // =====================================================================
  // Acumular / Ajustar puntos
  // =====================================================================

  abrirAcumular(cuenta: CuentaLealtad): void {
    this.cuentaAcumular.set(cuenta);
    this.modoAcumular.set('MONTO');
    this.modalError.set(null);
    this.acumularForm.reset({
      monto_compra: null,
      puntos: null,
      observaciones: '',
    });
    this.modalAcumularAbierto.set(true);
  }

  cerrarAcumular(): void {
    this.modalAcumularAbierto.set(false);
    this.cuentaAcumular.set(null);
    this.modalError.set(null);
  }

  guardarAcumular(): void {
    const cuenta = this.cuentaAcumular();
    if (!cuenta) return;
    const raw = this.acumularForm.getRawValue();
    this.modalError.set(null);

    if (this.modoAcumular() === 'MONTO') {
      const monto = Number(raw.monto_compra);
      if (!monto || monto <= 0) {
        this.modalError.set('Ingresa un monto de compra mayor a 0.');
        return;
      }
      this.saving.set(true);
      this.lealtadService
        .acumularPuntos(cuenta.id, {
          monto_compra: monto,
          observaciones: raw.observaciones.trim() || null,
        })
        .subscribe({
          next: (tx) => {
            this.saving.set(false);
            this.cerrarAcumular();
            this.exitoMensaje.set(
              `Se acumularon +${tx.puntos} pts a ${cuenta.miembro_nombre}. Nuevo saldo: ${tx.saldo_posterior} pts.`,
            );
            this.cargarCuentas();
          },
          error: (err: HttpErrorResponse) => {
            this.saving.set(false);
            this.modalError.set(this.extraerError(err, 'No se pudieron acumular los puntos.'));
          },
        });
      return;
    }

    const pts = Number(raw.puntos);
    if (!pts || pts === 0 || !Number.isInteger(pts)) {
      this.modalError.set('Ingresa una cantidad entera de puntos distinta de 0 (positiva o negativa).');
      return;
    }
    this.saving.set(true);
    this.lealtadService
      .acumularPuntos(cuenta.id, {
        puntos: pts,
        observaciones: raw.observaciones.trim() || null,
      })
      .subscribe({
        next: (tx) => {
          this.saving.set(false);
          this.cerrarAcumular();
          const signo = tx.puntos > 0 ? `+${tx.puntos}` : String(tx.puntos);
          this.exitoMensaje.set(
            `Ajuste registrado (${signo} pts) para ${cuenta.miembro_nombre}. Nuevo saldo: ${tx.saldo_posterior} pts.`,
          );
          this.cargarCuentas();
        },
        error: (err: HttpErrorResponse) => {
          this.saving.set(false);
          this.modalError.set(this.extraerError(err, 'No se pudo registrar el ajuste de puntos.'));
        },
      });
  }

  // =====================================================================
  // Redimir puntos
  // =====================================================================

  abrirRedimir(cuenta: CuentaLealtad): void {
    this.cuentaRedimir.set(cuenta);
    this.modalError.set(null);
    this.redimirForm.reset({
      puntos_redimidos: Math.min(100, Math.max(1, cuenta.saldo_puntos)),
      observaciones: '',
    });
    this.modalRedimirAbierto.set(true);
  }

  cerrarRedimir(): void {
    this.modalRedimirAbierto.set(false);
    this.cuentaRedimir.set(null);
    this.modalError.set(null);
  }

  descuentoEstimadoRedencion(): number {
    const cuenta = this.cuentaRedimir();
    if (!cuenta) return 0;
    const pts = Number(this.redimirForm.controls.puntos_redimidos.value) || 0;
    return Math.max(0, pts * (cuenta.valor_punto_redencion ?? 1));
  }

  guardarRedimir(): void {
    const cuenta = this.cuentaRedimir();
    if (!cuenta) return;
    const raw = this.redimirForm.getRawValue();
    const pts = Number(raw.puntos_redimidos);
    if (!pts || pts <= 0 || !Number.isInteger(pts)) {
      this.modalError.set('Ingresa una cantidad entera de puntos mayor a 0.');
      return;
    }
    if (pts > cuenta.saldo_puntos) {
      this.modalError.set(`Saldo insuficiente. Disponible: ${cuenta.saldo_puntos} pts.`);
      return;
    }

    this.saving.set(true);
    this.modalError.set(null);
    this.lealtadService
      .redimirPuntos(cuenta.id, {
        puntos_redimidos: pts,
        observaciones: raw.observaciones.trim() || null,
      })
      .subscribe({
        next: (red) => {
          this.saving.set(false);
          this.cerrarRedimir();
          this.exitoMensaje.set(
            `Redención exitosa: -${red.puntos_redimidos} pts (${red.moneda_simbolo || '$'} ${red.monto_descontado.toLocaleString('es-CO')} ${red.moneda_codigo || ''}). Nuevo saldo: ${red.saldo_posterior ?? 0} pts.`,
          );
          this.cargarCuentas();
        },
        error: (err: HttpErrorResponse) => {
          this.saving.set(false);
          this.modalError.set(this.extraerError(err, 'No se pudo procesar la redención de puntos.'));
        },
      });
  }

  private construirFiltros() {
    return {
      programa_id: this.filtroProgramaId(),
      estado: this.filtroEstado(),
      search: this.filtroSearch(),
    };
  }

  private extraerError(err: HttpErrorResponse, fallback: string): string {
    const body = err.error as { error?: { message?: string } } | null;
    return body?.error?.message ?? fallback;
  }
}
