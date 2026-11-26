export type TipoIdentificacion = 'NIT' | 'CC' | 'CE' | 'PASAPORTE' | 'RUT' | 'DNI' | 'OTRO';
export type NaturalezaJuridica = 'NATURAL' | 'JURIDICA';

export interface RolTercero {
  id: number;
  codigo: string;
  nombre_rol: string;
  descripcion: string | null;
  activo: boolean;
  es_sistema: boolean;
}

export interface RolAsignado {
  tercero_rol_id: number;
  rol_id: number;
  codigo: string;
  nombre_rol: string;
}

export interface Tercero {
  id: number;
  empresa_id: number;
  tipo_identificacion: TipoIdentificacion;
  numero_identificacion: string;
  razon_social: string;
  pais_id: number | null;
  pais_nombre?: string;
  naturaleza_juridica: NaturalezaJuridica;
  email: string | null;
  telefono: string | null;
  direccion: string | null;
  activo: boolean;
  roles: RolAsignado[];
  created_at: string;
  updated_at: string;
}

export interface ListaTercerosResponse {
  data: Tercero[];
  page: number;
  page_size: number;
  total: number;
}

export interface ListarTercerosFiltros {
  tipo_identificacion?: TipoIdentificacion | '';
  naturaleza_juridica?: NaturalezaJuridica | '';
  pais_id?: number;
  rol_id?: number;
  activo?: boolean | null;
  q?: string;
  page?: number;
  page_size?: number;
}

export interface CrearTerceroRequest {
  tipo_identificacion: TipoIdentificacion;
  numero_identificacion: string;
  razon_social: string;
  pais_id?: number | null;
  naturaleza_juridica: NaturalezaJuridica;
  email?: string | null;
  telefono?: string | null;
  direccion?: string | null;
  activo?: boolean;
  rol_ids: number[];
}

export interface ActualizarTerceroRequest {
  tipo_identificacion?: TipoIdentificacion;
  numero_identificacion?: string;
  razon_social?: string;
  pais_id?: number | null;
  limpiar_pais?: boolean;
  naturaleza_juridica?: NaturalezaJuridica;
  email?: string | null;
  telefono?: string | null;
  direccion?: string | null;
  activo?: boolean;
  rol_ids?: number[];
}
