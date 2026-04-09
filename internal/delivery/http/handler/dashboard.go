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

// GetDashboard godoc
// @Summary      Get dashboard data
// @Description  Returns dashboard statistics and metrics
// @Tags         dashboard
// @Produce      json
// @Security     BearerAuth
// @Param        range  query  int  false  "Date range in days"  default(30)
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  errorResponse
// @Failure      401  {object}  errorResponse
// @Failure      403  {object}  errorResponse
// @Failure      500  {object}  errorResponse
// @Router       /dashboard [get]
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
