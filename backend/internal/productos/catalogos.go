package productos

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"

	"tesserak/backend/internal/httpx"
)

var (
	errCategoriaNotFound     = errors.New("categoría no encontrada")
	errMarcaNotFound         = errors.New("marca no encontrada")
	errGrupoImpuestoNotFound = errors.New("grupo de impuesto no encontrado")

	ErrCategoriaNoEncontrada     = httpx.NewError(http.StatusNotFound, "categoria_no_encontrada", "la categoría de producto no existe")
	ErrMarcaNoEncontrada         = httpx.NewError(http.StatusNotFound, "marca_no_encontrada", "la marca no existe")
	ErrGrupoImpuestoNoEncontrado = httpx.NewError(http.StatusNotFound, "grupo_impuesto_no_encontrado", "el grupo de impuesto no existe")
	ErrTarifaInvalida            = httpx.NewError(http.StatusBadRequest, "tarifa_invalida", "la tarifa debe estar entre 0 y 100")
	ErrNombreDuplicado           = httpx.NewError(http.StatusConflict, "nombre_duplicado", "ya existe un registro activo con ese nombre en la empresa")
)

// =====================================================================
// Modelos
// =====================================================================

type UnidadMedida struct {
	ID          int32   `gorm:"column:id;primaryKey"`
	Nombre      string  `gorm:"column:nombre;not null"`
	Descripcion *string `gorm:"column:descripcion"`
}

func (UnidadMedida) TableName() string { return "unidades_medida" }

