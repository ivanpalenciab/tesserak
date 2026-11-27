// Package auth implementa autenticación de usuarios con JWT (HS256) y bcrypt.
package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

const (
	// RolAdministrador es el código del rol de sistema con acceso total a la empresa.
	RolAdministrador = "ADMINISTRADOR"
	// RolAdminSucursal es el código del rol con acceso administrativo en sus sucursales asignadas.
	RolAdminSucursal = "ADMINISTRADOR_SUCURSAL"
	// RolGestorVentas es el código del rol con permiso solo para crear/modificar ventas y clientes.
	RolGestorVentas = "GESTOR_VENTAS"
	// RolAsesorComercial es el código del rol con acceso completo a CRM.
	RolAsesorComercial = "ASESOR_COMERCIAL"
	// RolGestorCompras es el código del rol con permiso para crear y modificar compras.
	RolGestorCompras = "GESTOR_COMPRAS"
)

// Claims contiene la información firmada dentro del JWT.
type Claims struct {
	Sub               int64    `json:"sub"`
	EmpresaID         int32    `json:"empresa_id"`
	SucursalID        int32    `json:"sucursal_id"`
	SucursalIDs       []int32  `json:"sucursal_ids,omitempty"`
	Roles             []string `json:"roles,omitempty"`
	EsAdministrador   bool     `json:"es_administrador,omitempty"`
	EsAdminSucursal   bool     `json:"es_admin_sucursal,omitempty"`
	EsGestorVentas    bool     `json:"es_gestor_ventas,omitempty"`
	EsAsesorComercial bool     `json:"es_asesor_comercial,omitempty"`
	EsGestorCompras   bool     `json:"es_gestor_compras,omitempty"`
	Email             string   `json:"email"`
	Iat               int64    `json:"iat"`
	Exp               int64    `json:"exp"`
}

// IsAdmin indica si el token autenticado posee el rol de sistema ADMINISTRADOR.
func (c *Claims) IsAdmin() bool {
	if c == nil {
		return false
	}
	if c.EsAdministrador {
		return true
	}
	for _, r := range c.Roles {
		if strings.EqualFold(r, RolAdministrador) {
			return true
		}
	}
	return false
}

// IsAdminSucursal indica si el token posee el rol ADMINISTRADOR_SUCURSAL.
func (c *Claims) IsAdminSucursal() bool {
	if c == nil {
		return false
	}
	if c.EsAdminSucursal {
		return true
	}
	for _, r := range c.Roles {
		if strings.EqualFold(r, RolAdminSucursal) {
			return true
		}
	}
	return false
}

// CanManageUsers indica si el usuario puede administrar usuarios (ADMINISTRADOR o ADMINISTRADOR_SUCURSAL).
func (c *Claims) CanManageUsers() bool {
	return c.IsAdmin() || c.IsAdminSucursal()
}

// IsGestorVentas indica si el token posee el rol GESTOR_VENTAS.
func (c *Claims) IsGestorVentas() bool {
	if c == nil {
		return false
	}
	if c.EsGestorVentas {
		return true
	}
	for _, r := range c.Roles {
		if strings.EqualFold(r, RolGestorVentas) {
			return true
		}
	}
	return false
}

// HasRole indica si el token contiene el código de rol indicado.
func (c *Claims) HasRole(codigo string) bool {
	if c == nil {
		return false
	}
	for _, r := range c.Roles {
		if strings.EqualFold(r, codigo) {
			return true
		}
	}
	return false
}

// HasFullOperationalAccess indica si el usuario tiene acceso operativo completo
// (ADMINISTRADOR, ADMINISTRADOR_SUCURSAL, USUARIO estándar o token sin roles restringidos).
func (c *Claims) HasFullOperationalAccess() bool {
	if c == nil {
		return false
	}
	if c.IsAdmin() || c.IsAdminSucursal() || c.HasRole("USUARIO") {
		return true
	}
	return !c.IsGestorVentas() && !c.IsAsesorComercial() && !c.IsGestorCompras()
}

// IsGestorCompras indica si el token posee el rol GESTOR_COMPRAS.
func (c *Claims) IsGestorCompras() bool {
	if c == nil {
		return false
	}
	if c.EsGestorCompras {
		return true
	}
	return c.HasRole(RolGestorCompras)
}

// CanAccessCompras indica si el usuario puede consultar, crear y modificar compras.
func (c *Claims) CanAccessCompras() bool {
	return c.HasFullOperationalAccess() || c.IsGestorCompras()
}

// CanDeleteCompras indica si el usuario puede eliminar compras.
func (c *Claims) CanDeleteCompras() bool {
	return c.HasFullOperationalAccess()
}

