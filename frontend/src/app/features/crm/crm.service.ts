import { Injectable, inject } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable, map } from 'rxjs';

import { environment } from '../../../environments/environment';
import {
  ActualizarCampanaRequest,
  ActualizarEstadoLeadRequest,
  ActualizarEtapaOportunidadRequest,
  ActualizarGastoCampanaRequest,
  ActualizarLeadRequest,
  ActualizarMotivoDescarteRequest,
  ActualizarOrigenLeadRequest,
  ActualizarTipoCampanaRequest,
  AsociarLeadCampanaRequest,
  CambiarEstadoLeadRequest,
  Campana,
  CampanaAsociadaLead,
  ConvertirLeadRequest,
  CrearCampanaRequest,
  CrearEstadoLeadRequest,
  CrearEtapaOportunidadRequest,
  CrearGastoCampanaRequest,
  CrearLeadRequest,
  CrearMotivoDescarteRequest,
  CrearOrigenLeadRequest,
  CrearTipoCampanaRequest,
  EstadoCampana,
  EstadoLead,
  EtapaOportunidad,
  GastoCampana,
  IndicadoresCampana,
  Lead,
  ListaCampanasResponse,
  ListaGastosCampanaResponse,
  ListaLeadsCampanaResponse,
  ListaLeadsResponse,
  ListarCampanasFiltros,
  ListarLeadsFiltros,
  MotivoDescarteLead,
  OrigenLead,
  TipoCampana,
} from './crm.models';

interface CatalogoResponse<T> {
  data: T[];
}

@Injectable({ providedIn: 'root' })
export class CrmService {
  private readonly http = inject(HttpClient);
  private readonly api = `${environment.apiUrl}/crm`;

  listarOrigenes(empresaId: number, soloActivos?: boolean): Observable<OrigenLead[]> {
    let params = new HttpParams().set('empresa_id', String(empresaId));
    if (soloActivos !== undefined) {
      params = params.set('activo', String(soloActivos));
    }
    return this.http
      .get<CatalogoResponse<OrigenLead>>(`${this.api}/origenes`, { params })
      .pipe(map((r) => r.data ?? []));
  }

  crearOrigen(payload: CrearOrigenLeadRequest): Observable<OrigenLead> {
    return this.http.post<OrigenLead>(`${this.api}/origenes`, payload);
  }

  actualizarOrigen(id: number, payload: ActualizarOrigenLeadRequest): Observable<OrigenLead> {
    return this.http.patch<OrigenLead>(`${this.api}/origenes/${id}`, payload);
  }

  eliminarOrigen(id: number): Observable<void> {
    return this.http.delete<void>(`${this.api}/origenes/${id}`);
  }

  listarEstados(empresaId: number): Observable<EstadoLead[]> {
    const params = new HttpParams().set('empresa_id', String(empresaId));
    return this.http
      .get<CatalogoResponse<EstadoLead>>(`${this.api}/estados`, { params })
      .pipe(map((r) => r.data ?? []));
  }

  crearEstado(payload: CrearEstadoLeadRequest): Observable<EstadoLead> {
    return this.http.post<EstadoLead>(`${this.api}/estados`, payload);
  }

  actualizarEstado(id: number, payload: ActualizarEstadoLeadRequest): Observable<EstadoLead> {
    return this.http.patch<EstadoLead>(`${this.api}/estados/${id}`, payload);
  }

  eliminarEstado(id: number): Observable<void> {
    return this.http.delete<void>(`${this.api}/estados/${id}`);
  }

  listarEtapasOportunidad(empresaId: number, soloActivos?: boolean): Observable<EtapaOportunidad[]> {
    let params = new HttpParams().set('empresa_id', String(empresaId));
    if (soloActivos !== undefined) {
      params = params.set('activo', String(soloActivos));
    }
    return this.http
      .get<CatalogoResponse<EtapaOportunidad>>(`${this.api}/etapas-oportunidad`, { params })
      .pipe(map((r) => r.data ?? []));
  }

  crearEtapaOportunidad(payload: CrearEtapaOportunidadRequest): Observable<EtapaOportunidad> {
    return this.http.post<EtapaOportunidad>(`${this.api}/etapas-oportunidad`, payload);
  }

