package handler

import "github.com/gin-gonic/gin"

type CommentHandler struct{}

func (h *CommentHandler) ListComments(c *gin.Context)        {}
func (h *CommentHandler) CreateComment(c *gin.Context)       {}
func (h *CommentHandler) UpdateCommentStatus(c *gin.Context) {}
func (h *CommentHandler) DeleteComment(c *gin.Context)       {}
