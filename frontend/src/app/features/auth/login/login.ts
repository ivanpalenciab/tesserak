import { Component, inject, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { Router } from '@angular/router';
import { HttpErrorResponse } from '@angular/common/http';

import { AuthService } from '../../../core/auth/auth.service';
import { ApiErrorBody, LoginRequest } from '../../../core/auth/auth.models';

@Component({
  selector: 'app-login',
  standalone: true,
  imports: [ReactiveFormsModule],
  templateUrl: './login.html',
  styleUrl: './login.scss',
})
export class LoginComponent {
  private readonly fb = inject(FormBuilder);
  private readonly auth = inject(AuthService);
  private readonly router = inject(Router);

  readonly loading = signal(false);
  readonly showPassword = signal(false);
  readonly showEmpresaField = signal(false);
  readonly errorMessage = signal<string | null>(null);

  readonly form = this.fb.nonNullable.group({
    email: ['', [Validators.required, Validators.email, Validators.maxLength(255)]],
    password: ['', [Validators.required, Validators.minLength(6), Validators.maxLength(72)]],
    identificador_fiscal: ['', [Validators.maxLength(50)]],
    recordar: [true],
  });

  togglePassword(): void {
    this.showPassword.update((v) => !v);
  }

  toggleEmpresaField(): void {
    this.showEmpresaField.update((v) => !v);
    if (!this.showEmpresaField()) {
      this.form.controls.identificador_fiscal.setValue('');
    }
  }

  onSubmit(): void {
    this.errorMessage.set(null);

    if (this.form.invalid) {
      this.form.markAllAsTouched();
      return;
    }

    const raw = this.form.getRawValue();
    const payload: LoginRequest = {
      email: raw.email.trim(),
      password: raw.password,
    };
    if (raw.identificador_fiscal.trim()) {
      payload.identificador_fiscal = raw.identificador_fiscal.trim();
    }

    this.loading.set(true);
    this.auth.login(payload, raw.recordar).subscribe({
      next: () => {
        this.loading.set(false);
        void this.router.navigate(['/dashboard']);
      },
      error: (err: unknown) => {
        this.loading.set(false);
        this.handleError(err);
      },
    });
  }

  hasError(control: 'email' | 'password' | 'identificador_fiscal', error?: string): boolean {
    const ctrl = this.form.controls[control];
    if (!ctrl.touched) return false;
    return error ? ctrl.hasError(error) : ctrl.invalid;
  }

  private handleError(err: unknown): void {
    if (err instanceof HttpErrorResponse) {
      if (err.status === 0) {
        this.errorMessage.set(
          'No se pudo conectar con el servidor. Verifica que el backend esté en ejecución.',
        );
        return;
      }
      const body = err.error as ApiErrorBody | null;
      const code = body?.error?.code;

      if (code === 'empresa_requerida') {
        this.showEmpresaField.set(true);
        this.errorMessage.set(
          'Este correo está registrado en varias empresas. Ingresa el NIT o identificador fiscal de tu empresa.',
        );
        return;
      }
      if (body?.error?.message) {
        this.errorMessage.set(body.error.message);
        return;
      }
    }
    this.errorMessage.set('Ocurrió un error al iniciar sesión. Intenta nuevamente.');
  }
}
