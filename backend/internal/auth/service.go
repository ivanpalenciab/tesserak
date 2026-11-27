package auth

import (
	"context"
	"net/http"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"tesserak/backend/internal/httpx"
)

// usuarioAuthRow mapea los campos necesarios de la tabla "usuarios" para evitar
// acoplamiento circular con el paquete internal/usuarios.
type usuarioAuthRow struct {
	ID            int64          `gorm:"column:id;primaryKey"`
	Key           string         `gorm:"column:key"`
	EmpresaID     int32          `gorm:"column:empresa_id"`
	NombreUsuario string         `gorm:"column:nombre_usuario"`
	Nombres       string         `gorm:"column:nombres"`
	Apellidos     string         `gorm:"column:apellidos"`
	Email         string         `gorm:"column:email"`
	PasswordHash  string         `gorm:"column:password_hash"`
	Activo        bool           `gorm:"column:activo"`
	DeletedAt     gorm.DeletedAt `gorm:"column:deleted_at"`
}

func (usuarioAuthRow) TableName() string { return "usuarios" }

const tokenTTL = 12 * time.Hour

var (
	ErrCredencialesInvalidas = httpx.NewError(http.StatusUnauthorized, "credenciales_invalidas", "usuario/correo o contraseña incorrectos")
	ErrUsuarioInactivo       = httpx.NewError(http.StatusForbidden, "usuario_inactivo", "el usuario se encuentra inactivo")
	ErrEmpresaRequerida      = httpx.NewError(http.StatusConflict, "empresa_requerida", "el usuario o correo existe en varias empresas; indique el identificador fiscal de su empresa")
	ErrIdentificadorVacio    = httpx.NewError(http.StatusBadRequest, "validacion", "debe indicar un correo electrónico o nombre de usuario")
)

type LoginRequest struct {
	// Identificador acepta correo electrónico o nombre de usuario.
	// Se mantiene Email por retrocompatibilidad si un cliente aún envía "email".
	Identificador       string  `json:"identificador"        binding:"omitempty,max=255"`
	Email               string  `json:"email"                binding:"omitempty,max=255"`
	Password            string  `json:"password"             binding:"required,min=1,max=72"`
	IdentificadorFiscal *string `json:"identificador_fiscal" binding:"omitnil,max=50"`
}

type UsuarioAutenticado struct {
	ID              int64    `json:"id"`
	Key             string   `json:"key"`
	EmpresaID       int32    `json:"empresa_id"`
	SucursalID      int32    `json:"sucursal_id"`
	SucursalIDs     []int32  `json:"sucursal_ids"`
	Roles           []string `json:"roles"`
	EsAdministrador   bool     `json:"es_administrador"`
	EsAdminSucursal   bool     `json:"es_admin_sucursal"`
	EsGestorVentas    bool     `json:"es_gestor_ventas"`
	EsAsesorComercial bool     `json:"es_asesor_comercial"`
	EsGestorCompras   bool     `json:"es_gestor_compras"`
	NombreUsuario     string   `json:"nombre_usuario"`
	Nombres         string   `json:"nombres"`
	Apellidos       string   `json:"apellidos"`
	Email           string   `json:"email"`
}

type LoginResponse struct {
	AccessToken string             `json:"access_token"`
	TokenType   string             `json:"token_type"`
	ExpiresIn   int64              `json:"expires_in"`
	Usuario     UsuarioAutenticado `json:"usuario"`
}

type Service struct {
	db        *gorm.DB
	jwtSecret string
}

func NewService(db *gorm.DB, jwtSecret string) *Service {
	return &Service{db: db, jwtSecret: jwtSecret}
}

