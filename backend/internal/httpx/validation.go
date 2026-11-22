package httpx

import (
	"errors"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
)

type FieldError struct {
	Field string `json:"field"`
	Rule  string `json:"rule"`
	Param string `json:"param,omitempty"`
}

// BindError convierte errores de binding/validación de Gin en un AppError 400.
func BindError(err error) *AppError {
	var verrs validator.ValidationErrors
	if errors.As(err, &verrs) {
		fields := make([]FieldError, 0, len(verrs))
		for _, fe := range verrs {
			fields = append(fields, FieldError{
				Field: toSnake(fe.Field()),
				Rule:  fe.Tag(),
				Param: fe.Param(),
			})
		}
		return &AppError{
			Status:  http.StatusBadRequest,
			Code:    "validacion",
			Message: "datos de entrada inválidos",
			Details: fields,
		}
	}
	return &AppError{
		Status:  http.StatusBadRequest,
		Code:    "solicitud_invalida",
		Message: "cuerpo o parámetros de la solicitud inválidos: " + err.Error(),
	}
}

// toSnake convierte "NombreUsuario" -> "nombre_usuario" y "SucursalID" -> "sucursal_id".
func toSnake(s string) string {
	var b strings.Builder
	r := []rune(s)
	for i, c := range r {
		if c >= 'A' && c <= 'Z' {
			if i > 0 && (r[i-1] < 'A' || r[i-1] > 'Z') {
				b.WriteByte('_')
			}
			b.WriteRune(c + ('a' - 'A'))
			continue
		}
		b.WriteRune(c)
	}
	return b.String()
}
