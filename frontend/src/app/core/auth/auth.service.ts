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
  readonly esAdministrador = computed(() => {
    const u = this._usuario();
    if (!u) return false;
    if (u.es_administrador) return true;
    return (u.roles ?? []).some((r) => r.toUpperCase() === 'ADMINISTRADOR');
  });
  readonly esAdminSucursal = computed(() => {
    const u = this._usuario();
    if (!u) return false;
    if (u.es_admin_sucursal) return true;
    return (u.roles ?? []).some((r) => r.toUpperCase() === 'ADMINISTRADOR_SUCURSAL');
  });
  readonly puedeAdministrarUsuarios = computed(
    () => this.esAdministrador() || this.esAdminSucursal(),
  );
  readonly esGestorVentas = computed(() => {
    const u = this._usuario();
    if (!u) return false;
    if (u.es_gestor_ventas) return true;
    return (u.roles ?? []).some((r) => r.toUpperCase() === 'GESTOR_VENTAS');
  });
  readonly esAsesorComercial = computed(() => {
    const u = this._usuario();
    if (!u) return false;
    if (u.es_asesor_comercial) return true;
    return (u.roles ?? []).some((r) => r.toUpperCase() === 'ASESOR_COMERCIAL');
  });
  readonly esGestorCompras = computed(() => {
    const u = this._usuario();
    if (!u) return false;
    if (u.es_gestor_compras) return true;
    return (u.roles ?? []).some((r) => r.toUpperCase() === 'GESTOR_COMPRAS');
  });

  // Capacidades aditivas (unión de roles asignados al usuario)
  readonly tieneAccesoOperativoCompleto = computed(() => {
    const u = this._usuario();
    if (!u) return false;
    if (this.puedeAdministrarUsuarios()) return true;
    if ((u.roles ?? []).some((r) => r.toUpperCase() === 'USUARIO')) return true;
    return !this.esGestorVentas() && !this.esAsesorComercial() && !this.esGestorCompras();
  });

  readonly puedeAccederCRM = computed(
    () => this.tieneAccesoOperativoCompleto() || this.esAsesorComercial(),
  );
  readonly puedeAccederVentas = computed(
    () => this.tieneAccesoOperativoCompleto() || this.esGestorVentas(),
  );
  readonly puedeEliminarVentas = computed(() => this.tieneAccesoOperativoCompleto());
  readonly puedeAccederCompras = computed(
    () => this.tieneAccesoOperativoCompleto() || this.esGestorCompras(),
  );
  readonly puedeEliminarCompras = computed(() => this.tieneAccesoOperativoCompleto());
  readonly puedeAccederTerceros = computed(
    () => this.tieneAccesoOperativoCompleto() || this.esGestorVentas(),
  );
  readonly esSoloGestorVentas = computed(
    () => !this.tieneAccesoOperativoCompleto() && this.esGestorVentas(),
  );
  readonly esSoloAsesorComercial = computed(
    () => !this.tieneAccesoOperativoCompleto() && this.esAsesorComercial() && !this.esGestorVentas(),
  );
  readonly puedeAccederOperacion = computed(() => this.tieneAccesoOperativoCompleto());

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

  actualizarUsuarioLocal(parcial: Partial<UsuarioAutenticado>): void {
    const actual = this._usuario();
    if (!actual) return;
    const actualizado: UsuarioAutenticado = { ...actual, ...parcial };
    this._usuario.set(actualizado);
    if (localStorage.getItem(USER_KEY) !== null) {
      localStorage.setItem(USER_KEY, JSON.stringify(actualizado));
    }
    if (sessionStorage.getItem(USER_KEY) !== null) {
      sessionStorage.setItem(USER_KEY, JSON.stringify(actualizado));
    }
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
