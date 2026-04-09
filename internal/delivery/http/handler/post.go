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

// ListPosts godoc
// @Summary      List posts
// @Description  Returns a paginated list of blog posts
// @Tags         posts
// @Produce      json
// @Param        page    query  int     false  "Page number"     default(1)
// @Param        limit   query  int     false  "Items per page"  default(20)
// @Param        q       query  string  false  "Search query"
// @Param        tag     query  string  false  "Filter by tag"
// @Param        author  query  string  false  "Filter by author ID"
// @Param        sort    query  string  false  "Sort order"
// @Success      200  {object}  map[string]interface{}
// @Failure      500  {object}  errorResponse
// @Router       /posts [get]
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

// GetPost godoc
// @Summary      Get a post
// @Description  Returns a single published blog post by slug or UUID
// @Tags         posts
// @Produce      json
// @Param        id  path  string  true  "Post slug or UUID"
// @Success      200  {object}  domain.Post
// @Failure      404  {object}  errorResponse
// @Failure      500  {object}  errorResponse
// @Router       /posts/{id} [get]
func (h *PostHandler) GetPost(c *gin.Context) {
	p, err := h.getBySlug.Execute(c.Request.Context(), c.Param("id"), false, true)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, p)
}

// PreviewPost godoc
// @Summary      Preview a post (admin)
// @Description  Returns a post by slug or UUID regardless of publish status. Requires post:edit permission.
// @Tags         posts
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  string  true  "Post slug or UUID"
// @Success      200  {object}  domain.Post
// @Failure      401  {object}  errorResponse
// @Failure      403  {object}  errorResponse
// @Failure      404  {object}  errorResponse
// @Failure      500  {object}  errorResponse
// @Router       /posts/{id}/preview [get]
func (h *PostHandler) PreviewPost(c *gin.Context) {
	p, err := h.getBySlug.Execute(c.Request.Context(), c.Param("id"), true, false)
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

// CreatePost godoc
// @Summary      Create a post
// @Description  Creates a new blog post
// @Tags         posts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body  createPostRequest  true  "Post data"
// @Success      201  {object}  domain.Post
// @Failure      400  {object}  errorResponse
// @Failure      401  {object}  errorResponse
// @Failure      403  {object}  errorResponse
// @Failure      500  {object}  errorResponse
// @Router       /posts [post]
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

// UpdatePost godoc
// @Summary      Update a post
// @Description  Updates an existing blog post
// @Tags         posts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path  string           true  "Post ID"
// @Param        body  body  updatePostRequest  true  "Post data"
// @Success      200  {object}  domain.Post
// @Failure      400  {object}  errorResponse
// @Failure      401  {object}  errorResponse
// @Failure      403  {object}  errorResponse
// @Failure      404  {object}  errorResponse
// @Failure      500  {object}  errorResponse
// @Router       /posts/{id} [put]
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

// TogglePublish godoc
// @Summary      Toggle post publish status
// @Description  Publishes or unpublishes a blog post
// @Tags         posts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path  string               true  "Post ID"
// @Param        body  body  togglePublishRequest  true  "Publish flag"
// @Success      200  {object}  map[string]bool
// @Failure      400  {object}  errorResponse
// @Failure      401  {object}  errorResponse
// @Failure      403  {object}  errorResponse
// @Failure      404  {object}  errorResponse
// @Failure      500  {object}  errorResponse
// @Router       /posts/{id}/publish [patch]
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

// DeletePost godoc
// @Summary      Delete a post
// @Description  Deletes a blog post
// @Tags         posts
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  string  true  "Post ID"
// @Success      200  {object}  map[string]string
// @Failure      401  {object}  errorResponse
// @Failure      403  {object}  errorResponse
// @Failure      404  {object}  errorResponse
// @Failure      500  {object}  errorResponse
// @Router       /posts/{id} [delete]
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
