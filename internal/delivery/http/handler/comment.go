package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	_ "simple-blog-api/internal/domain"
	"simple-blog-api/internal/delivery/http/middleware"
	commentuc "simple-blog-api/internal/usecase/comment"
)

type CommentHandler struct {
	create       *commentuc.CreateCommentUsecase
	list         *commentuc.ListCommentsUsecase
	updateStatus *commentuc.UpdateStatusUsecase
	delete       *commentuc.DeleteCommentUsecase
}

func NewCommentHandler(
	create *commentuc.CreateCommentUsecase,
	list *commentuc.ListCommentsUsecase,
	updateStatus *commentuc.UpdateStatusUsecase,
	deleteUC *commentuc.DeleteCommentUsecase,
) *CommentHandler {
	return &CommentHandler{create: create, list: list, updateStatus: updateStatus, delete: deleteUC}
}

// ListComments godoc
// @Summary      List comments for a post
// @Description  Returns paginated comments for a given post
// @Tags         comments
// @Produce      json
// @Param        id     path   string  true   "Post ID"
// @Param        page   query  int     false  "Page number"     default(1)
// @Param        limit  query  int     false  "Items per page"  default(20)
// @Success      200  {object}  map[string]interface{}
// @Failure      404  {object}  errorResponse
// @Failure      500  {object}  errorResponse
// @Router       /posts/{id}/comments [get]
func (h *CommentHandler) ListComments(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	out, err := h.list.Execute(c.Request.Context(), c.Param("id"), page, limit)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"comments": out.Comments,
		"total":    out.Total,
		"page":     out.Page,
		"limit":    out.Limit,
	})
}

type createCommentRequest struct {
	Body string `json:"body" binding:"required"`
}

// CreateComment godoc
// @Summary      Create a comment
// @Description  Adds a comment to a post
// @Tags         comments
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path  string               true  "Post ID"
// @Param        body  body  createCommentRequest  true  "Comment body"
// @Success      201  {object}  domain.Comment
// @Failure      400  {object}  errorResponse
// @Failure      401  {object}  errorResponse
// @Failure      404  {object}  errorResponse
// @Failure      500  {object}  errorResponse
// @Router       /posts/{id}/comments [post]
func (h *CommentHandler) CreateComment(c *gin.Context) {
	var req createCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{err.Error(), "VALIDATION_ERROR"})
		return
	}
	actorID := c.GetString(middleware.ContextKeyUserID)
	actorEmail := c.GetString(middleware.ContextKeyEmail)
	comment, err := h.create.Execute(c.Request.Context(), commentuc.CreateCommentInput{
		PostID:      c.Param("id"),
		AuthorID:    actorID,
		AuthorEmail: actorEmail,
		Body:        req.Body,
	})
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, comment)
}

type updateCommentStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=approved rejected"`
}

// UpdateCommentStatus godoc
// @Summary      Update comment status
// @Description  Approves or rejects a comment
// @Tags         comments
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path  string                      true  "Comment ID"
// @Param        body  body  updateCommentStatusRequest  true  "Status (approved|rejected)"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  errorResponse
// @Failure      401  {object}  errorResponse
// @Failure      403  {object}  errorResponse
// @Failure      404  {object}  errorResponse
// @Failure      500  {object}  errorResponse
// @Router       /comments/{id}/status [patch]
func (h *CommentHandler) UpdateCommentStatus(c *gin.Context) {
	var req updateCommentStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{err.Error(), "VALIDATION_ERROR"})
		return
	}
	actorID := c.GetString(middleware.ContextKeyUserID)
	actorEmail := c.GetString(middleware.ContextKeyEmail)
	if err := h.updateStatus.Execute(c.Request.Context(),
		c.Param("id"), req.Status, actorID, actorEmail); err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": req.Status})
}

// DeleteComment godoc
// @Summary      Delete a comment
// @Description  Deletes a comment
// @Tags         comments
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  string  true  "Comment ID"
// @Success      200  {object}  map[string]string
// @Failure      401  {object}  errorResponse
// @Failure      403  {object}  errorResponse
// @Failure      404  {object}  errorResponse
// @Failure      500  {object}  errorResponse
// @Router       /comments/{id} [delete]
func (h *CommentHandler) DeleteComment(c *gin.Context) {
	actorID := c.GetString(middleware.ContextKeyUserID)
	actorEmail := c.GetString(middleware.ContextKeyEmail)
	if err := h.delete.Execute(c.Request.Context(),
		c.Param("id"), actorID, actorEmail); err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Comment deleted"})
}