type CategoriaProducto struct {
	ID        int32          `gorm:"column:id;primaryKey"`
	EmpresaID int32          `gorm:"column:empresa_id;not null"`
	Nombre    string         `gorm:"column:nombre;not null"`
	Activo    bool           `gorm:"column:activo;not null"`
	CreatedAt time.Time      `gorm:"column:created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (CategoriaProducto) TableName() string { return "categorias_producto" }

type Marca struct {
	ID        int32          `gorm:"column:id;primaryKey"`
	EmpresaID int32          `gorm:"column:empresa_id;not null"`
	Nombre    string         `gorm:"column:nombre;not null"`
	Activo    bool           `gorm:"column:activo;not null"`
	CreatedAt time.Time      `gorm:"column:created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (Marca) TableName() string { return "marcas" }

// =====================================================================
// DTOs
// =====================================================================

type UnidadMedidaResponse struct {
	ID          int32   `json:"id"`
	Nombre      string  `json:"nombre"`
	Descripcion *string `json:"descripcion"`
}

type CrearCatalogoNombreRequest struct {
	Nombre string `json:"nombre" binding:"required,min=1,max=100"`
	Activo *bool  `json:"activo"`
}

type ActualizarCatalogoNombreRequest struct {
	Nombre *string `json:"nombre" binding:"omitnil,min=1,max=100"`
	Activo *bool   `json:"activo"`
}

type ListarCatalogoQuery struct {
	Activo *bool  `form:"activo"`
	Q      string `form:"q" binding:"omitempty,max=100"`
}

type CatalogoNombreResponse struct {
	ID        int32     `json:"id"`
	EmpresaID int32     `json:"empresa_id"`
	Nombre    string    `json:"nombre"`
	Activo    bool      `json:"activo"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CrearGrupoImpuestoRequest struct {
	Nombre string   `json:"nombre" binding:"required,min=1,max=100"`
	Tarifa *float64 `json:"tarifa" binding:"required,gte=0,lte=100"`
	Activo *bool    `json:"activo"`
}

type ActualizarGrupoImpuestoRequest struct {
	Nombre *string  `json:"nombre" binding:"omitnil,min=1,max=100"`
	Tarifa *float64 `json:"tarifa" binding:"omitnil,gte=0,lte=100"`
	Activo *bool    `json:"activo"`
}

type GrupoImpuestoResponse struct {
	ID        int32     `json:"id"`
	EmpresaID int32     `json:"empresa_id"`
	Nombre    string    `json:"nombre"`
	Tarifa    float64   `json:"tarifa"`
	Activo    bool      `json:"activo"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// =====================================================================
// Repository
// =====================================================================

func (r *Repository) ListUnidadesMedida(ctx context.Context) ([]UnidadMedida, error) {
	var items []UnidadMedida
	err := r.db.WithContext(ctx).Order("nombre ASC").Find(&items).Error
	return items, err
}

func (r *Repository) ListTiposArticulo(ctx context.Context) ([]TipoArticulo, error) {
	var items []TipoArticulo
	err := r.db.WithContext(ctx).Order("id ASC").Find(&items).Error
	return items, err
}

func (r *Repository) ListMetodosValoracion(ctx context.Context) ([]MetodoValoracion, error) {
	var items []MetodoValoracion
	err := r.db.WithContext(ctx).Order("id ASC").Find(&items).Error
	return items, err
}

func (r *Repository) ListModalidadesServicio(ctx context.Context) ([]ModalidadServicio, error) {
	var items []ModalidadServicio
	err := r.db.WithContext(ctx).Order("id ASC").Find(&items).Error
	return items, err
}

func (r *Repository) CreateGrupoImpuesto(ctx context.Context, g *GrupoImpuesto) error {
	return r.db.WithContext(ctx).Create(g).Error
}

func (r *Repository) ListGruposImpuesto(ctx context.Context, empresaID int32, f ListarCatalogoQuery) ([]GrupoImpuesto, error) {
	q := r.db.WithContext(ctx).Where("empresa_id = ?", empresaID)
	if f.Activo != nil {
		q = q.Where("activo = ?", *f.Activo)
	}
	if s := strings.TrimSpace(f.Q); s != "" {
		q = q.Where("nombre ILIKE ?", "%"+escapeLike(s)+"%")
	}
	var items []GrupoImpuesto
	err := q.Order("tarifa ASC, nombre ASC").Find(&items).Error
	return items, err
}

func (r *Repository) FindGrupoImpuestoByID(ctx context.Context, empresaID, id int32) (*GrupoImpuesto, error) {
	var g GrupoImpuesto
	err := r.db.WithContext(ctx).Where("empresa_id = ? AND id = ?", empresaID, id).First(&g).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errGrupoImpuestoNotFound
	}
	return &g, err
}

func (r *Repository) UpdateGrupoImpuesto(ctx context.Context, g *GrupoImpuesto, cambios map[string]any) error {
	if len(cambios) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Model(g).Updates(cambios).Error
}

func (r *Repository) SoftDeleteGrupoImpuesto(ctx context.Context, empresaID, id int32) error {
	res := r.db.WithContext(ctx).Where("empresa_id = ? AND id = ?", empresaID, id).Delete(&GrupoImpuesto{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errGrupoImpuestoNotFound
	}
	return nil
}

func (r *Repository) CreateCategoria(ctx context.Context, c *CategoriaProducto) error {
	return r.db.WithContext(ctx).Create(c).Error
}

func (r *Repository) ListCategorias(ctx context.Context, empresaID int32, f ListarCatalogoQuery) ([]CategoriaProducto, error) {
	q := r.db.WithContext(ctx).Where("empresa_id = ?", empresaID)
	if f.Activo != nil {
		q = q.Where("activo = ?", *f.Activo)
	}
	if s := strings.TrimSpace(f.Q); s != "" {
		q = q.Where("nombre ILIKE ?", "%"+escapeLike(s)+"%")
	}
	var items []CategoriaProducto
	err := q.Order("nombre ASC").Find(&items).Error
	return items, err
}

func (r *Repository) FindCategoriaByID(ctx context.Context, empresaID, id int32) (*CategoriaProducto, error) {
	var c CategoriaProducto
	err := r.db.WithContext(ctx).Where("empresa_id = ? AND id = ?", empresaID, id).First(&c).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errCategoriaNotFound
	}
	return &c, err
}

func (r *Repository) UpdateCategoria(ctx context.Context, c *CategoriaProducto, cambios map[string]any) error {
	if len(cambios) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Model(c).Updates(cambios).Error
}

func (r *Repository) SoftDeleteCategoria(ctx context.Context, empresaID, id int32) error {
	res := r.db.WithContext(ctx).Where("empresa_id = ? AND id = ?", empresaID, id).Delete(&CategoriaProducto{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errCategoriaNotFound
	}
	return nil
}

func (r *Repository) CreateMarca(ctx context.Context, m *Marca) error {
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *Repository) ListMarcas(ctx context.Context, empresaID int32, f ListarCatalogoQuery) ([]Marca, error) {
	q := r.db.WithContext(ctx).Where("empresa_id = ?", empresaID)
	if f.Activo != nil {
		q = q.Where("activo = ?", *f.Activo)
	}
	if s := strings.TrimSpace(f.Q); s != "" {
		q = q.Where("nombre ILIKE ?", "%"+escapeLike(s)+"%")
	}
	var items []Marca
	err := q.Order("nombre ASC").Find(&items).Error
	return items, err
}

func (r *Repository) FindMarcaByID(ctx context.Context, empresaID, id int32) (*Marca, error) {
	var m Marca
	err := r.db.WithContext(ctx).Where("empresa_id = ? AND id = ?", empresaID, id).First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errMarcaNotFound
	}
	return &m, err
}

func (r *Repository) UpdateMarca(ctx context.Context, m *Marca, cambios map[string]any) error {
	if len(cambios) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Model(m).Updates(cambios).Error
}

func (r *Repository) SoftDeleteMarca(ctx context.Context, empresaID, id int32) error {
	res := r.db.WithContext(ctx).Where("empresa_id = ? AND id = ?", empresaID, id).Delete(&Marca{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errMarcaNotFound
	}
	return nil
}

// =====================================================================
// Service
// =====================================================================

func (s *Service) ListarUnidadesMedida(ctx context.Context) ([]UnidadMedidaResponse, error) {
	items, err := s.repo.ListUnidadesMedida(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]UnidadMedidaResponse, 0, len(items))
	for i := range items {
		out = append(out, UnidadMedidaResponse{
			ID:          items[i].ID,
			Nombre:      items[i].Nombre,
			Descripcion: items[i].Descripcion,
		})
	}
	return out, nil
}

func (s *Service) CrearCategoria(ctx context.Context, empresaID int32, req CrearCatalogoNombreRequest) (*CatalogoNombreResponse, error) {
	activo := true
	if req.Activo != nil {
		activo = *req.Activo
	}
	c := &CategoriaProducto{
		EmpresaID: empresaID,
		Nombre:    strings.TrimSpace(req.Nombre),
		Activo:    activo,
	}
	if err := s.repo.CreateCategoria(ctx, c); err != nil {
		return nil, mapCatalogoError(err)
	}
	resp := toCategoriaResponse(c)
	return &resp, nil
}

func (s *Service) ListarCategorias(ctx context.Context, empresaID int32, q ListarCatalogoQuery) ([]CatalogoNombreResponse, error) {
	items, err := s.repo.ListCategorias(ctx, empresaID, q)
	if err != nil {
		return nil, err
	}
	out := make([]CatalogoNombreResponse, 0, len(items))
	for i := range items {
		out = append(out, toCategoriaResponse(&items[i]))
	}
	return out, nil
}

func (s *Service) ObtenerCategoria(ctx context.Context, empresaID, id int32) (*CatalogoNombreResponse, error) {
	c, err := s.repo.FindCategoriaByID(ctx, empresaID, id)
	if err != nil {
		return nil, mapCatalogoError(err)
	}
	resp := toCategoriaResponse(c)
	return &resp, nil
}

func (s *Service) ActualizarCategoria(ctx context.Context, empresaID, id int32, req ActualizarCatalogoNombreRequest) (*CatalogoNombreResponse, error) {
	var res *CategoriaProducto
	err := s.repo.WithTx(ctx, func(tx *Repository) error {
		c, err := tx.FindCategoriaByID(ctx, empresaID, id)
		if err != nil {
			return err
		}
		cambios := map[string]any{}
		if req.Nombre != nil {
			cambios["nombre"] = strings.TrimSpace(*req.Nombre)
		}
		if req.Activo != nil {
			cambios["activo"] = *req.Activo
		}
		if err := tx.UpdateCategoria(ctx, c, cambios); err != nil {
			return err
		}
		res, err = tx.FindCategoriaByID(ctx, empresaID, id)
		return err
	})
	if err != nil {
		return nil, mapCatalogoError(err)
	}
	resp := toCategoriaResponse(res)
	return &resp, nil
}

func (s *Service) EliminarCategoria(ctx context.Context, empresaID, id int32) error {
	return mapCatalogoError(s.repo.SoftDeleteCategoria(ctx, empresaID, id))
}

func (s *Service) CrearMarca(ctx context.Context, empresaID int32, req CrearCatalogoNombreRequest) (*CatalogoNombreResponse, error) {
	activo := true
	if req.Activo != nil {
		activo = *req.Activo
	}
	m := &Marca{
		EmpresaID: empresaID,
		Nombre:    strings.TrimSpace(req.Nombre),
		Activo:    activo,
	}
	if err := s.repo.CreateMarca(ctx, m); err != nil {
		return nil, mapCatalogoError(err)
	}
	resp := toMarcaResponse(m)
	return &resp, nil
}

func (s *Service) ListarMarcas(ctx context.Context, empresaID int32, q ListarCatalogoQuery) ([]CatalogoNombreResponse, error) {
	items, err := s.repo.ListMarcas(ctx, empresaID, q)
	if err != nil {
		return nil, err
	}
	out := make([]CatalogoNombreResponse, 0, len(items))
	for i := range items {
		out = append(out, toMarcaResponse(&items[i]))
	}
	return out, nil
}

func (s *Service) ActualizarMarca(ctx context.Context, empresaID, id int32, req ActualizarCatalogoNombreRequest) (*CatalogoNombreResponse, error) {
	var res *Marca
	err := s.repo.WithTx(ctx, func(tx *Repository) error {
		m, err := tx.FindMarcaByID(ctx, empresaID, id)
		if err != nil {
			return err
		}
		cambios := map[string]any{}
		if req.Nombre != nil {
			cambios["nombre"] = strings.TrimSpace(*req.Nombre)
		}
		if req.Activo != nil {
			cambios["activo"] = *req.Activo
		}
		if err := tx.UpdateMarca(ctx, m, cambios); err != nil {
			return err
		}
		res, err = tx.FindMarcaByID(ctx, empresaID, id)
		return err
	})
	if err != nil {
		return nil, mapCatalogoError(err)
	}
	resp := toMarcaResponse(res)
	return &resp, nil
}

func (s *Service) EliminarMarca(ctx context.Context, empresaID, id int32) error {
	return mapCatalogoError(s.repo.SoftDeleteMarca(ctx, empresaID, id))
}

func (s *Service) CrearGrupoImpuesto(ctx context.Context, empresaID int32, req CrearGrupoImpuestoRequest) (*GrupoImpuestoResponse, error) {
	if req.Tarifa == nil || *req.Tarifa < 0 || *req.Tarifa > 100 {
		return nil, ErrTarifaInvalida
	}
	activo := true
	if req.Activo != nil {
		activo = *req.Activo
	}
	g := &GrupoImpuesto{
		EmpresaID: empresaID,
		Nombre:    strings.TrimSpace(req.Nombre),
		Tarifa:    *req.Tarifa,
		Activo:    activo,
	}
	if err := s.repo.CreateGrupoImpuesto(ctx, g); err != nil {
		return nil, mapCatalogoError(err)
	}
	resp := toGrupoImpuestoResponse(g)
	return &resp, nil
}

func (s *Service) ListarGruposImpuesto(ctx context.Context, empresaID int32, q ListarCatalogoQuery) ([]GrupoImpuestoResponse, error) {
	items, err := s.repo.ListGruposImpuesto(ctx, empresaID, q)
	if err != nil {
		return nil, err
	}
	out := make([]GrupoImpuestoResponse, 0, len(items))
	for i := range items {
		out = append(out, toGrupoImpuestoResponse(&items[i]))
	}
	return out, nil
}

func (s *Service) ActualizarGrupoImpuesto(ctx context.Context, empresaID, id int32, req ActualizarGrupoImpuestoRequest) (*GrupoImpuestoResponse, error) {
	if req.Tarifa != nil && (*req.Tarifa < 0 || *req.Tarifa > 100) {
		return nil, ErrTarifaInvalida
	}
	var res *GrupoImpuesto
	err := s.repo.WithTx(ctx, func(tx *Repository) error {
		g, err := tx.FindGrupoImpuestoByID(ctx, empresaID, id)
		if err != nil {
			return err
		}
		cambios := map[string]any{}
		if req.Nombre != nil {
			cambios["nombre"] = strings.TrimSpace(*req.Nombre)
		}
		if req.Tarifa != nil {
			cambios["tarifa"] = *req.Tarifa
		}
		if req.Activo != nil {
			cambios["activo"] = *req.Activo
		}
		if err := tx.UpdateGrupoImpuesto(ctx, g, cambios); err != nil {
			return err
		}
		res, err = tx.FindGrupoImpuestoByID(ctx, empresaID, id)
		return err
	})
	if err != nil {
		return nil, mapCatalogoError(err)
	}
	resp := toGrupoImpuestoResponse(res)
	return &resp, nil
}

func (s *Service) EliminarGrupoImpuesto(ctx context.Context, empresaID, id int32) error {
	return mapCatalogoError(s.repo.SoftDeleteGrupoImpuesto(ctx, empresaID, id))
}

func toGrupoImpuestoResponse(g *GrupoImpuesto) GrupoImpuestoResponse {
	return GrupoImpuestoResponse{
		ID:        g.ID,
		EmpresaID: g.EmpresaID,
		Nombre:    g.Nombre,
		Tarifa:    g.Tarifa,
		Activo:    g.Activo,
		CreatedAt: g.CreatedAt,
		UpdatedAt: g.UpdatedAt,
	}
}

func toCategoriaResponse(c *CategoriaProducto) CatalogoNombreResponse {
	return CatalogoNombreResponse{
		ID:        c.ID,
		EmpresaID: c.EmpresaID,
		Nombre:    c.Nombre,
		Activo:    c.Activo,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}

func toMarcaResponse(m *Marca) CatalogoNombreResponse {
	return CatalogoNombreResponse{
		ID:        m.ID,
		EmpresaID: m.EmpresaID,
		Nombre:    m.Nombre,
		Activo:    m.Activo,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

func mapCatalogoError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, errCategoriaNotFound) {
		return ErrCategoriaNoEncontrada
	}
	if errors.Is(err, errMarcaNotFound) {
		return ErrMarcaNoEncontrada
	}
	if errors.Is(err, errGrupoImpuestoNotFound) {
		return ErrGrupoImpuestoNoEncontrado
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return ErrNombreDuplicado
	}
	return mapError(err)
}

// =====================================================================
// Handlers
// =====================================================================

func (h *Handler) registerCatalogoRoutes(g *gin.RouterGroup) {
	g.GET("/unidades-medida", h.listarUnidadesMedida)
	g.GET("/tipos-articulo", h.listarTiposArticulo)
	g.GET("/metodos-valoracion", h.listarMetodosValoracion)
	g.GET("/modalidades-servicio", h.listarModalidadesServicio)

	imp := g.Group("/grupos-impuesto")
	imp.POST("", h.crearGrupoImpuesto)
	imp.GET("", h.listarGruposImpuesto)
	imp.PATCH("/:id", h.actualizarGrupoImpuesto)
	imp.DELETE("/:id", h.eliminarGrupoImpuesto)

	cats := g.Group("/categorias")
	cats.POST("", h.crearCategoria)
	cats.GET("", h.listarCategorias)
	cats.GET("/:id", h.obtenerCategoria)
	cats.PATCH("/:id", h.actualizarCategoria)
	cats.DELETE("/:id", h.eliminarCategoria)

	marcas := g.Group("/marcas")
	marcas.POST("", h.crearMarca)
	marcas.GET("", h.listarMarcas)
	marcas.PATCH("/:id", h.actualizarMarca)
	marcas.DELETE("/:id", h.eliminarMarca)
}

func (h *Handler) listarUnidadesMedida(c *gin.Context) {
	if _, ok := requireClaims(c); !ok {
		return
	}
	resp, err := h.svc.ListarUnidadesMedida(c.Request.Context())
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}

func (h *Handler) listarTiposArticulo(c *gin.Context) {
	if _, ok := requireClaims(c); !ok {
		return
	}
	items, err := h.svc.repo.ListTiposArticulo(c.Request.Context())
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	out := make([]gin.H, 0, len(items))
	for _, it := range items {
		out = append(out, gin.H{"id": it.ID, "codigo": it.Codigo, "nombre": it.Nombre})
	}
	c.JSON(http.StatusOK, gin.H{"data": out})
}

func (h *Handler) listarMetodosValoracion(c *gin.Context) {
	if _, ok := requireClaims(c); !ok {
		return
	}
	items, err := h.svc.repo.ListMetodosValoracion(c.Request.Context())
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	out := make([]gin.H, 0, len(items))
	for _, it := range items {
		out = append(out, gin.H{"id": it.ID, "codigo": it.Codigo, "nombre": it.Nombre})
	}
	c.JSON(http.StatusOK, gin.H{"data": out})
}

func (h *Handler) listarModalidadesServicio(c *gin.Context) {
	if _, ok := requireClaims(c); !ok {
		return
	}
	items, err := h.svc.repo.ListModalidadesServicio(c.Request.Context())
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	out := make([]gin.H, 0, len(items))
	for _, it := range items {
		out = append(out, gin.H{"id": it.ID, "codigo": it.Codigo, "nombre": it.Nombre})
	}
	c.JSON(http.StatusOK, gin.H{"data": out})
}

func (h *Handler) crearGrupoImpuesto(c *gin.Context) {
	claims, ok := requireClaims(c)
	if !ok {
		return
	}
	var req CrearGrupoImpuestoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondError(c, httpx.BindError(err))
		return
	}
	resp, err := h.svc.CrearGrupoImpuesto(c.Request.Context(), claims.EmpresaID, req)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *Handler) listarGruposImpuesto(c *gin.Context) {
	claims, ok := requireClaims(c)
	if !ok {
		return
	}
	var q ListarCatalogoQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		httpx.RespondError(c, httpx.BindError(err))
		return
	}
	resp, err := h.svc.ListarGruposImpuesto(c.Request.Context(), claims.EmpresaID, q)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}

