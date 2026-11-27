export type EstadoVenta = 'BORRADOR' | 'REGISTRADA' | 'PAGADA' | 'ANULADA';

export interface VentaDetalle {
  id: number;
  articulo_id: number;
  articulo_codigo?: string;
  articulo_nombre?: string;
  descripcion: string | null;
  cantidad: number;
  precio_unitario: number;
  subtotal: number;
  impuesto: number;
  total: number;
}

export interface Venta {
  id: number;
  empresa_id: number;
  sucursal_id: number;
  sucursal_nombre?: string;
  ubicacion_id: number;
  ubicacion_nombre?: string;
  cliente_id: number;
  cliente_nombre?: string;
  cliente_identificacion?: string;
  consecutivo: string;
  moneda_id: number | null;
  moneda_codigo?: string;
  estado: EstadoVenta;
  fecha_venta: string;
  fecha_vencimiento: string | null;
  subtotal: number;
  total_iva: number;
  puntos_redimidos?: number;
  descuento_lealtad?: number;
  total: number;
  observaciones: string | null;
  usuario_id: number;
  usuario_nombre?: string;
  detalles?: VentaDetalle[];
  created_at: string;
  updated_at: string;
}

export interface ListaVentasResponse {
  data: Venta[];
  page: number;
  page_size: number;
  total: number;
}

export interface VentaDetalleInput {
  articulo_id: number;
  descripcion?: string | null;
  cantidad: number;
  precio_unitario: number;
  impuesto?: number;
}

export interface CrearVentaRequest {
  ubicacion_id?: number | null;
  cliente_id?: number | null;
  estado?: EstadoVenta;
  fecha_venta: string;
  fecha_vencimiento?: string | null;
  observaciones?: string | null;
  puntos_redimir?: number | null;
  detalles: VentaDetalleInput[];
}

export interface ActualizarVentaRequest {
  ubicacion_id?: number | null;
  cliente_id?: number | null;
  estado?: EstadoVenta;
  fecha_venta?: string;
  fecha_vencimiento?: string | null;
  limpiar_fecha_vencimiento?: boolean;
  observaciones?: string | null;
  detalles?: VentaDetalleInput[];
}

export interface ListarVentasFiltros {
  estado?: EstadoVenta | '';
  cliente_id?: number;
  ubicacion_id?: number;
  fecha_desde?: string;
  fecha_hasta?: string;
  q?: string;
  page?: number;
  page_size?: number;
}
