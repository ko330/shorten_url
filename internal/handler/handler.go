package handler

import (
	"context"
	"net/http"

	"shorten_url/internal/shortener"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc *shortener.Shortener
}

func New(svc *shortener.Shortener) *Handler {
	return &Handler{svc: svc}
}

type shortenRequest struct {
	URL string `json:"url" binding:"required,url"`
}

type shortenResponse struct {
	ID       string `json:"id"`
	ShortURL string `json:"short_url"`
	URL      string `json:"url"`
}

func (h *Handler) Shorten(c *gin.Context) {
	var req shortenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id, shortURL, err := h.svc.Shorten(context.Background(), req.URL)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, shortenResponse{ID: id, ShortURL: shortURL, URL: req.URL})
}

func (h *Handler) Resolve(c *gin.Context) {
	id := c.Param("id")
	orig, err := h.svc.Resolve(context.Background(), id)
	if err == shortener.ErrNotFound {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Redirect(http.StatusFound, orig)
}

func (h *Handler) Health(c *gin.Context) {
	c.JSON(200, gin.H{"status": "ok"})
}
