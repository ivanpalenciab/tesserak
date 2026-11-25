package auth

import (
	"context"
	"net/http"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"tesserak/backend/internal/httpx"
	"tesserak/backend/internal/usuarios"
)

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
	ID            int64  `json:"id"`
	Key           string `json:"key"`
	EmpresaID     int32  `json:"empresa_id"`
	SucursalID    int32  `json:"sucursal_id"`
	NombreUsuario string `json:"nombre_usuario"`
	Nombres       string `json:"nombres"`
	Apellidos     string `json:"apellidos"`
	Email         string `json:"email"`
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
		Model(&usuarios.Usuario{}).
		Where("lower(usuarios.email) = ? OR lower(usuarios.nombre_usuario) = ?", identificador, identificador)

	if req.IdentificadorFiscal != nil && strings.TrimSpace(*req.IdentificadorFiscal) != "" {
		nit := strings.TrimSpace(*req.IdentificadorFiscal)
		q = q.Joins("JOIN empresas ON empresas.id = usuarios.empresa_id AND empresas.deleted_at IS NULL").
			Where("empresas.identificador_fiscal = ?", nit)
	}

	var encontrados []usuarios.Usuario
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

	now := time.Now().UTC()
	_ = s.db.WithContext(ctx).Model(&u).UpdateColumn("ultimo_login_at", now).Error

	claims := Claims{
		Sub:        u.ID,
		EmpresaID:  u.EmpresaID,
		SucursalID: u.SucursalID,
		Email:      u.Email,
		Iat:        now.Unix(),
		Exp:        now.Add(tokenTTL).Unix(),
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
			ID:            u.ID,
			Key:           u.Key,
			EmpresaID:     u.EmpresaID,
			SucursalID:    u.SucursalID,
			NombreUsuario: u.NombreUsuario,
			Nombres:       u.Nombres,
			Apellidos:     u.Apellidos,
			Email:         u.Email,
		},
	}, nil
}
