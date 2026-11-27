export interface RolSistema {
  id: number;
  codigo: string;
  nombre: string;
  descripcion: string | null;
  es_sistema: boolean;
  activo: boolean;
}

export interface SucursalOpcion {
  id: number;
  nombre: string;
  direccion: string | null;
  activo: boolean;
}

export interface UsuarioItem {
  id: number;
  key: string;
  empresa_id: number;
  sucursal_id: number;
  sucursal_ids: number[];
  roles: RolSistema[];
  es_administrador: boolean;
  es_admin_sucursal?: boolean;
  todas_sucursales: boolean;
  nombre_usuario: string;
  nombres: string;
  apellidos: string;
  email: string;
  activo: boolean;
  ultimo_login_at: string | null;
  created_at: string;
  updated_at: string;
}

export interface ListaUsuariosResponse {
  data: UsuarioItem[];
  page: number;
  page_size: number;
  total: number;
}

export interface CrearUsuarioRequest {
  rol_ids: number[];
  sucursal_ids?: number[];
  nombre_usuario: string;
  nombres: string;
  apellidos: string;
  email: string;
  password: string;
  activo?: boolean;
}

export interface ActualizarUsuarioRequest {
  rol_ids?: number[];
  sucursal_ids?: number[];
  nombre_usuario?: string;
  nombres?: string;
  apellidos?: string;
  email?: string;
  password?: string;
  activo?: boolean;
}

export interface ActualizarPerfilRequest {
  nombre_usuario?: string;
  nombres?: string;
  apellidos?: string;
  email?: string;
  password_actual?: string;
  password_nueva?: string;
}

export interface FiltrosUsuarios {
  sucursal_id?: number;
  activo?: boolean;
  q?: string;
  page?: number;
  page_size?: number;
}
