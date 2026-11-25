import { Injectable, inject } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable } from 'rxjs';

import { environment } from '../../../environments/environment';
import {
  ActualizarDetallePrecioRequest,
  ActualizarListaPrecioRequest,
  CrearDetallePrecioRequest,
  CrearListaPrecioRequest,
  DetallePrecio,
  DetallesPrecioPaginadosResponse,
  ListaPrecio,
  ListarDetallesPrecioFiltros,
  ListarListasPreciosFiltros,
  ListasPreciosPaginadasResponse,
} from './precios.models';

@Injectable({ providedIn: 'root' })
export class PreciosService {
  private readonly http = inject(HttpClient);
  private readonly api = `${environment.apiUrl}/listas-precios`;

  // =====================================================================
  // Listas de precios (Cabecera)
  // =====================================================================

  listarListas(filtros: ListarListasPreciosFiltros = {}): Observable<ListasPreciosPaginadasResponse> {
    let params = new HttpParams();
    if (filtros.tipo_lista) params = params.set('tipo_lista', filtros.tipo_lista);
    if (filtros.moneda_id) params = params.set('moneda_id', String(filtros.moneda_id));
    if (filtros.estado !== undefined && filtros.estado !== null) {
      params = params.set('estado', String(filtros.estado));
    }
    if (filtros.q?.trim()) params = params.set('q', filtros.q.trim());
    if (filtros.page) params = params.set('page', String(filtros.page));
    if (filtros.page_size) params = params.set('page_size', String(filtros.page_size));

    return this.http.get<ListasPreciosPaginadasResponse>(this.api, { params });
  }

  obtenerLista(id: number): Observable<ListaPrecio> {
    return this.http.get<ListaPrecio>(`${this.api}/${id}`);
  }

  crearLista(payload: CrearListaPrecioRequest): Observable<ListaPrecio> {
    return this.http.post<ListaPrecio>(this.api, payload);
  }

  actualizarLista(id: number, payload: ActualizarListaPrecioRequest): Observable<ListaPrecio> {
    return this.http.patch<ListaPrecio>(`${this.api}/${id}`, payload);
  }

  eliminarLista(id: number): Observable<void> {
    return this.http.delete<void>(`${this.api}/${id}`);
  }

  // =====================================================================
  // Detalles de lista de precios (lista_precios_detalle)
  // =====================================================================

  listarDetalles(
    listaPrecioId: number,
    filtros: ListarDetallesPrecioFiltros = {},
  ): Observable<DetallesPrecioPaginadosResponse> {
    let params = new HttpParams();
    if (filtros.articulo_id) params = params.set('articulo_id', String(filtros.articulo_id));
    if (filtros.unidad_medida_id) {
      params = params.set('unidad_medida_id', String(filtros.unidad_medida_id));
    }
    if (filtros.q?.trim()) params = params.set('q', filtros.q.trim());
    if (filtros.page) params = params.set('page', String(filtros.page));
    if (filtros.page_size) params = params.set('page_size', String(filtros.page_size));

    return this.http.get<DetallesPrecioPaginadosResponse>(`${this.api}/${listaPrecioId}/detalles`, {
      params,
    });
  }

  crearDetalle(
    listaPrecioId: number,
    payload: CrearDetallePrecioRequest,
  ): Observable<DetallePrecio> {
    return this.http.post<DetallePrecio>(`${this.api}/${listaPrecioId}/detalles`, payload);
  }

  actualizarDetalle(
    listaPrecioId: number,
    detalleId: number,
    payload: ActualizarDetallePrecioRequest,
  ): Observable<DetallePrecio> {
    return this.http.patch<DetallePrecio>(
      `${this.api}/${listaPrecioId}/detalles/${detalleId}`,
      payload,
    );
  }

  eliminarDetalle(listaPrecioId: number, detalleId: number): Observable<void> {
    return this.http.delete<void>(`${this.api}/${listaPrecioId}/detalles/${detalleId}`);
  }
}
