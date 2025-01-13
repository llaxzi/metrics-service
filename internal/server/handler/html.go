package handler

import (
	"github.com/gin-gonic/gin"
	"metrics-service/internal/server/service"
	"net/http"
)

type HTMLHandler interface {
	Get(ctx *gin.Context)
}

func NewHTMLHandler(service service.HTMLService) HTMLHandler {
	return &htmlHandler{service}
}

type htmlHandler struct {
	service service.HTMLService
}

func (h *htmlHandler) Get(ctx *gin.Context) {
	metricsHTML, err := h.service.GenerateHTML()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	ctx.Data(200, "text/html", []byte(metricsHTML))
}
