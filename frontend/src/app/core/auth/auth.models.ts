export interface LoginRequest {
  identificador: string;
  password: string;
  identificador_fiscal?: string;
}

export interface UsuarioAutenticado {
  id: number;
  key?: string;
  empresa_id: number;
  sucursal_id: number;
  sucursal_ids?: number[];
  roles?: string[];
  es_administrador?: boolean;
  es_admin_sucursal?: boolean;
  es_gestor_ventas?: boolean;
  es_asesor_comercial?: boolean;
  es_gestor_compras?: boolean;
  nombre_usuario: string;
  nombres: string;
  apellidos: string;
  email: string;
}

export interface LoginResponse {
  access_token: string;
  token_type: string;
  expires_in: number;
  usuario: UsuarioAutenticado;
}

export interface ApiErrorBody {
  error?: {
    code: string;
    message: string;
    details?: unknown;
  };
}
