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

// Claims contiene la información firmada dentro del JWT.
type Claims struct {
	Sub        int64  `json:"sub"`
	EmpresaID  int32  `json:"empresa_id"`
	SucursalID int32  `json:"sucursal_id"`
	Email      string `json:"email"`
	Iat        int64  `json:"iat"`
	Exp        int64  `json:"exp"`
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
