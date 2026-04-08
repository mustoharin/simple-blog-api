package handler

import (
    "net/http"
    "strconv"

    "github.com/gin-gonic/gin"

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
