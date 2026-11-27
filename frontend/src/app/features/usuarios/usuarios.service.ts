import { Injectable, inject } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable, map } from 'rxjs';

import { environment } from '../../../environments/environment';
import {
  ActualizarPerfilRequest,
  ActualizarUsuarioRequest,
  CrearUsuarioRequest,
  FiltrosUsuarios,
  ListaUsuariosResponse,
  RolSistema,
  SucursalOpcion,
  UsuarioItem,
} from './usuarios.models';

interface DataListResponse<T> {
  data: T[];
}

@Injectable({ providedIn: 'root' })
export class UsuariosService {
  private readonly http = inject(HttpClient);
  private readonly baseUrl = `${environment.apiUrl}/usuarios`;

  listarRoles(): Observable<RolSistema[]> {
    return this.http
      .get<DataListResponse<RolSistema>>(`${this.baseUrl}/roles`)
      .pipe(map((r) => r.data ?? []));
  }

  listarSucursales(): Observable<SucursalOpcion[]> {
    return this.http
      .get<DataListResponse<SucursalOpcion>>(`${this.baseUrl}/sucursales`)
      .pipe(map((r) => r.data ?? []));
  }

  listar(filtros: FiltrosUsuarios = {}): Observable<ListaUsuariosResponse> {
    let params = new HttpParams();
    if (filtros.sucursal_id) params = params.set('sucursal_id', String(filtros.sucursal_id));
    if (filtros.activo !== undefined) params = params.set('activo', String(filtros.activo));
    if (filtros.q?.trim()) params = params.set('q', filtros.q.trim());
    if (filtros.page) params = params.set('page', String(filtros.page));
    if (filtros.page_size) params = params.set('page_size', String(filtros.page_size));
    return this.http.get<ListaUsuariosResponse>(this.baseUrl, { params });
  }

  crear(payload: CrearUsuarioRequest): Observable<UsuarioItem> {
    return this.http.post<UsuarioItem>(this.baseUrl, payload);
  }

  actualizar(id: number, payload: ActualizarUsuarioRequest): Observable<UsuarioItem> {
    return this.http.patch<UsuarioItem>(`${this.baseUrl}/${id}`, payload);
  }

  eliminar(id: number): Observable<void> {
    return this.http.delete<void>(`${this.baseUrl}/${id}`);
  }

  obtenerPorId(id: number): Observable<UsuarioItem> {
    return this.http.get<UsuarioItem>(`${this.baseUrl}/${id}`);
  }

  obtenerMiPerfil(): Observable<UsuarioItem> {
    return this.http.get<UsuarioItem>(`${this.baseUrl}/mi-perfil`);
  }

  actualizarMiPerfil(payload: ActualizarPerfilRequest): Observable<UsuarioItem> {
    return this.http.patch<UsuarioItem>(`${this.baseUrl}/mi-perfil`, payload);
  }
}
