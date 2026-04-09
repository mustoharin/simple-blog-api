package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"simple-blog-api/internal/delivery/http/middleware"
	"simple-blog-api/internal/domain"
	postuc "simple-blog-api/internal/usecase/post"
)

type PostHandler struct {
	create        *postuc.CreatePostUsecase
	list          *postuc.ListPostsUsecase
	getBySlug     *postuc.GetBySlugUsecase
	update        *postuc.UpdatePostUsecase
	togglePublish *postuc.TogglePublishUsecase
	delete        *postuc.DeletePostUsecase
}

func NewPostHandler(
	create *postuc.CreatePostUsecase,
	list *postuc.ListPostsUsecase,
	getBySlug *postuc.GetBySlugUsecase,
	update *postuc.UpdatePostUsecase,
	togglePublish *postuc.TogglePublishUsecase,
	deleteUC *postuc.DeletePostUsecase,
) *PostHandler {
	return &PostHandler{
		create:        create,
		list:          list,
		getBySlug:     getBySlug,
		update:        update,
		togglePublish: togglePublish,
		delete:        deleteUC,
	}
}

func (h *PostHandler) ListPosts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	filter := domain.PostFilter{
		Query:    c.Query("q"),
		Tag:      c.Query("tag"),
		AuthorID: c.Query("author"),
		Sort:     c.Query("sort"),
		Page:     page,
		Limit:    limit,
	}
	out, err := h.list.Execute(c.Request.Context(), filter, true)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"posts": out.Posts,
		"total": out.Total,
		"page":  out.Page,
		"limit": out.Limit,
	})
}

func (h *PostHandler) GetPost(c *gin.Context) {
	p, err := h.getBySlug.Execute(c.Request.Context(), c.Param("id"), false)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, p)
}

type createPostRequest struct {
	Title         string   `json:"title"           binding:"required"`
	Slug          string   `json:"slug"`
	Content       string   `json:"content"`
	Excerpt       string   `json:"excerpt"`
	CoverImageURL string   `json:"cover_image_url"`
	TagIDs        []string `json:"tag_ids"`
}

func (h *PostHandler) CreatePost(c *gin.Context) {
	var req createPostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{err.Error(), "VALIDATION_ERROR"})
		return
	}
	actorID, _ := c.Get(middleware.ContextKeyUserID)
	actorEmail, _ := c.Get(middleware.ContextKeyEmail)
	if req.TagIDs == nil {
		req.TagIDs = []string{}
	}
	p, err := h.create.Execute(c.Request.Context(), postuc.CreatePostInput{
		Title:         req.Title,
		Slug:          req.Slug,
		Content:       req.Content,
		Excerpt:       req.Excerpt,
		CoverImageURL: req.CoverImageURL,
		AuthorID:      actorID.(string),
		AuthorEmail:   actorEmail.(string),
		TagIDs:        req.TagIDs,
	})
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, p)
}

type updatePostRequest struct {
	Title         string   `json:"title"`
	Slug          string   `json:"slug"`
	Content       string   `json:"content"`
	Excerpt       string   `json:"excerpt"`
	CoverImageURL string   `json:"cover_image_url"`
	TagIDs        []string `json:"tag_ids"`
}

func (h *PostHandler) UpdatePost(c *gin.Context) {
	var req updatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{err.Error(), "VALIDATION_ERROR"})
		return
	}
	actorID, _ := c.Get(middleware.ContextKeyUserID)
	actorEmail, _ := c.Get(middleware.ContextKeyEmail)
	if req.TagIDs == nil {
		req.TagIDs = []string{}
	}
	p, err := h.update.Execute(c.Request.Context(), postuc.UpdatePostInput{
		ID:            c.Param("id"),
		Title:         req.Title,
		Slug:          req.Slug,
		Content:       req.Content,
		Excerpt:       req.Excerpt,
		CoverImageURL: req.CoverImageURL,
		TagIDs:        req.TagIDs,
		ActorID:       actorID.(string),
		ActorEmail:    actorEmail.(string),
	})
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, p)
}

type togglePublishRequest struct {
	Published bool `json:"published"`
}

func (h *PostHandler) TogglePublish(c *gin.Context) {
	var req togglePublishRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{err.Error(), "VALIDATION_ERROR"})
		return
	}
	actorID, _ := c.Get(middleware.ContextKeyUserID)
	actorEmail, _ := c.Get(middleware.ContextKeyEmail)
	if err := h.togglePublish.Execute(c.Request.Context(), postuc.TogglePublishInput{
		PostID:     c.Param("id"),
		Published:  req.Published,
		ActorID:    actorID.(string),
		ActorEmail: actorEmail.(string),
	}); err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"published": req.Published})
}

func (h *PostHandler) DeletePost(c *gin.Context) {
	actorID, _ := c.Get(middleware.ContextKeyUserID)
	actorEmail, _ := c.Get(middleware.ContextKeyEmail)
	if err := h.delete.Execute(c.Request.Context(),
		c.Param("id"), actorID.(string), actorEmail.(string)); err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Post deleted"})
}