  actualizarEtapaOportunidad(
    id: number,
    payload: ActualizarEtapaOportunidadRequest,
  ): Observable<EtapaOportunidad> {
    return this.http.patch<EtapaOportunidad>(`${this.api}/etapas-oportunidad/${id}`, payload);
  }

  eliminarEtapaOportunidad(id: number): Observable<void> {
    return this.http.delete<void>(`${this.api}/etapas-oportunidad/${id}`);
  }

  listarMotivosDescarte(empresaId: number, soloActivos?: boolean): Observable<MotivoDescarteLead[]> {
    let params = new HttpParams().set('empresa_id', String(empresaId));
    if (soloActivos !== undefined) {
      params = params.set('activo', String(soloActivos));
    }
    return this.http
      .get<CatalogoResponse<MotivoDescarteLead>>(`${this.api}/motivos-descarte`, { params })
      .pipe(map((r) => r.data ?? []));
  }

  crearMotivoDescarte(payload: CrearMotivoDescarteRequest): Observable<MotivoDescarteLead> {
    return this.http.post<MotivoDescarteLead>(`${this.api}/motivos-descarte`, payload);
  }

  actualizarMotivoDescarte(
    id: number,
    payload: ActualizarMotivoDescarteRequest,
  ): Observable<MotivoDescarteLead> {
    return this.http.patch<MotivoDescarteLead>(`${this.api}/motivos-descarte/${id}`, payload);
  }

  eliminarMotivoDescarte(id: number): Observable<void> {
    return this.http.delete<void>(`${this.api}/motivos-descarte/${id}`);
  }

  listarLeads(filtros: ListarLeadsFiltros): Observable<ListaLeadsResponse> {
    let params = new HttpParams().set('empresa_id', String(filtros.empresa_id));

    if (filtros.sucursal_id) params = params.set('sucursal_id', String(filtros.sucursal_id));
    if (filtros.origen_id) params = params.set('origen_id', String(filtros.origen_id));
    if (filtros.estado_id) params = params.set('estado_id', String(filtros.estado_id));
    if (filtros.temperatura) params = params.set('temperatura', filtros.temperatura);
    if (filtros.convertido !== undefined && filtros.convertido !== null) {
      params = params.set('convertido', String(filtros.convertido));
    }
    if (filtros.q?.trim()) params = params.set('q', filtros.q.trim());
    if (filtros.page) params = params.set('page', String(filtros.page));
    if (filtros.page_size) params = params.set('page_size', String(filtros.page_size));

    return this.http.get<ListaLeadsResponse>(`${this.api}/leads`, { params });
  }

  crearLead(payload: CrearLeadRequest): Observable<Lead> {
    return this.http.post<Lead>(`${this.api}/leads`, payload);
  }

  actualizarLead(id: number, payload: ActualizarLeadRequest): Observable<Lead> {
    return this.http.patch<Lead>(`${this.api}/leads/${id}`, payload);
  }

  cambiarEstadoLead(id: number, payload: CambiarEstadoLeadRequest): Observable<Lead> {
    return this.http.patch<Lead>(`${this.api}/leads/${id}/estado`, payload);
  }

  convertirLead(id: number, payload: ConvertirLeadRequest = {}): Observable<Lead> {
    return this.http.post<Lead>(`${this.api}/leads/${id}/convertir`, payload);
  }

  eliminarLead(id: number): Observable<void> {
    return this.http.delete<void>(`${this.api}/leads/${id}`);
  }

  // =====================================================================
  // Tipos de campaña
  // =====================================================================

  listarTiposCampana(soloActivos?: boolean): Observable<TipoCampana[]> {
    let params = new HttpParams();
    if (soloActivos !== undefined) {
      params = params.set('activo', String(soloActivos));
    }
    return this.http
      .get<CatalogoResponse<TipoCampana>>(`${this.api}/tipos-campana`, { params })
      .pipe(map((r) => r.data ?? []));
  }

  crearTipoCampana(payload: CrearTipoCampanaRequest): Observable<TipoCampana> {
    return this.http.post<TipoCampana>(`${this.api}/tipos-campana`, payload);
  }

  actualizarTipoCampana(id: number, payload: ActualizarTipoCampanaRequest): Observable<TipoCampana> {
    return this.http.patch<TipoCampana>(`${this.api}/tipos-campana/${id}`, payload);
  }

