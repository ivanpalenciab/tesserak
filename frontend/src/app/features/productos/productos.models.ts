export type TipoArticuloCodigo = 'PRODUCTO' | 'SERVICIO';

export interface TipoArticulo {
  id: number;
  codigo: TipoArticuloCodigo;
  nombre: string;
}

export interface MetodoValoracion {
  id: number;
  codigo: string;
  nombre: string;
}

export interface ModalidadServicio {
  id: number;
  codigo: string;
  nombre: string;
}

export interface GrupoImpuesto {
  id: number;
  empresa_id: number;
  nombre: string;
  tarifa: number;
  activo: boolean;
  created_at?: string;
  updated_at?: string;
}

export interface CrearGrupoImpuestoRequest {
  nombre: string;
  tarifa: number;
  activo?: boolean;
}

export interface ActualizarGrupoImpuestoRequest {
  nombre?: string;
  tarifa?: number;
  activo?: boolean;
}

export interface CrearMarcaRequest {
  nombre: string;
  activo?: boolean;
}

export interface ActualizarMarcaRequest {
  nombre?: string;
  activo?: boolean;
}

export interface UnidadMedida {
  id: number;
  nombre: string;
  descripcion: string | null;
}

export interface CategoriaProducto {
  id: number;
  empresa_id: number;
  nombre: string;
  activo: boolean;
  created_at: string;
  updated_at: string;
}

export interface MarcaProducto {
  id: number;
  empresa_id: number;
  nombre: string;
  activo: boolean;
  created_at: string;
  updated_at: string;
}

export interface CrearCategoriaRequest {
  nombre: string;
  activo?: boolean;
}

export interface ActualizarCategoriaRequest {
  nombre?: string;
  activo?: boolean;
}

export interface ProductoDetalle {
  control_inventario: boolean;
  metodo_valoracion_id: number | null;
  permite_compra: boolean;
  permite_venta: boolean;
}

export interface ServicioDetalle {
  modalidad_id: number;
  duracion_estimada: number | null;
  requiere_programacion: boolean;
  control_consumos: boolean;
}

export interface Producto {
  id: number;
  empresa_id: number;
  codigo: string;
  nombre: string;
  descripcion: string | null;
  tipo_articulo_id: number;
  tipo_articulo_codigo?: TipoArticuloCodigo;
  unidad_medida_id: number;
  categoria_id: number;
  marca_id: number | null;
  grupo_impuesto_id: number;
  estado: boolean;
  producto?: ProductoDetalle | null;
  servicio?: ServicioDetalle | null;
  created_at: string;
  updated_at: string;
}

export interface ListaProductosResponse {
  data: Producto[];
  page: number;
  page_size: number;
  total: number;
}

export interface ListarProductosFiltros {
  tipo_articulo_id?: number;
  estado?: boolean | null;
  categoria_id?: number;
  marca_id?: number;
  unidad_medida_id?: number;
  q?: string;
  page?: number;
  page_size?: number;
}

export interface CrearProductoRequest {
  codigo: string;
  nombre: string;
  descripcion?: string;
  tipo_articulo_id: number;
  unidad_medida_id: number;
  categoria_id: number;
  marca_id?: number | null;
  grupo_impuesto_id: number;
  estado?: boolean;
  producto?: Partial<ProductoDetalle>;
  servicio?: Partial<ServicioDetalle>;
}

export interface ActualizarProductoRequest {
  codigo?: string;
  nombre?: string;
  descripcion?: string;
  unidad_medida_id?: number;
  categoria_id?: number;
  marca_id?: number | null;
  limpiar_marca?: boolean;
  grupo_impuesto_id?: number;
  estado?: boolean;
  producto?: Partial<ProductoDetalle>;
  servicio?: Partial<ServicioDetalle>;
}
