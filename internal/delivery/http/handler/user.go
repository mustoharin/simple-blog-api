package handler

import (
    "net/http"
    "strconv"

    "github.com/gin-gonic/gin"

    _ "simple-blog-api/internal/domain"
    "simple-blog-api/internal/delivery/http/middleware"
    "simple-blog-api/internal/usecase/user"
)

type UserHandler struct {
    create           *user.CreateUserUsecase
    list             *user.ListUsersUsecase
    update           *user.UpdateUserUsecase
    delete           *user.DeleteUserUsecase
    assignRole       *user.AssignRoleUsecase
    removeRole       *user.RemoveRoleUsecase
    resendInvitation *user.ResendInvitationUsecase
}

func NewUserHandler(
    create *user.CreateUserUsecase,
    list *user.ListUsersUsecase,
    update *user.UpdateUserUsecase,
    deleteUC *user.DeleteUserUsecase,
    assignRole *user.AssignRoleUsecase,
    removeRole *user.RemoveRoleUsecase,
    resendInvitation *user.ResendInvitationUsecase,
) *UserHandler {
    return &UserHandler{
        create:           create,
        list:             list,
        update:           update,
        delete:           deleteUC,
        assignRole:       assignRole,
        removeRole:       removeRole,
        resendInvitation: resendInvitation,
    }
}

// ListUsers godoc
// @Summary      List users
// @Description  Returns a paginated list of users
// @Tags         users
// @Produce      json
// @Security     BearerAuth
// @Param        page   query  int  false  "Page number"     default(1)
// @Param        limit  query  int  false  "Items per page"  default(20)
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {object}  errorResponse
// @Failure      403  {object}  errorResponse
// @Failure      500  {object}  errorResponse
// @Router       /users [get]
func (h *UserHandler) ListUsers(c *gin.Context) {
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
    out, err := h.list.Execute(c.Request.Context(), page, limit)
    if err != nil {
        respondError(c, err)
        return
    }
    c.JSON(http.StatusOK, gin.H{
        "users": out.Users,
        "total": out.Total,
        "page":  out.Page,
        "limit": out.Limit,
    })
}

type createUserRequest struct {
    Email       string `json:"email"        binding:"required,email"`
    DisplayName string `json:"display_name"`
}

// CreateUser godoc
// @Summary      Create a user
// @Description  Creates a new user and sends an invitation email
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body  createUserRequest  true  "User data"
// @Success      201  {object}  domain.User
// @Failure      400  {object}  errorResponse
// @Failure      401  {object}  errorResponse
// @Failure      403  {object}  errorResponse
// @Failure      409  {object}  errorResponse
// @Failure      500  {object}  errorResponse
// @Router       /users [post]
func (h *UserHandler) CreateUser(c *gin.Context) {
    var req createUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, errorResponse{err.Error(), "VALIDATION_ERROR"})
        return
    }
    actorID, _ := c.Get(middleware.ContextKeyUserID)
    actorEmail, _ := c.Get(middleware.ContextKeyEmail)
    u, err := h.create.Execute(c.Request.Context(), user.CreateUserInput{
        Email:       req.Email,
        DisplayName: req.DisplayName,
        ActorID:     actorID.(string),
        ActorEmail:  actorEmail.(string),
    })
    if err != nil {
        respondError(c, err)
        return
    }
    c.JSON(http.StatusCreated, u)
}

type updateUserRequest struct {
    DisplayName string `json:"display_name"`
    Bio         string `json:"bio"`
    AvatarURL   string `json:"avatar_url"`
}

// UpdateUser godoc
// @Summary      Update a user
// @Description  Updates a user's profile fields
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path  string            true  "User ID"
// @Param        body  body  updateUserRequest  true  "User data"
// @Success      200  {object}  domain.User
// @Failure      400  {object}  errorResponse
// @Failure      401  {object}  errorResponse
// @Failure      403  {object}  errorResponse
// @Failure      404  {object}  errorResponse
// @Failure      500  {object}  errorResponse
// @Router       /users/{id} [put]
func (h *UserHandler) UpdateUser(c *gin.Context) {
    var req updateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, errorResponse{err.Error(), "VALIDATION_ERROR"})
        return
    }
    actorID, _ := c.Get(middleware.ContextKeyUserID)
    actorEmail, _ := c.Get(middleware.ContextKeyEmail)
    u, err := h.update.Execute(c.Request.Context(), user.UpdateUserInput{
        ID:          c.Param("id"),
        DisplayName: req.DisplayName,
        Bio:         req.Bio,
        AvatarURL:   req.AvatarURL,
        ActorID:     actorID.(string),
        ActorEmail:  actorEmail.(string),
    })
    if err != nil {
        respondError(c, err)
        return
    }
    c.JSON(http.StatusOK, u)
}

