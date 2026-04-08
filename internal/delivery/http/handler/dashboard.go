package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	dashboarduc "simple-blog-api/internal/usecase/dashboard"
)

type DashboardHandler struct {
	get *dashboarduc.GetDashboardUsecase
}

func NewDashboardHandler(get *dashboarduc.GetDashboardUsecase) *DashboardHandler {
	return &DashboardHandler{get: get}
}

func (h *DashboardHandler) GetDashboard(c *gin.Context) {
	rangeDays := 30
	if r := c.Query("range"); r != "" {
		if v, err := strconv.Atoi(r); err == nil {
			rangeDays = v
		}
	}

	out, err := h.get.Execute(c.Request.Context(), rangeDays)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, out)
}
