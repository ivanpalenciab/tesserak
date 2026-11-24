import { Injectable, inject } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable, catchError, forkJoin, map, of } from 'rxjs';

import { environment } from '../../../environments/environment';

export interface EtapaPipelineResumen {
  estado_id: number;
  nombre: string;
  orden: number;
  es_estado_final: boolean;
  cantidad: number;
  valor_estimado: number;
}

export interface TemperaturaResumen {
  temperatura: string;
  cantidad: number;
  valor_estimado: number;
}

export interface ResumenCRM {
  empresa_id: number;
  total_leads: number;
  leads_convertidos: number;
  tasa_conversion_pct: number;
  valor_estimado_total: number;
  actividades_pendientes: number;
  actividades_vencidas: number;
  por_etapa: EtapaPipelineResumen[];
  por_temperatura: TemperaturaResumen[];
}

export interface ActividadResumen {
  id: number;
  lead_id: number;
  tipo_actividad: string;
  fecha_actividad: string;
  fecha_vencimiento: string | null;
  estado: string;
  asunto: string;
}

interface PaginatedResponse<T> {
  data: T[];
  total: number;
}

export interface HomeDashboardData {
  crm: ResumenCRM | null;
  actividades: ActividadResumen[];
  totalUsuarios: number | null;
  totalEmpresas: number | null;
  backendDisponible: boolean;
}

@Injectable({ providedIn: 'root' })
export class DashboardService {
  private readonly http = inject(HttpClient);
  private readonly api = environment.apiUrl;

  cargarResumen(empresaId: number): Observable<HomeDashboardData> {
    const paramsEmpresa = new HttpParams().set('empresa_id', String(empresaId));

    const crm$ = this.http
      .get<ResumenCRM>(`${this.api}/crm/resumen`, { params: paramsEmpresa })
      .pipe(catchError(() => of(null)));

    const actividades$ = this.http
      .get<PaginatedResponse<ActividadResumen>>(`${this.api}/crm/actividades`, {
        params: paramsEmpresa.set('estado', 'PENDIENTE').set('page_size', '5'),
      })
      .pipe(
        map((r) => r.data ?? []),
        catchError(() => of([] as ActividadResumen[])),
      );

    const usuarios$ = this.http
      .get<PaginatedResponse<unknown>>(`${this.api}/usuarios`, {
        params: paramsEmpresa.set('activo', 'true').set('page_size', '1'),
      })
      .pipe(
        map((r) => r.total ?? 0),
        catchError(() => of(null)),
      );

    const empresas$ = this.http
      .get<PaginatedResponse<unknown>>(`${this.api}/empresas`, {
        params: new HttpParams().set('activo', 'true').set('page_size', '1'),
      })
      .pipe(
        map((r) => r.total ?? 0),
        catchError(() => of(null)),
      );

    return forkJoin({
      crm: crm$,
      actividades: actividades$,
      totalUsuarios: usuarios$,
      totalEmpresas: empresas$,
    }).pipe(
      map((res) => ({
        ...res,
        backendDisponible: res.crm !== null || res.totalUsuarios !== null || res.totalEmpresas !== null,
      })),
    );
  }
}
