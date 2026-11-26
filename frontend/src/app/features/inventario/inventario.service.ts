import { Injectable, inject } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable, map } from 'rxjs';

import { environment } from '../../../environments/environment';
import {
  ActualizarUbicacionRequest,
  CrearMovimientoRequest,
  CrearUbicacionRequest,
  DocumentoInventario,
  ListaMovimientosResponse,
  ListaStockResponse,
  ListarMovimientosFiltros,
  ListarStockFiltros,
  SucursalOption,
  TipoMovimientoInventario,
  Ubicacion,
} from './inventario.models';

interface ListResponse<T> {
  data: T[];
}

@Injectable({ providedIn: 'root' })
export class InventarioService {
  private readonly http = inject(HttpClient);
  private readonly api = `${environment.apiUrl}/inventario`;

  listarSucursales(): Observable<SucursalOption[]> {
    return this.http
      .get<ListResponse<SucursalOption>>(`${this.api}/sucursales`)
      .pipe(map((r) => r.data ?? []));
  }

  listarTiposMovimiento(): Observable<TipoMovimientoInventario[]> {
    return this.http
      .get<ListResponse<TipoMovimientoInventario>>(`${this.api}/tipos-movimiento`)
      .pipe(map((r) => r.data ?? []));
  }

  listarUbicaciones(soloActivas?: boolean, q?: string): Observable<Ubicacion[]> {
    let params = new HttpParams();
    if (soloActivas !== undefined) params = params.set('activo', String(soloActivas));
    if (q?.trim()) params = params.set('q', q.trim());
    return this.http
      .get<ListResponse<Ubicacion>>(`${this.api}/ubicaciones`, { params })
      .pipe(map((r) => r.data ?? []));
  }

  crearUbicacion(payload: CrearUbicacionRequest): Observable<Ubicacion> {
    return this.http.post<Ubicacion>(`${this.api}/ubicaciones`, payload);
  }

  actualizarUbicacion(id: number, payload: ActualizarUbicacionRequest): Observable<Ubicacion> {
    return this.http.patch<Ubicacion>(`${this.api}/ubicaciones/${id}`, payload);
  }

  eliminarUbicacion(id: number): Observable<void> {
    return this.http.delete<void>(`${this.api}/ubicaciones/${id}`);
  }

  listarStock(filtros: ListarStockFiltros = {}): Observable<ListaStockResponse> {
    let params = new HttpParams();
    if (filtros.ubicacion_id) params = params.set('ubicacion_id', String(filtros.ubicacion_id));
    if (filtros.articulo_id) params = params.set('articulo_id', String(filtros.articulo_id));
    if (filtros.solo_con_stock !== undefined) {
      params = params.set('solo_con_stock', String(filtros.solo_con_stock));
    }
    if (filtros.q?.trim()) params = params.set('q', filtros.q.trim());
    if (filtros.page) params = params.set('page', String(filtros.page));
    if (filtros.page_size) params = params.set('page_size', String(filtros.page_size));
    return this.http.get<ListaStockResponse>(`${this.api}/stock`, { params });
  }

  crearMovimiento(payload: CrearMovimientoRequest): Observable<DocumentoInventario> {
    return this.http.post<DocumentoInventario>(`${this.api}/movimientos`, payload);
  }

  listarMovimientos(filtros: ListarMovimientosFiltros = {}): Observable<ListaMovimientosResponse> {
    let params = new HttpParams();
    if (filtros.tipo_movimiento_id) {
      params = params.set('tipo_movimiento_id', String(filtros.tipo_movimiento_id));
    }
    if (filtros.ubicacion_id) params = params.set('ubicacion_id', String(filtros.ubicacion_id));
    if (filtros.articulo_id) params = params.set('articulo_id', String(filtros.articulo_id));
    if (filtros.fecha_desde?.trim()) params = params.set('fecha_desde', filtros.fecha_desde.trim());
    if (filtros.fecha_hasta?.trim()) params = params.set('fecha_hasta', filtros.fecha_hasta.trim());
    if (filtros.q?.trim()) params = params.set('q', filtros.q.trim());
    if (filtros.page) params = params.set('page', String(filtros.page));
    if (filtros.page_size) params = params.set('page_size', String(filtros.page_size));
    return this.http.get<ListaMovimientosResponse>(`${this.api}/movimientos`, { params });
  }

  obtenerMovimiento(id: number): Observable<DocumentoInventario> {
    return this.http.get<DocumentoInventario>(`${this.api}/movimientos/${id}`);
  }
}
