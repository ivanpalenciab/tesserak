export type NaturalezaMovimiento = 'ENTRADA' | 'SALIDA' | 'TRASLADO';

export interface SucursalOption {
  id: number;
  nombre: string;
  activo: boolean;
}

export interface Ubicacion {
  id: number;
  empresa_id: number;
  sucursal_id: number;
  sucursal_nombre?: string;
  codigo: string;
  nombre: string;
  descripcion: string | null;
  activo: boolean;
  created_at: string;
  updated_at: string;
}

export interface CrearUbicacionRequest {
  sucursal_id: number;
  codigo: string;
  nombre: string;
  descripcion?: string | null;
  activo?: boolean;
}

export interface ActualizarUbicacionRequest {
  sucursal_id?: number;
  codigo?: string;
  nombre?: string;
  descripcion?: string | null;
  activo?: boolean;
}

export interface TipoMovimientoInventario {
  id: number;
  codigo: string;
  nombre: string;
  naturaleza: NaturalezaMovimiento;
  descripcion: string | null;
  activo: boolean;
}

export interface StockItem {
  id: number;
  empresa_id: number;
  articulo_id: number;
  articulo_codigo?: string;
  articulo_nombre?: string;
  ubicacion_id: number;
  ubicacion_codigo?: string;
  ubicacion_nombre?: string;
  cantidad: number;
  costo_promedio: number;
  valor_total: number;
  updated_at: string;
}

export interface ListaStockResponse {
  data: StockItem[];
  page: number;
  page_size: number;
  total: number;
}

export interface ListarStockFiltros {
  ubicacion_id?: number;
  articulo_id?: number;
  solo_con_stock?: boolean;
  q?: string;
  page?: number;
  page_size?: number;
}

export interface MovimientoLineaInput {
  articulo_id: number;
  ubicacion_origen_id?: number | null;
  ubicacion_destino_id?: number | null;
  cantidad: number;
  costo_unitario?: number | null;
}

export interface CrearMovimientoRequest {
  tipo_movimiento_id: number;
  fecha?: string;
  tercero_id?: number | null;
  motivo: string;
  observaciones?: string | null;
  detalles: MovimientoLineaInput[];
}

export interface MovimientoDetalle {
  id: number;
  articulo_id: number;
  articulo_codigo?: string;
  articulo_nombre?: string;
  ubicacion_origen_id: number | null;
  ubicacion_origen_nombre?: string;
  ubicacion_destino_id: number | null;
  ubicacion_destino_nombre?: string;
  cantidad: number;
  costo_unitario: number;
}

export interface DocumentoInventario {
  id: number;
  empresa_id: number;
  sucursal_id: number;
  sucursal_nombre?: string;
  tipo_movimiento_id: number;
  tipo_movimiento_codigo?: string;
  tipo_movimiento_nombre?: string;
  tipo_movimiento_naturaleza?: NaturalezaMovimiento;
  consecutivo: string;
  fecha: string;
  tercero_id: number | null;
  tercero_nombre?: string;
  motivo: string;
  observaciones: string | null;
  usuario_id: number;
  usuario_nombre?: string;
  estado: string;
  detalles?: MovimientoDetalle[];
  created_at: string;
}

export interface ListaMovimientosResponse {
  data: DocumentoInventario[];
  page: number;
  page_size: number;
  total: number;
}

export interface ListarMovimientosFiltros {
  tipo_movimiento_id?: number;
  ubicacion_id?: number;
  articulo_id?: number;
  fecha_desde?: string;
  fecha_hasta?: string;
  q?: string;
  page?: number;
  page_size?: number;
}
