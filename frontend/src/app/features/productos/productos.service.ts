import { Injectable, inject } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable, map } from 'rxjs';

import { environment } from '../../../environments/environment';
import {
  ActualizarCategoriaRequest,
  ActualizarGrupoImpuestoRequest,
  ActualizarMarcaRequest,
  ActualizarProductoRequest,
  CategoriaProducto,
  CrearCategoriaRequest,
  CrearGrupoImpuestoRequest,
  CrearMarcaRequest,
  CrearProductoRequest,
  GrupoImpuesto,
  ListaProductosResponse,
  ListarProductosFiltros,
  MarcaProducto,
  MetodoValoracion,
  ModalidadServicio,
  Producto,
  TipoArticulo,
  UnidadMedida,
} from './productos.models';

interface CatalogoListResponse<T> {
  data: T[];
}

@Injectable({ providedIn: 'root' })
export class ProductosService {
  private readonly http = inject(HttpClient);
  private readonly api = `${environment.apiUrl}/productos`;

  // =====================================================================
  // Catálogos auxiliares
  // =====================================================================

  listarTiposArticulo(): Observable<TipoArticulo[]> {
    return this.http
      .get<CatalogoListResponse<TipoArticulo>>(`${this.api}/tipos-articulo`)
      .pipe(map((r) => r.data ?? []));
  }

  listarMetodosValoracion(): Observable<MetodoValoracion[]> {
    return this.http
      .get<CatalogoListResponse<MetodoValoracion>>(`${this.api}/metodos-valoracion`)
      .pipe(map((r) => r.data ?? []));
  }

  listarModalidadesServicio(): Observable<ModalidadServicio[]> {
    return this.http
      .get<CatalogoListResponse<ModalidadServicio>>(`${this.api}/modalidades-servicio`)
      .pipe(map((r) => r.data ?? []));
  }

  listarGruposImpuesto(soloActivos?: boolean, q?: string): Observable<GrupoImpuesto[]> {
    let params = new HttpParams();
    if (soloActivos !== undefined) params = params.set('activo', String(soloActivos));
    if (q?.trim()) params = params.set('q', q.trim());
    return this.http
      .get<CatalogoListResponse<GrupoImpuesto>>(`${this.api}/grupos-impuesto`, { params })
      .pipe(map((r) => r.data ?? []));
  }

  crearGrupoImpuesto(payload: CrearGrupoImpuestoRequest): Observable<GrupoImpuesto> {
    return this.http.post<GrupoImpuesto>(`${this.api}/grupos-impuesto`, payload);
  }

  actualizarGrupoImpuesto(
    id: number,
    payload: ActualizarGrupoImpuestoRequest,
  ): Observable<GrupoImpuesto> {
    return this.http.patch<GrupoImpuesto>(`${this.api}/grupos-impuesto/${id}`, payload);
  }

  eliminarGrupoImpuesto(id: number): Observable<void> {
    return this.http.delete<void>(`${this.api}/grupos-impuesto/${id}`);
  }

  listarUnidadesMedida(): Observable<UnidadMedida[]> {
    return this.http
      .get<CatalogoListResponse<UnidadMedida>>(`${this.api}/unidades-medida`)
      .pipe(map((r) => r.data ?? []));
  }

  listarMarcas(soloActivas?: boolean, q?: string): Observable<MarcaProducto[]> {
    let params = new HttpParams();
    if (soloActivas !== undefined) params = params.set('activo', String(soloActivas));
    if (q?.trim()) params = params.set('q', q.trim());
    return this.http
      .get<CatalogoListResponse<MarcaProducto>>(`${this.api}/marcas`, { params })
      .pipe(map((r) => r.data ?? []));
  }

  crearMarca(payload: CrearMarcaRequest): Observable<MarcaProducto> {
    return this.http.post<MarcaProducto>(`${this.api}/marcas`, payload);
  }

  actualizarMarca(id: number, payload: ActualizarMarcaRequest): Observable<MarcaProducto> {
    return this.http.patch<MarcaProducto>(`${this.api}/marcas/${id}`, payload);
  }

  eliminarMarca(id: number): Observable<void> {
    return this.http.delete<void>(`${this.api}/marcas/${id}`);
  }

  // =====================================================================
  // Categorías (CRUD con soft delete)
  // =====================================================================

  listarCategorias(soloActivas?: boolean, q?: string): Observable<CategoriaProducto[]> {
    let params = new HttpParams();
    if (soloActivas !== undefined) params = params.set('activo', String(soloActivas));
    if (q?.trim()) params = params.set('q', q.trim());
    return this.http
      .get<CatalogoListResponse<CategoriaProducto>>(`${this.api}/categorias`, { params })
      .pipe(map((r) => r.data ?? []));
  }

  crearCategoria(payload: CrearCategoriaRequest): Observable<CategoriaProducto> {
    return this.http.post<CategoriaProducto>(`${this.api}/categorias`, payload);
  }

  actualizarCategoria(id: number, payload: ActualizarCategoriaRequest): Observable<CategoriaProducto> {
    return this.http.patch<CategoriaProducto>(`${this.api}/categorias/${id}`, payload);
  }

  eliminarCategoria(id: number): Observable<void> {
    return this.http.delete<void>(`${this.api}/categorias/${id}`);
  }

  // =====================================================================
  // Artículos (Productos y Servicios con soft delete)
  // =====================================================================

  listar(filtros: ListarProductosFiltros = {}): Observable<ListaProductosResponse> {
    let params = new HttpParams();
    if (filtros.tipo_articulo_id) {
      params = params.set('tipo_articulo_id', String(filtros.tipo_articulo_id));
    }
    if (filtros.estado !== undefined && filtros.estado !== null) {
      params = params.set('estado', String(filtros.estado));
    }
    if (filtros.categoria_id) params = params.set('categoria_id', String(filtros.categoria_id));
    if (filtros.marca_id) params = params.set('marca_id', String(filtros.marca_id));
    if (filtros.unidad_medida_id) {
      params = params.set('unidad_medida_id', String(filtros.unidad_medida_id));
    }
    if (filtros.q?.trim()) params = params.set('q', filtros.q.trim());
    if (filtros.page) params = params.set('page', String(filtros.page));
    if (filtros.page_size) params = params.set('page_size', String(filtros.page_size));

    return this.http.get<ListaProductosResponse>(this.api, { params });
  }

  crear(payload: CrearProductoRequest): Observable<Producto> {
    return this.http.post<Producto>(this.api, payload);
  }

  actualizar(id: number, payload: ActualizarProductoRequest): Observable<Producto> {
    return this.http.patch<Producto>(`${this.api}/${id}`, payload);
  }

  eliminar(id: number): Observable<void> {
    return this.http.delete<void>(`${this.api}/${id}`);
  }
}
