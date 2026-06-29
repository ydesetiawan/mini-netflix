package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ydesetiawan/mini-netflix/internal/model"
	"github.com/ydesetiawan/mini-netflix/internal/service"
)

type ContentHandler struct {
	svc *service.ContentService
}

func NewContentHandler(svc *service.ContentService) *ContentHandler {
	return &ContentHandler{svc: svc}
}

func (h *ContentHandler) Create(c *gin.Context) {
	var req model.CreateContentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	content, err := h.svc.Create(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, content)
}

func (h *ContentHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	content, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "content not found"})
		return
	}
	c.JSON(http.StatusOK, content)
}

func (h *ContentHandler) RecordWatch(c *gin.Context) {
	id := c.Param("id")
	userID, _ := c.Get("user_id")

	var body struct {
		Progress int `json:"progress"`
	}
	c.ShouldBindJSON(&body)

	ev := model.WatchEvent{
		ContentID: id,
		Progress:  body.Progress,
	}
	if uid, ok := userID.(string); ok {
		ev.UserID = uid
	}

	if err := h.svc.RecordWatch(c.Request.Context(), ev); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "watch event recorded"})
}

func (h *ContentHandler) Rate(c *gin.Context) {
	id := c.Param("id")
	userID := c.GetString("user_id")

	var req model.RateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.svc.Rate(c.Request.Context(), userID, id, req.Score); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "rating saved"})
}
