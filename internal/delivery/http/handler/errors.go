package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"simple-blog-api/internal/domain"
)

type errorResponse struct {
	Error string `json:"error"`
	Code  string `json:"code"`
}

func respondError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		c.JSON(http.StatusNotFound, errorResponse{"Resource not found", "NOT_FOUND"})
	case errors.Is(err, domain.ErrEmailAlreadyExists):
		c.JSON(http.StatusConflict, errorResponse{"Email already registered", "EMAIL_EXISTS"})
	case errors.Is(err, domain.ErrSlugAlreadyExists):
		c.JSON(http.StatusConflict, errorResponse{"Slug already exists", "SLUG_EXISTS"})
	case errors.Is(err, domain.ErrInvalidCredentials):
		c.JSON(http.StatusUnauthorized, errorResponse{"Invalid email or password", "INVALID_CREDENTIALS"})
	case errors.Is(err, domain.ErrAccountNotActivated):
		c.JSON(http.StatusForbidden, errorResponse{"Account not activated", "ACCOUNT_NOT_ACTIVATED"})
	case errors.Is(err, domain.ErrTokenExpired):
		c.JSON(http.StatusUnprocessableEntity, errorResponse{"Token has expired", "TOKEN_EXPIRED"})
	case errors.Is(err, domain.ErrTokenNotFound):
		c.JSON(http.StatusUnprocessableEntity, errorResponse{"Token not found", "TOKEN_NOT_FOUND"})
	case errors.Is(err, domain.ErrTokenUsed):
		c.JSON(http.StatusUnprocessableEntity, errorResponse{"Token already used", "TOKEN_USED"})
	case errors.Is(err, domain.ErrInvalidCaptcha):
		c.JSON(http.StatusUnprocessableEntity, errorResponse{"Invalid captcha", "INVALID_CAPTCHA"})
	case errors.Is(err, domain.ErrForbidden):
		c.JSON(http.StatusForbidden, errorResponse{"Forbidden", "FORBIDDEN"})
	case errors.Is(err, domain.ErrPasswordTooWeak):
		c.JSON(http.StatusUnprocessableEntity, errorResponse{err.Error(), "PASSWORD_TOO_WEAK"})
	case errors.Is(err, domain.ErrPasswordPwned):
		c.JSON(http.StatusUnprocessableEntity, errorResponse{err.Error(), "PASSWORD_PWNED"})
	case errors.Is(err, domain.ErrUserAlreadyActive):
		c.JSON(http.StatusConflict, errorResponse{"User is already active", "USER_ALREADY_ACTIVE"})
	case errors.Is(err, domain.ErrInvalidRange):
		c.JSON(http.StatusBadRequest, errorResponse{err.Error(), "INVALID_RANGE"})
	case errors.Is(err, domain.ErrInvalidMimeType):
		c.JSON(http.StatusUnprocessableEntity, errorResponse{err.Error(), "INVALID_MIME_TYPE"})
	case errors.Is(err, domain.ErrFileTooLarge):
		c.JSON(http.StatusRequestEntityTooLarge, errorResponse{err.Error(), "FILE_TOO_LARGE"})
	case errors.Is(err, domain.ErrTagNameAlreadyExists):
		c.JSON(http.StatusConflict, errorResponse{"Tag name already exists", "TAG_EXISTS"})
	default:
		c.JSON(http.StatusInternalServerError, errorResponse{"Internal server error", "INTERNAL_ERROR"})
	}
}
