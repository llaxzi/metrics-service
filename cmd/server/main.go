package main

import (
	"github.com/gin-gonic/gin"
	"log"
	"metrics-service/internal/server/handler"
	"metrics-service/internal/server/storage"
	"net/http"
)

func main() {

	server := gin.Default()
	server.LoadHTMLGlob("internal/server/templates/*")

	// Создаем хранилища
	metricsStorage := storage.NewMetricsStorage()

	// Создаем handler's
	counterHandler := handler.NewCounterHandler(metricsStorage)
	gaugeHandler := handler.NewGaugeHandler(metricsStorage)
	htmlHandler := handler.NewHTMLHandler(metricsStorage)

	// Рутинг
	server.POST("/update/counter/:metricName/:metricVal", counterHandler.Update)
	server.GET("/value/counter/:metricName", counterHandler.Get)

	server.POST("/update/gauge/:metricName/:metricVal", gaugeHandler.Update)
	server.GET("/value/gauge/:metricName", gaugeHandler.Get)

	server.GET("/", htmlHandler.Get)

	server.Any("/update/", func(ctx *gin.Context) {
		ctx.String(http.StatusBadRequest, "unsupported metric type")
	})

	err := server.Run("localhost:8080")
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

}
