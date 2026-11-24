export type TemperaturaLead = 'FRIO' | 'TIBIO' | 'CALIENTE';

export type EstadoCampana = 'BORRADOR' | 'PROGRAMADA' | 'ACTIVA' | 'PAUSADA' | 'FINALIZADA';

export type TipoAtribucionCampana = 'first_touch' | 'last_touch' | 'influencia';

export interface TipoCampana {
  id: number;
  empresa_id: number;
  nombre: string;
  descripcion: string | null;
  activo: boolean;
  created_at: string;
  updated_at: string;
}

export interface CrearTipoCampanaRequest {
  nombre: string;
  descripcion?: string;
  activo?: boolean;
}

export interface ActualizarTipoCampanaRequest {
  nombre?: string;
  descripcion?: string;
  activo?: boolean;
}

export interface Campana {
  id: number;
  empresa_id: number;
  tipo_campana_id: number;
  origen_id: number | null;
  moneda_id: number | null;
  nombre: string;
  descripcion: string | null;
  objetivo: string | null;
  fecha_inicio: string;
  fecha_fin: string | null;
  presupuesto: number | null;
  estado: EstadoCampana;
  created_by: number;
  created_at: string;
  updated_at: string;
}

export interface ListaCampanasResponse {
  data: Campana[];
  page: number;
  page_size: number;
  total: number;
}

export interface ListarCampanasFiltros {
  tipo_campana_id?: number;
  origen_id?: number;
  estado?: EstadoCampana | '';
  fecha_desde?: string;
  fecha_hasta?: string;
  q?: string;
  page?: number;
  page_size?: number;
}

export interface CrearCampanaRequest {
  tipo_campana_id: number;
  origen_id?: number | null;
  moneda_id?: number | null;
  nombre: string;
  descripcion?: string;
  objetivo?: string;
  fecha_inicio: string;
  fecha_fin?: string;
  presupuesto?: number | null;
  estado?: EstadoCampana;
}

export interface ActualizarCampanaRequest {
  tipo_campana_id?: number;
  origen_id?: number | null;
  moneda_id?: number | null;
  nombre?: string;
  descripcion?: string;
  objetivo?: string;
  fecha_inicio?: string;
  fecha_fin?: string;
  presupuesto?: number | null;
}

export interface GastoPorMonedaResumen {
  moneda_id: number;
  total_gastado: number;
  cantidad_gastos: number;
}

export interface IndicadoresCampana {
  campana_id: number;
  empresa_id: number;
  nombre: string;
  estado: EstadoCampana;
  moneda_id: number | null;
  presupuesto: number | null;
  total_leads: number;
  gasto_acumulado: number | null;
  presupuesto_disponible: number | null;
  gastos_por_moneda: GastoPorMonedaResumen[];
}

export interface AsociarLeadCampanaRequest {
  lead_id: number;
  tipo_atribucion?: TipoAtribucionCampana;
  utm_source?: string;
  utm_medium?: string;
  utm_campaign?: string;
  utm_content?: string;
  utm_term?: string;
  landing_page?: string;
  referrer_url?: string;
  external_campaign_id?: string;
  external_ad_id?: string;
}

export interface LeadAsociadoCampana {
  id: number;
  empresa_id: number;
  campana_id: number;
  lead_id: number;
  fecha_asociacion: string;
  tipo_atribucion: TipoAtribucionCampana | null;
  utm_source: string | null;
  utm_medium: string | null;
  utm_campaign: string | null;
  utm_content: string | null;
  utm_term: string | null;
  landing_page: string | null;
  referrer_url: string | null;
  external_campaign_id: string | null;
  external_ad_id: string | null;
  lead_nombre: string;
  lead_apellido: string;
  lead_email: string | null;
  lead_telefono: string | null;
  lead_nombre_empresa: string | null;
  lead_estado_id: number;
  lead_convertido: boolean;
}

export interface ListaLeadsCampanaResponse {
  data: LeadAsociadoCampana[];
  page: number;
  page_size: number;
  total: number;
}

export interface CampanaAsociadaLead {
  id: number;
  campana_id: number;
  lead_id: number;
  fecha_asociacion: string;
  tipo_atribucion: TipoAtribucionCampana | null;
  campana_nombre: string;
  campana_estado: EstadoCampana;
}

export interface GastoCampana {
  id: number;
  empresa_id: number;
  campana_id: number;
  fecha: string;
  concepto: string;
  categoria_gasto: string | null;
  valor: number;
  moneda_id: number;
  proveedor_tercero_empresa_rol_id: number | null;
  created_by: number;
  created_at: string;
  updated_at: string;
}

export interface ListaGastosCampanaResponse {
  data: GastoCampana[];
  page: number;
  page_size: number;
  total: number;
}

