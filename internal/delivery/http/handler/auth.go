package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"simple-blog-api/internal/delivery/http/middleware"
	"simple-blog-api/internal/usecase/auth"
)

type AuthHandler struct {
	register         *auth.RegisterUsecase
	login            *auth.LoginUsecase
	refresh          *auth.RefreshUsecase
	logout           *auth.LogoutUsecase
	forgotPassword   *auth.ForgotPasswordUsecase
	resetPassword    *auth.ResetPasswordUsecase
	acceptInvitation *auth.AcceptInvitationUsecase
}

func NewAuthHandler(
	register *auth.RegisterUsecase,
	login *auth.LoginUsecase,
	refresh *auth.RefreshUsecase,
	logout *auth.LogoutUsecase,
	forgotPassword *auth.ForgotPasswordUsecase,
	resetPassword *auth.ResetPasswordUsecase,
	acceptInvitation *auth.AcceptInvitationUsecase,
) *AuthHandler {
	return &AuthHandler{
		register:         register,
		login:            login,
		refresh:          refresh,
		logout:           logout,
		forgotPassword:   forgotPassword,
		resetPassword:    resetPassword,
		acceptInvitation: acceptInvitation,
	}
}

type registerRequest struct {
	Email       string `json:"email"    binding:"required,email"`
	Password    string `json:"password" binding:"required"`
	DisplayName string `json:"display_name"`
}

// Register godoc
// @Summary      Register a new user
// @Description  Creates a new user account
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body  registerRequest  true  "Registration request"
// @Success      201  {object}  map[string]string
// @Failure      400  {object}  errorResponse
// @Failure      409  {object}  errorResponse
// @Failure      422  {object}  errorResponse
// @Failure      500  {object}  errorResponse
// @Router       /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{err.Error(), "VALIDATION_ERROR"})
		return
	}
	err := h.register.Execute(c.Request.Context(), auth.RegisterInput{
		Email:       req.Email,
		Password:    req.Password,
		DisplayName: req.DisplayName,
	})
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Registration successful"})
}

type loginRequest struct {
	Email        string `json:"email"         binding:"required,email"`
	Password     string `json:"password"      binding:"required"`
	CaptchaToken string `json:"captcha_token" binding:"required"`
}

// Login godoc
// @Summary      Login
// @Description  Authenticates a user and returns JWT tokens
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body  loginRequest  true  "Login credentials"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  errorResponse
// @Failure      401  {object}  errorResponse
// @Failure      403  {object}  errorResponse
// @Failure      422  {object}  errorResponse
// @Failure      500  {object}  errorResponse
// @Router       /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{err.Error(), "VALIDATION_ERROR"})
		return
	}
	out, err := h.login.Execute(c.Request.Context(), auth.LoginInput{
		Email:        req.Email,
		Password:     req.Password,
		CaptchaToken: req.CaptchaToken,
		IPAddress:    c.ClientIP(),
		UserAgent:    c.Request.UserAgent(),
	})
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"access_token":  out.AccessToken,
		"refresh_token": out.RefreshToken,
	})
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// Refresh godoc
// @Summary      Refresh tokens
// @Description  Exchanges a refresh token for new JWT tokens
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body  refreshRequest  true  "Refresh token"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  errorResponse
// @Failure      422  {object}  errorResponse
// @Failure      500  {object}  errorResponse
// @Router       /auth/refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{err.Error(), "VALIDATION_ERROR"})
		return
	}
	out, err := h.refresh.Execute(c.Request.Context(), req.RefreshToken)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"access_token":  out.AccessToken,
		"refresh_token": out.RefreshToken,
	})
}

// Logout godoc
// @Summary      Logout
// @Description  Invalidates the provided refresh token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body  refreshRequest  true  "Refresh token to invalidate"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  errorResponse
// @Failure      401  {object}  errorResponse
// @Failure      500  {object}  errorResponse
// @Router       /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{err.Error(), "VALIDATION_ERROR"})
		return
	}
	userID, _ := c.Get(middleware.ContextKeyUserID)
	email, _ := c.Get(middleware.ContextKeyEmail)
	_ = h.logout.Execute(c.Request.Context(), req.RefreshToken,
		userID.(string), email.(string))
	c.JSON(http.StatusOK, gin.H{"message": "Logged out"})
}

type forgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// ForgotPassword godoc
// @Summary      Forgot password
// @Description  Sends a password reset email (always returns 200 to prevent enumeration)
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body  forgotPasswordRequest  true  "Email address"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  errorResponse
// @Router       /auth/forgot-password [post]
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req forgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{err.Error(), "VALIDATION_ERROR"})
		return
	}
	_ = h.forgotPassword.Execute(c.Request.Context(), req.Email)
	// Always 200 to prevent email enumeration
	c.JSON(http.StatusOK, gin.H{"message": "If that email exists, a reset link has been sent"})
}

type resetPasswordRequest struct {
	Token       string `json:"token"        binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

// ResetPassword godoc
// @Summary      Reset password
// @Description  Resets a user's password using a valid reset token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body  resetPasswordRequest  true  "Reset token and new password"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  errorResponse
// @Failure      422  {object}  errorResponse
// @Failure      500  {object}  errorResponse
// @Router       /auth/reset-password [post]
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req resetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{err.Error(), "VALIDATION_ERROR"})
		return
	}
	if err := h.resetPassword.Execute(c.Request.Context(), req.Token, req.NewPassword); err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Password reset successful"})
}

type acceptInvitationRequest struct {
	Token       string `json:"token"        binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

// AcceptInvitation godoc
// @Summary      Accept invitation
// @Description  Activates an invited user account using a valid invitation token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body  acceptInvitationRequest  true  "Invitation token and new password"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  errorResponse
// @Failure      422  {object}  errorResponse
// @Failure      500  {object}  errorResponse
// @Router       /auth/accept-invitation [post]
func (h *AuthHandler) AcceptInvitation(c *gin.Context) {
	var req acceptInvitationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{err.Error(), "VALIDATION_ERROR"})
		return
	}
	if err := h.acceptInvitation.Execute(c.Request.Context(), req.Token, req.NewPassword); err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Invitation accepted, account activated"})
}
