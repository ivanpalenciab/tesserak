export interface LoginRequest {
  identificador: string;
  password: string;
  identificador_fiscal?: string;
}

export interface UsuarioAutenticado {
  id: number;
  empresa_id: number;
  sucursal_id: number;
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
