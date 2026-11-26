import { Injectable, inject } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable, map } from 'rxjs';

import { environment } from '../../../environments/environment';
import {
  ActualizarTerceroRequest,
  CrearTerceroRequest,
  ListaTercerosResponse,
  ListarTercerosFiltros,
  RolTercero,
  Tercero,
} from './terceros.models';

interface ListResponse<T> {
  data: T[];
}

@Injectable({ providedIn: 'root' })
export class TercerosService {
  private readonly http = inject(HttpClient);
  private readonly api = `${environment.apiUrl}/terceros`;

  // =====================================================================
  // Catálogo global de roles (solo lectura desde el frontend)
  // =====================================================================

  listarRoles(soloActivos = true): Observable<RolTercero[]> {
    let params = new HttpParams();
    if (soloActivos !== undefined) params = params.set('activo', String(soloActivos));
    return this.http
      .get<ListResponse<RolTercero>>(`${this.api}/roles`, { params })
      .pipe(map((r) => r.data ?? []));
  }

  // =====================================================================
  // Terceros (tercero_maestro + asignación de roles)
  // =====================================================================

  listar(filtros: ListarTercerosFiltros = {}): Observable<ListaTercerosResponse> {
    let params = new HttpParams();
    if (filtros.tipo_identificacion) {
      params = params.set('tipo_identificacion', filtros.tipo_identificacion);
    }
    if (filtros.naturaleza_juridica) {
      params = params.set('naturaleza_juridica', filtros.naturaleza_juridica);
    }
    if (filtros.pais_id) params = params.set('pais_id', String(filtros.pais_id));
    if (filtros.rol_id) params = params.set('rol_id', String(filtros.rol_id));
    if (filtros.activo !== undefined && filtros.activo !== null) {
      params = params.set('activo', String(filtros.activo));
    }
    if (filtros.q?.trim()) params = params.set('q', filtros.q.trim());
    if (filtros.page) params = params.set('page', String(filtros.page));
    if (filtros.page_size) params = params.set('page_size', String(filtros.page_size));

    return this.http.get<ListaTercerosResponse>(this.api, { params });
  }

  obtener(id: number): Observable<Tercero> {
    return this.http.get<Tercero>(`${this.api}/${id}`);
  }

  crear(payload: CrearTerceroRequest): Observable<Tercero> {
    return this.http.post<Tercero>(this.api, payload);
  }

  actualizar(id: number, payload: ActualizarTerceroRequest): Observable<Tercero> {
    return this.http.patch<Tercero>(`${this.api}/${id}`, payload);
  }

  eliminar(id: number): Observable<void> {
    return this.http.delete<void>(`${this.api}/${id}`);
  }
}
