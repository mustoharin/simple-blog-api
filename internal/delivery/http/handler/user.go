package handler

import "github.com/gin-gonic/gin"

type UserHandler struct{}

func (h *UserHandler) ListUsers(c *gin.Context)        {}
func (h *UserHandler) CreateUser(c *gin.Context)       {}
func (h *UserHandler) UpdateUser(c *gin.Context)       {}
func (h *UserHandler) DeleteUser(c *gin.Context)       {}
func (h *UserHandler) AssignRole(c *gin.Context)       {}
func (h *UserHandler) RemoveRole(c *gin.Context)       {}
func (h *UserHandler) ResendInvitation(c *gin.Context) {}
