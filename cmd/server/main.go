package main

import (
	"github.com/gin-gonic/gin"
	"log"
	"metrics-service/internal/server/handler"
	"metrics-service/internal/server/middleware"
	"metrics-service/internal/server/storage"
)

func main() {

	// Обрабатываем аргументы командной строки
	parseFlags()

	// Создаем логгер
	mid := middleware.NewMiddleware()
	err := mid.InitializeZap(flagLogLevel)
	if err != nil {
		return
	}

	server := gin.Default()

	// Создаем хранилище
	metricsStorage := storage.NewMetricsStorage()

	// Создаем handler's

	metricsHandler := handler.NewMetricsHandler(metricsStorage)

	htmlHandler := handler.NewHTMLHandler(metricsStorage)

	// Роутинг

	// Для всех эндпоинтов используем логирование
	server.Use(mid.WithLogging())

	server.POST("/update/:metricType/:metricName/:metricVal", metricsHandler.Update)
	server.GET("/value/:metricType/:metricName", metricsHandler.Get)

	gzipGroup := server.Group("")
	gzipGroup.Use(mid.WithGzip())

	gzipGroup.GET("/", htmlHandler.Get)

	gzipGroup.POST("/update/", metricsHandler.UpdateJSON)
	gzipGroup.POST("/value/", metricsHandler.GetJSON)

	err = server.Run(flagRunAddr)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

}
