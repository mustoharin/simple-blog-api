package domain

import "errors"

var (
	ErrNotFound             = errors.New("resource not found")
	ErrEmailAlreadyExists   = errors.New("email already exists")
	ErrSlugAlreadyExists    = errors.New("slug already exists")
	ErrInvalidCredentials   = errors.New("invalid credentials")
	ErrAccountNotActivated  = errors.New("account not activated")
	ErrTokenExpired         = errors.New("token expired")
	ErrTokenUsed            = errors.New("token already used")
	ErrTokenNotFound        = errors.New("token not found")
	ErrInvalidCaptcha       = errors.New("invalid captcha")
	ErrForbidden            = errors.New("forbidden")
	ErrUserAlreadyActive    = errors.New("user is already active")
	ErrInvalidRange         = errors.New("invalid range; allowed values: 7, 30, 90")
	ErrInvalidMimeType      = errors.New("invalid file type; allowed: jpeg, png, gif, webp")
	ErrFileTooLarge         = errors.New("file too large; maximum 10MB")
	ErrPasswordTooWeak      = errors.New("password does not meet complexity requirements")
	ErrPasswordPwned        = errors.New("password has been found in a data breach; choose a different one")
	ErrTagNameAlreadyExists = errors.New("tag name already exists")
)
