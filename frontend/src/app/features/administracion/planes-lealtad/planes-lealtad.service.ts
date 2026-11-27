import { Injectable, inject } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable, map } from 'rxjs';

import { environment } from '../../../../environments/environment';
import {
  ActualizarProgramaPayload,
  ActualizarReglaPayload,
  AdhesionPayload,
  AdhesionPrograma,
  ConciliacionProgramaResponse,
  CrearProgramaPayload,
  CrearReglaPayload,
  ListProgramasResponse,
  ListarReglasResponse,
  MonedaLealtad,
  ProcesarVencimientosResponse,
  ProgramaLealtad,
  ReglaLealtad,
} from './planes-lealtad.models';

export interface FiltroProgramas {
  search?: string;
  alcance?: string;
  estado?: string;
  solo_adheridos?: boolean;
}

@Injectable({ providedIn: 'root' })
export class PlanesLealtadService {
  private readonly http = inject(HttpClient);
  private readonly baseUrl = `${environment.apiUrl}/lealtad`;

  listarMonedas(): Observable<MonedaLealtad[]> {
    return this.http
      .get<{ data: MonedaLealtad[] }>(`${this.baseUrl}/monedas`)
      .pipe(map((res) => res.data ?? []));
  }

  listarProgramas(filtro?: FiltroProgramas): Observable<ListProgramasResponse> {
    let params = new HttpParams();
    if (filtro?.search?.trim()) params = params.set('search', filtro.search.trim());
    if (filtro?.alcance) params = params.set('alcance', filtro.alcance);
    if (filtro?.estado) params = params.set('estado', filtro.estado);
    if (filtro?.solo_adheridos) params = params.set('solo_adheridos', 'true');

    return this.http.get<ListProgramasResponse>(`${this.baseUrl}/programas`, { params }).pipe(
      map((res) => ({
        data: res.data ?? [],
        es_admin_global: Boolean(res.es_admin_global),
      })),
    );
  }

  crearPrograma(payload: CrearProgramaPayload): Observable<ProgramaLealtad> {
    return this.http.post<ProgramaLealtad>(`${this.baseUrl}/programas`, payload);
  }

  actualizarPrograma(id: number, payload: ActualizarProgramaPayload): Observable<ProgramaLealtad> {
    return this.http.patch<ProgramaLealtad>(`${this.baseUrl}/programas/${id}`, payload);
  }

  eliminarPrograma(id: number): Observable<void> {
    return this.http.delete<void>(`${this.baseUrl}/programas/${id}`);
  }

  adherirEmpresa(programaId: number, payload: AdhesionPayload): Observable<AdhesionPrograma> {
    return this.http.post<AdhesionPrograma>(`${this.baseUrl}/programas/${programaId}/adherir`, payload);
  }

  actualizarAdhesion(programaId: number, payload: AdhesionPayload): Observable<AdhesionPrograma> {
    return this.http.patch<AdhesionPrograma>(`${this.baseUrl}/programas/${programaId}/adhesion`, payload);
  }

  desadherirEmpresa(programaId: number): Observable<void> {
    return this.http.delete<void>(`${this.baseUrl}/programas/${programaId}/adherir`);
  }

  listarEmpresasAdheridas(programaId: number): Observable<AdhesionPrograma[]> {
    return this.http
      .get<{ data: AdhesionPrograma[] }>(`${this.baseUrl}/programas/${programaId}/empresas`)
      .pipe(map((res) => res.data ?? []));
  }

  listarReglas(programaId: number): Observable<ListarReglasResponse> {
    return this.http.get<ListarReglasResponse>(`${this.baseUrl}/programas/${programaId}/reglas`).pipe(
      map((res) => ({
        data: res.data ?? [],
        puede_crear_global: Boolean(res.puede_crear_global),
        puede_crear_empresa: Boolean(res.puede_crear_empresa),
      })),
    );
  }

  crearRegla(programaId: number, payload: CrearReglaPayload): Observable<ReglaLealtad> {
    return this.http.post<ReglaLealtad>(`${this.baseUrl}/programas/${programaId}/reglas`, payload);
  }

  actualizarRegla(programaId: number, reglaId: number, payload: ActualizarReglaPayload): Observable<ReglaLealtad> {
    return this.http.patch<ReglaLealtad>(
      `${this.baseUrl}/programas/${programaId}/reglas/${reglaId}`,
      payload,
    );
  }

  eliminarRegla(programaId: number, reglaId: number): Observable<void> {
    return this.http.delete<void>(`${this.baseUrl}/programas/${programaId}/reglas/${reglaId}`);
  }

  obtenerConciliacion(
    programaId: number,
    fechaDesde = '',
    fechaHasta = '',
  ): Observable<ConciliacionProgramaResponse> {
    let params = new HttpParams();
    if (fechaDesde) params = params.set('fecha_desde', fechaDesde);
    if (fechaHasta) params = params.set('fecha_hasta', fechaHasta);
    return this.http.get<ConciliacionProgramaResponse>(
      `${this.baseUrl}/programas/${programaId}/conciliacion`,
      { params },
    );
  }

  procesarVencimientos(programaId: number): Observable<ProcesarVencimientosResponse> {
    return this.http.post<ProcesarVencimientosResponse>(
      `${this.baseUrl}/programas/${programaId}/procesar-vencimientos`,
      {},
    );
  }
}
