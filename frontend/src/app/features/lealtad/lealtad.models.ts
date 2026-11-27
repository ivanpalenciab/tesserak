export type EstadoCuentaLealtad = 'ACTIVA' | 'BLOQUEADA' | 'CERRADA';
export type TipoIdentificacionLealtad = 'CC' | 'NIT' | 'CE' | 'PASAPORTE' | 'RUT' | 'DNI' | 'OTRO';

export interface PaisLealtad {
  id: number;
  nombre: string;
  iso_alpha_2: string;
}

export interface ProgramaOperacionLealtad {
  id: number;
  codigo: string;
  nombre: string;
  alcance: 'LOCAL' | 'COALICION' | 'GLOBAL';
  estado: string;
  valor_punto_redencion: number;
  moneda_codigo?: string;
  moneda_simbolo?: string;
  es_programa_principal: boolean;
  permite_acumular: boolean;
  permite_redimir: boolean;
}

export interface ClienteDisponibleLealtad {
  id: number;
  pais_id?: number | null;
  tipo_identificacion: TipoIdentificacionLealtad;
  numero_identificacion: string;
  razon_social: string;
  email?: string | null;
  telefono?: string | null;
}

export interface MiembroLealtad {
  id: number;
  pais_id: number;
  pais_nombre?: string;
  pais_iso?: string;
  tipo_identificacion: TipoIdentificacionLealtad;
  numero_identificacion: string;
  nombre_completo: string;
  email?: string | null;
  telefono?: string | null;
  created_at: string;
  updated_at: string;
}

export interface CuentaLealtad {
  id: number;
  programa_id: number;
  programa_codigo?: string;
  programa_nombre?: string;
  programa_alcance?: string;
  valor_punto_redencion?: number;
  moneda_codigo?: string;
  moneda_simbolo?: string;
  valor_estimado_saldo: number;
  miembro_id: number;
  miembro_nombre?: string;
  miembro_tipo_doc?: string;
  miembro_num_doc?: string;
  miembro_pais_iso?: string;
  miembro_email?: string | null;
  miembro_telefono?: string | null;
  saldo_puntos: number;
  estado: EstadoCuentaLealtad;
  created_at: string;
  updated_at: string;
}

export interface RegistrarMiembroPayload {
  pais_id: number;
  tipo_identificacion: TipoIdentificacionLealtad;
  numero_identificacion: string;
  nombre_completo: string;
  email?: string | null;
  telefono?: string | null;
  programa_id: number;
}

export interface InscribirDesdeTerceroPayload {
  tercero_id: number;
  programa_id: number;
  pais_id?: number | null;
}

export interface ActualizarMiembroPayload {
  nombre_completo?: string;
  email?: string | null;
  telefono?: string | null;
}

export type TipoTransaccionLealtad =
  | 'ACUMULACION'
  | 'REDENCION'
  | 'AJUSTE_POSITIVO'
  | 'AJUSTE_NEGATIVO'
  | 'VENCIMIENTO'
  | 'REVERSO';

export interface TransaccionLealtad {
  id: number;
  cuenta_id: number;
  programa_id: number;
  empresa_id: number;
  empresa_nombre?: string;
  sucursal_id?: number | null;
  venta_id?: number | null;
  regla_id?: number | null;
  regla_nombre?: string;
  tipo_transaccion: TipoTransaccionLealtad;
  puntos: number;
  saldo_posterior: number;
  fecha_vencimiento?: string | null;
  observaciones?: string | null;
  created_by: number;
  created_by_nombre?: string;
  created_at: string;
}

export interface RedencionLealtad {
  id: number;
  cuenta_id: number;
  transaccion_id: number;
  empresa_id: number;
  empresa_nombre?: string;
  venta_id?: number | null;
  puntos_redimidos: number;
  monto_descontado: number;
  moneda_id: number;
  moneda_codigo?: string;
  moneda_simbolo?: string;
  estado: 'APLICADA' | 'ANULADA';
  observaciones?: string | null;
  saldo_posterior?: number;
  created_at: string;
}

export interface AcumularPuntosPayload {
  monto_compra?: number | null;
  puntos?: number | null;
  observaciones?: string | null;
}

export interface RedimirPuntosPayload {
  puntos_redimidos: number;
  venta_id?: number | null;
  observaciones?: string | null;
}

export interface CuentaClienteLealtad {
  tiene_cuenta: boolean;
  cuenta_id?: number;
  programa_id?: number;
  programa_codigo?: string;
  programa_nombre?: string;
  saldo_puntos: number;
  valor_punto_redencion: number;
  valor_maximo_descuento: number;
  moneda_codigo?: string;
  moneda_simbolo?: string;
  permite_acumular: boolean;
  permite_redimir: boolean;
}
