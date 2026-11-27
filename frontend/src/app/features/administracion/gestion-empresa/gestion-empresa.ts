import { Component, OnInit, inject, signal } from '@angular/core';
import { FormBuilder, FormGroup, ReactiveFormsModule, Validators } from '@angular/forms';
import { HttpErrorResponse } from '@angular/common/http';
import { catchError, forkJoin, of } from 'rxjs';

import { AuthService } from '../../../core/auth/auth.service';
import { ApiErrorBody } from '../../../core/auth/auth.models';
import { InventarioService } from '../../inventario/inventario.service';
import { SucursalOption, Ubicacion } from '../../inventario/inventario.models';
import { GestionEmpresaService } from '../gestion-empresa.service';
import {
  EmpresaDetalle,
  FrecuenciaReinicio,
  SecuenciaDocumento,
  SucursalDetalle,
  TipoDocumentoConsecutivo,
  UsuarioKeyDetalle,
} from '../gestion-empresa.models';

export interface SecuenciaItemForm {
  tipo: TipoDocumentoConsecutivo;
  secuenciaId: number | null;
  form: FormGroup;
  saving: boolean;
}

export const FRECUENCIAS_REINICIO: { value: FrecuenciaReinicio; label: string }[] = [
  { value: 'NUNCA', label: 'Nunca (correlativo continuo)' },
  { value: 'ANUAL', label: 'Anual (ej. PREF-2026-000001)' },
  { value: 'MENSUAL', label: 'Mensual (ej. PREF-2026-11-000001)' },
];

@Component({
  selector: 'app-gestion-empresa',
  standalone: true,
  imports: [ReactiveFormsModule],
  templateUrl: './gestion-empresa.html',
  styleUrl: './gestion-empresa.scss',
})
export class GestionEmpresaComponent implements OnInit {
  private readonly svc = inject(GestionEmpresaService);
  private readonly invSvc = inject(InventarioService);
  private readonly auth = inject(AuthService);
  private readonly fb = inject(FormBuilder);

  readonly esAdministrador = this.auth.esAdministrador;
  readonly frecuencias = FRECUENCIAS_REINICIO;

  readonly loading = signal(true);
  readonly errorMensaje = signal<string | null>(null);
  readonly exitoMensaje = signal<string | null>(null);

  readonly empresa = signal<EmpresaDetalle | null>(null);
  readonly usuarioDetalle = signal<UsuarioKeyDetalle | null>(null);
  readonly secuenciasItems = signal<SecuenciaItemForm[]>([]);
  readonly ubicaciones = signal<Ubicacion[]>([]);
  readonly sucursales = signal<SucursalOption[]>([]);
  readonly sucursalesEmpresa = signal<SucursalDetalle[]>([]);

  // Controles de sucursales (crear, editar, activar/desactivar, soft delete)
  readonly modalSucAbierto = signal(false);
  readonly savingSuc = signal(false);
  readonly modalSucError = signal<string | null>(null);
  readonly editandoSuc = signal<SucursalDetalle | null>(null);
  readonly sucAEliminar = signal<SucursalDetalle | null>(null);

  readonly formSuc = this.fb.group({
    nombre: ['', [Validators.required, Validators.maxLength(150)]],
    direccion: ['', [Validators.maxLength(255)]],
    activo: [true],
  });

  // Controles de ubicaciones (crear, editar, activar/desactivar, soft delete)
  readonly modalUbicAbierto = signal(false);
  readonly savingUbic = signal(false);
  readonly modalUbicError = signal<string | null>(null);
  readonly editandoUbic = signal<Ubicacion | null>(null);
  readonly ubicAEliminar = signal<Ubicacion | null>(null);

  readonly formUbic = this.fb.group({
    sucursal_id: [null as number | null, [Validators.required]],
    codigo: ['', [Validators.required, Validators.maxLength(40)]],
    nombre: ['', [Validators.required, Validators.maxLength(120)]],
    descripcion: ['', [Validators.maxLength(255)]],
    activo: [true],
  });

  // Controles de la key de usuario
  readonly keyVisible = signal(false);
  readonly keyCopiada = signal(false);
  readonly confirmRegenerarAbierto = signal(false);
  readonly regenerandoKey = signal(false);

  ngOnInit(): void {
    this.cargarDatos();
  }

