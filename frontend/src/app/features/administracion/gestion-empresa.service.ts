import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, map } from 'rxjs';

import { environment } from '../../../environments/environment';
import {
  ActualizarSecuenciaRequest,
  CrearSecuenciaRequest,
  EmpresaDetalle,
  SecuenciaDocumento,
  TipoDocumentoConsecutivo,
  UsuarioKeyDetalle,
} from './gestion-empresa.models';

interface DataListResponse<T> {
  data: T[];
}

@Injectable({ providedIn: 'root' })
export class GestionEmpresaService {
  private readonly http = inject(HttpClient);
  private readonly api = environment.apiUrl;

  obtenerEmpresa(empresaId: number): Observable<EmpresaDetalle> {
    return this.http.get<EmpresaDetalle>(`${this.api}/empresas/${empresaId}`);
  }

  listarTiposDocumento(): Observable<TipoDocumentoConsecutivo[]> {
    return this.http
      .get<DataListResponse<TipoDocumentoConsecutivo>>(`${this.api}/consecutivos/tipos-documento`)
      .pipe(map((r) => r.data ?? []));
  }

  listarSecuencias(): Observable<SecuenciaDocumento[]> {
    return this.http
      .get<DataListResponse<SecuenciaDocumento>>(`${this.api}/consecutivos/secuencias`)
      .pipe(map((r) => r.data ?? []));
  }

  crearSecuencia(payload: CrearSecuenciaRequest): Observable<SecuenciaDocumento> {
    return this.http.post<SecuenciaDocumento>(`${this.api}/consecutivos/secuencias`, payload);
  }

  actualizarSecuencia(
    id: number,
    payload: ActualizarSecuenciaRequest,
  ): Observable<SecuenciaDocumento> {
    return this.http.patch<SecuenciaDocumento>(`${this.api}/consecutivos/secuencias/${id}`, payload);
  }

  obtenerUsuario(usuarioId: number): Observable<UsuarioKeyDetalle> {
    return this.http.get<UsuarioKeyDetalle>(`${this.api}/usuarios/${usuarioId}`);
  }

  regenerarKeyUsuario(usuarioId: number): Observable<UsuarioKeyDetalle> {
    return this.http.post<UsuarioKeyDetalle>(
      `${this.api}/usuarios/${usuarioId}/regenerar-key`,
      {},
    );
  }
}
