// Package http provides shared HTTP helpers for Gin controllers, including
// a BindJSON wrapper that maps validator.ValidationErrors to stable error codes
// instead of leaking raw internal validator strings to clients.
package http

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/Gooowan/matchup/modules/core/types"
)

// fieldErrorCode maps a struct field name + validator tag to a stable error code
// suitable for frontend i18n lookup.
func fieldErrorCode(field, tag string) string {
	f := strings.ToLower(field)
	switch tag {
	case "required":
		switch f {
		case "email":
			return "REQUIRED_EMAIL"
		case "password":
			return "REQUIRED_PASSWORD"
		default:
			return "REQUIRED_FIELD"
		}
	case "email":
		return "INVALID_EMAIL"
	case "min":
		switch f {
		case "password", "newpassword":
			return "PASSWORD_TOO_SHORT"
		default:
			return "VALUE_TOO_SHORT"
		}
	case "max":
		return "VALUE_TOO_LONG"
	case "gte":
		return "VALUE_TOO_SMALL"
	case "lte":
		return "VALUE_TOO_LARGE"
	case "oneof":
		return "INVALID_OPTION"
	case "url", "http_url":
		return "INVALID_URL"
	case "e164":
		return "INVALID_PHONE"
	case "uuid", "uuid4":
		return "INVALID_ID"
	default:
		return "INVALID_FIELD"
	}
}

// FormatBindingError converts a binding/validation error into a stable error
// code and a generic user-friendly message. The error code is intended to be
// used by the frontend for i18n mapping.
func FormatBindingError(err error) (code string, msg string) {
	var ve validator.ValidationErrors
	if errors.As(err, &ve) && len(ve) > 0 {
		fe := ve[0]
		code = fieldErrorCode(fe.Field(), fe.Tag())
		return code, "Invalid request"
	}
	return "INVALID_REQUEST", "Invalid request"
}

// BindJSON calls ctx.ShouldBindJSON(dst). On failure it writes a 400 response
// with a stable error_code + generic message and returns false. Controllers
// should return immediately when BindJSON returns false.
func BindJSON(ctx *gin.Context, dst any) bool {
	if err := ctx.ShouldBindJSON(dst); err != nil {
		code, msg := FormatBindingError(err)
		ctx.JSON(http.StatusBadRequest, types.Resp{ErrorCode: code, Error: msg})
		return false
	}
	return true
}