  cargarDatos(): void {
    this.loading.set(true);
    this.errorMensaje.set(null);

    const sesion = this.auth.usuario();
    const empresaId = sesion?.empresa_id ?? null;
    const usuarioId = sesion?.id ?? null;

    forkJoin({
      empresa: empresaId
        ? this.svc.obtenerEmpresa(empresaId).pipe(catchError(() => of(null)))
        : of(null),
      usuario: usuarioId
        ? this.svc.obtenerUsuario(usuarioId).pipe(catchError(() => of(null)))
        : of(null),
      tipos: this.svc
        .listarTiposDocumento()
        .pipe(catchError(() => of([] as TipoDocumentoConsecutivo[]))),
      secuencias: this.svc
        .listarSecuencias()
        .pipe(catchError(() => of([] as SecuenciaDocumento[]))),
      ubicaciones: this.invSvc
        .listarUbicaciones()
        .pipe(catchError(() => of([] as Ubicacion[]))),
      sucursales: this.invSvc
        .listarSucursales()
        .pipe(catchError(() => of([] as SucursalOption[]))),
      sucursalesEmpresa: this.esAdministrador()
        ? this.svc.listarSucursales().pipe(catchError(() => of([] as SucursalDetalle[])))
        : of([] as SucursalDetalle[]),
    }).subscribe({
      next: ({
        empresa,
        usuario,
        tipos,
        secuencias,
        ubicaciones,
        sucursales,
        sucursalesEmpresa,
      }) => {
        this.empresa.set(empresa);
        this.usuarioDetalle.set(usuario);
        this.ubicaciones.set(ubicaciones);
        this.sucursales.set(sucursales);
        this.sucursalesEmpresa.set(sucursalesEmpresa);

        const secPorTipo = new Map(secuencias.map((s) => [s.tipo_documento_id, s]));
        const items: SecuenciaItemForm[] = tipos
          .filter((t) => t.activo)
          .map((tipo) => {
            const sec = secPorTipo.get(tipo.id);
            return {
              tipo,
              secuenciaId: sec?.id ?? null,
              saving: false,
              form: this.fb.group({
                prefijo: [
                  sec?.prefijo ?? tipo.prefijo_defecto,
                  [Validators.required, Validators.minLength(1), Validators.maxLength(10)],
                ],
                siguiente_numero: [
                  sec?.siguiente_numero ?? 1,
                  [Validators.required, Validators.min(1)],
                ],
                longitud_relleno: [
                  sec?.longitud_relleno ?? 6,
                  [Validators.required, Validators.min(1), Validators.max(20)],
                ],
                frecuencia_reinicio: [
                  (sec?.frecuencia_reinicio ?? 'NUNCA') as FrecuenciaReinicio,
                  [Validators.required],
                ],
                activo: [sec?.activo ?? true],
              }),
            };
          });

        this.secuenciasItems.set(items);
        this.loading.set(false);
      },
      error: (err) => {
        this.loading.set(false);
        this.errorMensaje.set(
          this.extraerMensajeError(err, 'No se pudo cargar la configuración de la empresa.'),
        );
      },
    });
  }

  // =====================================================================
  // Prefijos y consecutivos (Compras y Ventas)
  // =====================================================================

  vistaPrevia(item: SecuenciaItemForm): string {
    const v = item.form.getRawValue();
    const prefijo = String(v.prefijo ?? item.tipo.prefijo_defecto)
      .trim()
      .toUpperCase();
    const num = Math.max(1, Number(v.siguiente_numero ?? 1));
    const pad = Math.min(20, Math.max(1, Number(v.longitud_relleno ?? 6)));
    const freq = (v.frecuencia_reinicio ?? 'NUNCA') as FrecuenciaReinicio;
    const numStr = String(Math.floor(num)).padStart(pad, '0');

    const now = new Date();
    if (freq === 'ANUAL') {
      return `${prefijo}-${now.getFullYear()}-${numStr}`;
    }
    if (freq === 'MENSUAL') {
      const mes = String(now.getMonth() + 1).padStart(2, '0');
      return `${prefijo}-${now.getFullYear()}-${mes}-${numStr}`;
    }
    return `${prefijo}-${numStr}`;
  }

