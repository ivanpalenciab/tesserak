export type TipoListaPrecio = 'GENERAL' | 'MAYORISTA' | 'MINORISTA' | 'ESPECIAL';

export interface ListaPrecio {
  id: number;
  empresa_id: number;
  nombre: string;
  moneda_id: number;
  moneda_codigo?: string;
  moneda_simbolo?: string;
  tipo_lista: TipoListaPrecio;
  vigencia_desde: string;
  vigencia_hasta: string | null;
  estado: boolean;
  created_at: string;
  updated_at: string;
}

export interface ListasPreciosPaginadasResponse {
  data: ListaPrecio[];
  page: number;
  page_size: number;
  total: number;
}

export interface ListarListasPreciosFiltros {
  tipo_lista?: TipoListaPrecio | '';
  moneda_id?: number;
  estado?: boolean | null;
  q?: string;
  page?: number;
  page_size?: number;
}

export interface CrearListaPrecioRequest {
  nombre: string;
  moneda_id: number;
  tipo_lista: TipoListaPrecio;
  vigencia_desde: string;
  vigencia_hasta?: string | null;
  estado?: boolean;
}

export interface ActualizarListaPrecioRequest {
  nombre?: string;
  moneda_id?: number;
  tipo_lista?: TipoListaPrecio;
  vigencia_desde?: string;
  vigencia_hasta?: string | null;
  limpiar_vigencia_hasta?: boolean;
  estado?: boolean;
}

export interface DetallePrecio {
  id: number;
  lista_precio_id: number;
  articulo_id: number;
  articulo_codigo?: string;
  articulo_nombre?: string;
  precio: number;
  unidad_medida_id: number;
  unidad_medida_nombre?: string;
  cantidad_minima: number;
  vigencia_desde: string | null;
  vigencia_hasta: string | null;
  created_at: string;
  updated_at: string;
}

export interface DetallesPrecioPaginadosResponse {
  data: DetallePrecio[];
  page: number;
  page_size: number;
  total: number;
}

export interface ListarDetallesPrecioFiltros {
  articulo_id?: number;
  unidad_medida_id?: number;
  q?: string;
  page?: number;
  page_size?: number;
}

export interface CrearDetallePrecioRequest {
  articulo_id: number;
  precio: number;
  unidad_medida_id: number;
  cantidad_minima?: number;
  vigencia_desde?: string | null;
  vigencia_hasta?: string | null;
}

export interface ActualizarDetallePrecioRequest {
  articulo_id?: number;
  precio?: number;
  unidad_medida_id?: number;
  cantidad_minima?: number;
  vigencia_desde?: string | null;
  vigencia_hasta?: string | null;
  limpiar_vigencia_desde?: boolean;
  limpiar_vigencia_hasta?: boolean;
}
