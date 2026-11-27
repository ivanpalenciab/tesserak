export type AlcancePrograma = 'LOCAL' | 'COALICION' | 'GLOBAL';
export type EstadoPrograma = 'ACTIVO' | 'INACTIVO' | 'PAUSADO';
export type RolEmpresaPrograma = 'PROPIETARIA' | 'ALIADA';

export interface MonedaLealtad {
  id: number;
  codigo_iso: string;
  nombre_moneda: string;
  simbolo: string;
}

export interface AdhesionPrograma {
  id: number;
  programa_id: number;
  empresa_id: number;
  empresa_nombre?: string;
  empresa_nit?: string;
  rol_empresa: RolEmpresaPrograma;
  permite_acumular: boolean;
  permite_redimir: boolean;
  es_programa_principal: boolean;
  fecha_adhesion: string;
  fecha_baja?: string | null;
  activo: boolean;
}

export interface ProgramaLealtad {
  id: number;
  codigo: string;
  nombre: string;
  descripcion?: string | null;
  alcance: AlcancePrograma;
  empresa_propietaria_id?: number | null;
  empresa_propietaria_nombre?: string;
  usuario_administrador_id?: number | null;
  usuario_administrador_username?: string;
  usuario_administrador_nombre?: string;
  es_sistema: boolean;
  moneda_referencia_id: number;
  moneda_codigo?: string;
  moneda_nombre?: string;
  moneda_simbolo?: string;
  valor_punto_redencion: number;
  dias_vencimiento?: number | null;
  estado: EstadoPrograma;
  fecha_inicio: string;
  fecha_fin?: string | null;
  empresas_adheridas_count: number;
  adhesion?: AdhesionPrograma | null;
  puede_editar: boolean;
  puede_eliminar: boolean;
  puede_ver_empresas: boolean;
  puede_adherir: boolean;
  created_at: string;
  updated_at: string;
}

export interface ListProgramasResponse {
  data: ProgramaLealtad[];
  es_admin_global: boolean;
}

export interface CrearProgramaPayload {
  codigo: string;
  nombre: string;
  descripcion?: string | null;
  alcance: AlcancePrograma;
  moneda_referencia_id: number;
  valor_punto_redencion: number;
  dias_vencimiento?: number | null;
  estado: EstadoPrograma;
  fecha_inicio?: string | null;
  fecha_fin?: string | null;
}

export interface ActualizarProgramaPayload {
  codigo?: string;
  nombre?: string;
  descripcion?: string | null;
  moneda_referencia_id?: number;
  valor_punto_redencion?: number;
  dias_vencimiento?: number | null;
  limpiar_dias_vencimiento?: boolean;
  estado?: EstadoPrograma;
  fecha_inicio?: string | null;
  fecha_fin?: string | null;
  limpiar_fecha_fin?: boolean;
}

export interface AdhesionPayload {
  permite_acumular: boolean;
  permite_redimir: boolean;
  es_programa_principal: boolean;
}

export type TipoReglaLealtad = 'POR_MONTO' | 'BONO_FIJO' | 'MULTIPLICADOR';

export interface ReglaLealtad {
  id: number;
  programa_id: number;
  empresa_id?: number | null;
  empresa_nombre?: string;
  es_global: boolean;
  nombre: string;
  tipo_regla: TipoReglaLealtad;
  puntos_por_monto: number;
  monto_base: number;
  compra_minima: number;
  moneda_id: number;
  moneda_codigo?: string;
  moneda_simbolo?: string;
  prioridad: number;
  fecha_inicio: string;
  fecha_fin?: string | null;
  activo: boolean;
  puede_editar: boolean;
  puede_eliminar: boolean;
  created_at: string;
  updated_at: string;
}

export interface ListarReglasResponse {
  data: ReglaLealtad[];
  puede_crear_global: boolean;
  puede_crear_empresa: boolean;
}

export interface CrearReglaPayload {
  nombre: string;
  es_global: boolean;
  tipo_regla: TipoReglaLealtad;
  puntos_por_monto: number;
  monto_base: number;
  compra_minima: number;
  moneda_id: number;
  prioridad: number;
  fecha_inicio?: string | null;
  fecha_fin?: string | null;
  activo: boolean;
}

export interface ActualizarReglaPayload {
  nombre?: string;
  tipo_regla?: TipoReglaLealtad;
  puntos_por_monto?: number;
  monto_base?: number;
  compra_minima?: number;
  moneda_id?: number;
  prioridad?: number;
  fecha_inicio?: string | null;
  fecha_fin?: string | null;
  limpiar_fecha_fin?: boolean;
  activo?: boolean;
}

export interface ConciliacionEmpresaItem {
  empresa_id: number;
  empresa_nombre: string;
  empresa_nit: string;
  puntos_emitidos: number;
  puntos_redimidos: number;
  puntos_reversados: number;
  puntos_vencidos: number;
  balance_neto_puntos: number;
  valor_neto: number;
  total_transacciones: number;
}

export interface ConciliacionProgramaResponse {
  programa_id: number;
  programa_codigo: string;
  programa_nombre: string;
  valor_punto_redencion: number;
  moneda_codigo?: string;
  moneda_simbolo?: string;
  total_emitidos: number;
  total_redimidos: number;
  total_vencidos: number;
  empresas: ConciliacionEmpresaItem[];
}

export interface ProcesarVencimientosResponse {
  programa_id: number;
  cuentas_afectadas: number;
  lotes_procesados: number;
  puntos_vencidos_total: number;
  procesado_at: string;
}