  guardarSecuencia(item: SecuenciaItemForm): void {
    this.errorMensaje.set(null);
    if (item.form.invalid) {
      item.form.markAllAsTouched();
      return;
    }

    const v = item.form.getRawValue();
    const prefijo = String(v.prefijo ?? '')
      .trim()
      .toUpperCase();
    const siguienteNumero = Number(v.siguiente_numero ?? 1);
    const longitudRelleno = Number(v.longitud_relleno ?? 6);
    const frecuenciaReinicio = v.frecuencia_reinicio as FrecuenciaReinicio;
    const activo = Boolean(v.activo);

    item.saving = true;

    const req$ = item.secuenciaId
      ? this.svc.actualizarSecuencia(item.secuenciaId, {
          prefijo,
          siguiente_numero: siguienteNumero,
          longitud_relleno: longitudRelleno,
          frecuencia_reinicio: frecuenciaReinicio,
          activo,
        })
      : this.svc.crearSecuencia({
          tipo_documento_id: item.tipo.id,
          prefijo,
          siguiente_numero: siguienteNumero,
          longitud_relleno: longitudRelleno,
          frecuencia_reinicio: frecuenciaReinicio,
          activo,
        });

    req$.subscribe({
      next: (guardada) => {
        item.saving = false;
        item.secuenciaId = guardada.id;
        item.form.patchValue({
          prefijo: guardada.prefijo,
          siguiente_numero: guardada.siguiente_numero,
          longitud_relleno: guardada.longitud_relleno,
          frecuencia_reinicio: guardada.frecuencia_reinicio,
          activo: guardada.activo,
        });
        this.mostrarExito(
          `Prefijo de ${item.tipo.nombre} actualizado (${guardada.ejemplo_proximo}).`,
        );
      },
      error: (err) => {
        item.saving = false;
        this.errorMensaje.set(
          this.extraerMensajeError(err, `No se pudo guardar el prefijo de ${item.tipo.nombre}.`),
        );
      },
    });
  }

  // =====================================================================
  // Administración de Ubicaciones (crear, editar, activar/desactivar, soft delete)
  // =====================================================================

  abrirNuevaUbicacion(): void {
    this.editandoUbic.set(null);
    this.modalUbicError.set(null);
    this.formUbic.reset({
      sucursal_id: null,
      codigo: '',
      nombre: '',
      descripcion: '',
      activo: true,
    });
    this.modalUbicAbierto.set(true);
    if (this.sucursales().length === 0) {
      this.cargarSucursales();
    }
  }

  abrirEditarUbicacion(u: Ubicacion): void {
    this.editandoUbic.set(u);
    this.modalUbicError.set(null);
    this.formUbic.reset({
      sucursal_id: u.sucursal_id,
      codigo: u.codigo,
      nombre: u.nombre,
      descripcion: u.descripcion ?? '',
      activo: u.activo,
    });
    this.modalUbicAbierto.set(true);
    if (this.sucursales().length === 0) {
      this.cargarSucursales();
    }
  }

  private cargarSucursales(): void {
    this.invSvc.listarSucursales().subscribe({
      next: (list) => this.sucursales.set(list),
      error: (err) => {
        this.modalUbicError.set(
          this.extraerMensajeError(
            err,
            'No se pudieron cargar las sucursales disponibles. Verifica que el servidor backend esté reiniciado y actualizado.',
          ),
        );
      },
    });
  }

  cerrarModalUbic(): void {
    if (this.savingUbic()) return;
    this.modalUbicAbierto.set(false);
    this.editandoUbic.set(null);
    this.modalUbicError.set(null);
  }

