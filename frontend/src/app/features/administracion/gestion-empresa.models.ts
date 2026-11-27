export type FrecuenciaReinicio = 'NUNCA' | 'ANUAL' | 'MENSUAL';

export interface TipoDocumentoConsecutivo {
  id: number;
  codigo: string;
  nombre: string;
  descripcion: string | null;
  prefijo_defecto: string;
  activo: boolean;
}

export interface SecuenciaDocumento {
  id: number;
  empresa_id: number;
  tipo_documento_id: number;
  tipo_documento_codigo?: string;
  tipo_documento_nombre?: string;
  prefijo: string;
  siguiente_numero: number;
  longitud_relleno: number;
  frecuencia_reinicio: FrecuenciaReinicio;
  clave_periodo: string | null;
  activo: boolean;
  ejemplo_proximo: string;
  created_at: string;
  updated_at: string;
}

export interface CrearSecuenciaRequest {
  tipo_documento_id: number;
  prefijo: string;
  siguiente_numero?: number;
  longitud_relleno?: number;
  frecuencia_reinicio?: FrecuenciaReinicio;
  activo?: boolean;
}

export interface ActualizarSecuenciaRequest {
  prefijo?: string;
  siguiente_numero?: number;
  longitud_relleno?: number;
  frecuencia_reinicio?: FrecuenciaReinicio;
  activo?: boolean;
}

export interface EmpresaDetalle {
  id: number;
  nombre_legal: string;
  razon_social: string;
  identificador_fiscal: string;
  email: string | null;
  telefono: string | null;
  direccion: string | null;
  activo: boolean;
}

export interface UsuarioKeyDetalle {
  id: number;
  key: string;
  empresa_id: number;
  sucursal_id: number;
  sucursal_ids?: number[];
  nombre_usuario: string;
  nombres: string;
  apellidos: string;
  email: string;
  activo: boolean;
}

export interface SucursalDetalle {
  id: number;
  empresa_id: number;
  ciudad_id: number | null;
  nombre: string;
  direccion: string | null;
  activo: boolean;
  created_at: string;
  updated_at: string;
}

export interface CrearSucursalRequest {
  nombre: string;
  direccion?: string | null;
  activo?: boolean;
}

export interface ActualizarSucursalRequest {
  nombre?: string;
  direccion?: string | null;
  activo?: boolean;
}