  eliminarTipoCampana(id: number): Observable<void> {
    return this.http.delete<void>(`${this.api}/tipos-campana/${id}`);
  }

  // =====================================================================
  // Campañas
  // =====================================================================

  listarCampanas(pageSize = 100): Observable<Campana[]> {
    const params = new HttpParams().set('page_size', String(pageSize));
    return this.http
      .get<ListaCampanasResponse>(`${this.api}/campanas`, { params })
      .pipe(map((r) => r.data ?? []));
  }

  listarCampanasPaginado(filtros: ListarCampanasFiltros): Observable<ListaCampanasResponse> {
    let params = new HttpParams();
    if (filtros.tipo_campana_id) params = params.set('tipo_campana_id', String(filtros.tipo_campana_id));
    if (filtros.origen_id) params = params.set('origen_id', String(filtros.origen_id));
    if (filtros.estado) params = params.set('estado', filtros.estado);
    if (filtros.fecha_desde) params = params.set('fecha_desde', filtros.fecha_desde);
    if (filtros.fecha_hasta) params = params.set('fecha_hasta', filtros.fecha_hasta);
    if (filtros.q?.trim()) params = params.set('q', filtros.q.trim());
    if (filtros.page) params = params.set('page', String(filtros.page));
    if (filtros.page_size) params = params.set('page_size', String(filtros.page_size));

    return this.http.get<ListaCampanasResponse>(`${this.api}/campanas`, { params });
  }

  crearCampana(payload: CrearCampanaRequest): Observable<Campana> {
    return this.http.post<Campana>(`${this.api}/campanas`, payload);
  }

  actualizarCampana(id: number, payload: ActualizarCampanaRequest): Observable<Campana> {
    return this.http.patch<Campana>(`${this.api}/campanas/${id}`, payload);
  }

  cambiarEstadoCampana(id: number, estado: EstadoCampana): Observable<Campana> {
    return this.http.patch<Campana>(`${this.api}/campanas/${id}/estado`, { estado });
  }

  obtenerIndicadoresCampana(id: number): Observable<IndicadoresCampana> {
    return this.http.get<IndicadoresCampana>(`${this.api}/campanas/${id}/indicadores`);
  }

  // =====================================================================
  // Campaña - Leads
  // =====================================================================

  asociarLeadACampana(campanaId: number, payload: AsociarLeadCampanaRequest): Observable<unknown> {
    return this.http.post(`${this.api}/campanas/${campanaId}/leads`, payload);
  }

  listarLeadsDeCampana(campanaId: number, page = 1, pageSize = 20): Observable<ListaLeadsCampanaResponse> {
    const params = new HttpParams().set('page', String(page)).set('page_size', String(pageSize));
    return this.http.get<ListaLeadsCampanaResponse>(`${this.api}/campanas/${campanaId}/leads`, {
      params,
    });
  }

  desasociarLeadDeCampana(campanaId: number, leadId: number): Observable<void> {
    return this.http.delete<void>(`${this.api}/campanas/${campanaId}/leads/${leadId}`);
  }

  listarCampanasDeLead(leadId: number): Observable<CampanaAsociadaLead[]> {
    return this.http
      .get<CatalogoResponse<CampanaAsociadaLead>>(`${this.api}/leads/${leadId}/campanas`)
      .pipe(map((r) => r.data ?? []));
  }

  // =====================================================================
  // Gastos de campaña
  // =====================================================================

  listarGastosCampana(campanaId: number, page = 1, pageSize = 20): Observable<ListaGastosCampanaResponse> {
    const params = new HttpParams().set('page', String(page)).set('page_size', String(pageSize));
    return this.http.get<ListaGastosCampanaResponse>(`${this.api}/campanas/${campanaId}/gastos`, {
      params,
    });
  }

  crearGastoCampana(campanaId: number, payload: CrearGastoCampanaRequest): Observable<GastoCampana> {
    return this.http.post<GastoCampana>(`${this.api}/campanas/${campanaId}/gastos`, payload);
  }

  actualizarGastoCampana(id: number, payload: ActualizarGastoCampanaRequest): Observable<GastoCampana> {
    return this.http.patch<GastoCampana>(`${this.api}/gastos/${id}`, payload);
  }

  eliminarGastoCampana(id: number): Observable<void> {
    return this.http.delete<void>(`${this.api}/gastos/${id}`);
  }
}
