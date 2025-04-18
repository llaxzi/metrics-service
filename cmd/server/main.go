package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/llaxzi/retryables/v2"

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"

	apperrors "metrics-service/internal/server/errors"
	"metrics-service/internal/server/handler"
	"metrics-service/internal/server/middleware"
	storageP "metrics-service/internal/server/storage"
)

func main() {

	printBuildInfo()

	parseFlags()
	parseJSON()
	overrideEnv()

	// Создаем middleware (логгер, gzip)
	mid, err := middleware.NewMiddleware([]byte(flagHashKey), cryptoKeyPath)
	if err != nil {
		log.Fatalf("Failed to initialize middleware: %v", err)
	}

	err = mid.InitializeZap(flagLogLevel)
	if err != nil {
		log.Fatalf("Failed to initialize middleware: %v", err)
	}

	// Создаем storage
	storage, err := storageP.NewStorage(flagDatabaseDSN, flagFileStoragePath, flagRestore, flagStoreInterval)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	storageRetryer := retryables.NewRetryer(nil)
	storageRetryer.SetConditionFunc(func(err error) bool {
		return errors.Is(err, apperrors.ErrPgConnExc) || errors.Is(err, syscall.EBUSY)
	})

	// Подготавливаем бд
	err = storageRetryer.Retry(func() error {
		return storage.Bootstrap(context.Background())
	})
	if err != nil {
		log.Printf("Failed to bootstrap storage: %v", err)
	}
	defer storage.Close()

	isSync := flagStoreInterval <= 0
	// Сохранение данных на диск
	if flagDatabaseDSN == "" && !isSync {
		ticker := time.NewTicker(time.Duration(flagStoreInterval) * time.Second)
		defer ticker.Stop()
		go func() {
			for range ticker.C {
				err = storageRetryer.Retry(func() error {
					return storage.Save()
				})
				if err != nil {
					log.Printf("Failed to save metrics on disk: %v\n", err)
				}
			}
		}()
	}

	// Создаем handler's
	metricsHandler := handler.NewMetricsHandler(storage, storageRetryer, isSync)
	htmlHandler := handler.NewHTMLHandler(storage, storageRetryer)

	router := gin.Default()
	// Роутинг
	// Для всех эндпоинтов используем логирование
	router.Use(mid.WithLogging())

	router.POST("/update/:metricType/:metricName/:metricVal", metricsHandler.Update)
	router.GET("/value/:metricType/:metricName", metricsHandler.Get)
	router.GET("/ping", metricsHandler.Ping)

	// Группа для методов с gzip
	gzipGroup := router.Group("")
	gzipGroup.Use(mid.WithDecryptRSA(), mid.WithGzip())

	gzipGroup.GET("/", htmlHandler.Get)

	gzipGroup.POST("/update/", metricsHandler.UpdateJSON)
	gzipGroup.POST("/value/", metricsHandler.GetJSON)
	gzipGroup.POST("/updates/", metricsHandler.UpdateBatch)

	pprof.Register(router, "dev/pprof")

	srv := &http.Server{
		Addr:    flagRunAddr,
		Handler: router,
	}

	// через этот канал сообщим основному потоку, что соединения закрыты
	idleConnsClosed := make(chan struct{})
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT, os.Interrupt)
	go func() {
		<-sigint
		if err = srv.Shutdown(context.Background()); err != nil {
			log.Printf("HTTP server Shutdown: %v", err)
		}
		// сообщаем основному потоку,
		// что все сетевые соединения обработаны и закрыты
		close(idleConnsClosed)
	}()

	if err = srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("Failed to start server: %v", err)
	}

	<-idleConnsClosed
	fmt.Println("Server Shutdown gracefully")

}
