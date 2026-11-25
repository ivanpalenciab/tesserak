package productos

import "time"

type ProductoDetalleInput struct {
	ControlInventario  *bool  `json:"control_inventario"`
	MetodoValoracionID *int32 `json:"metodo_valoracion_id" binding:"omitnil,gt=0"`
	PermiteCompra      *bool  `json:"permite_compra"`
	PermiteVenta       *bool  `json:"permite_venta"`
}

type ServicioDetalleInput struct {
	ModalidadID          *int32   `json:"modalidad_id"          binding:"omitnil,gt=0"`
	DuracionEstimada     *float64 `json:"duracion_estimada"     binding:"omitnil,gt=0"`
	RequiereProgramacion *bool    `json:"requiere_programacion"`
	ControlConsumos      *bool    `json:"control_consumos"`
}

type CrearProductoRequest struct {
	Codigo          string                `json:"codigo"            binding:"required,min=1,max=80"`
	Nombre          string                `json:"nombre"            binding:"required,min=1,max=150"`
	Descripcion     *string               `json:"descripcion"       binding:"omitnil,max=5000"`
	TipoArticuloID  int32                 `json:"tipo_articulo_id"  binding:"required,gt=0"`
	UnidadMedidaID  int32                 `json:"unidad_medida_id"  binding:"required,gt=0"`
	CategoriaID     int32                 `json:"categoria_id"      binding:"required,gt=0"`
	MarcaID         *int32                `json:"marca_id"          binding:"omitnil,gt=0"`
	GrupoImpuestoID int32                 `json:"grupo_impuesto_id" binding:"required,gt=0"`
	Estado          *bool                 `json:"estado"`
	Producto        *ProductoDetalleInput `json:"producto"`
	Servicio        *ServicioDetalleInput `json:"servicio"`
}

type ActualizarProductoRequest struct {
	Codigo          *string               `json:"codigo"            binding:"omitnil,min=1,max=80"`
	Nombre          *string               `json:"nombre"            binding:"omitnil,min=1,max=150"`
	Descripcion     *string               `json:"descripcion"       binding:"omitnil,max=5000"`
	UnidadMedidaID  *int32                `json:"unidad_medida_id"  binding:"omitnil,gt=0"`
	CategoriaID     *int32                `json:"categoria_id"      binding:"omitnil,gt=0"`
	MarcaID         *int32                `json:"marca_id"          binding:"omitnil,gt=0"`
	LimpiarMarca    bool                  `json:"limpiar_marca"`
	GrupoImpuestoID *int32                `json:"grupo_impuesto_id" binding:"omitnil,gt=0"`
	Estado          *bool                 `json:"estado"`
	Producto        *ProductoDetalleInput `json:"producto"`
	Servicio        *ServicioDetalleInput `json:"servicio"`
}

type ListarProductosQuery struct {
	TipoArticuloID int32  `form:"tipo_articulo_id" binding:"omitempty,gt=0"`
	Estado         *bool  `form:"estado"`
	CategoriaID    int32  `form:"categoria_id"     binding:"omitempty,gt=0"`
	MarcaID        int32  `form:"marca_id"         binding:"omitempty,gt=0"`
	UnidadMedidaID int32  `form:"unidad_medida_id" binding:"omitempty,gt=0"`
	Q              string `form:"q"                binding:"omitempty,max=100"`
	Page           int    `form:"page"             binding:"omitempty,gte=1"`
	PageSize       int    `form:"page_size"        binding:"omitempty,gte=1,lte=100"`
}

type ProductoDetalleResponse struct {
	ControlInventario  bool   `json:"control_inventario"`
	MetodoValoracionID *int32 `json:"metodo_valoracion_id"`
	PermiteCompra      bool   `json:"permite_compra"`
	PermiteVenta       bool   `json:"permite_venta"`
}

type ServicioDetalleResponse struct {
	ModalidadID          int32    `json:"modalidad_id"`
	DuracionEstimada     *float64 `json:"duracion_estimada"`
	RequiereProgramacion bool     `json:"requiere_programacion"`
	ControlConsumos      bool     `json:"control_consumos"`
}

type ProductoResponse struct {
	ID                 int32                    `json:"id"`
	EmpresaID          int32                    `json:"empresa_id"`
	Codigo             string                   `json:"codigo"`
	Nombre             string                   `json:"nombre"`
	Descripcion        *string                  `json:"descripcion"`
	TipoArticuloID     int32                    `json:"tipo_articulo_id"`
	TipoArticuloCodigo string                   `json:"tipo_articulo_codigo,omitempty"`
	UnidadMedidaID     int32                    `json:"unidad_medida_id"`
	CategoriaID        int32                    `json:"categoria_id"`
	MarcaID            *int32                   `json:"marca_id"`
	GrupoImpuestoID    int32                    `json:"grupo_impuesto_id"`
	Estado             bool                     `json:"estado"`
	Producto           *ProductoDetalleResponse `json:"producto,omitempty"`
	Servicio           *ServicioDetalleResponse `json:"servicio,omitempty"`
	CreatedAt          time.Time                `json:"created_at"`
	UpdatedAt          time.Time                `json:"updated_at"`
}

type ListaProductosResponse struct {
	Data     []ProductoResponse `json:"data"`
	Page     int                `json:"page"`
	PageSize int                `json:"page_size"`
	Total    int64              `json:"total"`
}

func toResponse(a *Articulo) ProductoResponse {
	resp := ProductoResponse{
		ID:              a.ID,
		EmpresaID:       a.EmpresaID,
		Codigo:          a.Codigo,
		Nombre:          a.Nombre,
		Descripcion:     a.Descripcion,
		TipoArticuloID:  a.TipoArticuloID,
		UnidadMedidaID:  a.UnidadMedidaID,
		CategoriaID:     a.CategoriaID,
		MarcaID:         a.MarcaID,
		GrupoImpuestoID: a.GrupoImpuestoID,
		Estado:          a.Estado,
		CreatedAt:       a.CreatedAt,
		UpdatedAt:       a.UpdatedAt,
	}
	if a.TipoArticulo != nil {
		resp.TipoArticuloCodigo = a.TipoArticulo.Codigo
	}
	if a.Producto != nil {
		resp.Producto = &ProductoDetalleResponse{
			ControlInventario:  a.Producto.ControlInventario,
			MetodoValoracionID: a.Producto.MetodoValoracionID,
			PermiteCompra:      a.Producto.PermiteCompra,
			PermiteVenta:       a.Producto.PermiteVenta,
		}
	}
	if a.Servicio != nil {
		resp.Servicio = &ServicioDetalleResponse{
			ModalidadID:          a.Servicio.ModalidadID,
			DuracionEstimada:     a.Servicio.DuracionEstimada,
			RequiereProgramacion: a.Servicio.RequiereProgramacion,
			ControlConsumos:      a.Servicio.ControlConsumos,
		}
	}
	return resp
}
