// Package httpx contiene utilidades comunes para las respuestas HTTP.
package httpx

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

// AppError es un error de dominio con código HTTP asociado.
type AppError struct {
	Status  int    `json:"-"`
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

func (e *AppError) Error() string { return e.Message }

func NewError(status int, code, msg string) *AppError {
	return &AppError{Status: status, Code: code, Message: msg}
}

var (
	ErrNotFound = NewError(http.StatusNotFound, "not_found", "recurso no encontrado")
)

// Respond escribe un error en formato JSON. Si no es un AppError, responde 500.
func RespondError(c *gin.Context, err error) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		c.AbortWithStatusJSON(appErr.Status, gin.H{"error": appErr})
		return
	}
	_ = c.Error(err)
	c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": AppError{
		Code:    "internal_error",
		Message: "error interno del servidor",
	}})
}
