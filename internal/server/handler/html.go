package handler

import (
	"github.com/gin-gonic/gin"
	"metrics-service/internal/server/service"
)

type HTMLHandler interface {
	Get(ctx *gin.Context)
}

func NewHTMLHandler(service service.HtmlService) HTMLHandler {
	return &htmlHandler{service}
}

type htmlHandler struct {
	service service.HtmlService
}

func (h *htmlHandler) Get(ctx *gin.Context) {
	metricsHTML := h.service.GenerateHtml()
	ctx.Data(200, "text/html", []byte(metricsHTML))
}
