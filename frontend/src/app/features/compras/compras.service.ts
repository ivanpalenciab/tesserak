import { Injectable, inject } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable } from 'rxjs';

import { environment } from '../../../environments/environment';
import {
  ActualizarCompraRequest,
  Compra,
  CrearCompraRequest,
  ListaComprasResponse,
  ListarComprasFiltros,
} from './compras.models';

@Injectable({ providedIn: 'root' })
export class ComprasService {
  private readonly http = inject(HttpClient);
  private readonly api = `${environment.apiUrl}/compras`;

  listar(filtros: ListarComprasFiltros = {}): Observable<ListaComprasResponse> {
    let params = new HttpParams();
    if (filtros.sucursal_id) params = params.set('sucursal_id', String(filtros.sucursal_id));
    if (filtros.estado) params = params.set('estado', filtros.estado);
    if (filtros.moneda_id) params = params.set('moneda_id', String(filtros.moneda_id));
    if (filtros.fecha_desde?.trim()) params = params.set('fecha_desde', filtros.fecha_desde.trim());
    if (filtros.fecha_hasta?.trim()) params = params.set('fecha_hasta', filtros.fecha_hasta.trim());
    if (filtros.q?.trim()) params = params.set('q', filtros.q.trim());
    if (filtros.page) params = params.set('page', String(filtros.page));
    if (filtros.page_size) params = params.set('page_size', String(filtros.page_size));

    return this.http.get<ListaComprasResponse>(this.api, { params });
  }

  obtener(id: number): Observable<Compra> {
    return this.http.get<Compra>(`${this.api}/${id}`);
  }

  crear(payload: CrearCompraRequest): Observable<Compra> {
    return this.http.post<Compra>(this.api, payload);
  }

  actualizar(id: number, payload: ActualizarCompraRequest): Observable<Compra> {
    return this.http.patch<Compra>(`${this.api}/${id}`, payload);
  }

  eliminar(id: number): Observable<void> {
    return this.http.delete<void>(`${this.api}/${id}`);
  }
}
