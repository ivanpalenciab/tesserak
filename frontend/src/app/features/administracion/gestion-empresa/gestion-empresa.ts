import { Component, OnInit, inject, signal } from '@angular/core';
import { FormBuilder, FormGroup, ReactiveFormsModule, Validators } from '@angular/forms';
import { HttpErrorResponse } from '@angular/common/http';
import { catchError, forkJoin, of } from 'rxjs';

import { AuthService } from '../../../core/auth/auth.service';
import { ApiErrorBody } from '../../../core/auth/auth.models';
import { GestionEmpresaService } from '../gestion-empresa.service';
import {
  EmpresaDetalle,
  FrecuenciaReinicio,
  SecuenciaDocumento,
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
  private readonly auth = inject(AuthService);
  private readonly fb = inject(FormBuilder);

  readonly frecuencias = FRECUENCIAS_REINICIO;

  readonly loading = signal(true);
  readonly errorMensaje = signal<string | null>(null);
  readonly exitoMensaje = signal<string | null>(null);

  readonly empresa = signal<EmpresaDetalle | null>(null);
  readonly usuarioDetalle = signal<UsuarioKeyDetalle | null>(null);
  readonly secuenciasItems = signal<SecuenciaItemForm[]>([]);

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
    }).subscribe({
      next: ({ empresa, usuario, tipos, secuencias }) => {
        this.empresa.set(empresa);
        this.usuarioDetalle.set(usuario);

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