  guardarUbicacion(): void {
    this.modalUbicError.set(null);
    if (this.formUbic.invalid) {
      this.formUbic.markAllAsTouched();
      return;
    }

    const v = this.formUbic.getRawValue();
    this.savingUbic.set(true);
    const actual = this.editandoUbic();

    const req$ = actual
      ? this.invSvc.actualizarUbicacion(actual.id, {
          sucursal_id: Number(v.sucursal_id),
          codigo: (v.codigo ?? '').trim(),
          nombre: (v.nombre ?? '').trim(),
          descripcion: (v.descripcion ?? '').trim() || null,
          activo: Boolean(v.activo),
        })
      : this.invSvc.crearUbicacion({
          sucursal_id: Number(v.sucursal_id),
          codigo: (v.codigo ?? '').trim(),
          nombre: (v.nombre ?? '').trim(),
          descripcion: (v.descripcion ?? '').trim() || null,
          activo: Boolean(v.activo),
        });

    req$.subscribe({
      next: (u) => {
        this.savingUbic.set(false);
        this.modalUbicAbierto.set(false);
        this.editandoUbic.set(null);
        this.mostrarExito(`Ubicación "${u.nombre}" guardada correctamente.`);
        this.invSvc.listarUbicaciones().subscribe((list) => this.ubicaciones.set(list));
      },
      error: (err) => {
        this.savingUbic.set(false);
        this.modalUbicError.set(
          this.extraerMensajeError(err, 'No se pudo guardar la ubicación.'),
        );
      },
    });
  }

  toggleActivoUbicacion(u: Ubicacion): void {
    const nuevoEstado = !u.activo;
    this.invSvc.actualizarUbicacion(u.id, { activo: nuevoEstado }).subscribe({
      next: () => {
        this.mostrarExito(
          `Ubicación "${u.nombre}" ${nuevoEstado ? 'activada' : 'desactivada'}.`,
        );
        this.invSvc.listarUbicaciones().subscribe((list) => this.ubicaciones.set(list));
      },
      error: (err) => {
        this.errorMensaje.set(
          this.extraerMensajeError(err, 'No se pudo cambiar el estado de la ubicación.'),
        );
      },
    });
  }

  solicitarEliminarUbicacion(u: Ubicacion): void {
    this.ubicAEliminar.set(u);
  }

  cancelarEliminarUbicacion(): void {
    if (this.savingUbic()) return;
    this.ubicAEliminar.set(null);
  }

  confirmarEliminarUbicacion(): void {
    const target = this.ubicAEliminar();
    if (!target) return;
    this.savingUbic.set(true);
    this.invSvc.eliminarUbicacion(target.id).subscribe({
      next: () => {
        this.savingUbic.set(false);
        this.ubicAEliminar.set(null);
        this.mostrarExito(`Ubicación "${target.nombre}" eliminada.`);
        this.invSvc.listarUbicaciones().subscribe((list) => this.ubicaciones.set(list));
      },
      error: (err) => {
        this.savingUbic.set(false);
        this.ubicAEliminar.set(null);
        this.errorMensaje.set(
          this.extraerMensajeError(err, 'No se pudo eliminar la ubicación.'),
        );
      },
    });
  }

  // =====================================================================
  // Sucursales (Crear, Editar, Activar/Desactivar, Soft Delete)
  // =====================================================================

  abrirNuevaSucursal(): void {
    if (!this.esAdministrador()) return;
    this.editandoSuc.set(null);
    this.modalSucError.set(null);
    this.formSuc.reset({ nombre: '', direccion: '', activo: true });
    this.modalSucAbierto.set(true);
  }

  abrirEditarSucursal(s: SucursalDetalle): void {
    if (!this.esAdministrador()) return;
    this.editandoSuc.set(s);
    this.modalSucError.set(null);
    this.formSuc.reset({
      nombre: s.nombre,
      direccion: s.direccion ?? '',
      activo: s.activo,
    });
    this.modalSucAbierto.set(true);
  }

  cerrarModalSucursal(): void {
    if (this.savingSuc()) return;
    this.modalSucAbierto.set(false);
    this.editandoSuc.set(null);
    this.modalSucError.set(null);
  }

  guardarSucursal(): void {
    if (!this.esAdministrador()) return;
    if (this.formSuc.invalid) {
      this.formSuc.markAllAsTouched();
      return;
    }

    const v = this.formSuc.getRawValue();
    const payload = {
      nombre: (v.nombre ?? '').trim(),
      direccion: (v.direccion ?? '').trim() || null,
      activo: Boolean(v.activo),
    };

    this.savingSuc.set(true);
    this.modalSucError.set(null);
    const actual = this.editandoSuc();

    const obs = actual
      ? this.svc.actualizarSucursal(actual.id, payload)
      : this.svc.crearSucursal(payload);

    obs.subscribe({
      next: () => {
        this.savingSuc.set(false);
        this.modalSucAbierto.set(false);
        this.editandoSuc.set(null);
        this.mostrarExito(
          actual ? 'Sucursal actualizada correctamente.' : 'Sucursal creada correctamente.',
        );
        this.refrescarSucursales();
      },
      error: (err) => {
        this.savingSuc.set(false);
        this.modalSucError.set(
          this.extraerMensajeError(err, 'No se pudo guardar la sucursal.'),
        );
      },
    });
  }

