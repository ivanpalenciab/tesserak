import { Injectable, inject } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable } from 'rxjs';

import { environment } from '../../../environments/environment';
import {
  ActualizarVentaRequest,
  CrearVentaRequest,
  ListaVentasResponse,
  ListarVentasFiltros,
  Venta,
} from './ventas.models';

@Injectable({ providedIn: 'root' })
export class VentasService {
  private readonly http = inject(HttpClient);
  private readonly api = `${environment.apiUrl}/ventas`;

  listar(filtros: ListarVentasFiltros = {}): Observable<ListaVentasResponse> {
    let params = new HttpParams();
    if (filtros.estado) params = params.set('estado', filtros.estado);
    if (filtros.cliente_id) params = params.set('cliente_id', String(filtros.cliente_id));
    if (filtros.ubicacion_id) params = params.set('ubicacion_id', String(filtros.ubicacion_id));
    if (filtros.fecha_desde?.trim()) params = params.set('fecha_desde', filtros.fecha_desde.trim());
    if (filtros.fecha_hasta?.trim()) params = params.set('fecha_hasta', filtros.fecha_hasta.trim());
    if (filtros.q?.trim()) params = params.set('q', filtros.q.trim());
    if (filtros.page) params = params.set('page', String(filtros.page));
    if (filtros.page_size) params = params.set('page_size', String(filtros.page_size));
    return this.http.get<ListaVentasResponse>(this.api, { params });
  }

  obtener(id: number): Observable<Venta> {
    return this.http.get<Venta>(`${this.api}/${id}`);
  }

  crear(payload: CrearVentaRequest): Observable<Venta> {
    return this.http.post<Venta>(this.api, payload);
  }

  actualizar(id: number, payload: ActualizarVentaRequest): Observable<Venta> {
    return this.http.patch<Venta>(`${this.api}/${id}`, payload);
  }

  eliminar(id: number): Observable<void> {
    return this.http.delete<void>(`${this.api}/${id}`);
  }
}