// CanAccessCRM indica si el usuario puede operar el módulo CRM (unión aditiva de roles).
func (c *Claims) CanAccessCRM() bool {
	return c.HasFullOperationalAccess() || c.IsAsesorComercial()
}

// CanAccessVentas indica si el usuario puede consultar, crear y modificar ventas.
func (c *Claims) CanAccessVentas() bool {
	return c.HasFullOperationalAccess() || c.IsGestorVentas()
}

// CanDeleteVentas indica si el usuario puede eliminar ventas.
func (c *Claims) CanDeleteVentas() bool {
	return c.HasFullOperationalAccess()
}

// CanManageTercerosFull indica si el usuario puede gestionar todos los roles de terceros y eliminar.
func (c *Claims) CanManageTercerosFull() bool {
	return c.HasFullOperationalAccess()
}

// CanManageTercerosCliente indica si el usuario puede crear/modificar terceros (al menos como CLIENTE).
func (c *Claims) CanManageTercerosCliente() bool {
	return c.HasFullOperationalAccess() || c.IsGestorVentas()
}

// IsOnlyTercerosCliente indica si el usuario en Terceros está limitado al rol CLIENTE.
func (c *Claims) IsOnlyTercerosCliente() bool {
	return !c.CanManageTercerosFull() && c.CanManageTercerosCliente()
}

// CanManageOperacion indica si el usuario puede escribir en Compras, Inventario, Productos y Precios.
func (c *Claims) CanManageOperacion() bool {
	return c.HasFullOperationalAccess()
}

// IsOnlyGestorVentas indica si el usuario está restringido al rol GESTOR_VENTAS
// (sin privilegios de ADMINISTRADOR ni ADMINISTRADOR_SUCURSAL).
func (c *Claims) IsOnlyGestorVentas() bool {
	return c.IsGestorVentas() && !c.HasFullOperationalAccess()
}

// IsAsesorComercial indica si el token posee el rol ASESOR_COMERCIAL.
func (c *Claims) IsAsesorComercial() bool {
	if c == nil {
		return false
	}
	if c.EsAsesorComercial {
		return true
	}
	for _, r := range c.Roles {
		if strings.EqualFold(r, RolAsesorComercial) {
			return true
		}
	}
	return false
}

// IsOnlyAsesorComercial indica si el usuario está restringido al rol ASESOR_COMERCIAL
// (sin privilegios de ADMINISTRADOR ni ADMINISTRADOR_SUCURSAL).
func (c *Claims) IsOnlyAsesorComercial() bool {
	return c.IsAsesorComercial() && !c.CanManageUsers()
}

// HasSucursal indica si el usuario autenticado tiene acceso a la sucursal indicada.
func (c *Claims) HasSucursal(sucursalID int32) bool {
	if c == nil || sucursalID <= 0 {
		return false
	}
	if c.IsAdmin() {
		return true
	}
	for _, id := range c.SucursalIDs {
		if id == sucursalID {
			return true
		}
	}
	return c.SucursalID > 0 && c.SucursalID == sucursalID
}

// AllowedSucursalIDs devuelve los IDs de sucursales a los que el usuario tiene acceso.
// Devuelve nil para ADMINISTRADOR (acceso a todas las sucursales de la empresa).
// Si un usuario no administrador no tiene ninguna sucursal asignada, devuelve []int32{-1}
// para garantizar que no vea registros de otras sucursales (fail-closed).
func (c *Claims) AllowedSucursalIDs() []int32 {
	if c == nil || c.IsAdmin() {
		return nil
	}
	if len(c.SucursalIDs) > 0 {
		return c.SucursalIDs
	}
	if c.SucursalID > 0 {
		return []int32{c.SucursalID}
	}
	return []int32{-1}
}

var (
	ErrTokenInvalido = errors.New("token inválido")
	ErrTokenExpirado = errors.New("token expirado")
)

var jwtHeader = base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))

// SignHS256 genera un JWT firmado con HMAC-SHA256.
func SignHS256(claims Claims, secret string) (string, error) {
	payload, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	encodedPayload := base64.RawURLEncoding.EncodeToString(payload)
	signingInput := jwtHeader + "." + encodedPayload

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(signingInput))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return signingInput + "." + sig, nil
}

// VerifyHS256 verifica firma y expiración de un JWT HS256.
func VerifyHS256(token, secret string) (*Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, ErrTokenInvalido
	}
	signingInput := parts[0] + "." + parts[1]
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(signingInput))
	expectedSig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(parts[2]), []byte(expectedSig)) {
		return nil, ErrTokenInvalido
	}

	raw, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, ErrTokenInvalido
	}
	var c Claims
	if err := json.Unmarshal(raw, &c); err != nil {
		return nil, ErrTokenInvalido
	}
	if time.Now().UTC().Unix() > c.Exp {
		return nil, ErrTokenExpirado
	}
	return &c, nil
}