// DeleteUser godoc
// @Summary      Delete a user
// @Description  Deletes a user account
// @Tags         users
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  string  true  "User ID"
// @Success      200  {object}  map[string]string
// @Failure      401  {object}  errorResponse
// @Failure      403  {object}  errorResponse
// @Failure      404  {object}  errorResponse
// @Failure      500  {object}  errorResponse
// @Router       /users/{id} [delete]
func (h *UserHandler) DeleteUser(c *gin.Context) {
    actorID, _ := c.Get(middleware.ContextKeyUserID)
    actorEmail, _ := c.Get(middleware.ContextKeyEmail)
    if err := h.delete.Execute(c.Request.Context(),
        c.Param("id"), actorID.(string), actorEmail.(string)); err != nil {
        respondError(c, err)
        return
    }
    c.JSON(http.StatusOK, gin.H{"message": "User deleted"})
}

type assignRoleRequest struct {
    RoleID string `json:"role_id" binding:"required"`
}

// AssignRole godoc
// @Summary      Assign a role to a user
// @Description  Assigns a role to a user
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path  string            true  "User ID"
// @Param        body  body  assignRoleRequest  true  "Role ID"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  errorResponse
// @Failure      401  {object}  errorResponse
// @Failure      403  {object}  errorResponse
// @Failure      404  {object}  errorResponse
// @Failure      500  {object}  errorResponse
// @Router       /users/{id}/roles [post]
func (h *UserHandler) AssignRole(c *gin.Context) {
    var req assignRoleRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, errorResponse{err.Error(), "VALIDATION_ERROR"})
        return
    }
    actorID, _ := c.Get(middleware.ContextKeyUserID)
    actorEmail, _ := c.Get(middleware.ContextKeyEmail)
    if err := h.assignRole.Execute(c.Request.Context(),
        c.Param("id"), req.RoleID, actorID.(string), actorEmail.(string)); err != nil {
        respondError(c, err)
        return
    }
    c.JSON(http.StatusOK, gin.H{"message": "Role assigned"})
}

// RemoveRole godoc
// @Summary      Remove a role from a user
// @Description  Removes a role from a user
// @Tags         users
// @Produce      json
// @Security     BearerAuth
// @Param        id      path  string  true  "User ID"
// @Param        roleId  path  string  true  "Role ID"
// @Success      200  {object}  map[string]string
// @Failure      401  {object}  errorResponse
// @Failure      403  {object}  errorResponse
// @Failure      404  {object}  errorResponse
// @Failure      500  {object}  errorResponse
// @Router       /users/{id}/roles/{roleId} [delete]
func (h *UserHandler) RemoveRole(c *gin.Context) {
    actorID, _ := c.Get(middleware.ContextKeyUserID)
    actorEmail, _ := c.Get(middleware.ContextKeyEmail)
    if err := h.removeRole.Execute(c.Request.Context(),
        c.Param("id"), c.Param("roleId"), actorID.(string), actorEmail.(string)); err != nil {
        respondError(c, err)
        return
    }
    c.JSON(http.StatusOK, gin.H{"message": "Role removed"})
}

// ResendInvitation godoc
// @Summary      Resend invitation
// @Description  Resends the invitation email to a user
// @Tags         users
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  string  true  "User ID"
// @Success      200  {object}  map[string]string
// @Failure      401  {object}  errorResponse
// @Failure      403  {object}  errorResponse
// @Failure      404  {object}  errorResponse
// @Failure      409  {object}  errorResponse
// @Failure      500  {object}  errorResponse
// @Router       /users/{id}/resend-invitation [post]
func (h *UserHandler) ResendInvitation(c *gin.Context) {
    actorID, _ := c.Get(middleware.ContextKeyUserID)
    actorEmail, _ := c.Get(middleware.ContextKeyEmail)
    if err := h.resendInvitation.Execute(c.Request.Context(),
        c.Param("id"), actorID.(string), actorEmail.(string)); err != nil {
        respondError(c, err)
        return
    }
    c.JSON(http.StatusOK, gin.H{"message": "Invitation resent"})
}