  toggleActivoSucursal(s: SucursalDetalle): void {
    if (!this.esAdministrador()) return;
    this.errorMensaje.set(null);
    this.svc.actualizarSucursal(s.id, { activo: !s.activo }).subscribe({
      next: () => {
        this.mostrarExito(
          s.activo ? `Sucursal "${s.nombre}" desactivada.` : `Sucursal "${s.nombre}" activada.`,
        );
        this.refrescarSucursales();
      },
      error: (err) => {
        this.errorMensaje.set(
          this.extraerMensajeError(err, 'No se pudo actualizar el estado de la sucursal.'),
        );
      },
    });
  }

  solicitarEliminarSucursal(s: SucursalDetalle): void {
    if (!this.esAdministrador()) return;
    this.sucAEliminar.set(s);
  }

  cancelarEliminarSucursal(): void {
    if (this.savingSuc()) return;
    this.sucAEliminar.set(null);
  }

  confirmarEliminarSucursal(): void {
    const target = this.sucAEliminar();
    if (!target || !this.esAdministrador()) return;
    this.savingSuc.set(true);
    this.errorMensaje.set(null);

    this.svc.eliminarSucursal(target.id).subscribe({
      next: () => {
        this.savingSuc.set(false);
        this.sucAEliminar.set(null);
        this.mostrarExito(`Sucursal "${target.nombre}" eliminada correctamente.`);
        this.refrescarSucursales();
      },
      error: (err) => {
        this.savingSuc.set(false);
        this.sucAEliminar.set(null);
        this.errorMensaje.set(
          this.extraerMensajeError(err, 'No se pudo eliminar la sucursal.'),
        );
      },
    });
  }

  private refrescarSucursales(): void {
    this.svc.listarSucursales().subscribe((list) => this.sucursalesEmpresa.set(list));
    this.invSvc.listarSucursales().subscribe((list) => this.sucursales.set(list));
    this.invSvc.listarUbicaciones().subscribe((list) => this.ubicaciones.set(list));
  }

  // =====================================================================
  // Key de usuario (Ver, Copiar, Cambiar)
  // =====================================================================

  toggleVerKey(): void {
    this.keyVisible.update((v) => !v);
  }

  keyEnmascarada(key: string): string {
    if (!key) return '—';
    if (this.keyVisible()) return key;
    return `${key.slice(0, 8)}-••••-••••-••••-${key.slice(-4)}`;
  }

  copiarKey(): void {
    const key = this.usuarioDetalle()?.key;
    if (!key) return;
    if (navigator.clipboard?.writeText) {
      navigator.clipboard.writeText(key).catch(() => {});
    }
    this.keyCopiada.set(true);
    setTimeout(() => this.keyCopiada.set(false), 2500);
  }

  abrirConfirmarRegenerar(): void {
    this.confirmRegenerarAbierto.set(true);
  }

  cancelarRegenerar(): void {
    if (this.regenerandoKey()) return;
    this.confirmRegenerarAbierto.set(false);
  }

  confirmarRegenerarKey(): void {
    const usuarioId = this.usuarioDetalle()?.id ?? this.auth.usuario()?.id;
    if (!usuarioId) return;

    this.regenerandoKey.set(true);
    this.errorMensaje.set(null);

    this.svc.regenerarKeyUsuario(usuarioId).subscribe({
      next: (actualizado) => {
        this.regenerandoKey.set(false);
        this.confirmRegenerarAbierto.set(false);
        this.usuarioDetalle.set(actualizado);
        this.keyVisible.set(true);
        this.mostrarExito('La key de tu usuario ha sido regenerada correctamente.');
      },
      error: (err) => {
        this.regenerandoKey.set(false);
        this.confirmRegenerarAbierto.set(false);
        this.errorMensaje.set(
          this.extraerMensajeError(err, 'No se pudo cambiar la key del usuario.'),
        );
      },
    });
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
