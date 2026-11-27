import { Injectable, inject } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable, map } from 'rxjs';

import { environment } from '../../../environments/environment';
import {
  ActualizarMiembroPayload,
  AcumularPuntosPayload,
  ClienteDisponibleLealtad,
  CuentaClienteLealtad,
  CuentaLealtad,
  EstadoCuentaLealtad,
  InscribirDesdeTerceroPayload,
  MiembroLealtad,
  PaisLealtad,
  ProgramaOperacionLealtad,
  RedencionLealtad,
  RedimirPuntosPayload,
  RegistrarMiembroPayload,
  TransaccionLealtad,
} from './lealtad.models';

export interface FiltroCuentasLealtad {
  programa_id?: number | null;
  estado?: string;
  search?: string;
}

@Injectable({ providedIn: 'root' })
export class LealtadOperacionService {
  private readonly http = inject(HttpClient);
  private readonly baseUrl = `${environment.apiUrl}/lealtad`;

  listarPaises(): Observable<PaisLealtad[]> {
    return this.http
      .get<{ data: PaisLealtad[] }>(`${this.baseUrl}/operacion/paises`)
      .pipe(map((res) => res.data ?? []));
  }

  listarProgramasAdheridos(): Observable<ProgramaOperacionLealtad[]> {
    return this.http
      .get<{ data: ProgramaOperacionLealtad[] }>(`${this.baseUrl}/operacion/programas`)
      .pipe(map((res) => res.data ?? []));
  }

  listarClientesDisponibles(search = ''): Observable<ClienteDisponibleLealtad[]> {
    let params = new HttpParams();
    if (search.trim()) params = params.set('search', search.trim());
    return this.http
      .get<{ data: ClienteDisponibleLealtad[] }>(`${this.baseUrl}/operacion/clientes`, { params })
      .pipe(map((res) => res.data ?? []));
  }

  buscarMiembroPorDocumento(
    paisId: number,
    tipoIdentificacion: string,
    numeroIdentificacion: string,
  ): Observable<MiembroLealtad> {
    const params = new HttpParams()
      .set('pais_id', String(paisId))
      .set('tipo_identificacion', tipoIdentificacion)
      .set('numero_identificacion', numeroIdentificacion.trim());
    return this.http.get<MiembroLealtad>(`${this.baseUrl}/miembros/buscar`, { params });
  }

  registrarMiembro(payload: RegistrarMiembroPayload): Observable<CuentaLealtad> {
    return this.http.post<CuentaLealtad>(`${this.baseUrl}/miembros`, payload);
  }

  inscribirDesdeTercero(payload: InscribirDesdeTerceroPayload): Observable<CuentaLealtad> {
    return this.http.post<CuentaLealtad>(`${this.baseUrl}/miembros/desde-tercero`, payload);
  }

  actualizarMiembro(id: number, payload: ActualizarMiembroPayload): Observable<MiembroLealtad> {
    return this.http.patch<MiembroLealtad>(`${this.baseUrl}/miembros/${id}`, payload);
  }

  listarCuentas(filtro?: FiltroCuentasLealtad): Observable<CuentaLealtad[]> {
    let params = new HttpParams();
    if (filtro?.programa_id) params = params.set('programa_id', String(filtro.programa_id));
    if (filtro?.estado) params = params.set('estado', filtro.estado);
    if (filtro?.search?.trim()) params = params.set('search', filtro.search.trim());
    return this.http
      .get<{ data: CuentaLealtad[] }>(`${this.baseUrl}/cuentas`, { params })
      .pipe(map((res) => res.data ?? []));
  }

  actualizarEstadoCuenta(cuentaId: number, estado: EstadoCuentaLealtad): Observable<CuentaLealtad> {
    return this.http.patch<CuentaLealtad>(`${this.baseUrl}/cuentas/${cuentaId}/estado`, { estado });
  }

  listarTransacciones(cuentaId: number): Observable<TransaccionLealtad[]> {
    return this.http
      .get<{ data: TransaccionLealtad[] }>(`${this.baseUrl}/cuentas/${cuentaId}/transacciones`)
      .pipe(map((res) => res.data ?? []));
  }

  acumularPuntos(cuentaId: number, payload: AcumularPuntosPayload): Observable<TransaccionLealtad> {
    return this.http.post<TransaccionLealtad>(`${this.baseUrl}/cuentas/${cuentaId}/acumular`, payload);
  }

  redimirPuntos(cuentaId: number, payload: RedimirPuntosPayload): Observable<RedencionLealtad> {
    return this.http.post<RedencionLealtad>(`${this.baseUrl}/cuentas/${cuentaId}/redimir`, payload);
  }

  consultarCuentaCliente(terceroId: number): Observable<CuentaClienteLealtad> {
    return this.http.get<CuentaClienteLealtad>(`${this.baseUrl}/operacion/cuenta-cliente/${terceroId}`);
  }
}
