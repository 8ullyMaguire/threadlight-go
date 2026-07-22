package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/opencode-ai/polaris/internal/model"
	"github.com/opencode-ai/polaris/internal/services"
)

type PostHandler struct {
	postSvc   *services.PostService
	userSvc   *services.UserService
	mediaSvc  *services.MediaService
	creditSvc *services.CreditService
	configSvc *services.ConfigService
}

func NewPostHandler(postSvc *services.PostService, userSvc *services.UserService, mediaSvc *services.MediaService, creditSvc *services.CreditService, configSvc *services.ConfigService) *PostHandler {
	return &PostHandler{postSvc: postSvc, userSvc: userSvc, mediaSvc: mediaSvc, creditSvc: creditSvc, configSvc: configSvc}
}

func (h *PostHandler) Create(c *gin.Context) {
	var req model.CreatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	uid, _ := c.Get("user_id")

	resp, err := h.postSvc.CreatePost(c.Request.Context(), uid.(int64), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *PostHandler) List(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	posts, err := h.postSvc.ListPosts(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, posts)
}

func (h *PostHandler) GetCount(c *gin.Context) {
	count, err := h.postSvc.GetPostCount(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"count": count})
}

func (h *PostHandler) ListByAuthor(c *gin.Context) {
	authorID, err := strconv.ParseInt(c.Param("author_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid author_id"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	posts, err := h.postSvc.ListPostsByAuthor(c.Request.Context(), authorID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, posts)
}

func (h *PostHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid post id"})
		return
	}

	post, err := h.postSvc.GetPost(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, post)
}

func (h *PostHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid post id"})
		return
	}

	uid, _ := c.Get("user_id")

	var req model.UpdatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	post, err := h.postSvc.UpdatePost(c.Request.Context(), id, uid.(int64), req)
	if err == model.ErrForbidden {
		c.JSON(http.StatusForbidden, gin.H{"error": "you do not own this post"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, post)
}

func (h *PostHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid post id"})
		return
	}

	uid, _ := c.Get("user_id")

	if err := h.postSvc.ArchivePost(c.Request.Context(), id, uid.(int64)); err == model.ErrForbidden {
		c.JSON(http.StatusForbidden, gin.H{"error": "you do not own this post"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "post deleted"})
}

func (h *PostHandler) Archive(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid post id"})
		return
	}

	uid, _ := c.Get("user_id")

	if err := h.postSvc.ArchivePost(c.Request.Context(), id, uid.(int64)); err == model.ErrForbidden {
		c.JSON(http.StatusForbidden, gin.H{"error": "you do not own this post"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "post archived"})
}

func (h *PostHandler) ModRemove(c *gin.Context) {
	postID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid post id"})
		return
	}

	uid, _ := c.Get("user_id")
	user, err := h.userSvc.GetUserByID(c.Request.Context(), uid.(int64))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return
	}
	if user.TrustLevel < 3 {
		c.JSON(http.StatusForbidden, gin.H{"error": "moderator access required"})
		return
	}

	if err := h.postSvc.ArchivePost(c.Request.Context(), postID, uid.(int64)); err == model.ErrNotFound {
		c.JSON(http.StatusNotFound, gin.H{"error": "post not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "post removed by moderator"})
}

// UploadImage POST /api/v1/media/upload
func (h *PostHandler) UploadImage(c *gin.Context) {
	userID := c.GetInt64("user_id")

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required: " + err.Error()})
		return
	}
	defer file.Close()

	// Read credit action costs to check if image upload costs credits
	if h.configSvc != nil && h.creditSvc != nil {
		cfg, cfgErr := h.configSvc.Get(c.Request.Context())
		if cfgErr == nil && cfg.ImageStorageBackend != "none" {
			// Deduct credits for image upload
			actionCosts := &model.CreditActionCost{}
			if cfg.CreditActionCosts != nil {
				json.Unmarshal([]byte(*cfg.CreditActionCosts), actionCosts)
			}
			cost := int64(actionCosts.ImageUpload)
			if cost <= 0 {
				cost = 10 // default
			}
			if err := h.creditSvc.DeductCredits(c.Request.Context(), userID, cost, "image_upload", nil); err != nil {
				c.JSON(http.StatusPaymentRequired, gin.H{"error": "insufficient credits: " + err.Error()})
				return
			}
		}
	}

	media, err := h.mediaSvc.UploadImage(c.Request.Context(), userID, file, header)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "upload failed: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, media)
}
