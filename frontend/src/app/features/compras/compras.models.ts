export type EstadoCompra = 'BORRADOR' | 'REGISTRADA' | 'PAGADA' | 'ANULADA';

export interface CompraDetalle {
  id: number;
  compra_id: number;
  articulo_id: number;
  articulo_codigo?: string;
  articulo_nombre?: string;
  descripcion: string | null;
  cantidad: number;
  costo_unitario: number;
  subtotal: number;
  impuesto: number;
  total: number;
}

export interface Compra {
  id: number;
  empresa_id: number;
  sucursal_id: number;
  sucursal_nombre?: string;
  tercero_empresa_rol_id: number | null;
  proveedor_nombre: string | null;
  consecutivo: string;
  numero_factura_proveedor: string | null;
  moneda_id: number | null;
  moneda_codigo?: string;
  moneda_simbolo?: string;
  estado: EstadoCompra;
  fecha_compra: string;
  fecha_vencimiento: string | null;
  subtotal: number;
  total_iva: number;
  total: number;
  observaciones: string | null;
  usuario_id: number;
  usuario_nombre?: string;
  detalles?: CompraDetalle[];
  created_at: string;
  updated_at: string;
}

export interface ListaComprasResponse {
  data: Compra[];
  page: number;
  page_size: number;
  total: number;
}

export interface ListarComprasFiltros {
  sucursal_id?: number;
  estado?: EstadoCompra | '';
  moneda_id?: number;
  fecha_desde?: string;
  fecha_hasta?: string;
  q?: string;
  page?: number;
  page_size?: number;
}

export interface CompraDetalleInput {
  articulo_id: number;
  descripcion?: string | null;
  cantidad: number;
  costo_unitario: number;
  impuesto?: number;
}

export interface CrearCompraRequest {
  sucursal_id?: number;
  tercero_empresa_rol_id?: number | null;
  proveedor_nombre?: string | null;
  consecutivo?: string;
  numero_factura_proveedor?: string | null;
  moneda_id?: number | null;
  estado?: EstadoCompra;
  fecha_compra: string;
  fecha_vencimiento?: string | null;
  observaciones?: string | null;
  detalles: CompraDetalleInput[];
}

export interface ActualizarCompraRequest {
  sucursal_id?: number;
  tercero_empresa_rol_id?: number | null;
  proveedor_nombre?: string | null;
  consecutivo?: string;
  numero_factura_proveedor?: string | null;
  moneda_id?: number | null;
  estado?: EstadoCompra;
  fecha_compra?: string;
  fecha_vencimiento?: string | null;
  limpiar_fecha_vencimiento?: boolean;
  observaciones?: string | null;
  detalles?: CompraDetalleInput[];
}
