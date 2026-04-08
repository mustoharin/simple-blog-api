package handler

import "github.com/gin-gonic/gin"

type ProfileHandler struct{}

func (h *ProfileHandler) GetMe(c *gin.Context)            {}
func (h *ProfileHandler) UpdateMe(c *gin.Context)         {}
func (h *ProfileHandler) ChangePassword(c *gin.Context)   {}
