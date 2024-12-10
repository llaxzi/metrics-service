package main

import (
	"github.com/gin-gonic/gin"
	"log"
	"metrics-service/internal/server/handler"
	"metrics-service/internal/server/storage"
)

func main() {

	// Обрабатываем аргументы командной строки
	parseFlags()

	server := gin.Default()

	// Создаем хранилище
	metricsStorage := storage.NewMetricsStorage()

	// Создаем handler's

	metricsHandler := handler.NewMetricsHandler(metricsStorage)

	htmlHandler := handler.NewHTMLHandler(metricsStorage)

	// Роутинг
	server.POST("/update/:metricType/:metricName/:metricVal", metricsHandler.Update)

	server.GET("/value/:metricType/:metricName", metricsHandler.Get)

	server.GET("/", htmlHandler.Get)

	err := server.Run(flagRunAddr)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

}