func (s *Service) Login(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	identificador := strings.ToLower(strings.TrimSpace(req.Identificador))
	if identificador == "" {
		identificador = strings.ToLower(strings.TrimSpace(req.Email))
	}
	if identificador == "" {
		return nil, ErrIdentificadorVacio
	}

	q := s.db.WithContext(ctx).
		Model(&usuarioAuthRow{}).
		Where("lower(usuarios.email) = ? OR lower(usuarios.nombre_usuario) = ?", identificador, identificador)

	if req.IdentificadorFiscal != nil && strings.TrimSpace(*req.IdentificadorFiscal) != "" {
		nit := strings.TrimSpace(*req.IdentificadorFiscal)
		q = q.Joins("JOIN empresas ON empresas.id = usuarios.empresa_id AND empresas.deleted_at IS NULL").
			Where("empresas.identificador_fiscal = ?", nit)
	}

	var encontrados []usuarioAuthRow
	if err := q.Limit(2).Find(&encontrados).Error; err != nil {
		return nil, err
	}
	if len(encontrados) == 0 {
		return nil, ErrCredencialesInvalidas
	}
	if len(encontrados) > 1 {
		return nil, ErrEmpresaRequerida
	}

	u := encontrados[0]
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrCredencialesInvalidas
	}
	if !u.Activo {
		return nil, ErrUsuarioInactivo
	}

	// Consultar roles de sistema asignados al usuario
	var roles []string
	if err := s.db.WithContext(ctx).
		Table("usuario_roles ur").
		Select("rs.codigo").
		Joins("JOIN roles_sistema rs ON rs.id = ur.rol_id AND rs.activo = TRUE").
		Where("ur.usuario_id = ?", u.ID).
		Order("rs.id ASC").
		Pluck("rs.codigo", &roles).Error; err != nil {
		return nil, err
	}
	if roles == nil {
		roles = []string{}
	}
	esAdmin := false
	esAdminSucursal := false
	esGestorVentas := false
	esAsesorComercial := false
	esGestorCompras := false
	for _, r := range roles {
		if strings.EqualFold(r, RolAdministrador) {
			esAdmin = true
		}
		if strings.EqualFold(r, RolAdminSucursal) {
			esAdminSucursal = true
		}
		if strings.EqualFold(r, RolGestorVentas) {
			esGestorVentas = true
		}
		if strings.EqualFold(r, RolAsesorComercial) {
			esAsesorComercial = true
		}
		if strings.EqualFold(r, RolGestorCompras) {
			esGestorCompras = true
		}
	}

	// Resolver sucursales accesibles:
	// - Si es ADMINISTRADOR, tiene acceso automático a TODAS las sucursales activas de su empresa.
	// - En caso contrario, solo a las sucursales activas asignadas en usuario_sucursales.
	var sucursalIDs []int32
	if esAdmin {
		if err := s.db.WithContext(ctx).
			Table("sucursales s").
			Select("s.id").
			Joins("LEFT JOIN usuario_sucursales us ON us.sucursal_id = s.id AND us.usuario_id = ?", u.ID).
			Where("s.empresa_id = ? AND s.activo = TRUE AND s.deleted_at IS NULL", u.EmpresaID).
			Order("COALESCE(us.es_principal, FALSE) DESC, s.id ASC").
			Pluck("s.id", &sucursalIDs).Error; err != nil {
			return nil, err
		}
	} else {
		if err := s.db.WithContext(ctx).
			Table("usuario_sucursales us").
			Select("us.sucursal_id").
			Joins("JOIN sucursales s ON s.id = us.sucursal_id AND s.activo = TRUE AND s.deleted_at IS NULL").
			Where("us.usuario_id = ?", u.ID).
			Order("us.es_principal DESC, us.sucursal_id ASC").
			Pluck("us.sucursal_id", &sucursalIDs).Error; err != nil {
			return nil, err
		}
	}
	if sucursalIDs == nil {
		sucursalIDs = []int32{}
	}
	var sucursalPrincipal int32
	if len(sucursalIDs) > 0 {
		sucursalPrincipal = sucursalIDs[0]
	}

	now := time.Now().UTC()
	_ = s.db.WithContext(ctx).Model(&u).UpdateColumn("ultimo_login_at", now).Error

	claims := Claims{
		Sub:             u.ID,
		EmpresaID:       u.EmpresaID,
		SucursalID:      sucursalPrincipal,
		SucursalIDs:     sucursalIDs,
		Roles:           roles,
		EsAdministrador:   esAdmin,
		EsAdminSucursal:   esAdminSucursal,
		EsGestorVentas:    esGestorVentas,
		EsAsesorComercial: esAsesorComercial,
		EsGestorCompras:   esGestorCompras,
		Email:             u.Email,
		Iat:               now.Unix(),
		Exp:               now.Add(tokenTTL).Unix(),
	}
	token, err := SignHS256(claims, s.jwtSecret)
	if err != nil {
		return nil, err
	}

	return &LoginResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   int64(tokenTTL.Seconds()),
		Usuario: UsuarioAutenticado{
			ID:                u.ID,
			Key:               u.Key,
			EmpresaID:         u.EmpresaID,
			SucursalID:        sucursalPrincipal,
			SucursalIDs:       sucursalIDs,
			Roles:             roles,
			EsAdministrador:   esAdmin,
			EsAdminSucursal:   esAdminSucursal,
			EsGestorVentas:    esGestorVentas,
			EsAsesorComercial: esAsesorComercial,
			EsGestorCompras:   esGestorCompras,
			NombreUsuario:   u.NombreUsuario,
			Nombres:         u.Nombres,
			Apellidos:       u.Apellidos,
			Email:           u.Email,
		},
	}, nil
}
