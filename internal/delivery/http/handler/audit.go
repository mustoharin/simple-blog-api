package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	audituc "simple-blog-api/internal/usecase/audit"
)

type AuditHandler struct {
	list *audituc.ListAuditLogsUsecase
	get  *audituc.GetAuditLogUsecase
}

func NewAuditHandler(list *audituc.ListAuditLogsUsecase, get *audituc.GetAuditLogUsecase) *AuditHandler {
	return &AuditHandler{list: list, get: get}
}

func (h *AuditHandler) ListAuditLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	in := audituc.ListInput{
		ActorID:      c.Query("actor_id"),
		Action:       c.Query("action"),
		ResourceType: c.Query("resource_type"),
		ResourceID:   c.Query("resource_id"),
		Page:         page,
		Limit:        limit,
	}

	if fromStr := c.Query("from"); fromStr != "" {
		if t, err := time.Parse(time.RFC3339, fromStr); err == nil {
			in.From = &t
		}
	}
	if toStr := c.Query("to"); toStr != "" {
		if t, err := time.Parse(time.RFC3339, toStr); err == nil {
			in.To = &t
		}
	}

	out, err := h.list.Execute(c.Request.Context(), in)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"logs":  out.Logs,
		"total": out.Total,
		"page":  out.Page,
		"limit": out.Limit,
	})
}

func (h *AuditHandler) GetAuditLog(c *gin.Context) {
	log, err := h.get.Execute(c.Request.Context(), c.Param("id"))
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, log)
}
