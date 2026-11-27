import { Component, OnInit, computed, inject, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { HttpErrorResponse } from '@angular/common/http';
import { catchError, forkJoin, of } from 'rxjs';

import { AuthService } from '../../../core/auth/auth.service';
import { SucursalOpcion, UsuarioItem } from '../usuarios.models';
import { UsuariosService } from '../usuarios.service';

@Component({
  selector: 'app-mi-perfil',
  standalone: true,
  imports: [ReactiveFormsModule],
  templateUrl: './mi-perfil.html',
  styleUrl: './mi-perfil.scss',
})
export class MiPerfilComponent implements OnInit {
  private readonly fb = inject(FormBuilder);
  private readonly auth = inject(AuthService);
  private readonly usuariosSvc = inject(UsuariosService);

  readonly loading = signal(true);
  readonly perfil = signal<UsuarioItem | null>(null);
  readonly sucursales = signal<SucursalOpcion[]>([]);

  readonly savingInfo = signal(false);
  readonly infoSuccess = signal<string | null>(null);
  readonly infoError = signal<string | null>(null);

  readonly savingPassword = signal(false);
  readonly passwordSuccess = signal<string | null>(null);
  readonly passwordError = signal<string | null>(null);

  readonly sucursalesMap = computed(
    () => new Map(this.sucursales().map((s) => [s.id, s.nombre])),
  );

  readonly iniciales = computed(() => {
    const p = this.perfil();
    const n = (p?.nombres ?? this.auth.usuario()?.nombres ?? '').trim().charAt(0);
    const a = (p?.apellidos ?? this.auth.usuario()?.apellidos ?? '').trim().charAt(0);
    return (n + a).toUpperCase() || 'TS';
  });

  readonly infoForm = this.fb.nonNullable.group({
    nombres: ['', [Validators.required, Validators.maxLength(100)]],
    apellidos: ['', [Validators.required, Validators.maxLength(100)]],
    nombre_usuario: ['', [Validators.required, Validators.minLength(3), Validators.maxLength(50)]],
    email: ['', [Validators.required, Validators.email, Validators.maxLength(255)]],
  });

  readonly passwordForm = this.fb.nonNullable.group({
    password_actual: ['', [Validators.required, Validators.maxLength(72)]],
    password_nueva: ['', [Validators.required, Validators.minLength(8), Validators.maxLength(72)]],
    confirmar_password: ['', [Validators.required]],
  });

  ngOnInit(): void {
    this.cargarPerfil();
  }

  cargarPerfil(): void {
    this.loading.set(true);
    this.infoError.set(null);

    // Precargar inmediatamente los datos actuales de la sesión para que el formulario
    // siempre muestre la información actual del usuario.
    const sesion = this.auth.usuario();
    if (sesion) {
      this.infoForm.patchValue({
        nombres: sesion.nombres ?? '',
        apellidos: sesion.apellidos ?? '',
        nombre_usuario: sesion.nombre_usuario ?? '',
        email: sesion.email ?? '',
      });
      this.infoForm.markAsPristine();
    }

    const perfil$ = this.usuariosSvc.obtenerMiPerfil().pipe(
      catchError(() => {
        if (sesion?.id) {
          return this.usuariosSvc.obtenerPorId(sesion.id).pipe(catchError(() => of(null)));
        }
        return of(null);
      }),
    );

    forkJoin({
      perfil: perfil$,
      sucursales: this.usuariosSvc
        .listarSucursales()
        .pipe(catchError(() => of([] as SucursalOpcion[]))),
    }).subscribe({
      next: ({ perfil, sucursales }) => {
        this.sucursales.set(sucursales);
        if (perfil) {
          this.perfil.set(perfil);
          this.infoForm.patchValue({
            nombres: perfil.nombres,
            apellidos: perfil.apellidos,
            nombre_usuario: perfil.nombre_usuario,
            email: perfil.email,
          });
          this.infoForm.markAsPristine();
        } else if (sesion) {
          this.perfil.set({
            id: sesion.id,
            key: sesion.key ?? '',
            empresa_id: sesion.empresa_id,
            sucursal_id: sesion.sucursal_id,
            sucursal_ids: sesion.sucursal_ids ?? (sesion.sucursal_id ? [sesion.sucursal_id] : []),
            roles: (sesion.roles ?? []).map((codigo, idx) => ({
              id: idx + 1,
              codigo,
              nombre: codigo === 'ADMINISTRADOR' ? 'Administrador' : 'Usuario',
              descripcion: null,
              es_sistema: true,
              activo: true,
            })),
            es_administrador: Boolean(sesion.es_administrador),
            todas_sucursales: Boolean(sesion.es_administrador),
            nombre_usuario: sesion.nombre_usuario,
            nombres: sesion.nombres,
            apellidos: sesion.apellidos,
            email: sesion.email,
            activo: true,
            ultimo_login_at: null,
            created_at: '',
            updated_at: '',
          });
        } else {
          this.infoError.set('No se pudo cargar tu perfil.');
        }
        this.loading.set(false);
      },
      error: (err) => {
        this.infoError.set(this.extraerMensajeError(err, 'No se pudo cargar tu perfil.'));
        this.loading.set(false);
      },
    });
  }

  campoInfoInvalido(nombre: keyof typeof this.infoForm.controls): boolean {
    const ctrl = this.infoForm.controls[nombre];
    return ctrl.invalid && (ctrl.dirty || ctrl.touched);
  }

  campoPassInvalido(nombre: keyof typeof this.passwordForm.controls): boolean {
    const ctrl = this.passwordForm.controls[nombre];
    return ctrl.invalid && (ctrl.dirty || ctrl.touched);
  }

  guardarInformacion(): void {
    this.infoSuccess.set(null);
    this.infoError.set(null);

    if (this.infoForm.invalid) {
      this.infoForm.markAllAsTouched();
      return;
    }

    const v = this.infoForm.getRawValue();
    this.savingInfo.set(true);

    this.usuariosSvc
      .actualizarMiPerfil({
        nombres: v.nombres.trim(),
        apellidos: v.apellidos.trim(),
        nombre_usuario: v.nombre_usuario.trim(),
        email: v.email.trim(),
      })
      .subscribe({
        next: (actualizado) => {
          this.savingInfo.set(false);
          this.perfil.set(actualizado);
          this.auth.actualizarUsuarioLocal({
            nombres: actualizado.nombres,
            apellidos: actualizado.apellidos,
            nombre_usuario: actualizado.nombre_usuario,
            email: actualizado.email,
          });
          this.infoSuccess.set('Tu información personal se actualizó correctamente.');
        },
        error: (err) => {
          this.savingInfo.set(false);
          this.infoError.set(
            this.extraerMensajeError(err, 'No se pudo actualizar tu información personal.'),
          );
        },
      });
  }

  cambiarPassword(): void {
    this.passwordSuccess.set(null);
    this.passwordError.set(null);

    if (this.passwordForm.invalid) {
      this.passwordForm.markAllAsTouched();
      return;
    }

    const v = this.passwordForm.getRawValue();
    if (v.password_nueva !== v.confirmar_password) {
      this.passwordError.set('La nueva contraseña y su confirmación no coinciden.');
      return;
    }

    this.savingPassword.set(true);

    this.usuariosSvc
      .actualizarMiPerfil({
        password_actual: v.password_actual,
        password_nueva: v.password_nueva,
      })
      .subscribe({
        next: () => {
          this.savingPassword.set(false);
          this.passwordForm.reset({
            password_actual: '',
            password_nueva: '',
            confirmar_password: '',
          });
          this.passwordSuccess.set('Tu contraseña se actualizó correctamente.');
        },
        error: (err) => {
          this.savingPassword.set(false);
          this.passwordError.set(
            this.extraerMensajeError(err, 'No se pudo cambiar la contraseña.'),
          );
        },
      });
  }

  nombreSucursal(id: number): string {
    return this.sucursalesMap().get(id) ?? `Sucursal #${id}`;
  }

  private extraerMensajeError(err: unknown, fallback: string): string {
    if (err instanceof HttpErrorResponse) {
      const apiMsg = err.error?.error?.message;
      if (typeof apiMsg === 'string' && apiMsg.trim()) return apiMsg;
    }
    return fallback;
  }
}