export interface CrearGastoCampanaRequest {
  fecha: string;
  concepto: string;
  categoria_gasto?: string;
  valor: number;
  moneda_id: number;
  proveedor_tercero_empresa_rol_id?: number | null;
}

export interface ActualizarGastoCampanaRequest {
  fecha?: string;
  concepto?: string;
  categoria_gasto?: string;
  valor?: number;
  moneda_id?: number;
  proveedor_tercero_empresa_rol_id?: number | null;
}

export interface OrigenLead {
  id: number;
  empresa_id: number;
  nombre: string;
  activo: boolean;
  created_at: string;
  updated_at: string;
}

export interface CrearOrigenLeadRequest {
  empresa_id: number;
  nombre: string;
  activo?: boolean;
}

export interface ActualizarOrigenLeadRequest {
  nombre?: string;
  activo?: boolean;
}

export interface EstadoLead {
  id: number;
  empresa_id: number;
  nombre: string;
  orden: number;
  es_estado_final: boolean;
  created_at: string;
  updated_at: string;
}

export interface CrearEstadoLeadRequest {
  empresa_id: number;
  nombre: string;
  orden?: number;
  es_estado_final?: boolean;
}

export interface ActualizarEstadoLeadRequest {
  nombre?: string;
  orden?: number;
  es_estado_final?: boolean;
}

export interface EtapaOportunidad {
  id: number;
  empresa_id: number;
  nombre: string;
  orden: number;
  probabilidad_pct: number;
  es_ganada: boolean;
  es_perdida: boolean;
  activo: boolean;
  created_at: string;
  updated_at: string;
}

export interface CrearEtapaOportunidadRequest {
  empresa_id: number;
  nombre: string;
  orden?: number;
  probabilidad_pct?: number;
  es_ganada?: boolean;
  es_perdida?: boolean;
  activo?: boolean;
}

export interface ActualizarEtapaOportunidadRequest {
  nombre?: string;
  orden?: number;
  probabilidad_pct?: number;
  es_ganada?: boolean;
  es_perdida?: boolean;
  activo?: boolean;
}

export interface MotivoDescarteLead {
  id: number;
  empresa_id: number;
  descripcion: string;
  activo: boolean;
  created_at: string;
  updated_at: string;
}

export interface CrearMotivoDescarteRequest {
  empresa_id: number;
  descripcion: string;
  activo?: boolean;
}

export interface ActualizarMotivoDescarteRequest {
  descripcion?: string;
  activo?: boolean;
}

export interface Lead {
  id: number;
  empresa_id: number;
  sucursal_id: number | null;
  origen_id: number;
  estado_id: number;
  motivo_descarte_id: number | null;
  usuario_asignado_id: number | null;
  nombre: string;
  apellido: string;
  nombre_empresa: string | null;
  cargo: string | null;
  email: string | null;
  telefono: string | null;
  usuario_facebook: string | null;
  temperatura: TemperaturaLead;
  score: number;
  valor_estimado: number;
  moneda_id: number | null;
  convertido: boolean;
  convertido_tercero_empresa_rol_id: number | null;
  fecha_conversion: string | null;
  created_by: number;
  created_at: string;
  updated_at: string;
}

export interface ListaLeadsResponse {
  data: Lead[];
  page: number;
  page_size: number;
  total: number;
}

export interface ListarLeadsFiltros {
  empresa_id: number;
  sucursal_id?: number;
  origen_id?: number;
  estado_id?: number;
  temperatura?: TemperaturaLead | '';
  convertido?: boolean | null;
  q?: string;
  page?: number;
  page_size?: number;
}

export interface CrearLeadRequest {
  empresa_id: number;
  sucursal_id?: number | null;
  origen_id: number;
  estado_id: number;
  motivo_descarte_id?: number | null;
  usuario_asignado_id?: number | null;
  nombre: string;
  apellido: string;
  nombre_empresa?: string | null;
  cargo?: string | null;
  email?: string | null;
  telefono?: string | null;
  usuario_facebook?: string | null;
  temperatura?: TemperaturaLead;
  score?: number;
  valor_estimado?: number;
  created_by: number;
}

export interface ActualizarLeadRequest {
  sucursal_id?: number | null;
  origen_id?: number;
  estado_id?: number;
  motivo_descarte_id?: number | null;
  usuario_asignado_id?: number | null;
  nombre?: string;
  apellido?: string;
  nombre_empresa?: string | null;
  cargo?: string | null;
  email?: string | null;
  telefono?: string | null;
  usuario_facebook?: string | null;
  temperatura?: TemperaturaLead;
  score?: number;
  valor_estimado?: number;
}

export interface CambiarEstadoLeadRequest {
  estado_id: number;
  motivo_descarte_id?: number | null;
}

export interface ConvertirLeadRequest {
  estado_id?: number;
}
