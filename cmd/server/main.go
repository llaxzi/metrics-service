package main

import (
	"errors"
	"github.com/gin-gonic/gin"
	"log"
	apperrors "metrics-service/internal/server/errors"
	"metrics-service/internal/server/handler"
	"metrics-service/internal/server/middleware"
	"metrics-service/internal/server/repository"
	"metrics-service/internal/server/retry"
	"metrics-service/internal/server/service"
	"metrics-service/internal/server/storage"
	"syscall"
	"time"
)

var (
	mid            middleware.Middleware
	metricsStorage storage.MetricsStorage
	repo           repository.Repository
	diskW          storage.DiskWriter
)

func init() {
	// Обрабатываем аргументы командной строки
	parseFlags()
	// Создаем middleware (логгер, gzip)
	mid = middleware.NewMiddleware()
	err := mid.InitializeZap(flagLogLevel)
	if err != nil {
		return
	}
	// Создаем хранилище
	metricsStorage = storage.NewMetricsStorage()
	// Создаем repository
	repo, err = repository.NewRepository(flagDatabaseDSN)
	if err != nil {
		log.Fatalf("Failed to initialize repository: %v", err)
	}
	// Создаем diskWriter
	diskW, err = storage.NewDiskWriter(metricsStorage, flagFileStoragePath)
	if err != nil {
		log.Fatalf("Failed to create disk writer: %v", err)
	}
}

func main() {

	// Загружаем storage из файла, если необходимо
	if flagDatabaseDSN == "" && flagRestore {
		reader, err := storage.NewDiskReader(metricsStorage, flagFileStoragePath)
		if err != nil {
			log.Fatalf("Failed to restore metrics from %v: %v", flagFileStoragePath, err)
		}
		err = reader.Load()
		if err != nil {
			log.Fatalf("Failed to restore metrics from %v: %v", flagFileStoragePath, err)
		}
		err = reader.Close()
		if err != nil {
			log.Fatalf("Failed to close diskReader: %v", err)
		}
		log.Printf("Read metrics from file: %v\n", flagFileStoragePath)
	}

	serviceRetryer := retry.NewRetryer()
	serviceRetryer.SetConditionFunc(func(err error) bool {
		return errors.Is(err, apperrors.ErrPgConnExc)
	})

	// Подготавливаем бд
	if err := serviceRetryer.Retry(repo.Bootstrap); err != nil {
		log.Printf("Failed to set up sql environment: %v", err)
	}
	defer func(repo repository.Repository) {
		err := repo.Close()
		if err != nil {
			log.Fatalf("Failed to close repository: %v", err)
		}
	}(repo)

	isStoreInterval := flagStoreInterval > 0

	saveRetryer := retry.NewRetryer()
	saveRetryer.SetConditionFunc(func(err error) bool {
		return errors.Is(err, apperrors.ErrPgConnExc) || errors.Is(err, syscall.EBUSY)
	})

	// Сохранение данных на диск
	if flagDatabaseDSN == "" && isStoreInterval {
		ticker := time.NewTicker(time.Duration(flagStoreInterval) * time.Second)
		defer ticker.Stop()
		go func() {
			for range ticker.C {
				err := saveRetryer.Retry(diskW.Save)
				if err != nil {
					log.Printf("Failed to save metrics on disk: %v\n", err)
				}
			}
		}()
	}

	// Создаем service'ы
	metricsService := service.NewMetricsService(metricsStorage, diskW, repo, flagDatabaseDSN != "", flagDatabaseDSN == "" && !isStoreInterval, serviceRetryer)
	htmlService := service.NewHTMLService(metricsStorage, repo, flagDatabaseDSN != "")
	// Создаем handler's
	metricsHandler := handler.NewMetricsHandler(metricsService)
	htmlHandler := handler.NewHTMLHandler(htmlService)

	server := gin.Default()
	// Роутинг
	// Для всех эндпоинтов используем логирование
	server.Use(mid.WithLogging())

	server.POST("/update/:metricType/:metricName/:metricVal", metricsHandler.Update)
	server.GET("/value/:metricType/:metricName", metricsHandler.Get)
	server.GET("/ping", metricsHandler.Ping)

	// Группа для методов с gzip
	gzipGroup := server.Group("")
	gzipGroup.Use(mid.WithGzip())

	gzipGroup.GET("/", htmlHandler.Get)

	gzipGroup.POST("/update/", metricsHandler.UpdateJSON)
	gzipGroup.POST("/value/", metricsHandler.GetJSON)
	gzipGroup.POST("/updates/", metricsHandler.UpdateBatch)

	err := server.Run(flagRunAddr)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

}
