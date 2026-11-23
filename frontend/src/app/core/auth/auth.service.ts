import { Injectable, computed, inject, signal } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, tap } from 'rxjs';
import { Router } from '@angular/router';

import { environment } from '../../../environments/environment';
import { LoginRequest, LoginResponse, UsuarioAutenticado } from './auth.models';

const TOKEN_KEY = 'tesserak_token';
const USER_KEY = 'tesserak_user';

@Injectable({ providedIn: 'root' })
export class AuthService {
  private readonly http = inject(HttpClient);
  private readonly router = inject(Router);

  private readonly _token = signal<string | null>(this.readStorage(TOKEN_KEY));
  private readonly _usuario = signal<UsuarioAutenticado | null>(this.readStoredUser());

  readonly token = this._token.asReadonly();
  readonly usuario = this._usuario.asReadonly();
  readonly isAuthenticated = computed(() => Boolean(this._token() && this._usuario()));

  login(payload: LoginRequest, recordar = true): Observable<LoginResponse> {
    return this.http.post<LoginResponse>(`${environment.apiUrl}/auth/login`, payload).pipe(
      tap((res) => {
        const storage = recordar ? localStorage : sessionStorage;
        this.clearStorage();
        storage.setItem(TOKEN_KEY, res.access_token);
        storage.setItem(USER_KEY, JSON.stringify(res.usuario));
        this._token.set(res.access_token);
        this._usuario.set(res.usuario);
      }),
    );
  }

  logout(): void {
    this.clearStorage();
    this._token.set(null);
    this._usuario.set(null);
    void this.router.navigate(['/login']);
  }

  private readStorage(key: string): string | null {
    return localStorage.getItem(key) ?? sessionStorage.getItem(key);
  }

  private readStoredUser(): UsuarioAutenticado | null {
    const raw = this.readStorage(USER_KEY);
    if (!raw) return null;
    try {
      return JSON.parse(raw) as UsuarioAutenticado;
    } catch {
      return null;
    }
  }

  private clearStorage(): void {
    localStorage.removeItem(TOKEN_KEY);
    localStorage.removeItem(USER_KEY);
    sessionStorage.removeItem(TOKEN_KEY);
    sessionStorage.removeItem(USER_KEY);
  }
}
