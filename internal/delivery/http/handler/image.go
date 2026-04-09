package handler

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"simple-blog-api/internal/delivery/http/middleware"
	imageuc "simple-blog-api/internal/usecase/image"
)

type ImageHandler struct {
	upload *imageuc.UploadImageUsecase
	delete *imageuc.DeleteImageUsecase
}

func NewImageHandler(upload *imageuc.UploadImageUsecase, deleteUC *imageuc.DeleteImageUsecase) *ImageHandler {
	return &ImageHandler{upload: upload, delete: deleteUC}
}

// UploadImage godoc
// @Summary      Upload an image
// @Description  Uploads an image file (max 10MB)
// @Tags         images
// @Accept       mpfd
// @Produce      json
// @Security     BearerAuth
// @Param        file  formData  file  true  "Image file"
// @Success      201  {object}  map[string]interface{}
// @Failure      400  {object}  errorResponse
// @Failure      401  {object}  errorResponse
// @Failure      403  {object}  errorResponse
// @Failure      413  {object}  errorResponse
// @Failure      422  {object}  errorResponse
// @Failure      500  {object}  errorResponse
// @Router       /images [post]
func (h *ImageHandler) UploadImage(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{"file field is required", "VALIDATION_ERROR"})
		return
	}
	defer file.Close()

	data, err := io.ReadAll(io.LimitReader(file, 11*1024*1024))
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse{"failed to read file", "INTERNAL_ERROR"})
		return
	}

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	actorID, _ := c.Get(middleware.ContextKeyUserID)
	actorEmail, _ := c.Get(middleware.ContextKeyEmail)

	img, err := h.upload.Execute(c.Request.Context(), imageuc.UploadInput{
		Filename:      header.Filename,
		ContentType:   contentType,
		Data:          data,
		UploaderID:    actorID.(string),
		UploaderEmail: actorEmail.(string),
	})
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, img)
}

// DeleteImage godoc
// @Summary      Delete an image
// @Description  Deletes an image by ID
// @Tags         images
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  string  true  "Image ID"
// @Success      200  {object}  map[string]string
// @Failure      401  {object}  errorResponse
// @Failure      403  {object}  errorResponse
// @Failure      404  {object}  errorResponse
// @Failure      500  {object}  errorResponse
// @Router       /images/{id} [delete]
func (h *ImageHandler) DeleteImage(c *gin.Context) {
	actorID, _ := c.Get(middleware.ContextKeyUserID)
	actorEmail, _ := c.Get(middleware.ContextKeyEmail)
	if err := h.delete.Execute(c.Request.Context(),
		c.Param("id"), actorID.(string), actorEmail.(string)); err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Image deleted"})
}