func (h *Handler) actualizarGrupoImpuesto(c *gin.Context) {
	claims, ok := requireClaims(c)
	if !ok {
		return
	}
	id, err := parseID32(c, "id")
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	var req ActualizarGrupoImpuestoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondError(c, httpx.BindError(err))
		return
	}
	resp, err := h.svc.ActualizarGrupoImpuesto(c.Request.Context(), claims.EmpresaID, id, req)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) eliminarGrupoImpuesto(c *gin.Context) {
	claims, ok := requireClaims(c)
	if !ok {
		return
	}
	id, err := parseID32(c, "id")
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	if err := h.svc.EliminarGrupoImpuesto(c.Request.Context(), claims.EmpresaID, id); err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) crearCategoria(c *gin.Context) {
	claims, ok := requireClaims(c)
	if !ok {
		return
	}
	var req CrearCatalogoNombreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondError(c, httpx.BindError(err))
		return
	}
	resp, err := h.svc.CrearCategoria(c.Request.Context(), claims.EmpresaID, req)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *Handler) listarCategorias(c *gin.Context) {
	claims, ok := requireClaims(c)
	if !ok {
		return
	}
	var q ListarCatalogoQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		httpx.RespondError(c, httpx.BindError(err))
		return
	}
	resp, err := h.svc.ListarCategorias(c.Request.Context(), claims.EmpresaID, q)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}

