package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"simple-blog-api/internal/delivery/http/middleware"
	"simple-blog-api/internal/usecase/profile"
)

type ProfileHandler struct {
	getMe          *profile.GetMeUsecase
	updateMe       *profile.UpdateMeUsecase
	changePassword *profile.ChangePasswordUsecase
}

func NewProfileHandler(
	getMe *profile.GetMeUsecase,
	updateMe *profile.UpdateMeUsecase,
	changePassword *profile.ChangePasswordUsecase,
) *ProfileHandler {
	return &ProfileHandler{
		getMe:          getMe,
		updateMe:       updateMe,
		changePassword: changePassword,
	}
}

// GetMe godoc
// @Summary      Get current user profile
// @Description  Returns the authenticated user's profile
// @Tags         profile
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {object}  errorResponse
// @Failure      500  {object}  errorResponse
// @Router       /me [get]
func (h *ProfileHandler) GetMe(c *gin.Context) {
	userID, _ := c.Get(middleware.ContextKeyUserID)
	u, err := h.getMe.Execute(c.Request.Context(), userID.(string))
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, u)
}

type updateMeRequest struct {
	DisplayName string `json:"display_name"`
	Bio         string `json:"bio"`
	AvatarURL   string `json:"avatar_url"`
}

// UpdateMe godoc
// @Summary      Update current user profile
// @Description  Updates the authenticated user's profile
// @Tags         profile
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body  updateMeRequest  true  "Profile data"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  errorResponse
// @Failure      401  {object}  errorResponse
// @Failure      500  {object}  errorResponse
// @Router       /me [patch]
func (h *ProfileHandler) UpdateMe(c *gin.Context) {
	var req updateMeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{err.Error(), "VALIDATION_ERROR"})
		return
	}
	userID, _ := c.Get(middleware.ContextKeyUserID)
	u, err := h.updateMe.Execute(c.Request.Context(), profile.UpdateMeInput{
		UserID:      userID.(string),
		DisplayName: req.DisplayName,
		Bio:         req.Bio,
		AvatarURL:   req.AvatarURL,
	})
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, u)
}

type changePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password"     binding:"required"`
}

// ChangePassword godoc
// @Summary      Change password
// @Description  Changes the authenticated user's password
// @Tags         profile
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body  changePasswordRequest  true  "Current and new password"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  errorResponse
// @Failure      401  {object}  errorResponse
// @Failure      422  {object}  errorResponse
// @Failure      500  {object}  errorResponse
// @Router       /me/change-password [post]
func (h *ProfileHandler) ChangePassword(c *gin.Context) {
	var req changePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{err.Error(), "VALIDATION_ERROR"})
		return
	}
	userID, _ := c.Get(middleware.ContextKeyUserID)
	if err := h.changePassword.Execute(
		c.Request.Context(),
		userID.(string),
		req.CurrentPassword,
		req.NewPassword,
	); err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
}
