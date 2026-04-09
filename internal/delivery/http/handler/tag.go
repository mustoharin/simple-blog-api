package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"simple-blog-api/internal/delivery/http/middleware"
	taguc "simple-blog-api/internal/usecase/tag"
)

type TagHandler struct {
	create *taguc.CreateTagUsecase
	list   *taguc.ListTagsUsecase
	delete *taguc.DeleteTagUsecase
}

func NewTagHandler(
	create *taguc.CreateTagUsecase,
	list *taguc.ListTagsUsecase,
	deleteUC *taguc.DeleteTagUsecase,
) *TagHandler {
	return &TagHandler{create: create, list: list, delete: deleteUC}
}

func (h *TagHandler) ListTags(c *gin.Context) {
	tags, err := h.list.Execute(c.Request.Context())
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"tags": tags})
}

type createTagRequest struct {
	Name string `json:"name" binding:"required"`
}

func (h *TagHandler) CreateTag(c *gin.Context) {
	var req createTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{err.Error(), "VALIDATION_ERROR"})
		return
	}
	actorID, _ := c.Get(middleware.ContextKeyUserID)
	actorEmail, _ := c.Get(middleware.ContextKeyEmail)
	t, err := h.create.Execute(c.Request.Context(), req.Name, actorID.(string), actorEmail.(string))
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, t)
}

func (h *TagHandler) DeleteTag(c *gin.Context) {
	actorID, _ := c.Get(middleware.ContextKeyUserID)
	actorEmail, _ := c.Get(middleware.ContextKeyEmail)
	if err := h.delete.Execute(c.Request.Context(),
		c.Param("id"), actorID.(string), actorEmail.(string)); err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Tag deleted"})
}