func (h *Handler) obtenerCategoria(c *gin.Context) {
	claims, ok := requireClaims(c)
	if !ok {
		return
	}
	id, err := parseID32(c, "id")
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	resp, err := h.svc.ObtenerCategoria(c.Request.Context(), claims.EmpresaID, id)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) actualizarCategoria(c *gin.Context) {
	claims, ok := requireClaims(c)
	if !ok {
		return
	}
	id, err := parseID32(c, "id")
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	var req ActualizarCatalogoNombreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondError(c, httpx.BindError(err))
		return
	}
	resp, err := h.svc.ActualizarCategoria(c.Request.Context(), claims.EmpresaID, id, req)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) eliminarCategoria(c *gin.Context) {
	claims, ok := requireClaims(c)
	if !ok {
		return
	}
	id, err := parseID32(c, "id")
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	if err := h.svc.EliminarCategoria(c.Request.Context(), claims.EmpresaID, id); err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) crearMarca(c *gin.Context) {
	claims, ok := requireClaims(c)
	if !ok {
		return
	}
	var req CrearCatalogoNombreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondError(c, httpx.BindError(err))
		return
	}
	resp, err := h.svc.CrearMarca(c.Request.Context(), claims.EmpresaID, req)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *Handler) listarMarcas(c *gin.Context) {
	claims, ok := requireClaims(c)
	if !ok {
		return
	}
	var q ListarCatalogoQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		httpx.RespondError(c, httpx.BindError(err))
		return
	}
	resp, err := h.svc.ListarMarcas(c.Request.Context(), claims.EmpresaID, q)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}

func (h *Handler) actualizarMarca(c *gin.Context) {
	claims, ok := requireClaims(c)
	if !ok {
		return
	}
	id, err := parseID32(c, "id")
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	var req ActualizarCatalogoNombreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondError(c, httpx.BindError(err))
		return
	}
	resp, err := h.svc.ActualizarMarca(c.Request.Context(), claims.EmpresaID, id, req)
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) eliminarMarca(c *gin.Context) {
	claims, ok := requireClaims(c)
	if !ok {
		return
	}
	id, err := parseID32(c, "id")
	if err != nil {
		httpx.RespondError(c, err)
		return
	}
	if err := h.svc.EliminarMarca(c.Request.Context(), claims.EmpresaID, id); err != nil {
		httpx.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
