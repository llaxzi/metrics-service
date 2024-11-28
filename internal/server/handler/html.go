package handler

import (
	"github.com/gin-gonic/gin"
	"metrics-service/internal/server/storage"
	"net/http"
)

type HTMLHandler interface {
	Get(ctx *gin.Context)
}

func NewHTMLHandler(storage storage.MetricsStorage) HTMLHandler {
	return &htmlHandler{storage}
}

type htmlHandler struct {
	storage storage.MetricsStorage
}

func (h *htmlHandler) Get(ctx *gin.Context) {
	if ctx.GetHeader("Content-Type") != "text/plain" {

		ctx.String(http.StatusUnsupportedMediaType, "unsupported content type")
		return
	}

	metrics := h.storage.GetMetrics()
	ctx.HTML(200, "metrics.html", metrics)

}
