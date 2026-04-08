package handler

import "github.com/gin-gonic/gin"

type AuditHandler struct{}

func (h *AuditHandler) ListAuditLogs(c *gin.Context) {}
func (h *AuditHandler) GetAuditLog(c *gin.Context)   {}
